package generator

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"schedule-optimizer/internal/cache"
	"schedule-optimizer/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

// setupRealData loads actual course data from the database for benchmarking.
func setupRealData(b *testing.B) (*Service, string) {
	b.Helper()

	dbPath := os.Getenv("TEST_DB_PATH")
	if dbPath == "" {
		dbPath = "../../../data/schedule.db"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		b.Skipf("Could not open database: %v", err)
	}

	queries := store.New(db)
	scheduleCache := cache.NewScheduleCache(queries)

	term := "202540"
	if err := scheduleCache.LoadTerm(context.Background(), term); err != nil {
		b.Skipf("Could not load term %s: %v", term, err)
	}

	service := NewService(scheduleCache, queries)
	return service, term
}

// BenchmarkGenerate_RealData_5Courses benchmarks with 5 high-section courses.
func BenchmarkGenerate_RealData_5Courses(b *testing.B) {
	service, term := setupRealData(b)
	ctx := context.Background()

	req := GenerateRequest{
		Term: term,
		CourseSpecs: []CourseSpec{
			{Subject: "ENG", CourseNumber: "101"},
			{Subject: "CHEM", CourseNumber: "161"},
			{Subject: "BIOL", CourseNumber: "101"},
			{Subject: "MATH", CourseNumber: "112"},
			{Subject: "PHYS", CourseNumber: "161"},
		},
		MinCourses: 2,
		MaxCourses: 5,
	}

	b.ResetTimer()
	for b.Loop() {
		resp, err := service.Generate(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
		_ = resp
	}
}

// BenchmarkGenerate_RealData_8Courses benchmarks with 8 courses.
func BenchmarkGenerate_RealData_8Courses(b *testing.B) {
	service, term := setupRealData(b)
	ctx := context.Background()

	req := GenerateRequest{
		Term: term,
		CourseSpecs: []CourseSpec{
			{Subject: "ENG", CourseNumber: "101"},
			{Subject: "CHEM", CourseNumber: "161"},
			{Subject: "BIOL", CourseNumber: "101"},
			{Subject: "MATH", CourseNumber: "112"},
			{Subject: "PHYS", CourseNumber: "161"},
			{Subject: "GEOL", CourseNumber: "101"},
			{Subject: "MATH", CourseNumber: "124"},
			{Subject: "BIOL", CourseNumber: "204"},
		},
		MinCourses: 2,
		MaxCourses: 8,
	}

	b.ResetTimer()
	for b.Loop() {
		resp, err := service.Generate(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
		_ = resp
	}
}

// BenchmarkGenerate_RealData_10Courses benchmarks with 10 courses (realistic student load).
func BenchmarkGenerate_RealData_10Courses(b *testing.B) {
	service, term := setupRealData(b)
	ctx := context.Background()

	req := GenerateRequest{
		Term: term,
		CourseSpecs: []CourseSpec{
			{Subject: "ENG", CourseNumber: "101"},
			{Subject: "CHEM", CourseNumber: "161"},
			{Subject: "BIOL", CourseNumber: "101"},
			{Subject: "MATH", CourseNumber: "112"},
			{Subject: "PHYS", CourseNumber: "161"},
			{Subject: "GEOL", CourseNumber: "101"},
			{Subject: "MATH", CourseNumber: "124"},
			{Subject: "BIOL", CourseNumber: "204"},
			{Subject: "HNRS", CourseNumber: "103"},
			{Subject: "COMM", CourseNumber: "101"},
		},
		MinCourses: 2,
		MaxCourses: 8,
	}

	b.ResetTimer()
	for b.Loop() {
		resp, err := service.Generate(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
		_ = resp
	}
}

// BenchmarkGenerate_RealData_13Courses benchmarks with max input courses.
func BenchmarkGenerate_RealData_13Courses(b *testing.B) {
	service, term := setupRealData(b)
	ctx := context.Background()

	req := GenerateRequest{
		Term: term,
		CourseSpecs: []CourseSpec{
			{Subject: "ENG", CourseNumber: "101"},
			{Subject: "CHEM", CourseNumber: "161"},
			{Subject: "BIOL", CourseNumber: "101"},
			{Subject: "MATH", CourseNumber: "112"},
			{Subject: "PHYS", CourseNumber: "161"},
			{Subject: "GEOL", CourseNumber: "101"},
			{Subject: "MATH", CourseNumber: "124"},
			{Subject: "BIOL", CourseNumber: "204"},
			{Subject: "HNRS", CourseNumber: "103"},
			{Subject: "CHEM", CourseNumber: "163"},
			{Subject: "CSCI", CourseNumber: "141"},
			{Subject: "COMM", CourseNumber: "101"},
			{Subject: "ANTH", CourseNumber: "201"},
		},
		MinCourses: 3,
		MaxCourses: 8,
	}

	b.ResetTimer()
	for b.Loop() {
		resp, err := service.Generate(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
		_ = resp
	}
}

// BenchmarkGenerate_RealData_WithRequired benchmarks with required courses.
func BenchmarkGenerate_RealData_WithRequired(b *testing.B) {
	service, term := setupRealData(b)
	ctx := context.Background()

	req := GenerateRequest{
		Term: term,
		CourseSpecs: []CourseSpec{
			{Subject: "MATH", CourseNumber: "112", Required: true, AllowedCRNs: []string{"40346"}}, // Force a specific section
			{Subject: "CHEM", CourseNumber: "161"},
			{Subject: "BIOL", CourseNumber: "101"},
			{Subject: "PHYS", CourseNumber: "161"},
			{Subject: "GEOL", CourseNumber: "101"},
		},
		MinCourses: 2,
		MaxCourses: 6,
	}

	b.ResetTimer()
	for b.Loop() {
		resp, err := service.Generate(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
		_ = resp
	}
}

// TestGenerate_RealData_Stats prints generation stats for analysis.
func TestGenerate_RealData_Stats(t *testing.T) {
	dbPath := os.Getenv("TEST_DB_PATH")
	if dbPath == "" {
		dbPath = "../../../data/schedule.db"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Skipf("Could not open database: %v", err)
	}

	queries := store.New(db)
	scheduleCache := cache.NewScheduleCache(queries)

	term := "202540"
	if err := scheduleCache.LoadTerm(context.Background(), term); err != nil {
		t.Skipf("Could not load term %s: %v", term, err)
	}

	service := NewService(scheduleCache, queries)
	ctx := context.Background()

	testCases := []struct {
		name  string
		specs []CourseSpec
		min   int
		max   int
	}{
		{"5 courses", []CourseSpec{
			{Subject: "ENG", CourseNumber: "101"},
			{Subject: "CHEM", CourseNumber: "161"},
			{Subject: "BIOL", CourseNumber: "101"},
			{Subject: "MATH", CourseNumber: "112"},
			{Subject: "PHYS", CourseNumber: "161"},
		}, 2, 5},
		{"8 courses", []CourseSpec{
			{Subject: "ENG", CourseNumber: "101"},
			{Subject: "CHEM", CourseNumber: "161"},
			{Subject: "BIOL", CourseNumber: "101"},
			{Subject: "MATH", CourseNumber: "112"},
			{Subject: "PHYS", CourseNumber: "161"},
			{Subject: "GEOL", CourseNumber: "101"},
			{Subject: "MATH", CourseNumber: "124"},
			{Subject: "BIOL", CourseNumber: "204"},
		}, 2, 8},
		{"10 courses", []CourseSpec{
			{Subject: "ENG", CourseNumber: "101"},
			{Subject: "CHEM", CourseNumber: "161"},
			{Subject: "BIOL", CourseNumber: "101"},
			{Subject: "MATH", CourseNumber: "112"},
			{Subject: "PHYS", CourseNumber: "161"},
			{Subject: "GEOL", CourseNumber: "101"},
			{Subject: "MATH", CourseNumber: "124"},
			{Subject: "BIOL", CourseNumber: "204"},
			{Subject: "HNRS", CourseNumber: "103"},
			{Subject: "COMM", CourseNumber: "101"},
		}, 2, 8},
		{"13 courses", []CourseSpec{
			{Subject: "ENG", CourseNumber: "101"},
			{Subject: "CHEM", CourseNumber: "161"},
			{Subject: "BIOL", CourseNumber: "101"},
			{Subject: "MATH", CourseNumber: "112"},
			{Subject: "PHYS", CourseNumber: "161"},
			{Subject: "GEOL", CourseNumber: "101"},
			{Subject: "MATH", CourseNumber: "124"},
			{Subject: "BIOL", CourseNumber: "204"},
			{Subject: "HNRS", CourseNumber: "103"},
			{Subject: "CHEM", CourseNumber: "163"},
			{Subject: "CSCI", CourseNumber: "141"},
			{Subject: "COMM", CourseNumber: "101"},
			{Subject: "ANTH", CourseNumber: "201"},
		}, 3, 8},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := GenerateRequest{
				Term:        term,
				CourseSpecs: tc.specs,
				MinCourses:  tc.min,
				MaxCourses:  tc.max,
			}

			resp, err := service.Generate(ctx, req)
			if err != nil {
				t.Fatal(err)
			}

			totalSections := 0
			for _, cr := range resp.CourseResults {
				totalSections += cr.Count
			}

			t.Logf("Courses: %d, Sections: %d, Generated: %d, Time: %.2fms",
				len(tc.specs), totalSections, resp.Stats.TotalGenerated, resp.Stats.TimeMs)
		})
	}
}
