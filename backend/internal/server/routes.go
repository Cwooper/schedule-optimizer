package server

import (
	"schedule-optimizer/internal/api"
	"schedule-optimizer/internal/static"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up all API routes and static file serving.
func RegisterRoutes(r *gin.Engine, h *api.Handlers) {
	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/health", h.Health)
		apiGroup.GET("/terms", h.GetTerms)
		apiGroup.GET("/subjects", h.GetSubjects)
		apiGroup.GET("/course/:subject/:courseNumber", h.GetCourse)
		apiGroup.GET("/search", h.Search)
		apiGroup.GET("/crn/:crn", h.GetCRN)
		apiGroup.POST("/courses/validate", h.ValidateCourses)
		apiGroup.POST("/generate", h.Generate)
	}

	// Serve static files (compiled frontend)
	r.NoRoute(gin.WrapH(static.Handler()))
}
