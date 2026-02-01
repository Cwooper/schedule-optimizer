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
		api.GET("/course/:subject/:courseNumber", h.GetCourse)
		api.GET("/courses", h.SearchCourses)
		api.GET("/crn/:crn", h.GetCRN)
		api.POST("/courses/validate", h.ValidateCourses)
		api.POST("/generate", h.Generate)
	}
}
