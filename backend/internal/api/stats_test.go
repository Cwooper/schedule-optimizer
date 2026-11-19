package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"schedule-optimizer/internal/stats"
	"schedule-optimizer/internal/stats/grades"
	"schedule-optimizer/internal/testutil"

	"github.com/gin-gonic/gin"
)

func TestStatsStatus(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	gradeService := grades.NewService(db, queries, "")
	r := gin.New()
	h := NewHandlers(db, nil, nil, queries, nil, gradeService, stats.NewService(gradeService, queries))
	r.GET("/stats/course/:subject/:number", h.GetStatsCourse)
	r.GET("/stats/section/:term/:crn", h.GetStatsSection)

	status := func(path string) int {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		return w.Code
	}
	if got := status("/stats/course/CSCI/305"); got != http.StatusServiceUnavailable {
		t.Errorf("before load: %d, want 503", got)
	}
	if err := gradeService.LoadFromDB(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/stats/course/CSCI/305", "/stats/section/202340/100"} {
		if got := status(path); got != http.StatusNotFound {
			t.Errorf("%s: %d, want 404", path, got)
		}
	}
}
