package jobs

import (
	"time"

	"schedule-optimizer/internal/termcode"
)

// TermPhase represents the lifecycle phase of a term for scraping purposes.
type TermPhase int

const (
	PhasePast               TermPhase = iota // Term has ended
	PhasePreRegistration                     // Registration not yet open (scrape daily)
	PhaseActiveRegistration                  // Registration open or term in progress (scrape frequently)
)

func (p TermPhase) String() string {
	switch p {
	case PhasePast:
		return "past"
	case PhasePreRegistration:
		return "pre-registration"
	case PhaseActiveRegistration:
		return "active-registration"
	default:
		return "unknown"
	}
}

// GetTermPhase determines the current lifecycle phase of a term.
//
// Phase detection uses approximate term dates:
//   - Winter: Jan 5 - Mar 20
//   - Spring: Apr 1 - Jun 12
//   - Summer: Jun 23 - Aug 21
//   - Fall: Sep 24 - Dec 12
//
// Registration is estimated to open 40 days before term start.
// Any term not yet in active registration is treated as pre-registration so
// the daily job scrapes every future term Banner exposes — Banner controls
// how far ahead terms are surfaced.
// TODO(#37): Scrape actual registration dates from Banner for precise detection.
func GetTermPhase(termCode string, now time.Time) TermPhase {
	year, quarter, err := termcode.Parse(termCode)
	if err != nil {
		return PhasePast // Invalid codes treated as past
	}

	termStart, termEnd := getTermDates(year, quarter)
	regStart := termStart.AddDate(0, 0, -40)

	if now.After(termEnd) {
		return PhasePast
	}
	if now.After(regStart) {
		return PhaseActiveRegistration
	}
	return PhasePreRegistration
}

// getTermDates returns approximate start and end dates for a term.
func getTermDates(year, quarter int) (start, end time.Time) {
	loc := time.Local
	switch quarter {
	case termcode.Winter:
		return time.Date(year, 1, 5, 0, 0, 0, 0, loc),
			time.Date(year, 3, 20, 23, 59, 59, 0, loc)
	case termcode.Spring:
		return time.Date(year, 4, 1, 0, 0, 0, 0, loc),
			time.Date(year, 6, 12, 23, 59, 59, 0, loc)
	case termcode.Summer:
		return time.Date(year, 6, 23, 0, 0, 0, 0, loc),
			time.Date(year, 8, 21, 23, 59, 59, 0, loc)
	case termcode.Fall:
		return time.Date(year, 9, 24, 0, 0, 0, 0, loc),
			time.Date(year, 12, 12, 23, 59, 59, 0, loc)
	default:
		return time.Time{}, time.Time{}
	}
}

// GetPastTermCutoff returns the oldest term code worth scraping.
// Terms older than yearsBack years from now are not scraped.
func GetPastTermCutoff(now time.Time, yearsBack int) string {
	cutoffYear := now.Year() - yearsBack
	return termcode.Make(cutoffYear, termcode.Winter)
}

// IsTermInRange checks if a term code is within the scraping range.
// Returns true if termCode >= cutoff (newer or equal).
func IsTermInRange(termCode string, cutoff string) bool {
	return termCode >= cutoff
}
