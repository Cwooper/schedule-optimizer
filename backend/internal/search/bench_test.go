package search

import (
	"context"
	"testing"

	"schedule-optimizer/internal/testutil"
)

func BenchmarkSearch_SubjectFilter(b *testing.B) {
	db, queries := testutil.SetupTestDB(b)
	defer db.Close()
	testutil.SeedTestData(b, db)

	svc := NewService(db, queries)
	ctx := context.Background()
	req := SearchRequest{
		Term:    "202520",
		Subject: "CSCI",
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := svc.Search(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSearch_TitleToken(b *testing.B) {
	db, queries := testutil.SetupTestDB(b)
	defer db.Close()
	testutil.SeedTestData(b, db)

	svc := NewService(db, queries)
	ctx := context.Background()
	req := SearchRequest{
		Term:  "202520",
		Title: "data",
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := svc.Search(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSearch_AllTimeScope(b *testing.B) {
	db, queries := testutil.SetupTestDB(b)
	defer db.Close()
	testutil.SeedTestData(b, db)

	svc := NewService(db, queries)
	ctx := context.Background()
	req := SearchRequest{
		Subject:      "CSCI",
		CourseNumber: "247",
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := svc.Search(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSearch_CombinedFilters(b *testing.B) {
	db, queries := testutil.SetupTestDB(b)
	defer db.Close()
	testutil.SeedTestData(b, db)

	svc := NewService(db, queries)
	ctx := context.Background()
	req := SearchRequest{
		Term:         "202520",
		Subject:      "CSCI",
		CourseNumber: "2*",
		Title:        "data",
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := svc.Search(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
