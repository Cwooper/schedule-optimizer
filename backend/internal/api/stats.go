package api

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"schedule-optimizer/internal/stats"

	"github.com/gin-gonic/gin"
)

// GetStatsCourses lists every course with grade data, for the Statistics
// tab's pickers.
func (h *Handlers) GetStatsCourses(c *gin.Context) {
	resp, err := h.stats.Courses()
	respondStats(c, resp, err, "list courses")
}

func (h *Handlers) GetStatsCourse(c *gin.Context) {
	subject := strings.ToUpper(strings.TrimSpace(c.Param("subject")))
	number := strings.ToUpper(strings.TrimSpace(c.Param("number")))
	resp, err := h.stats.Course(c.Request.Context(), subject, number)
	respondStats(c, resp, err, "build course stats")
}

func (h *Handlers) GetStatsSection(c *gin.Context) {
	resp, err := h.stats.Section(c.Request.Context(), c.Param("term"), c.Param("crn"))
	respondStats(c, resp, err, "build section stats")
}

func respondStats(c *gin.Context, resp any, err error, action string) {
	switch {
	case err == nil:
		c.JSON(http.StatusOK, resp)
	case errors.Is(err, stats.ErrNotLoaded):
		// The client shows this message as is.
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Statistics are still loading"})
	case errors.Is(err, stats.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		slog.Error("Failed to "+action, "path", c.Request.URL.Path, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to " + action})
	}
}
