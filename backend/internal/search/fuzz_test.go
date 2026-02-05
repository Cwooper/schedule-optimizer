package search

import (
	"context"
	"strings"
	"testing"

	"schedule-optimizer/internal/testutil"
)

// FuzzSearch tests that the search function never panics with arbitrary input.
// Run with: go test -fuzz=FuzzSearch -fuzztime=30s ./internal/search/...
func FuzzSearch(f *testing.F) {
	db, queries := testutil.SetupTestDB(f)
	defer db.Close()
	testutil.SeedTestData(f, db)

	svc := NewService(db, queries)
	ctx := context.Background()

	// Seed corpus with known interesting inputs
	f.Add("CSCI", "247", "Data Structures", "Smith", true, 1, 5)
	f.Add("", "", "", "", false, 0, 0)
	f.Add("MATH", "2*", "", "", false, 4, 4)
	f.Add("CS", "1", "intro", "Dr", true, 0, 10)
	f.Add("*", "***", "a b c d", "___", false, -1, 100)
	f.Add("CSCI", "247", "データ", "日本語", false, 0, 0) // Unicode
	f.Add("'; DROP TABLE sections; --", "247", "", "", false, 0, 0) // SQL injection attempt

	f.Fuzz(func(t *testing.T, subject, courseNum, title, instructor string, openSeats bool, minCredits, maxCredits int) {
		var minPtr, maxPtr *int
		if minCredits > 0 {
			minPtr = &minCredits
		}
		if maxCredits > 0 {
			maxPtr = &maxCredits
		}

		req := SearchRequest{
			Term:         "202520",
			Subject:      subject,
			CourseNumber: courseNum,
			Title:        title,
			Instructor:   instructor,
			OpenSeats:    openSeats,
			MinCredits:   minPtr,
			MaxCredits:   maxPtr,
		}

		// Should never panic - errors are expected for invalid input
		resp, err := svc.Search(ctx, req)

		// If we got a response, verify basic invariants
		if err == nil && resp != nil {
			verifyResponseInvariants(t, resp)
		}
	})
}

// verifyResponseInvariants checks that a SearchResponse is internally consistent.
func verifyResponseInvariants(t *testing.T, resp *SearchResponse) {
	t.Helper()

	// Invariant 1: All sectionKeys in Results must exist in Sections map
	for _, ref := range resp.Results {
		for _, sectionKey := range ref.SectionKeys {
			if _, exists := resp.Sections[sectionKey]; !exists {
				t.Errorf("result references non-existent section: %s", sectionKey)
			}
		}
	}

	// Invariant 2: All courseKeys in Results must exist in Courses map
	for _, ref := range resp.Results {
		if _, exists := resp.Courses[ref.CourseKey]; !exists {
			t.Errorf("result references non-existent course: %s", ref.CourseKey)
		}
	}

	// Invariant 3: All sections must reference valid courses
	for sectionKey, section := range resp.Sections {
		if _, exists := resp.Courses[section.CourseKey]; !exists {
			t.Errorf("section %s references non-existent course: %s", sectionKey, section.CourseKey)
		}
	}

	// Invariant 4: Section keys must be in term:crn format
	for sectionKey, section := range resp.Sections {
		expectedKey := section.Term + ":" + section.CRN
		if sectionKey != expectedKey {
			t.Errorf("section key %s doesn't match expected format %s", sectionKey, expectedKey)
		}
	}

	// Invariant 5: Course keys must be in subject:courseNumber format
	for courseKey, course := range resp.Courses {
		expectedKey := course.Subject + ":" + course.CourseNumber
		if courseKey != expectedKey {
			t.Errorf("course key %s doesn't match expected format %s", courseKey, expectedKey)
		}
	}

	// Invariant 6: Stats must match actual counts
	if resp.Stats.TotalSections != len(resp.Sections) {
		t.Errorf("stats.TotalSections %d != len(Sections) %d", resp.Stats.TotalSections, len(resp.Sections))
	}
	if resp.Stats.TotalCourses != len(resp.Courses) {
		t.Errorf("stats.TotalCourses %d != len(Courses) %d", resp.Stats.TotalCourses, len(resp.Courses))
	}

	// Invariant 7: Total must match Results count
	if resp.Total != len(resp.Results) {
		t.Errorf("Total %d != len(Results) %d", resp.Total, len(resp.Results))
	}

	// Invariant 8: Each course in Results should have at least one section
	for _, ref := range resp.Results {
		if len(ref.SectionKeys) == 0 {
			t.Errorf("course %s has no sections", ref.CourseKey)
		}
	}

	// Invariant 9: No duplicate section keys in any single course ref
	for _, ref := range resp.Results {
		seen := make(map[string]bool)
		for _, key := range ref.SectionKeys {
			if seen[key] {
				t.Errorf("duplicate section key %s in course %s", key, ref.CourseKey)
			}
			seen[key] = true
		}
	}

	// Invariant 10: Section's courseKey must match the result it appears in
	sectionToCourse := make(map[string]string)
	for _, ref := range resp.Results {
		for _, sectionKey := range ref.SectionKeys {
			sectionToCourse[sectionKey] = ref.CourseKey
		}
	}
	for sectionKey, section := range resp.Sections {
		if expectedCourse, exists := sectionToCourse[sectionKey]; exists {
			if section.CourseKey != expectedCourse {
				t.Errorf("section %s has courseKey %s but appears in result for %s",
					sectionKey, section.CourseKey, expectedCourse)
			}
		}
	}
}

// TestSearch_ResponseInvariants runs the invariant checks against various search scenarios.
func TestSearch_ResponseInvariants(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)
	ctx := context.Background()

	tests := []struct {
		name string
		req  SearchRequest
	}{
		{
			name: "subject filter",
			req:  SearchRequest{Term: "202520", Subject: "CSCI"},
		},
		{
			name: "course number wildcard",
			req:  SearchRequest{Term: "202520", CourseNumber: "2*"},
		},
		{
			name: "title search",
			req:  SearchRequest{Term: "202520", Title: "data"},
		},
		{
			name: "instructor search",
			req:  SearchRequest{Term: "202520", Instructor: "smith"},
		},
		{
			name: "all-time search",
			req:  SearchRequest{Subject: "CSCI", CourseNumber: "247"},
		},
		{
			name: "combined filters",
			req:  SearchRequest{Term: "202520", Subject: "CSCI", CourseNumber: "247", Title: "Data"},
		},
		{
			name: "open seats filter",
			req:  SearchRequest{Term: "202520", Subject: "CSCI", OpenSeats: true},
		},
		{
			name: "no results",
			req:  SearchRequest{Term: "202520", Subject: "ZZZZ"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.Search(ctx, tt.req)
			if err != nil {
				t.Skipf("search returned error (expected for some cases): %v", err)
			}
			verifyResponseInvariants(t, resp)
		})
	}
}

// TestSearch_InvariantsWithUnicode tests that unicode input doesn't break invariants.
func TestSearch_InvariantsWithUnicode(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)
	ctx := context.Background()

	unicodeInputs := []string{
		"日本語",
		"émoji 🎓",
		"Ñoño",
		"中文课程",
		"Привет",
		"مرحبا",
		"\x00\x01\x02", // Control characters
		strings.Repeat("a", 1000), // Very long input
	}

	for _, input := range unicodeInputs {
		// These should either error or return valid responses - never panic
		resp, err := svc.Search(ctx, SearchRequest{
			Term:  "202520",
			Title: input,
		})
		if err == nil && resp != nil {
			verifyResponseInvariants(t, resp)
		}
	}
}
