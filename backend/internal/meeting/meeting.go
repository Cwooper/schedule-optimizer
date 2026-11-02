// Package meeting holds the meeting-time shape every section payload shares,
// so the schedule cache, search, and CRN lookup cannot drift apart.
package meeting

import (
	"database/sql"
	"time"

	"schedule-optimizer/internal/store"
)

// Time is one meeting pattern of a section.
type Time struct {
	Days      [7]bool `json:"days"` // [Sun, Mon, Tue, Wed, Thu, Fri, Sat]
	StartTime string  `json:"startTime,omitempty"`
	EndTime   string  `json:"endTime,omitempty"`
	Building  string  `json:"building,omitempty"`
	Room      string  `json:"room,omitempty"`
	StartDate string  `json:"startDate,omitempty"` // YYYY-MM-DD
	EndDate   string  `json:"endDate,omitempty"`   // YYYY-MM-DD, inclusive
	// ScheduleType is Banner's code for the meeting ("L", "B" for lab, ...).
	// It only feeds final exam matching.
	ScheduleType string `json:"-"`
}

// IsLab reports whether Banner marks the meeting as a lab.
func (t Time) IsLab() bool { return t.ScheduleType == "B" }

// FromStore converts a stored meeting row.
func FromStore(m *store.MeetingTime) Time {
	return Time{
		Days: [7]bool{
			flag(m.Sunday), flag(m.Monday), flag(m.Tuesday), flag(m.Wednesday),
			flag(m.Thursday), flag(m.Friday), flag(m.Saturday),
		},
		StartTime:    m.StartTime.String,
		EndTime:      m.EndTime.String,
		Building:     m.Building.String,
		Room:         m.Room.String,
		StartDate:    isoDate(m.StartDate),
		EndDate:      isoDate(m.EndDate),
		ScheduleType: m.ScheduleType.String,
	}
}

// FromStoreAll converts rows, returning an empty (not nil) slice so the JSON
// is [] rather than null.
func FromStoreAll(rows []*store.MeetingTime) []Time {
	out := make([]Time, 0, len(rows))
	for _, m := range rows {
		out = append(out, FromStore(m))
	}
	return out
}

func flag(n sql.NullInt64) bool { return n.Valid && n.Int64 != 0 }

// isoDate converts Banner's MM/DD/YYYY to YYYY-MM-DD, or "" if unreadable.
func isoDate(s sql.NullString) string {
	t, err := time.Parse("01/02/2006", s.String)
	if err != nil {
		return ""
	}
	return t.Format(time.DateOnly)
}
