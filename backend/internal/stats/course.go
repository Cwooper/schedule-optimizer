package stats

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"slices"

	"schedule-optimizer/internal/stats/grades"
	"schedule-optimizer/internal/store"
	"schedule-optimizer/internal/termcode"
)

type CourseView struct {
	Subject    string `json:"subject"`
	Number     string `json:"number"`
	Title      string `json:"title"`
	Metric     Metric `json:"metric"`
	Summary    Stats  `json:"summary"`
	Sections   int    `json:"sections"`
	Professors int    `json:"professors"`
	// Terms holds each graded term, oldest first.
	Terms []TermValue `json:"terms"`
	// Seasons holds the metric per quarter over every year: Fall, Winter,
	// Spring, Summer.
	Seasons []SeasonValue `json:"seasons"`
	// Offered is every term the course ran or is scheduled in, oldest first.
	Offered       []OfferedTerm  `json:"offered"`
	Next          *NextTerm      `json:"next,omitempty"`
	ProfessorRows []ProfessorRow `json:"professorRows"`
}

// TermValue is one term's graded sections; Value is nil when they are too
// small to report.
type TermValue struct {
	Term     string   `json:"term"`
	Value    *float64 `json:"value"`
	Sections int      `json:"sections"`
	Students int      `json:"students"`
}

// OfferedTerm counts a term's sections, Banner's and the grade file's, each
// CRN once: a term can be partly scraped, or older than Banner keeps.
type OfferedTerm struct {
	Term     string `json:"term"`
	Sections int    `json:"sections"`
}

type SeasonValue struct {
	Quarter int      `json:"quarter"`
	Value   *float64 `json:"value"`
}

// NextTerm is the first term not yet started that offers the course.
type NextTerm struct {
	Term       string   `json:"term"`
	Professors []string `json:"professors"`
}

type ProfessorRow struct {
	Name     string `json:"name"` // "" groups sections with no professor listed
	Sections int    `json:"sections"`
	Stats
	LastTaught  string       `json:"lastTaught,omitempty"`
	TeachesNext bool         `json:"teachesNext,omitempty"`
	History     []SectionRow `json:"history"` // newest first
}

type SectionRow struct {
	Term string `json:"term"`
	CRN  string `json:"crn"`
	Stats
}

var seasonOrder = []int{termcode.Fall, termcode.Winter, termcode.Spring, termcode.Summer}

// Course builds a course's view. It returns ErrNotFound for a course with no
// grade data.
func (s *Service) Course(ctx context.Context, subject, number string) (*CourseView, error) {
	idx, err := s.index()
	if err != nil {
		return nil, err
	}
	course := idx.Course(grades.CourseKey{Subject: subject, Number: number})
	if course == nil {
		return nil, ErrNotFound
	}
	offerings, err := s.queries.GetCourseOfferings(ctx, store.GetCourseOfferingsParams{Subject: subject, CourseNumber: number})
	if err != nil {
		return nil, fmt.Errorf("load offerings: %w", err)
	}
	started, err := s.termStarted(ctx)
	if err != nil {
		return nil, err
	}
	return buildCourseView(course, readOfferings(offerings, started)), nil
}

// offeringFacts is what Banner says about a course's terms.
type offeringFacts struct {
	crns       map[string]map[string]bool // term -> CRNs
	lastTaught map[string]string          // professor -> newest started term
	next       *NextTerm
	nextKeys   map[string]bool // professors teaching the next term
}

// readOfferings reads a course's Banner sections, in term order.
func readOfferings(rows []*store.GetCourseOfferingsRow, started func(string) bool) offeringFacts {
	f := offeringFacts{crns: make(map[string]map[string]bool), lastTaught: make(map[string]string), nextKeys: make(map[string]bool)}
	for _, r := range rows {
		addCRN(f.crns, r.Term, r.Crn)
		key := r.InstructorName.String
		if started(r.Term) {
			if key != "" {
				f.lastTaught[key] = max(f.lastTaught[key], r.Term)
			}
			continue
		}
		if f.next == nil {
			f.next = &NextTerm{Term: r.Term, Professors: []string{}}
		}
		if r.Term != f.next.Term || key == "" || f.nextKeys[key] {
			continue
		}
		f.nextKeys[key] = true
		f.next.Professors = append(f.next.Professors, key)
	}
	return f
}

