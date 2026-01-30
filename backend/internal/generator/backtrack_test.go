package generator

import (
	"context"
	"testing"

	"schedule-optimizer/internal/cache"
)

func makeTestSection(id int64, crn string, meetings []cache.MeetingTime) *cache.Course {
	return &cache.Course{
		ID:           id,
		CRN:          crn,
		MeetingTimes: meetings,
	}
}

func TestBacktrack_NoGroups(t *testing.T) {
	ctx := context.Background()
	schedules := backtrack(ctx, backtrackParams{
		groups:     nil,
		minCourses: 1,
		maxCourses: 3,
		limit:      100,
	})
	if len(schedules) != 0 {
		t.Errorf("Expected 0 schedules, got %d", len(schedules))
	}
}

func TestBacktrack_SingleGroupSingleSection(t *testing.T) {
	ctx := context.Background()
	section := makeTestSection(1, "12345", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "0950"},
	})
	mask := FromMeetingTimes(section.MeetingTimes)

	groups := []courseGroup{
		{courseKey: "CSCI:241", sections: []*sectionData{{course: section, mask: mask}}},
	}

	schedules := backtrack(ctx, backtrackParams{
		groups:     groups,
		minCourses: 1,
		maxCourses: 1,
		limit:      100,
	})
	if len(schedules) != 1 {
		t.Errorf("Expected 1 schedule, got %d", len(schedules))
	}
	if len(schedules[0].Courses) != 1 {
		t.Errorf("Expected 1 course in schedule, got %d", len(schedules[0].Courses))
	}
}

func TestBacktrack_TwoGroupsNoConflict(t *testing.T) {
	ctx := context.Background()

	// CSCI 241 on Monday 9-10am
	section1 := makeTestSection(1, "12345", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
	})
	// MATH 204 on Monday 10-11am (no conflict)
	section2 := makeTestSection(2, "12346", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "1000", EndTime: "1100"},
	})

	groups := []courseGroup{
		{courseKey: "CSCI:241", sections: []*sectionData{{course: section1, mask: FromMeetingTimes(section1.MeetingTimes)}}},
		{courseKey: "MATH:204", sections: []*sectionData{{course: section2, mask: FromMeetingTimes(section2.MeetingTimes)}}},
	}

	schedules := backtrack(ctx, backtrackParams{
		groups:     groups,
		minCourses: 1,
		maxCourses: 2,
		limit:      100,
	})
	// Should produce: [CSCI], [MATH], [CSCI+MATH] = 3 schedules
	if len(schedules) != 3 {
		t.Errorf("Expected 3 schedules, got %d", len(schedules))
	}
}

func TestBacktrack_TwoGroupsWithConflict(t *testing.T) {
	ctx := context.Background()

	// CSCI 241 on Monday 9-10am
	section1 := makeTestSection(1, "12345", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
	})
	// MATH 204 on Monday 9:30-10:30am (conflicts!)
	section2 := makeTestSection(2, "12346", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0930", EndTime: "1030"},
	})

	groups := []courseGroup{
		{courseKey: "CSCI:241", sections: []*sectionData{{course: section1, mask: FromMeetingTimes(section1.MeetingTimes)}}},
		{courseKey: "MATH:204", sections: []*sectionData{{course: section2, mask: FromMeetingTimes(section2.MeetingTimes)}}},
	}

	schedules := backtrack(ctx, backtrackParams{
		groups:     groups,
		minCourses: 1,
		maxCourses: 2,
		limit:      100,
	})
	// Should produce only single-course schedules: [CSCI], [MATH] = 2 schedules
	// No combined schedule due to conflict
	if len(schedules) != 2 {
		t.Errorf("Expected 2 schedules, got %d", len(schedules))
	}
	for _, s := range schedules {
		if len(s.Courses) > 1 {
			t.Error("Should not have any combined schedules due to conflict")
		}
	}
}

func TestBacktrack_MinCoursesFiltering(t *testing.T) {
	ctx := context.Background()

	section1 := makeTestSection(1, "12345", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
	})
	section2 := makeTestSection(2, "12346", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "1000", EndTime: "1100"},
	})

	groups := []courseGroup{
		{courseKey: "CSCI:241", sections: []*sectionData{{course: section1, mask: FromMeetingTimes(section1.MeetingTimes)}}},
		{courseKey: "MATH:204", sections: []*sectionData{{course: section2, mask: FromMeetingTimes(section2.MeetingTimes)}}},
	}

	// Require minimum 2 courses
	schedules := backtrack(ctx, backtrackParams{
		groups:     groups,
		minCourses: 2,
		maxCourses: 2,
		limit:      100,
	})
	if len(schedules) != 1 {
		t.Errorf("Expected 1 schedule with min=2, got %d", len(schedules))
	}
	if len(schedules[0].Courses) != 2 {
		t.Errorf("Expected 2 courses in schedule, got %d", len(schedules[0].Courses))
	}
}

