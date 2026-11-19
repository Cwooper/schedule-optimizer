package stats

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"schedule-optimizer/internal/calendar"
	"schedule-optimizer/internal/meeting"
	"schedule-optimizer/internal/stats/grades"
	"schedule-optimizer/internal/store"
)

type SectionView struct {
	Term      string `json:"term"`
	CRN       string `json:"crn"`
	Subject   string `json:"subject"`
	Number    string `json:"number"`
	Title     string `json:"title"`
	Professor string `json:"professor"`
	Metric    Metric `json:"metric"`
	Section   Stats  `json:"section"`
	Course    Stats  `json:"course"`
	// Rank is 1 for the best metric among Dots, shared on ties; 0 when this
	// section is too small to rank.
	Rank int `json:"rank"`
	// Dots holds every reportable section of the course.
	Dots []Dot `json:"dots"`
	// Siblings are the course's graded sections that term, this one included.
	Siblings []SiblingRow `json:"siblings"`
	Class    *Class       `json:"class,omitempty"`
}

type Dot struct {
	Term  string  `json:"term"`
	CRN   string  `json:"crn"`
	Value float64 `json:"value"`
	Mine  bool    `json:"mine,omitempty"` // taught by this section's professor
}

type SiblingRow struct {
	CRN       string `json:"crn"`
	Professor string `json:"professor"`
	Stats
}

// Class is Banner's listing of the section.
type Class struct {
	Meetings []meeting.Time  `json:"meetings"`
	Credits  int             `json:"credits"`
	Capacity int             `json:"capacity"`
	Final    *calendar.Final `json:"final,omitempty"`
}

// Section builds a section's view. It returns ErrNotFound for a section with
// no grade data.
func (s *Service) Section(ctx context.Context, term, crn string) (*SectionView, error) {
	idx, err := s.index()
	if err != nil {
		return nil, err
	}
	sec := idx.Section(term, crn)
	if sec == nil {
		return nil, ErrNotFound
	}
	v := buildSectionView(idx, idx.Course(sec.Course), sec)
	if v.Class, err = s.class(ctx, term, crn); err != nil {
		return nil, err
	}
	return v, nil
}

func buildSectionView(idx *grades.Index, course *grades.Course, sec *grades.Section) *SectionView {
	total, students := sumStudents(course.Sections)
	metric := metricOf(total)
	v := &SectionView{
		Term:      sec.Term,
		CRN:       sec.CRN,
		Subject:   course.Subject,
		Number:    course.Number,
		Title:     course.Title,
		Professor: sec.Professor,
		Metric:    metric,
		Section:   newStats(sec.Counts, sec.Students),
		Course:    newStats(total, students),
		Dots:      []Dot{},
		Siblings:  []SiblingRow{},
	}
	for _, o := range course.Sections {
		if o.Term == sec.Term {
			v.Siblings = append(v.Siblings, SiblingRow{
				CRN:       o.CRN,
				Professor: o.Professor,
				Stats:     newStats(o.Counts, o.Students),
			})
		}
		if val := metric.value(o.Counts); val != nil {
			v.Dots = append(v.Dots, Dot{
				Term:  o.Term,
				CRN:   o.CRN,
				Value: *val,
				Mine:  sec.Professor != "" && o.Professor == sec.Professor,
			})
		}
	}
	v.Rank = rank(v.Dots, metric.value(sec.Counts))
	return v
}

// rank is 1 + the number of dots strictly above value, so ties share a rank.
func rank(dots []Dot, value *float64) int {
	if value == nil {
		return 0
	}
	r := 1
	for _, d := range dots {
		if d.Value > *value {
			r++
		}
	}
	return r
}

// class returns nil when Banner has no such section.
func (s *Service) class(ctx context.Context, term, crn string) (*Class, error) {
	row, err := s.queries.GetSectionWithInstructorByTermAndCRN(ctx, store.GetSectionWithInstructorByTermAndCRNParams{Term: term, Crn: crn})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load section: %w", err)
	}
	rows, err := s.queries.GetMeetingTimesBySection(ctx, row.ID)
	if err != nil {
		return nil, fmt.Errorf("load meetings: %w", err)
	}
	meetings := meeting.FromStoreAll(rows)
	finals, _ := calendar.SectionFinals(ctx, s.queries, []string{term})
	return &Class{
		Meetings: meetings,
		Credits:  int(row.CreditHoursLow.Int64),
		Capacity: int(row.MaxEnrollment.Int64),
		Final:    finals[term].FinalFor(row.ScheduleType.String, meetings),
	}, nil
}
