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
	dailyHour int
	lastRun   time.Time
}

func NewCalendarJob(scraper *scraper.Scraper, cal *calendar.Service, dailyHour int) *CalendarJob {
	return &CalendarJob{scraper: scraper, calendar: cal, dailyHour: dailyHour}
}

func (j *CalendarJob) Name() string { return "calendar" }

func (j *CalendarJob) ShouldRun(_ context.Context, now time.Time) bool {
	return j.lastRun.IsZero() || j.lastRun.Before(dailyBoundary(now, j.dailyHour))
}

func (j *CalendarJob) Run(ctx context.Context, now time.Time) error {
	j.lastRun = now
	_, err := j.scraper.ScrapeTerms(ctx)
	return errors.Join(err, j.calendar.ScrapeAll(ctx))
}
