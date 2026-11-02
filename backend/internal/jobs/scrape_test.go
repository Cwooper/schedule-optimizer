package jobs

import (
	"context"
	"database/sql"
	"slices"
	"testing"
	"time"

	"schedule-optimizer/internal/calendar"
	"schedule-optimizer/internal/store"
	"schedule-optimizer/internal/testutil"
)

// scrapeDuration is how long each fake scrape takes. last_scraped_at lands
// this long after the scrape starts, as it does when Banner pages take
// minutes to fetch.
const scrapeDuration = 3 * time.Minute

// fakeScraper stands in for Banner: each scrape advances the simulated clock
// by scrapeDuration and writes last_scraped_at at that time.
type fakeScraper struct {
	t       *testing.T
	db      *sql.DB
	clock   time.Time
	scrapes map[string][]time.Time // scrape start times per term
	order   []string
}

func (f *fakeScraper) ScrapeTerm(ctx context.Context, term string) (int, error) {
	f.scrapes[term] = append(f.scrapes[term], f.clock)
	f.order = append(f.order, term)
	f.clock = f.clock.Add(scrapeDuration)
	if _, err := f.db.ExecContext(ctx, "UPDATE terms SET last_scraped_at = ? WHERE code = ?", f.clock.UTC(), term); err != nil {
		f.t.Fatal(err)
	}
	return 0, nil
}

// simulate runs the job through the service's check loop once a minute from
// start for the given span, with calendar tables empty so phases come from
// estimates. Ticks are UTC, as on the server. Like a time.Ticker, the loop
// skips ticks that fall inside a run.
func simulate(t *testing.T, codes []string, start time.Time, span time.Duration, seed func(*sql.DB)) *fakeScraper {
	t.Helper()
	db, queries := testutil.SetupTestDB(t)
	ctx := context.Background()
	for _, code := range codes {
		if err := queries.UpsertTerm(ctx, store.UpsertTermParams{Code: code, Description: code}); err != nil {
			t.Fatal(err)
		}
	}
	if seed != nil {
		seed(db)
	}

	fake := &fakeScraper{t: t, db: db, scrapes: make(map[string][]time.Time)}
	svc := NewService(time.Minute)
	svc.Register(NewScrapeJob(queries, fake, 5, 8, 3))
	start = start.UTC()
	for now := start; now.Before(start.Add(span)); {
		fake.clock = now
		svc.checkJobs(ctx, now)
		now = now.Add(time.Minute)
		for now.Before(fake.clock) {
			now = now.Add(time.Minute)
		}
	}
	return fake
}

func pacific(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, calendar.Pacific)
}

// TestScrapeCadence deploys at 10:00 Pacific on 2026-11-10, a week into
// Winter 2027 registration by the estimates, and runs three days.
func TestScrapeCadence(t *testing.T) {
	codes := []string{
		"202720", // pre-registration
		"202710", // registration
		"202640", // in session
		"202630", // past
		"202010", // past, outside the 5-year window
	}
	scraped := simulate(t, codes, pacific(2026, 11, 10, 10, 0), 72*time.Hour, nil)

	if got := scraped.order[:4]; !slices.Equal(got, []string{"202710", "202640", "202720", "202630"}) {
		t.Errorf("startup scraped %v, want registration first and past last", got)
	}

	registration := scraped.scrapes["202710"]
	for i := 1; i < len(registration); i++ {
		if gap := registration[i].Sub(registration[i-1]); gap < 8*time.Hour || gap > 8*time.Hour+scrapeDuration+time.Minute {
			t.Errorf("registration scrapes %d and %d are %v apart, want 8h plus the scrape time", i-1, i, gap)
		}
	}
	if len(registration) < 9 {
		t.Errorf("registration term scraped %d times in 72h, want 9", len(registration))
	}

	// The startup scrape, then each day at 3am: a mid-day restart must not
	// push the next daily scrape out a day.
	for _, code := range []string{"202640", "202720"} {
		var got []string
		for _, s := range scraped.scrapes[code] {
			got = append(got, s.In(calendar.Pacific).Format("Jan 2 15h"))
		}
		want := []string{"Nov 10 10h", "Nov 11 03h", "Nov 12 03h", "Nov 13 03h"}
		if !slices.Equal(got, want) {
			t.Errorf("term %s scraped at %v, want %v", code, got, want)
		}
	}

	if n := len(scraped.scrapes["202630"]); n != 1 {
		t.Errorf("past term scraped %d times, want once", n)
	}
	if n := len(scraped.scrapes["202010"]); n != 0 {
		t.Errorf("term outside the window scraped %d times, want never", n)
	}
}

func TestDailyScrapeAcrossDST(t *testing.T) {
	tests := []struct {
		name  string
		code  string // in session through the span
		start time.Time
		want  []string
	}{
		{"spring forward", "202710", pacific(2027, 3, 12, 12, 0),
			[]string{"Mar 12 12h", "Mar 13 03h", "Mar 14 03h", "Mar 15 03h"}},
		{"fall back", "202740", pacific(2027, 11, 5, 12, 0),
			[]string{"Nov 5 12h", "Nov 6 03h", "Nov 7 03h", "Nov 8 03h"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scraped := simulate(t, []string{tt.code}, tt.start, 72*time.Hour, nil)
			var got []string
			for _, s := range scraped.scrapes[tt.code] {
				got = append(got, s.In(calendar.Pacific).Format("Jan 2 15h"))
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("scraped at %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScrapeContinuesWhenScheduleFailsToLoad(t *testing.T) {
	scraped := simulate(t, []string{"202710"}, pacific(2026, 11, 10, 10, 0), time.Minute, func(db *sql.DB) {
		if _, err := db.Exec("INSERT INTO term_dates VALUES ('202710', 'not a date', '2027-03-19')"); err != nil {
			t.Fatal(err)
		}
	})
	if n := len(scraped.scrapes["202710"]); n != 1 {
		t.Errorf("term scraped %d times, want once on estimated dates", n)
	}
}

func TestDailyBoundary(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{"before the hour", pacific(2026, 11, 10, 2, 59), pacific(2026, 11, 9, 3, 0)},
		{"on the hour", pacific(2026, 11, 10, 3, 0), pacific(2026, 11, 10, 3, 0)},
		{"after the hour", pacific(2026, 11, 10, 23, 0), pacific(2026, 11, 10, 3, 0)},
		{"Berlin noon is Pacific 3am", time.Date(2026, 1, 15, 12, 0, 0, 0, time.FixedZone("CET", 3600)), pacific(2026, 1, 15, 3, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dailyBoundary(tt.now, 3); !got.Equal(tt.want) {
				t.Errorf("dailyBoundary(%v) = %v, want %v", tt.now, got, tt.want)
			}
		})
	}
}
