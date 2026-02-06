package api

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"schedule-optimizer/internal/cache"
	"schedule-optimizer/internal/generator"
	"schedule-optimizer/internal/jobs"
	"schedule-optimizer/internal/search"
	"schedule-optimizer/internal/store"

	"github.com/gin-gonic/gin"
)

const maxTermsReturned = 8 // ~2 years of quarters

// Handlers contains all HTTP handler dependencies.
type Handlers struct {
	db        *sql.DB
	cache     *cache.ScheduleCache
	generator *generator.Service
	queries   *store.Queries
	search    *search.Service
}

// Response types for type-safe JSON serialization

type MeetingTimeInfo struct {
	Days      []bool `json:"days"` // [Sun, Mon, Tue, Wed, Thu, Fri, Sat]
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Building  string `json:"building"`
	Room      string `json:"room"`
}

type SectionResponse struct {
	CRN            string            `json:"crn"`
	Term           string            `json:"term"`
	Subject        string            `json:"subject"`
	CourseNumber   string            `json:"courseNumber"`
	Title          string            `json:"title"`
	Instructor     string            `json:"instructor"`
	Credits        int               `json:"credits"`
	Enrollment     int64             `json:"enrollment"`
	MaxEnrollment  int64             `json:"maxEnrollment"`
	SeatsAvailable int64             `json:"seatsAvailable"`
	WaitCount      int64             `json:"waitCount"`
	IsOpen         bool              `json:"isOpen"`
	MeetingTimes   []MeetingTimeInfo `json:"meetingTimes"`
}

type CRNResponse struct {
	Section *SectionResponse `json:"section"`
}

type CourseInfo struct {
	Subject      string `json:"subject"`
	CourseNumber string `json:"courseNumber"`
	Title        string `json:"title"`
	Credits      int    `json:"credits"`
}

type CourseSectionInfo struct {
	CRN            string `json:"crn"`
	Instructor     string `json:"instructor"`
	Enrollment     int64  `json:"enrollment"`
	MaxEnrollment  int64  `json:"maxEnrollment"`
	SeatsAvailable int64  `json:"seatsAvailable"`
	WaitCount      int64  `json:"waitCount"`
	IsOpen         bool   `json:"isOpen"`
}

type CourseResponse struct {
	Course       *CourseInfo         `json:"course"`
	Sections     []CourseSectionInfo `json:"sections,omitempty"`
	SectionCount int                 `json:"sectionCount,omitempty"`
}

// Helper functions for converting sql.Null types

func fromNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func fromNullInt64(ni sql.NullInt64) int {
	if ni.Valid {
		return int(ni.Int64)
	}
	return 0
}

func fromNullInt64Raw(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}

func fromNullInt64ToBool(ni sql.NullInt64) bool {
	return ni.Valid && ni.Int64 != 0
}

