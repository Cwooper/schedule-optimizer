package api

import (
	"log/slog"
	"net/http"

	"schedule-optimizer/internal/calendar"

	"github.com/gin-gonic/gin"
)

// GetCalendar returns a term's registration phases, holidays, deadlines, and
// finals week.
func (h *Handlers) GetCalendar(c *gin.Context) {
	term := c.Param("term")
	if !h.validateTerm(c, term) {
		return
	}
	cal, ok, err := calendar.LoadTerm(c.Request.Context(), h.queries, term)
	if err != nil {
		slog.Error("Failed to load term calendar", "term", term, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load calendar"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Term not found: " + term})
		return
	}
	c.JSON(http.StatusOK, cal)
}
