package generator

import "math"

// Scoring constants for day boundaries.
const (
	dayBeginMins = 8 * 60  // 8:00 AM
	dayEndMins   = 17 * 60 // 5:00 PM
)

// dayStats tracks timing information for a single day.
type dayStats struct {
	totalClassTime int // Total minutes of class time
	firstStart     int // Earliest class start (minutes from midnight)
	lastEnd        int // Latest class end (minutes from midnight)
}

// scoreSchedule computes and sets the Score and Weights for a schedule.
func scoreSchedule(s *Schedule) {
	weights := []Weight{
		{Name: "GPA", Value: weighGPA(s)},
		{Name: "Gap", Value: weighGap(s)},
		{Name: "Start", Value: weighStart(s)},
		{Name: "End", Value: weighEnd(s)},
	}
	s.Weights = weights

	var total float64
	var count int
	for _, w := range weights {
		if w.Value > 0 {
			total += w.Value
			count++
		}
	}
	if count > 0 {
		s.Score = math.Round(total/float64(count)*100) / 100
	}
}

// weighGap scores based on gaps between classes.
// Formula: 1 - (gap_time / total_span) per day, averaged across all active days.
// Higher score = fewer gaps = more compact schedule.
func weighGap(s *Schedule) float64 {
	days := make(map[int]*dayStats)

	for _, c := range s.Courses {
		for _, mt := range c.MeetingTimes {
			start := parseTimeToMins(mt.StartTime)
			end := parseTimeToMins(mt.EndTime)
			if start < 0 || end < 0 {
				continue
			}
			duration := end - start

			for day := range 5 {
				// Days[0]=Sun, Days[1]=Mon, etc.
				if mt.Days[day+1] {
					if days[day] == nil {
						days[day] = &dayStats{firstStart: start, lastEnd: end}
					}
					ds := days[day]
					ds.totalClassTime += duration
					if start < ds.firstStart {
						ds.firstStart = start
					}
					if end > ds.lastEnd {
						ds.lastEnd = end
					}
				}
			}
		}
	}

	var totalScore float64
	var activeDays int
	for _, ds := range days {
		if ds.totalClassTime > 0 {
			activeDays++
			span := ds.lastEnd - ds.firstStart
			if span > 0 {
				gapTime := span - ds.totalClassTime
				gapRatio := float64(gapTime) / float64(span)
				totalScore += 1.0 - gapRatio
			} else {
				totalScore += 1.0
			}
		}
	}

	if activeDays == 0 {
		return 0
	}
	return math.Round(totalScore/float64(activeDays)*100) / 100
}

// weighStart scores based on how late classes start.
// Linear scale: 8am = 0, 5pm = 1.
// Higher score = later start times.
func weighStart(s *Schedule) float64 {
	earliest := max(findEarliestStart(s), dayBeginMins)
	if earliest > dayEndMins {
		return 1.0
	}
	score := float64(earliest-dayBeginMins) / float64(dayEndMins-dayBeginMins)
	return math.Round(score*100) / 100
}

// weighEnd scores based on how early classes end.
// Linear scale: 8am = 1, 5pm = 0.
// Higher score = earlier end times.
func weighEnd(s *Schedule) float64 {
	latest := min(findLatestEnd(s), dayEndMins)
	if latest < dayBeginMins {
		return 1.0
	}
	score := 1.0 - float64(latest-dayBeginMins)/float64(dayEndMins-dayBeginMins)
	return math.Round(score*100) / 100
}

// weighGPA scores based on instructor GPA data.
// Currently returns 0 (excluded from average) until GPA data is integrated.
func weighGPA(_ *Schedule) float64 {
	// TODO: integrate GPA data from CSV import
	return 0
}

// findEarliestStart returns the earliest class start time across all days.
func findEarliestStart(s *Schedule) int {
	earliest := 24 * 60 // Start with end of day
	for _, c := range s.Courses {
		for _, mt := range c.MeetingTimes {
			start := parseTimeToMins(mt.StartTime)
			if start >= 0 && start < earliest {
				earliest = start
			}
		}
	}
	return earliest
}

// findLatestEnd returns the latest class end time across all days.
func findLatestEnd(s *Schedule) int {
	latest := 0
	for _, c := range s.Courses {
		for _, mt := range c.MeetingTimes {
			end := parseTimeToMins(mt.EndTime)
			if end > latest {
				latest = end
			}
		}
	}
	return latest
}
