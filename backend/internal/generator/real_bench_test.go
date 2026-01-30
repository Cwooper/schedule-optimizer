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
		Term:       term,
		Courses:    []string{"ENG 101", "CHEM 161", "BIOL 101", "MATH 112", "PHYS 161"},
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
		Term:       term,
		Courses:    []string{"ENG 101", "CHEM 161", "BIOL 101", "MATH 112", "PHYS 161", "GEOL 101", "MATH 124", "BIOL 204"},
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
		Term:       term,
		Courses:    []string{"ENG 101", "CHEM 161", "BIOL 101", "MATH 112", "PHYS 161", "GEOL 101", "MATH 124", "BIOL 204", "HNRS 103", "COMM 101"},
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
		Term:       term,
		Courses:    []string{"ENG 101", "CHEM 161", "BIOL 101", "MATH 112", "PHYS 161", "GEOL 101", "MATH 124", "BIOL 204", "HNRS 103", "CHEM 163", "CSCI 141", "COMM 101", "ANTH 201"},
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

// BenchmarkGenerate_RealData_WithForced benchmarks with forced CRNs.
func BenchmarkGenerate_RealData_WithForced(b *testing.B) {
	service, term := setupRealData(b)
	ctx := context.Background()

	req := GenerateRequest{
		Term:       term,
		Courses:    []string{"CHEM 161", "BIOL 101", "MATH 112", "PHYS 161", "GEOL 101"},
		ForcedCRNs: []string{"40346"}, // Force a specific MATH 112 section
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
		name    string
		courses []string
		min     int
		max     int
	}{
		{"5 courses", []string{"ENG 101", "CHEM 161", "BIOL 101", "MATH 112", "PHYS 161"}, 2, 5},
		{"8 courses", []string{"ENG 101", "CHEM 161", "BIOL 101", "MATH 112", "PHYS 161", "GEOL 101", "MATH 124", "BIOL 204"}, 2, 8},
		{"10 courses", []string{"ENG 101", "CHEM 161", "BIOL 101", "MATH 112", "PHYS 161", "GEOL 101", "MATH 124", "BIOL 204", "HNRS 103", "COMM 101"}, 2, 8},
		{"13 courses", []string{"ENG 101", "CHEM 161", "BIOL 101", "MATH 112", "PHYS 161", "GEOL 101", "MATH 124", "BIOL 204", "HNRS 103", "CHEM 163", "CSCI 141", "COMM 101", "ANTH 201"}, 3, 8},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := GenerateRequest{
				Term:       term,
				Courses:    tc.courses,
				MinCourses: tc.min,
				MaxCourses: tc.max,
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
				len(tc.courses), totalSections, resp.Stats.TotalGenerated, resp.Stats.TimeMs)
		})
	}
}
