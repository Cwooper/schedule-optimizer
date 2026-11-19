package generator

import (
	"cmp"
	"context"
	"slices"
	"time"

	"schedule-optimizer/internal/cache"
	"schedule-optimizer/internal/store"
)

// Service handles schedule generation using bitmask-based conflict detection.
type Service struct {
	queries *store.Queries
}

// NewService creates a new schedule generator service.
func NewService(q *store.Queries) *Service {
	return &Service{queries: q}
}

// Generate produces all valid schedule combinations for the requested courses
// from one snapshot of the term, so a cache reload mid-request cannot mix
// two loads.
func (s *Service) Generate(ctx context.Context, term *cache.TermData, req GenerateRequest) (*GenerateResponse, error) {
	start := time.Now()

	blockedMask := FromBlockedTimes(req.BlockedTimes)

	// Separate required vs optional specs
	var requiredSpecs, optionalSpecs []CourseSpec
	for _, spec := range req.CourseSpecs {
		if spec.Required {
			requiredSpecs = append(requiredSpecs, spec)
		} else {
			optionalSpecs = append(optionalSpecs, spec)
		}
	}

	// Build course groups for all specs
	requiredGroups, reqAsyncs, reqResults := s.buildCourseGroups(ctx, term, requiredSpecs, blockedMask)
	optionalGroups, optAsyncs, optResults := s.buildCourseGroups(ctx, term, optionalSpecs, blockedMask)

	// Check if any required course has no valid sections
	if len(requiredGroups) < len(requiredSpecs) {
		// At least one required course has no scheduleable sections - no valid schedules
		return &GenerateResponse{
			Schedules:     nil,
			Asyncs:        append(reqAsyncs, optAsyncs...),
			CourseResults: append(reqResults, optResults...),
			Stats: GenerateStats{
				TotalGenerated: 0,
				TimeMs:         float64(time.Since(start).Microseconds()) / 1000,
			},
		}, nil
	}

	// Fewest sections first prunes best. The course key breaks ties so the order,
	// and with it compareSchedules' CRN tiebreak, doesn't depend on the order
	// courses were requested in.
	byPruning := func(a, b courseGroup) int {
		return cmp.Or(cmp.Compare(len(a.sections), len(b.sections)), cmp.Compare(a.courseKey, b.courseKey))
	}
	slices.SortFunc(requiredGroups, byPruning)
	slices.SortFunc(optionalGroups, byPruning)

	allGroups := append(requiredGroups, optionalGroups...)
	numRequired := len(requiredGroups)
	totalCourses := len(allGroups)

	// Default minCourses to totalCourses if not specified (0), but at least numRequired
	effectiveMin := req.MinCourses
	userSetMin := effectiveMin > 0
	if effectiveMin == 0 {
		effectiveMin = totalCourses
	}
	effectiveMin = max(effectiveMin, numRequired)

	// When user didn't set min, allow one fewer course as fallback
	fallbackMin := effectiveMin
	if !userSetMin && effectiveMin > numRequired && effectiveMin > 1 {
		fallbackMin = effectiveMin - 1
	}

	minCourses, maxCourses := clampBounds(fallbackMin, req.MaxCourses, totalCourses)

	prefs := DefaultPreferences()
	if req.Preferences != nil {
		prefs = *req.Preferences
	}
	r := newRanker(newScorer(prefs), MaxSchedulesToReturn, effectiveMin)
	backtrack(ctx, backtrackParams{
		groups:      allGroups,
		numRequired: numRequired,
		minCourses:  minCourses,
		maxCourses:  maxCourses,
		limit:       MaxSchedulesToGenerate,
	}, r.visit)
	schedules := r.results()
	totalGenerated := r.total

	// If we used a fallback min and there are full-count schedules, drop the
	// shorter ones (prefer a full course load). Course count ranks first, so the
	// kept schedules hold every full one that made the cut.
	if !userSetMin && fallbackMin < effectiveMin && r.full > 0 {
		schedules = slices.DeleteFunc(schedules, func(s Schedule) bool {
			return len(s.Courses) < effectiveMin
		})
		totalGenerated = r.full
	}

	return &GenerateResponse{
		Schedules:     schedules,
		Asyncs:        append(reqAsyncs, optAsyncs...),
		CourseResults: append(reqResults, optResults...),
		Stats: GenerateStats{
			TotalGenerated: totalGenerated,
			TimeMs:         float64(time.Since(start).Microseconds()) / 1000,
		},
	}, nil
}

