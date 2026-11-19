package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"schedule-optimizer/internal/api"
	"schedule-optimizer/internal/stats"
	"schedule-optimizer/internal/stats/grades"
	"schedule-optimizer/internal/testutil"

	"github.com/gin-gonic/gin"
)

func TestRoutesKeepSlashSubjects(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	_, err := db.Exec(`
		INSERT INTO terms (code, description) VALUES ('202440', 'Fall 2024');
		INSERT INTO sections (id, term, crn, subject, course_number, title) VALUES
			(1, '202440', '100', 'A/HI', '201', 'Art History'),
			(2, '202440', '200', 'CSCI', '241', 'Data Structures');
		INSERT INTO grade_rows (term, crn, subject, course_number, title, professor, students_enrolled, grade_count, cnt_a) VALUES
			('202440', '100', 'A/HI', '201', 'Art History', '', 20, 20, 20),
			('202440', '200', 'CSCI', '241', 'Data Structures', '', 20, 20, 20);
	`)
	if err != nil {
		t.Fatal(err)
	}
	gradeService := grades.NewService(db, queries, "")
	if err := gradeService.LoadFromDB(context.Background()); err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	RegisterRoutes(r, api.NewHandlers(db, nil, nil, queries, nil, gradeService, stats.NewService(gradeService, queries)))

	tests := []struct{ path, subject string }{
		{"/api/stats/course/A%2FHI/201", "A/HI"},
		{"/api/stats/course/CSCI/241", "CSCI"},
		{"/api/course/A%2FHI/201?term=202440", "A/HI"},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if w.Code != http.StatusOK {
			t.Errorf("%s: %d, want 200", tt.path, w.Code)
			continue
		}
		// The stats view has the subject at the top, the course lookup under course.
		var body struct {
			Subject string
			Course  struct{ Subject string }
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if got := body.Subject + body.Course.Subject; got != tt.subject {
			t.Errorf("%s: subject %q, want %q", tt.path, got, tt.subject)
		}
	}
}
