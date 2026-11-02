package calendar

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"schedule-optimizer/internal/meeting"
	"schedule-optimizer/internal/store"
)

// GuessReason says why a section's final is a guess rather than certain.
type GuessReason string

const (
	ReasonLab      GuessReason = "lab"      // lecture and lab meet at different times
	ReasonMultiple GuessReason = "multiple" // more than one meeting time
	ReasonWeekly   GuessReason = "weekly"   // meets once a week; many such classes have no final
)

// Final is a section's final exam slot.
type Final struct {
	Date    string      `json:"date"`  // YYYY-MM-DD
	Start   string      `json:"start"` // HHMM
	End     string      `json:"end"`   // HHMM
	Certain bool        `json:"certain"`
	Reason  GuessReason `json:"reason,omitempty"`
}

// FinalsTable is one term's finals schedule.
type FinalsTable struct {
	slots  []finalMapping
	monday time.Time // first day of finals week
}

// LoadFinals reads the finals tables of terms in one query. A term with no
// published table (every summer, and terms not yet published) is absent.
func LoadFinals(ctx context.Context, q *store.Queries, terms []string) (map[string]*FinalsTable, error) {
	rows, err := q.GetFinalMappingsByTerms(ctx, terms)
	if err != nil {
		return nil, fmt.Errorf("loading finals: %w", err)
	}
	tables := make(map[string]*FinalsTable)
	for _, r := range rows {
		examDate, err := parseDate(r.ExamDate)
		if err != nil {
			return nil, fmt.Errorf("term %s exam date: %w", r.TermCode, err)
		}
		t := tables[r.TermCode]
		if t == nil {
			t = &FinalsTable{}
			tables[r.TermCode] = t
		}
		t.add(finalMapping{
			termCode:   r.TermCode,
			rangeStart: r.TimeRangeStart,
			rangeEnd:   r.TimeRangeEnd,
			hasTuTh:    r.HasTuth,
			examDate:   examDate,
			examStart:  r.ExamStartTime,
			examEnd:    r.ExamEndTime,
		})
	}
	return tables, nil
}

func (t *FinalsTable) add(m finalMapping) {
	t.slots = append(t.slots, m)
	monday := m.examDate.AddDate(0, 0, -int(m.examDate.Weekday()-time.Monday))
	if t.monday.IsZero() || monday.Before(t.monday) {
		t.monday = monday
	}
}

// Week returns the first and last weekday of finals week.
func (t *FinalsTable) Week() (start, end time.Time) {
	return t.monday, t.monday.AddDate(0, 0, 4)
}

// eveningStart is when the registrar's table stops: classes starting then or
// later take their exam at their own time on their first weekday of finals
// week, capped at two hours.
const (
	eveningStart   = 18 * 60
	maxExamMinutes = 120
)

// FinalFor matches a section to its exam slot, following the special
// conditions on registrar.wwu.edu/calendars/finals. It returns nil when the
// section has no final to show: no table (t is nil), lab-only, asynchronous,
// only one-off or ended meetings, or a start time the table does not cover.
// The registrar holds a lab-only class's exam at its lab time, but most labs
// have none, so a guess there would mostly be wrong.
//
// Only one weekly meeting time, on two or more days, without a lab is certain.
// The registrar bases a lecture-and-lab class on the lecture, but Banner's
// per-meeting lab codes are unreliable (some PHYS 161 sections mark the
// lecture as the lab), so those, and other multi-time sections, are guesses
// from the pattern with the most weekly minutes.
func (t *FinalsTable) FinalFor(scheduleType string, meetings []meeting.Time) *Final {
	if t == nil || scheduleType == "Lab" {
		return nil
	}
	patterns := t.patterns(meetings)
	chosen, ok := choosePattern(patterns)
	if !ok {
		return nil
	}

	final := t.slotFor(chosen)
	if final == nil {
		return nil
	}
	hasLab := slices.ContainsFunc(patterns, func(p pattern) bool { return p.lab }) || scheduleType == "Lecture/Lab"
	switch {
	case len(patterns) > 1 && hasLab, chosen.lab:
		final.Reason = ReasonLab
	case len(patterns) > 1:
		final.Reason = ReasonMultiple
	case chosen.dayCount() == 1:
		final.Reason = ReasonWeekly
	default:
		final.Certain = true
	}
	return final
}

