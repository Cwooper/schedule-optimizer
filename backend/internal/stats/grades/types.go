package grades

import "schedule-optimizer/internal/store"

// GradeAggregate holds pre-computed grade statistics at a specific aggregation level.
type GradeAggregate struct {
	Level        string // "course_professor", "course", "professor", "subject"
	Subject      string
	CourseNumber string
	Instructor   string

	Sections int
	Students int
	Counts   Counts

	GPA      float64  // 0 without letter grades
	PassRate *float64 // Counts.PassFail
}

func (a *GradeAggregate) addRow(row *store.GradeRow) {
	a.Sections++
	a.Students += int(row.GradeCount)
	a.Counts.Add(countsOf(row))
}

// finish derives GPA and PassRate from the summed counts.
func (a *GradeAggregate) finish() {
	a.GPA = 0
	if gpa := a.Counts.GPA(); gpa != nil {
		a.GPA = *gpa
	}
	a.PassRate = a.Counts.PassFail()
}
