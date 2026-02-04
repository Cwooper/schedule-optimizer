package generator

import (
	"context"
	"slices"
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
		{"A&S 101", "A&S:101"},     // Subject with special char
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
		{"defaults", 0, 0, 5, 1, 5},                           // Uses defaults, clamped to numCourses
		{"normal", 2, 4, 5, 2, 4},                             // Normal case
		{"min exceeds num", 6, 8, 5, 5, 5},                    // Min clamped to numCourses
		{"max exceeds num", 1, 10, 5, 1, 5},                   // Max clamped to numCourses
		{"min exceeds max", 4, 2, 5, 4, 4},                    // Max raised to min
		{"zero min uses default", 0, 3, 5, 1, 3},              // MinReq=0 uses default
		{"negative uses default", -1, 0, 5, 1, 5},             // Negative values use defaults
		{"exact match", 3, 3, 3, 3, 3},                        // All equal
		{"required courses set min", 3, 5, 5, 3, 5},           // Simulates 3 required courses (min already set by caller)
		{"required courses with low user min", 2, 4, 5, 2, 4}, // User min=2 with required already factored in by caller
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

// simulateFallback replicates the min-course fallback logic from Generate()
// to test it in isolation without needing a cache or database.
func simulateFallback(
	groups []courseGroup,
	numRequired int,
	reqMinCourses int, // 0 = user didn't set
	reqMaxCourses int,
) []Schedule {
	totalCourses := len(groups)

	effectiveMin := reqMinCourses
	userSetMin := effectiveMin > 0
	if effectiveMin == 0 {
		effectiveMin = totalCourses
	}
	effectiveMin = max(effectiveMin, numRequired)

	fallbackMin := effectiveMin
	if !userSetMin && effectiveMin > numRequired && effectiveMin > 1 {
		fallbackMin = effectiveMin - 1
	}

	minCourses, maxCourses := clampBounds(fallbackMin, reqMaxCourses, totalCourses)

	ctx := context.Background()
	schedules := backtrack(ctx, backtrackParams{
		groups:      groups,
		numRequired: numRequired,
		minCourses:  minCourses,
		maxCourses:  maxCourses,
		limit:       MaxSchedulesToGenerate,
	})

	if !userSetMin && fallbackMin < effectiveMin && len(schedules) > 0 {
		fullCount := 0
		for _, s := range schedules {
			if len(s.Courses) >= effectiveMin {
				fullCount++
			}
		}
		if fullCount > 0 {
			schedules = slices.DeleteFunc(schedules, func(s Schedule) bool {
				return len(s.Courses) < effectiveMin
			})
		}
	}

	return schedules
}

func TestFallback_NoFullSchedules_KeepsFallback(t *testing.T) {
	// 3 courses that all conflict with each other — no 3-course combo works,
	// but each pair has one non-conflicting combo available via alternate sections.

	// Course A: Mon 9-10
	sectionA := makeTestSection(1, "A1", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
	})
	// Course B: Mon 9-10 (conflicts with A) + alternate Tue 9-10
	sectionB1 := makeTestSection(2, "B1", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
	})
	sectionB2 := makeTestSection(3, "B2", []cache.MeetingTime{
		{Days: [7]bool{false, false, true, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
	})
	// Course C: Mon 9-10 (conflicts with A) + Tue 9-10 (conflicts with B2)
	sectionC1 := makeTestSection(4, "C1", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
	})
	sectionC2 := makeTestSection(5, "C2", []cache.MeetingTime{
		{Days: [7]bool{false, false, true, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
	})

	groups := []courseGroup{
		{courseKey: "A:100", sections: []*sectionData{
			{course: sectionA, mask: FromMeetingTimes(sectionA.MeetingTimes)},
		}},
		{courseKey: "B:200", sections: []*sectionData{
			{course: sectionB1, mask: FromMeetingTimes(sectionB1.MeetingTimes)},
			{course: sectionB2, mask: FromMeetingTimes(sectionB2.MeetingTimes)},
		}},
		{courseKey: "C:300", sections: []*sectionData{
			{course: sectionC1, mask: FromMeetingTimes(sectionC1.MeetingTimes)},
			{course: sectionC2, mask: FromMeetingTimes(sectionC2.MeetingTimes)},
		}},
	}

	// With MinCourses=0 (user didn't set), fallback allows 2-course schedules.
	// No 3-course combo works (A is Mon 9-10; B2 is Tue 9-10; C2 is also Tue 9-10 — B2+C2 conflict).
	// So fallback 2-course schedules should be returned.
	schedules := simulateFallback(groups, 0, 0, 0)
	if len(schedules) == 0 {
		t.Fatal("Expected fallback to produce schedules when no full-count combo works")
	}
	for _, s := range schedules {
		if len(s.Courses) < 2 {
			t.Errorf("Fallback schedules should have at least 2 courses, got %d", len(s.Courses))
		}
	}

	// With MinCourses=3 (user explicitly set), no fallback — should get 0 results.
	schedules = simulateFallback(groups, 0, 3, 0)
	if len(schedules) != 0 {
		t.Errorf("User-set min=3 should produce 0 schedules when no 3-course combo works, got %d", len(schedules))
	}
}

func TestFallback_FullSchedulesExist_FiltersFallback(t *testing.T) {
	// 3 non-conflicting courses — full 3-course combos exist.
	// Fallback allows 2-course schedules, but they should be filtered out
	// because full-count schedules are available.

	sectionA := makeTestSection(1, "A1", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0800", EndTime: "0900"},
	})
	sectionB := makeTestSection(2, "B1", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
	})
	sectionC := makeTestSection(3, "C1", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "1000", EndTime: "1100"},
	})

	groups := []courseGroup{
		{courseKey: "A:100", sections: []*sectionData{{course: sectionA, mask: FromMeetingTimes(sectionA.MeetingTimes)}}},
		{courseKey: "B:200", sections: []*sectionData{{course: sectionB, mask: FromMeetingTimes(sectionB.MeetingTimes)}}},
		{courseKey: "C:300", sections: []*sectionData{{course: sectionC, mask: FromMeetingTimes(sectionC.MeetingTimes)}}},
	}

	schedules := simulateFallback(groups, 0, 0, 0)

	// All returned schedules should have exactly 3 courses (2-course fallbacks filtered out)
	if len(schedules) == 0 {
		t.Fatal("Expected schedules when all courses are compatible")
	}
	for _, s := range schedules {
		if len(s.Courses) != 3 {
			t.Errorf("Expected only 3-course schedules (fallback filtered), got schedule with %d courses", len(s.Courses))
		}
	}
}

