package main

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func main() {
	slog.Info("Starting server...")

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	port := 8080
	slog.Info(fmt.Sprintf("Server is running on port %d", port))
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		slog.Error("Failed to run server", slog.String("error", err.Error()))
	}
}