func TestBacktrack_Limit(t *testing.T) {
	ctx := context.Background()

	// Create 3 sections per course = 3^2 = 9 possible combinations (plus partial schedules)
	var sections1 []*sectionData
	for i := range 3 {
		s := makeTestSection(int64(i), "", []cache.MeetingTime{
			{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "0950"},
		})
		sections1 = append(sections1, &sectionData{course: s, mask: FromMeetingTimes(s.MeetingTimes)})
	}

	var sections2 []*sectionData
	for i := range 3 {
		s := makeTestSection(int64(10+i), "", []cache.MeetingTime{
			{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "1000", EndTime: "1050"},
		})
		sections2 = append(sections2, &sectionData{course: s, mask: FromMeetingTimes(s.MeetingTimes)})
	}

	groups := []courseGroup{
		{courseKey: "CSCI:241", sections: sections1},
		{courseKey: "MATH:204", sections: sections2},
	}

	// Limit to 5 schedules
	schedules := backtrack(ctx, backtrackParams{
		groups:     groups,
		minCourses: 1,
		maxCourses: 2,
		limit:      5,
	})
	if len(schedules) > 5 {
		t.Errorf("Expected at most 5 schedules, got %d", len(schedules))
	}
}

func TestBacktrack_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	section := makeTestSection(1, "12345", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "0950"},
	})

	groups := []courseGroup{
		{courseKey: "CSCI:241", sections: []*sectionData{{course: section, mask: FromMeetingTimes(section.MeetingTimes)}}},
	}

	schedules := backtrack(ctx, backtrackParams{
		groups:     groups,
		minCourses: 1,
		maxCourses: 1,
		limit:      100,
	})
	if len(schedules) != 0 {
		t.Errorf("Expected 0 schedules when context cancelled, got %d", len(schedules))
	}
}

func TestBacktrack_MultipleSectionsPerCourse(t *testing.T) {
	ctx := context.Background()

	// Course with two sections at different times
	section1a := makeTestSection(1, "12345", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "0950"},
	})
	section1b := makeTestSection(2, "12346", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "1000", EndTime: "1050"},
	})

	groups := []courseGroup{
		{courseKey: "CSCI:241", sections: []*sectionData{
			{course: section1a, mask: FromMeetingTimes(section1a.MeetingTimes)},
			{course: section1b, mask: FromMeetingTimes(section1b.MeetingTimes)},
		}},
	}

	schedules := backtrack(ctx, backtrackParams{
		groups:     groups,
		minCourses: 1,
		maxCourses: 1,
		limit:      100,
	})
	// Should produce 2 schedules (one for each section)
	if len(schedules) != 2 {
		t.Errorf("Expected 2 schedules, got %d", len(schedules))
	}
}

func TestBacktrack_WithForcedSections(t *testing.T) {
	ctx := context.Background()

	// Forced section on Monday 8-9am
	forced := makeTestSection(1, "11111", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0800", EndTime: "0900"},
	})
	forcedMask := FromMeetingTimes(forced.MeetingTimes)

	// Optional section on Monday 9-10am (no conflict with forced)
	optional := makeTestSection(2, "22222", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
	})

	groups := []courseGroup{
		{courseKey: "CSCI:241", sections: []*sectionData{{course: optional, mask: FromMeetingTimes(optional.MeetingTimes)}}},
	}

	schedules := backtrack(ctx, backtrackParams{
		groups:     groups,
		forced:     []*sectionData{{course: forced, mask: forcedMask}},
		forcedMask: forcedMask,
		minCourses: 1,
		maxCourses: 2,
		limit:      100,
	})

	// Should produce schedules that all include the forced section
	if len(schedules) != 2 { // [forced], [forced+optional]
		t.Errorf("Expected 2 schedules, got %d", len(schedules))
	}

	for _, s := range schedules {
		found := false
		for _, c := range s.Courses {
			if c.CRN == "11111" {
				found = true
				break
			}
		}
		if !found {
			t.Error("All schedules should include forced section")
		}
	}
}
