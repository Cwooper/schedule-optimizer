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
	cache   *cache.ScheduleCache
	queries *store.Queries
}

// NewService creates a new schedule generator service.
func NewService(c *cache.ScheduleCache, q *store.Queries) *Service {
	return &Service{cache: c, queries: q}
}

// Generate produces all valid schedule combinations for the requested courses.
func (s *Service) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
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
	requiredGroups, reqAsyncs, reqResults := s.buildCourseGroups(ctx, req.Term, requiredSpecs, blockedMask)
	optionalGroups, optAsyncs, optResults := s.buildCourseGroups(ctx, req.Term, optionalSpecs, blockedMask)

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

	// Sort groups by section count (smallest first = better pruning)
	// Required groups first, then optional groups
	slices.SortFunc(requiredGroups, func(a, b courseGroup) int {
		return len(a.sections) - len(b.sections)
	})
	slices.SortFunc(optionalGroups, func(a, b courseGroup) int {
		return len(a.sections) - len(b.sections)
	})

	allGroups := append(requiredGroups, optionalGroups...)
	numRequired := len(requiredGroups)
	totalCourses := len(allGroups)

	// Default minCourses to totalCourses if not specified (0), but at least numRequired
	effectiveMin := req.MinCourses
	if effectiveMin == 0 {
		effectiveMin = totalCourses
	}
	effectiveMin = max(effectiveMin, numRequired)

	minCourses, maxCourses := clampBounds(effectiveMin, req.MaxCourses, totalCourses)

	schedules := backtrack(ctx, backtrackParams{
		groups:      allGroups,
		numRequired: numRequired,
		minCourses:  minCourses,
		maxCourses:  maxCourses,
		limit:       MaxSchedulesToGenerate,
	})

	for i := range schedules {
		scoreSchedule(&schedules[i])
	}
	slices.SortFunc(schedules, func(a, b Schedule) int {
		return cmp.Compare(b.Score, a.Score) // Descending by score
	})

	totalGenerated := len(schedules)

	// Only return top N schedules to client
	if len(schedules) > MaxSchedulesToReturn {
		schedules = schedules[:MaxSchedulesToReturn]
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

// buildCourseGroups fetches sections from cache, filters by blocked times and allowed CRNs, and groups by course.
func (s *Service) buildCourseGroups(ctx context.Context, term string, specs []CourseSpec, blockedMask TimeMask) ([]courseGroup, []*cache.Course, []CourseResult) {
	var groups []courseGroup
	var asyncs []*cache.Course
	var results []CourseResult

	for _, spec := range specs {
		courseKey := spec.Subject + ":" + spec.CourseNumber
		displayName := spec.Subject + " " + spec.CourseNumber
		sections := s.cache.GetCoursesByCourseCode(term, courseKey)

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
