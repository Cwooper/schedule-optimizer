package jobs

import (
	"context"
	"time"

	"schedule-optimizer/internal/calendar"
)

// CalendarScrapeJob scrapes academic calendar data (term dates, finals, holidays).
// Runs once daily.
type CalendarScrapeJob struct {
	calendarService *calendar.Service
	interval        time.Duration
	lastRun         time.Time
}

func NewCalendarScrapeJob(calendarService *calendar.Service) *CalendarScrapeJob {
	return &CalendarScrapeJob{
		calendarService: calendarService,
		interval:        24 * time.Hour,
	}
}

func (j *CalendarScrapeJob) Name() string { return "calendar-scrape" }

func (j *CalendarScrapeJob) ShouldRun(now time.Time) bool {
	if j.lastRun.IsZero() {
		return true
	}
	return now.Sub(j.lastRun) >= j.interval-scrapeTolerance
}

func (j *CalendarScrapeJob) Run(ctx context.Context, now time.Time) error {
	j.lastRun = now
	return j.calendarService.ScrapeAll(ctx)
}
