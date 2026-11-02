package termcode

import (
	"slices"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		code        string
		wantYear    int
		wantQuarter int
		wantErr     bool
	}{
		{"202510", 2025, Winter, false},
		{"202520", 2025, Spring, false},
		{"202530", 2025, Summer, false},
		{"202540", 2025, Fall, false},
		{"202550", 0, 0, true},
		{"20251", 0, 0, true},
		{"2025100", 0, 0, true},
		{"abcd10", 0, 0, true},
		{"2025ab", 0, 0, true},
		{"", 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			year, quarter, err := Parse(tt.code)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse(%q) error = %v, wantErr %v", tt.code, err, tt.wantErr)
			}
			if year != tt.wantYear || quarter != tt.wantQuarter {
				t.Errorf("Parse(%q) = (%d, %d), want (%d, %d)", tt.code, year, quarter, tt.wantYear, tt.wantQuarter)
			}
		})
	}
}

func TestFromDescription(t *testing.T) {
	tests := []struct {
		desc string
		want string
	}{
		{"Winter 2026", "202610"},
		{"spring 2026", "202620"},
		{"Summer 2026", "202630"},
		{"Fall 2026", "202640"},
		{"  Fall   2026 ", "202640"},
		{"Autumn 2026", ""},
		{"Winter", ""},
		{"Winter 2026 (View Only)", ""},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if got := FromDescription(tt.desc); got != tt.want {
				t.Errorf("FromDescription(%q) = %q, want %q", tt.desc, got, tt.want)
			}
		})
	}
}

func TestOrdinal(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"202510", "202520", 1},
		{"202540", "202610", 1},
		{"202510", "202610", 4},
		{"202630", "202510", -6},
	}
	for _, tt := range tests {
		a, errA := Ordinal(tt.a)
		b, errB := Ordinal(tt.b)
		if errA != nil || errB != nil {
			t.Fatalf("Ordinal(%q, %q) errors: %v, %v", tt.a, tt.b, errA, errB)
		}
		if got := b - a; got != tt.want {
			t.Errorf("Ordinal(%q) - Ordinal(%q) = %d, want %d", tt.b, tt.a, got, tt.want)
		}
	}
}

func TestCurrent(t *testing.T) {
	tests := []struct {
		month time.Month
		want  string
	}{
		{time.January, "202510"},
		{time.March, "202510"},
		{time.April, "202520"},
		{time.May, "202520"},
		{time.June, "202530"},
		{time.August, "202530"},
		{time.September, "202540"},
		{time.December, "202540"},
	}
	for _, tt := range tests {
		t.Run(tt.month.String(), func(t *testing.T) {
			now := time.Date(2025, tt.month, 15, 0, 0, 0, 0, time.UTC)
			if got := Current(now); got != tt.want {
				t.Errorf("Current(%v) = %q, want %q", now, got, tt.want)
			}
		})
	}
}

func TestAcademicYear(t *testing.T) {
	got := AcademicYear(2025)
	want := []string{"202440", "202510", "202520", "202530"}
	if !slices.Equal(got, want) {
		t.Errorf("AcademicYear(2025) = %v, want %v", got, want)
	}
}
