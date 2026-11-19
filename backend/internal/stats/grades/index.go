package grades

import (
	"cmp"
	"slices"
	"strings"

	"schedule-optimizer/internal/store"
)

// CourseKey is a course as Banner lists it.
type CourseKey struct {
	Subject string
	Number  string
}

// Section is one graded section.
type Section struct {
	Term      string
	CRN       string
	Course    CourseKey
	Professor string // a professor key (see Index.ProfessorKey); "" when unlisted
	Students  int    // enrolled, from the grade file
	Counts    Counts
}

// Course is every graded section of one course, newest first.
type Course struct {
	CourseKey
	Title    string
	Sections []*Section
}

type Subject struct {
	Code string
	Name string // Banner's description; "" for codes Banner never used
}

// Index is the grade file keyed for the stats views. It is built once per
// load and never changed, so readers share it without locking.
//
// The grade file's subject is a department code (LANG 101 covers every
// first-year language), so each row is keyed on the Banner course it ran as,
// found by its (term, CRN). Cross-listed courses (CSCI 247 and CISS 247) stay
// separate, as Banner lists them.
type Index struct {
	courses  map[CourseKey]*Course
	list     []*Course // by subject, then number
	sections map[[2]string]*Section
	levels   map[levelKey]Counts
	subjects []Subject
	names    map[string]string // professor key -> display name
	keys     map[string]string // Banner name -> professor key
}

type levelKey struct {
	subject string
	digit   byte
}

// indexSource is what buildIndex reads, one query each.
type indexSource struct {
	rows     []*store.GetGradeRowsWithBannerRow
	mappings []*store.InstructorMapping
	titles   []*store.GetLatestCourseTitlesRow
	subjects []*store.GetLatestSubjectDescriptionsRow
}

func buildIndex(src indexSource) *Index {
	idx := &Index{
		courses:  make(map[CourseKey]*Course),
		sections: make(map[[2]string]*Section, len(src.rows)),
		levels:   make(map[levelKey]Counts),
	}
	idx.names, idx.keys = professorNames(src.mappings)

	keys := courseKeys(src.rows)
	gradeTitles := make(map[CourseKey]string)
	newest := make(map[CourseKey]string)
	for i, r := range src.rows {
		g := &r.GradeRow
		sec := &Section{
			Term:      g.Term,
			CRN:       g.Crn,
			Course:    keys[i],
			Professor: strings.TrimSpace(g.Professor),
			Students:  int(g.StudentsEnrolled),
			Counts:    countsOf(g),
		}
		if sec.Professor == "" && r.BannerInstructor.String != "" {
			sec.Professor = idx.ProfessorKey(r.BannerInstructor.String)
		}
		c := idx.courses[sec.Course]
		if c == nil {
			c = &Course{CourseKey: sec.Course}
			idx.courses[sec.Course] = c
		}
		c.Sections = append(c.Sections, sec)
		idx.sections[[2]string{sec.Term, sec.CRN}] = sec
		if sec.Term >= newest[sec.Course] {
			newest[sec.Course], gradeTitles[sec.Course] = sec.Term, g.Title
		}
		if lk, ok := levelOf(sec.Course); ok {
			counts := idx.levels[lk]
			counts.Add(sec.Counts)
			idx.levels[lk] = counts
		}
	}

	// Grade titles are abbreviated ("Experimentl Approaches Resrch") and
	// courses get renamed, so Banner's newest title wins.
	bannerTitles := make(map[CourseKey]string, len(src.titles))
	for _, t := range src.titles {
		bannerTitles[CourseKey{t.Subject, t.CourseNumber}] = t.Title
	}
	descriptions := make(map[string]string, len(src.subjects))
	for _, s := range src.subjects {
		descriptions[s.Subject] = s.SubjectDescription.String
	}

	for k, c := range idx.courses {
		c.Title = cmp.Or(bannerTitles[k], gradeTitles[k])
		slices.SortFunc(c.Sections, func(a, b *Section) int {
			return cmp.Or(cmp.Compare(b.Term, a.Term), cmp.Compare(a.CRN, b.CRN))
		})
		idx.list = append(idx.list, c)
	}
	slices.SortFunc(idx.list, func(a, b *Course) int {
		return cmp.Or(cmp.Compare(a.Subject, b.Subject), cmp.Compare(a.Number, b.Number))
	})
	for i, c := range idx.list {
		if i == 0 || c.Subject != idx.list[i-1].Subject {
			idx.subjects = append(idx.subjects, Subject{Code: c.Subject, Name: descriptions[c.Subject]})
		}
	}
	return idx
}

