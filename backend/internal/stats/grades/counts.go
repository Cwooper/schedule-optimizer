package grades

import "schedule-optimizer/internal/store"

// Grade indexes Counts. The order is the grade file's column order, which
// the frontend's distribution charts share.
type Grade int

const (
	GradeA Grade = iota
	GradeAMinus
	GradeBPlus
	GradeB
	GradeBMinus
	GradeCPlus
	GradeC
	GradeCMinus
	GradeDPlus
	GradeD
	GradeDMinus
	GradeF
	GradeW
	GradeP
	GradeNP
	GradeS
	GradeU
	numGrades
)

var gradePoints = [GradeF + 1]float64{4, 3.7, 3.3, 3, 2.7, 2.3, 2, 1.7, 1.3, 1, 0.7, 0}

// MinReportable is the fewest grades a scope needs before a rate is shown:
// letter grades for its GPA, non-W grades for its pass rate. Smaller classes
// swing too far on one student to compare.
const MinReportable = 10

// Counts holds grade counts summed over any set of sections. GPA and rates
// always come from the sums, never from averaging per-section values.
type Counts [numGrades]int

func countsOf(r *store.GradeRow) Counts {
	return Counts{
		int(r.CntA), int(r.CntAm), int(r.CntBp), int(r.CntB), int(r.CntBm),
		int(r.CntCp), int(r.CntC), int(r.CntCm), int(r.CntDp), int(r.CntD),
		int(r.CntDm), int(r.CntF), int(r.CntW), int(r.CntP), int(r.CntNp),
		int(r.CntS), int(r.CntU),
	}
}

func (c *Counts) Add(o Counts) {
	for i, n := range o {
		c[i] += n
	}
}

func (c Counts) sum(from, to Grade) int {
	n := 0
	for _, v := range c[from : to+1] {
		n += v
	}
	return n
}

func (c Counts) letters() int { return c.sum(GradeA, GradeF) }

// graded is every grade but W.
func (c Counts) graded() int { return c.sum(GradeA, GradeU) - c[GradeW] }

// GPA is nil without letter grades: P, S, and W carry no grade points.
func (c Counts) GPA() *float64 {
	n := c.letters()
	if n == 0 {
		return nil
	}
	var pts float64
	for g, p := range gradePoints {
		pts += float64(c[g]) * p
	}
	return ratio(pts, n)
}

// Pass is (A through C-, P, S) over every non-W grade. For an S/U-only
// course it equals PassFail, so the generator and stats views agree.
func (c Counts) Pass() *float64 {
	n := c.graded()
	if n == 0 {
		return nil
	}
	return ratio(float64(c.sum(GradeA, GradeCMinus)+c[GradeP]+c[GradeS]), n)
}

func (c Counts) Withdrew() *float64 {
	n := c.sum(GradeA, GradeU)
	if n == 0 {
		return nil
	}
	return ratio(float64(c[GradeW]), n)
}

// PassFail is the generator's pass rate, (S+P) / (S+U+P+NP), nil without
// pass/fail grades.
func (c Counts) PassFail() *float64 {
	n := c[GradeS] + c[GradeU] + c[GradeP] + c[GradeNP]
	if n == 0 {
		return nil
	}
	return ratio(float64(c[GradeS]+c[GradeP]), n)
}

// ReportsGPA reports whether the scope has enough letter grades to show
// its GPA (see MinReportable).
func (c Counts) ReportsGPA() bool { return c.letters() >= MinReportable }

// ReportsPass reports whether the scope has enough non-W grades to show its
// pass rate (see MinReportable).
func (c Counts) ReportsPass() bool { return c.graded() >= MinReportable }

func ratio(num float64, den int) *float64 { return new(num / float64(den)) }
