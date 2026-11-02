package calendar

import (
	"context"
	"reflect"
	"testing"
)

// seedWinter2027 stores Winter 2027 as the registrar fixtures publish it.
func seedWinter2027(t *testing.T, svc *Service) {
	t.Helper()
	ctx := context.Background()
	dates, _, err := parseTermDates(readFixture(t, "term_dates.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range dates {
		if d.termCode == "202710" {
			if err := svc.storeTermDates(ctx, []termDates{d}); err != nil {
				t.Fatal(err)
			}
		}
	}
	important, err := parseImportantDates(readFixture(t, "important_dates_202710.html"), "202710")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.replaceImportantDates(ctx, "202710", important); err != nil {
		t.Fatal(err)
	}
	var finals []finalMapping
	for _, m := range parseFinalsFixture(t) {
		if m.termCode == "202710" {
			finals = append(finals, m)
		}
	}
	if err := svc.replaceFinals(ctx, "202710", finals); err != nil {
		t.Fatal(err)
	}
}

func TestLoadTerm(t *testing.T) {
	svc, queries := newTestService(t, "202710")
	seedWinter2027(t, svc)

	cal, ok, err := LoadTerm(context.Background(), queries, "202710")
	if err != nil || !ok {
		t.Fatalf("LoadTerm: ok %v, err %v", ok, err)
	}

	if cal.Start != "2027-01-05" || cal.End != "2027-03-19" || cal.RegistrationOpens != "2026-11-03" || cal.Estimated {
		t.Errorf("dates = %s to %s, registration %s, estimated %v", cal.Start, cal.End, cal.RegistrationOpens, cal.Estimated)
	}
	wantPhases := []RegistrationPhase{
		{Phase: "I", DateRange: DateRange{"2026-11-03", "2026-11-16"}},
		{Phase: "II", DateRange: DateRange{"2026-11-17", "2027-01-04"}},
		{Phase: "III", DateRange: DateRange{"2027-01-05", "2027-01-11"}},
	}
	if !reflect.DeepEqual(cal.Registration, wantPhases) {
		t.Errorf("registration = %+v, want %+v", cal.Registration, wantPhases)
	}
	// Spring Break starts the day after the term ends, so it is not listed.
	var holidays []string
	for _, h := range cal.Holidays {
		holidays = append(holidays, h.Start+" "+h.Name)
	}
	wantHolidays := []string{"2027-01-18 Martin Luther King Jr. Day", "2027-02-15 Presidents' Day"}
	if !reflect.DeepEqual(holidays, wantHolidays) {
		t.Errorf("holidays = %q, want %q", holidays, wantHolidays)
	}
	if cal.Finals == nil || *cal.Finals != (DateRange{"2027-03-15", "2027-03-19"}) {
		t.Errorf("finals = %+v, want Mar 15-19", cal.Finals)
	}

	kinds := make(map[DeadlineKind]int)
	for _, d := range cal.Deadlines {
		kinds[d.Kind]++
	}
	// Add without an override; drop with full refund, without a fee, without
	// a W, and for half refund; withdraw from a class and from WWU.
	if kinds[DeadlineAdd] != 1 || kinds[DeadlineDrop] != 4 || kinds[DeadlineWithdraw] != 2 {
		t.Errorf("deadline kinds = %v", kinds)
	}
}

func TestLoadTermFinalsWeekPrefersThePage(t *testing.T) {
	svc, queries := newTestService(t, "202710")
	seedWinter2027(t, svc)
	// The table's exams fall Mar 15-19; a page row that disagrees wins.
	res, err := svc.db.Exec(`UPDATE important_dates SET date = '2027-03-16', end_date = '2027-03-20'
		WHERE term_code = '202710' AND description LIKE 'Final Exam%'`)
	if err != nil {
		t.Fatal(err)
	}
	if n, _ := res.RowsAffected(); n != 1 {
		t.Fatalf("updated %d final exams rows, want 1", n)
	}

	cal, _, err := LoadTerm(context.Background(), queries, "202710")
	if err != nil {
		t.Fatal(err)
	}
	if cal.Finals == nil || *cal.Finals != (DateRange{"2027-03-16", "2027-03-20"}) {
		t.Errorf("finals = %+v, want the page's Mar 16-20", cal.Finals)
	}
}

func TestLoadTermEstimated(t *testing.T) {
	_, queries := newTestService(t, "202730")

	cal, ok, err := LoadTerm(context.Background(), queries, "202730")
	if err != nil || !ok {
		t.Fatalf("LoadTerm: ok %v, err %v", ok, err)
	}
	// Summer's estimate: Jun 23 - Aug 21, registration 56 days ahead.
	want := TermCalendar{
		Term: "202730", Start: "2027-06-23", End: "2027-08-21", RegistrationOpens: "2027-04-28", Estimated: true,
		Registration: []RegistrationPhase{}, Holidays: []Holiday{}, Deadlines: []Deadline{},
	}
	if !reflect.DeepEqual(*cal, want) {
		t.Errorf("LoadTerm = %+v, want %+v", *cal, want)
	}

	if _, ok, err := LoadTerm(context.Background(), queries, "bogus"); ok || err != nil {
		t.Errorf("invalid code: ok %v, err %v", ok, err)
	}
}

func TestDeadlineKind(t *testing.T) {
	tests := map[string]DeadlineKind{
		"Last day to add a class without an override":                                         DeadlineAdd,
		"Last day to add or drop a class without incurring a $10 per credit late fee":         DeadlineDrop,
		"Last day to withdraw from WWU or drop a class with full refund (ends at 11:59 p.m.)": DeadlineDrop,
		"Last day to withdraw from a class. Online registration via Web4U ends at 11:59 p.m.": DeadlineWithdraw,
		"Last day to change to-or-from Audit":                                                 DeadlineOther,
		"Tuition and fees, housing and meal plan charges due":                                 DeadlineOther,
	}
	for description, want := range tests {
		if got := deadlineKind(description); got != want {
			t.Errorf("deadlineKind(%q) = %s, want %s", description, got, want)
		}
	}
}
