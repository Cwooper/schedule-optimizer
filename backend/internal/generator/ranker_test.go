package generator

import (
	"context"
	"slices"
	"testing"
)

// collectAll materializes every schedule backtrack visits.
func collectAll(ctx context.Context, p backtrackParams) []Schedule {
	var out []Schedule
	backtrack(ctx, p, func(sel []*sectionData) {
		out = append(out, buildSchedule(sel))
	})
	return out
}

func TestRanker_MatchesFullSort(t *testing.T) {
	groups := makeTestGroups(6, 4, true)
	for i, g := range groups {
		for j, s := range g.sections {
			s.course.GPA = 2 + float64((i*3+j)%5)/2
			s.course.IsOpen = (i+j)%3 != 0
		}
	}
	params := backtrackParams{groups: groups, minCourses: 1, maxCourses: 6, limit: MaxSchedulesToGenerate}
	sc := newScorer(DefaultPreferences())

	var visits [][]*sectionData
	backtrack(context.Background(), params, func(sel []*sectionData) {
		visits = append(visits, slices.Clone(sel))
	})
	want := make([]Schedule, len(visits))
	for i, sel := range visits {
		want[i] = buildSchedule(sel)
		want[i].Score = sc.score(&want[i])
	}
	slices.SortFunc(want, compareSchedules)

	// Pruning before the buffer is sorted drops good schedules that arrive
	// after a better one, so interleave the worst and best.
	worstFirst := slices.Clone(visits)
	slices.SortFunc(worstFirst, func(a, b []*sectionData) int {
		return -compareSchedules(scored(sc, a), scored(sc, b))
	})
	var interleaved [][]*sectionData
	for lo, hi := 0, len(worstFirst)-1; lo <= hi; lo, hi = lo+1, hi-1 {
		interleaved = append(interleaved, worstFirst[lo])
		if lo != hi {
			interleaved = append(interleaved, worstFirst[hi])
		}
	}
	orders := map[string][][]*sectionData{
		"enumeration": visits, "worst first": worstFirst, "interleaved": interleaved,
	}

	for name, order := range orders {
		for _, n := range []int{1, 7, len(want) / 2, len(want) - 1, len(want), len(want) + 10} {
			r := newRanker(sc, n, 6)
			for _, sel := range order {
				r.visit(sel)
			}
			got := r.results()

			wantN := want[:min(n, len(want))]
			if len(got) != len(wantN) {
				t.Fatalf("%s, n=%d: got %d schedules, want %d", name, n, len(got), len(wantN))
			}
			for i := range got {
				if compareSchedules(got[i], wantN[i]) != 0 {
					t.Fatalf("%s, n=%d: schedule %d differs from full sort", name, n, i)
				}
			}
			if r.total != len(want) {
				t.Errorf("%s, n=%d: total = %d, want %d", name, n, r.total, len(want))
			}
		}
	}
}

func scored(sc scorer, sel []*sectionData) Schedule {
	s := buildSchedule(sel)
	s.Score = sc.score(&s)
	return s
}
