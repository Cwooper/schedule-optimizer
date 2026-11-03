// Package meeting holds the meeting-time shape every section payload shares,
// so the payloads cannot drift apart.
package meeting

import (
	"database/sql"
	"strings"
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
	Lab       bool    `json:"lab,omitempty"`       // ScheduleType "B"
	// ScheduleType is Banner's code for the meeting ("L", "B" for lab, ...).
	// It only feeds final exam matching.
	ScheduleType string `json:"-"`
}

const (
	// LectureType is Banner's meeting type for a lecture.
	LectureType = "L"
	labType     = "B"
)

// Location is the building and room, without Banner's ASNC and SYNC
// placeholders, which mark online meetings rather than rooms.
func (t Time) Location() string {
	parts := make([]string, 0, 2)
	for _, p := range []string{t.Building, t.Room} {
		if p != "" && p != "ASNC" && p != "SYNC" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, " ")
}

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
		Lab:          m.ScheduleType.String == labType,
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
