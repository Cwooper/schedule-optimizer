package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"schedule-optimizer/internal/testutil"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestValidateCourses(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)
	defer db.Close()

	h := NewHandlers(db, nil, nil, queries, nil)

	tests := []struct {
		name           string
		request        ValidateCoursesRequest
		wantStatus     int
		wantResults    []CourseValidationResult
		wantErrContain string
	}{
		{
			name: "all courses exist",
			request: ValidateCoursesRequest{
				Term: "202520",
				Courses: []struct {
					Subject      string `json:"subject" binding:"required"`
					CourseNumber string `json:"courseNumber" binding:"required"`
				}{
					{Subject: "CSCI", CourseNumber: "247"},
					{Subject: "MATH", CourseNumber: "204"},
				},
			},
			wantStatus: http.StatusOK,
			wantResults: []CourseValidationResult{
				{Subject: "CSCI", CourseNumber: "247", Exists: true, Title: "Data Structures", SectionCount: 1},
				{Subject: "MATH", CourseNumber: "204", Exists: true, Title: "Linear Algebra", SectionCount: 1},
			},
		},
		{
			name: "some courses do not exist",
			request: ValidateCoursesRequest{
				Term: "202520",
				Courses: []struct {
					Subject      string `json:"subject" binding:"required"`
					CourseNumber string `json:"courseNumber" binding:"required"`
				}{
					{Subject: "CSCI", CourseNumber: "247"},
					{Subject: "CSCI", CourseNumber: "999"},
				},
			},
			wantStatus: http.StatusOK,
			wantResults: []CourseValidationResult{
				{Subject: "CSCI", CourseNumber: "247", Exists: true, Title: "Data Structures", SectionCount: 1},
				{Subject: "CSCI", CourseNumber: "999", Exists: false},
			},
		},
		{
			name: "no courses exist in term",
			request: ValidateCoursesRequest{
				Term: "202520",
				Courses: []struct {
					Subject      string `json:"subject" binding:"required"`
					CourseNumber string `json:"courseNumber" binding:"required"`
				}{
					{Subject: "PHYS", CourseNumber: "101"},
				},
			},
			wantStatus: http.StatusOK,
			wantResults: []CourseValidationResult{
				{Subject: "PHYS", CourseNumber: "101", Exists: false},
			},
		},
		{
			name: "empty courses array",
			request: ValidateCoursesRequest{
				Term: "202520",
				Courses: []struct {
					Subject      string `json:"subject" binding:"required"`
					CourseNumber string `json:"courseNumber" binding:"required"`
				}{},
			},
			wantStatus:  http.StatusOK,
			wantResults: []CourseValidationResult{},
		},
		{
			name: "normalizes input to uppercase",
			request: ValidateCoursesRequest{
				Term: "202520",
				Courses: []struct {
					Subject      string `json:"subject" binding:"required"`
					CourseNumber string `json:"courseNumber" binding:"required"`
				}{
					{Subject: "csci", CourseNumber: "247"},
				},
			},
			wantStatus: http.StatusOK,
			wantResults: []CourseValidationResult{
				{Subject: "CSCI", CourseNumber: "247", Exists: true, Title: "Data Structures", SectionCount: 1},
			},
		},
		{
			name: "course exists in different term",
			request: ValidateCoursesRequest{
				Term: "202510",
				Courses: []struct {
					Subject      string `json:"subject" binding:"required"`
					CourseNumber string `json:"courseNumber" binding:"required"`
				}{
					{Subject: "CSCI", CourseNumber: "247"},
					{Subject: "CSCI", CourseNumber: "301"}, // Only in 202520
				},
			},
			wantStatus: http.StatusOK,
			wantResults: []CourseValidationResult{
				{Subject: "CSCI", CourseNumber: "247", Exists: true, Title: "Data Structures", SectionCount: 1},
				{Subject: "CSCI", CourseNumber: "301", Exists: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/courses/validate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			h.ValidateCourses(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantErrContain != "" {
				if !bytes.Contains(w.Body.Bytes(), []byte(tt.wantErrContain)) {
					t.Errorf("body = %s, want to contain %q", w.Body.String(), tt.wantErrContain)
				}
				return
			}

			var resp ValidateCoursesResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if len(resp.Results) != len(tt.wantResults) {
				t.Fatalf("got %d results, want %d", len(resp.Results), len(tt.wantResults))
			}

			for i, want := range tt.wantResults {
				got := resp.Results[i]
				if got.Subject != want.Subject {
					t.Errorf("results[%d].Subject = %q, want %q", i, got.Subject, want.Subject)
				}
				if got.CourseNumber != want.CourseNumber {
					t.Errorf("results[%d].CourseNumber = %q, want %q", i, got.CourseNumber, want.CourseNumber)
				}
				if got.Exists != want.Exists {
					t.Errorf("results[%d].Exists = %v, want %v", i, got.Exists, want.Exists)
				}
				if want.Exists {
					if got.Title != want.Title {
						t.Errorf("results[%d].Title = %q, want %q", i, got.Title, want.Title)
					}
					if got.SectionCount != want.SectionCount {
						t.Errorf("results[%d].SectionCount = %d, want %d", i, got.SectionCount, want.SectionCount)
					}
				}
			}
		})
	}
}

func TestValidateCourses_BatchLimit(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)
	defer db.Close()

	h := NewHandlers(db, nil, nil, queries, nil)

	// Create request with 21 courses (exceeds limit of 20)
	courses := make([]struct {
		Subject      string `json:"subject" binding:"required"`
		CourseNumber string `json:"courseNumber" binding:"required"`
	}, 21)
	for i := range courses {
		courses[i] = struct {
			Subject      string `json:"subject" binding:"required"`
			CourseNumber string `json:"courseNumber" binding:"required"`
		}{Subject: "CSCI", CourseNumber: "100"}
	}

	req := ValidateCoursesRequest{
		Term:    "202520",
		Courses: courses,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/courses/validate", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	h.ValidateCourses(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("Maximum 20 courses")) {
		t.Errorf("body = %s, want to contain 'Maximum 20 courses'", w.Body.String())
	}
}
