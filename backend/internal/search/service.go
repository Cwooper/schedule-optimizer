package search

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"schedule-optimizer/internal/jobs"
	"schedule-optimizer/internal/store"
)

var (
	ErrNoFilters      = errors.New("at least one search filter is required (subject, courseNumber, title, or instructor)")
	ErrTooManyTokens  = errors.New("too many search terms (max 3 words per field)")
	ErrInvalidTerm    = errors.New("invalid term code")
	ErrInvalidYear    = errors.New("invalid academic year")
	ErrWildcardOnly   = errors.New("search filter cannot be only wildcards")
	ErrFilterTooShort = errors.New("search filter must be at least 2 characters (excluding wildcards)")
)

// Service handles course search operations.
type Service struct {
	db      *sql.DB
	queries *store.Queries
	scorers []Scorer
}

// NewService creates a new search service.
func NewService(db *sql.DB, queries *store.Queries) *Service {
	return &Service{
		db:      db,
		queries: queries,
		scorers: []Scorer{
			NewRecencyScorer(),
			NewMatchQualityScorer(),
		},
	}
}

// Search performs a course search with the given parameters.
func (s *Service) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	startTime := time.Now()

	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	titleTokens, err := splitTokens(req.Title, MaxTokens)
	if err != nil {
		return nil, fmt.Errorf("title: %w", err)
	}
	instrTokens, err := splitTokens(req.Instructor, MaxTokens)
	if err != nil {
		return nil, fmt.Errorf("instructor: %w", err)
	}

	terms, err := s.resolveTerms(ctx, req)
	if err != nil {
		return nil, err
	}

	courseNumberPattern := buildCourseNumberPattern(req.CourseNumber)

	var allResults []*store.SearchSectionsRow
	var sectionLimitHit bool

	if len(terms) == 0 {
		// All-time search (no term filter)
		results, err := s.executeSearch(ctx, nil, req, courseNumberPattern, titleTokens, instrTokens, MaxSectionFetch)
		if err != nil {
			return nil, err
		}
		allResults = results
	} else if len(terms) == 1 {
		// Single term search
		results, err := s.executeSearch(ctx, &terms[0], req, courseNumberPattern, titleTokens, instrTokens, MaxSectionFetch)
		if err != nil {
			return nil, err
		}
		allResults = results
	} else {
		// Multi-term search (year scope) - search each term and combine
		for _, term := range terms {
			termCopy := term
			results, err := s.executeSearch(ctx, &termCopy, req, courseNumberPattern, titleTokens, instrTokens, MaxSectionFetch)
			if err != nil {
				return nil, err
			}
			allResults = append(allResults, results...)
		}
		// Safety limit after combining
		if len(allResults) > MaxSectionFetch {
			allResults = allResults[:MaxSectionFetch]
		}
	}

	if len(allResults) >= MaxSectionFetch {
		sectionLimitHit = true
	}

	// Convert to sectionRow for scoring and grouping
	rows := make([]*sectionRow, len(allResults))
	for i, row := range allResults {
		rows[i] = rowToSectionRow(row)
	}

	// Apply scoring to all sections
	for _, row := range rows {
		var totalScore float64
		for _, scorer := range s.scorers {
			totalScore += scorer.Score(row, &req)
		}
		row.RelevanceScore = totalScore
	}

	return s.buildResponse(rows, sectionLimitHit, startTime)
}

