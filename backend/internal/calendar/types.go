package calendar

import "time"

// Parser output. Every time.Time is midnight Pacific on a calendar date.

type termDates struct {
	termCode string
	start    time.Time
	end      time.Time // last day, inclusive
}

type finalMapping struct {
	termCode string
	// Classes starting from rangeStart through rangeEnd (HHMM, inclusive)
	// take this exam slot.
	rangeStart string
	rangeEnd   string
	hasTuTh    bool
	examDate   time.Time
	examStart  string // HHMM
	examEnd    string // HHMM
}

type category string

const (
	categoryAcademic     category = "academic"
	categoryRegistration category = "registration"
	categoryDeadline     category = "deadline"
	categoryHoliday      category = "holiday"
)

type importantDate struct {
	date              time.Time
	end               time.Time // zero unless the row is a range
	description       string
	category          category
	opensRegistration bool
}

// termOption is one entry of the important dates page's term dropdown.
type termOption struct {
	nodeID   string // value of the page's field_quarter_target_id filter
	termCode string
}

// termError is a parse failure confined to one term. The term is skipped
// rather than stored from a partial parse, so its previous rows stand. code
// is "" when the failure cannot be tied to a term.
type termError struct {
	code string
	err  error
}