// pattern is one weekly meeting time. Banner splits a pattern across rows
// (TR as a T row and an R row) and repeats rows, so rows sharing a time and
// lab flag merge.
type pattern struct {
	start, end int // minutes after midnight
	days       [7]bool
	lab        bool
}

func (p pattern) dayCount() int {
	n := 0
	for _, d := range p.days {
		if d {
			n++
		}
	}
	return n
}

func (p pattern) weeklyMinutes() int { return (p.end - p.start) * p.dayCount() }

func (p pattern) meetsWeekday() bool { return slices.Contains(p.days[1:6], true) }

// patterns keeps the timed, repeating meetings still running in the last
// week of classes; a short session that ended earlier has no final.
func (t *FinalsTable) patterns(meetings []meeting.Time) []pattern {
	lastWeek := t.monday.AddDate(0, 0, -7).Format(time.DateOnly)
	var out []pattern
	for _, m := range meetings {
		start, okStart := clockMinutes(m.StartTime)
		end, okEnd := clockMinutes(m.EndTime)
		if !okStart || !okEnd || end <= start || !slices.Contains(m.Days[:], true) {
			continue
		}
		if m.StartDate != "" && m.StartDate == m.EndDate {
			continue // a one-off meeting
		}
		if m.EndDate != "" && m.EndDate < lastWeek {
			continue
		}
		i := slices.IndexFunc(out, func(p pattern) bool {
			return p.start == start && p.end == end && p.lab == m.IsLab()
		})
		if i < 0 {
			out = append(out, pattern{start: start, end: end, lab: m.IsLab()})
			i = len(out) - 1
		}
		for d, on := range m.Days {
			out[i].days[d] = out[i].days[d] || on
		}
	}
	return out
}

// choosePattern picks the weekday pattern with the most weekly minutes,
// preferring non-lab patterns. Ties go to the earlier start so the pick does
// not depend on row order.
func choosePattern(patterns []pattern) (pattern, bool) {
	var best pattern
	found := false
	better := func(p pattern) bool {
		switch {
		case !found:
			return true
		case p.lab != best.lab:
			return !p.lab
		case p.weeklyMinutes() != best.weeklyMinutes():
			return p.weeklyMinutes() > best.weeklyMinutes()
		default:
			return p.start < best.start
		}
	}
	for _, p := range patterns {
		if p.meetsWeekday() && better(p) {
			best, found = p, true
		}
	}
	return best, found
}

func (t *FinalsTable) slotFor(p pattern) *Final {
	if p.start >= eveningStart {
		weekday := slices.Index(p.days[1:6], true) // 0 = Monday
		return &Final{
			Date:  t.monday.AddDate(0, 0, weekday).Format(time.DateOnly),
			Start: hhmm(p.start),
			End:   hhmm(min(p.end, p.start+maxExamMinutes)),
		}
	}
	tuth := p.days[time.Tuesday] || p.days[time.Thursday]
	start := hhmm(p.start)
	for _, s := range t.slots {
		if s.hasTuTh == tuth && s.rangeStart <= start && start <= s.rangeEnd {
			return &Final{Date: s.examDate.Format(time.DateOnly), Start: s.examStart, End: s.examEnd}
		}
	}
	return nil
}

// clockMinutes parses Banner's HHMM.
func clockMinutes(s string) (int, bool) {
	if len(s) != 4 {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 || n/100 > 23 || n%100 > 59 {
		return 0, false
	}
	return n/100*60 + n%100, true
}
