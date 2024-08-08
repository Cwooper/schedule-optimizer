package models

import (
	"github.com/cwooper/schedule-optimizer/internal/utils"
)

var (
	dayBegin       = toMins(800)  // Converts the start of the day to mins
	dayEnd         = toMins(1700) // Converts the end of the day to mins
	startSlope     = 1.0 / float64(dayEnd-dayBegin)
	startIntercept = float64(dayBegin) / float64(dayEnd-dayBegin) * -1
	endSlope       = startSlope * -1
	endIntercept   = 1 + (startIntercept * -1)
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

// Creates a new schedule with the courses and initializes the weights
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
	foundWeights := 0
	for i := range s.Weights {
		weighFunc, exists := weightMap[s.Weights[i].Name]
		if exists {
			s.Weights[i].Value = weighFunc(s)
			if s.Weights[i].Value != 0.0 {
				totalWeight += s.Weights[i].Value
				foundWeights += 1
			}
		} // weights already initialized to 0
	}
	s.Score = utils.Round(totalWeight / float64(foundWeights))
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
	daySchedules := initializeDaySchedules()
	updateDaySchedules(s, daySchedules)
	averageGapScore := weighDaySchedules(daySchedules)
	return utils.Round(averageGapScore)
}

// TODO untested
func weighStart(s *Schedule) float64 {
	startTime, _ := findStartEndTimes(s) // Already converted to mins
	if startTime == -1 {
		return 0.0
	}

	if startTime <= dayBegin || startTime >= dayEnd { // Before 08:00 or After 17:00 is error
		return 0.0
	} else {
		return startSlope*float64(startTime) + startIntercept
	}
}

// TODO untested
func weighEnd(s *Schedule) float64 {
	_, endTime := findStartEndTimes(s) // Already converted to mins
	if endTime == -1 {
		return 0.0
	}

	if endTime <= dayBegin || endTime >= dayEnd { // Before 08:00 or After 17:00 is error
		return 0.0
	} else {
		return endSlope*float64(endTime) + endIntercept
	}
}

// Helper function to convert time string to minutes since midnight
func toMins(time int) int {
	return int(time/100)*60 + (time % 100)
}

// Helper function to find the earliest start time and latest end time of the schedule
func findStartEndTimes(s *Schedule) (int, int) {
	endTime := -1
	startTime := -1
	for _, course := range s.Courses {
		for _, session := range course.Sessions {
			if session.IsAsync || session.IsTimeTBD {
				continue // There are no times for async or timetbd sessions
			}
			sessionStart := toMins(session.StartTime)
			sessionEnd := toMins(session.EndTime)
			if startTime == -1 || sessionStart < startTime {
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
	totalCourseTime int // Total time spent in courses
	firstCourseTime int // Start time of the first course
	lastCourseTime  int // End time of the last course
}

func initializeDaySchedules() map[rune]*daySchedule {
	daySchedules := make(map[rune]*daySchedule)
	for _, day := range []rune{'M', 'T', 'W', 'R', 'F'} {
		daySchedules[day] = &daySchedule{
			totalCourseTime: 0,
			firstCourseTime: -1,
			lastCourseTime:  -1,
		}
	}
	return daySchedules
}

// Won't add session to the schedules if it's async or tbd time
func updateDaySchedules(s *Schedule, daySchedules map[rune]*daySchedule) {
	for _, course := range s.Courses {
		for _, session := range course.Sessions {
			if !session.IsAsync && !session.IsTimeTBD {
				updateSessionSchedule(session, daySchedules)
			}
		}
	}
}

func updateSessionSchedule(session Session, daySchedules map[rune]*daySchedule) {
	sessionStartMins := toMins(session.StartTime)
	sessionEndMins := toMins(session.EndTime)
	sessionDuration := sessionEndMins - sessionStartMins

	for _, day := range session.Days {
		schedule := daySchedules[day]
		schedule.totalCourseTime += sessionDuration
		updateDayBoundaries(schedule, sessionStartMins, sessionEndMins)
	}
}

func updateDayBoundaries(schedule *daySchedule, startMins, endMins int) {
	if schedule.firstCourseTime == -1 || startMins < schedule.firstCourseTime {
		schedule.firstCourseTime = startMins
	}
	if schedule.lastCourseTime == -1 || endMins > schedule.lastCourseTime {
		schedule.lastCourseTime = endMins
	}
}

// Weighs the daySchedules based on their total time vs gap time
func weighDaySchedules(daySchedules map[rune]*daySchedule) float64 {
	totalGapScore := 0.0
	activeDays := 0

	for _, schedule := range daySchedules {
		if schedule.totalCourseTime > 0 {
			activeDays++
			totalDayTime := schedule.lastCourseTime - schedule.firstCourseTime
			gapTime := totalDayTime - schedule.totalCourseTime
			gapRatio := float64(gapTime) / float64(totalDayTime)
			totalGapScore += 1 - gapRatio // Higher score for less gap time
		}
	}

	if activeDays == 0 {
		return 0
	}

	return totalGapScore / float64(activeDays)
}
