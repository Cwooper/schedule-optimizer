package scraper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"schedule-optimizer/internal/testutil"
)

func TestScrapeTerm(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/classSearch/getTerms", func(w http.ResponseWriter, r *http.Request) {
		terms := []TermResponse{
			{Code: "202520", Description: "Spring 2025"},
			{Code: "202510", Description: "Winter 2025"},
		}
		json.NewEncoder(w).Encode(terms)
	})

	mux.HandleFunc("/term/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/searchResults/searchResults", func(w http.ResponseWriter, r *http.Request) {
		offset := r.URL.Query().Get("pageOffset")

		var courses []CourseData
		if offset == "0" {
			courses = []CourseData{
				makeMockCourse("20001", "CSCI", "247", "Data Structures"),
				makeMockCourse("20002", "CSCI", "301", "Algorithms"),
			}
		}

		resp := APIResponse{
			Success:    true,
			TotalCount: 2,
			Data:       courses,
		}
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	db, queries := testutil.SetupTestDB(t)
	defer db.Close()

	scraper, err := newScraperWithBaseURL(queries, 2, server.URL)
	if err != nil {
		t.Fatalf("failed to create scraper: %v", err)
	}

	ctx := context.Background()
	stored, err := scraper.ScrapeTerm(ctx, "202520")
	if err != nil {
		t.Fatalf("ScrapeTerm failed: %v", err)
	}

	if stored != 2 {
		t.Errorf("expected 2 sections stored, got %d", stored)
	}

	sections, err := queries.GetSectionsByTerm(ctx, "202520")
	if err != nil {
		t.Fatalf("failed to get sections: %v", err)
	}
	if len(sections) != 2 {
		t.Errorf("expected 2 sections in DB, got %d", len(sections))
	}

	term, err := queries.GetTermByCode(ctx, "202520")
	if err != nil {
		t.Fatalf("failed to get term: %v", err)
	}
	if term.Description != "Spring 2025" {
		t.Errorf("expected term description 'Spring 2025', got %q", term.Description)
	}
	if !term.LastScrapedAt.Valid {
		t.Error("expected last_scraped_at to be set")
	}
}

func TestScrapeTerm_Pagination(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/classSearch/getTerms", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]TermResponse{{Code: "202520", Description: "Spring 2025"}})
	})

	mux.HandleFunc("/term/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/searchResults/searchResults", func(w http.ResponseWriter, r *http.Request) {
		offset := r.URL.Query().Get("pageOffset")

		var courses []CourseData
		switch offset {
		case "0":
			courses = []CourseData{makeMockCourse("20001", "CSCI", "101", "Intro CS")}
		case "500":
			courses = []CourseData{makeMockCourse("20002", "CSCI", "247", "Data Structures")}
		case "1000":
			courses = []CourseData{makeMockCourse("20003", "CSCI", "301", "Algorithms")}
		}

		resp := APIResponse{
			Success:    true,
			TotalCount: 1200, // forces pagination
			Data:       courses,
		}
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	db, queries := testutil.SetupTestDB(t)
	defer db.Close()

	scraper, err := newScraperWithBaseURL(queries, 2, server.URL)
	if err != nil {
		t.Fatalf("failed to create scraper: %v", err)
	}

	ctx := context.Background()
	stored, err := scraper.ScrapeTerm(ctx, "202520")
	if err != nil {
		t.Fatalf("ScrapeTerm failed: %v", err)
	}

	if stored != 3 {
		t.Errorf("expected 3 sections stored from pagination, got %d", stored)
	}
}

func TestScrapeTerm_PartialFailure(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/classSearch/getTerms", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]TermResponse{{Code: "202520", Description: "Spring 2025"}})
	})

	mux.HandleFunc("/term/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/searchResults/searchResults", func(w http.ResponseWriter, r *http.Request) {
		offset := r.URL.Query().Get("pageOffset")

		// intentionally fail second page to test partial success
		if offset == "500" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var courses []CourseData
		if offset == "0" {
			courses = []CourseData{makeMockCourse("20001", "CSCI", "101", "Intro CS")}
		}

		resp := APIResponse{
			Success:    true,
			TotalCount: 600,
			Data:       courses,
		}
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	db, queries := testutil.SetupTestDB(t)
	defer db.Close()

	scraper, err := newScraperWithBaseURL(queries, 1, server.URL)
	if err != nil {
		t.Fatalf("failed to create scraper: %v", err)
	}

	ctx := context.Background()
	stored, err := scraper.ScrapeTerm(ctx, "202520")
	if err != nil {
		t.Fatalf("expected partial success, got error: %v", err)
	}

	if stored != 1 {
		t.Errorf("expected 1 section stored despite page failure, got %d", stored)
	}
}

func TestScrapeTerms(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/classSearch/getTerms", func(w http.ResponseWriter, r *http.Request) {
		terms := []TermResponse{
			{Code: "202520", Description: "Spring 2025"},
			{Code: "202510", Description: "Winter 2025"},
			{Code: "202440", Description: "Fall 2024"},
		}
		json.NewEncoder(w).Encode(terms)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	db, queries := testutil.SetupTestDB(t)
	defer db.Close()

	scraper, err := newScraperWithBaseURL(queries, 1, server.URL)
	if err != nil {
		t.Fatalf("failed to create scraper: %v", err)
	}

	ctx := context.Background()
	terms, err := scraper.ScrapeTerms(ctx)
	if err != nil {
		t.Fatalf("ScrapeTerms failed: %v", err)
	}

	if len(terms) != 3 {
		t.Errorf("expected 3 terms returned, got %d", len(terms))
	}

	dbTerms, err := queries.GetTerms(ctx)
	if err != nil {
		t.Fatalf("failed to get terms from DB: %v", err)
	}
	if len(dbTerms) != 3 {
		t.Errorf("expected 3 terms in DB, got %d", len(dbTerms))
	}
}

func makeMockCourse(crn, subject, number, title string) CourseData {
	credits := 4.0
	return CourseData{
		Term:                  "202520",
		CourseReferenceNumber: crn,
		Subject:               subject,
		SubjectDescription:    subject + " Department",
		CourseNumber:          number,
		CourseTitle:           title,
		CreditHourLow:         &credits,
		MaximumEnrollment:     30,
		Enrollment:            25,
		SeatsAvailable:        5,
		OpenSection:           true,
		Faculty: []FacultyData{
			{DisplayName: "Dr. Test", EmailAddress: "test@wwu.edu", PrimaryIndicator: true},
		},
		MeetingsFaculty: []MeetingsFaculty{
			{
				MeetingTime: MeetingTimeData{
					BeginTime: "1000",
					EndTime:   "1050",
					Monday:    true,
					Wednesday: true,
					Friday:    true,
					Building:  "CF",
					Room:      "105",
				},
			},
		},
	}
}
