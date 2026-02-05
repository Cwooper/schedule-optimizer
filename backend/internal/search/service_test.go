package search

import (
	"context"
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

	if len(resp.Sections) != 2 {
		t.Errorf("expected 2 CSCI sections, got %d", len(resp.Sections))
	}

	for _, s := range resp.Sections {
		if s.Subject != "CSCI" {
			t.Errorf("expected subject CSCI, got %s", s.Subject)
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

	if len(resp.Sections) != 1 {
		t.Errorf("expected 1 section for exact 247, got %d", len(resp.Sections))
	}

	if len(resp.Sections) > 0 && resp.Sections[0].CourseNumber != "247" {
		t.Errorf("expected course number 247, got %s", resp.Sections[0].CourseNumber)
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

	// Should match 247, 204, 301 (all start with 2 or 3... wait, 301 starts with 3)
	// Actually 247 and 204 start with 2, 301 starts with 3
	if len(resp.Sections) != 2 {
		t.Errorf("expected 2 sections for 2*, got %d", len(resp.Sections))
	}

	for _, s := range resp.Sections {
		if s.CourseNumber[0] != '2' {
			t.Errorf("expected course number starting with 2, got %s", s.CourseNumber)
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

	if len(resp.Sections) != 2 {
		t.Errorf("expected 2 sections for auto-wildcard 2, got %d", len(resp.Sections))
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

	if len(resp.Sections) != 1 {
		t.Errorf("expected 1 section matching 'data', got %d", len(resp.Sections))
	}

	if len(resp.Sections) > 0 && resp.Sections[0].Title != "Data Structures" {
		t.Errorf("expected 'Data Structures', got %s", resp.Sections[0].Title)
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

	if len(resp.Sections) > 0 && resp.Sections[0].Instructor != "Dr. Smith" {
		t.Errorf("expected 'Dr. Smith', got %s", resp.Sections[0].Instructor)
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

	for _, s := range resp.Sections {
		if s.SeatsAvailable <= 0 {
			t.Errorf("expected section with open seats, got %d available", s.SeatsAvailable)
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

	for _, s := range resp.Sections {
		if s.Credits < minCredits {
			t.Errorf("expected credits >= %d, got %d", minCredits, s.Credits)
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

	if len(resp.Sections) != 1 {
		t.Errorf("expected 1 section, got %d", len(resp.Sections))
	}
}

func TestSearch_Pagination(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	defer db.Close()
	testutil.SeedTestData(t, db)

	svc := NewService(db, queries)

	// Get all results
	resp, err := svc.Search(context.Background(), SearchRequest{
		Term:    "202520",
		Subject: "CSCI",
		Limit:   10,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	totalSections := len(resp.Sections)

	// Get with limit 1
	resp, err = svc.Search(context.Background(), SearchRequest{
		Term:    "202520",
		Subject: "CSCI",
		Limit:   1,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Sections) != 1 {
		t.Errorf("expected 1 section with limit 1, got %d", len(resp.Sections))
	}

	if totalSections > 1 && !resp.HasMore {
		t.Error("expected HasMore to be true")
	}

	// Get with offset 1
	resp, err = svc.Search(context.Background(), SearchRequest{
		Term:    "202520",
		Subject: "CSCI",
		Limit:   10,
		Offset:  1,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if totalSections > 1 && len(resp.Sections) != totalSections-1 {
		t.Errorf("expected %d sections with offset 1, got %d", totalSections-1, len(resp.Sections))
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

	if len(resp.Sections) != 0 {
		t.Errorf("expected 0 sections, got %d", len(resp.Sections))
	}

	if resp.Total != 0 {
		t.Errorf("expected total 0, got %d", resp.Total)
	}
}

func TestSearch_MeetingTimes(t *testing.T) {
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

	if len(resp.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(resp.Sections))
	}

	section := resp.Sections[0]
	if len(section.MeetingTimes) == 0 {
		t.Error("expected meeting times to be populated")
	}

	if len(section.MeetingTimes) > 0 {
		mt := section.MeetingTimes[0]
		if mt.StartTime != "1000" || mt.EndTime != "1050" {
			t.Errorf("expected meeting time 1000-1050, got %s-%s", mt.StartTime, mt.EndTime)
		}
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

	// Should find sections from multiple terms (202520 and 202510)
	if len(resp.Sections) != 2 {
		t.Errorf("expected 2 sections across terms, got %d", len(resp.Sections))
	}

	// Should have relevance scores
	for _, s := range resp.Sections {
		if s.RelevanceScore == 0 {
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
	for _, s := range resp.Sections {
		if s.RelevanceScore == 0 {
			t.Error("expected relevance score to be set for single-term search (from match quality)")
		}
	}
}

// Scorer tests

func TestRecencyScorer(t *testing.T) {
	scorer := NewRecencyScorer()

	// More recent terms should score higher for all-time searches
	currentSection := &SectionResult{Term: "202520"}
	olderSection := &SectionResult{Term: "202510"}

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
	exactSection := &SectionResult{CourseNumber: "247", Subject: "CSCI"}
	partialSection := &SectionResult{CourseNumber: "2471", Subject: "CSCI"}

	req := &SearchRequest{CourseNumber: "247", Subject: "CSCI"}

	exactScore := scorer.Score(exactSection, req)
	partialScore := scorer.Score(partialSection, req)

	if exactScore <= partialScore {
		t.Errorf("expected exact match score > partial match score, got %f <= %f", exactScore, partialScore)
	}
}

func TestCompositeScorer(t *testing.T) {
	recency := NewRecencyScorer()
	quality := NewMatchQualityScorer()
	composite := NewCompositeScorer([]Scorer{recency, quality}, []float64{1.0, 1.0})

	section := &SectionResult{Term: "202520", CourseNumber: "247", Subject: "CSCI"}
	req := &SearchRequest{CourseNumber: "247", Subject: "CSCI"}

	compositeScore := composite.Score(section, req)
	recencyScore := recency.Score(section, req)
	qualityScore := quality.Score(section, req)

	expectedScore := recencyScore + qualityScore
	if compositeScore != expectedScore {
		t.Errorf("expected composite score %f, got %f", expectedScore, compositeScore)
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
