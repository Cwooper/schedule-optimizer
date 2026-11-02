package jobs

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"schedule-optimizer/internal/calendar"
	"schedule-optimizer/internal/scraper"
	"schedule-optimizer/internal/store"
	"schedule-optimizer/internal/termcode"
)

// scrapeTolerance accounts for job execution time when checking scrape intervals.
// Without this buffer, a job that takes N minutes would cause the next interval
// check to fail by N minutes, potentially skipping every other cycle.
const scrapeTolerance = 30 * time.Minute

func dueAfter(last, now time.Time, interval time.Duration) bool {
	return last.IsZero() || now.Sub(last) >= interval-scrapeTolerance
}

// termScraper scrapes the terms in the phases one job owns. Each phase
// belongs to exactly one job, so no term is scraped twice on its cadence:
//
//	pre-registration  daily       DailyScrapeJob
//	registration      every 8h    ActiveScrapeJob
//	in session        daily       DailyScrapeJob
//	past              once        PastTermBackfillJob
type termScraper struct {
	queries       *store.Queries
	scraper       *scraper.Scraper
	pastTermYears int
	phases        []calendar.Phase
	// due reports whether a term last scraped at the given time (zero if
	// never) needs scraping now.
	due func(lastScraped, now time.Time) bool
}

func (s *termScraper) run(ctx context.Context, now time.Time) error {
	schedule, err := calendar.LoadSchedule(ctx, s.queries)
	if err != nil {
		return err
	}
	terms, err := s.queries.GetTerms(ctx)
	if err != nil {
		return err
	}
	for _, term := range s.selectTerms(terms, schedule, now) {
		slog.Info("Scraping term", "term", term.Code, "description", term.Description)
		stored, err := s.scraper.ScrapeTerm(ctx, term.Code)
		if err != nil {
			slog.Error("Failed to scrape term", "term", term.Code, "error", err)
			continue
		}
		slog.Info("Term scrape complete", "term", term.Code, "sections", stored)
	}
	return nil
}

func (s *termScraper) selectTerms(terms []*store.Term, schedule calendar.Schedule, now time.Time) []*store.Term {
	oldest := oldestScrapedTerm(now, s.pastTermYears)
	var selected []*store.Term
	for _, term := range terms {
		if term.Code < oldest || !slices.Contains(s.phases, schedule.Phase(term.Code, now)) {
			continue
		}
		var last time.Time
		if term.LastScrapedAt.Valid {
			last = term.LastScrapedAt.Time
		}
		if s.due(last, now) {
			selected = append(selected, term)
		}
	}
	return selected
}

// oldestScrapedTerm is Winter of the year pastTermYears back; older terms
// are never scraped.
func oldestScrapedTerm(now time.Time, pastTermYears int) string {
	return termcode.Make(now.In(calendar.Pacific).Year()-pastTermYears, termcode.Winter)
}

// PastTermBackfillJob scrapes, once on startup, past terms that have never
// been scraped. A past term's sections no longer change.
type PastTermBackfillJob struct {
	terms  termScraper
	hasRun bool
}

func NewPastTermBackfillJob(queries *store.Queries, scraper *scraper.Scraper, pastTermYears int) *PastTermBackfillJob {
	return &PastTermBackfillJob{terms: termScraper{
		queries:       queries,
		scraper:       scraper,
		pastTermYears: pastTermYears,
		phases:        []calendar.Phase{calendar.PhasePast},
		due:           func(last, _ time.Time) bool { return last.IsZero() },
	}}
}

func (j *PastTermBackfillJob) Name() string { return "past-term-backfill" }

func (j *PastTermBackfillJob) ShouldRun(time.Time) bool { return !j.hasRun }

func (j *PastTermBackfillJob) Run(ctx context.Context, now time.Time) error {
	j.hasRun = true
	return j.terms.run(ctx, now)
}

// ActiveScrapeJob scrapes terms in registration, when seats move fastest,
// at a configurable interval (default 8 hours).
type ActiveScrapeJob struct {
	terms    termScraper
	interval time.Duration
	lastRun  time.Time
}

func NewActiveScrapeJob(queries *store.Queries, scraper *scraper.Scraper, pastTermYears int, intervalHours int) *ActiveScrapeJob {
	interval := time.Duration(intervalHours) * time.Hour
	return &ActiveScrapeJob{
		interval: interval,
		terms: termScraper{
			queries:       queries,
			scraper:       scraper,
			pastTermYears: pastTermYears,
			phases:        []calendar.Phase{calendar.PhaseRegistration},
			due:           func(last, now time.Time) bool { return dueAfter(last, now, interval) },
		},
	}
}

func (j *ActiveScrapeJob) Name() string { return "active-scrape" }

func (j *ActiveScrapeJob) ShouldRun(now time.Time) bool { return dueAfter(j.lastRun, now, j.interval) }

func (j *ActiveScrapeJob) Run(ctx context.Context, now time.Time) error {
	j.lastRun = now
	return j.terms.run(ctx, now)
}

// DailyScrapeJob scrapes terms before registration (every future term
// Banner exposes, since Banner decides how far ahead to publish) and terms
// in session, once a day at a configurable Pacific hour, plus once on
// startup.
type DailyScrapeJob struct {
	terms       termScraper
	targetHour  int
	lastRunDate time.Time
}

func NewDailyScrapeJob(queries *store.Queries, scraper *scraper.Scraper, pastTermYears int, targetHour int) *DailyScrapeJob {
	return &DailyScrapeJob{
		targetHour: targetHour,
		terms: termScraper{
			queries:       queries,
			scraper:       scraper,
			pastTermYears: pastTermYears,
			phases:        []calendar.Phase{calendar.PhasePreRegistration, calendar.PhaseInSession},
			due:           func(last, now time.Time) bool { return dueAfter(last, now, 24*time.Hour) },
		},
	}
}

func (j *DailyScrapeJob) Name() string { return "daily-scrape" }

func (j *DailyScrapeJob) ShouldRun(now time.Time) bool {
	if j.lastRunDate.IsZero() {
		return true
	}
	local := now.In(calendar.Pacific)
	return local.Hour() == j.targetHour && !j.lastRunDate.Equal(pacificDay(local))
}

func (j *DailyScrapeJob) Run(ctx context.Context, now time.Time) error {
	j.lastRunDate = pacificDay(now)
	return j.terms.run(ctx, now)
}

func pacificDay(t time.Time) time.Time {
	t = t.In(calendar.Pacific)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, calendar.Pacific)
}