func TestFallback_UserSetMin_NoFallback(t *testing.T) {
	// Same 3 non-conflicting courses but user explicitly sets min=2.
	// Should include 2-course and 3-course schedules (no fallback filtering).
	sectionA := makeTestSection(1, "A1", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0800", EndTime: "0900"},
	})
	sectionB := makeTestSection(2, "B1", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
	})
	sectionC := makeTestSection(3, "C1", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "1000", EndTime: "1100"},
	})

	groups := []courseGroup{
		{courseKey: "A:100", sections: []*sectionData{{course: sectionA, mask: FromMeetingTimes(sectionA.MeetingTimes)}}},
		{courseKey: "B:200", sections: []*sectionData{{course: sectionB, mask: FromMeetingTimes(sectionB.MeetingTimes)}}},
		{courseKey: "C:300", sections: []*sectionData{{course: sectionC, mask: FromMeetingTimes(sectionC.MeetingTimes)}}},
	}

	schedules := simulateFallback(groups, 0, 2, 0)

	// User set min=2, so both 2- and 3-course schedules should appear
	has2 := false
	has3 := false
	for _, s := range schedules {
		if len(s.Courses) == 2 {
			has2 = true
		}
		if len(s.Courses) == 3 {
			has3 = true
		}
	}
	if !has2 || !has3 {
		t.Errorf("User-set min=2 should produce both 2-course and 3-course schedules, has2=%v has3=%v", has2, has3)
	}
}

func TestFallback_RequiredCourses_FloorRespected(t *testing.T) {
	// 2 required + 1 optional, all non-conflicting.
	// effectiveMin = max(totalCourses=3, numRequired=2) = 3
	// fallbackMin should be 2, but since numRequired=2, the guard
	// (effectiveMin > numRequired) allows it: 3 > 2 → fallbackMin = 2.
	// If all 3 fit, fallback schedules should be filtered out.

	sectionA := makeTestSection(1, "A1", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0800", EndTime: "0900"},
	})
	sectionB := makeTestSection(2, "B1", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
	})
	sectionC := makeTestSection(3, "C1", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "1000", EndTime: "1100"},
	})

	groups := []courseGroup{
		// First 2 are required
		{courseKey: "REQ:100", sections: []*sectionData{{course: sectionA, mask: FromMeetingTimes(sectionA.MeetingTimes)}}},
		{courseKey: "REQ:200", sections: []*sectionData{{course: sectionB, mask: FromMeetingTimes(sectionB.MeetingTimes)}}},
		{courseKey: "OPT:300", sections: []*sectionData{{course: sectionC, mask: FromMeetingTimes(sectionC.MeetingTimes)}}},
	}

	schedules := simulateFallback(groups, 2, 0, 0)

	// All 3 fit, so only 3-course schedules should remain
	for _, s := range schedules {
		if len(s.Courses) != 3 {
			t.Errorf("Expected only 3-course schedules when all fit, got %d", len(s.Courses))
		}
	}

	// Both required courses must be in every schedule
	for _, s := range schedules {
		crns := make(map[string]bool)
		for _, c := range s.Courses {
			crns[c.CRN] = true
		}
		if !crns["A1"] || !crns["B1"] {
			t.Error("Required courses missing from schedule")
		}
	}
}

func TestFallback_SingleCourse_NoFallback(t *testing.T) {
	// With only 1 course, effectiveMin=1. Guard (effectiveMin > 1) prevents
	// fallback to 0, which would be nonsensical.
	section := makeTestSection(1, "A1", []cache.MeetingTime{
		{Days: [7]bool{false, true, false, false, false, false, false}, StartTime: "0900", EndTime: "1000"},
	})

	groups := []courseGroup{
		{courseKey: "A:100", sections: []*sectionData{{course: section, mask: FromMeetingTimes(section.MeetingTimes)}}},
	}

	schedules := simulateFallback(groups, 0, 0, 0)
	if len(schedules) != 1 {
		t.Errorf("Expected 1 schedule for single course, got %d", len(schedules))
	}
	if len(schedules[0].Courses) != 1 {
		t.Errorf("Expected 1 course in schedule, got %d", len(schedules[0].Courses))
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
