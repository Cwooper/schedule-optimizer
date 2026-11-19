package grades

import "testing"

func TestCountsRates(t *testing.T) {
	tests := []struct {
		name                string
		c                   Counts
		gpa, pass, withdrew *float64
		passFail            *float64
		reportable          bool
	}{
		{
			name: "letters",
			// (5*4 + 5*3 + 5*2 + 5*0) / 20 = 2.25; pass is A..C- = 15 of 20
			c:   Counts{GradeA: 5, GradeB: 5, GradeC: 5, GradeF: 5},
			gpa: new(2.25), pass: new(0.75), withdrew: new(0.0), reportable: true,
		},
		{
			name: "D is not a pass",
			c:    Counts{GradeCMinus: 6, GradeDPlus: 4},
			gpa:  new(1.54), pass: new(0.6), withdrew: new(0.0), reportable: true,
		},
		{
			name: "P and NP mixed into letter rows count toward pass, not GPA",
			c:    Counts{GradeA: 8, GradeF: 2, GradeP: 6, GradeNP: 4, GradeW: 5},
			// GPA from the 10 letters only; pass (8 + 6) / 20; W 5 / 25
			gpa: new(3.2), pass: new(0.7), withdrew: new(0.2), passFail: new(0.6), reportable: true,
		},
		{
			name: "S/U only has no GPA, and pass equals the generator's pass rate",
			c:    Counts{GradeS: 18, GradeU: 2, GradeW: 1},
			pass: new(0.9), withdrew: new(1.0 / 21), passFail: new(0.9), reportable: true,
		},
		{
			name:     "all W",
			c:        Counts{GradeW: 12},
			withdrew: new(1.0),
		},
		{
			name: "nine letters is too small, whatever else it has",
			c:    Counts{GradeA: 9, GradeS: 30},
			gpa:  new(4.0), pass: new(1.0), withdrew: new(0.0), passFail: new(1.0),
		},
		{
			name: "ten non-W grades report a scope with no letters",
			c:    Counts{GradeP: 7, GradeNP: 3, GradeW: 4},
			pass: new(0.7), withdrew: new(4.0 / 14), passFail: new(0.7), reportable: true,
		},
		{
			name: "empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkRate(t, "GPA", tt.c.GPA(), tt.gpa)
			checkRate(t, "Pass", tt.c.Pass(), tt.pass)
			checkRate(t, "Withdrew", tt.c.Withdrew(), tt.withdrew)
			checkRate(t, "PassFail", tt.c.PassFail(), tt.passFail)
			if got := tt.c.Reportable(); got != tt.reportable {
				t.Errorf("Reportable() = %v, want %v", got, tt.reportable)
			}
		})
	}
}

//go:fix inline

func checkRate(t *testing.T, name string, got, want *float64) {
	t.Helper()
	switch {
	case got == nil && want == nil:
	case got == nil || want == nil:
		t.Errorf("%s() = %v, want %v", name, deref(got), deref(want))
	case !almostEqual(*got, *want, 1e-9):
		t.Errorf("%s() = %v, want %v", name, *got, *want)
	}
}

func deref(v *float64) any {
	if v == nil {
		return "nil"
	}
	return *v
}
