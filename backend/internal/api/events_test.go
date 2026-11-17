package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"schedule-optimizer/internal/testutil"

	"github.com/gin-gonic/gin"
)

const testSession = "550e8400-e29b-41d4-a716-446655440000"

func TestRecordEvent(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	h := NewHandlers(db, nil, nil, queries, nil, nil)
	r := gin.New()
	r.POST("/api/events", h.RecordEvent)

	event := func(name, target, term string) string {
		return `{"session_id":"` + testSession + `","name":"` + name +
			`","target":"` + target + `","term":"` + term + `"}`
	}

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantTarget sql.NullString
		wantTerm   sql.NullString
	}{
		{
			name:       "visit with term",
			body:       event("visit", "desktop", "202520"),
			wantStatus: http.StatusNoContent,
			wantTarget: sql.NullString{String: "desktop", Valid: true},
			wantTerm:   sql.NullString{String: "202520", Valid: true},
		},
		{
			name:       "targetless event stores nulls",
			body:       `{"session_id":"550e8400e29b41d4a716446655440000","name":"finals_view"}`,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "every ics option in order",
			body:       event("ics_export", "classes,finals,holidays,reg1,reg2,reg3,deadlines", "202540"),
			wantStatus: http.StatusNoContent,
			wantTarget: sql.NullString{String: "classes,finals,holidays,reg1,reg2,reg3,deadlines", Valid: true},
			wantTerm:   sql.NullString{String: "202540", Valid: true},
		},
		{
			name:       "single ics option",
			body:       event("ics_export", "reg2", ""),
			wantStatus: http.StatusNoContent,
			wantTarget: sql.NullString{String: "reg2", Valid: true},
		},
		{name: "ics options out of order", body: event("ics_export", "finals,classes", ""), wantStatus: http.StatusBadRequest},
		{name: "ics option repeated", body: event("ics_export", "classes,classes", ""), wantStatus: http.StatusBadRequest},
		{name: "unknown ics option", body: event("ics_export", "classes,grades", ""), wantStatus: http.StatusBadRequest},
		{name: "trailing comma in ics target", body: event("ics_export", "classes,", ""), wantStatus: http.StatusBadRequest},
		{name: "empty ics target", body: event("ics_export", "", ""), wantStatus: http.StatusBadRequest},
		{name: "unknown name", body: event("signup", "", ""), wantStatus: http.StatusBadRequest},
		{name: "target outside name's set", body: event("tab", "finals", ""), wantStatus: http.StatusBadRequest},
		{name: "missing required target", body: event("png_export", "", ""), wantStatus: http.StatusBadRequest},
		{name: "target on targetless event", body: event("feedback_open", "desktop", ""), wantStatus: http.StatusBadRequest},
		{name: "invalid quarter in term", body: event("visit", "mobile", "202550"), wantStatus: http.StatusBadRequest},
		{name: "short term", body: event("visit", "mobile", "2025"), wantStatus: http.StatusBadRequest},
		{
			name:       "invalid session id",
			body:       `{"session_id":"'; DROP TABLE events; --","name":"visit","target":"mobile"}`,
			wantStatus: http.StatusBadRequest,
		},
		{name: "missing session id", body: `{"name":"visit","target":"mobile"}`, wantStatus: http.StatusBadRequest},
		{name: "malformed json", body: `{"name":`, wantStatus: http.StatusBadRequest},
		{name: "trailing data after json", body: event("visit", "mobile", "") + `x`, wantStatus: http.StatusBadRequest},
		{name: "signed term", body: event("visit", "mobile", "-12310"), wantStatus: http.StatusBadRequest},
		{
			name:       "oversize body",
			body:       `{"session_id":"` + testSession + `","name":"visit","target":"mobile","pad":"` + strings.Repeat("x", maxEventBodyBytes) + `"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	stored := 0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/events", strings.NewReader(tt.body)))
			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantStatus != http.StatusNoContent {
				return
			}

			stored++
			waitForEventCount(t, db, stored)
			var session, name string
			var target, term sql.NullString
			if err := db.QueryRow(`SELECT session_id, name, target, term FROM events ORDER BY id DESC LIMIT 1`).
				Scan(&session, &name, &target, &term); err != nil {
				t.Fatal(err)
			}
			var sent EventRequest
			if err := json.Unmarshal([]byte(tt.body), &sent); err != nil {
				t.Fatal(err)
			}
			if session != sent.SessionID || name != sent.Name || target != tt.wantTarget || term != tt.wantTerm {
				t.Errorf("stored (%s, %s, %v, %v); want (%s, %s, %v, %v)",
					session, name, target, term, sent.SessionID, sent.Name, tt.wantTarget, tt.wantTerm)
			}
		})
	}
}

// waitForEventCount waits out RecordEvent's fire-and-forget insert.
func waitForEventCount(t *testing.T, db *sql.DB, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM events`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n == want {
			return
		}
		if n > want || time.Now().After(deadline) {
			t.Fatalf("events = %d, want %d", n, want)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
