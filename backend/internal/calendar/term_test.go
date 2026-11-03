package calendar

import (
	"context"
	"reflect"
	"strings"
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
	var phases []string
	for _, r := range cal.Registration {
		phases = append(phases, r.Phase+" "+r.Start+" "+r.End)
	}
	wantPhases := []string{"I 2026-11-03 2026-11-16", "II 2026-11-17 2027-01-04", "III 2027-01-05 2027-01-11"}
	if !reflect.DeepEqual(phases, wantPhases) {
		t.Errorf("registration = %q, want %q", phases, wantPhases)
	}
	if name := cal.Registration[0].Name; !strings.HasPrefix(name, "Phase I - registration for continuing") {
		t.Errorf("Phase I name = %q, want the registrar's wording", name)
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

func TestClassifyDeadline(t *testing.T) {
	tests := []struct {
		description string
		kind        DeadlineKind
		title       string
	}{
		{"Last day to add a class without an override", DeadlineAdd, "Add deadline"},
		{"Last day to add or drop a class without incurring a $10 per credit late fee", DeadlineDrop, "Late fee starts after"},
		{`Last day to drop a class without incurring a "W" grade on official transcript ("XM" code appears on unofficial academic history only)`,
			DeadlineDrop, "Drop deadline (no W)"},
		{"Last day to withdraw from WWU or drop a class with full refund (ends at 11:59 p.m.)", DeadlineDrop, "Full refund deadline"},
		{"Last day for half-tuition refund when dropping a class (fees are not refunded)", DeadlineDrop, "Half refund deadline"},
		{"Last day to withdraw from a class. Online registration via Web4U ends at 11:59 p.m.", DeadlineWithdraw, "Withdraw deadline"},
		{"Last day to withdraw from WWU - for fall quarter only", DeadlineWithdraw, "WWU withdrawal deadline"},
		// Summer's wording.
		{"Last day to withdraw for the quarter or drop a class with full refund (ends at 11:59 p.m.)", DeadlineDrop, "Full refund deadline"},
		{"Last day to add or drop a class on an existing schedule without incurring a $10 per credit late add/drop service fee.",
			DeadlineDrop, "Late fee starts after"},
		{"Last day to drop a class without incurring a \u201cW\u201d grade on official transcript", DeadlineDrop, "Drop deadline (no W)"},
		{"Last day to withdraw from a class (Six-week Session)", DeadlineWithdraw, "Withdraw deadline (Six-week Session)"},
		{"Last day to withdraw from WWU (Nine-week Session)", DeadlineWithdraw, "WWU withdrawal deadline (Nine-week Session)"},
		{"Last day to change to-or-from Pass/No Pass (Nine-week Session)", DeadlineOther, "Last day to change to-or-from Pass/No Pass (Nine-week Session)"},

		// Unknown wording sorts by its verb and keeps its own text, shortened.
		{"Last day to drop a lab section. Fees apply after", DeadlineDrop, "Last day to drop a lab section"},
		{"Last day to withdraw from housing (ends at 5 p.m.)", DeadlineWithdraw, "Last day to withdraw from housing"},
		{"Last day to change to-or-from Audit", DeadlineOther, "Last day to change to-or-from Audit"},
		{"Waitlisting ends at 11:59 p.m.", DeadlineOther, "Waitlisting ends at 11:59 p.m."},
		{"Priority date for readmission after being away one quarter or longer other than summer quarter",
			DeadlineOther, "Priority date for readmission after being away one quarter..."},
	}
	for _, tt := range tests {
		kind, title := classifyDeadline(tt.description)
		if kind != tt.kind || title != tt.title {
			t.Errorf("classifyDeadline(%q) = %s, %q; want %s, %q", tt.description, kind, title, tt.kind, tt.title)
		}
	}
}
