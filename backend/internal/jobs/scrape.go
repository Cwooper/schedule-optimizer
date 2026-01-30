package jobs

import (
	"context"
	"log/slog"
	"time"

	"schedule-optimizer/internal/scraper"
	"schedule-optimizer/internal/store"
)

// scrapeTolerance accounts for job execution time when checking scrape intervals.
// Without this buffer, a job that takes N minutes would cause the next interval
// check to fail by N minutes, potentially skipping every other cycle.
const scrapeTolerance = 30 * time.Minute

// BootstrapJob initializes term data and runs any other jobs that need to catch up.
// Runs once on startup to ensure the system has current data immediately.
type BootstrapJob struct {
	scraper   *scraper.Scraper
	otherJobs []Job
	hasRun    bool
}

func NewBootstrapJob(scraper *scraper.Scraper, otherJobs []Job) *BootstrapJob {
	return &BootstrapJob{
		scraper:   scraper,
		otherJobs: otherJobs,
	}
}

func (j *BootstrapJob) Name() string { return "bootstrap" }

func (j *BootstrapJob) ShouldRun(now time.Time) bool {
	return !j.hasRun
}

func (j *BootstrapJob) Run(ctx context.Context, now time.Time) error {
	j.hasRun = true

	if _, err := j.scraper.ScrapeTerms(ctx); err != nil {
		slog.Error("Failed to fetch available terms", "error", err)
	}

	for _, job := range j.otherJobs {
		if job.ShouldRun(now) {
			slog.Info("Bootstrap triggering job", "job", job.Name())
			if err := job.Run(ctx, now); err != nil {
				slog.Error("Bootstrap job failed", "job", job.Name(), "error", err)
			}
		}
	}

	return nil
}

// PastTermBackfillJob scrapes past terms that have never been scraped.
// Runs once on startup to backfill historical data.
type PastTermBackfillJob struct {
	queries       *store.Queries
	scraper       *scraper.Scraper
	pastTermYears int
	hasRun        bool
}

func NewPastTermBackfillJob(queries *store.Queries, scraper *scraper.Scraper, pastTermYears int) *PastTermBackfillJob {
	return &PastTermBackfillJob{
		queries:       queries,
		scraper:       scraper,
		pastTermYears: pastTermYears,
	}
}

func (j *PastTermBackfillJob) Name() string { return "past-term-backfill" }

func (j *PastTermBackfillJob) ShouldRun(now time.Time) bool {
	return !j.hasRun
}

func (j *PastTermBackfillJob) Run(ctx context.Context, now time.Time) error {
	j.hasRun = true
	cutoff := GetPastTermCutoff(now, j.pastTermYears)

	terms, err := j.queries.GetTermsNeverScraped(ctx)
	if err != nil {
		return err
	}

	for _, term := range terms {
		if !IsTermInRange(term.Code, cutoff) {
			continue
		}

		phase := GetTermPhase(term.Code, now)
		if phase != PhasePast {
			continue
		}

		slog.Info("Scraping past term", "term", term.Code, "description", term.Description)
		stored, err := j.scraper.ScrapeTerm(ctx, term.Code)
		if err != nil {
			slog.Error("Failed to scrape past term", "term", term.Code, "error", err)
			continue
		}
		slog.Info("Past term scrape complete", "term", term.Code, "sections", stored)
	}

	return nil
}

// ActiveScrapeJob scrapes terms in active registration phase.
// Runs at a configurable interval (default 8 hours).
type ActiveScrapeJob struct {
	queries       *store.Queries
	scraper       *scraper.Scraper
	pastTermYears int
	interval      time.Duration
	lastRun       time.Time
}

func NewActiveScrapeJob(queries *store.Queries, scraper *scraper.Scraper, pastTermYears int, intervalHours int) *ActiveScrapeJob {
	return &ActiveScrapeJob{
		queries:       queries,
		scraper:       scraper,
		pastTermYears: pastTermYears,
		interval:      time.Duration(intervalHours) * time.Hour,
	}
}

func (j *ActiveScrapeJob) Name() string { return "active-scrape" }

func (j *ActiveScrapeJob) ShouldRun(now time.Time) bool {
	if j.lastRun.IsZero() {
		return true
	}
	return now.Sub(j.lastRun) >= j.interval-scrapeTolerance
}

func (j *ActiveScrapeJob) Run(ctx context.Context, now time.Time) error {
	j.lastRun = now
	cutoff := GetPastTermCutoff(now, j.pastTermYears)

	terms, err := j.queries.GetTerms(ctx)
	if err != nil {
		return err
	}

	for _, term := range terms {
		if !IsTermInRange(term.Code, cutoff) {
			continue
		}

		phase := GetTermPhase(term.Code, now)
		if phase != PhaseActiveRegistration {
			continue
		}

		if !j.shouldScrapeTerm(term, now) {
			continue
		}

		slog.Info("Scraping active term", "term", term.Code, "description", term.Description)
		stored, err := j.scraper.ScrapeTerm(ctx, term.Code)
		if err != nil {
			slog.Error("Failed to scrape term", "term", term.Code, "error", err)
			continue
		}
		slog.Info("Active term scrape complete", "term", term.Code, "sections", stored)
	}

	return nil
}

func (j *ActiveScrapeJob) shouldScrapeTerm(term *store.Term, now time.Time) bool {
	if !term.LastScrapedAt.Valid {
		return true
	}
	return now.Sub(term.LastScrapedAt.Time) >= j.interval-scrapeTolerance
}

// DailyScrapeJob scrapes terms in pre-registration phase.
// Runs once daily at a configurable hour.
type DailyScrapeJob struct {
	queries       *store.Queries
	scraper       *scraper.Scraper
	pastTermYears int
	targetHour    int
	lastRunDate   time.Time
}

func NewDailyScrapeJob(queries *store.Queries, scraper *scraper.Scraper, pastTermYears int, targetHour int) *DailyScrapeJob {
	return &DailyScrapeJob{
		queries:       queries,
		scraper:       scraper,
		pastTermYears: pastTermYears,
		targetHour:    targetHour,
	}
}

func (j *DailyScrapeJob) Name() string { return "daily-scrape" }

func (j *DailyScrapeJob) ShouldRun(now time.Time) bool {
	if now.Hour() != j.targetHour {
		return false
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return !j.lastRunDate.Equal(today)
}

func (j *DailyScrapeJob) Run(ctx context.Context, now time.Time) error {
	j.lastRunDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	cutoff := GetPastTermCutoff(now, j.pastTermYears)

	terms, err := j.queries.GetTerms(ctx)
	if err != nil {
		return err
	}

	for _, term := range terms {
		if !IsTermInRange(term.Code, cutoff) {
			continue
		}

		phase := GetTermPhase(term.Code, now)
		if phase != PhasePreRegistration {
			continue
		}

		if !j.shouldScrapeTerm(term, now) {
			continue
		}

		slog.Info("Scraping pre-registration term", "term", term.Code, "description", term.Description)
		stored, err := j.scraper.ScrapeTerm(ctx, term.Code)
		if err != nil {
			slog.Error("Failed to scrape term", "term", term.Code, "error", err)
			continue
		}
		slog.Info("Pre-registration term scrape complete", "term", term.Code, "sections", stored)
	}

	return nil
}

func (j *DailyScrapeJob) shouldScrapeTerm(term *store.Term, now time.Time) bool {
	if !term.LastScrapedAt.Valid {
		return true
	}
	return now.Sub(term.LastScrapedAt.Time) >= 24*time.Hour-scrapeTolerance
}
