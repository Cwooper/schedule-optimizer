package generator

import (
	"github.com/cwooper/schedule-optimizer/internal/models"
)

type ScheduleRequest struct {
	Courses   []models.Course
	Conflicts []Conflict
	Min       int
	Max       int
}

// Returns all possible schedules given the scheduleRequest
func Combinations(scheduleRequest ScheduleRequest) []models.Schedule {
	courses := scheduleRequest.Courses
	conflicts := createConflictMap(scheduleRequest.Conflicts)
	min := scheduleRequest.Min
	max := scheduleRequest.Max

	if max > len(courses) {
		max = len(courses)
	}

	result := [][]models.Course{}
	var combine func(int, []models.Course)

	combine = func(start int, current []models.Course) {
		if len(current) >= min {
			temp := make([]models.Course, len(current))
			copy(temp, current)
			result = append(result, temp)
		}

		if len(current) == max {
			return
		}

		for i := start; i < len(courses); i++ {
			if !hasConflict(current, courses[i], conflicts) {
				current = append(current, courses[i])
				combine(i+1, current)
				current = current[:len(current)-1]
			}
		}
	}

	combine(0, []models.Course{})

	return courseListsToSchedules(result)
}

// Creates a two-way map of conflicts for an asympotically efficient runtime
func createConflictMap(conflicts []Conflict) map[int]map[int]bool {
	conflictMap := make(map[int]map[int]bool)
	for _, conflict := range conflicts {
		if _, exists := conflictMap[conflict.First]; !exists {
			conflictMap[conflict.First] = make(map[int]bool)
		}
		if _, exists := conflictMap[conflict.Second]; !exists {
			conflictMap[conflict.Second] = make(map[int]bool)
		}
		conflictMap[conflict.First][conflict.Second] = true
		conflictMap[conflict.Second][conflict.First] = true
	}
	return conflictMap
}

// Returns true if any of the current courses in current conflict with the course
func hasConflict(current []models.Course, course models.Course, conflicts map[int]map[int]bool) bool {
	for _, c := range current {
		if conflicts[c.CRN][course.CRN] {
			return true
		}
	}
	return false
}

// Turns an array of course lists into an array of schedules
func courseListsToSchedules(courseLists [][]models.Course) []models.Schedule {
	schedules := make([]models.Schedule, len(courseLists))
	for i, courses := range courseLists {
		schedules[i] = *models.NewSchedule(courses)
	}
	return schedules
}
