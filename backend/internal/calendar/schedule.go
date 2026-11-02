package calendar

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	_ "time/tzdata" // the deploy box may lack a zoneinfo database

	"schedule-optimizer/internal/store"
	"schedule-optimizer/internal/termcode"
)

// Pacific is WWU's time zone. The registrar publishes bare local dates and
// the server runs elsewhere, so every calendar date is midnight Pacific.
var Pacific = mustLoadLocation("America/Los_Angeles")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

const dateLayout = "2006-01-02"

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, Pacific)
}

func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation(dateLayout, s, Pacific)
}

// Phase is where a term is in its lifecycle, which sets how often its
// sections are scraped and which term the frontend opens to.
type Phase int

const (
	PhasePreRegistration Phase = iota // before Phase I registration opens
	PhaseRegistration                 // Phase I through the first week of classes
	PhaseInSession                    // classes underway, registration settled
	PhasePast                         // term over
)

func (p Phase) String() string {
	switch p {
	case PhasePreRegistration:
		return "pre-registration"
	case PhaseRegistration:
		return "registration"
	case PhaseInSession:
		return "in-session"
	case PhasePast:
		return "past"
	}
	return fmt.Sprintf("Phase(%d)", int(p))
}

// registrationGraceDays keeps a term in registration through its first
// week, while Phase III open registration and adding without an override
// run, so seats still churn. Drops without a W run about five days longer
// and get the daily cadence.
const registrationGraceDays = 7

// TermDates are a term's phase boundaries, each at midnight Pacific.
type TermDates struct {
	RegistrationOpens time.Time
	Start             time.Time
	End               time.Time // midnight after the last day, exclusive
	Estimated         bool      // some boundary fell back to an estimate
}

func (d TermDates) Phase(now time.Time) Phase {
	switch {
	case !now.Before(d.End):
		return PhasePast
	case !now.Before(d.Start.AddDate(0, 0, registrationGraceDays)):
		return PhaseInSession
	case !now.Before(d.RegistrationOpens):
		return PhaseRegistration
	default:
		return PhasePreRegistration
	}
}

type estimate struct {
	startMonth time.Month
	startDay   int
	endMonth   time.Month
	endDay     int
	leadDays   int // registration opens this many days before the start
}

// estimates stand in for dates the registrar has not published, or that a
// fresh database has not scraped yet. Leads are the 2026-27 Phase I leads
// (63, 35, 49, 134 days) plus a week: opening early only costs extra
// scrapes, while opening late serves stale seat counts mid-registration.
var estimates = map[int]estimate{
	termcode.Winter: {time.January, 5, time.March, 20, 70},
	termcode.Spring: {time.April, 1, time.June, 12, 42},
	termcode.Summer: {time.June, 23, time.August, 21, 56},
	termcode.Fall:   {time.September, 24, time.December, 12, 141},
}

// Schedule resolves term dates from scraped data, falling back per boundary
// to estimates.
type Schedule struct {
	scraped map[string]scrapedDates
}

// scrapedDates holds what the registrar published; zero fields are unknown.
type scrapedDates struct {
	start, lastDay, registrationOpens time.Time
}

// LoadSchedule reads every term's scraped dates in one query.
func LoadSchedule(ctx context.Context, q *store.Queries) (Schedule, error) {
	rows, err := q.GetTermSchedules(ctx)
	if err != nil {
		return Schedule{}, fmt.Errorf("loading term schedules: %w", err)
	}
	s := Schedule{scraped: make(map[string]scrapedDates, len(rows))}
	for _, r := range rows {
		var d scrapedDates
		if d.start, err = parseNullDate(r.StartDate); err != nil {
			return Schedule{}, fmt.Errorf("term %s start: %w", r.TermCode, err)
		}
		if d.lastDay, err = parseNullDate(r.EndDate); err != nil {
			return Schedule{}, fmt.Errorf("term %s end: %w", r.TermCode, err)
		}
		if d.registrationOpens, err = parseNullDate(r.RegistrationOpens); err != nil {
			return Schedule{}, fmt.Errorf("term %s registration: %w", r.TermCode, err)
		}
		s.scraped[r.TermCode] = d
	}
	return s, nil
}

func parseNullDate(s sql.NullString) (time.Time, error) {
	if !s.Valid {
		return time.Time{}, nil
	}
	return parseDate(s.String)
}

// Dates returns a term's boundaries. ok is false only for an invalid code.
func (s Schedule) Dates(code string) (d TermDates, ok bool) {
	year, quarter, err := termcode.Parse(code)
	if err != nil {
		return TermDates{}, false
	}
	e := estimates[quarter]
	known := s.scraped[code]

	if !known.start.IsZero() && !known.lastDay.IsZero() {
		d.Start, d.End = known.start, known.lastDay.AddDate(0, 0, 1)
	} else {
		d.Start, d.End = date(year, e.startMonth, e.startDay), date(year, e.endMonth, e.endDay+1)
		d.Estimated = true
	}

	if !known.registrationOpens.IsZero() && known.registrationOpens.Before(d.Start) {
		d.RegistrationOpens = known.registrationOpens
	} else {
		d.RegistrationOpens = d.Start.AddDate(0, 0, -e.leadDays)
		d.Estimated = true
	}
	return d, true
}

// Phase treats an invalid code as past so nothing ever scrapes it as current.
func (s Schedule) Phase(code string, now time.Time) Phase {
	d, ok := s.Dates(code)
	if !ok {
		return PhasePast
	}
	return d.Phase(now)
}

// DefaultTerm picks the term the frontend should open to: the newest term in
// registration, since students are choosing its classes (in May that is Fall
// over Summer); else the term in session; else the soonest upcoming term;
// else the newest term. Returns "" for no codes.
func (s Schedule) DefaultTerm(codes []string, now time.Time) string {
	var registering, inSession, upcoming, newest string
	for _, code := range codes {
		newest = max(newest, code)
		switch s.Phase(code, now) {
		case PhaseRegistration:
			registering = max(registering, code)
		case PhaseInSession:
			inSession = max(inSession, code)
		case PhasePreRegistration:
			if upcoming == "" || code < upcoming {
				upcoming = code
			}
		}
	}
	for _, pick := range []string{registering, inSession, upcoming} {
		if pick != "" {
			return pick
		}
	}
	return newest
}
