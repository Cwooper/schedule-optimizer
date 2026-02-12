package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"schedule-optimizer/internal/api"
	"schedule-optimizer/internal/cache"
	"schedule-optimizer/internal/config"
	"schedule-optimizer/internal/db"
	"schedule-optimizer/internal/generator"
	"schedule-optimizer/internal/jobs"
	"schedule-optimizer/internal/search"
	"schedule-optimizer/internal/stats"
	"schedule-optimizer/internal/stats/grades"
	"schedule-optimizer/internal/store"

	"github.com/gin-gonic/gin"
)

// Run initializes and starts the server, blocking until shutdown.
func Run() {
	cfg := config.Load()
	config.SetupLogging(cfg)

	slog.Info("Starting server...")

	database, err := db.Open(cfg.DatabasePath)
	if err != nil {
		slog.Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	queries := store.New(database)

	// Validate that migrations have been applied
	if _, err := queries.CheckSchemaExists(context.Background()); err != nil {
		slog.Error("Database schema not initialized - run 'make migrate-up' first", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	gradeService := grades.NewService(database, queries, cfg.GradeDataPath)
	_ = stats.NewService(gradeService) // TODO: pass to handlers when stats endpoints are added

	jobsService := jobs.Setup(ctx, cfg, queries, gradeService)

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	SetupMiddleware(r, cfg)

	scheduleCache := cache.NewScheduleCache(queries, gradeService)
	generatorService := generator.NewService(scheduleCache, queries)
	searchService := search.NewService(database, queries, gradeService)
	handlers := api.NewHandlers(database, scheduleCache, generatorService, queries, searchService)

	RegisterRoutes(r, handlers)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	go func() {
		slog.Info("Server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown error", "error", err)
	}

	if jobsService != nil {
		jobsService.Stop()
	}

	slog.Info("Server stopped")
}
