package calendar

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"schedule-optimizer/internal/store"
	"schedule-optimizer/internal/testutil"
)

func newTestService(t *testing.T, terms ...string) (*Service, *store.Queries) {
	t.Helper()
	db, queries := testutil.SetupTestDB(t)
	for _, code := range terms {
		if err := queries.UpsertTerm(context.Background(), store.UpsertTermParams{Code: code, Description: code}); err != nil {
			t.Fatalf("seed term: %v", err)
		}
	}
	return &Service{db: db, queries: queries}, queries
}

func TestReplaceImportantDatesRemovesStaleRows(t *testing.T) {
	ctx := context.Background()
	svc, queries := newTestService(t, "202710", "202640")

	first := []importantDate{
		{date: date(2026, 11, 3), end: date(2026, 11, 16), description: "Phase I - registration", category: categoryRegistration, opensRegistration: true},
		{date: date(2027, 1, 18), description: "MLK Day", category: categoryHoliday},
	}
	if err := svc.replaceImportantDates(ctx, "202710", first); err != nil {
		t.Fatal(err)
	}
	other := []importantDate{{date: date(2026, 9, 7), description: "Labor Day", category: categoryHoliday}}
	if err := svc.replaceImportantDates(ctx, "202640", other); err != nil {
		t.Fatal(err)
	}

	renamed := []importantDate{
		{date: date(2026, 11, 3), end: date(2026, 11, 16), description: "Phase I - registration (revised)", category: categoryRegistration, opensRegistration: true},
		{date: date(2027, 1, 18), description: "MLK Day", category: categoryHoliday},
		{date: date(2027, 1, 18), description: "MLK Day", category: categoryHoliday}, // the page repeating a row
	}
	if err := svc.replaceImportantDates(ctx, "202710", renamed); err != nil {
		t.Fatal(err)
	}

	rows, err := queries.GetImportantDatesByTerm(ctx, "202710")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, r := range rows {
		got = append(got, r.Date+" "+r.Description)
	}
	want := []string{"2026-11-03 Phase I - registration (revised)", "2027-01-18 MLK Day"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("rows = %q, want %q", got, want)
	}
	if !rows[0].OpensRegistration || rows[0].EndDate != (sql.NullString{String: "2026-11-16", Valid: true}) {
		t.Errorf("phase I row = %+v, want registration opening ending 2026-11-16", rows[0])
	}

	untouched, err := queries.GetImportantDatesByTerm(ctx, "202640")
	if err != nil {
		t.Fatal(err)
	}
	if len(untouched) != 1 {
		t.Errorf("other term has %d rows, want 1", len(untouched))
	}
}

func TestReplaceFinalsRollsBackOnFailure(t *testing.T) {
	ctx := context.Background()
	svc, queries := newTestService(t, "202640")

	slot := finalMapping{rangeStart: "0800", rangeEnd: "0829", hasTuTh: true, examDate: date(2026, 12, 7), examStart: "1300", examEnd: "1500"}
	if err := svc.replaceFinals(ctx, "202640", []finalMapping{slot}); err != nil {
		t.Fatal(err)
	}

	moved := slot
	moved.examStart, moved.examEnd = "0800", "1000"
	// The second row violates UNIQUE(term_code, time_range_start, has_tuth).
	if err := svc.replaceFinals(ctx, "202640", []finalMapping{moved, moved}); err == nil {
		t.Fatal("want error for duplicate slot")
	}

	rows, err := queries.GetFinalMappingsByTerm(ctx, "202640")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ExamStartTime != "1300" {
		t.Errorf("rows = %+v, want the original 1300 slot kept", rows)
	}
}

