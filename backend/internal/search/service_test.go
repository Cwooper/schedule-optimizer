package search

import (
	"context"
	"strings"
	"testing"

	"schedule-optimizer/internal/testutil"
)

func TestSearch_NoFilters(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	_, err := svc.Search(context.Background(), SearchRequest{
		Term: "202520",
	})

	if err != ErrNoFilters {
		t.Errorf("expected ErrNoFilters, got %v", err)
	}
}

func TestSearch_WildcardOnlyFilter(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	// Wildcard-only course number should be rejected
	_, err := svc.Search(context.Background(), SearchRequest{
		Term:         "202520",
		CourseNumber: "*",
	})

	if err != ErrFilterTooShort {
		t.Errorf("expected ErrFilterTooShort for wildcard-only courseNumber, got %v", err)
	}

	// Wildcard-only title should be rejected
	_, err = svc.Search(context.Background(), SearchRequest{
		Term:  "202520",
		Title: "***",
	})

	if err != ErrFilterTooShort {
		t.Errorf("expected ErrFilterTooShort for wildcard-only title, got %v", err)
	}

	// Underscore-only should also be rejected (SQL single-char wildcard)
	_, err = svc.Search(context.Background(), SearchRequest{
		Term:       "202520",
		Instructor: "___",
	})

	if err != ErrFilterTooShort {
		t.Errorf("expected ErrFilterTooShort for underscore-only instructor, got %v", err)
	}
}

func TestSearch_FilterMinimumLength(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	// Single character subject should be rejected (min 2)
	_, err := svc.Search(context.Background(), SearchRequest{
		Term:    "202520",
		Subject: "C",
	})

	if err != ErrFilterTooShort {
		t.Errorf("expected ErrFilterTooShort for single-char subject, got %v", err)
	}

	// Single digit course number should be allowed (convenience for level search)
	_, err = svc.Search(context.Background(), SearchRequest{
		Term:         "202520",
		CourseNumber: "2",
	})

	if err != nil {
		t.Errorf("single digit courseNumber should be allowed, got %v", err)
	}

	// Two-character subject should be allowed
	_, err = svc.Search(context.Background(), SearchRequest{
		Term:    "202520",
		Subject: "CS",
	})

	if err != nil {
		t.Errorf("two-char subject should be allowed, got %v", err)
	}
}

func TestSearch_TooManyTokens(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	_, err := svc.Search(context.Background(), SearchRequest{
		Term:  "202520",
		Title: "one two three four",
	})

	if err == nil || err.Error() != "title: too many search terms (max 3 words per field)" {
		t.Errorf("expected too many tokens error for title, got %v", err)
	}

	_, err = svc.Search(context.Background(), SearchRequest{
		Term:       "202520",
		Instructor: "a b c d",
	})

	if err == nil || err.Error() != "instructor: too many search terms (max 3 words per field)" {
		t.Errorf("expected too many tokens error for instructor, got %v", err)
	}
}

func TestSearch_InvalidTerm(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	_, err := svc.Search(context.Background(), SearchRequest{
		Term:    "999999",
		Subject: "CSCI",
	})

	if err != ErrInvalidTerm {
		t.Errorf("expected ErrInvalidTerm, got %v", err)
	}
}

func TestSearch_SubjectFilter(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:    "202520",
		Subject: "csci", // lowercase to test case insensitivity
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 2 sections (CSCI 247 and CSCI 301)
	if len(resp.Sections) != 2 {
		t.Errorf("expected 2 CSCI sections, got %d", len(resp.Sections))
	}

	// Should have 2 courses
	if len(resp.Courses) != 2 {
		t.Errorf("expected 2 CSCI courses, got %d", len(resp.Courses))
	}

	// Verify all sections reference valid courses
	for crn, section := range resp.Sections {
		if _, exists := resp.Courses[section.CourseKey]; !exists {
			t.Errorf("section %s references non-existent course %s", crn, section.CourseKey)
		}
	}

	// Verify all courses are CSCI
	for _, course := range resp.Courses {
		if course.Subject != "CSCI" {
			t.Errorf("expected subject CSCI, got %s", course.Subject)
		}
	}
}

func TestSearch_CourseNumberExact(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:         "202520",
		CourseNumber: "247",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Errorf("expected 1 course for exact 247, got %d", len(resp.Results))
	}

	if len(resp.Results) > 0 {
		courseKey := resp.Results[0].CourseKey
		course, exists := resp.Courses[courseKey]
		if !exists {
			t.Error("course key from results not found in courses map")
		} else if course.CourseNumber != "247" {
			t.Errorf("expected course number 247, got %s", course.CourseNumber)
		}
	}
}

