package api

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"schedule-optimizer/internal/calendar"
	"schedule-optimizer/internal/meeting"
	"schedule-optimizer/internal/search"
	"schedule-optimizer/internal/store"
	"strings"

	"github.com/gin-gonic/gin"
)

type SectionResponse struct {
	CRN            string          `json:"crn"`
	Term           string          `json:"term"`
	Subject        string          `json:"subject"`
	CourseNumber   string          `json:"courseNumber"`
	Title          string          `json:"title"`
	Instructor     string          `json:"instructor"`
	Credits        int             `json:"credits"`
	Enrollment     int64           `json:"enrollment"`
	MaxEnrollment  int64           `json:"maxEnrollment"`
	SeatsAvailable int64           `json:"seatsAvailable"`
	WaitCount      int64           `json:"waitCount"`
	IsOpen         bool            `json:"isOpen"`
	MeetingTimes   []meeting.Time  `json:"meetingTimes"`
	Final          *calendar.Final `json:"final,omitempty"`
	GPA            float64         `json:"gpa,omitempty"`
	GPASource      string          `json:"gpaSource,omitempty"`
	PassRate       *float64        `json:"passRate,omitempty"`
}

type CRNResponse struct {
	Section *SectionResponse `json:"section"`
}

type CourseInfo struct {
	Subject      string   `json:"subject"`
	CourseNumber string   `json:"courseNumber"`
	Title        string   `json:"title"`
	Credits      int      `json:"credits"`
	GPA          float64  `json:"gpa,omitempty"`
	PassRate     *float64 `json:"passRate,omitempty"`
}

type CourseSectionInfo struct {
	CRN            string   `json:"crn"`
	Instructor     string   `json:"instructor"`
	Enrollment     int64    `json:"enrollment"`
	MaxEnrollment  int64    `json:"maxEnrollment"`
	SeatsAvailable int64    `json:"seatsAvailable"`
	WaitCount      int64    `json:"waitCount"`
	IsOpen         bool     `json:"isOpen"`
	GPA            float64  `json:"gpa,omitempty"`
	GPASource      string   `json:"gpaSource,omitempty"`
	PassRate       *float64 `json:"passRate,omitempty"`
}

type CourseResponse struct {
	Course       *CourseInfo         `json:"course"`
	Sections     []CourseSectionInfo `json:"sections,omitempty"`
	SectionCount int                 `json:"sectionCount,omitempty"`
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

// GetCRN looks up a specific CRN. Accepts optional ?term= query param; without
// it, terms are searched newest first.
func (h *Handlers) GetCRN(c *gin.Context) {
	crn := c.Param("crn")
	if crn == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "crn is required"})
		return
	}

	var terms []string
	if term := c.Query("term"); term != "" {
		if !h.validateTerm(c, term) {
			return
		}
		terms = []string{term}
	} else {
		all, err := h.queries.GetTerms(c.Request.Context())
		if err != nil {
			slog.Error("Failed to fetch terms for CRN lookup", "crn", crn, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch terms"})
			return
		}
		for _, t := range all {
			terms = append(terms, t.Code)
		}
	}

	for _, term := range terms {
		section, err := h.lookupSection(c.Request.Context(), term, crn)
		if err != nil {
			slog.Error("Failed to lookup CRN", "crn", crn, "term", term, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lookup CRN"})
			return
		}
		if section != nil {
			c.JSON(http.StatusOK, CRNResponse{Section: section})
			return
		}
	}
	c.JSON(http.StatusOK, CRNResponse{Section: nil})
}

// lookupSection returns nil, nil when the term has no such CRN.
func (h *Handlers) lookupSection(ctx context.Context, term, crn string) (*SectionResponse, error) {
	s, err := h.queries.GetSectionWithInstructorByTermAndCRN(ctx, store.GetSectionWithInstructorByTermAndCRNParams{
		Term: term,
		Crn:  crn,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rows, err := h.queries.GetMeetingTimesBySection(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	finals, _ := calendar.SectionFinals(ctx, h.queries, []string{term})

	meetings := meeting.FromStoreAll(rows)
	resp := &SectionResponse{
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
		MeetingTimes:   meetings,
		Final:          finals[term].FinalFor(s.ScheduleType.String, meetings),
	}
	if h.grades != nil && h.grades.IsLoaded() {
		resp.GPA, resp.PassRate, resp.GPASource = h.grades.LookupSectionGPA(resp.Subject, resp.CourseNumber, resp.Instructor)
	}
	return resp, nil
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
		si := CourseSectionInfo{
			CRN:            s.Crn,
			Instructor:     fromNullString(s.InstructorName),
			Enrollment:     fromNullInt64Raw(s.Enrollment),
			MaxEnrollment:  fromNullInt64Raw(s.MaxEnrollment),
			SeatsAvailable: fromNullInt64Raw(s.SeatsAvailable),
			WaitCount:      fromNullInt64Raw(s.WaitCount),
			IsOpen:         fromNullInt64ToBool(s.IsOpen),
		}
		if h.grades != nil && h.grades.IsLoaded() {
			si.GPA, si.PassRate, si.GPASource = h.grades.LookupSectionGPA(subject, courseNumber, fromNullString(s.InstructorName))
		}
		sectionList = append(sectionList, si)
	}

	first := sections[0]
	courseInfo := &CourseInfo{
		Subject:      first.Subject,
		CourseNumber: first.CourseNumber,
		Title:        first.Title,
		Credits:      fromNullInt64(first.CreditHoursLow),
	}
	if h.grades != nil && h.grades.IsLoaded() {
		courseInfo.GPA, courseInfo.PassRate, _ = h.grades.LookupCourseGPA(subject, courseNumber)
	}
	c.JSON(http.StatusOK, CourseResponse{
		Course:       courseInfo,
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
