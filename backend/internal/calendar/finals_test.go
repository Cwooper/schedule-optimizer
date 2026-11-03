package calendar

import (
	"testing"

	"schedule-optimizer/internal/meeting"
)

func fallFinals(t *testing.T) *FinalsTable {
	t.Helper()
	table := &FinalsTable{}
	for _, m := range parseFinalsFixture(t) {
		if m.termCode == "202640" {
			table.add(m)
		}
	}
	return table
}

// days builds a meeting's day flags from Banner letters (U M T W R F S).
func days(letters string) [7]bool {
	var d [7]bool
	for _, l := range letters {
		for i, c := range "UMTWRFS" {
			if l == c {
				d[i] = true
			}
		}
	}
	return d
}

// mt is a term-long Fall 2026 meeting.
func mt(dayLetters, start, end string) meeting.Time {
	return meeting.Time{Days: days(dayLetters), StartTime: start, EndTime: end, StartDate: "2026-09-23", EndDate: "2026-12-11", ScheduleType: "L"}
}

func lab(m meeting.Time) meeting.Time { return typed(m, "B") }

func typed(m meeting.Time, scheduleType string) meeting.Time {
	m.ScheduleType = scheduleType
	return m
}

func dated(m meeting.Time, start, end string) meeting.Time {
	m.StartDate, m.EndDate = start, end
	return m
}

