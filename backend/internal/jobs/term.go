package jobs

import (
	"fmt"
	"strconv"
	"time"
)

// TermPhase represents the lifecycle phase of a term for scraping purposes.
type TermPhase int

const (
	PhasePast               TermPhase = iota // Term has ended
	PhasePreRegistration                     // Registration not yet open (scrape daily)
	PhaseActiveRegistration                  // Registration open or term in progress (scrape frequently)
	PhaseFuture                              // Too far in the future to scrape
)

func (p TermPhase) String() string {
	switch p {
	case PhasePast:
		return "past"
	case PhasePreRegistration:
		return "pre-registration"
	case PhaseActiveRegistration:
		return "active-registration"
	case PhaseFuture:
		return "future"
	default:
		return "unknown"
	}
}

// Term code quarters
const (
	QuarterWinter = 10
	QuarterSpring = 20
	QuarterSummer = 30
	QuarterFall   = 40
)

// ParseTermCode extracts year and quarter from a term code.
// Term codes are formatted as YYYYQQ where QQ is 10 (Winter), 20 (Spring), 30 (Summer), or 40 (Fall).
// Example: "202520" -> 2025, 20 (Spring 2025)
func ParseTermCode(code string) (year int, quarter int, err error) {
	if len(code) != 6 {
		return 0, 0, fmt.Errorf("invalid term code length: %s", code)
	}

	year, err = strconv.Atoi(code[:4])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid year in term code: %s", code)
	}

	quarter, err = strconv.Atoi(code[4:])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid quarter in term code: %s", code)
	}

	if quarter != QuarterWinter && quarter != QuarterSpring && quarter != QuarterSummer && quarter != QuarterFall {
		return 0, 0, fmt.Errorf("invalid quarter value %d in term code: %s", quarter, code)
	}

	return year, quarter, nil
}

// MakeTermCode creates a term code from year and quarter.
func MakeTermCode(year, quarter int) string {
	return fmt.Sprintf("%d%02d", year, quarter)
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
// TODO(#37): Scrape actual registration dates from Banner for precise detection.
func GetTermPhase(termCode string, now time.Time) TermPhase {
	year, quarter, err := ParseTermCode(termCode)
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
	// Pre-registration: within 60 days before registration opens
	preRegStart := regStart.AddDate(0, 0, -60)
	if now.After(preRegStart) {
		return PhasePreRegistration
	}
	return PhaseFuture
}

// getTermDates returns approximate start and end dates for a term.
func getTermDates(year, quarter int) (start, end time.Time) {
	loc := time.Local
	switch quarter {
	case QuarterWinter:
		return time.Date(year, 1, 5, 0, 0, 0, 0, loc),
			time.Date(year, 3, 20, 23, 59, 59, 0, loc)
	case QuarterSpring:
		return time.Date(year, 4, 1, 0, 0, 0, 0, loc),
			time.Date(year, 6, 12, 23, 59, 59, 0, loc)
	case QuarterSummer:
		return time.Date(year, 6, 23, 0, 0, 0, 0, loc),
			time.Date(year, 8, 21, 23, 59, 59, 0, loc)
	case QuarterFall:
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
	return MakeTermCode(cutoffYear, QuarterWinter)
}

// IsTermInRange checks if a term code is within the scraping range.
// Returns true if termCode >= cutoff (newer or equal).
func IsTermInRange(termCode string, cutoff string) bool {
	return termCode >= cutoff
}

// NextQuarter returns the next quarter after the given one.
// Wraps from Fall to next year's Winter.
func NextQuarter(year, quarter int) (int, int) {
	switch quarter {
	case QuarterWinter:
		return year, QuarterSpring
	case QuarterSpring:
		return year, QuarterSummer
	case QuarterSummer:
		return year, QuarterFall
	case QuarterFall:
		return year + 1, QuarterWinter
	default:
		return year, quarter
	}
}

// CurrentTermCode returns the term code for the current or most recent term.
func CurrentTermCode(now time.Time) string {
	year := now.Year()
	month := now.Month()

	var quarter int
	switch {
	case month >= 9: // Sep-Dec
		quarter = QuarterFall
	case month >= 6: // Jun-Aug
		quarter = QuarterSummer
	case month >= 4: // Apr-May
		quarter = QuarterSpring
	default: // Jan-Mar
		quarter = QuarterWinter
	}

	return MakeTermCode(year, quarter)
}

// GetAcademicYearTerms returns the four term codes for an academic year.
// Academic year N runs from Fall of year N-1 through Summer of year N.
// Example: GetAcademicYearTerms(2025) returns ["202440", "202510", "202520", "202530"]
func GetAcademicYearTerms(academicYear int) []string {
	return []string{
		MakeTermCode(academicYear-1, QuarterFall),
		MakeTermCode(academicYear, QuarterWinter),
		MakeTermCode(academicYear, QuarterSpring),
		MakeTermCode(academicYear, QuarterSummer),
	}
}
