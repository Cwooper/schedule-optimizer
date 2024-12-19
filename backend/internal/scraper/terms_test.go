// term_test.go

package scraper

import (
	"testing"
)

func TestParseTermCode(t *testing.T) {
	tests := []struct {
		name         string
		termCode     string
		wantAcadYear string
		wantQuarter  int
		wantError    bool
	}{
		{
			name:         "Summer Quarter 24",
			termCode:     "202430",
			wantAcadYear: "2324",
			wantQuarter:  30,
			wantError:    false,
		},
		{
			name:         "Fall Quarter 24",
			termCode:     "202440",
			wantAcadYear: "2425",
			wantQuarter:  40,
			wantError:    false,
		},
		{
			name:         "Winter Quarter 25",
			termCode:     "202510",
			wantAcadYear: "2425",
			wantQuarter:  10,
			wantError:    false,
		},
		{
			name:         "Spring Quarter 25",
			termCode:     "202520",
			wantAcadYear: "2425",
			wantQuarter:  20,
			wantError:    false,
		},
		{
			name:         "Invalid Quarter 25",
			termCode:     "202450",
			wantAcadYear: "",
			wantQuarter:  0,
			wantError:    true,
		},
		{
			name:         "Invalid Length",
			termCode:     "2024",
			wantAcadYear: "",
			wantQuarter:  0,
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTermCode(tt.termCode)
			if (err != nil) != tt.wantError {
				t.Errorf("ParseTermCode() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				if got.AcadYear != tt.wantAcadYear {
					t.Errorf("ParseTermCode(%v) acadYear = %v, want %v", tt.termCode, got.AcadYear, tt.wantAcadYear)
				}
				if got.Quarter != tt.wantQuarter {
					t.Errorf("ParseTermCode(%v) quarter = %v, want %v", tt.termCode, got.Quarter, tt.wantQuarter)
				}
			}
		})
	}
}
