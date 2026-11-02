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
func Setup(ctx context.Context, cfg *config.Config, db *sql.DB, gradeService *grades.Service) *Service {
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

	// The service runs due jobs in registration order, and every job is due
	// on its first check, so startup refreshes the term list and calendar
	// before any scrape job reads them.
	service := NewService(time.Minute)
	service.Register(NewGradeImportJob(gradeService))
	service.Register(NewCalendarJob(sc, calendar.NewService(db)))
	service.Register(NewPastTermBackfillJob(queries, sc, cfg.PastTermYears))
	service.Register(NewActiveScrapeJob(queries, sc, cfg.PastTermYears, cfg.ActiveScrapeHours))
	service.Register(NewDailyScrapeJob(queries, sc, cfg.PastTermYears, cfg.DailyScrapeHour))

	go service.Start(ctx)

	return service
}
