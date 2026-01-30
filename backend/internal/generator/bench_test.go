package generator

import (
	"context"
	"testing"

	"schedule-optimizer/internal/cache"
)

// BenchmarkFromMeetingTimes measures bitmask creation performance.
func BenchmarkFromMeetingTimes(b *testing.B) {
	meetings := []cache.MeetingTime{
		{Days: [7]bool{false, true, false, true, false, true, false}, StartTime: "0900", EndTime: "0950"},
		{Days: [7]bool{false, false, true, false, true, false, false}, StartTime: "1000", EndTime: "1150"},
	}

	for b.Loop() {
		FromMeetingTimes(meetings)
	}
}

// BenchmarkTimeMaskConflicts measures conflict detection performance.
func BenchmarkTimeMaskConflicts(b *testing.B) {
	m1 := FromMeetingTimes([]cache.MeetingTime{
		{Days: [7]bool{false, true, false, true, false, true, false}, StartTime: "0900", EndTime: "0950"},
	})
	m2 := FromMeetingTimes([]cache.MeetingTime{
		{Days: [7]bool{false, true, false, true, false, true, false}, StartTime: "1000", EndTime: "1050"},
	})

	for b.Loop() {
		m1.Conflicts(m2)
	}
}

// BenchmarkTimeMaskMerge measures mask merging performance.
func BenchmarkTimeMaskMerge(b *testing.B) {
	m1 := FromMeetingTimes([]cache.MeetingTime{
		{Days: [7]bool{false, true, false, true, false, true, false}, StartTime: "0900", EndTime: "0950"},
	})
	m2 := FromMeetingTimes([]cache.MeetingTime{
		{Days: [7]bool{false, true, false, true, false, true, false}, StartTime: "1000", EndTime: "1050"},
	})

	for b.Loop() {
		m1.Merge(m2)
	}
}

// BenchmarkBacktrack_SmallNoConflict benchmarks backtracking with 3 courses, no conflicts.
func BenchmarkBacktrack_SmallNoConflict(b *testing.B) {
	groups := makeTestGroups(3, 2, false)
	ctx := context.Background()
	params := backtrackParams{groups: groups, minCourses: 1, maxCourses: 3, limit: 2000}

	b.ResetTimer()
	for b.Loop() {
		backtrack(ctx, params)
	}
}

// BenchmarkBacktrack_MediumNoConflict benchmarks backtracking with 6 courses, no conflicts.
func BenchmarkBacktrack_MediumNoConflict(b *testing.B) {
	groups := makeTestGroups(6, 3, false)
	ctx := context.Background()
	params := backtrackParams{groups: groups, minCourses: 1, maxCourses: 6, limit: 2000}

	b.ResetTimer()
	for b.Loop() {
		backtrack(ctx, params)
	}
}

// BenchmarkBacktrack_LargeNoConflict benchmarks backtracking with 8 courses, no conflicts.
func BenchmarkBacktrack_LargeNoConflict(b *testing.B) {
	groups := makeTestGroups(8, 4, false)
	ctx := context.Background()
	params := backtrackParams{groups: groups, minCourses: 1, maxCourses: 8, limit: 2000}

	b.ResetTimer()
	for b.Loop() {
		backtrack(ctx, params)
	}
}

// BenchmarkBacktrack_WithConflicts benchmarks with some conflicting sections.
func BenchmarkBacktrack_WithConflicts(b *testing.B) {
	groups := makeTestGroups(6, 4, true)
	ctx := context.Background()
	params := backtrackParams{groups: groups, minCourses: 1, maxCourses: 6, limit: 2000}

	b.ResetTimer()
	for b.Loop() {
		backtrack(ctx, params)
	}
}

// BenchmarkScoreSchedule measures scoring performance.
func BenchmarkScoreSchedule(b *testing.B) {
	schedule := &Schedule{
		Courses: []*cache.Course{
			{MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, true, false, true, false, true, false}, StartTime: "0900", EndTime: "0950"},
			}},
			{MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, true, false, true, false, true, false}, StartTime: "1000", EndTime: "1050"},
			}},
			{MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, false, true, false, true, false, false}, StartTime: "1100", EndTime: "1150"},
			}},
		},
	}

	for b.Loop() {
		scoreSchedule(schedule)
	}
}

// makeTestGroups creates test course groups for benchmarking.
// If withConflicts is true, some sections will have overlapping times.
func makeTestGroups(numCourses, sectionsPerCourse int, withConflicts bool) []courseGroup {
	groups := make([]courseGroup, numCourses)

	// Base times: spread across the day to avoid conflicts by default
	baseTimes := []string{"0800", "0900", "1000", "1100", "1200", "1300", "1400", "1500"}

	for i := range numCourses {
		groups[i].courseKey = "TEST:" + string(rune('A'+i))
		groups[i].sections = make([]*sectionData, sectionsPerCourse)

		for j := range sectionsPerCourse {
			var startTime string
			if withConflicts && j%2 == 1 {
				// Create some conflicts by using overlapping times
				startTime = baseTimes[i%len(baseTimes)]
			} else {
				// Use non-overlapping times based on course index
				startTime = baseTimes[(i+j)%len(baseTimes)]
			}

			// End time is 50 minutes after start
			endTime := addMinutes(startTime, 50)

			course := &cache.Course{
				ID:  int64(i*100 + j),
				CRN: string(rune('0' + i)) + string(rune('0' + j)),
				MeetingTimes: []cache.MeetingTime{
					{
						Days:      [7]bool{false, true, false, true, false, true, false}, // MWF
						StartTime: startTime,
						EndTime:   endTime,
					},
				},
			}

			groups[i].sections[j] = &sectionData{
				course: course,
				mask:   FromMeetingTimes(course.MeetingTimes),
			}
		}
	}

	return groups
}

// addMinutes adds minutes to a time string and returns the new time.
func addMinutes(t string, mins int) string {
	totalMins := parseTimeToMins(t) + mins
	hours := totalMins / 60
	minutes := totalMins % 60
	return string(rune('0'+hours/10)) + string(rune('0'+hours%10)) +
		string(rune('0'+minutes/10)) + string(rune('0'+minutes%10))
}
