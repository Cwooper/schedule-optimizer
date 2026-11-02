package meeting

import (
	"database/sql"
	"testing"

	"schedule-optimizer/internal/store"
)

func TestFromStore(t *testing.T) {
	str := func(s string) sql.NullString { return sql.NullString{String: s, Valid: s != ""} }
	got := FromStore(&store.MeetingTime{
		StartTime: str("0900"), EndTime: str("0950"),
		StartDate: str("09/23/2026"), EndDate: str("TBA"),
		Tuesday: sql.NullInt64{Int64: 1, Valid: true}, Thursday: sql.NullInt64{Int64: 0, Valid: true},
		ScheduleType: str("B"),
	})
	want := Time{
		Days:      [7]bool{false, false, true, false, false, false, false},
		StartTime: "0900", EndTime: "0950",
		StartDate: "2026-09-23", EndDate: "",
		ScheduleType: "B",
	}
	if got != want {
		t.Errorf("FromStore = %+v, want %+v", got, want)
	}
}
