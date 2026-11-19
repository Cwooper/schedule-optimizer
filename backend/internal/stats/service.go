// Package stats builds the Statistics tab's course and section views from
// the grade index (grades.Index) and Banner's sections, which are read per
// request since scrapes change them.
package stats

import (
	"context"
	"errors"
	"math"
	"time"

	"schedule-optimizer/internal/calendar"
	"schedule-optimizer/internal/stats/grades"
	"schedule-optimizer/internal/store"
)

var (
	ErrNotLoaded = errors.New("grade data is not loaded yet")
	ErrNotFound  = errors.New("no grade data")
)

type Service struct {
	grades  *grades.Service
	queries *store.Queries
	now     func() time.Time
}

func NewService(gradeService *grades.Service, queries *store.Queries) *Service {
	return &Service{grades: gradeService, queries: queries, now: time.Now}
}

func (s *Service) index() (*grades.Index, error) {
	idx := s.grades.Index()
	if idx == nil {
		return nil, ErrNotLoaded
	}
	return idx, nil
}

// Stats is one scope's grade numbers. Grades and GPA are omitted for a scope
// too small to report (see grades.MinReportable).
type Stats struct {
	Students int            `json:"students"`
	Grades   *grades.Counts `json:"grades,omitempty"`
	GPA      *float64       `json:"gpa"`
	Pass     *float64       `json:"pass"`
	Withdrew *float64       `json:"withdrew"`
}

func newStats(c grades.Counts, students int) Stats {
	st := Stats{Students: students, Pass: round(c.Pass()), Withdrew: round(c.Withdrew())}
	if c.Reportable() {
		st.Grades = &c
		st.GPA = round(c.GPA())
	}
	return st
}

// Metric is what a course's sections are compared by: GPA, or the pass
// rate for a course with no letter grades (S/U only).
type Metric string

const (
	MetricGPA  Metric = "gpa"
	MetricPass Metric = "pass"
)

func metricOf(course grades.Counts) Metric {
	if course.GPA() != nil {
		return MetricGPA
	}
	return MetricPass
}

// value is nil when the scope is too small to report.
func (m Metric) value(c grades.Counts) *float64 {
	if !c.Reportable() {
		return nil
	}
	if m == MetricGPA {
		return round(c.GPA())
	}
	return round(c.Pass())
}

// round trims floats to what the views can show, which keeps payloads small.
func round(v *float64) *float64 {
	if v == nil {
		return nil
	}
	return new(math.Round(*v*1e4) / 1e4)
}

// CourseOption is one course in the pickers.
type CourseOption struct {
	Subject string `json:"subject"`
	Number  string `json:"number"`
	Title   string `json:"title"`
}

type SubjectOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CoursesResponse struct {
	Subjects []SubjectOption `json:"subjects"`
	Courses  []CourseOption  `json:"courses"`
}

// Courses lists every course with grade data, for the pickers.
func (s *Service) Courses() (*CoursesResponse, error) {
	idx, err := s.index()
	if err != nil {
		return nil, err
	}
	resp := &CoursesResponse{
		Subjects: make([]SubjectOption, 0, len(idx.Subjects())),
		Courses:  make([]CourseOption, 0, len(idx.Courses())),
	}
	for _, sub := range idx.Subjects() {
		resp.Subjects = append(resp.Subjects, SubjectOption{Code: sub.Code, Name: sub.Name})
	}
	for _, c := range idx.Courses() {
		resp.Courses = append(resp.Courses, CourseOption{Subject: c.Subject, Number: c.Number, Title: c.Title})
	}
	return resp, nil
}

// termStarted reports whether a term's classes have begun, by the registrar's
// dates where published and estimates otherwise.
func (s *Service) termStarted(ctx context.Context) (func(term string) bool, error) {
	schedule, err := calendar.LoadSchedule(ctx, s.queries)
	if err != nil {
		return nil, err
	}
	now := s.now()
	return func(term string) bool {
		d, ok := schedule.Dates(term)
		return ok && !now.Before(d.Start)
	}, nil
}

func displayName(idx *grades.Index, key string) string {
	if key == "" {
		return ""
	}
	return idx.DisplayName(key)
}

func sumStudents(secs []*grades.Section) (grades.Counts, int) {
	var c grades.Counts
	n := 0
	for _, s := range secs {
		c.Add(s.Counts)
		n += s.Students
	}
	return c, n
}
