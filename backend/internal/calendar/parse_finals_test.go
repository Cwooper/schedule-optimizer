package calendar

import (
	"strings"
	"testing"
)

func parseFinalsFixture(t *testing.T) []finalMapping {
	t.Helper()
	mappings, failed, err := parseFinals(readFixture(t, "finals.html"))
	if err != nil || len(failed) > 0 {
		t.Fatalf("parseFinals: %v, failed terms %v", err, failed)
	}
	return mappings
}

func TestParseFinals(t *testing.T) {
	mappings := parseFinalsFixture(t)

	perTerm := make(map[string]int)
	for _, m := range mappings {
		perTerm[m.termCode]++
	}
	// 10 start-time rows x 2 day-pattern columns per term.
	for _, code := range []string{"202640", "202710", "202720"} {
		if perTerm[code] != 20 {
			t.Errorf("term %s: %d mappings, want 20", code, perTerm[code])
		}
	}

	tests := []struct {
		name       string
		code       string
		rangeStart string
		hasTuTh    bool
		want       finalMapping
	}{
		{
			name: "morning class, afternoon exam", code: "202640", rangeStart: "0800", hasTuTh: true,
			want: finalMapping{rangeEnd: "0829", examStart: "1300", examEnd: "1500"},
		},
		{
			name: "exam spanning noon under one PM marker", code: "202640", rangeStart: "0830", hasTuTh: false,
			want: finalMapping{rangeEnd: "0929", examStart: "1030", examEnd: "1230"},
		},
		{
			name: "class range spanning noon", code: "202640", rangeStart: "1130", hasTuTh: true,
			want: finalMapping{rangeEnd: "1229", examStart: "0800", examEnd: "1000"},
		},
		{
			name: "evening exam", code: "202640", rangeStart: "1630", hasTuTh: false,
			want: finalMapping{rangeEnd: "1759", examStart: "1800", examEnd: "2000"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, m := range mappings {
				if m.termCode != tt.code || m.rangeStart != tt.rangeStart || m.hasTuTh != tt.hasTuTh {
					continue
				}
				if m.rangeEnd != tt.want.rangeEnd || m.examStart != tt.want.examStart || m.examEnd != tt.want.examEnd {
					t.Errorf("got range end %s, exam %s-%s; want %s, %s-%s",
						m.rangeEnd, m.examStart, m.examEnd, tt.want.rangeEnd, tt.want.examStart, tt.want.examEnd)
				}
				return
			}
			t.Fatal("mapping not found")
		})
	}
}

func TestParseFinalsExamDates(t *testing.T) {
	mappings := parseFinalsFixture(t)
	// Fall 2026 finals week is Mon Dec 7 - Fri Dec 11.
	want := map[string]string{
		"0800": "2026-12-07", // Monday
		"0930": "2026-12-08", // Tuesday
		"1030": "2026-12-09", // Wednesday
		"1330": "2026-12-10", // Thursday
		"1130": "2026-12-11", // Friday
	}
	for _, m := range mappings {
		if m.termCode != "202640" || !m.hasTuTh {
			continue
		}
		if w, ok := want[m.rangeStart]; ok {
			if got := m.examDate.Format(dateLayout); got != w {
				t.Errorf("class at %s: exam date %s, want %s", m.rangeStart, got, w)
			}
			delete(want, m.rangeStart)
		}
	}
	for start := range want {
		t.Errorf("no T/Th mapping for class at %s", start)
	}
}

// finalsItem is one switcher item with a two-column table: a T/Th column,
// then a neither column.
func finalsItem(label, heading string, rows ...[3]string) string {
	var b strings.Builder
	b.WriteString(`<wwu-switcher-item label="` + label + `"><h2>` + heading + `</h2><table><thead><tr>` +
		`<th>Class start time</th><th>Tuesday or Thursday</th><th>neither Tuesday nor Thursday</th></tr></thead><tbody>`)
	for _, r := range rows {
		b.WriteString(`<tr><th>` + r[0] + `</th><td>` + r[1] + `</td><td>` + r[2] + `</td></tr>`)
	}
	b.WriteString(`</tbody></table></wwu-switcher-item>`)
	return b.String()
}

func TestParseFinalsFailsOnlyTheMalformedTerm(t *testing.T) {
	good := [3]string{"Between 8:00 and 8:29 AM", "Monday 1:00 - 3:00 PM", "Monday 3:30 - 5:30 PM"}
	html := finalsItem("Fall 2026", "December 7\u201311", good,
		// En dashes, as the registrar already uses in the week heading.
		[3]string{"Between 8:30 and 9:29 AM", "Monday 8:00 \u2013 10:00 AM", "Monday 10:30\u201312:30 PM"}) +
		finalsItem("Winter 2027", "March 15-19", good,
			[3]string{"Between 8:30 and 9:29 AM", "Thursday 8:00 - 10:00 AM", "TBA"}) +
		finalsItem("Spring 2027", "Schedule coming soon") +
		finalsItem("Summer 2027", "June 28-July 2", good,
			[3]string{"Between 8:30 and 9:29 AM", "Thursday 8:00 - 10:00 AM"})

	mappings, failed, err := parseFinals([]byte(html))
	if err != nil {
		t.Fatalf("parseFinals: %v", err)
	}
	perTerm := make(map[string]int)
	for _, m := range mappings {
		perTerm[m.termCode]++
	}
	if len(perTerm) != 1 || perTerm["202640"] != 4 {
		t.Errorf("mappings per term = %v, want only Fall 2026's 4", perTerm)
	}
	var failedCodes []string
	for _, f := range failed {
		failedCodes = append(failedCodes, f.code)
	}
	if strings.Join(failedCodes, ",") != "202710,202720,202730" {
		t.Errorf("failed terms = %v, want 202710 (bad cell), 202720 (no table), 202730 (missing cell)", failedCodes)
	}
}

func TestTimeRange(t *testing.T) {
	tests := []struct {
		start, end, marker string
		wantStart, wantEnd string
	}{
		{"8:00", "10:00", "AM", "0800", "1000"},
		{"1:00", "3:00", "PM", "1300", "1500"},
		{"10:30", "12:30", "PM", "1030", "1230"},
		{"11:30", "12:29", "PM", "1130", "1229"},
		{"12:30", "1:29", "PM", "1230", "1329"},
		{"6:00", "8:00", "pm", "1800", "2000"},
	}
	for _, tt := range tests {
		t.Run(tt.start+"-"+tt.end+tt.marker, func(t *testing.T) {
			s, e := timeRange(tt.start, tt.end, tt.marker)
			if s != tt.wantStart || e != tt.wantEnd {
				t.Errorf("timeRange = %s-%s, want %s-%s", s, e, tt.wantStart, tt.wantEnd)
			}
		})
	}
}
