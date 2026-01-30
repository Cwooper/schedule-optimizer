package jobs

import (
	"context"
	"time"

	"schedule-optimizer/internal/config"
	"schedule-optimizer/internal/scraper"
	"schedule-optimizer/internal/store"
)

// Setup creates and starts the jobs service if enabled in config.
// Returns nil if jobs are disabled. The context controls job lifecycle.
func Setup(ctx context.Context, cfg *config.Config, queries *store.Queries) *Service {
	if !cfg.JobsEnabled {
		return nil
	}

	sc, err := scraper.NewScraper(queries, cfg.ScraperConcurrency)
	if err != nil {
		// Log and continue without jobs rather than crashing
		return nil
	}

	pastTermJob := NewPastTermBackfillJob(queries, sc, cfg.PastTermYears)
	activeJob := NewActiveScrapeJob(queries, sc, cfg.PastTermYears, cfg.ActiveScrapeHours)
	dailyJob := NewDailyScrapeJob(queries, sc, cfg.PastTermYears, cfg.DailyScrapeHour)
	bootstrapJob := NewBootstrapJob(sc, []Job{pastTermJob, activeJob, dailyJob})

	service := NewService(time.Minute)
	service.Register(bootstrapJob)
	service.Register(pastTermJob)
	service.Register(activeJob)
	service.Register(dailyJob)

	go service.Start(ctx)

	return service
}
