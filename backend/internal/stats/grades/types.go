package grades

// GradeAggregate holds pre-computed grade statistics at a specific aggregation level.
type GradeAggregate struct {
	Level        string // "course_professor", "course", "professor", "subject"
	Subject      string
	CourseNumber string
	Instructor   string

	Sections int
	Students int

	CntA  int
	CntAM int
	CntBP int
	CntB  int
	CntBM int
	CntCP int
	CntC  int
	CntCM int
	CntDP int
	CntD  int
	CntDM int
	CntF  int
	CntW  int
	CntP  int
	CntNP int
	CntS  int
	CntU  int

	GPA      float64
	PassRate *float64 // nil if no S/U/P/NP data
}

// computeGPA calculates weighted GPA from letter grade counts.
// Returns 0 if no letter grades exist.
func computeGPA(a, am, bp, b, bm, cp, c, cm, dp, d, dm, f int) float64 {
	total := a + am + bp + b + bm + cp + c + cm + dp + d + dm + f
	if total == 0 {
		return 0
	}
	weighted := float64(a)*4.0 + float64(am)*3.7 +
		float64(bp)*3.3 + float64(b)*3.0 + float64(bm)*2.7 +
		float64(cp)*2.3 + float64(c)*2.0 + float64(cm)*1.7 +
		float64(dp)*1.3 + float64(d)*1.0 + float64(dm)*0.7 +
		float64(f)*0.0
	return weighted / float64(total)
}

// computePassRate calculates (S+P) / (S+U+P+NP).
// Returns nil if no S/U/P/NP data exists.
func computePassRate(s, u, p, np int) *float64 {
	total := s + u + p + np
	if total == 0 {
		return nil
	}
	rate := float64(s+p) / float64(total)
	return &rate
}
