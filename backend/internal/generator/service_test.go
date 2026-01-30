package generator

import (
	"testing"

	"schedule-optimizer/internal/cache"
)

func TestNormalizeCourseKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"CSCI 241", "CSCI:241"},
		{"csci 241", "CSCI:241"},
		{"CSCI:241", "CSCI:241"},
		{"csci:241", "CSCI:241"},
		{"CSCI241", "CSCI:241"},
		{"csci241", "CSCI:241"},
		{"  CSCI  241  ", "CSCI:241"},
		{"MATH 204", "MATH:204"},
		{"ENG 101", "ENG:101"},
		{"BIOL 101L", "BIOL:101L"}, // Lab suffix preserved
		{"A&S 101", "A&S:101"},    // Subject with special char
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeCourseKey(tt.input)
			if got != tt.want {
				t.Errorf("normalizeCourseKey(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSplitCourseKey(t *testing.T) {
	tests := []struct {
		input       string
		wantSubject string
		wantNumber  string
	}{
		{"CSCI:241", "CSCI", "241"},
		{"MATH:204", "MATH", "204"},
		{"INVALID", "INVALID", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			subject, number := splitCourseKey(tt.input)
			if subject != tt.wantSubject || number != tt.wantNumber {
				t.Errorf("splitCourseKey(%q) = (%q, %q), want (%q, %q)",
					tt.input, subject, number, tt.wantSubject, tt.wantNumber)
			}
		})
	}
}

func TestClampBounds(t *testing.T) {
	tests := []struct {
		name       string
		minReq     int
		maxReq     int
		numCourses int
		wantMin    int
		wantMax    int
	}{
		{"defaults", 0, 0, 5, 1, 5},                // Uses defaults, clamped to numCourses
		{"normal", 2, 4, 5, 2, 4},                  // Normal case
		{"min exceeds num", 6, 8, 5, 5, 5},         // Min clamped to numCourses
		{"max exceeds num", 1, 10, 5, 1, 5},        // Max clamped to numCourses
		{"min exceeds max", 4, 2, 5, 4, 4},         // Max raised to min
		{"zero min uses default", 0, 3, 5, 1, 3},   // MinReq=0 uses default
		{"negative uses default", -1, 0, 5, 1, 5},  // Negative values use defaults
		{"exact match", 3, 3, 3, 3, 3},             // All equal
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotMax := clampBounds(tt.minReq, tt.maxReq, tt.numCourses)
			if gotMin != tt.wantMin || gotMax != tt.wantMax {
				t.Errorf("clampBounds(%d, %d, %d) = (%d, %d), want (%d, %d)",
					tt.minReq, tt.maxReq, tt.numCourses, gotMin, gotMax, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestIsAsyncOrTBD(t *testing.T) {
	tests := []struct {
		name     string
		meetings []struct {
			days  [7]bool
			start string
			end   string
		}
		want bool
	}{
		{
			name:     "no meetings",
			meetings: nil,
			want:     true,
		},
		{
			name: "TBD times",
			meetings: []struct {
				days  [7]bool
				start string
				end   string
			}{
				{days: [7]bool{false, true, false, false, false, false, false}, start: "", end: ""},
			},
			want: true,
		},
		{
			name: "no weekdays",
			meetings: []struct {
				days  [7]bool
				start string
				end   string
			}{
				{days: [7]bool{true, false, false, false, false, false, true}, start: "0900", end: "1000"}, // Only Sun/Sat
			},
			want: true,
		},
		{
			name: "valid scheduled",
			meetings: []struct {
				days  [7]bool
				start string
				end   string
			}{
				{days: [7]bool{false, true, false, true, false, true, false}, start: "0900", end: "0950"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var course cache.Course
			for _, m := range tt.meetings {
				course.MeetingTimes = append(course.MeetingTimes, cache.MeetingTime{
					Days:      m.days,
					StartTime: m.start,
					EndTime:   m.end,
				})
			}

			got := isAsyncOrTBD(&course)
			if got != tt.want {
				t.Errorf("isAsyncOrTBD() = %v, want %v", got, tt.want)
			}
		})
	}
}
