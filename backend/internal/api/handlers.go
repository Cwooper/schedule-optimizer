package api

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"schedule-optimizer/internal/cache"
	"schedule-optimizer/internal/generator"
	"schedule-optimizer/internal/jobs"
	"schedule-optimizer/internal/store"

	"github.com/gin-gonic/gin"
)

const maxTermsReturned = 8 // ~2 years of quarters

// Handlers contains all HTTP handler dependencies.
type Handlers struct {
	cache     *cache.ScheduleCache
	generator *generator.Service
	queries   *store.Queries
}

// Response types for type-safe JSON serialization

type SectionResponse struct {
	CRN          string `json:"crn"`
	Term         string `json:"term"`
	Subject      string `json:"subject"`
	CourseNumber string `json:"courseNumber"`
	Title        string `json:"title"`
	Instructor   string `json:"instructor"`
	Credits      int    `json:"credits"`
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

// NewHandlers creates a new Handlers instance with all dependencies.
func NewHandlers(cache *cache.ScheduleCache, generator *generator.Service, queries *store.Queries) *Handlers {
	return &Handlers{
		cache:     cache,
		generator: generator,
		queries:   queries,
	}
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

	if err := h.cache.LoadTermIfNeeded(c.Request.Context(), req.Term); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load term: " + err.Error()})
		return
	}

	resp, err := h.generator.Generate(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
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
// Prefers terms in active registration phase, falls back to current term by date.
func determineCurrentTerm(terms []*store.Term) string {
	if len(terms) == 0 {
		return ""
	}

	now := time.Now()

	// Find the first term in active registration (terms are sorted DESC by code)
	for _, t := range terms {
		phase := jobs.GetTermPhase(t.Code, now)
		if phase == jobs.PhaseActiveRegistration {
			return t.Code
		}
	}

	// Fall back to current term by calendar date
	currentByDate := jobs.CurrentTermCode(now)
	for _, t := range terms {
		if t.Code == currentByDate {
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

// SearchCourses searches for courses/sections with filters.
func (h *Handlers) SearchCourses(c *gin.Context) {
	// TODO: implement with real data from store
	c.JSON(http.StatusOK, gin.H{
		"sections": []gin.H{
			{
				"crn":           "41328",
				"term":          "202520",
				"subject":       "CSCI",
				"courseNumber":  "247",
				"title":         "Computer Systems",
				"credits":       5,
				"instructor":    "See-Mong Tan",
				"maxEnrollment": 30,
				"enrollment":    28,
				"seatsAvailable": 2,
				"waitCount":     0,
				"isOpen":        true,
				"meetingTimes": []gin.H{
					{"days": []bool{false, true, false, true, false, true, false}, "startTime": "0900", "endTime": "0950", "building": "CF", "room": "225"},
				},
			},
		},
		"total":   1,
		"hasMore": false,
	})
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

		c.JSON(http.StatusOK, CRNResponse{Section: formatSectionResponse(section)})
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
			c.JSON(http.StatusOK, CRNResponse{Section: formatSectionResponse(section)})
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

func formatSectionResponse(s *store.GetSectionWithInstructorByTermAndCRNRow) *SectionResponse {
	return &SectionResponse{
		CRN:          s.Crn,
		Term:         s.Term,
		Subject:      s.Subject,
		CourseNumber: s.CourseNumber,
		Title:        s.Title,
		Instructor:   fromNullString(s.InstructorName),
		Credits:      fromNullInt64(s.CreditHoursLow),
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
