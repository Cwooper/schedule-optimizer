package calendar

import (
	"os"
	"testing"
)

func TestParseDrupalTermSelect(t *testing.T) {
	html, err := os.ReadFile("testdata/important_dates_select.html")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	options, err := parseDrupalTermSelect(html)
	if err != nil {
		t.Fatalf("parseDrupalTermSelect: %v", err)
	}

	if len(options) == 0 {
		t.Fatal("expected at least one term option")
	}

	// Verify known entries
	found := make(map[string]DrupalTermOption)
	for _, o := range options {
		found[o.TermCode] = o
	}

	if opt, ok := found["202610"]; !ok {
		t.Error("missing Winter 2026")
	} else if opt.NodeID != "274" {
		t.Errorf("Winter 2026 node ID = %s, want 274", opt.NodeID)
	}

	if opt, ok := found["202620"]; !ok {
		t.Error("missing Spring 2026")
	} else if opt.NodeID != "275" {
		t.Errorf("Spring 2026 node ID = %s, want 275", opt.NodeID)
	}

	t.Logf("Found %d term options", len(options))
	for _, o := range options {
		t.Logf("  %s -> %s (%s)", o.NodeID, o.Description, o.TermCode)
	}
}

func TestParseImportantDatesAjax(t *testing.T) {
	jsonData, err := os.ReadFile("testdata/important_dates_ajax.json")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	holidays, dates, err := parseImportantDatesAjax(jsonData, "202620")
	if err != nil {
		t.Fatalf("parseImportantDatesAjax: %v", err)
	}

	t.Logf("Parsed %d holidays, %d important dates", len(holidays), len(dates))

	// Spring 2026 should have at least Memorial Day and Juneteenth
	if len(holidays) < 1 {
		t.Error("expected at least 1 holiday")
	}

	for _, h := range holidays {
		t.Logf("  Holiday: %s - %s", h.Date.Format("2006-01-02"), h.Description)
	}

	// Should have registration phases, deadlines, etc.
	if len(dates) < 5 {
		t.Errorf("expected at least 5 important dates, got %d", len(dates))
	}

	// Check categories are being assigned
	categories := make(map[string]int)
	for _, d := range dates {
		categories[d.Category]++
		t.Logf("  [%s] %s - %s", d.Category, d.Date.Format("2006-01-02"), d.Description)
	}

	if categories["registration"] == 0 {
		t.Error("expected at least one registration date")
	}
	if categories["academic"] == 0 {
		t.Error("expected at least one academic date")
	}
}

func TestCategorizeCaption(t *testing.T) {
	tests := []struct {
		caption string
		want    string
	}{
		{"Start of Quarter", "academic"},
		{"Registration", "registration"},
		{"Holidays", "holiday"},
		{"Finals Week", "academic"},
		{"Tuition and Fees", "deadline"},
		{"End of Quarter", "academic"},
		{"Commencement", "academic"},
		{"Break", "academic"},
	}

	for _, tt := range tests {
		t.Run(tt.caption, func(t *testing.T) {
			got := categorizeCaption(tt.caption)
			if got != tt.want {
				t.Errorf("categorizeCaption(%q) = %q, want %q", tt.caption, got, tt.want)
			}
		})
	}
}
