package generator

import (
	"math"
	"slices"
)

const (
	weekdays = 5
	// Idle stretches this short are walking between classes, not a gap.
	passingTimeMins = 15
	gapWidthMins    = 60
)

// startEndWidths maps a Start or End level to the falloff width in minutes,
// so a stricter level also punishes the same miss harder.
var startEndWidths = [MaxImportance + 1]float64{0, 180, 120, 60}

// scorer ranks schedules by the user's preferences. Each factor scores in
// [0, 1]; the schedule's score is their weighted mean.
type scorer struct {
	start, end, gap, gpa, seats float64 // weights; 0 = off
	startWidth, endWidth        float64
	idealStart, idealEnd        int // minutes from midnight
}

// newScorer expects validated preferences (see Preferences.UnmarshalJSON).
func newScorer(p Preferences) scorer {
	return scorer{
		start:      float64(p.Start),
		end:        float64(p.End),
		gap:        float64(p.Gap),
		gpa:        float64(p.GPA),
		seats:      float64(p.Seats),
		startWidth: startEndWidths[p.Start],
		endWidth:   startEndWidths[p.End],
		idealStart: parseTimeToMins(p.IdealStart),
		idealEnd:   parseTimeToMins(p.IdealEnd),
	}
}

// score returns the weighted mean of the enabled factors that have data, or 0
// when none do.
func (sc scorer) score(s *Schedule) float64 {
	var sum, weights float64
	add := func(w, score float64, hasData bool) {
		if w > 0 && hasData {
			sum += w * score
			weights += w
		}
	}

	if sc.start > 0 || sc.end > 0 || sc.gap > 0 {
		week, busy := weekStats(s)
		add(sc.start, sc.startScore(week), busy)
		add(sc.end, sc.endScore(week), busy)
		add(sc.gap, gapScore(week), busy)
	}
	if sc.gpa > 0 {
		score, ok := gpaScore(s)
		add(sc.gpa, score, ok)
	}
	if sc.seats > 0 {
		add(sc.seats, seatsScore(s), len(s.Courses) > 0)
	}

	if weights == 0 {
		return 0
	}
	return sum / weights
}

// falloff maps a miss of d minutes to a score in [0, 1]: flat near zero, so
// small misses barely count, then dropping steeply around width k.
func falloff(d, k float64) float64 {
	if d <= 0 {
		return 1
	}
	x := d / k
	return 1 - math.Tanh(x*x)
}

// dayStats summarizes one weekday's classes, in minutes from midnight.
type dayStats struct {
	busy        bool
	first, last int
	idle        int // time between classes, excluding passing time
}

// weekStats builds Mon-Fri stats in one pass over the schedule's timed
// meetings, reporting whether any weekday has a class.
func weekStats(s *Schedule) ([weekdays]dayStats, bool) {
	// Each meeting-day packs into one key ordered by (day, start), which sorts
	// faster than a struct with a comparison func; minutes fit in 11 bits.
	var buf [32]uint32
	keys := buf[:0]
	for _, c := range s.Courses {
		for _, mt := range c.MeetingTimes {
			start, end := parseTimeToMins(mt.StartTime), parseTimeToMins(mt.EndTime)
			if start < 0 || end < 0 {
				continue
			}
			for day := range weekdays {
				if mt.Days[day+1] { // Days[0] is Sunday
					keys = append(keys, uint32(day)<<22|uint32(start)<<11|uint32(end))
				}
			}
		}
	}
	slices.Sort(keys)

	var week [weekdays]dayStats
	for _, k := range keys {
		ds := &week[k>>22]
		start, end := int(k>>11&0x7ff), int(k&0x7ff)
		if !ds.busy {
			*ds = dayStats{busy: true, first: start, last: end}
			continue
		}
		// Sorted by start, so ds.last is the end of the merged block so far.
		if gap := start - ds.last; gap > passingTimeMins {
			ds.idle += gap
		}
		ds.last = max(ds.last, end)
	}
	return week, len(keys) > 0
}

// startScore averages over all five weekdays, not just class days, so adding a
// good day can't mask a bad one. Starting later than ideal costs nothing.
func (sc scorer) startScore(week [weekdays]dayStats) float64 {
	var total float64
	for _, ds := range week {
		if !ds.busy {
			total++
			continue
		}
		total += falloff(float64(sc.idealStart-ds.first), sc.startWidth)
	}
	return total / weekdays
}

// endScore mirrors startScore: finishing earlier than ideal costs nothing.
func (sc scorer) endScore(week [weekdays]dayStats) float64 {
	var total float64
	for _, ds := range week {
		if !ds.busy {
			total++
			continue
		}
		total += falloff(float64(ds.last-sc.idealEnd), sc.endWidth)
	}
	return total / weekdays
}

func gapScore(week [weekdays]dayStats) float64 {
	var idle int
	for _, ds := range week {
		idle += ds.idle
	}
	return falloff(float64(idle)/weekdays, gapWidthMins)
}

// gpaScore maps a 2.0 GPA to 0 and 4.0 to 1. A course without grade data
// counts as 3.0, so missing data neither helps nor hurts a schedule.
func gpaScore(s *Schedule) (float64, bool) {
	var total float64
	var hasData bool
	for _, c := range s.Courses {
		if c.GPA <= 0 {
			total += 0.5
			continue
		}
		hasData = true
		total += min(max((c.GPA-2)/2, 0), 1)
	}
	if !hasData {
		return 0, false
	}
	return total / float64(len(s.Courses)), true
}

func seatsScore(s *Schedule) float64 {
	var open int
	for _, c := range s.Courses {
		if c.IsOpen {
			open++
		}
	}
	return float64(open) / float64(len(s.Courses))
}
