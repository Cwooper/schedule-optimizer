package generator

import (
	"testing"

	"schedule-optimizer/internal/cache"
)

func TestToResponse_BasicTransformation(t *testing.T) {
	input := &GenerateResponse{
		Schedules: []Schedule{
			{
				Courses: []*cache.Course{
					{
						CRN:          "12345",
						Term:         "202510",
						Subject:      "CSCI",
						CourseNumber: "241",
						Title:        "Data Structures",
						Credits:      4,
						Instructor:   "Smith",
						Enrollment:   25,
						MaxEnrollment: 30,
						SeatsAvailable: 5,
						MeetingTimes: []cache.MeetingTime{
							{Days: [7]bool{false, true, false, true, false, false, false}, StartTime: "0900", EndTime: "0950"},
						},
					},
				},
				Score:   0.85,
				Weights: []Weight{{Name: "gaps", Value: 0.5}},
			},
		},
		CourseResults: []CourseResult{{Name: "CSCI 241", Status: StatusFound, Count: 3}},
		Stats:         GenerateStats{TotalGenerated: 1, TimeMs: 10.5},
	}

	resp := input.ToResponse()

	// Check courses map
	if len(resp.Courses) != 1 {
		t.Errorf("expected 1 course, got %d", len(resp.Courses))
	}
	course, ok := resp.Courses["CSCI:241"]
	if !ok {
		t.Fatal("expected course CSCI:241 in map")
	}
	if course.Title != "Data Structures" {
		t.Errorf("expected title 'Data Structures', got %q", course.Title)
	}

	// Check sections map
	if len(resp.Sections) != 1 {
		t.Errorf("expected 1 section, got %d", len(resp.Sections))
	}
	section, ok := resp.Sections["12345"]
	if !ok {
		t.Fatal("expected section 12345 in map")
	}
	if section.CourseKey != "CSCI:241" {
		t.Errorf("expected courseKey 'CSCI:241', got %q", section.CourseKey)
	}
	if section.Instructor != "Smith" {
		t.Errorf("expected instructor 'Smith', got %q", section.Instructor)
	}

	// Check schedules
	if len(resp.Schedules) != 1 {
		t.Errorf("expected 1 schedule, got %d", len(resp.Schedules))
	}
	if len(resp.Schedules[0].CRNs) != 1 || resp.Schedules[0].CRNs[0] != "12345" {
		t.Errorf("expected schedule CRNs [12345], got %v", resp.Schedules[0].CRNs)
	}
	if resp.Schedules[0].Score != 0.85 {
		t.Errorf("expected score 0.85, got %f", resp.Schedules[0].Score)
	}
}

func TestToResponse_Deduplication(t *testing.T) {
	// Same course appears in 3 different schedules
	course := &cache.Course{
		CRN:          "12345",
		Term:         "202510",
		Subject:      "CSCI",
		CourseNumber: "241",
		Title:        "Data Structures",
		Credits:      4,
	}

	input := &GenerateResponse{
		Schedules: []Schedule{
			{Courses: []*cache.Course{course}, Score: 0.9},
			{Courses: []*cache.Course{course}, Score: 0.8},
			{Courses: []*cache.Course{course}, Score: 0.7},
		},
		CourseResults: []CourseResult{},
		Stats:         GenerateStats{TotalGenerated: 3, TimeMs: 5.0},
	}

	resp := input.ToResponse()

	// Should have exactly 1 course and 1 section despite 3 schedules
	if len(resp.Courses) != 1 {
		t.Errorf("expected 1 course after dedup, got %d", len(resp.Courses))
	}
	if len(resp.Sections) != 1 {
		t.Errorf("expected 1 section after dedup, got %d", len(resp.Sections))
	}
	if len(resp.Schedules) != 3 {
		t.Errorf("expected 3 schedules, got %d", len(resp.Schedules))
	}
}

func TestToResponse_MultipleSections(t *testing.T) {
	// Two different sections of the same course in one schedule
	input := &GenerateResponse{
		Schedules: []Schedule{
			{
				Courses: []*cache.Course{
					{CRN: "11111", Term: "202510", Subject: "CSCI", CourseNumber: "241", Title: "Data Structures", Credits: 4},
					{CRN: "22222", Term: "202510", Subject: "MATH", CourseNumber: "204", Title: "Linear Algebra", Credits: 5},
				},
				Score: 0.9,
			},
		},
		CourseResults: []CourseResult{},
		Stats:         GenerateStats{TotalGenerated: 1, TimeMs: 5.0},
	}

	resp := input.ToResponse()

	if len(resp.Courses) != 2 {
		t.Errorf("expected 2 courses, got %d", len(resp.Courses))
	}
	if len(resp.Sections) != 2 {
		t.Errorf("expected 2 sections, got %d", len(resp.Sections))
	}
	if len(resp.Schedules[0].CRNs) != 2 {
		t.Errorf("expected 2 CRNs in schedule, got %d", len(resp.Schedules[0].CRNs))
	}
}