func TestSearch_CourseNumberWildcard(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	// Test explicit wildcard "2*"
	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:         "202520",
		CourseNumber: "2*",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should match 247 and 204 (both start with 2)
	if len(resp.Results) != 2 {
		t.Errorf("expected 2 courses for 2*, got %d", len(resp.Results))
	}

	for _, ref := range resp.Results {
		course := resp.Courses[ref.CourseKey]
		if course.CourseNumber[0] != '2' {
			t.Errorf("expected course number starting with 2, got %s", course.CourseNumber)
		}
	}
}

func TestSearch_CourseNumberAutoWildcard(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	// Test auto-wildcard for 1-2 digit input
	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:         "202520",
		CourseNumber: "2", // Should become "2%"
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Results) != 2 {
		t.Errorf("expected 2 courses for auto-wildcard 2, got %d", len(resp.Results))
	}
}

func TestSearch_TitleToken(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:  "202520",
		Title: "data",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Errorf("expected 1 course matching 'data', got %d", len(resp.Results))
	}

	if len(resp.Results) > 0 {
		course := resp.Courses[resp.Results[0].CourseKey]
		if course.Title != "Data Structures" {
			t.Errorf("expected 'Data Structures', got %s", course.Title)
		}
	}
}

func TestSearch_InstructorToken(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:       "202520",
		Instructor: "smith",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Sections) != 1 {
		t.Errorf("expected 1 section with instructor Smith, got %d", len(resp.Sections))
	}

	for _, section := range resp.Sections {
		if section.Instructor != "Dr. Smith" {
			t.Errorf("expected 'Dr. Smith', got %s", section.Instructor)
		}
	}
}

func TestSearch_InstructorMultipleTokens(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	// "Dr Smith" should match "Dr. Smith" (both tokens must match)
	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:       "202520",
		Instructor: "Dr Smith",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Sections) != 1 {
		t.Errorf("expected 1 section with instructor 'Dr Smith', got %d", len(resp.Sections))
	}
}

func TestSearch_OpenSeatsFilter(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	// Without open_seats filter
	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:    "202520",
		Subject: "CSCI",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	totalSections := len(resp.Sections)

	// With open_seats filter
	resp, err = svc.Search(context.Background(), SearchRequest{
		Term:      "202520",
		Subject:   "CSCI",
		OpenSeats: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have fewer or equal sections with open seats
	if len(resp.Sections) >= totalSections && totalSections > 1 {
		t.Errorf("expected fewer sections with open_seats filter, got %d (total: %d)", len(resp.Sections), totalSections)
	}

	for _, section := range resp.Sections {
		if section.SeatsAvailable <= 0 {
			t.Errorf("expected section with open seats, got %d available", section.SeatsAvailable)
		}
	}
}

func TestSearch_CreditRange(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	minCredits := 5
	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:       "202520",
		Subject:    "MATH",
		MinCredits: &minCredits,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, course := range resp.Courses {
		if course.Credits < minCredits {
			t.Errorf("expected credits >= %d, got %d", minCredits, course.Credits)
		}
	}
}

func TestSearch_CombinedFilters(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:         "202520",
		Subject:      "CSCI",
		CourseNumber: "247",
		Title:        "Data",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Errorf("expected 1 course, got %d", len(resp.Results))
	}
}

func TestSearch_NoResults(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:    "202520",
		Subject: "NONEXISTENT",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Results))
	}

	if len(resp.Courses) != 0 {
		t.Errorf("expected 0 courses, got %d", len(resp.Courses))
	}

	if len(resp.Sections) != 0 {
		t.Errorf("expected 0 sections, got %d", len(resp.Sections))
	}

	if resp.Total != 0 {
		t.Errorf("expected total 0, got %d", resp.Total)
	}
}

func TestSearch_NoMeetingTimes(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:         "202520",
		CourseNumber: "247",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 course, got %d", len(resp.Results))
	}

	// Verify section exists and has expected fields but no meetingTimes
	sectionKey := resp.Results[0].SectionKeys[0]
	section, exists := resp.Sections[sectionKey]
	if !exists {
		t.Fatal("section key from results not found in sections map")
	}

	if section.CRN == "" {
		t.Error("expected section to have a CRN")
	}
	if section.Term != "202520" {
		t.Errorf("expected term 202520, got %s", section.Term)
	}
}

