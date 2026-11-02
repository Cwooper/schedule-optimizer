package api

import (
	"context"
	"database/sql"
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
				CourseNumber:    spec.CourseNumber,
				Required:        boolToInt64(spec.Required),
			}); err != nil {
				slog.Warn("Failed to log generation course", "error", err)
			}
		}
	}()

	c.JSON(http.StatusOK, resp.ToResponse())
}
