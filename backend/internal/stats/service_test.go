package stats

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"slices"
	"strconv"
	"testing"
	"time"

	"schedule-optimizer/internal/stats/grades"
	"schedule-optimizer/internal/store"
	"schedule-optimizer/internal/testutil"

	_ "github.com/mattn/go-sqlite3"
)

// seededService has CSCI 305 graded in Fall 2023 and Fall 2024, taught
// ungraded in Summer 2025, and scheduled for Winter 2027, which has not
// started on 2026-11-19.
func seededService(t *testing.T) *Service {
	t.Helper()
	db, queries := testutil.SetupTestDB(t)
	_, err := db.Exec(`
		INSERT INTO terms (code, description) VALUES ('202340', 'Fall 2023'), ('202440', 'Fall 2024'), ('202520', 'Summer 2025'), ('202710', 'Winter 2027');
		INSERT INTO term_dates (term_code, start_date, end_date) VALUES ('202520', '2025-06-23', '2025-08-15'), ('202710', '2027-01-05', '2027-03-19');
		INSERT INTO sections (id, term, crn, subject, course_number, title, credit_hours_low, credit_hours_high, max_enrollment) VALUES
			(1, '202340', '100', 'CSCI', '305', 'Old Title', 2, 4, 40),
			(2, '202340', '101', 'CSCI', '305', 'Old Title', 4, 4, 40),
			(3, '202440', '200', 'CSCI', '305', 'Old Title', 4, 4, 40),
			(6, '202520', '400', 'CSCI', '305', 'Old Title', 4, 4, 40),
			(4, '202710', '300', 'CSCI', '305', 'Data Structures and Algorithms I', 4, 4, 40),
			(5, '202710', '301', 'CSCI', '305', 'Data Structures and Algorithms I', 4, 4, 40);
		-- Section 1's second primary row must not count it twice.
		INSERT INTO instructors (section_id, name, is_primary) VALUES
			(1, 'Smith, Jordan', 1), (2, 'Lee, Kara', 1), (3, 'Smith, Jordan', 1),
			(4, 'Smith, Jordan', 1), (5, 'Gu, Xuhang', 1), (6, 'Smith, Jordan', 1),
			(1, 'Zed, Other', 1);
		INSERT INTO meeting_times (section_id, start_time, end_time, building, room, monday, wednesday, friday)
			VALUES (1, '1000', '1050', 'CF', '225', 1, 1, 1);
		INSERT INTO grade_rows (term, crn, subject, course_number, title, professor, students_enrolled, grade_count, cnt_a, cnt_b, cnt_w) VALUES
			('202340', '100', 'CSCI', '305', 'Analysis of Algs', 'Jordan Smith', 38, 21, 20, 0, 1),
			('202340', '101', 'CSCI', '305', 'Analysis of Algs', 'Kara Lee', 30, 20, 10, 10, 0),
			('202440', '200', 'CSCI', '305', 'Analysis of Algs', 'Jordan Smith', 36, 15, 15, 0, 0),
			('202440', '299', 'CSCI', '305', 'Analysis of Algs', '', 12, 5, 0, 5, 0);
	`)
	if err != nil {
		t.Fatal(err)
	}
	g := grades.NewService(db, queries, "")
	if err := g.LoadFromDB(context.Background()); err != nil {
		t.Fatal(err)
	}
	s := NewService(g, queries)
	s.now = func() time.Time { return time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC) }
	return s
}

