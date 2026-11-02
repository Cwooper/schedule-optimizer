package api

import (
	"log/slog"
	"net/http"
	"schedule-optimizer/internal/calendar"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// TermResponse is a term with its phase and dates (YYYY-MM-DD, end
// inclusive), estimated when the registrar has not published them.
type TermResponse struct {
	Code              string `json:"code"`
	Name              string `json:"name"`
	Phase             string `json:"phase"`
	Start             string `json:"start"`
	End               string `json:"end"`
	RegistrationOpens string `json:"registrationOpens"`
	Estimated         bool   `json:"estimated"`
}

// GetTerms returns available terms for dropdowns.
func (h *Handlers) GetTerms(c *gin.Context) {
	terms, err := h.queries.GetTerms(c.Request.Context())
	if err != nil {
		slog.Error("Failed to fetch terms", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch terms"})
		return
	}

	schedule, err := calendar.LoadSchedule(c.Request.Context(), h.queries)
	if err != nil {
		// Estimated dates still give a sensible default term.
		slog.Error("Failed to load term schedule", "error", err)
	}
	now := time.Now()

	limit := min(len(terms), maxTermsReturned)
	result := make([]TermResponse, 0, limit)
	for _, t := range terms[:limit] {
		r := TermResponse{
			Code: t.Code,
			Name: strings.TrimSuffix(t.Description, " (View Only)"),
		}
		if d, ok := schedule.Dates(t.Code); ok {
			r.Phase = d.Phase(now).String()
			r.Start = d.Start.Format(time.DateOnly)
			r.End = d.End.AddDate(0, 0, -1).Format(time.DateOnly)
			r.RegistrationOpens = d.RegistrationOpens.Format(time.DateOnly)
			r.Estimated = d.Estimated
		}
		result = append(result, r)
	}

	codes := make([]string, len(terms))
	for i, t := range terms {
		codes[i] = t.Code
	}

	c.JSON(http.StatusOK, gin.H{
		"terms":   result,
		"current": schedule.DefaultTerm(codes, now),
	})
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
