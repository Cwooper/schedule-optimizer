package main

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

	"schedule-optimizer/internal/config"
	"schedule-optimizer/internal/db"
	"schedule-optimizer/internal/jobs"
	"schedule-optimizer/internal/scraper"
	"schedule-optimizer/internal/store"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lmittmann/tint"
)

func main() {
	// TODO: split this main function out into starting a server that is
	// initialized in internal, this function should then have <50 lines
	// of setting up middleware, db, cache, routes, etc. all in their
	// own functions/packages/files.
	slog.Info("Starting server...")

	cfg := config.Load()

	if cfg.Environment != "production" {
		slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		})))
	}

	database, err := db.Open(cfg.DatabasePath)
	if err != nil {
		slog.Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	queries := store.New(database)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var jobsService *jobs.Service
	if cfg.JobsEnabled {
		sc, err := scraper.NewScraper(queries, cfg.ScraperConcurrency)
		if err != nil {
			slog.Error("Failed to create scraper", "error", err)
			os.Exit(1)
		}

		pastTermJob := jobs.NewPastTermBackfillJob(queries, sc, cfg.PastTermYears)
		activeJob := jobs.NewActiveScrapeJob(queries, sc, cfg.PastTermYears, cfg.ActiveScrapeHours)
		dailyJob := jobs.NewDailyScrapeJob(queries, sc, cfg.PastTermYears, cfg.DailyScrapeHour)
		bootstrapJob := jobs.NewBootstrapJob(sc, []jobs.Job{pastTermJob, activeJob, dailyJob})

		jobsService = jobs.NewService(time.Minute)
		jobsService.Register(bootstrapJob)
		jobsService.Register(pastTermJob)
		jobsService.Register(activeJob)
		jobsService.Register(dailyJob)
		go jobsService.Start(ctx)
	}

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