func TestSearch_AllTimeScope(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	// Search without term (all-time)
	resp, err := svc.Search(context.Background(), SearchRequest{
		Subject:      "CSCI",
		CourseNumber: "247",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find 1 course (CSCI 247) with sections from multiple terms (202520 and 202510)
	if len(resp.Results) != 1 {
		t.Errorf("expected 1 course across terms, got %d", len(resp.Results))
	}

	// Should have 2 sections (one per term)
	if len(resp.Sections) != 2 {
		t.Errorf("expected 2 sections across terms, got %d", len(resp.Sections))
	}

	// Course ref should have 2 section keys (one per term)
	if len(resp.Results) > 0 && len(resp.Results[0].SectionKeys) != 2 {
		t.Errorf("expected 2 section keys for course, got %d", len(resp.Results[0].SectionKeys))
	}

	// Should have relevance scores
	for _, ref := range resp.Results {
		if ref.RelevanceScore == 0 {
			t.Error("expected relevance score to be set for all-time search")
		}
	}
}

func TestSearch_SingleTermScoring(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	// Single term search should still have relevance scores (from MatchQualityScorer)
	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:         "202520",
		CourseNumber: "247",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have relevance score from match quality (not recency since single term)
	for _, ref := range resp.Results {
		if ref.RelevanceScore == 0 {
			t.Error("expected relevance score to be set for single-term search (from match quality)")
		}
	}
}

func TestSearch_CourseGrouping(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	// Search for CSCI without term - should group sections by course
	resp, err := svc.Search(context.Background(), SearchRequest{
		Subject: "CSCI",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify results structure
	if resp.Courses == nil {
		t.Fatal("courses map should not be nil")
	}
	if resp.Sections == nil {
		t.Fatal("sections map should not be nil")
	}
	if resp.Results == nil {
		t.Fatal("results slice should not be nil")
	}

	// Each result should reference a valid course
	for _, ref := range resp.Results {
		if _, exists := resp.Courses[ref.CourseKey]; !exists {
			t.Errorf("result references non-existent course: %s", ref.CourseKey)
		}
		// Each section key should reference a valid section
		for _, sectionKey := range ref.SectionKeys {
			section, exists := resp.Sections[sectionKey]
			if !exists {
				t.Errorf("result section key references non-existent section: %s", sectionKey)
			}
			// Section's courseKey should match
			if section.CourseKey != ref.CourseKey {
				t.Errorf("section courseKey %s doesn't match result courseKey %s", section.CourseKey, ref.CourseKey)
			}
		}
	}

	// Total should match results count
	if resp.Total != len(resp.Results) {
		t.Errorf("total %d doesn't match results length %d", resp.Total, len(resp.Results))
	}
}

func TestSearch_ResponseStructure(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:    "202520",
		Subject: "CSCI",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify CourseInfo structure
	for key, course := range resp.Courses {
		expectedKey := course.Subject + ":" + course.CourseNumber
		if key != expectedKey {
			t.Errorf("course key %s doesn't match expected %s", key, expectedKey)
		}
		if course.Subject == "" {
			t.Error("course subject should not be empty")
		}
		if course.CourseNumber == "" {
			t.Error("course number should not be empty")
		}
		if course.Title == "" {
			t.Error("course title should not be empty")
		}
	}

	// Verify SectionInfo structure
	for sectionKey, section := range resp.Sections {
		expectedKey := section.Term + ":" + section.CRN
		if sectionKey != expectedKey {
			t.Errorf("section key %s doesn't match expected %s", sectionKey, expectedKey)
		}
		if section.CourseKey == "" {
			t.Error("section courseKey should not be empty")
		}
		if section.Term == "" {
			t.Error("section term should not be empty")
		}
	}

	// Verify CourseRef structure
	for _, ref := range resp.Results {
		if ref.CourseKey == "" {
			t.Error("result courseKey should not be empty")
		}
		if len(ref.SectionKeys) == 0 {
			t.Error("result should have at least one section key")
		}
	}
}

// Scorer tests

func TestRecencyScorer(t *testing.T) {
	scorer := NewRecencyScorer()

	// More recent terms should score higher for all-time searches
	currentSection := &sectionRow{Term: "202520"}
	olderSection := &sectionRow{Term: "202510"}

	allTimeReq := &SearchRequest{} // No term = all-time
	currentScore := scorer.Score(currentSection, allTimeReq)
	olderScore := scorer.Score(olderSection, allTimeReq)

	if currentScore <= olderScore {
		t.Errorf("expected current term score > older term score, got %f <= %f", currentScore, olderScore)
	}

	// Single-term searches should return 0 (recency is meaningless)
	singleTermReq := &SearchRequest{Term: "202520"}
	singleTermScore := scorer.Score(currentSection, singleTermReq)

	if singleTermScore != 0 {
		t.Errorf("expected 0 for single-term search, got %f", singleTermScore)
	}
}

func TestMatchQualityScorer(t *testing.T) {
	scorer := NewMatchQualityScorer()

	// Exact match should score higher than partial
	exactSection := &sectionRow{CourseNumber: "247", Subject: "CSCI"}
	partialSection := &sectionRow{CourseNumber: "2471", Subject: "CSCI"}

	req := &SearchRequest{CourseNumber: "247", Subject: "CSCI"}

	exactScore := scorer.Score(exactSection, req)
	partialScore := scorer.Score(partialSection, req)

	if exactScore <= partialScore {
		t.Errorf("expected exact match score > partial match score, got %f <= %f", exactScore, partialScore)
	}
}

func TestSearch_YearScope(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	// Search with year scope (2025 = Fall 2024 through Summer 2025)
	// Our test data has terms 202520 (Spring 2025) and 202510 (Winter 2025)
	resp, err := svc.Search(context.Background(), SearchRequest{
		Year:         2025,
		Subject:      "CSCI",
		CourseNumber: "247",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find sections from both terms in the academic year
	if len(resp.Sections) < 1 {
		t.Errorf("expected at least 1 section for year scope, got %d", len(resp.Sections))
	}
}

func TestSearch_SectionKeyUniqueness(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	// Search across terms for CSCI 247 which exists in both 202520 and 202510
	resp, err := svc.Search(context.Background(), SearchRequest{
		Subject:      "CSCI",
		CourseNumber: "247",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 2 sections (one per term), not 1 (which would indicate collision)
	if len(resp.Sections) != 2 {
		t.Errorf("expected 2 sections (one per term), got %d - possible CRN collision", len(resp.Sections))
	}

	// Verify section keys are in term:crn format
	for sectionKey := range resp.Sections {
		if !strings.Contains(sectionKey, ":") {
			t.Errorf("section key %s should be in term:crn format", sectionKey)
		}
	}

	// Verify each section key is unique
	seen := make(map[string]bool)
	for _, ref := range resp.Results {
		for _, key := range ref.SectionKeys {
			if seen[key] {
				t.Errorf("duplicate section key: %s", key)
			}
			seen[key] = true
		}
	}
}

func TestSearch_StatsPopulated(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:    "202520",
		Subject: "CSCI",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify stats are populated
	if resp.Stats.TotalSections != len(resp.Sections) {
		t.Errorf("stats.TotalSections %d doesn't match sections count %d", resp.Stats.TotalSections, len(resp.Sections))
	}
	if resp.Stats.TotalCourses != len(resp.Courses) {
		t.Errorf("stats.TotalCourses %d doesn't match courses count %d", resp.Stats.TotalCourses, len(resp.Courses))
	}
	if resp.Stats.TimeMs <= 0 {
		t.Error("stats.TimeMs should be positive")
	}
}

// Helper function tests

func TestSplitTokens(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected int
		wantErr  bool
	}{
		{"", 3, 0, false},
		{"hello", 3, 1, false},
		{"hello world", 3, 2, false},
		{"hello-world", 3, 2, false},
		{"hello world again", 3, 3, false},
		{"a b c d", 3, 0, true}, // Too many tokens
		{"See-Mong Tan", 3, 3, false},
	}

	for _, tt := range tests {
		tokens, err := splitTokens(tt.input, tt.max)
		if tt.wantErr {
			if err == nil {
				t.Errorf("splitTokens(%q, %d) expected error", tt.input, tt.max)
			}
		} else {
			if err != nil {
				t.Errorf("splitTokens(%q, %d) unexpected error: %v", tt.input, tt.max, err)
			}
			if len(tokens) != tt.expected {
				t.Errorf("splitTokens(%q, %d) = %d tokens, expected %d", tt.input, tt.max, len(tokens), tt.expected)
			}
		}
	}
}

func TestBuildCourseNumberPattern(t *testing.T) {
	tests := []struct {
		input    string
		expected *string
	}{
		// Empty
		{"", nil},
		// Exact match (3+ chars, no wildcard)
		{"247", strPtr("247")},
		{"301", strPtr("301")},
		// Auto-wildcard for 1-2 chars
		{"2", strPtr("2%")},
		{"20", strPtr("20%")},
		// Explicit suffix wildcard
		{"2*", strPtr("2%")},
		{"2%", strPtr("2%")},
		{"30*", strPtr("30%")},
		// Prefix wildcard
		{"*97", strPtr("%97")},
		{"*01", strPtr("%01")},
		// Internal wildcard
		{"2*7", strPtr("2%7")},
		{"4*7X", strPtr("4%7X")},
		// Multiple wildcards
		{"*97*", strPtr("%97%")},
		{"*9*X", strPtr("%9%X")},
		// Mixed * and %
		{"2*7%", strPtr("2%7%")},
	}

	for _, tt := range tests {
		result := buildCourseNumberPattern(tt.input)
		if tt.expected == nil {
			if result != nil {
				t.Errorf("buildCourseNumberPattern(%q) = %v, expected nil", tt.input, *result)
			}
		} else {
			if result == nil {
				t.Errorf("buildCourseNumberPattern(%q) = nil, expected %v", tt.input, *tt.expected)
			} else if *result != *tt.expected {
				t.Errorf("buildCourseNumberPattern(%q) = %v, expected %v", tt.input, *result, *tt.expected)
			}
		}
	}
}

func strPtr(s string) *string {
	return &s
}