func TestCourse(t *testing.T) {
	v, err := seededService(t).Course(context.Background(), "CSCI", "305")
	if err != nil {
		t.Fatal(err)
	}
	if v.Title != "Data Structures and Algorithms I" || v.Sections != 4 || v.Professors != 2 {
		t.Errorf("header = %q, %d sections, %d professors", v.Title, v.Sections, v.Professors)
	}
	// Fall 2024 counts graded CRN 299, which Banner does not list.
	wantOffered := []OfferedTerm{{"202340", 2}, {"202440", 2}, {"202520", 1}, {"202710", 2}}
	if !slices.Equal(v.Offered, wantOffered) {
		t.Errorf("offered = %v", v.Offered)
	}
	if v.Next == nil || v.Next.Term != "202710" || !slices.Equal(v.Next.Professors, []string{"Smith, Jordan", "Gu, Xuhang"}) {
		t.Errorf("next = %+v", v.Next)
	}

	type row struct {
		name       string
		sections   int
		lastTaught string
		next       bool
		reported   bool
	}
	want := []row{
		// Summer 2025 is in Banner only; Winter 2027 has not started.
		{"Smith, Jordan", 2, "202520", true, true},
		{"Lee, Kara", 1, "202340", false, true},
		// Five letter grades are too few to report.
		{"", 1, "", false, false},
	}
	var got []row
	for _, p := range v.ProfessorRows {
		got = append(got, row{p.Name, p.Sections, p.LastTaught, p.TeachesNext, p.GPA != nil && p.Grades != nil})
	}
	if !slices.Equal(got, want) {
		t.Errorf("professor rows =\n%+v\nwant\n%+v", got, want)
	}
	if h := v.ProfessorRows[0].History; h[0].Term != "202440" || h[1].Term != "202340" {
		t.Errorf("history not newest first: %+v", h)
	}
	// A term's GPA sums its sections' counts, the small one included:
	// Fall 2024 is (15*4 + 5*3) / 20.
	type term struct {
		term               string
		value              float64
		sections, students int
	}
	var terms []term
	for _, tv := range v.Terms {
		terms = append(terms, term{tv.Term, *tv.Value, tv.Sections, tv.Students})
	}
	if !slices.Equal(terms, []term{{"202340", 3.75, 2, 68}, {"202440", 3.75, 2, 48}}) {
		t.Errorf("terms = %+v", terms)
	}
	// Every section is a Fall one: 225 grade points over 60 letters.
	var seasons []string
	for _, sv := range v.Seasons {
		val := "-"
		if sv.Value != nil {
			val = strconv.FormatFloat(*sv.Value, 'f', 2, 64)
		}
		seasons = append(seasons, strconv.Itoa(sv.Quarter)+":"+val)
	}
	if want := []string{"40:3.75", "10:-", "20:-", "30:-"}; !slices.Equal(seasons, want) {
		t.Errorf("seasons = %v, want %v", seasons, want)
	}
}

func TestMetric(t *testing.T) {
	tests := []struct {
		name   string
		course grades.Counts
		want   Metric
	}{
		{"letters", grades.Counts{grades.GradeA: 10}, MetricGPA},
		{"S/U only", grades.Counts{grades.GradeS: 10}, MetricPass},
		{"too few letters for a GPA, enough grades for a pass rate", grades.Counts{grades.GradeA: 9, grades.GradeP: 30}, MetricPass},
		{"too small for either, with letters", grades.Counts{grades.GradeA: 5, grades.GradeS: 2}, MetricGPA},
		{"too small for either, without letters", grades.Counts{grades.GradeS: 5}, MetricPass},
	}
	for _, tt := range tests {
		if got := metricOf(tt.course); got != tt.want {
			t.Errorf("%s: metric %q, want %q", tt.name, got, tt.want)
		}
	}
}

// A scope too small for the metric hides it everywhere: a pass-rate
// course hides Pass as a GPA course hides GPA, while a GPA course's Pass
// column stays.
func TestNewStatsHidesUnreportedMetric(t *testing.T) {
	tests := []struct {
		name                 string
		c                    grades.Counts
		metric               Metric
		gpa, pass, histogram bool
	}{
		{"GPA course, small section", grades.Counts{grades.GradeA: 5}, MetricGPA, false, true, false},
		{"GPA course, S/U section", grades.Counts{grades.GradeS: 12}, MetricGPA, false, true, false},
		{"pass course, small section", grades.Counts{grades.GradeS: 6, grades.GradeW: 6}, MetricPass, false, false, false},
		{"pass course, section with a few letters", grades.Counts{grades.GradeA: 2, grades.GradeS: 10}, MetricPass, false, true, true},
	}
	for _, tt := range tests {
		st := newStats(tt.c, 12, tt.metric)
		if (st.GPA != nil) != tt.gpa || (st.Pass != nil) != tt.pass || (st.Grades != nil) != tt.histogram {
			t.Errorf("%s: gpa %v, pass %v, grades %v", tt.name, st.GPA != nil, st.Pass != nil, st.Grades != nil)
		}
	}
}

