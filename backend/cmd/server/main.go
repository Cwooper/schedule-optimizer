package main

import (
	"fmt"
	"log/slog"

	"schedule-optimizer/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	slog.Info("Starting server...")

	cfg := config.Load()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}
	r.Use(cors.New(corsConfig))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	slog.Info(fmt.Sprintf("Server is running on port %s", cfg.Port))
	if err := r.Run(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		slog.Error("Failed to run server", slog.String("error", err.Error()))
	}
}
