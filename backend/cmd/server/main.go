package main

import (
	"fmt"
	"log/slog"
	"os"

	"schedule-optimizer/internal/config"
	"schedule-optimizer/internal/db"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// TODO: split this main function out into starting a server that is
	// initialized in internal, this function should then have <50 lines
	// of setting up middleware, db, cache, routes, etc. all in their
	// own functions/packages/files.
	slog.Info("Starting server...")

	cfg := config.Load()

	// Initialize database connection
	database, err := db.Open(cfg.DatabasePath)
	if err != nil {
		slog.Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

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