func TestOfferedTermsUnionsCRNs(t *testing.T) {
	banner := map[string]map[string]bool{
		"202340": {"1": true, "2": true},
		"202710": {"5": true},
	}
	graded := []*grades.Section{
		{Term: "202240", CRN: "9"},
		{Term: "202340", CRN: "2"}, // in Banner too
		{Term: "202340", CRN: "3"}, // missing from a partial scrape
	}
	want := []OfferedTerm{{"202240", 1}, {"202340", 3}, {"202710", 1}}
	if got := offeredTerms(banner, graded); !slices.Equal(got, want) {
		t.Errorf("offered = %v, want %v", got, want)
	}
}

func TestSection(t *testing.T) {
	s := seededService(t)
	ctx := context.Background()
	tests := []struct {
		crn, term string
		rank      int
		siblings  []string
		mine      int
	}{
		// 100 and 200 both have a 4.0, so they share first; 101 has 3.5.
		{"100", "202340", 1, []string{"100", "101"}, 2},
		{"200", "202440", 1, []string{"200", "299"}, 2},
		{"101", "202340", 3, []string{"100", "101"}, 1},
		// Too small to rank, and left out of the dots.
		{"299", "202440", 0, []string{"200", "299"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.crn, func(t *testing.T) {
			v, err := s.Section(ctx, tt.term, tt.crn)
			if err != nil {
				t.Fatal(err)
			}
			if v.Rank != tt.rank {
				t.Errorf("rank = %d, want %d", v.Rank, tt.rank)
			}
			if len(v.Dots) != 3 {
				t.Errorf("dots = %d, want 3", len(v.Dots))
			}
			mine := 0
			for _, d := range v.Dots {
				if d.Mine {
					mine++
				}
			}
			if mine != tt.mine {
				t.Errorf("mine = %d, want %d", mine, tt.mine)
			}
			var sibs []string
			for _, r := range v.Siblings {
				sibs = append(sibs, r.CRN)
			}
			if !slices.Equal(sibs, tt.siblings) {
				t.Errorf("siblings = %v, want %v", sibs, tt.siblings)
			}
		})
	}

	v, err := s.Section(ctx, "202340", "100")
	if err != nil {
		t.Fatal(err)
	}
	if v.Class == nil || v.Class.Capacity != 40 || v.Class.Credits != 2 || v.Class.CreditsHigh != 4 || len(v.Class.Meetings) != 1 {
		t.Errorf("class = %+v", v.Class)
	}
}

func TestNotFoundAndNotLoaded(t *testing.T) {
	ctx := context.Background()
	s := seededService(t)
	if _, err := s.Course(ctx, "CSCI", "999"); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown course: %v", err)
	}
	if _, err := s.Section(ctx, "202340", "999"); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown section: %v", err)
	}

	db, queries := testutil.SetupTestDB(t)
	unloaded := NewService(grades.NewService(db, queries, ""), queries)
	if _, err := unloaded.Courses(); !errors.Is(err, ErrNotLoaded) {
		t.Errorf("before load: %v", err)
	}
}

func BenchmarkCourseLargest(b *testing.B) {
	path := os.Getenv("TEST_DB_PATH")
	if path == "" {
		path = "../../data/schedule.db"
	}
	db, err := sql.Open("sqlite3", "file:"+path+"?mode=ro")
	if err != nil {
		b.Skipf("open database: %v", err)
	}
	defer db.Close()
	queries := store.New(db)
	g := grades.NewService(db, queries, "")
	ctx := context.Background()
	if err := g.LoadFromDB(ctx); err != nil || g.Index().Course(grades.CourseKey{Subject: "ENG", Number: "101"}) == nil {
		b.Skipf("no grade data: %v", err)
	}
	s := NewService(g, queries)
	for b.Loop() {
		if _, err := s.Course(ctx, "ENG", "101"); err != nil {
			b.Fatal(err)
		}
	}
}