func buildCourseView(course *grades.Course, facts offeringFacts) *CourseView {
	total, students := sumStudents(course.Sections)
	metric := metricOf(total)
	v := &CourseView{
		Subject:  course.Subject,
		Number:   course.Number,
		Title:    course.Title,
		Metric:   metric,
		Summary:  newStats(total, students, metric),
		Sections: len(course.Sections),
		Next:     facts.next,
	}
	v.Terms, v.Seasons = termValues(course.Sections, metric)
	v.Offered = offeredTerms(facts.crns, course.Sections)
	v.ProfessorRows = professorRows(course.Sections, facts, metric)
	for _, p := range v.ProfessorRows {
		if p.Name != "" {
			v.Professors++
		}
	}
	return v
}

func termValues(secs []*grades.Section, metric Metric) ([]TermValue, []SeasonValue) {
	byTerm := make(map[string][]*grades.Section)
	bySeason := make(map[int]grades.Counts)
	for _, s := range secs {
		byTerm[s.Term] = append(byTerm[s.Term], s)
		_, q, err := termcode.Parse(s.Term)
		if err != nil {
			continue
		}
		c := bySeason[q]
		c.Add(s.Counts)
		bySeason[q] = c
	}
	terms := make([]TermValue, 0, len(byTerm))
	for term, g := range byTerm {
		c, students := sumStudents(g)
		terms = append(terms, TermValue{Term: term, Value: metric.value(c), Sections: len(g), Students: students})
	}
	slices.SortFunc(terms, func(a, b TermValue) int { return cmp.Compare(a.Term, b.Term) })
	seasons := make([]SeasonValue, 0, len(seasonOrder))
	for _, q := range seasonOrder {
		seasons = append(seasons, SeasonValue{Quarter: q, Value: metric.value(bySeason[q])})
	}
	return terms, seasons
}

func addCRN(crns map[string]map[string]bool, term, crn string) {
	if crns[term] == nil {
		crns[term] = make(map[string]bool)
	}
	crns[term][crn] = true
}

// offeredTerms counts each term's CRNs from Banner and the grade file,
// oldest first.
func offeredTerms(banner map[string]map[string]bool, graded []*grades.Section) []OfferedTerm {
	crns := make(map[string]map[string]bool, len(banner))
	for term, set := range banner {
		crns[term] = maps.Clone(set)
	}
	for _, s := range graded {
		addCRN(crns, s.Term, s.CRN)
	}
	out := make([]OfferedTerm, 0, len(crns))
	for _, term := range slices.Sorted(maps.Keys(crns)) {
		out = append(out, OfferedTerm{Term: term, Sections: len(crns[term])})
	}
	return out
}

// professorRows groups sections by professor, most sections first, with
// unlisted sections last.
func professorRows(secs []*grades.Section, facts offeringFacts, metric Metric) []ProfessorRow {
	groups := make(map[string][]*grades.Section)
	var order []string
	for _, s := range secs {
		if groups[s.Professor] == nil {
			order = append(order, s.Professor)
		}
		groups[s.Professor] = append(groups[s.Professor], s)
	}
	rows := make([]ProfessorRow, 0, len(order))
	for _, key := range order {
		g := groups[key]
		counts, students := sumStudents(g)
		row := ProfessorRow{
			Name:     key,
			Sections: len(g),
			Stats:    newStats(counts, students, metric),
			History:  make([]SectionRow, 0, len(g)),
		}
		if key != "" {
			// Banner runs past the grade file, so it knows later terms.
			row.LastTaught = max(facts.lastTaught[key], g[0].Term)
			row.TeachesNext = facts.nextKeys[key]
		}
		for _, s := range g {
			row.History = append(row.History, SectionRow{Term: s.Term, CRN: s.CRN, Stats: newStats(s.Counts, s.Students, metric)})
		}
		rows = append(rows, row)
	}
	slices.SortStableFunc(rows, func(a, b ProfessorRow) int {
		if (a.Name == "") != (b.Name == "") {
			if a.Name == "" {
				return 1
			}
			return -1
		}
		return cmp.Or(cmp.Compare(b.Sections, a.Sections), cmp.Compare(a.Name, b.Name))
	})
	return rows
}
