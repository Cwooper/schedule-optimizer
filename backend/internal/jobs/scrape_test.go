package jobs

import (
	"testing"
	"time"
)

func TestBootstrapJob_ShouldRun(t *testing.T) {
	job := &BootstrapJob{}
	now := time.Now()

	if !job.ShouldRun(now) {
		t.Error("BootstrapJob should run when hasRun is false")
	}

	job.hasRun = true
	if job.ShouldRun(now) {
		t.Error("BootstrapJob should not run when hasRun is true")
	}
}

func TestPastTermBackfillJob_ShouldRun(t *testing.T) {
	job := &PastTermBackfillJob{}
	now := time.Now()

	if !job.ShouldRun(now) {
		t.Error("PastTermBackfillJob should run when hasRun is false")
	}

	job.hasRun = true
	if job.ShouldRun(now) {
		t.Error("PastTermBackfillJob should not run when hasRun is true")
	}
}

func TestActiveScrapeJob_ShouldRun(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		lastRun  time.Time
		now      time.Time
		want     bool
	}{
		{
			name:     "first run (zero lastRun)",
			interval: 8 * time.Hour,
			lastRun:  time.Time{},
			now:      time.Date(2025, 1, 15, 10, 0, 0, 0, time.Local),
			want:     true,
		},
		{
			name:     "interval not elapsed",
			interval: 8 * time.Hour,
			lastRun:  time.Date(2025, 1, 15, 10, 0, 0, 0, time.Local),
			now:      time.Date(2025, 1, 15, 14, 0, 0, 0, time.Local), // 4 hours later
			want:     false,
		},
		{
			name:     "interval elapsed exactly",
			interval: 8 * time.Hour,
			lastRun:  time.Date(2025, 1, 15, 10, 0, 0, 0, time.Local),
			now:      time.Date(2025, 1, 15, 18, 0, 0, 0, time.Local), // 8 hours later
			want:     true,
		},
		{
			name:     "within tolerance window",
			interval: 8 * time.Hour,
			lastRun:  time.Date(2025, 1, 15, 10, 0, 0, 0, time.Local),
			now:      time.Date(2025, 1, 15, 17, 35, 0, 0, time.Local), // 7h35m later (within 30m tolerance)
			want:     true,
		},
		{
			name:     "just outside tolerance window",
			interval: 8 * time.Hour,
			lastRun:  time.Date(2025, 1, 15, 10, 0, 0, 0, time.Local),
			now:      time.Date(2025, 1, 15, 17, 25, 0, 0, time.Local), // 7h25m later (outside 30m tolerance)
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &ActiveScrapeJob{
				interval: tt.interval,
				lastRun:  tt.lastRun,
			}
			got := job.ShouldRun(tt.now)
			if got != tt.want {
				t.Errorf("ShouldRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDailyScrapeJob_ShouldRun(t *testing.T) {
	tests := []struct {
		name        string
		targetHour  int
		lastRunDate time.Time
		now         time.Time
		want        bool
	}{
		{
			name:        "first run (zero lastRunDate)",
			targetHour:  3,
			lastRunDate: time.Time{},
			now:         time.Date(2025, 1, 15, 10, 0, 0, 0, time.Local),
			want:        true,
		},
		{
			name:        "wrong hour after first run",
			targetHour:  3,
			lastRunDate: time.Date(2025, 1, 14, 0, 0, 0, 0, time.Local),
			now:         time.Date(2025, 1, 15, 10, 0, 0, 0, time.Local),
			want:        false,
		},
		{
			name:        "correct hour, never run",
			targetHour:  3,
			lastRunDate: time.Time{},
			now:         time.Date(2025, 1, 15, 3, 30, 0, 0, time.Local),
			want:        true,
		},
		{
			name:        "correct hour, already run today",
			targetHour:  3,
			lastRunDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local),
			now:         time.Date(2025, 1, 15, 3, 45, 0, 0, time.Local),
			want:        false,
		},
		{
			name:        "correct hour, ran yesterday",
			targetHour:  3,
			lastRunDate: time.Date(2025, 1, 14, 0, 0, 0, 0, time.Local),
			now:         time.Date(2025, 1, 15, 3, 15, 0, 0, time.Local),
			want:        true,
		},
		{
			name:        "minute 0 of target hour",
			targetHour:  3,
			lastRunDate: time.Time{},
			now:         time.Date(2025, 1, 15, 3, 0, 0, 0, time.Local),
			want:        true,
		},
		{
			name:        "minute 59 of target hour",
			targetHour:  3,
			lastRunDate: time.Time{},
			now:         time.Date(2025, 1, 15, 3, 59, 0, 0, time.Local),
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &DailyScrapeJob{
				targetHour:  tt.targetHour,
				lastRunDate: tt.lastRunDate,
			}
			got := job.ShouldRun(tt.now)
			if got != tt.want {
				t.Errorf("ShouldRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScrapeTolerance(t *testing.T) {
	if scrapeTolerance != 30*time.Minute {
		t.Errorf("scrapeTolerance = %v, want 30m", scrapeTolerance)
	}
}
