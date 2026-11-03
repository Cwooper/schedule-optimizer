package calendar

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"schedule-optimizer/internal/store"
)

// TermCalendar is one term's dates for display and calendar export. Dates are
// YYYY-MM-DD, and every end is inclusive.
type TermCalendar struct {
	Term              string              `json:"term"`
	Start             string              `json:"start"`
	End               string              `json:"end"`
	RegistrationOpens string              `json:"registrationOpens"`
	Estimated         bool                `json:"estimated"`
	Registration      []RegistrationPhase `json:"registration"`
	Holidays          []Holiday           `json:"holidays"`
	Deadlines         []Deadline          `json:"deadlines"`
	Finals            *DateRange          `json:"finals"`
}

type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type RegistrationPhase struct {
	Phase string `json:"phase"` // "I", "II", or "III"
	Name  string `json:"name"`
	DateRange
}

type Holiday struct {
	Name string `json:"name"`
	DateRange
}

type DeadlineKind string

const (
	DeadlineAdd      DeadlineKind = "add"
	DeadlineDrop     DeadlineKind = "drop"
	DeadlineWithdraw DeadlineKind = "withdraw"
	DeadlineOther    DeadlineKind = "other"
)

type Deadline struct {
	Name  string       `json:"name"`
	Title string       `json:"title"` // short form of Name, for calendar events
	Date  string       `json:"date"`
	Kind  DeadlineKind `json:"kind"`
}

// LoadTerm assembles a term's calendar from the scraped rows, falling back
// to estimated term and registration dates. Holidays outside the term's
// first-to-last day are dropped (the page lists Labor Day and Winter Break
// under Fall). ok is false for an invalid term code.
func LoadTerm(ctx context.Context, q *store.Queries, code string) (cal *TermCalendar, ok bool, err error) {
	schedule, err := LoadSchedule(ctx, q)
	if err != nil {
		return nil, false, err
	}
	dates, ok := schedule.Dates(code)
	if !ok {
		return nil, false, nil
	}
	rows, err := q.GetImportantDatesByTerm(ctx, code)
	if err != nil {
		return nil, false, fmt.Errorf("loading important dates: %w", err)
	}
	finals, err := LoadFinals(ctx, q, []string{code})
	if err != nil {
		return nil, false, err
	}

	lastDay := dates.End.AddDate(0, 0, -1)
	cal = &TermCalendar{
		Term:              code,
		Start:             dates.Start.Format(time.DateOnly),
		End:               lastDay.Format(time.DateOnly),
		RegistrationOpens: dates.RegistrationOpens.Format(time.DateOnly),
		Estimated:         dates.Estimated,
		Registration:      []RegistrationPhase{},
		Holidays:          []Holiday{},
		Deadlines:         []Deadline{},
	}
	if t := finals[code]; t != nil {
		start, end := t.Week()
		cal.Finals = &DateRange{Start: start.Format(time.DateOnly), End: end.Format(time.DateOnly)}
	}

	seenPhase := make(map[string]bool)
	for _, r := range rows {
		span := DateRange{Start: r.Date, End: r.Date}
		if r.EndDate.Valid {
			span.End = r.EndDate.String
		}
		switch category(r.Category) {
		case categoryRegistration:
			// Rows are ordered by date, so the first row per phase is its opening.
			if phase := registrationPhase(r.Description); phase != "" && !seenPhase[phase] {
				seenPhase[phase] = true
				cal.Registration = append(cal.Registration, RegistrationPhase{Phase: phase, Name: r.Description, DateRange: span})
			}
		case categoryHoliday:
			if span.End >= cal.Start && span.Start <= cal.End {
				cal.Holidays = append(cal.Holidays, Holiday{Name: r.Description, DateRange: span})
			}
		case categoryDeadline:
			kind, title := classifyDeadline(r.Description)
			cal.Deadlines = append(cal.Deadlines, Deadline{Name: r.Description, Title: title, Date: r.Date, Kind: kind})
		case categoryAcademic:
			// The page's own finals week wins over the one inferred from the table.
			if finalExamsRegex.MatchString(r.Description) {
				cal.Finals = &span
			}
		}
	}
	return cal, true, nil
}

