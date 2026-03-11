package calendar

import (
	"os"
	"testing"
)

func TestParseTermDates(t *testing.T) {
	html, err := os.ReadFile("testdata/term_dates.html")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	dates, err := parseTermDates(html)
	if err != nil {
		t.Fatalf("parseTermDates: %v", err)
	}

	if len(dates) == 0 {
		t.Fatal("expected at least one term date entry")
	}

	// Verify known entries from the 2025-2026 academic year
	expected := map[string]struct {
		start string
		end   string
	}{
		"202440": {"2024-09-25", "2024-12-13"},
		"202510": {"2025-01-07", "2025-03-21"},
		"202520": {"2025-04-01", "2025-06-13"},
		"202530": {"2025-06-24", "2025-08-22"},
		"202540": {"2025-09-24", "2025-12-12"},
		"202610": {"2026-01-06", "2026-03-20"},
		"202620": {"2026-03-31", "2026-06-12"},
		"202630": {"2026-06-23", "2026-08-21"},
	}

	found := make(map[string]TermDates)
	for _, d := range dates {
		found[d.TermCode] = d
	}

	for code, want := range expected {
		got, ok := found[code]
		if !ok {
			t.Errorf("missing term %s", code)
			continue
		}
		gotStart := got.StartDate.Format("2006-01-02")
		gotEnd := got.EndDate.Format("2006-01-02")
		if gotStart != want.start {
			t.Errorf("term %s: start = %s, want %s", code, gotStart, want.start)
		}
		if gotEnd != want.end {
			t.Errorf("term %s: end = %s, want %s", code, gotEnd, want.end)
		}
	}

	t.Logf("Parsed %d term date entries", len(dates))
}

func TestBuildTermCode(t *testing.T) {
	tests := []struct {
		quarter string
		month   int
		year    int
		want    string
	}{
		{"Fall", 9, 2024, "202440"},
		{"Winter", 1, 2025, "202510"},
		{"Spring", 4, 2025, "202520"},
		{"Summer", 6, 2025, "202530"},
		{"fall", 9, 2026, "202640"},
		{"invalid", 1, 2025, ""},
	}

	for _, tt := range tests {
		t.Run(tt.quarter, func(t *testing.T) {
			got := buildTermCode(tt.quarter, tt.year)
			if got != tt.want {
				t.Errorf("buildTermCode(%q, %d) = %q, want %q", tt.quarter, tt.year, got, tt.want)
			}
		})
	}
}

func TestDescriptionToTermCode(t *testing.T) {
	tests := []struct {
		desc string
		want string
	}{
		{"Winter 2026", "202610"},
		{"Spring 2026", "202620"},
		{"Summer 2026", "202630"},
		{"Fall 2026", "202640"},
		{"Winter 2027", "202710"},
		{"invalid", ""},
		{"", ""},
		{"Winter", ""},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := descriptionToTermCode(tt.desc)
			if got != tt.want {
				t.Errorf("descriptionToTermCode(%q) = %q, want %q", tt.desc, got, tt.want)
			}
		})
	}
}

func TestParseFlexDate(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"9/25/2024", "2024-09-25"},
		{"1/7/25", "2025-01-07"},
		{"12/13/2024", "2024-12-13"},
		{"6/24/2025", "2025-06-24"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseFlexDate(tt.input)
			if err != nil {
				t.Fatalf("parseFlexDate(%q): %v", tt.input, err)
			}
			if got.Format("2006-01-02") != tt.want {
				t.Errorf("parseFlexDate(%q) = %s, want %s", tt.input, got.Format("2006-01-02"), tt.want)
			}
		})
	}
}
