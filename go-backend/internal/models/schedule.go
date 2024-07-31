package models

import (
	"math"

	"schedule-optimizer/internal/utils"
)

// WeighFunc is a function type for weight calculation functions
type WeighFunc func(*Schedule) float64

// weightMap maps weight names to their corresponding weighing functions
var weightMap = map[string]WeighFunc{
	"GPA":   weighGPA,
	"GAP":   weighGap,
	"Start": weighStart,
	"End":   weighEnd,
}

// A weight for the schedule
type Weight struct {
	Name  string
	Value float64
}

// Holds a single displayable schedule object
type Schedule struct {
	Courses []Course // All courses in this schedule
	Weights []Weight // All current weights for this schedule
	Score   float64  // The average of the weights for an average score
}

func NewSchedule(courses []Course) *Schedule {
	s := &Schedule{
		Courses: courses,
		Weights: make([]Weight, 0, len(weightMap)),
		Score:   0,
	}
	s.initializeWeights()
	s.weigh()
	return s
}

// initializeWeights initializes the Weights slice with all weights from weightMap
func (s *Schedule) initializeWeights() {
	for name := range weightMap {
		s.Weights = append(s.Weights, Weight{Name: name, Value: 0})
	}
}

// Weighs this schedule by evaluating and storing the internal weights
func (s *Schedule) weigh() {
	var totalWeight float64
	for i := range s.Weights {
		weighFunc, exists := weightMap[s.Weights[i].Name]
		if exists {
			s.Weights[i].Value = weighFunc(s)
			totalWeight += s.Weights[i].Value
		} // weights already initialized to 0
	}
	s.Score = utils.Round(totalWeight / float64(len(s.Weights)))
}

// ----------------------- Weighing Logic -----------------------

// TODO untested
func weighGPA(s *Schedule) float64 {
	totalGPA := 0.0
	numCourses := 0
	for _, course := range s.Courses {
		if course.GPA > 0 {
			totalGPA += float64(course.GPA)
			numCourses++
		}
	}

	if numCourses == 0 {
		return 0
	}

	averageGPA := totalGPA / float64(numCourses)
	gpaScore := averageGPA / 4.0 // Normalize weight based on 4.0 scale
	return utils.Round(gpaScore)
}

// TODO untested
func weighGap(s *Schedule) float64 {
	daySchedules := make(map[rune]*daySchedule)
	for _, day := range []rune{'M', 'T', 'W', 'R', 'F'} {
		daySchedules[day] = &daySchedule{
			totalClassTime: 0,
			firstClassTime: -1,
			lastClassTime:  -1,
		}
	}

	for _, course := range s.Courses {
		for _, session := range course.Sessions {
			sessionStartMins := toMins(session.StartTime)
			sessionEndMins := toMins(session.EndTime)
			sessionDuration := sessionEndMins - sessionStartMins

			for _, day := range session.Days {
				// Increase the total amount of time taken on this day
				schedule := daySchedules[day]
				schedule.totalClassTime += sessionDuration

				// Change min and max values if necessary
				if schedule.firstClassTime == -1 || sessionStartMins < schedule.firstClassTime {
					schedule.firstClassTime = sessionStartMins
				}
				if schedule.lastClassTime == -1 || sessionEndMins > schedule.lastClassTime {
					schedule.lastClassTime = sessionEndMins
				}
			}
		}
	}

	totalGapScore := 0.0
	activeDays := 0

	for _, schedule := range daySchedules {
		if schedule.totalClassTime > 0 {
			activeDays++
			totalDayTime := schedule.lastClassTime - schedule.firstClassTime
			gapTime := totalDayTime - schedule.totalClassTime
			gapRatio := float64(gapTime) / float64(totalDayTime)
			totalGapScore += 1 - gapRatio // Higher score for less gap time
		}
	}

	if activeDays == 0 {
		return 0
	}

	averageGapScore := totalGapScore / float64(activeDays)
	return utils.Round(averageGapScore)
}

// TODO untested
func weighStart(s *Schedule) float64 {
	startTime, _ := findStartEndTimes(s)

	dayBegin := toMins(800)
	dayEnd := toMins(1700)

	if startTime < dayBegin || startTime > dayEnd { // Before 08:00 or After 17:00 is error
		return 0.0
	} else {
		ratio := 1 / (math.Log10(float64(dayEnd - dayBegin)))
		return ratio * math.Log10(float64(startTime-dayBegin))
	}
}

// TODO untested
func weighEnd(s *Schedule) float64 {
	_, endTime := findStartEndTimes(s)

	dayBegin := toMins(800)
	dayEnd := toMins(1700)

	if endTime < dayBegin || endTime > dayEnd { // Before 08:00 or After 17:00 is error
		return 0.0
	} else {
		ratio := 1 / (math.Log10(float64(dayEnd - dayBegin)))
		return ratio * math.Log10(float64(dayEnd-endTime))
	}
}

// Helper function to convert time string to minutes since midnight
func toMins(time int) int {
	return int(time/100)*60 + (time % 100)
}

// Helper function to find the earliest start time and latest end time of the schedule
func findStartEndTimes(s *Schedule) (int, int) {
	var startTime, endTime int
	for _, course := range s.Courses {
		for _, session := range course.Sessions {
			sessionStart := toMins(session.StartTime)
			sessionEnd := toMins(session.EndTime)
			if startTime == 0 || sessionStart < startTime {
				startTime = sessionStart
			}
			if sessionEnd > endTime {
				endTime = sessionEnd
			}
		}
	}
	return startTime, endTime
}

// daySchedule holds the schedule information for a single day
type daySchedule struct {
	totalClassTime int // Total time spent in classes
	firstClassTime int // Start time of the first class
	lastClassTime  int // End time of the last class
}
