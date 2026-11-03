package jobs

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"schedule-optimizer/internal/calendar"
	"schedule-optimizer/internal/config"
	"schedule-optimizer/internal/scraper"
	"schedule-optimizer/internal/stats/grades"
	"schedule-optimizer/internal/store"
)

// Setup creates and starts the jobs service if enabled in config.
// Returns nil if jobs are disabled. The context controls job lifecycle.
func Setup(ctx context.Context, cfg *config.Config, db *sql.DB, gradeService *grades.Service, cache CacheInvalidator) *Service {
	if !cfg.JobsEnabled {
		return nil
	}

	queries := store.New(db)
	sc, err := scraper.NewScraper(queries, cfg.ScraperConcurrency)
	if err != nil {
		// Serve without background scraping rather than crash.
		slog.Error("Failed to create scraper; jobs disabled", "error", err)
		return nil
	}

	// The service runs due jobs in registration order, and the calendar job
	// is due on startup and at the daily hour, so the term list and calendar
	// are fresh before the scrape job reads them at those checks.
	service := NewService(time.Minute)
	service.Register(NewGradeImportJob(gradeService, cache))
	service.Register(NewCalendarJob(sc, calendar.NewService(db), cache, cfg.DailyScrapeHour))
	service.Register(NewScrapeJob(queries, sc, cache, cfg.PastTermYears, cfg.ActiveScrapeHours, cfg.DailyScrapeHour))

	go service.Start(ctx)

	return service
}
