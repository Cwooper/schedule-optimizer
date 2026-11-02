package jobs

import (
	"testing"
	"time"
)

func TestGetTermPhase(t *testing.T) {
	tests := []struct {
		name      string
		termCode  string
		now       time.Time
		wantPhase TermPhase
	}{
		// Winter 2025: Jan 5 - Mar 20, reg opens Dec 26 2024 (40 days before)
		{
			name:      "Winter active during term",
			termCode:  "202510",
			now:       time.Date(2025, 2, 15, 0, 0, 0, 0, time.Local),
			wantPhase: PhaseActiveRegistration,
		},
		{
			name:      "Winter active during registration before term",
			termCode:  "202510",
			now:       time.Date(2024, 12, 28, 0, 0, 0, 0, time.Local),
			wantPhase: PhaseActiveRegistration,
		},
		{
			name:      "Winter past after term ends",
			termCode:  "202510",
			now:       time.Date(2025, 3, 25, 0, 0, 0, 0, time.Local),
			wantPhase: PhasePast,
		},
		{
			name:      "Winter pre-registration",
			termCode:  "202510",
			now:       time.Date(2024, 11, 15, 0, 0, 0, 0, time.Local),
			wantPhase: PhasePreRegistration,
		},
		{
			name:      "Winter far before registration still pre-registration",
			termCode:  "202510",
			now:       time.Date(2024, 9, 1, 0, 0, 0, 0, time.Local),
			wantPhase: PhasePreRegistration,
		},
		{
			name:      "Fall many months before registration still pre-registration",
			termCode:  "202640",
			now:       time.Date(2026, 5, 15, 0, 0, 0, 0, time.Local),
			wantPhase: PhasePreRegistration,
		},

		// Spring 2025: Apr 1 - Jun 12, reg opens Feb 20 (40 days before)
		{
			name:      "Spring active during registration",
			termCode:  "202520",
			now:       time.Date(2025, 3, 1, 0, 0, 0, 0, time.Local),
			wantPhase: PhaseActiveRegistration,
		},
		{
			name:      "Spring active during term",
			termCode:  "202520",
			now:       time.Date(2025, 5, 1, 0, 0, 0, 0, time.Local),
			wantPhase: PhaseActiveRegistration,
		},
		{
			name:      "Spring past",
			termCode:  "202520",
			now:       time.Date(2025, 6, 20, 0, 0, 0, 0, time.Local),
			wantPhase: PhasePast,
		},

		// Fall 2025: Sep 24 - Dec 12, reg opens Aug 15 (40 days before)
		{
			name:      "Fall active during registration",
			termCode:  "202540",
			now:       time.Date(2025, 8, 20, 0, 0, 0, 0, time.Local),
			wantPhase: PhaseActiveRegistration,
		},
		{
			name:      "Fall pre-registration",
			termCode:  "202540",
			now:       time.Date(2025, 7, 1, 0, 0, 0, 0, time.Local),
			wantPhase: PhasePreRegistration,
		},

		// Invalid term code
		{
			name:      "Invalid code treated as past",
			termCode:  "invalid",
			now:       time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local),
			wantPhase: PhasePast,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTermPhase(tt.termCode, tt.now)
			if got != tt.wantPhase {
				t.Errorf("GetTermPhase(%q, %v) = %v, want %v", tt.termCode, tt.now, got, tt.wantPhase)
			}
		})
	}
}

func TestGetPastTermCutoff(t *testing.T) {
	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.Local)

	tests := []struct {
		yearsBack int
		want      string
	}{
		{5, "202010"}, // 5 years back from 2025 = 2020 Winter
		{1, "202410"}, // 1 year back from 2025 = 2024 Winter
		{0, "202510"}, // 0 years back = 2025 Winter
	}

	for _, tt := range tests {
		got := GetPastTermCutoff(now, tt.yearsBack)
		if got != tt.want {
			t.Errorf("GetPastTermCutoff(%v, %d) = %v, want %v", now, tt.yearsBack, got, tt.want)
		}
	}
}

func TestIsTermInRange(t *testing.T) {
	tests := []struct {
		termCode string
		cutoff   string
		want     bool
	}{
		{"202520", "202010", true},  // 2025 Spring >= 2020 Winter
		{"202010", "202010", true},  // Equal
		{"201940", "202010", false}, // 2019 Fall < 2020 Winter
		{"202540", "202520", true},  // 2025 Fall >= 2025 Spring
	}

	for _, tt := range tests {
		got := IsTermInRange(tt.termCode, tt.cutoff)
		if got != tt.want {
			t.Errorf("IsTermInRange(%q, %q) = %v, want %v", tt.termCode, tt.cutoff, got, tt.want)
		}
	}
}

func TestTermPhaseString(t *testing.T) {
	tests := []struct {
		phase TermPhase
		want  string
	}{
		{PhasePast, "past"},
		{PhasePreRegistration, "pre-registration"},
		{PhaseActiveRegistration, "active-registration"},
		{TermPhase(99), "unknown"},
	}

	for _, tt := range tests {
		got := tt.phase.String()
		if got != tt.want {
			t.Errorf("TermPhase(%d).String() = %v, want %v", tt.phase, got, tt.want)
		}
	}
}
