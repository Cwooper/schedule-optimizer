// Package gpa for processing and updating protobufs with gpas
package gpa

import (
	"fmt"
	"strings"

	"github.com/cwooper/schedule-optimizer/internal/models"
	"github.com/cwooper/schedule-optimizer/internal/utils"
)

// Finds the instructor in the gpaData, converts name from last, first
func findInstructor(name string, subject string, gpaData models.GPAData) (string, bool) {
	if name == "Staff" || name == "TBD" || name == "N/A" {
		return "", false
	}

	var firstName, lastName string

	// Turn a name into simply "First Last" from "Last, First Middle"
	if match := utils.CommaNameRegexp.FindStringSubmatch(name); len(match) == 3 {
		lastName, firstName = match[1], match[2]
	} else {
		return "", false
	}

	// Create variations of the name to check
	variations := []string{
		fmt.Sprintf("%s %s", firstName, lastName),
		fmt.Sprintf("%s %s", lastName, firstName),
	}

	// Check for exact matches first
	for _, variation := range variations {
		if _, ok := gpaData.Professors[variation]; ok {
			return variation, true
		}
	}

	// Extract subject prefix (e.g., "CSCI" from "CSCI 347")
	matches := utils.SubjectRegexp.FindStringSubmatch(subject)
	if len(matches) < 2 {
		return "", false
	}
	subjectPrefix := matches[1] // First capture group contains the subject prefix

	// Check for professors with same last name who have taught the subject
	lastNameLower := strings.ToLower(lastName)
	if professors, ok := gpaData.LastNameIndex[lastNameLower]; ok {
		for _, profName := range professors {
			// Check if this professor has taught the subject
			if subjects, ok := gpaData.ProfessorSubjects[profName]; ok {
				if _, taught := subjects[subjectPrefix]; taught {
					return profName, true
				}
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

	if newSubject != "" { // Replace the subject if the GPA data had a different subject
		subject = strings.Replace(subject, courseName, newSubject, 1)
	}

	// Find the instructor
	instructorName, found := findInstructor(course.Sessions[0].Instructor, subject, gpaData)
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
		return courseGPA // We found the perfect match
	}

	// Use subject GPA if we couldn't find an instructor to match
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
