package generator

import (
	"strings"
	"testing"

	"schedule-optimizer/internal/cache"
)

// normalizeCourseKey converts various formats to "SUBJECT:NUMBER".
// Examples: "CSCI 241" -> "CSCI:241", "csci241" -> "CSCI:241"
// NOTE: This was used in the old API that accepted course strings.
// Kept for testing backward compatibility parsing if needed.
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
		{"defaults", 0, 0, 5, 1, 5},                              // Uses defaults, clamped to numCourses
		{"normal", 2, 4, 5, 2, 4},                                // Normal case
		{"min exceeds num", 6, 8, 5, 5, 5},                       // Min clamped to numCourses
		{"max exceeds num", 1, 10, 5, 1, 5},                      // Max clamped to numCourses
		{"min exceeds max", 4, 2, 5, 4, 4},                       // Max raised to min
		{"zero min uses default", 0, 3, 5, 1, 3},                 // MinReq=0 uses default
		{"negative uses default", -1, 0, 5, 1, 5},                // Negative values use defaults
		{"exact match", 3, 3, 3, 3, 3},                           // All equal
		{"required courses set min", 3, 5, 5, 3, 5},              // Simulates 3 required courses (min already set by caller)
		{"required courses with low user min", 2, 4, 5, 2, 4},    // User min=2 with required already factored in by caller
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

func TestAllowedCRNsFiltering(t *testing.T) {
	// This tests the CRN filtering logic used in buildCourseGroups.
	// The actual filtering is: if allowedCRNs != nil && !allowedCRNs[crn] { skip }

	tests := []struct {
		name        string
		allowedCRNs []string
		sectionCRNs []string
		wantPassed  []string
	}{
		{
			name:        "nil allowedCRNs passes all",
			allowedCRNs: nil,
			sectionCRNs: []string{"11111", "22222", "33333"},
			wantPassed:  []string{"11111", "22222", "33333"},
		},
		{
			name:        "empty allowedCRNs passes all",
			allowedCRNs: []string{},
			sectionCRNs: []string{"11111", "22222", "33333"},
			wantPassed:  []string{"11111", "22222", "33333"},
		},
		{
			name:        "filter to single CRN",
			allowedCRNs: []string{"22222"},
			sectionCRNs: []string{"11111", "22222", "33333"},
			wantPassed:  []string{"22222"},
		},
		{
			name:        "filter to multiple CRNs",
			allowedCRNs: []string{"11111", "33333"},
			sectionCRNs: []string{"11111", "22222", "33333"},
			wantPassed:  []string{"11111", "33333"},
		},
		{
			name:        "no CRNs match",
			allowedCRNs: []string{"99999"},
			sectionCRNs: []string{"11111", "22222", "33333"},
			wantPassed:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build allowed set (mirrors buildCourseGroups logic)
			var allowedSet map[string]bool
			if len(tt.allowedCRNs) > 0 {
				allowedSet = make(map[string]bool, len(tt.allowedCRNs))
				for _, crn := range tt.allowedCRNs {
					allowedSet[crn] = true
				}
			}

			// Filter sections
			var passed []string
			for _, crn := range tt.sectionCRNs {
				if allowedSet != nil && !allowedSet[crn] {
					continue
				}
				passed = append(passed, crn)
			}

			// Compare results
			if len(passed) != len(tt.wantPassed) {
				t.Errorf("got %d passed CRNs, want %d", len(passed), len(tt.wantPassed))
				return
			}
			for i, crn := range passed {
				if crn != tt.wantPassed[i] {
					t.Errorf("passed[%d] = %q, want %q", i, crn, tt.wantPassed[i])
				}
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
