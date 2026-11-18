package api

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"schedule-optimizer/internal/cache"
	"schedule-optimizer/internal/generator"
	"schedule-optimizer/internal/testutil"

	"github.com/gin-gonic/gin"
)

func TestGeneratePreferences(t *testing.T) {
	const valid = `{"start":3,"end":0,"gap":1,"gpa":2,"seats":0,"idealStart":"0930","idealEnd":"1500"}`
	tests := []struct {
		name        string
		preferences string // JSON value, or "" to omit the field
		wantStatus  int
		wantLogged  sql.NullString
	}{
		{name: "omitted uses defaults", wantStatus: http.StatusOK},
		{name: "explicit null uses defaults", preferences: `null`, wantStatus: http.StatusOK},
		{
			name:        "valid, with levels at 0",
			preferences: valid,
			wantStatus:  http.StatusOK,
			wantLogged:  sql.NullString{String: valid, Valid: true},
		},
		{
			name:        "colon time is accepted",
			preferences: strings.Replace(valid, `"0930"`, `"09:30"`, 1),
			wantStatus:  http.StatusOK,
			wantLogged:  sql.NullString{String: strings.Replace(valid, `"0930"`, `"09:30"`, 1), Valid: true},
		},
		{name: "omitted level", preferences: `{"start":3,"end":0,"gap":1,"gpa":2,"idealStart":"0930","idealEnd":"1500"}`, wantStatus: http.StatusBadRequest},
		{name: "omitted time", preferences: `{"start":3,"end":0,"gap":1,"gpa":2,"seats":0,"idealStart":"0930"}`, wantStatus: http.StatusBadRequest},
		{name: "empty object", preferences: `{}`, wantStatus: http.StatusBadRequest},
		{name: "level above range", preferences: strings.Replace(valid, `"start":3`, `"start":4`, 1), wantStatus: http.StatusBadRequest},
		{name: "negative level", preferences: strings.Replace(valid, `"gap":1`, `"gap":-1`, 1), wantStatus: http.StatusBadRequest},
		{name: "fractional level", preferences: strings.Replace(valid, `"gap":1`, `"gap":1.5`, 1), wantStatus: http.StatusBadRequest},
		{name: "hour out of range", preferences: strings.Replace(valid, `"1500"`, `"2500"`, 1), wantStatus: http.StatusBadRequest},
		{name: "garbage time", preferences: strings.Replace(valid, `"0930"`, `"soon"`, 1), wantStatus: http.StatusBadRequest},
		{name: "empty time", preferences: strings.Replace(valid, `"0930"`, `""`, 1), wantStatus: http.StatusBadRequest},
		{name: "level as string", preferences: strings.Replace(valid, `"gpa":2`, `"gpa":"2"`, 1), wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, queries := testutil.SetupTestDB(t)
			testutil.SeedTestData(t, db)
			h := NewHandlers(db, cache.NewScheduleCache(queries, nil), generator.NewService(queries), queries, nil, nil)
			r := gin.New()
			r.POST("/api/generate", h.Generate)

			body := `{"term":"202520","courseSpecs":[{"subject":"CSCI","courseNumber":"247"}]`
			if tt.preferences != "" {
				body += `,"preferences":` + tt.preferences
			}
			body += `}`

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(body)))
			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", w.Code, tt.wantStatus, w.Body)
			}
			if tt.wantStatus != http.StatusOK {
				return
			}

			// The handler logs in the background after responding.
			var logged sql.NullString
			deadline := time.Now().Add(2 * time.Second)
			for {
				err := db.QueryRow(`SELECT preferences FROM generation_logs`).Scan(&logged)
				if err == nil {
					break
				}
				if err != sql.ErrNoRows || time.Now().After(deadline) {
					t.Fatalf("reading generation log: %v", err)
				}
				time.Sleep(10 * time.Millisecond)
			}
			if logged != tt.wantLogged {
				t.Errorf("logged preferences = %+v, want %+v", logged, tt.wantLogged)
			}
		})
	}
}
