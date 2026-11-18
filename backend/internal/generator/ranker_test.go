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

	want := collectAll(context.Background(), params)
	for i := range want {
		want[i].Score = sc.score(&want[i])
	}
	slices.SortFunc(want, compareSchedules)

	for _, n := range []int{1, 7, 50, len(want), len(want) + 10} {
		r := newRanker(sc, n, 6)
		backtrack(context.Background(), params, r.visit)
		got := r.results()

		wantN := want[:min(n, len(want))]
		if len(got) != len(wantN) {
			t.Fatalf("n=%d: got %d schedules, want %d", n, len(got), len(wantN))
		}
		for i := range got {
			if compareSchedules(got[i], wantN[i]) != 0 {
				t.Fatalf("n=%d: schedule %d differs from full sort", n, i)
			}
		}
		if r.total != len(want) {
			t.Errorf("n=%d: total = %d, want %d", n, r.total, len(want))
		}
	}
}
