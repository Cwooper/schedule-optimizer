package generator

import (
	"context"

	"schedule-optimizer/internal/cache"
)

// backtrackParams configures the backtracking algorithm.
type backtrackParams struct {
	groups       []courseGroup
	forced       []*sectionData // Sections that must be in every schedule
	forcedMask   TimeMask       // Combined mask of forced sections
	minCourses   int
	maxCourses   int
	limit        int
}

// backtrack finds all valid schedule combinations using recursive backtracking.
// It explores course groups in order, pruning branches that cannot lead to valid schedules.
// All generated schedules include the forced sections.
func backtrack(ctx context.Context, p backtrackParams) []Schedule {
	maxFromGroups := min(p.maxCourses - len(p.forced), len(p.groups))
	minFromGroups := max(p.minCourses - len(p.forced), 0)

	initialCap := min(p.limit, 100)
	results := make([]Schedule, 0, initialCap)
	current := make([]*sectionData, 0, maxFromGroups)
	currentMask := p.forcedMask

	var generate func(groupIdx int)
	generate = func(groupIdx int) {
		if ctx.Err() != nil || len(results) >= p.limit {
			return
		}

		// Record valid schedule if we have enough courses (including forced)
		if len(current) >= minFromGroups {
			results = append(results, buildSchedule(p.forced, current))
		}

		// Stop if we've reached max courses or exhausted all groups
		if len(current) >= maxFromGroups || groupIdx >= len(p.groups) {
			return
		}

		// Pruning: can we still reach the minimum?
		remaining := len(p.groups) - groupIdx
		if len(current)+remaining < minFromGroups {
			return
		}

		// Try each remaining group
		for g := groupIdx; g < len(p.groups); g++ {
			for _, section := range p.groups[g].sections {
				if currentMask.Conflicts(section.mask) {
					continue
				}

				current = append(current, section)
				oldMask := currentMask
				currentMask = currentMask.Merge(section.mask)

				generate(g + 1)

				current = current[:len(current)-1]
				currentMask = oldMask
			}
		}
	}

	generate(0)
	return results
}

// buildSchedule creates a Schedule from forced and selected sections.
func buildSchedule(forced, selected []*sectionData) Schedule {
	courses := make([]*cache.Course, 0, len(forced)+len(selected))
	for _, s := range forced {
		courses = append(courses, s.course)
	}
	for _, s := range selected {
		courses = append(courses, s.course)
	}
	return Schedule{Courses: courses}
}
