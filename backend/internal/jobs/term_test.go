package jobs

import (
	"testing"
	"time"
)

func TestParseTermCode(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		wantYear    int
		wantQuarter int
		wantErr     bool
	}{
		{"Winter 2025", "202510", 2025, QuarterWinter, false},
		{"Spring 2025", "202520", 2025, QuarterSpring, false},
		{"Summer 2025", "202530", 2025, QuarterSummer, false},
		{"Fall 2025", "202540", 2025, QuarterFall, false},
		{"Winter 2020", "202010", 2020, QuarterWinter, false},
		{"Invalid length", "20251", 0, 0, true},
		{"Invalid quarter", "202515", 0, 0, true},
		{"Invalid year", "abcd10", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			year, quarter, err := ParseTermCode(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTermCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if year != tt.wantYear {
				t.Errorf("ParseTermCode() year = %v, want %v", year, tt.wantYear)
			}
			if quarter != tt.wantQuarter {
				t.Errorf("ParseTermCode() quarter = %v, want %v", quarter, tt.wantQuarter)
			}
		})
	}
}

func TestMakeTermCode(t *testing.T) {
	tests := []struct {
		year    int
		quarter int
		want    string
	}{
		{2025, QuarterWinter, "202510"},
		{2025, QuarterSpring, "202520"},
		{2025, QuarterSummer, "202530"},
		{2025, QuarterFall, "202540"},
	}

	for _, tt := range tests {
		got := MakeTermCode(tt.year, tt.quarter)
		if got != tt.want {
			t.Errorf("MakeTermCode(%d, %d) = %v, want %v", tt.year, tt.quarter, got, tt.want)
		}
	}
}

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
			name:      "Winter future",
			termCode:  "202510",
			now:       time.Date(2024, 9, 1, 0, 0, 0, 0, time.Local),
			wantPhase: PhaseFuture,
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

func TestNextQuarter(t *testing.T) {
	tests := []struct {
		year        int
		quarter     int
		wantYear    int
		wantQuarter int
	}{
		{2025, QuarterWinter, 2025, QuarterSpring},
		{2025, QuarterSpring, 2025, QuarterSummer},
		{2025, QuarterSummer, 2025, QuarterFall},
		{2025, QuarterFall, 2026, QuarterWinter},
	}

	for _, tt := range tests {
		gotYear, gotQuarter := NextQuarter(tt.year, tt.quarter)
		if gotYear != tt.wantYear || gotQuarter != tt.wantQuarter {
			t.Errorf("NextQuarter(%d, %d) = (%d, %d), want (%d, %d)",
				tt.year, tt.quarter, gotYear, gotQuarter, tt.wantYear, tt.wantQuarter)
		}
	}
}

func TestCurrentTermCode(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
		want string
	}{
		{"January - Winter", time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local), "202510"},
		{"March - Winter", time.Date(2025, 3, 15, 0, 0, 0, 0, time.Local), "202510"},
		{"April - Spring", time.Date(2025, 4, 15, 0, 0, 0, 0, time.Local), "202520"},
		{"May - Spring", time.Date(2025, 5, 15, 0, 0, 0, 0, time.Local), "202520"},
		{"June - Summer", time.Date(2025, 6, 15, 0, 0, 0, 0, time.Local), "202530"},
		{"August - Summer", time.Date(2025, 8, 15, 0, 0, 0, 0, time.Local), "202530"},
		{"September - Fall", time.Date(2025, 9, 15, 0, 0, 0, 0, time.Local), "202540"},
		{"December - Fall", time.Date(2025, 12, 15, 0, 0, 0, 0, time.Local), "202540"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CurrentTermCode(tt.now)
			if got != tt.want {
				t.Errorf("CurrentTermCode(%v) = %v, want %v", tt.now, got, tt.want)
			}
		})
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
		{PhaseFuture, "future"},
		{TermPhase(99), "unknown"},
	}

	for _, tt := range tests {
		got := tt.phase.String()
		if got != tt.want {
			t.Errorf("TermPhase(%d).String() = %v, want %v", tt.phase, got, tt.want)
		}
	}
}