// Analytics helper: matches UUID v4 format (with or without dashes) or 32-char hex
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$`)

func getSessionID(c *gin.Context) sql.NullString {
	id := c.GetHeader("X-Session-ID")
	if id == "" || !uuidPattern.MatchString(strings.ToLower(id)) {
		return sql.NullString{}
	}
	return sql.NullString{String: id, Valid: true}
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullInt64(n int) sql.NullInt64 {
	if n == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(n), Valid: true}
}

func nullIntPtr(p *int) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*p), Valid: true}
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// NewHandlers creates a new Handlers instance with all dependencies.
func NewHandlers(db *sql.DB, cache *cache.ScheduleCache, generator *generator.Service, queries *store.Queries, searchSvc *search.Service) *Handlers {
	return &Handlers{
		db:        db,
		cache:     cache,
		generator: generator,
		queries:   queries,
		search:    searchSvc,
	}
}

// validateTerm checks if a term exists and sends an error response if not.
// Returns true if the term is valid, false if an error response was sent.
func (h *Handlers) validateTerm(c *gin.Context, term string) bool {
	_, err := h.queries.GetTermByCode(c.Request.Context(), term)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Term not found: " + term})
		return false
	}
	if err != nil {
		slog.Error("Failed to validate term", "term", term, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate term"})
		return false
	}
	return true
}

// Health returns server health status.
func (h *Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// Generate creates schedule combinations for requested courses.
func (h *Handlers) Generate(c *gin.Context) {
	var req generator.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("Invalid generate request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !h.validateTerm(c, req.Term) {
		return
	}

	if err := h.cache.LoadTermIfNeeded(c.Request.Context(), req.Term); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load term: " + err.Error()})
		return
	}

	resp, err := h.generator.Generate(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sessionID := getSessionID(c)
	go func() {
		logID, err := h.queries.LogGeneration(context.Background(), store.LogGenerationParams{
			SessionID:          sessionID,
			Term:               req.Term,
			CoursesCount:       int64(len(req.CourseSpecs)),
			SchedulesGenerated: int64(resp.Stats.TotalGenerated),
			MinCourses:         nullInt64(req.MinCourses),
			MaxCourses:         nullInt64(req.MaxCourses),
			BlockedTimesCount:  int64(len(req.BlockedTimes)),
			DurationMs:         sql.NullFloat64{Float64: resp.Stats.TimeMs, Valid: true},
		})
		if err != nil {
			slog.Warn("Failed to log generation", "error", err)
			return
		}
		for _, spec := range req.CourseSpecs {
			if err := h.queries.LogGenerationCourse(context.Background(), store.LogGenerationCourseParams{
				GenerationLogID: logID,
				Subject:         spec.Subject,
				CourseNumber:     spec.CourseNumber,
				Required:        boolToInt64(spec.Required),
			}); err != nil {
				slog.Warn("Failed to log generation course", "error", err)
			}
		}
	}()

	c.JSON(http.StatusOK, resp.ToResponse())
}

// TermResponse represents a term in the API response.
type TermResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// GetTerms returns available terms for dropdowns.
func (h *Handlers) GetTerms(c *gin.Context) {
	terms, err := h.queries.GetTerms(c.Request.Context())
	if err != nil {
		slog.Error("Failed to fetch terms", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch terms"})
		return
	}

	// Limit to recent terms and clean up names
	limit := min(len(terms), maxTermsReturned)
	result := make([]TermResponse, 0, limit)
	for _, t := range terms[:limit] {
		name := strings.TrimSuffix(t.Description, " (View Only)")
		result = append(result, TermResponse{
			Code: t.Code,
			Name: name,
		})
	}

	current := determineCurrentTerm(terms)

	c.JSON(http.StatusOK, gin.H{
		"terms":   result,
		"current": current,
	})
}

// determineCurrentTerm selects the best default term for the frontend.
// Prefers pre-registration (upcoming term), then active registration, then most recent.
func determineCurrentTerm(terms []*store.Term) string {
	if len(terms) == 0 {
		return ""
	}

	now := time.Now()

	// Find the first term in pre-registration (registration opening soon)
	for _, t := range terms {
		phase := jobs.GetTermPhase(t.Code, now)
		if phase == jobs.PhasePreRegistration {
			return t.Code
		}
	}

	// Fall back to active registration (current term still registering)
	for _, t := range terms {
		phase := jobs.GetTermPhase(t.Code, now)
		if phase == jobs.PhaseActiveRegistration {
			return t.Code
		}
	}

	// If nothing matches, return the most recent term
	return terms[0].Code
}

// SubjectResponse represents a subject in the API response.
type SubjectResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// GetSubjects returns subject list for dropdowns.
// Accepts optional ?term= query param to filter subjects by term.
func (h *Handlers) GetSubjects(c *gin.Context) {
	term := c.Query("term")

	if term != "" && !h.validateTerm(c, term) {
		return
	}

	var subjects []SubjectResponse

	if term != "" {
		rows, err := h.queries.GetSubjectsWithDescriptionsByTerm(c.Request.Context(), term)
		if err != nil {
			slog.Error("Failed to fetch subjects by term", "term", term, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subjects"})
			return
		}
		subjects = make([]SubjectResponse, 0, len(rows))
		for _, row := range rows {
			name := row.Subject
			if row.SubjectDescription.Valid && row.SubjectDescription.String != "" {
				name = row.SubjectDescription.String
			}
			subjects = append(subjects, SubjectResponse{
				Code: row.Subject,
				Name: name,
			})
		}
	} else {
		codes, err := h.queries.GetDistinctSubjects(c.Request.Context())
		if err != nil {
			slog.Error("Failed to fetch distinct subjects", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subjects"})
			return
		}
		subjects = make([]SubjectResponse, 0, len(codes))
		for _, code := range codes {
			subjects = append(subjects, SubjectResponse{
				Code: code,
				Name: code,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"subjects": subjects,
	})
}

// Search searches for courses/sections with filters.
func (h *Handlers) Search(c *gin.Context) {
	var req search.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.search.Search(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, search.ErrNoFilters):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, search.ErrTooManyTokens):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, search.ErrInvalidTerm):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, search.ErrInvalidYear):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, search.ErrWildcardOnly):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, search.ErrFilterTooShort):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			slog.Error("Search failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		}
		return
	}

	sessionID := getSessionID(c)
	scope := "all"
	if req.Term != "" {
		scope = "term"
	} else if req.Year != 0 {
		scope = "year"
	}
	go func() {
		if err := h.queries.LogSearch(context.Background(), store.LogSearchParams{
			SessionID:    sessionID,
			Term:         nullStr(req.Term),
			Scope:        nullStr(scope),
			Subject:      nullStr(req.Subject),
			CourseNumber: nullStr(req.CourseNumber),
			Title:        nullStr(req.Title),
			Instructor:   nullStr(req.Instructor),
			OpenSeats:    boolToInt64(req.OpenSeats),
			MinCredits:   nullIntPtr(req.MinCredits),
			MaxCredits:   nullIntPtr(req.MaxCredits),
			ResultsCount: int64(result.Total),
			DurationMs:   sql.NullFloat64{Float64: result.Stats.TimeMs, Valid: true},
		}); err != nil {
			slog.Warn("Failed to log search", "error", err)
		}
	}()

	c.JSON(http.StatusOK, result)
}

// GetCRN looks up a specific CRN.
// Accepts optional ?term= query param. If not provided, searches the most recent term first.
func (h *Handlers) GetCRN(c *gin.Context) {
	crn := c.Param("crn")
	if crn == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "crn is required"})
		return
	}

	term := c.Query("term")

	// If term is specified, search only that term
	if term != "" {
		if !h.validateTerm(c, term) {
			return
		}
		section, err := h.queries.GetSectionWithInstructorByTermAndCRN(c.Request.Context(), store.GetSectionWithInstructorByTermAndCRNParams{
			Term: term,
			Crn:  crn,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusOK, CRNResponse{Section: nil})
				return
			}
			slog.Error("Failed to lookup CRN", "crn", crn, "term", term, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lookup CRN"})
			return
		}

		meetings, err := h.queries.GetMeetingTimesBySection(c.Request.Context(), section.ID)
		if err != nil {
			slog.Error("Failed to fetch meeting times", "crn", crn, "term", term, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch meeting times"})
			return
		}

		c.JSON(http.StatusOK, CRNResponse{Section: formatSectionResponse(section, meetings)})
		return
	}

	// No term specified: search through terms from newest to oldest
	terms, err := h.queries.GetTerms(c.Request.Context())
	if err != nil {
		slog.Error("Failed to fetch terms for CRN lookup", "crn", crn, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch terms"})
		return
	}

	for _, t := range terms {
		section, err := h.queries.GetSectionWithInstructorByTermAndCRN(c.Request.Context(), store.GetSectionWithInstructorByTermAndCRNParams{
			Term: t.Code,
			Crn:  crn,
		})
		if err == nil {
			meetings, err := h.queries.GetMeetingTimesBySection(c.Request.Context(), section.ID)
			if err != nil {
				slog.Error("Failed to fetch meeting times", "crn", crn, "term", t.Code, "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch meeting times"})
				return
			}
			c.JSON(http.StatusOK, CRNResponse{Section: formatSectionResponse(section, meetings)})
			return
		}
		if !errors.Is(err, sql.ErrNoRows) {
			slog.Error("Failed to lookup CRN across terms", "crn", crn, "currentTerm", t.Code, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lookup CRN"})
			return
		}
	}

	c.JSON(http.StatusOK, CRNResponse{Section: nil})
}

func formatSectionResponse(s *store.GetSectionWithInstructorByTermAndCRNRow, meetings []*store.MeetingTime) *SectionResponse {
	meetingTimes := make([]MeetingTimeInfo, 0, len(meetings))
	for _, m := range meetings {
		meetingTimes = append(meetingTimes, MeetingTimeInfo{
			Days: []bool{
				m.Sunday.Valid && m.Sunday.Int64 != 0,
				m.Monday.Valid && m.Monday.Int64 != 0,
				m.Tuesday.Valid && m.Tuesday.Int64 != 0,
				m.Wednesday.Valid && m.Wednesday.Int64 != 0,
				m.Thursday.Valid && m.Thursday.Int64 != 0,
				m.Friday.Valid && m.Friday.Int64 != 0,
				m.Saturday.Valid && m.Saturday.Int64 != 0,
			},
			StartTime: fromNullString(m.StartTime),
			EndTime:   fromNullString(m.EndTime),
			Building:  fromNullString(m.Building),
			Room:      fromNullString(m.Room),
		})
	}

	return &SectionResponse{
		CRN:            s.Crn,
		Term:           s.Term,
		Subject:        s.Subject,
		CourseNumber:   s.CourseNumber,
		Title:          s.Title,
		Instructor:     fromNullString(s.InstructorName),
		Credits:        fromNullInt64(s.CreditHoursLow),
		Enrollment:     fromNullInt64Raw(s.Enrollment),
		MaxEnrollment:  fromNullInt64Raw(s.MaxEnrollment),
		SeatsAvailable: fromNullInt64Raw(s.SeatsAvailable),
		WaitCount:      fromNullInt64Raw(s.WaitCount),
		IsOpen:         fromNullInt64ToBool(s.IsOpen),
		MeetingTimes:   meetingTimes,
	}
}

// GetCourse returns course info and all sections for a course.
func (h *Handlers) GetCourse(c *gin.Context) {
	subject := strings.ToUpper(strings.TrimSpace(c.Param("subject")))
	courseNumber := strings.ToUpper(strings.TrimSpace(c.Param("courseNumber")))
	term := strings.TrimSpace(c.Query("term"))

	if subject == "" || courseNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subject and courseNumber are required"})
		return
	}
	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "term query parameter is required"})
		return
	}
	if !h.validateTerm(c, term) {
		return
	}

	sections, err := h.queries.GetSectionsWithInstructorByCourse(c.Request.Context(), store.GetSectionsWithInstructorByCourseParams{
		Term:         term,
		Subject:      subject,
		CourseNumber: courseNumber,
	})
	if err != nil {
		slog.Error("Failed to get course sections", "term", term, "subject", subject, "courseNumber", courseNumber, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get course"})
		return
	}

	if len(sections) == 0 {
		c.JSON(http.StatusOK, CourseResponse{Course: nil})
		return
	}

	// Build sections response
	sectionList := make([]CourseSectionInfo, 0, len(sections))
	for _, s := range sections {
		sectionList = append(sectionList, CourseSectionInfo{
			CRN:            s.Crn,
			Instructor:     fromNullString(s.InstructorName),
			Enrollment:     fromNullInt64Raw(s.Enrollment),
			MaxEnrollment:  fromNullInt64Raw(s.MaxEnrollment),
			SeatsAvailable: fromNullInt64Raw(s.SeatsAvailable),
			WaitCount:      fromNullInt64Raw(s.WaitCount),
			IsOpen:         fromNullInt64ToBool(s.IsOpen),
		})
	}

	first := sections[0]
	c.JSON(http.StatusOK, CourseResponse{
		Course: &CourseInfo{
			Subject:      first.Subject,
			CourseNumber: first.CourseNumber,
			Title:        first.Title,
			Credits:      fromNullInt64(first.CreditHoursLow),
		},
		Sections:     sectionList,
		SectionCount: len(sections),
	})
}

const maxValidateCourses = 20

// ValidateCoursesRequest is the request body for batch course validation.
type ValidateCoursesRequest struct {
	Term    string `json:"term" binding:"required"`
	Courses []struct {
		Subject      string `json:"subject" binding:"required"`
		CourseNumber string `json:"courseNumber" binding:"required"`
	} `json:"courses" binding:"required"`
}

// CourseValidationResult represents the validation result for a single course.
type CourseValidationResult struct {
	Subject      string `json:"subject"`
	CourseNumber string `json:"courseNumber"`
	Exists       bool   `json:"exists"`
	Title        string `json:"title,omitempty"`
	SectionCount int    `json:"sectionCount,omitempty"`
}

// ValidateCoursesResponse is the response for batch course validation.
type ValidateCoursesResponse struct {
	Results []CourseValidationResult `json:"results"`
}

// ValidateCourses checks if courses exist in a given term using a single batch query.
func (h *Handlers) ValidateCourses(c *gin.Context) {
	var req ValidateCoursesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("Invalid validate courses request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !h.validateTerm(c, req.Term) {
		return
	}

	if len(req.Courses) == 0 {
		c.JSON(http.StatusOK, ValidateCoursesResponse{Results: []CourseValidationResult{}})
		return
	}

	if len(req.Courses) > maxValidateCourses {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 20 courses per request"})
		return
	}

	// Normalize inputs and build query
	type courseKey struct {
		subject      string
		courseNumber string
	}
	courses := make([]courseKey, 0, len(req.Courses))
	for _, course := range req.Courses {
		courses = append(courses, courseKey{
			subject:      strings.ToUpper(strings.TrimSpace(course.Subject)),
			courseNumber: strings.ToUpper(strings.TrimSpace(course.CourseNumber)),
		})
	}

	// Build batch query with OR conditions
	// Query: SELECT subject, course_number, COUNT(*) as section_count, MAX(title) as title
	//        FROM sections WHERE term = ? AND ((subject = ? AND course_number = ?) OR ...)
	//        GROUP BY subject, course_number
	args := make([]any, 0, 1+len(courses)*2)
	args = append(args, req.Term)

	conditions := make([]string, 0, len(courses))
	for _, ck := range courses {
		conditions = append(conditions, "(subject = ? AND course_number = ?)")
		args = append(args, ck.subject, ck.courseNumber)
	}

	query := "SELECT subject, course_number, COUNT(*) as section_count, MAX(title) as title " +
		"FROM sections WHERE term = ? AND (" + strings.Join(conditions, " OR ") + ") " +
		"GROUP BY subject, course_number"

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		slog.Error("Failed to validate courses", "term", req.Term, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate courses"})
		return
	}
	defer rows.Close()

	// Collect results from query into a map
	found := make(map[courseKey]CourseValidationResult)
	for rows.Next() {
		var subject, courseNumber, title string
		var sectionCount int
		if err := rows.Scan(&subject, &courseNumber, &sectionCount, &title); err != nil {
			slog.Error("Failed to scan validation row", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate courses"})
			return
		}
		found[courseKey{subject: subject, courseNumber: courseNumber}] = CourseValidationResult{
			Subject:      subject,
			CourseNumber: courseNumber,
			Exists:       true,
			Title:        title,
			SectionCount: sectionCount,
		}
	}
	if err := rows.Err(); err != nil {
		slog.Error("Failed to iterate validation rows", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate courses"})
		return
	}

	// Build response in request order, marking missing courses as not existing
	results := make([]CourseValidationResult, 0, len(courses))
	for _, ck := range courses {
		if result, ok := found[ck]; ok {
			results = append(results, result)
		} else {
			results = append(results, CourseValidationResult{
				Subject:      ck.subject,
				CourseNumber: ck.courseNumber,
				Exists:       false,
			})
		}
	}

	c.JSON(http.StatusOK, ValidateCoursesResponse{Results: results})
}

// AnnouncementResponse wraps the announcement for JSON serialization.
type AnnouncementResponse struct {
	Announcement *AnnouncementInfo `json:"announcement"`
}

// AnnouncementInfo is the public shape of an announcement.
type AnnouncementInfo struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Type  string `json:"type"`
}

// GetAnnouncement returns the current active announcement, if any.
func (h *Handlers) GetAnnouncement(c *gin.Context) {
	row, err := h.queries.GetActiveAnnouncement(c.Request.Context())
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusOK, AnnouncementResponse{Announcement: nil})
		return
	}
	if err != nil {
		slog.Error("Failed to fetch announcement", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch announcement"})
		return
	}

	c.JSON(http.StatusOK, AnnouncementResponse{
		Announcement: &AnnouncementInfo{
			ID:    row.ID,
			Title: row.Title,
			Body:  row.Body,
			Type:  row.Type,
		},
	})
}

const maxFeedbackLength = 1000

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

// FeedbackRequest is the request body for submitting feedback.
type FeedbackRequest struct {
	Message string `json:"message" binding:"required"`
}

// SubmitFeedback stores user feedback in the database.
func (h *Handlers) SubmitFeedback(c *gin.Context) {
	var req FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	// Strip HTML tags, then validate
	msg := htmlTagPattern.ReplaceAllString(req.Message, "")
	msg = strings.TrimSpace(msg)

	if msg == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message cannot be empty"})
		return
	}
	if len(msg) > maxFeedbackLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message exceeds 1000 character limit"})
		return
	}

	sessionID := getSessionID(c)
	go func() {
		if err := h.queries.InsertFeedback(context.Background(), store.InsertFeedbackParams{
			SessionID: sessionID,
			Message:   msg,
		}); err != nil {
			slog.Warn("Failed to store feedback", "error", err)
		}
	}()

	c.Status(http.StatusNoContent)
}
