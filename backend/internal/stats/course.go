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
	Offered []OfferedTerm `json:"offered"`
	// Upcoming is every term not yet started that offers the course, soonest
	// first.
	Upcoming      []UpcomingTerm `json:"upcoming"`
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

type UpcomingTerm struct {
	Term       string   `json:"term"`
	Professors []string `json:"professors"`
}

// Row is one group of sections in a view's table, so every view (a course's
// professors, later a professor's or subject's courses) shares one builder.
type Row struct {
	Name     string `json:"name"` // "" groups sections with no name listed
	Sections int    `json:"sections"`
	Stats
	History []SectionRow `json:"history"` // newest first
}

type ProfessorRow struct {
	Row
	LastTaught string `json:"lastTaught,omitempty"`
	// NextTerm is the soonest upcoming term the professor teaches.
	NextTerm string `json:"nextTerm,omitempty"`
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
	upcoming   []UpcomingTerm
	nextTerm   map[string]string // professor -> soonest upcoming term
}

// readOfferings reads a course's Banner sections, in term order.
func readOfferings(rows []*store.GetCourseOfferingsRow, started func(string) bool) offeringFacts {
	f := offeringFacts{crns: make(map[string]map[string]bool), lastTaught: make(map[string]string), upcoming: []UpcomingTerm{}, nextTerm: make(map[string]string)}
	for _, r := range rows {
		addCRN(f.crns, r.Term, r.Crn)
		key := r.InstructorName.String
		if started(r.Term) {
			if key != "" {
				f.lastTaught[key] = max(f.lastTaught[key], r.Term)
			}
			continue
		}
		if n := len(f.upcoming); n == 0 || f.upcoming[n-1].Term != r.Term {
			f.upcoming = append(f.upcoming, UpcomingTerm{Term: r.Term, Professors: []string{}})
		}
		u := &f.upcoming[len(f.upcoming)-1]
		if key == "" || slices.Contains(u.Professors, key) {
			continue
		}
		u.Professors = append(u.Professors, key)
		if f.nextTerm[key] == "" {
			f.nextTerm[key] = r.Term
		}
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
		Upcoming: facts.upcoming,
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

func professorRows(secs []*grades.Section, facts offeringFacts, metric Metric) []ProfessorRow {
	groups := groupRows(secs, func(s *grades.Section) string { return s.Professor }, metric)
	rows := make([]ProfessorRow, len(groups))
	for i, g := range groups {
		rows[i].Row = g
		if g.Name != "" {
			// Banner runs past the grade file, so it knows later terms.
			rows[i].LastTaught = max(facts.lastTaught[g.Name], g.History[0].Term)
			rows[i].NextTerm = facts.nextTerm[g.Name]
		}
	}
	return rows
}

// groupRows groups newest-first sections by key, most sections first, with
// the unnamed group last.
func groupRows(secs []*grades.Section, key func(*grades.Section) string, metric Metric) []Row {
	groups := make(map[string][]*grades.Section)
	var order []string
	for _, s := range secs {
		k := key(s)
		if groups[k] == nil {
			order = append(order, k)
		}
		groups[k] = append(groups[k], s)
	}
	rows := make([]Row, 0, len(order))
	for _, k := range order {
		g := groups[k]
		counts, students := sumStudents(g)
		row := Row{Name: k, Sections: len(g), Stats: newStats(counts, students, metric), History: make([]SectionRow, 0, len(g))}
		for _, s := range g {
			row.History = append(row.History, SectionRow{Term: s.Term, CRN: s.CRN, Stats: newStats(s.Counts, s.Students, metric)})
		}
		rows = append(rows, row)
	}
	slices.SortStableFunc(rows, func(a, b Row) int {
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
