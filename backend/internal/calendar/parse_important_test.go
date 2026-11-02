package calendar

import (
	"strings"
	"testing"
	"time"
)

func TestParseTermSelect(t *testing.T) {
	options, err := parseTermSelect(readFixture(t, "important_dates_select.html"))
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
			fixture: "important_dates_202710.json",
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
			fixture: "important_dates_202640.json",
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

func TestParseImportantDatesWithoutPhaseOne(t *testing.T) {
	body := []byte(`[{"command":"insert","method":"replaceWith","data":` +
		`"<table><caption>Registration</caption><tbody>` +
		`<tr><td>Tue, Nov 17</td><td>Phase II - registration</td></tr>` +
		`<tr><td>Tue, Jan 5</td><td>Phase III - open registration</td></tr>` +
		`</tbody></table>"}]`)
	dates, err := parseImportantDates(body, "202710")
	if err != nil {
		t.Fatalf("parseImportantDates: %v", err)
	}
	for _, d := range dates {
		if d.opensRegistration {
			t.Errorf("%q flagged as opening registration", d.description)
		}
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
