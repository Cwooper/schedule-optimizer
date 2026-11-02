package jobs

import (
	"context"
	"errors"
	"time"

	"schedule-optimizer/internal/calendar"
	"schedule-optimizer/internal/scraper"
)

// CalendarJob refreshes Banner's term list, then the registrar calendar, on
// startup and daily at the scrape hour. The list goes first because calendar
// rows are stored only for terms Banner lists.
type CalendarJob struct {
	scraper   *scraper.Scraper
	calendar  *calendar.Service
	cache     CacheInvalidator
	dailyHour int
	lastRun   time.Time
}

func NewCalendarJob(scraper *scraper.Scraper, cal *calendar.Service, cache CacheInvalidator, dailyHour int) *CalendarJob {
	return &CalendarJob{scraper: scraper, calendar: cal, cache: cache, dailyHour: dailyHour}
}

func (j *CalendarJob) Name() string { return "calendar" }

func (j *CalendarJob) ShouldRun(_ context.Context, now time.Time) bool {
	return j.lastRun.IsZero() || j.lastRun.Before(dailyBoundary(now, j.dailyHour))
}

func (j *CalendarJob) Run(ctx context.Context, now time.Time) error {
	j.lastRun = now
	_, err := j.scraper.ScrapeTerms(ctx)
	err = errors.Join(err, j.calendar.ScrapeAll(ctx))
	// ScrapeAll rewrites every published term's finals, even when one source
	// fails, so every cached term may hold old ones.
	j.cache.InvalidateAll()
	return err
}
