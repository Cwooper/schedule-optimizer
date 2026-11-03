package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"schedule-optimizer/internal/testutil"

	"github.com/gin-gonic/gin"
)

func getJSON(t *testing.T, h gin.HandlerFunc, route, url string) (int, map[string]any) {
	t.Helper()
	r := gin.New()
	r.GET(route, h)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, url, nil))
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode %s: %v (%s)", url, err, w.Body.String())
	}
	return w.Code, body
}

func TestGetCalendar(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)
	h := NewHandlers(db, nil, nil, queries, nil, nil)
	testutil.SeedSpring2025Calendar(t, db)

	code, body := getJSON(t, h.GetCalendar, "/calendar/:term", "/calendar/202520")
	if code != http.StatusOK {
		t.Fatalf("status %d: %v", code, body)
	}
	want := map[string]any{
		"term":              "202520",
		"start":             "2025-03-31",
		"end":               "2025-06-13",
		"registrationOpens": "2025-02-25",
		"estimated":         false,
		"registration":      []any{map[string]any{"phase": "I", "name": "Phase I - registration", "start": "2025-02-25", "end": "2025-03-10"}},
		"holidays":          []any{map[string]any{"name": "Memorial Day", "start": "2025-05-26", "end": "2025-05-26"}},
		"deadlines": []any{map[string]any{
			"name": "Last day to withdraw from a class", "title": "Withdraw deadline", "date": "2025-05-11", "kind": "withdraw",
		}},
		"finals": map[string]any{"start": "2025-06-09", "end": "2025-06-13"},
	}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("GET /calendar/202520 =\n%v\nwant\n%v", body, want)
	}

	if code, _ := getJSON(t, h.GetCalendar, "/calendar/:term", "/calendar/209910"); code != http.StatusNotFound {
		t.Errorf("unknown term: status %d, want 404", code)
	}
}

func TestGetCRNIncludesFinal(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)
	h := NewHandlers(db, nil, nil, queries, nil, nil)
	testutil.SeedSpring2025Calendar(t, db)

	_, body := getJSON(t, h.GetCRN, "/crn/:crn", "/crn/20001?term=202520")
	section, _ := body["section"].(map[string]any)
	want := map[string]any{"date": "2025-06-10", "start": "1030", "end": "1230", "certain": true, "location": "CF 105"}
	if !reflect.DeepEqual(section["final"], want) {
		t.Errorf("final = %v, want %v", section["final"], want)
	}

	// Winter 2025 has no finals table, so its sections carry no final.
	_, body = getJSON(t, h.GetCRN, "/crn/:crn", "/crn/10001?term=202510")
	section, _ = body["section"].(map[string]any)
	if _, has := section["final"]; has || section == nil {
		t.Errorf("section without a finals table = %v, want no final", section)
	}
}

// A finals table that fails to load costs the final, not the lookup.
func TestGetCRNWithoutLoadableFinals(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)
	h := NewHandlers(db, nil, nil, queries, nil, nil)
	if _, err := db.Exec(`INSERT INTO finals_mappings (term_code, time_range_start, time_range_end, has_tuth, exam_date, exam_start_time, exam_end_time)
		VALUES ('202520', '0930', '1029', 0, 'June 10', '1030', '1230')`); err != nil {
		t.Fatal(err)
	}

	code, body := getJSON(t, h.GetCRN, "/crn/:crn", "/crn/20001?term=202520")
	section, _ := body["section"].(map[string]any)
	if _, has := section["final"]; code != http.StatusOK || section == nil || has {
		t.Errorf("status %d, section %v; want the section without a final", code, section)
	}
}

func TestGetTermsIncludesDates(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)
	h := NewHandlers(db, nil, nil, queries, nil, nil)
	testutil.SeedSpring2025Calendar(t, db)

	_, body := getJSON(t, h.GetTerms, "/terms", "/terms")
	terms, _ := body["terms"].([]any)
	if len(terms) != 2 {
		t.Fatalf("terms = %v", body["terms"])
	}
	want := map[string]any{
		"code": "202520", "name": "Spring 2025", "phase": "past",
		"start": "2025-03-31", "end": "2025-06-13", "registrationOpens": "2025-02-25", "estimated": false,
	}
	if !reflect.DeepEqual(terms[0], want) {
		t.Errorf("terms[0] = %v, want %v", terms[0], want)
	}
	if winter := terms[1].(map[string]any); winter["estimated"] != true {
		t.Errorf("Winter 2025 has no scraped dates, want estimated: %v", winter)
	}
}
