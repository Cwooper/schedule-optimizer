package calendar

import (
	"strings"
	"testing"
	"time"
)

func TestParseTermSelect(t *testing.T) {
	options, err := parseTermSelect(readFixture(t, "important_dates_202640.html"))
	if err != nil {
		t.Fatalf("parseTermSelect: %v", err)
	}
	want := []termOption{
		{nodeID: "292", termCode: "202640"},
		{nodeID: "293", termCode: "202710"},
		{nodeID: "294", termCode: "202720"},
		{nodeID: "298", termCode: "202730"},
	}
	if len(options) != len(want) {
		t.Fatalf("got %d options, want %d: %+v", len(options), len(want), options)
	}
	for i := range want {
		if options[i] != want[i] {
			t.Errorf("option %d = %+v, want %+v", i, options[i], want[i])
		}
	}
}

func TestParseImportantDates(t *testing.T) {
	type row struct {
		date, end   string // end "" for a single date
		category    category
		opensReg    bool
		description string // prefix
	}
	tests := []struct {
		name    string
		fixture string
		code    string
		want    []row
	}{
		{
			name:    "winter rolls autumn dates back a year",
			fixture: "important_dates_202710.html",
			code:    "202710",
			want: []row{
				{"2026-11-03", "2026-11-16", categoryRegistration, true, "Phase I - registration"},
				{"2026-11-17", "2027-01-04", categoryRegistration, false, "Phase II"},
				{"2026-10-02", "", categoryDeadline, false, "Last day to apply for a bachelor's degree (before"},
				{"2027-01-05", "", categoryAcademic, false, "First Day of Classes"},
				{"2027-05-16", "", categoryDeadline, false, "Last day to make a reservation for June"},
				{"2027-03-20", "2027-03-29", categoryHoliday, false, "Spring Break"},
			},
		},
		{
			name:    "fall range runs into the next year",
			fixture: "important_dates_202640.html",
			code:    "202640",
			want: []row{
				{"2026-05-12", "2026-07-13", categoryRegistration, true, "Phase I - registration"},
				{"2026-12-12", "2027-01-04", categoryHoliday, false, "Winter Break"},
				{"2026-12-07", "2026-12-11", categoryAcademic, false, "Final Exams"},
				{"2026-07-21", "2026-07-27", categoryRegistration, false, "Advising & Registration - Transfer"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dates, err := parseImportantDates(readFixture(t, tt.fixture), tt.code)
			if err != nil {
				t.Fatalf("parseImportantDates: %v", err)
			}
			opening := 0
			for _, d := range dates {
				if d.opensRegistration {
					opening++
				}
			}
			if opening != 1 {
				t.Errorf("%d rows open registration, want 1", opening)
			}
			for _, w := range tt.want {
				d, ok := findDate(dates, w.description)
				if !ok {
					t.Errorf("no row starting %q", w.description)
					continue
				}
				gotEnd := ""
				if !d.end.IsZero() {
					gotEnd = d.end.Format(dateLayout)
				}
				if got := d.date.Format(dateLayout); got != w.date || gotEnd != w.end {
					t.Errorf("%q: dates %s..%s, want %s..%s", w.description, got, gotEnd, w.date, w.end)
				}
				if d.category != w.category || d.opensRegistration != w.opensReg {
					t.Errorf("%q: category %s opens %v, want %s opens %v",
						w.description, d.category, d.opensRegistration, w.category, w.opensReg)
				}
			}
		})
	}
}

func findDate(dates []importantDate, prefix string) (importantDate, bool) {
	for _, d := range dates {
		if strings.HasPrefix(d.description, prefix) {
			return d, true
		}
	}
	return importantDate{}, false
}

// importantPage is the important dates page filtered to Winter 2027, with
// one table of the given caption and rows (date cell, description).
func importantPage(caption string, rows ...[2]string) []byte {
	var b strings.Builder
	b.WriteString(`<select name="field_quarter_target_id"><option value="292">Fall 2026</option>` +
		`<option value="293" selected="selected">Winter 2027</option></select>` +
		`<table><caption>` + caption + `</caption><tbody>`)
	for _, r := range rows {
		b.WriteString(`<tr><td>` + r[0] + `</td><td>` + r[1] + `</td></tr>`)
	}
	b.WriteString(`</tbody></table>`)
	return []byte(b.String())
}

func TestParseImportantDatesFlagsEarliestPhaseOne(t *testing.T) {
	tests := []struct {
		name string
		rows [][2]string
		want string // description of the flagged row, "" for none
	}{
		{"no phase I", [][2]string{
			{"Tue, Nov 17", "Phase II - registration"},
			{"Tue, Jan 5", "Phase III - open registration"},
		}, ""},
		{"earlier phase I listed second", [][2]string{
			{"Tue, Nov 17", "Phase I - registration for new students"},
			{"Tue, Nov 3", "Phase I - registration for continuing students"},
			{"Tue, Nov 10", "Phase I - registration for transfers"},
		}, "Phase I - registration for continuing students"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dates, err := parseImportantDates(importantPage("Registration", tt.rows...), "202710")
			if err != nil {
				t.Fatalf("parseImportantDates: %v", err)
			}
			var flagged []string
			for _, d := range dates {
				if d.opensRegistration {
					flagged = append(flagged, d.description)
				}
			}
			if strings.Join(flagged, "|") != tt.want {
				t.Errorf("flagged %q, want %q", flagged, tt.want)
			}
		})
	}
}

