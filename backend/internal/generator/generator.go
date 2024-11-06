// Package generator for generating schedules in responses
package generator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/cwooper/schedule-optimizer/internal/models"
	"github.com/cwooper/schedule-optimizer/internal/utils"
	"github.com/cwooper/schedule-optimizer/pkg/protoutils"
)

var CourseNamePattern = regexp.MustCompile(`^[A-Z/ ]{2,4} \d{3}[A-Z]?$`) // e.g. CSCI 497A

// Conflict structure to hold the conflicts in an array of courses
type Conflict struct {
	First  int // CRN of first course
	Second int // CRN of second course
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
func GenerateResponse(req models.RawRequest) *models.Response {
	g := NewGenerator()

	if len(req.Courses) > utils.MAX_INPUT_COURSES || len(req.Forced) > utils.MAX_INPUT_COURSES {
		errString := fmt.Sprintf("Cannot input more than %d courses", utils.MAX_INPUT_COURSES)
		g.response.Errors = append(g.response.Errors, errString)
		return g.response
	}

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
	g.generateCourses(req.Term)
	if len(g.response.Errors) > 0 {
		return g.response
	}

	// Verify that the forced courses were found in the output
	for i, forceName := range g.cleanedForcedNames {
		found := false
		for _, course := range g.courses {
			if forceName == course.Subject {
				found = true
				break
			}
		}
		if !found {
			g.cleanedForcedNames[i] = ""
			g.response.Warnings = append(g.response.Warnings, "Could not find "+forceName)
		}
	}

	// TODO: find conflicting asyncs (by course name)
	// 		 then find maximum possible asyncs together
	//       then subtract len(possibleAsyncs) from req.Min (clamp to 1)
	if len(g.response.Asyncs) > 0 { // currently works to alleviate some
		g.response.Warnings = append(g.response.Warnings,
			"ASYNC course generated below. It is not included in the schedule preview.")
		// req.Min -= 1 // should be handled by clampBounds
	}

	// Clamp and verify that the bounds work
	req.Min, req.Max = clampBounds(req.Min, req.Max)
	if len(g.courses) < req.Min {
		errString := fmt.Sprintf("Could not generate schedules with Minimum set to %d "+
			" (%d courses found).", req.Min, len(g.courses))
		g.response.Errors = append(g.response.Errors, errString)
		return g.response
	}
	if len(req.Forced) > req.Max {
		errString := fmt.Sprintf("Cannot force more than %d courses.", req.Max)
		g.response.Errors = append(g.response.Errors, errString)
		return g.response
	}

	// Generates the conflicts and the schedules
	g.generateConflicts()
	g.generateSchedules(req)

	return g.response
}

// Checks if the string array already contains the given string
func containsString(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// Cleans the course names by stripping any possible whitespace and verifying
// that the course names are valid
func (g *Generator) cleanCourseNames(courses []string) []string {
	cleanedNames := make([]string, 0)
	for _, courseName := range courses {
		courseName = strings.TrimSpace(courseName)
		if CourseNamePattern.MatchString(courseName) && !containsString(cleanedNames, courseName) {
			cleanedNames = append(cleanedNames, courseName)
		} else { // Invalid Course Name
			g.response.Warnings = append(g.response.Warnings,
				"Invalid course name: "+courseName)
		}
	}

	return cleanedNames
}

// Fills the courses array in the generator with courses in the database
func (g *Generator) generateCourses(term string) {
	// Load the term protobuf
	termFile := filepath.Join(utils.DataDirectory, term+".pb")

	proto, err := utils.LoadCoursesProtobuf(termFile)
	if err != nil {
		g.response.Warnings = append(g.response.Warnings,
			"Term does not exist: "+term)
		return
	}

	courseList := protoutils.ProtoToCourses(proto)
	courses := make([]models.Course, 0)

	// Search for the necessary courses of cleaned names
	for _, courseRequest := range g.cleanedCourseNames {
		found := false
		for _, course := range courseList {
			if course.Subject == courseRequest {
				found = true
				// If the course doesn't have a time to test, don't include it
				if models.HasAsyncOrTBD(course) {
					g.response.Asyncs = append(g.response.Asyncs, course)
				} else {
					courses = append(courses, course)
				}
			}
		}
		if !found {
			g.response.Warnings = append(g.response.Warnings,
				courseRequest+" is not offered this term.")
		}
	}

	// Verify that we found the forced course
	for _, forceRequest := range g.cleanedForcedNames {
		found := false
		for _, course := range courses {
			if course.Subject == forceRequest {
				found = true
				break
			}
		}
		if !found {
			errString := "Could not force " + forceRequest + " (could not find course)"
			g.response.Warnings = append(g.response.Warnings, errString)
		}
	}

	g.courses = courses
}

// Fills the conflicts array in the generator with known conflict pairs
// Must be called after filling the generators courses with objects
func (g *Generator) generateConflicts() {
	for i := 0; i < len(g.courses)-1; i++ {
		for j := i + 1; j < len(g.courses); j++ {
			if g.courses[i].Conflicts(g.courses[j]) {
				g.conflicts = append(g.conflicts, Conflict{
					First:  g.courses[i].CRN,
					Second: g.courses[j].CRN,
				})
			}
		}
	}
}

// Generates Schedules after finding conflicts and getting courses
func (g *Generator) generateSchedules(req models.RawRequest) {
	scheduleRequest := ScheduleRequest{
		Courses:   g.courses,
		Conflicts: g.conflicts,
		Min:       req.Min,
		Max:       req.Max,
	}

	schedules := Combinations(scheduleRequest)
	if len(schedules) == 0 {
		g.response.Errors = append(g.response.Errors, "No possible schedules found")
		return
	}

	// Remove schedules that don't have the forced courses
	schedules = filterSchedules(schedules, g.cleanedForcedNames)
	if len(schedules) == 0 {
		g.response.Errors = append(g.response.Errors, "No possible schedules found after forcing courses")
		return
	}

	// Store and sort the schedules
	g.response.Schedules = schedules
	sort.Slice(g.response.Schedules, func(i, j int) bool {
		return g.response.Schedules[i].Score > g.response.Schedules[j].Score
	})
}

// filterSchedules filters schedules based on forced courses
func filterSchedules(schedules []models.Schedule, forcedCourses []string) []models.Schedule {
	filtered := make([]models.Schedule, 0, len(schedules))

	for _, schedule := range schedules {
		if containsAllForcedCourses(schedule, forcedCourses) {
			filtered = append(filtered, schedule)
		}
	}

	return filtered
}

// containsAllForcedCourses checks if a schedule contains all forced courses
func containsAllForcedCourses(schedule models.Schedule, forcedCourses []string) bool {
	for _, forcedCourse := range forcedCourses {
		found := false
		for _, course := range schedule.Courses {
			if course.Subject == forcedCourse {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
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
