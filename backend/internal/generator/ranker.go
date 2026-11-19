package generator

import "slices"

// ranker keeps the best n schedules by compareSchedules as backtracking visits
// them, so memory stays bounded however many combinations a request has.
type ranker struct {
	sc   scorer
	n    int
	kept   []Schedule
	sorted bool // kept[:n] is the best n so far, in order; set by the first trim
	// scratch scores a candidate without allocating; only schedules that make
	// the cut get their own copy.
	scratch     Schedule
	fullCourses int // schedules with at least this many courses count as full
	total, full int
}

func newRanker(sc scorer, n, fullCourses int) *ranker {
	return &ranker{sc: sc, n: n, fullCourses: fullCourses}
}

func (r *ranker) visit(sel []*sectionData) {
	r.total++
	if len(sel) >= r.fullCourses {
		r.full++
	}

	r.scratch.Courses = r.scratch.Courses[:0]
	for _, s := range sel {
		r.scratch.Courses = append(r.scratch.Courses, s.course)
	}
	r.scratch.Score = r.sc.score(&r.scratch)

	// Once sorted, kept[n-1] only improves with later trims, so anything that
	// can't beat it now never will.
	if r.sorted && compareSchedules(r.scratch, r.kept[r.n-1]) >= 0 {
		return
	}
	r.kept = append(r.kept, Schedule{Courses: slices.Clone(r.scratch.Courses), Score: r.scratch.Score})
	if len(r.kept) >= 2*r.n {
		r.trim()
	}
}

func (r *ranker) trim() {
	slices.SortFunc(r.kept, compareSchedules)
	keep := min(len(r.kept), r.n)
	clear(r.kept[keep:])
	r.kept = r.kept[:keep]
	r.sorted = keep == r.n
}

// results returns the kept schedules, best first.
func (r *ranker) results() []Schedule {
	r.trim()
	return r.kept
}