// buildResponse groups sections by course and builds the normalized response.
func (s *Service) buildResponse(rows []*sectionRow, sectionLimitHit bool, startTime time.Time) (*SearchResponse, error) {
	courses := make(map[string]CourseInfo)
	sections := make(map[string]SectionInfo)

	// Track course refs with their max scores for ordering
	type courseRefData struct {
		courseKey    string
		sectionKeys []string
		maxScore    float64
	}
	courseRefMap := make(map[string]*courseRefData)

	for _, row := range rows {
		courseKey := row.Subject + ":" + row.CourseNumber
		sectionKey := row.Term + ":" + row.CRN // Unique across terms

		// Add course if not seen
		if _, exists := courses[courseKey]; !exists {
			courses[courseKey] = CourseInfo{
				Subject:      row.Subject,
				CourseNumber: row.CourseNumber,
				Title:        row.Title,
				Credits:      row.Credits,
				CreditsHigh:  row.CreditsHigh,
			}
			courseRefMap[courseKey] = &courseRefData{
				courseKey:    courseKey,
				sectionKeys: []string{},
				maxScore:    0,
			}
		}

		// Add section (keyed by term:crn for cross-term uniqueness)
		sections[sectionKey] = SectionInfo{
			CRN:             row.CRN,
			Term:            row.Term,
			CourseKey:       courseKey,
			Instructor:      row.Instructor,
			InstructorEmail: row.InstructorEmail,
			Enrollment:      row.Enrollment,
			MaxEnrollment:   row.MaxEnrollment,
			SeatsAvailable:  row.SeatsAvailable,
			WaitCount:       row.WaitCount,
			IsOpen:          row.IsOpen,
			Campus:          row.Campus,
			ScheduleType:    row.ScheduleType,
		}

		// Update course ref
		ref := courseRefMap[courseKey]
		ref.sectionKeys = append(ref.sectionKeys, sectionKey)
		if row.RelevanceScore > ref.maxScore {
			ref.maxScore = row.RelevanceScore
		}
	}

	// Build results slice sorted by relevance score
	results := make([]CourseRef, 0, len(courseRefMap))
	for _, ref := range courseRefMap {
		results = append(results, CourseRef{
			CourseKey:      ref.courseKey,
			SectionKeys:    ref.sectionKeys,
			RelevanceScore: ref.maxScore,
		})
	}

	// Sort by relevance score descending
	slices.SortFunc(results, func(a, b CourseRef) int {
		if a.RelevanceScore > b.RelevanceScore {
			return -1
		}
		if a.RelevanceScore < b.RelevanceScore {
			return 1
		}
		return 0
	})

	// Build warning from whichever limits were hit
	courseLimitHit := len(results) > MaxCourseResults
	if courseLimitHit {
		results = results[:MaxCourseResults]

		// Remove sections and courses that are no longer in results
		keepCourses := make(map[string]bool)
		keepSectionKeys := make(map[string]bool)
		for _, ref := range results {
			keepCourses[ref.CourseKey] = true
			for _, sectionKey := range ref.SectionKeys {
				keepSectionKeys[sectionKey] = true
			}
		}

		for key := range courses {
			if !keepCourses[key] {
				delete(courses, key)
			}
		}
		for sectionKey := range sections {
			if !keepSectionKeys[sectionKey] {
				delete(sections, sectionKey)
			}
		}
	}

	var warning string
	switch {
	case sectionLimitHit && courseLimitHit:
		warning = SectionWarningMessage + " " + CourseWarningMessage
	case sectionLimitHit:
		warning = SectionWarningMessage
	case courseLimitHit:
		warning = CourseWarningMessage
	}

	return &SearchResponse{
		Courses:  courses,
		Sections: sections,
		Results:  results,
		Total:    len(results),
		Warning:  warning,
		Stats: SearchStats{
			TotalSections:   len(sections),
			TotalCourses:    len(courses),
			TimeMs:          float64(time.Since(startTime).Microseconds()) / 1000.0,
			SectionLimitHit: sectionLimitHit,
		},
	}, nil
}

// validateRequest checks that the request has valid filters.
func (s *Service) validateRequest(req SearchRequest) error {
	// Check each filter for validity (not just wildcards, minimum meaningful content)
	subjectValid := isValidFilter(req.Subject, 2)
	courseNumValid := isValidFilter(req.CourseNumber, 1) // Allow single digit for level search
	titleValid := isValidFilter(req.Title, 2)
	instructorValid := isValidFilter(req.Instructor, 2)

	hasFilter := subjectValid || courseNumValid || titleValid || instructorValid

	if !hasFilter {
		// Check if they provided filters that were rejected
		if req.Subject != "" || req.CourseNumber != "" || req.Title != "" || req.Instructor != "" {
			return ErrFilterTooShort
		}
		return ErrNoFilters
	}
	return nil
}

// isValidFilter checks if a filter value is meaningful (not just wildcards).
// Returns true if the filter has at least minChars non-wildcard characters.
func isValidFilter(value string, minChars int) bool {
	if value == "" {
		return false
	}
	// Strip wildcards and check remaining length
	cleaned := strings.ReplaceAll(value, "*", "")
	cleaned = strings.ReplaceAll(cleaned, "%", "")
	cleaned = strings.ReplaceAll(cleaned, "_", "") // Also treat _ as wildcard
	return len(cleaned) >= minChars
}

// resolveTerms converts scope parameters to actual term codes.
func (s *Service) resolveTerms(ctx context.Context, req SearchRequest) ([]string, error) {
	if req.Term != "" {
		// Validate term exists
		_, err := s.queries.GetTermByCode(ctx, req.Term)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrInvalidTerm
			}
			return nil, err
		}
		return []string{req.Term}, nil
	}

	if req.Year > 0 {
		if req.Year < 2000 || req.Year > 2100 {
			return nil, ErrInvalidYear
		}
		return jobs.GetAcademicYearTerms(req.Year), nil
	}

	// All-time: no term filter
	return nil, nil
}