var (
	phaseRegex      = regexp.MustCompile(`(?i)^phase\s+(iii|ii|i|1|2|3)\b`)
	finalExamsRegex = regexp.MustCompile(`(?i)^final exam`)
	romanPhases     = map[string]string{"1": "I", "2": "II", "3": "III"}
)

func registrationPhase(description string) string {
	m := phaseRegex.FindStringSubmatch(description)
	if m == nil {
		return ""
	}
	if roman, ok := romanPhases[m[1]]; ok {
		return roman
	}
	return strings.ToUpper(m[1])
}

// deadlineRules name the registrar's deadlines by their wording; the first
// rule whose phrases all appear wins. Order matters: "withdraw from WWU or
// drop a class with full refund" is the refund deadline, not the withdrawal
// one, and "add or drop ... late fee" is neither the add nor the drop one.
// Summer words its fee "late add/drop service fee".
var deadlineRules = []struct {
	phrases []string // lowercase
	kind    DeadlineKind
	title   string
}{
	{[]string{"full refund"}, DeadlineDrop, "Full refund deadline"},
	{[]string{"half", "refund"}, DeadlineDrop, "Half refund deadline"},
	{[]string{"add or drop", "fee"}, DeadlineDrop, "Late fee starts after"},
	{[]string{"drop", "grade"}, DeadlineDrop, "Drop deadline (no W)"},
	{[]string{"withdraw from wwu"}, DeadlineWithdraw, "WWU withdrawal deadline"},
	{[]string{"withdraw from a class"}, DeadlineWithdraw, "Withdraw deadline"},
	{[]string{"to add a class"}, DeadlineAdd, "Add deadline"},
}

// maxTitleLen caps a fallback title; the full text goes in the description.
const maxTitleLen = 60

// sessionRegex finds the session of a summer deadline, which the page
// repeats once per session: "Last day to withdraw from a class (Six-week
// Session)".
var sessionRegex = regexp.MustCompile(`(?i)(\([^()]*\bsession\))\s*$`)

// classifyDeadline sorts the page's free-text deadlines and names each
// briefly, keeping a summer session so its deadlines stay apart.
func classifyDeadline(description string) (DeadlineKind, string) {
	kind, title := matchDeadline(description)
	if m := sessionRegex.FindStringSubmatch(description); m != nil {
		title += " " + m[1]
	}
	return kind, title
}

// matchDeadline applies deadlineRules. Wording no rule knows still sorts by
// its verb, so a reworded deadline keeps its kind, and is titled by its own
// text, shortened.
func matchDeadline(description string) (DeadlineKind, string) {
	lower := strings.ToLower(description)
	for _, r := range deadlineRules {
		if !slices.ContainsFunc(r.phrases, func(p string) bool { return !strings.Contains(lower, p) }) {
			return r.kind, r.title
		}
	}
	kind := DeadlineOther
	switch {
	case strings.Contains(lower, "drop"):
		kind = DeadlineDrop
	case strings.Contains(lower, "withdraw"):
		kind = DeadlineWithdraw
	case strings.Contains(lower, "to add"):
		kind = DeadlineAdd
	}
	return kind, shortTitle(description)
}

var trailingNoteRegex = regexp.MustCompile(`\s*\([^()]*\)$`)

// shortTitle keeps the first sentence of text without a trailing
// parenthetical, cut at a word boundary past maxTitleLen.
func shortTitle(text string) string {
	title := strings.TrimSpace(text)
	if i := strings.Index(title, ". "); i >= 0 {
		title = title[:i]
	}
	title = trailingNoteRegex.ReplaceAllString(title, "")
	if len(title) <= maxTitleLen {
		return title
	}
	cut := strings.LastIndex(title[:maxTitleLen+1], " ")
	if cut <= 0 {
		return title
	}
	return strings.TrimRight(title[:cut], " ,;:-") + "..."
}
