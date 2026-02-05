package search

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"schedule-optimizer/internal/jobs"
	"schedule-optimizer/internal/store"
)

var (
	ErrNoFilters      = errors.New("at least one search filter is required (subject, courseNumber, title, or instructor)")
	ErrTooManyTokens  = errors.New("too many search terms (max 3 words per field)")
	ErrInvalidTerm    = errors.New("invalid term code")
	ErrInvalidYear    = errors.New("invalid academic year")
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
	// Validate request
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	// Apply defaults
	limit := req.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	// Split tokens for title and instructor
	titleTokens, err := splitTokens(req.Title, MaxTokens)
	if err != nil {
		return nil, fmt.Errorf("title: %w", err)
	}
	instrTokens, err := splitTokens(req.Instructor, MaxTokens)
	if err != nil {
		return nil, fmt.Errorf("instructor: %w", err)
	}

	// Resolve terms based on scope
	terms, err := s.resolveTerms(ctx, req)
	if err != nil {
		return nil, err
	}

	// Build course number pattern
	courseNumberPattern := buildCourseNumberPattern(req.CourseNumber)

	// Execute search for each term (or once if specific term or all-time)
	var allResults []*store.SearchSectionsRow

	if len(terms) == 0 {
		// All-time search (no term filter)
		results, err := s.executeSearch(ctx, nil, req, courseNumberPattern, titleTokens, instrTokens, limit+1, req.Offset)
		if err != nil {
			return nil, err
		}
		allResults = results
	} else if len(terms) == 1 {
		// Single term search
		results, err := s.executeSearch(ctx, &terms[0], req, courseNumberPattern, titleTokens, instrTokens, limit+1, req.Offset)
		if err != nil {
			return nil, err
		}
		allResults = results
	} else {
		// Multi-term search (year scope) - search each term and combine
		// Fetch extra to account for offset, since we can't push offset to individual term queries
		fetchPerTerm := limit + req.Offset + 1
		for _, term := range terms {
			termCopy := term
			results, err := s.executeSearch(ctx, &termCopy, req, courseNumberPattern, titleTokens, instrTokens, fetchPerTerm, 0)
			if err != nil {
				return nil, err
			}
			allResults = append(allResults, results...)
		}
		// Results are already ordered by term DESC within each batch
		// Apply offset after combining
		if req.Offset > 0 && len(allResults) > req.Offset {
			allResults = allResults[req.Offset:]
		} else if req.Offset > 0 {
			allResults = nil // Offset beyond available results
		}
		// Trim to limit+1 (for hasMore detection)
		if len(allResults) > limit+1 {
			allResults = allResults[:limit+1]
		}
	}

	// Check if there are more results
	hasMore := len(allResults) > limit
	if hasMore {
		allResults = allResults[:limit]
	}

	// Convert to SectionResult
	sections := make([]*SectionResult, len(allResults))
	sectionIDs := make([]int64, len(allResults))
	for i, row := range allResults {
		sections[i] = rowToSectionResult(row)
		sectionIDs[i] = row.ID
	}

	// Fetch meeting times for all sections
	if len(sectionIDs) > 0 {
		meetingsBySection, err := s.fetchMeetingTimes(ctx, sectionIDs)
		if err != nil {
			slog.Warn("Failed to fetch meeting times for search results", "error", err)
		} else {
			for i, id := range sectionIDs {
				sections[i].MeetingTimes = meetingsBySection[id]
			}
		}
	}

	// Apply scoring for all-time searches
	isAllTime := req.Term == "" && req.Year == 0
	if isAllTime && len(s.scorers) > 0 {
		for _, section := range sections {
			var totalScore float64
			for _, scorer := range s.scorers {
				totalScore += scorer.Score(section, &req)
			}
			section.RelevanceScore = totalScore
		}
		// Sort by relevance score descending
		sortByRelevance(sections)
	}

	// Build response
	response := &SearchResponse{
		Sections: sections,
		Total:    len(sections),
		HasMore:  hasMore,
	}

	// Add warning for broad queries
	if hasMore || len(sections) >= limit {
		response.Warning = WarningMessage
	}

	return response, nil
}

// validateRequest checks that the request has at least one filter.
func (s *Service) validateRequest(req SearchRequest) error {
	hasFilter := req.Subject != "" ||
		req.CourseNumber != "" ||
		req.Title != "" ||
		req.Instructor != ""

	if !hasFilter {
		return ErrNoFilters
	}
	return nil
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
	limit, offset int,
) ([]*store.SearchSectionsRow, error) {
	params := store.SearchSectionsParams{
		ResultLimit:  int64(limit),
		ResultOffset: int64(offset),
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

// fetchMeetingTimes retrieves meeting times for multiple sections efficiently.
func (s *Service) fetchMeetingTimes(ctx context.Context, sectionIDs []int64) (map[int64][]MeetingTimeInfo, error) {
	result := make(map[int64][]MeetingTimeInfo)

	// Fetch meeting times for each section
	// TODO: Optimize with a batch query if performance becomes an issue
	for _, sectionID := range sectionIDs {
		meetings, err := s.queries.GetMeetingTimesBySection(ctx, sectionID)
		if err != nil {
			continue // Skip sections with no meeting times
		}

		for _, m := range meetings {
			mt := MeetingTimeInfo{
				Days: []bool{
					m.Sunday.Valid && m.Sunday.Int64 == 1,
					m.Monday.Valid && m.Monday.Int64 == 1,
					m.Tuesday.Valid && m.Tuesday.Int64 == 1,
					m.Wednesday.Valid && m.Wednesday.Int64 == 1,
					m.Thursday.Valid && m.Thursday.Int64 == 1,
					m.Friday.Valid && m.Friday.Int64 == 1,
					m.Saturday.Valid && m.Saturday.Int64 == 1,
				},
				StartTime: nullString(m.StartTime),
				EndTime:   nullString(m.EndTime),
				Building:  nullString(m.Building),
				Room:      nullString(m.Room),
			}
			result[sectionID] = append(result[sectionID], mt)
		}
	}

	return result, nil
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
// Auto-wildcards 1-2 digit inputs: "2" -> "2%", "20" -> "20%"
// Explicit wildcards: "2*" -> "2%"
// Exact match: "201" -> "201"
func buildCourseNumberPattern(input string) *string {
	if input == "" {
		return nil
	}

	input = strings.ToUpper(input)

	// Handle explicit wildcards
	if strings.HasSuffix(input, "*") || strings.HasSuffix(input, "%") {
		pattern := strings.TrimSuffix(strings.TrimSuffix(input, "*"), "%") + "%"
		return &pattern
	}

	// Auto-wildcard for 1-2 digit inputs
	if len(input) <= 2 {
		pattern := input + "%"
		return &pattern
	}

	// Exact match
	return &input
}

// rowToSectionResult converts a database row to a SectionResult.
func rowToSectionResult(row *store.SearchSectionsRow) *SectionResult {
	return &SectionResult{
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

// sortByRelevance sorts sections by relevance score in descending order.
func sortByRelevance(sections []*SectionResult) {
	slices.SortFunc(sections, func(a, b *SectionResult) int {
		if a.RelevanceScore > b.RelevanceScore {
			return -1
		}
		if a.RelevanceScore < b.RelevanceScore {
			return 1
		}
		return 0
	})
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