func TestLoadSchedule(t *testing.T) {
	ctx := context.Background()
	svc, queries := newTestService(t, "202640", "202710", "202720")

	if err := svc.storeTermDates(ctx, []termDates{
		{termCode: "202640", start: date(2026, 9, 23), end: date(2026, 12, 11)},
		{termCode: "202710", start: date(2027, 1, 5), end: date(2027, 3, 19)},
	}); err != nil {
		t.Fatal(err)
	}
	if err := svc.replaceImportantDates(ctx, "202710", []importantDate{
		{date: date(2026, 11, 3), description: "Phase I", category: categoryRegistration, opensRegistration: true},
		{date: date(2026, 11, 17), description: "Phase II", category: categoryRegistration},
	}); err != nil {
		t.Fatal(err)
	}
	// Registration published before term dates.
	if err := svc.replaceImportantDates(ctx, "202720", []importantDate{
		{date: date(2027, 2, 23), description: "Phase I", category: categoryRegistration, opensRegistration: true},
	}); err != nil {
		t.Fatal(err)
	}

	schedule, err := LoadSchedule(ctx, queries)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		code                     string
		registration, start, end string
	}{
		{"202640", "2026-05-05", "2026-09-23", "2026-12-12"}, // registration estimated from the real start
		{"202710", "2026-11-03", "2027-01-05", "2027-03-20"},
		{"202720", "2027-02-23", "2027-04-01", "2027-06-13"}, // term dates estimated
	}
	for _, tt := range tests {
		d, ok := schedule.Dates(tt.code)
		if !ok {
			t.Fatalf("Dates(%s) not ok", tt.code)
		}
		got := [3]string{d.RegistrationOpens.Format(dateLayout), d.Start.Format(dateLayout), d.End.Format(dateLayout)}
		want := [3]string{tt.registration, tt.start, tt.end}
		if got != want {
			t.Errorf("Dates(%s) = %v, want %v", tt.code, got, want)
		}
		if d.Start.Location() != Pacific || d.Start.Hour() != 0 {
			t.Errorf("Dates(%s).Start = %v, want midnight Pacific", tt.code, d.Start)
		}
	}
}

// registrar serves the fixtures the way registrar.wwu.edu does. Spring 2027's
// important dates fail, so one source erroring must not block the rest.
func registrar(t *testing.T) *httptest.Server {
	t.Helper()
	pages := map[string]string{
		termDatesPath:      "term_dates.html",
		finalsPath:         "finals.html",
		importantDatesPath: "important_dates_select.html",
	}
	ajax := map[string]string{"292": "important_dates_202640.json", "293": "important_dates_202710.json"}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, ok := pages[r.URL.Path]
		if r.URL.Path == drupalAjaxPath {
			name, ok = ajax[r.URL.Query().Get("field_quarter_target_id")]
		}
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		body, err := os.ReadFile("testdata/" + name)
		if err != nil {
			t.Errorf("read fixture: %v", err)
		}
		w.Write(body)
	}))
}

func TestScrapeAll(t *testing.T) {
	ctx := context.Background()
	srv := registrar(t)
	defer srv.Close()

	// Summer 2027 is on the registrar site but not yet in Banner.
	svc, queries := newTestService(t, "202630", "202640", "202710", "202720")
	svc.client = newClient(srv.URL)

	err := svc.ScrapeAll(ctx)
	if err == nil || !strings.Contains(err.Error(), "202720") || strings.Contains(err.Error(), "202730") {
		t.Errorf("ScrapeAll error = %v, want only Spring 2027's important dates failure", err)
	}

	schedule, err := LoadSchedule(ctx, queries)
	if err != nil {
		t.Fatal(err)
	}
	if len(schedule.scraped) != 4 {
		t.Errorf("scraped dates for %d terms, want the 4 known", len(schedule.scraped))
	}
	winter, _ := schedule.Dates("202710")
	if got := winter.RegistrationOpens.Format(dateLayout); got != "2026-11-03" {
		t.Errorf("Winter 2027 registration opens %s, want 2026-11-03", got)
	}

	for code, want := range map[string]int{"202640": 20, "202710": 20, "202720": 20, "202630": 0} {
		rows, err := queries.GetFinalMappingsByTerm(ctx, code)
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != want {
			t.Errorf("term %s: %d finals rows, want %d", code, len(rows), want)
		}
	}
	for code, wantAny := range map[string]bool{"202640": true, "202710": true, "202720": false} {
		rows, err := queries.GetImportantDatesByTerm(ctx, code)
		if err != nil {
			t.Fatal(err)
		}
		if (len(rows) > 0) != wantAny {
			t.Errorf("term %s: %d important dates, want any = %v", code, len(rows), wantAny)
		}
	}
}
