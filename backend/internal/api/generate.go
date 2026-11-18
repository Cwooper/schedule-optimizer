package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"schedule-optimizer/internal/generator"
	"schedule-optimizer/internal/store"

	"github.com/gin-gonic/gin"
)

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

	term, err := h.cache.Term(c.Request.Context(), req.Term)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load term: " + err.Error()})
		return
	}

	resp, err := h.generator.Generate(c.Request.Context(), term, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sessionID := getSessionID(c)
	var prefs sql.NullString
	if req.Preferences != nil {
		// Marshaling a struct of ints and strings cannot fail.
		b, _ := json.Marshal(req.Preferences)
		prefs = sql.NullString{String: string(b), Valid: true}
	}
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
			Preferences:        prefs,
		})
		if err != nil {
			slog.Warn("Failed to log generation", "error", err)
			return
		}
		for _, spec := range req.CourseSpecs {
			if err := h.queries.LogGenerationCourse(context.Background(), store.LogGenerationCourseParams{
				GenerationLogID: logID,
				Subject:         spec.Subject,
				CourseNumber:    spec.CourseNumber,
				Required:        boolToInt64(spec.Required),
			}); err != nil {
				slog.Warn("Failed to log generation course", "error", err)
			}
		}
	}()

	c.JSON(http.StatusOK, resp.ToResponse())
}
