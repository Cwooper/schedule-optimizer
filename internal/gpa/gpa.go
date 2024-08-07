// Package gpa for processing and updating protobufs with gpas
package gpa

import (
	"fmt"
	"strings"

	"schedule-optimizer/internal/models"
	"schedule-optimizer/internal/utils"
)

// Finds the instructor in the gpaData, converts name from last, first
func findInstructor(name string, professors models.Professors) (string, bool) {
	var firstName, lastName string

	if match := utils.CommaNameRegexp.FindStringSubmatch(name); len(match) == 3 {
		lastName, firstName = match[1], match[2]
	} else {
		return "", false
	}

	// Create variations of the name to check
	variations := []string{
		fmt.Sprintf("%s %s", firstName, lastName),
		fmt.Sprintf("%s %s", lastName, firstName),
		lastName,
	}

	// Check for exact matches first
	for _, variation := range variations {
		if _, ok := professors[variation]; ok {
			return variation, true
		}
	}

	// If no exact match, try partial matches
	for profName := range professors {
		profNameLower := strings.ToLower(profName)
		for _, variation := range variations {
			if strings.Contains(profNameLower, strings.ToLower(variation)) {
				return profName, true
			}
		}
	}

	return "", false
}

// Finds and processes the GPA of a course, returns the GPA for that course
func findGPA(course models.Course, gpaData models.GPAData) float64 {
	// Pull out the subjects course and convert it via utils.CourseSubjectMapping if necessary
	subject := course.Subject
	courseName := string(utils.SubjectRegexp.Find([]byte(course.Subject)))
	newSubject := utils.CourseSubjectMapping[string(subject)]

	if newSubject != "" {
		subject = strings.Replace(subject, courseName, newSubject, 1)
	}

	// Find the instructor
	instructorName, found := findInstructor(course.Sessions[0].Instructor, gpaData.Professors)
	if !found {
		// If instructor not found, fall back to subject GPA
		if subjectGPA, ok := gpaData.Subjects[subject]; ok {
			return subjectGPA
		}
		return 0.0 // Default GPA if neither instructor nor subject found
	}

	// Check for course-specific GPA
	courseKey := models.CourseKey(subject, instructorName)
	if courseGPA, ok := gpaData.CourseGPAs[courseKey]; ok {
		return courseGPA
	}

	// Fall back to instructor's overall GPA
	if instructorGPA, ok := gpaData.Professors[instructorName]; ok {
		return instructorGPA
	}

	// If all else fails, use subject GPA
	if subjectGPA, ok := gpaData.Subjects[subject]; ok {
		return subjectGPA
	}

	return 0.0 // Default GPA if no match found
}

// Adds the GPA variable to each course
func GenerateGPA(courseList *[]models.Course) error {
	// Get the GPAData map
	gpaData, err := GetGPAData()
	if err != nil {
		return fmt.Errorf("failed to get gpa data: %w", err)
	}

	// Create a map of courseKey to GPA
	gpaMap := make(map[string]float64, len(gpaData.CourseGPAs))

	// Iterate through every course
	for i, course := range *courseList {
		// If CourseKey is in the map, use that GPA
		courseKey := models.CourseKey(course.Subject, course.Sessions[0].Instructor)
		if gpa, ok := gpaMap[courseKey]; ok {
			(*courseList)[i].GPA = gpa
			continue
		}
		
		// If the course isn't in the map, then process the gpa, and add course to map
		courseGpa := findGPA(course, gpaData)
		gpaMap[courseKey] = courseGpa
		(*courseList)[i].GPA = courseGpa
	}

	return nil
}
