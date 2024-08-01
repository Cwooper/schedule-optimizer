// Package generator for generating schedules in responses
package generator

import (
	"fmt"
	"regexp"
	"strings"

	"schedule-optimizer/internal/models"
	"schedule-optimizer/internal/utils"
)

var CourseNamePattern = regexp.MustCompile(`^[A-Z/ ]{2,4} \d{3}[A-Z]?$`) // e.g. CSCI 497A

// Conflict structure to hold the conflicts in an array of courses
type Conflict struct {
	First  models.Course
	Second models.Course
}

// The generator
type Generator struct {
	response           *models.Response // Holds the response to the user
	conflicts          []Conflict       // List of conflict pairs
	courses            []models.Course  // list of Course objects
	cleanedCourseNames []string         // Courses as names
	cleanedForcedNames []string         // Courses to force as names
}

func NewGenerator() *Generator {
	return &Generator{
		response:           &models.Response{},       // Response to the frontend
		conflicts:          make([]Conflict, 0),      // Course conflicts
		courses:            make([]models.Course, 0), // Cleaned courses
		cleanedCourseNames: make([]string, 0),        // Cleaned list of course names
		cleanedForcedNames: make([]string, 0),        // Cleaned list of forced names
	}
}

// Generates schedules based upon the desired list of courses
// Takes into account the requested min and max schedule size
// Forces the forced courses into the returned schedule
// This takes can takes a "dirty request" directly from a user
func (g *Generator) GenerateResponse(req models.ScheduleRequest) *models.Response {
	// Clean the course Names
	g.cleanedCourseNames = g.cleanCourseNames(req.Courses)
	if len(g.response.Errors) > 0 {
		return g.response
	}
	g.cleanedForcedNames = g.cleanCourseNames(req.Forced)
	if len(g.response.Errors) > 0 {
		return g.response
	}

	// Generate the courses based on the cleaned Names
	g.generateCourses()
	if len(g.response.Errors) > 0 {
		return g.response
	}

	// Clamp and verify that the bounds work
	min, max := clampBounds(req.Min, req.Max)
	if len(g.courses) < min {
		errString := fmt.Sprintf("Cannot generate %d minimum size schedule "+
			"with %d courses", req.Min, len(g.courses))
		g.response.Errors = append(g.response.Errors, errString)
		return g.response
	}
	if len(req.Forced) > req.Max {
		errString := fmt.Sprintf("Cannot force more than %d courses", max)
		g.response.Errors = append(g.response.Errors, errString)
		return g.response
	}

	// Generates the conflicts and the schedules
	g.generateConflicts()
	g.generateSchedules(req)

	return g.response
}

// Cleans the course names by stripping any possible whitespace and verifying
// that the course names are valid
func (g *Generator) cleanCourseNames(courses []string) []string {
	cleanedNames := make([]string, len(courses))
	for _, courseName := range courses {
		courseName = strings.TrimSpace(courseName)
		if CourseNamePattern.MatchString(courseName) {
			cleanedNames = append(cleanedNames, courseName)
		} else { // Invalid Course Name
			g.response.Warnings = append(g.response.Warnings, "Invalid course name: "+courseName)
		}
	}

	return cleanedNames
}

// Fills the courses array in the generator with courses in the database
func (g *Generator) generateCourses() {
	// Gets the courses from the given data/term proto buf
}

// Fills the conflicts array in the generator with known conflict pairs
// Must be called after filling the generators courses with objects
func (g *Generator) generateConflicts() {
	for i := 0; i < len(g.courses)-1; i++ {
		for j := i + 1; j < len(g.courses); j++ {
			if g.courses[i].Conflicts(g.courses[j]) {
				g.conflicts = append(g.conflicts, Conflict{
					First:  g.courses[i],
					Second: g.courses[j],
				})
			}
		}
	}
}

// Generates Schedules after finding conflicts and getting courses
func (g *Generator) generateSchedules(req models.ScheduleRequest) {

}

// Clamp the bounds in case they are out of bounds
func clampBounds(min, max int) (int, int) {
	if min < utils.MIN_COURSES {
		min = utils.MIN_COURSES
	} else if min > utils.MAX_COURSES {
		min = utils.MAX_COURSES
	}

	if max > utils.MAX_COURSES {
		max = utils.MAX_COURSES
	} else if max < min {
		max = min
	}

	return min, max
}