func TestToResponse_AsyncSections(t *testing.T) {
	input := &GenerateResponse{
		Schedules: []Schedule{},
		Asyncs: []*cache.Course{
			{CRN: "99999", Term: "202510", Subject: "CSCI", CourseNumber: "101", Title: "Intro to CS", Credits: 4},
		},
		CourseResults: []CourseResult{},
		Stats:         GenerateStats{TotalGenerated: 0, TimeMs: 1.0},
	}

	resp := input.ToResponse()

	// Async sections should be in sections map
	if len(resp.Sections) != 1 {
		t.Errorf("expected 1 section from asyncs, got %d", len(resp.Sections))
	}
	if _, ok := resp.Sections["99999"]; !ok {
		t.Error("expected async section 99999 in sections map")
	}

	// Async CRNs should be in asyncs array
	if len(resp.Asyncs) != 1 || resp.Asyncs[0] != "99999" {
		t.Errorf("expected asyncs [99999], got %v", resp.Asyncs)
	}

	// Course should be in courses map
	if len(resp.Courses) != 1 {
		t.Errorf("expected 1 course from asyncs, got %d", len(resp.Courses))
	}
}

func TestToResponse_EmptyResponse(t *testing.T) {
	input := &GenerateResponse{
		Schedules:     []Schedule{},
		Asyncs:        []*cache.Course{},
		CourseResults: []CourseResult{},
		Stats:         GenerateStats{TotalGenerated: 0, TimeMs: 0.5},
	}

	resp := input.ToResponse()

	// Empty maps, not nil
	if resp.Courses == nil {
		t.Error("courses should be empty map, not nil")
	}
	if len(resp.Courses) != 0 {
		t.Errorf("expected 0 courses, got %d", len(resp.Courses))
	}

	if resp.Sections == nil {
		t.Error("sections should be empty map, not nil")
	}
	if len(resp.Sections) != 0 {
		t.Errorf("expected 0 sections, got %d", len(resp.Sections))
	}

	if resp.Schedules == nil {
		t.Error("schedules should be empty slice, not nil")
	}
	if len(resp.Schedules) != 0 {
		t.Errorf("expected 0 schedules, got %d", len(resp.Schedules))
	}

	if resp.Asyncs == nil {
		t.Error("asyncs should be empty slice, not nil")
	}
}

func TestToResponse_NilAsyncs(t *testing.T) {
	input := &GenerateResponse{
		Schedules:     []Schedule{},
		Asyncs:        nil, // nil instead of empty slice
		CourseResults: []CourseResult{},
		Stats:         GenerateStats{TotalGenerated: 0, TimeMs: 0.5},
	}

	resp := input.ToResponse()

	// Should still get empty slice, not nil
	if resp.Asyncs == nil {
		t.Error("asyncs should be empty slice even when input is nil")
	}
}

func BenchmarkToResponse_2000Schedules(b *testing.B) {
	// Create 5 courses
	courses := make([]*cache.Course, 5)
	for i := range 5 {
		courses[i] = &cache.Course{
			CRN:           string(rune('1' + i)) + "0000",
			Term:          "202510",
			Subject:       "CSCI",
			CourseNumber:  string(rune('1' + i)) + "00",
			Title:         "Course " + string(rune('A'+i)),
			Credits:       4,
			Instructor:    "Instructor",
			Enrollment:    25,
			MaxEnrollment: 30,
			SeatsAvailable: 5,
			MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, true, false, true, false, false, false}, StartTime: "0900", EndTime: "0950"},
			},
		}
	}

	// Create 2000 schedules
	schedules := make([]Schedule, 2000)
	for i := range 2000 {
		schedules[i] = Schedule{
			Courses: courses,
			Score:   float64(i) / 2000.0,
			Weights: []Weight{{Name: "gaps", Value: 0.5}, {Name: "start", Value: 0.3}},
		}
	}

	input := &GenerateResponse{
		Schedules:     schedules,
		CourseResults: []CourseResult{{Name: "Test", Status: StatusFound}},
		Stats:         GenerateStats{TotalGenerated: 2000, TimeMs: 100.0},
	}

	b.ResetTimer()
	for b.Loop() {
		_ = input.ToResponse()
	}
}
