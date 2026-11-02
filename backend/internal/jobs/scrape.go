package jobs

import (
	"cmp"
	"context"
	"log/slog"
	"slices"
	"time"

	"schedule-optimizer/internal/calendar"
	"schedule-optimizer/internal/store"
	"schedule-optimizer/internal/termcode"
)

// termScraper is the part of *scraper.Scraper the job uses. ScrapeTerm sets
// the term's last_scraped_at when it succeeds, even on an empty term.
type termScraper interface {
	ScrapeTerm(ctx context.Context, term string) (int, error)
}

// CacheInvalidator is the schedule cache, which holds sections, finals, and
// grades in memory and must drop what a job rewrites.
type CacheInvalidator interface {
	Invalidate(term string)
	InvalidateAll()
}

// ScrapeJob scrapes each term's sections on a cadence set by its phase:
//
//	registration      interval after the last scrape, while seats move fastest
//	pre-registration  daily, after dailyHour Pacific
//	in session        daily, after dailyHour Pacific
//	past              once
//
// A term never scraped is due in any phase. Pre-registration covers every
// future term Banner lists, since Banner decides how far ahead to publish.
// The cadence comes from the phase and last_scraped_at, so a restart resumes
// it from the database. A term whose scrape failed waits retryBackoff before
// its next attempt.
type ScrapeJob struct {
	queries       *store.Queries
	scraper       termScraper
	cache         CacheInvalidator
	pastTermYears int
	interval      time.Duration
	dailyHour     int

	// failedAt is when each term's last scrape failed. It lives in memory: a
	// restart costs at most one early retry per term.
	failedAt map[string]time.Time
	// scheduleErr is the last LoadSchedule error logged, so a persistent
	// failure logs once rather than on every check.
	scheduleErr string
}

// retryBackoff spaces attempts at a failing term. Without it a term stays
// due, since only a successful scrape moves last_scraped_at, and a Banner
// outage would cost a full session on every one-minute check.
const retryBackoff = time.Hour

func NewScrapeJob(queries *store.Queries, scraper termScraper, cache CacheInvalidator, pastTermYears, intervalHours, dailyHour int) *ScrapeJob {
	return &ScrapeJob{
		queries:       queries,
		scraper:       scraper,
		cache:         cache,
		pastTermYears: pastTermYears,
		interval:      time.Duration(intervalHours) * time.Hour,
		dailyHour:     dailyHour,
		failedAt:      make(map[string]time.Time),
	}
}

func (j *ScrapeJob) Name() string { return "scrape" }

func (j *ScrapeJob) ShouldRun(ctx context.Context, now time.Time) bool {
	terms, err := j.dueTerms(ctx, now)
	if err != nil {
		slog.Error("Failed to select terms to scrape", "error", err)
	}
	return len(terms) > 0
}

func (j *ScrapeJob) Run(ctx context.Context, now time.Time) error {
	terms, err := j.dueTerms(ctx, now)
	if err != nil {
		return err
	}
	for _, term := range terms {
		if err := ctx.Err(); err != nil {
			return err
		}
		slog.Info("Scraping term", "term", term.Code, "description", term.Description)
		stored, err := j.scraper.ScrapeTerm(ctx, term.Code)
		if err != nil {
			j.failedAt[term.Code] = now
			slog.Error("Failed to scrape term", "term", term.Code, "error", err, "retry_after", retryBackoff)
			continue
		}
		j.cache.Invalidate(term.Code)
		slog.Info("Term scrape complete", "term", term.Code, "sections", stored)
	}
	return nil
}

// scrapeOrder puts terms in registration first, so a backfill of past terms
// on a fresh database does not hold up the seat counts students are waiting on.
var scrapeOrder = map[calendar.Phase]int{
	calendar.PhaseRegistration:    0,
	calendar.PhaseInSession:       1,
	calendar.PhasePreRegistration: 2,
	calendar.PhasePast:            3,
}

func (j *ScrapeJob) dueTerms(ctx context.Context, now time.Time) ([]*store.Term, error) {
	schedule, err := calendar.LoadSchedule(ctx, j.queries)
	// The zero Schedule falls back to estimated dates, which still place
	// every term in a phase.
	j.logScheduleErr(err)
	terms, err := j.queries.GetTerms(ctx)
	if err != nil {
		return nil, err
	}

	oldest := oldestScrapedTerm(now, j.pastTermYears)
	var due []*store.Term
	phases := make(map[string]calendar.Phase)
	for _, term := range terms {
		if term.Code < oldest || now.Sub(j.failedAt[term.Code]) < retryBackoff {
			continue
		}
		phase := schedule.Phase(term.Code, now)
		if j.due(phase, term.LastScrapedAt.Time, now) {
			due = append(due, term)
			phases[term.Code] = phase
		}
	}
	slices.SortStableFunc(due, func(a, b *store.Term) int {
		return cmp.Compare(scrapeOrder[phases[a.Code]], scrapeOrder[phases[b.Code]])
	})
	return due, nil
}

func (j *ScrapeJob) logScheduleErr(err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	if msg == j.scheduleErr {
		return
	}
	if err != nil {
		slog.Error("Failed to load term schedule; using estimates", "error", err)
	} else {
		slog.Info("Term schedule loaded after earlier failure")
	}
	j.scheduleErr = msg
}

// due reports whether a term in phase, last scraped at last (zero if never),
// needs scraping at now.
func (j *ScrapeJob) due(phase calendar.Phase, last, now time.Time) bool {
	switch {
	case last.IsZero():
		return true
	case phase == calendar.PhaseRegistration:
		return now.Sub(last) >= j.interval
	case phase == calendar.PhasePast:
		return false
	default:
		return last.Before(dailyBoundary(now, j.dailyHour))
	}
}

// dailyBoundary is the latest instant at or before now when Pacific clocks
// read hour:00. Anchoring daily work to it rather than to 24 hours elapsed
// keeps the cadence through DST changes and after a restart mid-day.
func dailyBoundary(now time.Time, hour int) time.Time {
	local := now.In(calendar.Pacific)
	boundary := time.Date(local.Year(), local.Month(), local.Day(), hour, 0, 0, 0, calendar.Pacific)
	if boundary.After(now) {
		boundary = time.Date(local.Year(), local.Month(), local.Day()-1, hour, 0, 0, 0, calendar.Pacific)
	}
	return boundary
}

// oldestScrapedTerm is Winter of the year pastTermYears back; older terms
// are never scraped.
func oldestScrapedTerm(now time.Time, pastTermYears int) string {
	return termcode.Make(now.In(calendar.Pacific).Year()-pastTermYears, termcode.Winter)
}
