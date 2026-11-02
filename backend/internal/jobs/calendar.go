package jobs

import (
	"context"
	"errors"
	"time"

	"schedule-optimizer/internal/calendar"
	"schedule-optimizer/internal/scraper"
)

// CalendarJob refreshes Banner's term list and the registrar's calendar
// daily. It is registered ahead of the scrape jobs because they read both:
// the term list says what to scrape, the calendar which phase each term is
// in. Calendar rows are stored only for terms Banner lists, so the list is
// refreshed first.
type CalendarJob struct {
	scraper  *scraper.Scraper
	calendar *calendar.Service
	lastRun  time.Time
}

func NewCalendarJob(scraper *scraper.Scraper, cal *calendar.Service) *CalendarJob {
	return &CalendarJob{scraper: scraper, calendar: cal}
}

func (j *CalendarJob) Name() string { return "calendar" }

func (j *CalendarJob) ShouldRun(now time.Time) bool { return dueAfter(j.lastRun, now, 24*time.Hour) }

func (j *CalendarJob) Run(ctx context.Context, now time.Time) error {
	j.lastRun = now
	_, err := j.scraper.ScrapeTerms(ctx)
	return errors.Join(err, j.calendar.ScrapeAll(ctx))
}
