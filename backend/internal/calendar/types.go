package calendar

import "time"

// TermDates holds the start/end dates for a term and its finals week.
type TermDates struct {
	TermCode    string
	StartDate   time.Time
	EndDate     time.Time
	FinalsStart time.Time // zero if unknown
	FinalsEnd   time.Time // zero if unknown
}

// FinalMapping maps a class start time range + day pattern to a final exam slot.
type FinalMapping struct {
	TermCode       string
	TimeRangeStart string // "0800" (24h format)
	TimeRangeEnd   string // "0829"
	HasTuTh        bool   // true = T/Th column, false = non-T/Th column
	ExamDate       time.Time
	ExamStartTime  string // "0800"
	ExamEndTime    string // "1000"
}

// Holiday represents a no-class day within a term.
type Holiday struct {
	TermCode    string
	Date        time.Time
	Description string
}

// ImportantDate represents a key deadline or academic date within a term.
type ImportantDate struct {
	TermCode    string
	Date        time.Time
	Description string
	Category    string // "deadline", "registration", "academic"
}

// DrupalTermOption maps a Drupal node ID to a term.
// Scraped from the <select> on the important-dates-deadlines page.
type DrupalTermOption struct {
	NodeID      string // e.g. "275"
	Description string // e.g. "Spring 2026"
	TermCode    string // e.g. "202620"
}
