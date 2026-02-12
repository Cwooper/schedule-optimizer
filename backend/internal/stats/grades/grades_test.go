package grades

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"schedule-optimizer/internal/store"
)

// setupTestDB creates an in-memory SQLite database with both migrations applied.
func setupTestDB(t testing.TB) (*sql.DB, *store.Queries) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	root := getProjectRoot()
	for _, migration := range []string{
		"migrations/000001_initial_schema.up.sql",
		"migrations/000002_add_grade_tables.up.sql",
	} {
		schema, err := os.ReadFile(filepath.Join(root, migration))
		if err != nil {
			t.Fatalf("failed to read %s: %v", migration, err)
		}
		if _, err := db.Exec(string(schema)); err != nil {
			t.Fatalf("failed to apply %s: %v", migration, err)
		}
	}

	return db, store.New(db)
}

func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("could not find project root")
		}
		dir = parent
	}
}

// almostEqual compares floats within a tolerance.
func almostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

func TestComputeGPA(t *testing.T) {
	tests := []struct {
		name                               string
		a, am, bp, b, bm, cp, c, cm       int
		dp, d, dm, f                       int
		want                               float64
	}{
		{
			name: "all As",
			a:    10,
			want: 4.0,
		},
		{
			name: "all Fs",
			f:    5,
			want: 0.0,
		},
		{
			name: "no grades",
			want: 0.0,
		},
		{
			name: "mixed grades",
			a:    5, b: 5, c: 5, f: 5,
			// (5*4.0 + 5*3.0 + 5*2.0 + 5*0.0) / 20 = 45/20 = 2.25
			want: 2.25,
		},
		{
			name: "single A-minus",
			am:   1,
			want: 3.7,
		},
		{
			name: "B+/B- average",
			bp:   1, bm: 1,
			// (3.3 + 2.7) / 2 = 3.0
			want: 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeGPA(tt.a, tt.am, tt.bp, tt.b, tt.bm,
				tt.cp, tt.c, tt.cm, tt.dp, tt.d, tt.dm, tt.f)
			if !almostEqual(got, tt.want, 0.001) {
				t.Errorf("computeGPA() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComputePassRate(t *testing.T) {
	tests := []struct {
		name       string
		s, u, p, np int
		wantNil    bool
		want       float64
	}{
		{
			name:    "no data",
			wantNil: true,
		},
		{
			name: "all pass",
			s:    10, p: 5,
			want: 1.0,
		},
		{
			name: "all fail",
			u:    3, np: 7,
			want: 0.0,
		},
		{
			name: "mixed",
			s:    6, u: 2, p: 4, np: 8,
			// (6+4) / (6+2+4+8) = 10/20 = 0.5
			want: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computePassRate(tt.s, tt.u, tt.p, tt.np)
			if tt.wantNil {
				if got != nil {
					t.Errorf("computePassRate() = %v, want nil", *got)
				}
				return
			}
			if got == nil {
				t.Fatal("computePassRate() = nil, want non-nil")
			}
			if !almostEqual(*got, tt.want, 0.001) {
				t.Errorf("computePassRate() = %v, want %v", *got, tt.want)
			}
		})
	}
}

func TestPickBest(t *testing.T) {
	tests := []struct {
		name       string
		candidates map[string]int
		wantKey    string
		wantCount  int
	}{
		{
			name:       "single candidate",
			candidates: map[string]int{"CSCI": 10},
			wantKey:    "CSCI",
			wantCount:  10,
		},
		{
			name:       "clear winner",
			candidates: map[string]int{"CSCI": 100, "CS": 5},
			wantKey:    "CSCI",
			wantCount:  100,
		},
		{
			name:       "empty map",
			candidates: map[string]int{},
			wantKey:    "",
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, count := pickBest(tt.candidates)
			if key != tt.wantKey {
				t.Errorf("pickBest() key = %q, want %q", key, tt.wantKey)
			}
			if count != tt.wantCount {
				t.Errorf("pickBest() count = %d, want %d", count, tt.wantCount)
			}
		})
	}
}

func TestLookupSectionGPA(t *testing.T) {
	db, queries := setupTestDB(t)

	svc := NewService(db, queries, "")

	// Populate maps directly for unit testing (bypass DB pipeline)
	svc.mu.Lock()
	svc.subjectMap = map[string]string{
		"CSCI": "CS",   // Banner "CSCI" maps to grade-file "CS"
		"MATH": "MATH", // identity mapping
	}
	svc.instructorMap = map[string]string{
		"Dr. Smith": "Smith, John",
	}
	pr95 := 0.95
	pr80 := 0.80
	svc.courseProfAgg = map[string]*GradeAggregate{
		"CS:247:Smith, John": {GPA: 3.5},
		"CS:590:Smith, John": {GPA: 0, PassRate: &pr95},
		"CS:301:Smith, John": {GPA: 3.1, PassRate: &pr80},
	}
	svc.courseAgg = map[string]*GradeAggregate{
		"CS:247":   {GPA: 3.2},
		"MATH:204": {GPA: 2.9},
		"CS:590":   {GPA: 0, PassRate: &pr95},
		"CS:301":   {GPA: 3.0, PassRate: &pr80},
	}
	svc.professorAgg = map[string]*GradeAggregate{
		"Smith, John": {GPA: 3.4},
	}
	svc.subjectAgg = map[string]*GradeAggregate{
		"CS": {GPA: 3.1},
	}
	svc.loaded = true
	svc.mu.Unlock()

	t.Run("course+professor match", func(t *testing.T) {
		gpa, passRate, source := svc.LookupSectionGPA("CSCI", "247", "Dr. Smith")
		if !almostEqual(gpa, 3.5, 0.001) {
			t.Errorf("GPA = %v, want 3.5", gpa)
		}
		if passRate != nil {
			t.Errorf("passRate = %v, want nil", *passRate)
		}
		if source != "course_professor" {
			t.Errorf("source = %q, want %q", source, "course_professor")
		}
	})

	t.Run("falls back to course when instructor unmapped", func(t *testing.T) {
		gpa, _, source := svc.LookupSectionGPA("CSCI", "247", "Dr. Unknown")
		if !almostEqual(gpa, 3.2, 0.001) {
			t.Errorf("GPA = %v, want 3.2", gpa)
		}
		if source != "course" {
			t.Errorf("source = %q, want %q", source, "course")
		}
	})

	t.Run("falls back to course when no prof aggregate", func(t *testing.T) {
		gpa, _, source := svc.LookupSectionGPA("MATH", "204", "Dr. Brown")
		if !almostEqual(gpa, 2.9, 0.001) {
			t.Errorf("GPA = %v, want 2.9", gpa)
		}
		if source != "course" {
			t.Errorf("source = %q, want %q", source, "course")
		}
	})

	t.Run("no data returns zero and empty source", func(t *testing.T) {
		gpa, passRate, source := svc.LookupSectionGPA("PHYS", "101", "Anyone")
		if gpa != 0 {
			t.Errorf("GPA = %v, want 0", gpa)
		}
		if passRate != nil {
			t.Errorf("passRate = %v, want nil", *passRate)
		}
		if source != "" {
			t.Errorf("source = %q, want empty", source)
		}
	})

	t.Run("pure S/U course returns pass rate", func(t *testing.T) {
		gpa, passRate, source := svc.LookupSectionGPA("CSCI", "590", "Dr. Smith")
		if gpa != 0 {
			t.Errorf("GPA = %v, want 0", gpa)
		}
		if passRate == nil {
			t.Fatal("passRate = nil, want non-nil")
		}
		if !almostEqual(*passRate, 0.95, 0.001) {
			t.Errorf("passRate = %v, want 0.95", *passRate)
		}
		if source != "course_professor" {
			t.Errorf("source = %q, want %q", source, "course_professor")
		}
	})

	t.Run("pure S/U course falls back to course level", func(t *testing.T) {
		gpa, passRate, source := svc.LookupSectionGPA("CSCI", "590", "Dr. Unknown")
		if gpa != 0 {
			t.Errorf("GPA = %v, want 0", gpa)
		}
		if passRate == nil {
			t.Fatal("passRate = nil, want non-nil")
		}
		if !almostEqual(*passRate, 0.95, 0.001) {
			t.Errorf("passRate = %v, want 0.95", *passRate)
		}
		if source != "course" {
			t.Errorf("source = %q, want %q", source, "course")
		}
	})

	t.Run("mixed course returns both GPA and pass rate", func(t *testing.T) {
		gpa, passRate, source := svc.LookupSectionGPA("CSCI", "301", "Dr. Smith")
		if !almostEqual(gpa, 3.1, 0.001) {
			t.Errorf("GPA = %v, want 3.1", gpa)
		}
		if passRate == nil {
			t.Fatal("passRate = nil, want non-nil")
		}
		if !almostEqual(*passRate, 0.80, 0.001) {
			t.Errorf("passRate = %v, want 0.80", *passRate)
		}
		if source != "course_professor" {
			t.Errorf("source = %q, want %q", source, "course_professor")
		}
	})
}

func TestLookupCourseGPA(t *testing.T) {
	db, queries := setupTestDB(t)

	svc := NewService(db, queries, "")

	svc.mu.Lock()
	svc.subjectMap = map[string]string{"CSCI": "CS"}
	svc.courseAgg = map[string]*GradeAggregate{
		"CS:247": {GPA: 3.2},
	}
	svc.loaded = true
	svc.mu.Unlock()

	t.Run("found", func(t *testing.T) {
		gpa, passRate, ok := svc.LookupCourseGPA("CSCI", "247")
		if !ok {
			t.Fatal("expected found=true")
		}
		if !almostEqual(gpa, 3.2, 0.001) {
			t.Errorf("GPA = %v, want 3.2", gpa)
		}
		if passRate != nil {
			t.Errorf("passRate = %v, want nil", *passRate)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, _, ok := svc.LookupCourseGPA("PHYS", "101")
		if ok {
			t.Error("expected found=false")
		}
	})

	t.Run("identity mapping when no subject mapping exists", func(t *testing.T) {
		svc.mu.Lock()
		svc.courseAgg["MATH:204"] = &GradeAggregate{GPA: 2.8}
		svc.mu.Unlock()

		gpa, _, ok := svc.LookupCourseGPA("MATH", "204")
		if !ok {
			t.Fatal("expected found=true with identity mapping")
		}
		if !almostEqual(gpa, 2.8, 0.001) {
			t.Errorf("GPA = %v, want 2.8", gpa)
		}
	})

	t.Run("pure S/U course returns pass rate", func(t *testing.T) {
		pr := 0.99
		svc.mu.Lock()
		svc.courseAgg["CS:591"] = &GradeAggregate{GPA: 0, PassRate: &pr}
		svc.mu.Unlock()

		gpa, passRate, ok := svc.LookupCourseGPA("CSCI", "591")
		if !ok {
			t.Fatal("expected found=true")
		}
		if gpa != 0 {
			t.Errorf("GPA = %v, want 0", gpa)
		}
		if passRate == nil {
			t.Fatal("passRate = nil, want non-nil")
		}
		if !almostEqual(*passRate, 0.99, 0.001) {
			t.Errorf("passRate = %v, want 0.99", *passRate)
		}
	})
}

func TestMapSubjectAndInstructor(t *testing.T) {
	db, queries := setupTestDB(t)
	svc := NewService(db, queries, "")

	svc.mu.Lock()
	svc.subjectMap = map[string]string{"CSCI": "CS"}
	svc.instructorMap = map[string]string{
		"Dr. Smith":       "Smith, John",
		"O'Neil, Gregory": "Gregory O'Neil",
		"Montaño, Manuel": "Manuel Montaño",
	}
	svc.mu.Unlock()

	t.Run("mapSubject with mapping", func(t *testing.T) {
		if got := svc.mapSubject("CSCI"); got != "CS" {
			t.Errorf("mapSubject(CSCI) = %q, want CS", got)
		}
	})

	t.Run("mapSubject identity fallback", func(t *testing.T) {
		if got := svc.mapSubject("MATH"); got != "MATH" {
			t.Errorf("mapSubject(MATH) = %q, want MATH", got)
		}
	})

	t.Run("mapInstructor with mapping", func(t *testing.T) {
		if got := svc.mapInstructor("Dr. Smith"); got != "Smith, John" {
			t.Errorf("mapInstructor(Dr. Smith) = %q, want Smith, John", got)
		}
	})

	t.Run("mapInstructor unescapes HTML entities", func(t *testing.T) {
		// Banner stores names with HTML entities like O&#39;Neil
		if got := svc.mapInstructor("O&#39;Neil, Gregory"); got != "Gregory O'Neil" {
			t.Errorf("mapInstructor(O&#39;Neil) = %q, want Gregory O'Neil", got)
		}
		// Named entities like &ntilde;
		if got := svc.mapInstructor("Monta&ntilde;o, Manuel"); got != "Manuel Montaño" {
			t.Errorf("mapInstructor(Monta&ntilde;o) = %q, want Manuel Montaño", got)
		}
	})

	t.Run("mapInstructor returns empty for unknown", func(t *testing.T) {
		if got := svc.mapInstructor("Unknown"); got != "" {
			t.Errorf("mapInstructor(Unknown) = %q, want empty", got)
		}
	})
}

func TestIsLoaded(t *testing.T) {
	db, queries := setupTestDB(t)
	svc := NewService(db, queries, "")

	if svc.IsLoaded() {
		t.Error("new service should not be loaded")
	}

	svc.mu.Lock()
	svc.loaded = true
	svc.mu.Unlock()

	if !svc.IsLoaded() {
		t.Error("service should be loaded after setting flag")
	}
}

func TestImportAndComputeIdempotency(t *testing.T) {
	db, queries := setupTestDB(t)
	ctx := context.Background()

	// Seed grade_rows so import is skipped
	_, err := db.Exec(`INSERT INTO grade_rows (term, crn, subject, course_number, title, professor,
		students_enrolled, grade_count, cnt_a, cnt_am, cnt_bp, cnt_b, cnt_bm,
		cnt_cp, cnt_c, cnt_cm, cnt_dp, cnt_d, cnt_dm, cnt_f, cnt_w, cnt_p, cnt_np, cnt_s, cnt_u)
		VALUES ('202520', '20001', 'CSCI', '247', 'Data Structures', 'Smith',
		30, 28, 10, 5, 3, 4, 2, 1, 1, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0)`)
	if err != nil {
		t.Fatalf("seed grade_rows: %v", err)
	}

	svc := NewService(db, queries, "nonexistent.xlsx")

	// Should not fail even though xlsx doesn't exist — import is skipped (count > 0)
	if err := svc.ImportAndCompute(ctx); err != nil {
		t.Fatalf("ImportAndCompute failed: %v", err)
	}

	if !svc.IsLoaded() {
		t.Error("service should be loaded after ImportAndCompute")
	}

	// Aggregates should have been computed from the seeded row
	aggCount, _ := queries.GetGradeAggregateCount(ctx)
	if aggCount == 0 {
		t.Error("expected aggregates to be computed")
	}
}

func TestLoadFromDB(t *testing.T) {
	db, queries := setupTestDB(t)
	ctx := context.Background()

	// Seed mappings and aggregates directly
	_, err := db.Exec(`
		INSERT INTO subject_mappings (banner_subject, grade_subject, match_count) VALUES ('CSCI', 'CS', 50);
		INSERT INTO instructor_mappings (banner_name, grade_name, match_count) VALUES ('Dr. Smith', 'Smith, John', 30);
		INSERT INTO grade_aggregates (level, subject, course_number, instructor,
			sections_count, students_count, cnt_a, cnt_am, cnt_bp, cnt_b, cnt_bm,
			cnt_cp, cnt_c, cnt_cm, cnt_dp, cnt_d, cnt_dm, cnt_f, cnt_w, cnt_p, cnt_np, cnt_s, cnt_u,
			gpa, pass_rate)
		VALUES
			('course_professor', 'CS', '247', 'Smith, John', 5, 120, 40, 20, 15, 15, 10, 5, 5, 3, 2, 2, 1, 2, 0, 0, 0, 0, 0, 3.45, NULL),
			('course', 'CS', '247', '', 10, 250, 80, 40, 30, 30, 20, 10, 10, 6, 4, 4, 2, 4, 0, 0, 0, 0, 0, 3.30, NULL),
			('subject', 'CS', '', '', 50, 1200, 400, 200, 150, 150, 100, 50, 50, 30, 20, 20, 10, 20, 0, 0, 0, 0, 0, 3.20, NULL);
	`)
	if err != nil {
		t.Fatalf("seed data: %v", err)
	}

	svc := NewService(db, queries, "")
	if err := svc.LoadFromDB(ctx); err != nil {
		t.Fatalf("LoadFromDB failed: %v", err)
	}

	if !svc.IsLoaded() {
		t.Error("service should be loaded")
	}

	// Verify subject mapping loaded
	if got := svc.mapSubject("CSCI"); got != "CS" {
		t.Errorf("mapSubject(CSCI) = %q, want CS", got)
	}

	// Verify lookup works end-to-end
	gpa, passRate, source := svc.LookupSectionGPA("CSCI", "247", "Dr. Smith")
	if !almostEqual(gpa, 3.45, 0.01) {
		t.Errorf("GPA = %v, want 3.45", gpa)
	}
	if passRate != nil {
		t.Errorf("passRate = %v, want nil", *passRate)
	}
	if source != "course_professor" {
		t.Errorf("source = %q, want course_professor", source)
	}
}
