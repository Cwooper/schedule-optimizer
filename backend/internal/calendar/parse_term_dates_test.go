package calendar

import (
	"os"
	"strings"
	"testing"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return b
}

func TestParseTermDates(t *testing.T) {
	dates, failed, err := parseTermDates(readFixture(t, "term_dates.html"))
	if err != nil || len(failed) > 0 {
		t.Fatalf("parseTermDates: %v, failed %v", err, failed)
	}

	// Fall 2025 through Summer 2033.
	if len(dates) != 32 {
		t.Errorf("got %d terms, want 32", len(dates))
	}

	got := make(map[string]termDates, len(dates))
	for _, d := range dates {
		if _, dup := got[d.termCode]; dup {
			t.Errorf("term %s parsed twice", d.termCode)
		}
		got[d.termCode] = d
	}

	tests := []struct {
		name       string
		code       string
		start, end string
	}{
		{"four-digit year", "202540", "2025-09-24", "2025-12-12"},
		{"two-digit start year", "202610", "2026-01-06", "2026-03-20"},
		{"winter after fall in academic year", "202710", "2027-01-05", "2027-03-19"},
		{"nbsp between label and date", "202920", "2029-04-03", "2029-06-15"},
		{"no space before dash", "203020", "2030-04-02", "2030-06-14"},
		{"last listed", "203330", "2033-06-21", "2033-08-19"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, ok := got[tt.code]
			if !ok {
				t.Fatalf("missing term %s", tt.code)
			}
			if s := d.start.Format(dateLayout); s != tt.start {
				t.Errorf("start = %s, want %s", s, tt.start)
			}
			if e := d.end.Format(dateLayout); e != tt.end {
				t.Errorf("end = %s, want %s", e, tt.end)
			}
		})
	}
}

func TestParseTermDatesSkipsImplausibleSpans(t *testing.T) {
	html := `<div class="field--name-body">
		<p>Fall: 9/23/2026 - 12/11/2026</p>
		<p>Winter: 1/5/2027 - 3/19/2026</p>
		<p>Spring: 3/30/2026 - 6/11/2027</p>
		<p>Summer: 6/22/2027 ` + "\u2013" + ` 8/20/2027</p>
	</div>`
	dates, failed, err := parseTermDates([]byte(html))
	if err != nil {
		t.Fatal(err)
	}
	var kept, skipped []string
	for _, d := range dates {
		kept = append(kept, d.termCode)
	}
	for _, f := range failed {
		skipped = append(skipped, f.code)
	}
	if strings.Join(kept, ",") != "202640,202730" {
		t.Errorf("kept %v, want Fall 2026 and the en-dashed Summer 2027", kept)
	}
	// Spring's typo'd start year files it under 2026.
	if strings.Join(skipped, ",") != "202710,202620" {
		t.Errorf("skipped %v, want reversed Winter 2027 and year-long Spring", skipped)
	}
}

func TestParseTermDatesRejectsPageWithoutDates(t *testing.T) {
	_, _, err := parseTermDates([]byte(`<div class="field--name-body"><p>Coming soon</p></div>`))
	if err == nil {
		t.Error("want error for a page with no term dates")
	}
}
