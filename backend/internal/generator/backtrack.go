package generator

import (
	"context"

	"schedule-optimizer/internal/cache"
)

// backtrackParams configures the backtracking algorithm.
type backtrackParams struct {
	groups      []courseGroup
	numRequired int // First N groups are required (must all be in every schedule)
	minCourses  int
	maxCourses  int
	limit       int
}

// backtrack finds all valid schedule combinations using recursive backtracking.
// The first numRequired groups are required (must all be in every schedule).
// Remaining groups are optional. It explores groups in order, pruning branches
// that cannot lead to valid schedules.
func backtrack(ctx context.Context, p backtrackParams) []Schedule {
	initialCap := min(p.limit, 100)
	results := make([]Schedule, 0, initialCap)
	current := make([]*sectionData, 0, p.maxCourses)
	var currentMask TimeMask

	var generate func(groupIdx int)
	generate = func(groupIdx int) {
		if ctx.Err() != nil || len(results) >= p.limit {
			return
		}

		// For required groups, we must pick exactly one from each
		if groupIdx < p.numRequired {
			// Try each section in this required group
			for _, section := range p.groups[groupIdx].sections {
				if currentMask.Conflicts(section.mask) {
					continue
				}

				current = append(current, section)
				oldMask := currentMask
				currentMask = currentMask.Merge(section.mask)

				generate(groupIdx + 1)

				current = current[:len(current)-1]
				currentMask = oldMask
			}
			return
		}

		// We've filled all required groups, now handle optional groups
		// Record valid schedule if we have enough courses
		if len(current) >= p.minCourses {
			results = append(results, buildSchedule(current))
		}

		// Stop if we've reached max courses or exhausted all groups
		if len(current) >= p.maxCourses || groupIdx >= len(p.groups) {
			return
		}

		// Pruning: can we still reach the minimum?
		remaining := len(p.groups) - groupIdx
		if len(current)+remaining < p.minCourses {
			return
		}

		// Try each remaining optional group
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

// buildSchedule creates a Schedule from selected sections.
func buildSchedule(selected []*sectionData) Schedule {
	courses := make([]*cache.Course, 0, len(selected))
	for _, s := range selected {
		courses = append(courses, s.course)
	}
	return Schedule{Courses: courses}
}
