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

// GetTerms returns available terms for dropdowns.
func (h *Handlers) GetTerms(c *gin.Context) {
	// TODO: implement with real data from store
	c.JSON(http.StatusOK, gin.H{
		"terms": []gin.H{
			{"code": "202620", "name": "Summer 2026"},
			{"code": "202610", "name": "Spring 2026"},
			{"code": "202540", "name": "Fall 2025"},
			{"code": "202530", "name": "Summer 2025"},
			{"code": "202520", "name": "Winter 2025"},
		},
		"current": "202520",
	})
}

// GetSubjects returns subject list for dropdowns.
func (h *Handlers) GetSubjects(c *gin.Context) {
	// TODO: implement with real data from store
	c.JSON(http.StatusOK, gin.H{
		"subjects": []gin.H{
			{"code": "CSCI", "name": "Computer Science"},
			{"code": "MATH", "name": "Mathematics"},
			{"code": "ENG", "name": "English"},
			{"code": "PHYS", "name": "Physics"},
		},
	})
}

// ValidateCourse checks if a course exists for a given term.
func (h *Handlers) ValidateCourse(c *gin.Context) {
	// TODO: implement with real data from store
	term := c.Query("term")
	subject := c.Query("subject")
	courseNumber := c.Query("courseNumber")

	if term == "" || subject == "" || courseNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "term, subject, and courseNumber are required"})
		return
	}

	// Stub: pretend course exists
	c.JSON(http.StatusOK, gin.H{
		"exists":       true,
		"title":        "Stub Course Title",
		"sectionCount": 3,
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
func (h *Handlers) GetCRN(c *gin.Context) {
	crn := c.Param("crn")

	// TODO: implement with real data from store
	if crn == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "crn is required"})
		return
	}

	// Stub: pretend CRN exists
	c.JSON(http.StatusOK, gin.H{
		"section": gin.H{
			"crn":          crn,
			"term":         "202520",
			"subject":      "CSCI",
			"courseNumber": "247",
			"title":        "Computer Systems",
			"instructor":   "See-Mong Tan",
			"credits":      5,
		},
	})
}
