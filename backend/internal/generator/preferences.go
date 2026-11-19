package generator

import (
	"encoding/json"
	"errors"
	"fmt"
)

// MaxImportance is the highest importance level a preference can take; 0 turns
// its factor off.
const MaxImportance = 3

// Preferences sets how much each scoring factor counts when ranking schedules.
type Preferences struct {
	Start      int    `json:"start"`
	End        int    `json:"end"`
	Gap        int    `json:"gap"`
	GPA        int    `json:"gpa"`
	Seats      int    `json:"seats"`
	IdealStart string `json:"idealStart"` // "HHMM", as in BlockedTime
	IdealEnd   string `json:"idealEnd"`
}

// DefaultPreferences applies when a request omits preferences. The frontend
// store mirrors these values.
func DefaultPreferences() Preferences {
	return Preferences{
		Start:      2,
		End:        2,
		Gap:        2,
		GPA:        3,
		Seats:      0,
		IdealStart: "1000",
		IdealEnd:   "1400",
	}
}

// UnmarshalJSON requires every field, so an omitted level is an error rather
// than a silent 0 (off), and rejects out-of-range levels and unparseable times.
func (p *Preferences) UnmarshalJSON(data []byte) error {
	var raw struct {
		Start      *int    `json:"start"`
		End        *int    `json:"end"`
		Gap        *int    `json:"gap"`
		GPA        *int    `json:"gpa"`
		Seats      *int    `json:"seats"`
		IdealStart *string `json:"idealStart"`
		IdealEnd   *string `json:"idealEnd"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var errs []error
	level := func(name string, v *int) int {
		switch {
		case v == nil:
			errs = append(errs, fmt.Errorf("preferences.%s is required", name))
		case *v < 0 || *v > MaxImportance:
			errs = append(errs, fmt.Errorf("preferences.%s must be 0-%d, got %d", name, MaxImportance, *v))
		default:
			return *v
		}
		return 0
	}
	timeOfDay := func(name string, v *string) string {
		switch {
		case v == nil:
			errs = append(errs, fmt.Errorf("preferences.%s is required", name))
		case parseTimeToMins(*v) < 0:
			errs = append(errs, fmt.Errorf("preferences.%s must be a time of day as HHMM", name))
		default:
			return *v
		}
		return ""
	}

	parsed := Preferences{
		Start:      level("start", raw.Start),
		End:        level("end", raw.End),
		Gap:        level("gap", raw.Gap),
		GPA:        level("gpa", raw.GPA),
		Seats:      level("seats", raw.Seats),
		IdealStart: timeOfDay("idealStart", raw.IdealStart),
		IdealEnd:   timeOfDay("idealEnd", raw.IdealEnd),
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}
	*p = parsed
	return nil
}