func TestFinalFor(t *testing.T) {
	table := fallFinals(t)
	certain := func(date, start, end string) *Final {
		return &Final{Date: date, Start: start, End: end, Certain: true}
	}
	guess := func(date, start, end string, reason GuessReason) *Final {
		return &Final{Date: date, Start: start, End: end, Reason: reason}
	}

	tests := []struct {
		name         string
		scheduleType string
		meetings     []meeting.Time
		want         *Final
	}{
		{"MWF lecture", "Lecture", []meeting.Time{mt("MWF", "0900", "0950")},
			certain("2026-12-07", "1030", "1230")},
		{"Tuesday or Thursday picks the other column", "Lecture", []meeting.Time{mt("MTWF", "0900", "0950")},
			certain("2026-12-07", "0800", "1000")},
		{"TR split across rows merges", "Lecture", []meeting.Time{mt("T", "1400", "1550"), mt("R", "1400", "1550")},
			certain("2026-12-10", "0800", "1000")},
		{"repeated row merges", "Lecture", []meeting.Time{mt("MW", "1000", "1150"), mt("MW", "1000", "1150")},
			certain("2026-12-08", "1030", "1230")},
		{"range end is inclusive", "Lecture", []meeting.Time{mt("MWF", "0929", "1020")},
			certain("2026-12-07", "1030", "1230")},
		{"an async row beside a timed one is ignored", "Lecture", []meeting.Time{mt("MWF", "0900", "0950"), {}},
			certain("2026-12-07", "1030", "1230")},

		{"lab meeting: final follows the lecture", "Lecture", []meeting.Time{mt("MWF", "1300", "1350"), lab(mt("R", "0800", "0950"))},
			guess("2026-12-08", "1530", "1730", ReasonLab)},
		{"lecture/lab without lab codes takes the most weekly minutes", "Lecture/Lab",
			[]meeting.Time{mt("T", "1100", "1250"), mt("MWF", "0800", "0850")},
			guess("2026-12-07", "1530", "1730", ReasonLab)},
		{"only lab-coded patterns", "Lecture", []meeting.Time{lab(mt("MW", "1000", "1150"))},
			guess("2026-12-08", "1030", "1230", ReasonLab)},
		// MIS 320 CRN 10105: an L row for MW and two C rows (two rooms) for R.
		{"lecture rows set the pattern beside other types at the same time", "Lecture/Lab",
			[]meeting.Time{mt("MW", "1000", "1120"), typed(mt("R", "1000", "1120"), "C"), typed(mt("R", "1000", "1120"), "C")},
			guess("2026-12-08", "1030", "1230", ReasonLab)},
		// MIS 320 CRN 41981: as above but the C rows are on F, which does not
		// move the exam.
		{"other types at the lecture's time are one time when the exam is the same", "Lecture/Lab",
			[]meeting.Time{mt("MW", "1000", "1120"), typed(mt("F", "1000", "1120"), "C"), typed(mt("F", "1000", "1120"), "C")},
			certain("2026-12-08", "1030", "1230")},
		// KIN 416 CRN 40471 and MFGE 454 CRN 42743: the lecture has a Tuesday
		// or Thursday either way.
		{"another type adding a day keeps the lecture's column", "Lecture/Lab",
			[]meeting.Time{mt("MWR", "0800", "0850"), typed(mt("F", "0800", "0850"), "C")},
			certain("2026-12-07", "1300", "1500")},
		{"another type adding a Tuesday to a Thursday lecture", "Lecture/Lab",
			[]meeting.Time{typed(mt("T", "1500", "1650"), "C"), mt("WR", "1500", "1650")},
			certain("2026-12-09", "1300", "1500")},
		{"another type moving an evening exam's day stays a guess", "Lecture",
			[]meeting.Time{mt("WR", "1900", "2150"), typed(mt("M", "1900", "2150"), "C")},
			guess("2026-12-09", "1900", "2100", ReasonMultiple)},
		{"another type on the lecture's one day is still weekly", "Lecture",
			[]meeting.Time{mt("W", "1400", "1650"), typed(mt("W", "1400", "1650"), "C")},
			guess("2026-12-10", "1030", "1230", ReasonWeekly)},
		{"a lab-coded row at the lecture's time stays apart", "Lecture",
			[]meeting.Time{mt("MW", "1000", "1120"), lab(mt("F", "1000", "1120"))},
			guess("2026-12-08", "1030", "1230", ReasonLab)},
		{"lecture rows win over longer rows of another type", "Lecture",
			[]meeting.Time{mt("MW", "1000", "1050"), typed(mt("MTWRF", "1300", "1350"), "C")},
			guess("2026-12-08", "1030", "1230", ReasonMultiple)},
		{"one pattern under a lecture/lab type is certain", "Lecture/Lab", []meeting.Time{mt("MWF", "0900", "0950")},
			certain("2026-12-07", "1030", "1230")},
		{"multiple times take the most weekly minutes", "Lecture", []meeting.Time{mt("W", "0800", "0950"), mt("MTWF", "1200", "1250")},
			guess("2026-12-11", "0800", "1000", ReasonMultiple)},
		{"equal minutes tie to the earlier start", "Seminar", []meeting.Time{mt("T", "1400", "1550"), mt("R", "1000", "1150")},
			guess("2026-12-08", "0800", "1000", ReasonMultiple)},
		{"once a week", "Seminar", []meeting.Time{mt("W", "1400", "1650")},
			guess("2026-12-10", "1030", "1230", ReasonWeekly)},

		{"evening class: first weekday at its own time", "Lecture", []meeting.Time{mt("TR", "1800", "1920")},
			certain("2026-12-08", "1800", "1920")},
		{"evening exam capped at two hours", "Lecture", []meeting.Time{mt("TR", "1800", "2100")},
			certain("2026-12-08", "1800", "2000")},
		{"once-a-week evening class", "Lecture", []meeting.Time{mt("W", "1900", "2150")},
			guess("2026-12-09", "1900", "2100", ReasonWeekly)},
		{"weekend pattern loses to the weekday one", "Seminar", []meeting.Time{mt("S", "1200", "1650"), mt("M", "1800", "2150")},
			guess("2026-12-07", "1800", "2000", ReasonMultiple)},
		{"4:30-5:59 neither Tuesday nor Thursday", "Lecture", []meeting.Time{mt("MW", "1700", "1820")},
			certain("2026-12-07", "1800", "2000")},
		{"4:30-5:59 with Tuesday or Thursday", "Lecture", []meeting.Time{mt("TR", "1630", "1750")},
			certain("2026-12-08", "1800", "2000")},
		{"ends in the last week of classes", "Lecture", []meeting.Time{dated(mt("MWF", "0900", "0950"), "2026-09-23", "2026-12-04")},
			certain("2026-12-07", "1030", "1230")},

		{"lab-only section", "Lab", []meeting.Time{lab(mt("TR", "1000", "1150"))}, nil},
		{"asynchronous", "Lecture", []meeting.Time{{}}, nil},
		{"no meetings", "Lecture", nil, nil},
		{"one-off meeting", "Lecture", []meeting.Time{dated(mt("S", "0900", "1600"), "2026-10-10", "2026-10-10")}, nil},
		{"starts after the last week of classes", "Lecture", []meeting.Time{dated(mt("MW", "0900", "0950"), "2026-12-14", "2026-12-18")}, nil},
		{"short session ends before finals week", "Lecture", []meeting.Time{dated(mt("MTWRF", "0900", "0950"), "2026-09-23", "2026-10-30")}, nil},
		{"weekend only", "Seminar", []meeting.Time{mt("S", "0900", "1150")}, nil},
		{"before the table's first row", "Lecture", []meeting.Time{mt("MWF", "0730", "0820")}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := table.FinalFor(tt.scheduleType, tt.meetings)
			switch {
			case got == nil && tt.want == nil:
			case got == nil || tt.want == nil || *got != *tt.want:
				t.Errorf("FinalFor = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestFinalForWithoutTable(t *testing.T) {
	var summer *FinalsTable
	if got := summer.FinalFor("Lecture", []meeting.Time{mt("MWF", "0900", "0950")}); got != nil {
		t.Errorf("FinalFor without a finals table = %+v, want nil", got)
	}
}
