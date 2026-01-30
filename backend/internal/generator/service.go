package generator

import (
	"cmp"
	"context"
	"slices"
	"strings"
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

	// Build forced sections first
	forced, forcedMask, forcedConflict := s.buildForcedSections(req.Term, req.ForcedCRNs, blockedMask)
	if forcedConflict {
		// Forced CRNs conflict with each other or blocked times - no valid schedules
		return &GenerateResponse{
			Schedules:     nil,
			CourseResults: nil,
			Stats: GenerateStats{
				TotalGenerated: 0,
				TimeMs:         float64(time.Since(start).Microseconds()) / 1000,
			},
		}, nil
	}

	groups, asyncs, courseResults := s.buildCourseGroups(ctx, req.Term, req.Courses, forcedMask)

	// Sort groups by section count (smallest first = better pruning)
	slices.SortFunc(groups, func(a, b courseGroup) int {
		return len(a.sections) - len(b.sections)
	})

	totalCourses := len(req.Courses) + len(forced)
	minCourses, maxCourses := clampBounds(req.MinCourses, req.MaxCourses, totalCourses)

	schedules := backtrack(ctx, backtrackParams{
		groups:     groups,
		forced:     forced,
		forcedMask: forcedMask,
		minCourses: minCourses,
		maxCourses: maxCourses,
		limit:      MaxSchedulesToGenerate,
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
		Asyncs:        asyncs,
		CourseResults: courseResults,
		Stats: GenerateStats{
			TotalGenerated: totalGenerated,
			TimeMs:         float64(time.Since(start).Microseconds()) / 1000,
		},
	}, nil
}

// buildForcedSections looks up forced CRNs and builds their combined mask.
// Returns the sections, combined mask, and whether there's a conflict.
func (s *Service) buildForcedSections(term string, crns []string, blockedMask TimeMask) ([]*sectionData, TimeMask, bool) {
	if len(crns) == 0 {
		return nil, blockedMask, false
	}

	forced := make([]*sectionData, 0, len(crns))
	combinedMask := blockedMask

	for _, crn := range crns {
		course, ok := s.cache.GetCourse(term, crn)
		if !ok {
			continue // CRN not found, skip
		}

		mask := FromMeetingTimes(course.MeetingTimes)

		// Check for conflict with blocked times or other forced sections
		if combinedMask.Conflicts(mask) {
			return nil, TimeMask{}, true
		}

		combinedMask = combinedMask.Merge(mask)
		forced = append(forced, &sectionData{
			course: course,
			mask:   mask,
		})
	}

	return forced, combinedMask, false
}

// buildCourseGroups fetches sections from cache, filters by blocked times, and groups by course.
func (s *Service) buildCourseGroups(ctx context.Context, term string, courseNames []string, blockedMask TimeMask) ([]courseGroup, []*cache.Course, []CourseResult) {
	var groups []courseGroup
	var asyncs []*cache.Course
	var results []CourseResult

	for _, name := range courseNames {
		courseKey := normalizeCourseKey(name)
		sections := s.cache.GetCoursesByCourseCode(term, courseKey)

		if len(sections) == 0 {
			subject, courseNum := splitCourseKey(courseKey)
			exists, _ := s.queries.CourseExistsAnyTerm(ctx, store.CourseExistsAnyTermParams{
				Subject:      subject,
				CourseNumber: courseNum,
			})
			if exists > 0 {
				results = append(results, CourseResult{Name: name, Status: StatusNotOffered})
			} else {
				results = append(results, CourseResult{Name: name, Status: StatusNotExists})
			}
			continue
		}

		var group courseGroup
		group.courseKey = courseKey
		var asyncCount, blockedCount int

		for _, sec := range sections {
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
				Name:   name,
				Status: StatusFound,
				Count:  len(group.sections),
			})
		} else if asyncCount > 0 {
			results = append(results, CourseResult{Name: name, Status: StatusAsyncOnly})
		} else if blockedCount > 0 {
			results = append(results, CourseResult{Name: name, Status: StatusBlocked})
		}
	}

	return groups, asyncs, results
}

// clampBounds ensures min/max are within valid ranges.
func clampBounds(minReq, maxReq, numCourses int) (int, int) {
	minCourses := max(minReq, DefaultMinCourses)
	minCourses = min(minCourses, numCourses)

	maxCourses := maxReq
	if maxCourses < 1 {
		maxCourses = DefaultMaxCourses
	}
	maxCourses = min(maxCourses, numCourses)
	maxCourses = max(maxCourses, minCourses)

	return minCourses, maxCourses
}

// normalizeCourseKey converts various formats to "SUBJECT:NUMBER".
// Examples: "CSCI 241" -> "CSCI:241", "csci241" -> "CSCI:241"
func normalizeCourseKey(name string) string {
	name = strings.ToUpper(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, ":", " ")

	// Split on space first
	parts := strings.Fields(name)
	if len(parts) == 2 {
		return parts[0] + ":" + parts[1]
	}

	// No space - try to split letters from numbers
	var subject, number strings.Builder
	inNumber := false
	for _, r := range name {
		if r >= '0' && r <= '9' {
			inNumber = true
			number.WriteRune(r)
		} else if !inNumber {
			subject.WriteRune(r)
		}
	}

	if subject.Len() > 0 && number.Len() > 0 {
		return subject.String() + ":" + number.String()
	}

	return name
}

// splitCourseKey splits "CSCI:241" into ("CSCI", "241").
func splitCourseKey(key string) (string, string) {
	parts := strings.Split(key, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return key, ""
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
