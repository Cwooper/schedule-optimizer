package generator

import (
	"math"
	"testing"

	"schedule-optimizer/internal/cache"
)

func makeScheduleWithMeetings(meetings []cache.MeetingTime) *Schedule {
	return &Schedule{
		Courses: []*cache.Course{
			{MeetingTimes: meetings},
		},
	}
}

func TestWeighGap_SingleClass(t *testing.T) {
	// Single class - no gaps possible, should be 1.0
	s := makeScheduleWithMeetings([]cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
	})

	score := weighGap(s)
	if score != 1.0 {
		t.Errorf("Single class should have gap score 1.0, got %v", score)
	}
}

func TestWeighGap_NoGap(t *testing.T) {
	// Back-to-back classes on same day - no gap
	s := &Schedule{
		Courses: []*cache.Course{
			{MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
			}},
			{MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "1000", EndTime: "1100"},
			}},
		},
	}

	score := weighGap(s)
	if score != 1.0 {
		t.Errorf("Back-to-back classes should have gap score 1.0, got %v", score)
	}
}

func TestWeighGap_WithGap(t *testing.T) {
	// Two 1-hour classes with 1-hour gap
	// Span = 3 hours, class time = 2 hours, gap = 1 hour
	// Gap ratio = 1/3 = 0.33, score = 1 - 0.33 = 0.67
	s := &Schedule{
		Courses: []*cache.Course{
			{MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
			}},
			{MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "1100", EndTime: "1200"},
			}},
		},
	}

	score := weighGap(s)
	expected := 0.67 // rounded
	if math.Abs(score-expected) > 0.01 {
		t.Errorf("Expected gap score ~%v, got %v", expected, score)
	}
}

func TestWeighGap_MultipleDays(t *testing.T) {
	// Monday: 9-10 (no gap, score 1.0)
	// Tuesday: 9-10 and 11-12 (1hr gap in 3hr span, score ~0.67)
	// Average: (1.0 + 0.67) / 2 = 0.835
	s := &Schedule{
		Courses: []*cache.Course{
			{MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
			}},
			{MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, false, true, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
			}},
			{MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, false, true, false, false, false, false}, StartTime: "1100", EndTime: "1200"},
			}},
		},
	}

	score := weighGap(s)
	expected := 0.83 // (1.0 + 0.67) / 2, rounded
	if math.Abs(score-expected) > 0.02 {
		t.Errorf("Expected gap score ~%v, got %v", expected, score)
	}
}

func TestWeighStart(t *testing.T) {
	tests := []struct {
		name      string
		startTime string
		wantMin   float64
		wantMax   float64
	}{
		{"8am start", "0800", 0.0, 0.01},
		{"5pm start", "1700", 0.99, 1.0},
		{"Noon start", "1200", 0.43, 0.46}, // ~4 hours / 9 hours
		{"Before 8am", "0700", 0.0, 0.01},  // Clamped to 8am
		{"After 5pm", "1800", 0.99, 1.0},   // Returns 1.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := makeScheduleWithMeetings([]cache.MeetingTime{
				{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: tt.startTime, EndTime: "1800"},
			})
			score := weighStart(s)
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("weighStart with %s: got %v, want between %v and %v", tt.startTime, score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestWeighEnd(t *testing.T) {
	tests := []struct {
		name    string
		endTime string
		wantMin float64
		wantMax float64
	}{
		{"5pm end", "1700", 0.0, 0.01},
		{"8am end", "0800", 0.99, 1.0},
		{"Noon end", "1200", 0.54, 0.57}, // ~5 hours / 9 hours
		{"After 5pm", "1800", 0.0, 0.01}, // Clamped to 5pm
		{"Before 8am", "0700", 0.99, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := makeScheduleWithMeetings([]cache.MeetingTime{
				{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0700", EndTime: tt.endTime},
			})
			score := weighEnd(s)
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("weighEnd with %s: got %v, want between %v and %v", tt.endTime, score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestScoreSchedule(t *testing.T) {
	s := &Schedule{
		Courses: []*cache.Course{
			{MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
			}},
		},
	}

	scoreSchedule(s)

	if len(s.Weights) != 4 {
		t.Errorf("Expected 4 weights, got %d", len(s.Weights))
	}

	// Check weight names
	expectedNames := map[string]bool{"GPA": true, "Gap": true, "Start": true, "End": true}
	for _, w := range s.Weights {
		if !expectedNames[w.Name] {
			t.Errorf("Unexpected weight name: %s", w.Name)
		}
	}

	// Score should be average of non-zero weights
	if s.Score <= 0 || s.Score > 1 {
		t.Errorf("Score should be between 0 and 1, got %v", s.Score)
	}
}

func TestFindEarliestStart(t *testing.T) {
	s := &Schedule{
		Courses: []*cache.Course{
			{MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "1000", EndTime: "1100"},
			}},
			{MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "0950"},
			}},
		},
	}

	earliest := findEarliestStart(s)
	expected := 9 * 60 // 9:00 AM in minutes
	if earliest != expected {
		t.Errorf("Expected earliest start %d, got %d", expected, earliest)
	}
}

func TestFindLatestEnd(t *testing.T) {
	s := &Schedule{
		Courses: []*cache.Course{
			{MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
			}},
			{MeetingTimes: []cache.MeetingTime{
				{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "1000", EndTime: "1100"},
			}},
		},
	}

	latest := findLatestEnd(s)
	expected := 11 * 60 // 11:00 AM in minutes
	if latest != expected {
		t.Errorf("Expected latest end %d, got %d", expected, latest)
	}
}