// courseKeys returns each row's Banner course. A row with no Banner section
// takes the one Banner course its grade code always joined to elsewhere
// (HHD 104 is always PE 104), else keeps its grade code.
func courseKeys(rows []*store.GetGradeRowsWithBannerRow) []CourseKey {
	joined := make(map[CourseKey]map[CourseKey]bool)
	for _, r := range rows {
		if r.BannerSubject.Valid {
			code := CourseKey{r.GradeRow.Subject, r.GradeRow.CourseNumber}
			if joined[code] == nil {
				joined[code] = make(map[CourseKey]bool, 1)
			}
			joined[code][CourseKey{r.BannerSubject.String, r.BannerCourseNumber.String}] = true
		}
	}
	keys := make([]CourseKey, len(rows))
	for i, r := range rows {
		if r.BannerSubject.Valid {
			keys[i] = CourseKey{r.BannerSubject.String, r.BannerCourseNumber.String}
			continue
		}
		code := CourseKey{r.GradeRow.Subject, r.GradeRow.CourseNumber}
		keys[i] = code
		if banner := joined[code]; len(banner) == 1 {
			for k := range banner {
				keys[i] = k
			}
		}
	}
	return keys
}

// minTrustedMatches is how many CRN joins make a Banner-to-grade name mapping
// trustworthy even when the names share no surname (a name change). Below
// it, such a mapping is usually a co-taught section joining the wrong
// instructor.
const minTrustedMatches = 3

// professorNames returns each grade-file professor's display name (Banner's
// "Last, First", the most-joined one when there are several) and each
// Banner name's professor key, from the trusted mappings.
func professorNames(mappings []*store.InstructorMapping) (names, keys map[string]string) {
	names = make(map[string]string)
	keys = make(map[string]string)
	best := make(map[string]*store.InstructorMapping)
	for _, m := range mappings {
		if m.MatchCount < minTrustedMatches && !sharesSurname(m.BannerName, m.GradeName) {
			continue
		}
		keys[m.BannerName] = m.GradeName
		b := best[m.GradeName]
		if b == nil || m.MatchCount > b.MatchCount || (m.MatchCount == b.MatchCount && m.BannerName < b.BannerName) {
			best[m.GradeName] = m
		}
	}
	for grade, m := range best {
		names[grade] = m.BannerName
	}
	return names, keys
}

// sharesSurname reports whether a Banner "Last, First" name's surname
// appears in a grade-file "First Last" name.
func sharesSurname(banner, grade string) bool {
	last, _, _ := strings.Cut(banner, ",")
	last = strings.TrimSpace(last)
	return last != "" && strings.Contains(strings.ToLower(grade), strings.ToLower(last))
}

func levelOf(k CourseKey) (levelKey, bool) {
	if k.Number == "" || k.Number[0] < '0' || k.Number[0] > '9' {
		return levelKey{}, false
	}
	return levelKey{k.Subject, k.Number[0]}, true
}

// Courses returns every course with grade data, by subject then number.
func (x *Index) Courses() []*Course { return x.list }

// Subjects returns the subjects of Courses, in order.
func (x *Index) Subjects() []Subject { return x.subjects }

// Course returns nil for a course with no grade data.
func (x *Index) Course(k CourseKey) *Course { return x.courses[k] }

// Section returns nil for a section with no grade data.
func (x *Index) Section(term, crn string) *Section { return x.sections[[2]string{term, crn}] }

// Level sums every course in k's subject at its hundreds level (CSCI 3xx).
func (x *Index) Level(k CourseKey) (Counts, bool) {
	lk, ok := levelOf(k)
	if !ok {
		return Counts{}, false
	}
	c, ok := x.levels[lk]
	return c, ok
}

// ProfessorKey returns the key a Banner instructor's grade rows group under:
// their grade-file name, or the Banner name itself when unmapped.
func (x *Index) ProfessorKey(bannerName string) string {
	return cmp.Or(x.keys[bannerName], bannerName)
}

// DisplayName returns a professor key's name in Banner's "Last, First" form
// when one is known.
func (x *Index) DisplayName(key string) string {
	return cmp.Or(x.names[key], key)
}
