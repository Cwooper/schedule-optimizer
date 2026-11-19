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
	Professor string // see professorNames; "" when unlisted
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
	subjects []Subject
	first    string
	last     string
}

// indexSource is what buildIndex reads, one query each.
type indexSource struct {
	rows     []*store.GetGradeRowsWithBannerRow
	titles   []*store.GetLatestCourseTitlesRow
	subjects []*store.GetLatestSubjectDescriptionsRow
}

func buildIndex(src indexSource) *Index {
	idx := &Index{
		courses:  make(map[CourseKey]*Course),
		sections: make(map[[2]string]*Section, len(src.rows)),
	}
	names := professorNames(src.rows)
	keys := courseKeys(src.rows)
	gradeTitles := make(map[CourseKey]string)
	newest := make(map[CourseKey]string)
	for i, r := range src.rows {
		g := &r.GradeRow
		sec := &Section{
			Term:      g.Term,
			CRN:       g.Crn,
			Course:    keys[i],
			Professor: names[strings.TrimSpace(g.Professor)],
			Students:  int(g.StudentsEnrolled),
			Counts:    countsOf(g),
		}
		if sec.Professor == "" {
			sec.Professor = r.BannerInstructor.String
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
		if idx.first == "" || sec.Term < idx.first {
			idx.first = sec.Term
		}
		idx.last = max(idx.last, sec.Term)
	}

	titles := bannerTitles(src.titles)
	descriptions := make(map[string]string, len(src.subjects))
	for _, s := range src.subjects {
		descriptions[s.Subject] = s.SubjectDescription.String
	}

	for k, c := range idx.courses {
		c.Title = cmp.Or(titles[k], gradeTitles[k])
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

// bannerTitles picks each course's title from its newest Banner term: the
// title most of its sections carry. Grade titles are abbreviated
// ("Experimentl Approaches Resrch") and courses get renamed, so Banner's
// wins, except on a tie (a topics course whose sections each have their
// own), which is left out so the grade file's title stands.
func bannerTitles(rows []*store.GetLatestCourseTitlesRow) map[CourseKey]string {
	type best struct {
		title    string
		sections int64
		tied     bool
	}
	top := make(map[CourseKey]best, len(rows))
	for _, t := range rows {
		k := CourseKey{t.Subject, t.CourseNumber}
		switch b := top[k]; {
		case t.Sections > b.sections:
			top[k] = best{t.Title, t.Sections, false}
		case t.Sections == b.sections:
			b.tied = true
			top[k] = b
		}
	}
	titles := make(map[CourseKey]string, len(top))
	for k, b := range top {
		if !b.tied {
			titles[k] = b.title
		}
	}
	return titles
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

// minTrustedMatches is how many CRN joins make a grade-file name's Banner
// name trustworthy even when they share no surname (a name change). Below
// it, such a pair is usually a co-taught section joining the other
// instructor.
const minTrustedMatches = 3

// professorNames maps each grade-file professor ("First Last") to the name
// their sections group and display under: the Banner "Last, First" their
// sections most often joined among the trusted ones, else the grade name
// itself.
// Grouping by the Banner name merges grade-file variants of one person
// (Grace Dunbar Miller and Grace Dunbar-Miller, or a maiden name), and lets
// Banner's later terms match by name alone. A missing key maps to "".
func professorNames(rows []*store.GetGradeRowsWithBannerRow) map[string]string {
	joins := make(map[string]map[string]int)
	for _, r := range rows {
		grade := strings.TrimSpace(r.GradeRow.Professor)
		if grade == "" {
			continue
		}
		if joins[grade] == nil {
			joins[grade] = make(map[string]int)
		}
		if banner := r.BannerInstructor.String; banner != "" {
			joins[grade][banner]++
		}
	}
	names := make(map[string]string, len(joins))
	for grade, banners := range joins {
		names[grade] = grade
		best, n := "", 0
		for b, c := range banners {
			if c < minTrustedMatches && !sharesSurname(b, grade) {
				continue
			}
			if c > n || (c == n && b < best) {
				best, n = b, c
			}
		}
		if n > 0 {
			names[grade] = best
		}
	}
	return names
}

// sharesSurname reports whether a Banner "Last, First" name's surname
// appears in a grade-file "First Last" name.
func sharesSurname(banner, grade string) bool {
	last, _, _ := strings.Cut(banner, ",")
	last = strings.TrimSpace(last)
	return last != "" && strings.Contains(strings.ToLower(grade), strings.ToLower(last))
}

// Courses returns every course with grade data, by subject then number.
func (x *Index) Courses() []*Course { return x.list }

// Subjects returns the subjects of Courses, in order.
func (x *Index) Subjects() []Subject { return x.subjects }

// Course returns nil for a course with no grade data.
func (x *Index) Course(k CourseKey) *Course { return x.courses[k] }

// Section returns nil for a section with no grade data.
func (x *Index) Section(term, crn string) *Section { return x.sections[[2]string{term, crn}] }

// Terms returns the first and last terms with grade data.
func (x *Index) Terms() (first, last string) { return x.first, x.last }
