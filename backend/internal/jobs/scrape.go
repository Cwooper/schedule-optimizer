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
// the term's last_scraped_at when it finishes.
type termScraper interface {
	ScrapeTerm(ctx context.Context, term string) (int, error)
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
// Due-ness comes from the phase and last_scraped_at alone, so a restart
// resumes the cadence from the database.
type ScrapeJob struct {
	queries       *store.Queries
	scraper       termScraper
	pastTermYears int
	interval      time.Duration
	dailyHour     int
}

func NewScrapeJob(queries *store.Queries, scraper termScraper, pastTermYears, intervalHours, dailyHour int) *ScrapeJob {
	return &ScrapeJob{
		queries:       queries,
		scraper:       scraper,
		pastTermYears: pastTermYears,
		interval:      time.Duration(intervalHours) * time.Hour,
		dailyHour:     dailyHour,
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
			slog.Error("Failed to scrape term", "term", term.Code, "error", err)
			continue
		}
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
	if err != nil {
		// The zero Schedule falls back to estimated dates, which still
		// place every term in a phase.
		slog.Error("Failed to load term schedule; using estimates", "error", err)
	}
	terms, err := j.queries.GetTerms(ctx)
	if err != nil {
		return nil, err
	}

	oldest := oldestScrapedTerm(now, j.pastTermYears)
	var due []*store.Term
	phases := make(map[string]calendar.Phase)
	for _, term := range terms {
		if term.Code < oldest {
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
