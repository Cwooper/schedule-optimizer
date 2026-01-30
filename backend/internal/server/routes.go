package server

import (
	"schedule-optimizer/internal/api"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up all API routes.
func RegisterRoutes(r *gin.Engine, h *api.Handlers) {
	r.GET("/health", h.Health)
	r.POST("/generate", h.Generate)
}
