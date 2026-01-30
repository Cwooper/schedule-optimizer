package api

import (
	"net/http"

	"schedule-optimizer/internal/cache"
	"schedule-optimizer/internal/generator"

	"github.com/gin-gonic/gin"
)

// Handlers contains all HTTP handler dependencies.
type Handlers struct {
	cache     *cache.ScheduleCache
	generator *generator.Service
}

// NewHandlers creates a new Handlers instance with all dependencies.
func NewHandlers(cache *cache.ScheduleCache, generator *generator.Service) *Handlers {
	return &Handlers{
		cache:     cache,
		generator: generator,
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
