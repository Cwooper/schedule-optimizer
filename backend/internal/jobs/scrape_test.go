package jobs

import (
	"database/sql"
	"slices"
	"testing"
	"time"

	"schedule-optimizer/internal/calendar"
	"schedule-optimizer/internal/store"
)

func TestDueAfter(t *testing.T) {
	last := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		last    time.Time
		elapsed time.Duration
		want    bool
	}{
		{"never run", time.Time{}, 0, true},
		{"interval not elapsed", last, 4 * time.Hour, false},
		{"interval elapsed exactly", last, 8 * time.Hour, true},
		{"within tolerance window", last, 7*time.Hour + 35*time.Minute, true},
		{"just outside tolerance window", last, 7*time.Hour + 25*time.Minute, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dueAfter(tt.last, last.Add(tt.elapsed), 8*time.Hour); got != tt.want {
				t.Errorf("dueAfter = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDailyScrapeJob_ShouldRun(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	pt := func(day, hour, minute int) time.Time {
		return time.Date(2025, 1, day, hour, minute, 0, 0, calendar.Pacific)
	}
	tests := []struct {
		name        string
		lastRunDate time.Time
		now         time.Time
		want        bool
	}{
		{"first check always runs", time.Time{}, pt(15, 10, 0), true},
		{"wrong hour", pt(14, 0, 0), pt(15, 10, 0), false},
		{"target hour, ran yesterday", pt(14, 0, 0), pt(15, 3, 15), true},
		{"target hour, already ran today", pt(15, 0, 0), pt(15, 3, 45), false},
		{"last minute of target hour", pt(14, 0, 0), pt(15, 3, 59), true},
		{"3am in Berlin is not 3am Pacific", pt(14, 0, 0), time.Date(2025, 1, 15, 3, 0, 0, 0, berlin), false},
		{"noon in Berlin is 3am Pacific", pt(14, 0, 0), time.Date(2025, 1, 15, 12, 0, 0, 0, berlin), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &DailyScrapeJob{targetHour: 3, lastRunDate: tt.lastRunDate}
			if got := job.ShouldRun(tt.now); got != tt.want {
				t.Errorf("ShouldRun(%v) = %v, want %v", tt.now, got, tt.want)
			}
		})
	}
}

// TestJobsSplitTermsByPhase runs every job's term selection over the same
// terms on 2026-11-10, a week into Winter 2027 registration, with term
// dates estimated (an empty schedule).
func TestJobsSplitTermsByPhase(t *testing.T) {
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, calendar.Pacific)
	scraped := func(ago time.Duration) sql.NullTime {
		return sql.NullTime{Time: now.Add(-ago), Valid: true}
	}
	terms := []*store.Term{
		{Code: "202720", LastScrapedAt: scraped(2 * time.Hour)},  // pre-registration, fresh
		{Code: "202710", LastScrapedAt: scraped(9 * time.Hour)},  // registration, stale
		{Code: "202640", LastScrapedAt: scraped(25 * time.Hour)}, // in session, stale
		{Code: "202630"}, // past, never scraped
		{Code: "202620", LastScrapedAt: scraped(90 * 24 * time.Hour)}, // past, scraped
		{Code: "202010"}, // past, never scraped, inside a 6-year window only
	}

	tests := []struct {
		name string
		job  termScraper
		want []string
	}{
		{"backfill", NewPastTermBackfillJob(nil, nil, 5).terms, []string{"202630"}},
		{"backfill wider window", NewPastTermBackfillJob(nil, nil, 6).terms, []string{"202630", "202010"}},
		{"active", NewActiveScrapeJob(nil, nil, 5, 8).terms, []string{"202710"}},
		{"daily", NewDailyScrapeJob(nil, nil, 5, 3).terms, []string{"202640"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			for _, term := range tt.job.selectTerms(terms, calendar.Schedule{}, now) {
				got = append(got, term.Code)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("selected %v, want %v", got, tt.want)
			}
		})
	}
}
