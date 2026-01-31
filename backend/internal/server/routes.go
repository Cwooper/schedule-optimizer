package server

import (
	"schedule-optimizer/internal/api"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up all API routes.
func RegisterRoutes(r *gin.Engine, h *api.Handlers) {
	api := r.Group("/api")
	{
		api.GET("/health", h.Health)
		api.GET("/terms", h.GetTerms)
		api.GET("/subjects", h.GetSubjects)
		api.GET("/course/validate", h.ValidateCourse)
		api.GET("/courses", h.SearchCourses)
		api.GET("/crn/:crn", h.GetCRN)
		api.POST("/generate", h.Generate)
	}
}
