package calendar

import (
	"os"
	"testing"
)

func TestParseFinals(t *testing.T) {
	html, err := os.ReadFile("testdata/finals.html")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	mappings, err := parseFinals(html)
	if err != nil {
		t.Fatalf("parseFinals: %v", err)
	}

	// 3 terms × 10 time slots × 2 columns = 60 mappings
	if len(mappings) < 30 {
		t.Errorf("expected at least 30 mappings, got %d", len(mappings))
	}

	// Count per term
	termCounts := make(map[string]int)
	for _, m := range mappings {
		termCounts[m.TermCode]++
	}

	for term, count := range termCounts {
		// Each term should have 20 mappings (10 time slots × 2 columns)
		if count != 20 {
			t.Errorf("term %s: got %d mappings, want 20", term, count)
		}
	}

	t.Logf("Parsed %d finals mappings across %d terms", len(mappings), len(termCounts))

	// Verify a specific known mapping: Winter 2026, 8:30-9:29 AM, T/Th → Thursday 8:00-10:00 AM
	found := false
	for _, m := range mappings {
		if m.TermCode == "202610" && m.TimeRangeStart == "0830" && m.HasTuTh {
			found = true
			wantDate := "2026-03-19"
			gotDate := m.ExamDate.Format("2006-01-02")
			if gotDate != wantDate {
				t.Errorf("Winter 2026 8:30 T/Th: exam date = %s, want %s", gotDate, wantDate)
			}
			if m.ExamStartTime != "0800" {
				t.Errorf("Winter 2026 8:30 T/Th: exam start = %s, want 0800", m.ExamStartTime)
			}
			if m.ExamEndTime != "1000" {
				t.Errorf("Winter 2026 8:30 T/Th: exam end = %s, want 1000", m.ExamEndTime)
			}
			break
		}
	}
	if !found {
		t.Error("expected Winter 2026 8:30 T/Th mapping not found")
	}
}

func TestTo24h(t *testing.T) {
	tests := []struct {
		time string
		ampm string
		want string
	}{
		{"8:00", "AM", "0800"},
		{"8:29", "AM", "0829"},
		{"12:30", "PM", "1230"},
		{"1:00", "PM", "1300"},
		{"6:00", "PM", "1800"},
		{"12:00", "AM", "0000"},
		{"12:00", "PM", "1200"},
		{"11:29", "AM", "1129"},
	}

	for _, tt := range tests {
		t.Run(tt.time+"_"+tt.ampm, func(t *testing.T) {
			got := to24h(tt.time, tt.ampm)
			if got != tt.want {
				t.Errorf("to24h(%q, %q) = %q, want %q", tt.time, tt.ampm, got, tt.want)
			}
		})
	}
}

func TestBuildFinalsWeekMap(t *testing.T) {
	// March 16, 2026 is a Monday
	monday := makeDate(2026, 3, 16)
	m := buildFinalsWeekMap(monday)

	if got := m["monday"].Day(); got != 16 {
		t.Errorf("monday = %d, want 16", got)
	}
	if got := m["friday"].Day(); got != 20 {
		t.Errorf("friday = %d, want 20", got)
	}
}
