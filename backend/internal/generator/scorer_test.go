package generator

import (
	"context"
	"math"
	"slices"
	"strings"
	"testing"

	"schedule-optimizer/internal/cache"
	"schedule-optimizer/internal/meeting"
)

var (
	mon = [7]bool{1: true}
	tue = [7]bool{2: true}
	mwf = [7]bool{1: true, 3: true, 5: true}
	sat = [7]bool{6: true}
)

func mt(days [7]bool, start, end string) meeting.Time {
	return meeting.Time{Days: days, StartTime: start, EndTime: end}
}

// scheduleOf gives each meeting its own course.
func scheduleOf(meetings ...meeting.Time) *Schedule {
	s := &Schedule{}
	for _, m := range meetings {
		s.Courses = append(s.Courses, &cache.Course{MeetingTimes: []meeting.Time{m}})
	}
	return s
}

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-3 }

func crnsOf(s Schedule) []string {
	crns := make([]string, 0, len(s.Courses))
	for _, c := range s.Courses {
		crns = append(crns, c.CRN)
	}
	return crns
}

func TestFalloff(t *testing.T) {
	tests := []struct {
		name string
		d, k float64
		want float64
	}{
		{"negative miss", -30, 60, 1},
		{"no miss", 0, 60, 1},
		{"half width", 30, 60, 1 - math.Tanh(0.25)},
		{"at width", 60, 60, 1 - math.Tanh(1)},
		{"twice width", 120, 60, 1 - math.Tanh(4)},
		{"wider curve forgives the same miss", 60, 180, 1 - math.Tanh(1.0/9)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := falloff(tt.d, tt.k); !approx(got, tt.want) {
				t.Errorf("falloff(%v, %v) = %v, want %v", tt.d, tt.k, got, tt.want)
			}
		})
	}
}

