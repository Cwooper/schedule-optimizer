package calendar

import (
	"testing"
	"time"
)

func pacific(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, Pacific)
}

// testSchedule is 2026-27 as the registrar published it.
var testSchedule = Schedule{scraped: map[string]scrapedDates{
	"202640": {start: date(2026, 9, 23), lastDay: date(2026, 12, 11), registrationOpens: date(2026, 5, 12)},
	"202710": {start: date(2027, 1, 5), lastDay: date(2027, 3, 19), registrationOpens: date(2026, 11, 3)},
	"202720": {start: date(2027, 3, 30), lastDay: date(2027, 6, 11), registrationOpens: date(2027, 2, 23)},
	"202730": {start: date(2027, 6, 22), lastDay: date(2027, 8, 20), registrationOpens: date(2027, 5, 4)},
}}

var berlin = mustLoadLocation("Europe/Berlin")

func TestSchedulePhase(t *testing.T) {
	tests := []struct {
		name     string
		schedule Schedule
		code     string
		now      time.Time
		want     Phase
	}{
		{"before Phase I", testSchedule, "202710", pacific(2026, 11, 2, 23, 59), PhasePreRegistration},
		{"Phase I opens", testSchedule, "202710", pacific(2026, 11, 3, 0, 0), PhaseRegistration},
		{"Berlin morning is still the Pacific day before", testSchedule, "202710",
			time.Date(2026, 11, 3, 8, 59, 0, 0, berlin), PhasePreRegistration},
		{"Berlin 9am is Pacific midnight", testSchedule, "202710",
			time.Date(2026, 11, 3, 9, 0, 0, 0, berlin), PhaseRegistration},
		{"first week of classes", testSchedule, "202710", pacific(2027, 1, 11, 23, 59), PhaseRegistration},
		{"one week in", testSchedule, "202710", pacific(2027, 1, 12, 0, 0), PhaseInSession},
		{"last day of term", testSchedule, "202710", pacific(2027, 3, 19, 23, 59), PhaseInSession},
		{"day after last day", testSchedule, "202710", pacific(2027, 3, 20, 0, 0), PhasePast},
		{"fall registration months ahead", testSchedule, "202640", pacific(2026, 5, 12, 12, 0), PhaseRegistration},

		// No scraped dates: estimated Winter 2027 is Jan 5 - Mar 20, registration Oct 27.
		{"estimate before registration", Schedule{}, "202710", pacific(2026, 10, 26, 12, 0), PhasePreRegistration},
		{"estimate registration", Schedule{}, "202710", pacific(2026, 10, 27, 0, 0), PhaseRegistration},
		{"estimate in session", Schedule{}, "202710", pacific(2027, 2, 1, 0, 0), PhaseInSession},
		{"estimate past", Schedule{}, "202710", pacific(2027, 3, 21, 0, 0), PhasePast},
		// A fresh database must not treat the current term as past, or the
		// backfill would scrape it once and never again.
		{"fresh database mid-term", Schedule{}, "202640", pacific(2026, 10, 15, 12, 0), PhaseInSession},

		{"registration date after start ignored", Schedule{scraped: map[string]scrapedDates{
			"202710": {registrationOpens: date(2027, 2, 1)},
		}}, "202710", pacific(2026, 12, 1, 0, 0), PhaseRegistration},
		{"invalid code", testSchedule, "202750", pacific(2027, 1, 1, 0, 0), PhasePast},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.schedule.Phase(tt.code, tt.now); got != tt.want {
				t.Errorf("Phase(%s, %v) = %v, want %v", tt.code, tt.now, got, tt.want)
			}
		})
	}
}

func TestDefaultTerm(t *testing.T) {
	all := []string{"202740", "202730", "202720", "202710", "202640", "202630"}
	tests := []struct {
		name  string
		codes []string
		now   time.Time
		want  string
	}{
		{"in session, next registration not open", all, pacific(2026, 10, 15, 12, 0), "202640"},
		// Winter 2027 Phase I opens Nov 3.
		{"next registration opens in 15 days", all, pacific(2026, 10, 19, 23, 59), "202640"},
		{"next registration opens in 14 days", all, pacific(2026, 10, 20, 0, 0), "202710"},
		{"next registration opens tomorrow", all, pacific(2026, 11, 2, 12, 0), "202710"},
		{"Berlin morning counts the Pacific date", all, time.Date(2026, 10, 20, 8, 0, 0, 0, berlin), "202640"},
		{"registering for next term mid-term", all, pacific(2026, 11, 3, 1, 0), "202710"},
		{"between terms", all, pacific(2026, 12, 20, 12, 0), "202710"},
		{"first week of the term", all, pacific(2027, 1, 8, 12, 0), "202710"},
		{"term in session", all, pacific(2027, 1, 20, 12, 0), "202710"},
		{"spring registration", all, pacific(2027, 3, 1, 12, 0), "202720"},
		// Summer opens May 4 and Fall 2027 (estimated) May 6.
		{"soonest of two opening", all, pacific(2027, 4, 25, 12, 0), "202730"},
		{"newest of two registering terms", all, pacific(2027, 5, 20, 12, 0), "202740"},
		{"soonest upcoming", []string{"202720", "202710"}, pacific(2026, 6, 1, 12, 0), "202710"},
		{"all past", []string{"202520", "202530"}, pacific(2026, 10, 15, 12, 0), "202530"},
		{"no terms", nil, pacific(2026, 10, 15, 12, 0), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testSchedule.DefaultTerm(tt.codes, tt.now); got != tt.want {
				t.Errorf("DefaultTerm = %q, want %q", got, tt.want)
			}
		})
	}

	sameDay := Schedule{scraped: map[string]scrapedDates{
		"202730": {registrationOpens: date(2027, 5, 4)},
		"202740": {registrationOpens: date(2027, 5, 4)},
	}}
	if got := sameDay.DefaultTerm(all, pacific(2027, 4, 25, 12, 0)); got != "202740" {
		t.Errorf("two opening the same day: DefaultTerm = %q, want the newer, 202740", got)
	}
}