// compareSchedules ranks more courses first, since scores average over the
// whole week and would otherwise favor lighter schedules; then by score, then
// by CRNs so equal scores keep a stable order across requests.
func compareSchedules(a, b Schedule) int {
	if c := cmp.Compare(len(b.Courses), len(a.Courses)); c != 0 {
		return c
	}
	if c := cmp.Compare(b.Score, a.Score); c != 0 {
		return c
	}
	return slices.CompareFunc(a.Courses, b.Courses, func(x, y *cache.Course) int {
		return cmp.Compare(x.CRN, y.CRN)
	})
}

// buildCourseGroups fetches sections from the term, filters by blocked times and allowed CRNs, and groups by course.
func (s *Service) buildCourseGroups(ctx context.Context, term *cache.TermData, specs []CourseSpec, blockedMask TimeMask) ([]courseGroup, []*cache.Course, []CourseResult) {
	var groups []courseGroup
	var asyncs []*cache.Course
	var results []CourseResult

	for _, spec := range specs {
		courseKey := spec.Subject + ":" + spec.CourseNumber
		displayName := spec.Subject + " " + spec.CourseNumber
		sections := term.ByCourseCode[courseKey]

		if len(sections) == 0 {
			exists, _ := s.queries.CourseExistsAnyTerm(ctx, store.CourseExistsAnyTermParams{
				Subject:      spec.Subject,
				CourseNumber: spec.CourseNumber,
			})
			if exists > 0 {
				results = append(results, CourseResult{Name: displayName, Status: StatusNotOffered})
			} else {
				results = append(results, CourseResult{Name: displayName, Status: StatusNotExists})
			}
			continue
		}

		// Build allowed CRN set if specified
		var allowedCRNs map[string]bool
		if len(spec.AllowedCRNs) > 0 {
			allowedCRNs = make(map[string]bool, len(spec.AllowedCRNs))
			for _, crn := range spec.AllowedCRNs {
				allowedCRNs[crn] = true
			}
		}

		var group courseGroup
		group.courseKey = courseKey
		var asyncCount, blockedCount, filteredCount int

		for _, sec := range sections {
			// Filter by allowed CRNs if specified
			if allowedCRNs != nil && !allowedCRNs[sec.CRN] {
				filteredCount++
				continue
			}

			if isAsyncOrTBD(sec) {
				asyncs = append(asyncs, sec)
				asyncCount++
				continue
			}

			mask := FromMeetingTimes(sec.MeetingTimes)

			if blockedMask.Conflicts(mask) {
				blockedCount++
				continue
			}

			group.sections = append(group.sections, &sectionData{
				course: sec,
				mask:   mask,
			})
		}

		if len(group.sections) > 0 {
			groups = append(groups, group)
			results = append(results, CourseResult{
				Name:   displayName,
				Status: StatusFound,
				Count:  len(group.sections),
			})
		} else if asyncCount > 0 {
			results = append(results, CourseResult{Name: displayName, Status: StatusAsyncOnly})
		} else if filteredCount > 0 && blockedCount == 0 {
			// All sections filtered by AllowedCRNs (none of the specified CRNs exist)
			results = append(results, CourseResult{Name: displayName, Status: StatusCRNFiltered})
		} else if blockedCount > 0 {
			results = append(results, CourseResult{Name: displayName, Status: StatusBlocked})
		}
	}

	return groups, asyncs, results
}

// clampBounds ensures min/max are within valid ranges.
func clampBounds(minReq, maxReq, numCourses int) (int, int) {
	minCourses := max(minReq, 1)
	minCourses = min(minCourses, numCourses)

	maxCourses := maxReq
	if maxCourses < 1 {
		maxCourses = DefaultMaxCourses
	}
	maxCourses = min(maxCourses, numCourses)
	maxCourses = max(maxCourses, minCourses)

	return minCourses, maxCourses
}

// isAsyncOrTBD returns true if the course has no scheduled meeting times.
func isAsyncOrTBD(c *cache.Course) bool {
	if len(c.MeetingTimes) == 0 {
		return true
	}
	for _, mt := range c.MeetingTimes {
		if mt.StartTime != "" && mt.EndTime != "" {
			// Has at least one scheduled time
			hasDay := false
			for i := 1; i <= 5; i++ { // Mon-Fri
				if mt.Days[i] {
					hasDay = true
					break
				}
			}
			if hasDay {
				return false
			}
		}
	}
	return true
}