func TestWeekStats(t *testing.T) {
	tests := []struct {
		name     string
		schedule *Schedule
		wantBusy bool
		want     [weekdays]dayStats
	}{
		{
			name:     "no timed meetings",
			schedule: scheduleOf(mt(mon, "", ""), mt(sat, "0900", "1000")),
		},
		{
			name:     "gap at the passing-time threshold is ignored",
			schedule: scheduleOf(mt(mon, "0900", "0950"), mt(mon, "1005", "1100")),
			wantBusy: true,
			want:     [weekdays]dayStats{{busy: true, first: 540, last: 660}},
		},
		{
			name:     "gap past the threshold counts in full",
			schedule: scheduleOf(mt(mon, "0900", "0950"), mt(mon, "1010", "1100")),
			wantBusy: true,
			want:     [weekdays]dayStats{{busy: true, first: 540, last: 660, idle: 20}},
		},
		{
			name: "overlapping meetings merge before measuring gaps",
			schedule: scheduleOf(
				mt(mon, "1130", "1200"), // out of order on purpose
				mt(mon, "0900", "1000"),
				mt(mon, "0930", "1100"),
			),
			wantBusy: true,
			want:     [weekdays]dayStats{{busy: true, first: 540, last: 720, idle: 30}},
		},
		{
			name:     "meeting contained in another adds no gap",
			schedule: scheduleOf(mt(mon, "0900", "1200"), mt(mon, "1000", "1050"), mt(mon, "1300", "1350")),
			wantBusy: true,
			want:     [weekdays]dayStats{{busy: true, first: 540, last: 830, idle: 60}},
		},
		{
			name:     "days are independent and days off stay empty",
			schedule: scheduleOf(mt(mwf, "0800", "0850"), mt(tue, "1300", "1350"), mt(mon, "1200", "1250")),
			wantBusy: true,
			want: [weekdays]dayStats{
				{busy: true, first: 480, last: 770, idle: 190},
				{busy: true, first: 780, last: 830},
				{busy: true, first: 480, last: 530},
				{},
				{busy: true, first: 480, last: 530},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, busy := weekStats(tt.schedule)
			if busy != tt.wantBusy {
				t.Errorf("busy = %v, want %v", busy, tt.wantBusy)
			}
			if got != tt.want {
				t.Errorf("weekStats = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestStartEndScores(t *testing.T) {
	prefs := func(start, end int) Preferences {
		p := DefaultPreferences()
		p.Start, p.End = start, end
		return p
	}
	tests := []struct {
		name      string
		prefs     Preferences
		schedule  *Schedule
		wantStart float64
		wantEnd   float64
	}{
		{
			name:      "inside the ideal window",
			prefs:     prefs(2, 2),
			schedule:  scheduleOf(mt(mwf, "1000", "1050"), mt(mwf, "1300", "1400")),
			wantStart: 1,
			wantEnd:   1,
		},
		{
			// 8am is 120 minutes early on 3 of 5 days; Firm width is 120.
			name:      "early start on class days only",
			prefs:     prefs(2, 2),
			schedule:  scheduleOf(mt(mwf, "0800", "0850")),
			wantStart: (3*(1-math.Tanh(1)) + 2) / 5,
			wantEnd:   1,
		},
		{
			name:      "strict level punishes the same miss harder",
			prefs:     prefs(3, 3),
			schedule:  scheduleOf(mt(mwf, "0800", "0850")),
			wantStart: (3*(1-math.Tanh(4)) + 2) / 5,
			wantEnd:   1,
		},
		{
			name:      "loose level forgives it",
			prefs:     prefs(1, 1),
			schedule:  scheduleOf(mt(mwf, "0800", "0850")),
			wantStart: (3*(1-math.Tanh(4.0/9)) + 2) / 5,
			wantEnd:   1,
		},
		{
			// 60 minutes past 2pm on Monday only.
			name:      "late end",
			prefs:     prefs(2, 2),
			schedule:  scheduleOf(mt(mon, "1400", "1500")),
			wantStart: 1,
			wantEnd:   (1 - math.Tanh(0.25) + 4) / 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := newScorer(tt.prefs)
			week, _ := weekStats(tt.schedule)
			if got := sc.startScore(week); !approx(got, tt.wantStart) {
				t.Errorf("startScore = %v, want %v", got, tt.wantStart)
			}
			if got := sc.endScore(week); !approx(got, tt.wantEnd) {
				t.Errorf("endScore = %v, want %v", got, tt.wantEnd)
			}
		})
	}
}

func TestGapScore(t *testing.T) {
	tests := []struct {
		name     string
		schedule *Schedule
		want     float64
	}{
		{"back to back", scheduleOf(mt(mwf, "0900", "0950"), mt(mwf, "1000", "1050")), 1},
		// 3 days x 60 idle minutes = 180 a week, 36 a day.
		{"hour gap three days", scheduleOf(mt(mwf, "0900", "0950"), mt(mwf, "1050", "1150")), 1 - math.Tanh(0.36)},
		// Two hours idle on one day spreads to 24 a day, so it hurts less.
		{"two hour gap one day", scheduleOf(mt(mon, "0900", "1000"), mt(mon, "1200", "1300")), 1 - math.Tanh(0.16)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			week, _ := weekStats(tt.schedule)
			if got := gapScore(week); !approx(got, tt.want) {
				t.Errorf("gapScore = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGPAScore(t *testing.T) {
	withGPAs := func(gpas ...float64) *Schedule {
		s := &Schedule{}
		for _, g := range gpas {
			s.Courses = append(s.Courses, &cache.Course{GPA: g})
		}
		return s
	}
	tests := []struct {
		name     string
		schedule *Schedule
		want     float64
		wantData bool
	}{
		{"no grade data", withGPAs(0, 0), 0, false},
		{"4.0 is full marks", withGPAs(4.0), 1, true},
		{"3.0 is the midpoint", withGPAs(3.0), 0.5, true},
		{"below 2.0 clamps to 0", withGPAs(1.5), 0, true},
		{"missing counts as 3.0", withGPAs(4.0, 0), 0.75, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := gpaScore(tt.schedule)
			if ok != tt.wantData || !approx(got, tt.want) {
				t.Errorf("gpaScore = (%v, %v), want (%v, %v)", got, ok, tt.want, tt.wantData)
			}
		})
	}
}

func TestSeatsScore(t *testing.T) {
	s := &Schedule{Courses: []*cache.Course{{IsOpen: true}, {IsOpen: false}, {IsOpen: true}, {IsOpen: false}}}
	if got := seatsScore(s); got != 0.5 {
		t.Errorf("seatsScore = %v, want 0.5", got)
	}
}

func TestScore(t *testing.T) {
	only := func(set func(*Preferences)) Preferences {
		p := Preferences{IdealStart: "1000", IdealEnd: "1400"}
		set(&p)
		return p
	}
	full := &cache.Course{IsOpen: false, MeetingTimes: []meeting.Time{mt(mwf, "1000", "1050")}}
	open := &cache.Course{IsOpen: true, GPA: 4.0, MeetingTimes: []meeting.Time{mt(tue, "1000", "1050")}}

	tests := []struct {
		name     string
		prefs    Preferences
		schedule *Schedule
		want     float64
	}{
		{"all factors off", Preferences{IdealStart: "1000", IdealEnd: "1400"}, &Schedule{Courses: []*cache.Course{open}}, 0},
		{"a zero factor still counts", only(func(p *Preferences) { p.Seats = 1; p.Start = 1 }), &Schedule{Courses: []*cache.Course{full}}, 0.5},
		{
			name:     "weights scale each factor",
			prefs:    only(func(p *Preferences) { p.Seats = 3; p.Start = 1 }),
			schedule: &Schedule{Courses: []*cache.Course{full, open}},
			want:     (3*0.5 + 1*1) / 4.0,
		},
		{
			name:     "factor without data drops out",
			prefs:    only(func(p *Preferences) { p.GPA = 3; p.Seats = 1 }),
			schedule: &Schedule{Courses: []*cache.Course{{IsOpen: true}}},
			want:     1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newScorer(tt.prefs).score(tt.schedule); !approx(got, tt.want) {
				t.Errorf("score = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScore_EarlierStartRanksLower(t *testing.T) {
	sc := newScorer(DefaultPreferences())
	eight := sc.score(scheduleOf(mt(mwf, "0800", "0850")))
	nine := sc.score(scheduleOf(mt(mwf, "0900", "0950")))
	if eight >= nine {
		t.Errorf("8am score %v should be below 9am score %v", eight, nine)
	}
}

func TestCompareSchedules(t *testing.T) {
	course := func(crn string) *cache.Course { return &cache.Course{CRN: crn} }
	schedules := []Schedule{
		{Courses: []*cache.Course{course("2"), course("9")}, Score: 0.9},
		{Courses: []*cache.Course{course("1"), course("3"), course("5")}, Score: 0.2},
		{Courses: []*cache.Course{course("1"), course("8")}, Score: 0.9},
		{Courses: []*cache.Course{course("1"), course("4")}, Score: 0.95},
	}
	slices.SortFunc(schedules, compareSchedules)

	var got []string
	for _, s := range schedules {
		got = append(got, strings.Join(crnsOf(s), ""))
	}
	want := []string{"135", "14", "18", "29"}
	if !slices.Equal(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

// TestGenerate_PreferencesChangeRanking runs the real Generate path: an early
// pair of sections and a late pair, ranked first by Start, then by End.
func TestGenerate_PreferencesChangeRanking(t *testing.T) {
	section := func(crn, subject, start, end string) *cache.Course {
		return &cache.Course{
			CRN: crn, Subject: subject, CourseNumber: "100",
			MeetingTimes: []meeting.Time{mt(mwf, start, end)},
		}
	}
	term := &cache.TermData{ByCourseCode: map[string][]*cache.Course{
		"A:100": {section("A-early", "A", "0800", "0850"), section("A-late", "A", "1400", "1450")},
		"B:100": {section("B-early", "B", "0900", "0950"), section("B-late", "B", "1500", "1550")},
	}}
	specs := []CourseSpec{{Subject: "A", CourseNumber: "100"}, {Subject: "B", CourseNumber: "100"}}

	tests := []struct {
		name      string
		start     int
		end       int
		wantFirst []string
	}{
		{"strict start prefers late classes", 3, 0, []string{"A-late", "B-late"}},
		{"strict end prefers early classes", 0, 3, []string{"A-early", "B-early"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefs := Preferences{Start: tt.start, End: tt.end, IdealStart: "1000", IdealEnd: "1400"}
			resp, err := NewService(nil).Generate(context.Background(), term, GenerateRequest{
				CourseSpecs: specs,
				Preferences: &prefs,
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(resp.Schedules) != 4 {
				t.Fatalf("got %d schedules, want 4", len(resp.Schedules))
			}
			first, second := resp.Schedules[0], resp.Schedules[1]
			crns := crnsOf(first)
			slices.Sort(crns)
			if !slices.Equal(crns, tt.wantFirst) {
				t.Errorf("first schedule = %v, want %v", crns, tt.wantFirst)
			}
			if margin := first.Score - second.Score; margin < 0.3 {
				t.Errorf("first leads second by %v, want a clear margin (>= 0.3)", margin)
			}
		})
	}
}
