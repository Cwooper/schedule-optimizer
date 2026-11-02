package calendar

import "testing"

func TestParseFinals(t *testing.T) {
	mappings, err := parseFinals(readFixture(t, "finals.html"))
	if err != nil {
		t.Fatalf("parseFinals: %v", err)
	}

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
	mappings, err := parseFinals(readFixture(t, "finals.html"))
	if err != nil {
		t.Fatalf("parseFinals: %v", err)
	}
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