func TestParseImportantDatesRejectsPartialTerm(t *testing.T) {
	phaseOne := [2]string{"Tue, Nov 3 \u2013 Mon, Nov 16", "Phase I - registration"}
	tests := []struct {
		name    string
		page    []byte
		code    string
		wantErr bool
	}{
		{"en-dashed range parses", importantPage("Registration", phaseOne), "202710", false},
		{"blank row ignored", importantPage("Registration", phaseOne, [2]string{" ", ""}), "202710", false},
		{"unreadable date", importantPage("Registration", phaseOne, [2]string{"Nov 17 to Jan 4", "Phase II"}), "202710", true},
		{"date without description", importantPage("Registration", phaseOne, [2]string{"Tue, Nov 17", ""}), "202710", true},
		{"page shows another term", importantPage("Registration", phaseOne), "202640", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dates, err := parseImportantDates(tt.page, tt.code)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, want error %v", err, tt.wantErr)
			}
			if err == nil && dates[0].end.Format(dateLayout) != "2026-11-16" {
				t.Errorf("phase I ends %s, want 2026-11-16", dates[0].end.Format(dateLayout))
			}
		})
	}
}

func TestParseMonthDay(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		anchor time.Time
		want   string
	}{
		{"same year", "Tue, Mar 31", date(2026, time.April, 1), "2026-03-31"},
		{"winter registration in previous autumn", "Tue, Nov 3", date(2027, time.January, 5), "2026-11-03"},
		{"winter deadline the next spring", "Sun, May 16", date(2027, time.January, 5), "2027-05-16"},
		{"spring deadline in previous december", "Fri, Dec 4", date(2027, time.April, 1), "2026-12-04"},
		{"fall break ending in january", "Mon, Jan 4", date(2026, time.September, 24), "2027-01-04"},
		{"fall registration in may", "Tue, May 12", date(2026, time.September, 24), "2026-05-12"},
		{"no weekday", "Nov 3", date(2027, time.January, 5), "2026-11-03"},
		{"leap day picks the leap year", "Thu, Feb 29", date(2028, time.April, 1), "2028-02-29"},
		{"not a date", "TBA", date(2027, time.January, 5), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseMonthDay(tt.text, tt.anchor)
			if tt.want == "" {
				if ok {
					t.Errorf("parseMonthDay(%q) = %s, want no date", tt.text, got.Format(dateLayout))
				}
				return
			}
			if !ok || got.Format(dateLayout) != tt.want {
				t.Errorf("parseMonthDay(%q) = %s (ok %v), want %s", tt.text, got.Format(dateLayout), ok, tt.want)
			}
		})
	}
}

func TestCategorizeCaption(t *testing.T) {
	tests := []struct {
		caption string
		want    category
	}{
		{"Start of Quarter", categoryAcademic},
		{"End of Quarter", categoryAcademic},
		{"Registration", categoryRegistration},
		{"Holidays", categoryHoliday},
		{"Last Day Deadlines", categoryDeadline},
		{"Six-Week Course Deadlines", categoryDeadline},
	}
	for _, tt := range tests {
		if got := categorizeCaption(tt.caption); got != tt.want {
			t.Errorf("categorizeCaption(%q) = %s, want %s", tt.caption, got, tt.want)
		}
	}
}