// executeSearch runs the sqlc query with the given parameters.
func (s *Service) executeSearch(
	ctx context.Context,
	term *string,
	req SearchRequest,
	courseNumberPattern *string,
	titleTokens, instrTokens []any,
	limit int,
) ([]*store.SearchSectionsRow, error) {
	params := store.SearchSectionsParams{
		ResultLimit: int64(limit),
	}

	// Set term filter
	if term != nil {
		params.Term = *term
	}

	// Set subject filter
	if req.Subject != "" {
		params.Subject = strings.ToUpper(req.Subject)
	}

	// Set course number pattern
	if courseNumberPattern != nil {
		params.CourseNumber = *courseNumberPattern
	}

	// Set title tokens
	if len(titleTokens) >= 1 {
		params.TitleT1 = titleTokens[0]
	}
	if len(titleTokens) >= 2 {
		params.TitleT2 = titleTokens[1]
	}
	if len(titleTokens) >= 3 {
		params.TitleT3 = titleTokens[2]
	}

	// Set instructor tokens
	if len(instrTokens) >= 1 {
		params.InstrT1 = instrTokens[0]
	}
	if len(instrTokens) >= 2 {
		params.InstrT2 = instrTokens[1]
	}
	if len(instrTokens) >= 3 {
		params.InstrT3 = instrTokens[2]
	}

	// Set open seats filter
	if req.OpenSeats {
		params.OpenSeats = 1
	}

	// Set credit range
	if req.MinCredits != nil {
		params.MinCredits = int64(*req.MinCredits)
	}
	if req.MaxCredits != nil {
		params.MaxCredits = int64(*req.MaxCredits)
	}

	return s.queries.SearchSections(ctx, params)
}

// splitTokens splits input on spaces and hyphens into up to max tokens.
func splitTokens(input string, max int) ([]any, error) {
	if input == "" {
		return nil, nil
	}

	tokens := strings.FieldsFunc(strings.ToLower(input), func(r rune) bool {
		return r == ' ' || r == '-'
	})

	if len(tokens) > max {
		return nil, ErrTooManyTokens
	}

	result := make([]any, len(tokens))
	for i, t := range tokens {
		result[i] = t
	}
	return result, nil
}

// buildCourseNumberPattern converts user input to a LIKE pattern.
// Wildcards (*) can appear anywhere and are converted to SQL %.
// Examples:
//   - "247"   -> "247"  (exact match)
//   - "2"     -> "2%"   (auto-wildcard for 1-2 chars)
//   - "2*"    -> "2%"   (suffix wildcard)
//   - "*97"   -> "%97"  (prefix wildcard - matches 197, 297, 397, etc.)
//   - "2*7"   -> "2%7"  (internal wildcard - matches 207, 217, 247, etc.)
//   - "*97*"  -> "%97%" (both - matches anything containing 97)
//   - "4*7X"  -> "4%7X" (matches 407X, 417X, 497X, etc.)
func buildCourseNumberPattern(input string) *string {
	if input == "" {
		return nil
	}

	input = strings.ToUpper(input)

	// Check if input contains any wildcards
	hasWildcard := strings.Contains(input, "*") || strings.Contains(input, "%")

	if hasWildcard {
		// Replace * with % for SQL LIKE
		pattern := strings.ReplaceAll(input, "*", "%")
		return &pattern
	}

	// Auto-wildcard for 1-2 character inputs (convenience for level searches)
	if len(input) <= 2 {
		pattern := input + "%"
		return &pattern
	}

	// Exact match
	return &input
}

// rowToSectionRow converts a database row to the internal sectionRow type.
func rowToSectionRow(row *store.SearchSectionsRow) *sectionRow {
	return &sectionRow{
		ID:              row.ID,
		CRN:             row.Crn,
		Term:            row.Term,
		Subject:         row.Subject,
		CourseNumber:    row.CourseNumber,
		Title:           row.Title,
		Credits:         int(nullInt(row.CreditHoursLow)),
		CreditsHigh:     int(nullInt(row.CreditHoursHigh)),
		Enrollment:      int(nullInt(row.Enrollment)),
		MaxEnrollment:   int(nullInt(row.MaxEnrollment)),
		SeatsAvailable:  int(nullInt(row.SeatsAvailable)),
		WaitCount:       int(nullInt(row.WaitCount)),
		IsOpen:          nullInt(row.IsOpen) == 1,
		Campus:          nullString(row.Campus),
		ScheduleType:    nullString(row.ScheduleType),
		Instructor:      nullString(row.InstructorName),
		InstructorEmail: nullString(row.InstructorEmail),
	}
}

// Helper functions for null types
func nullString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func nullInt(i sql.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	}
	return 0
}
