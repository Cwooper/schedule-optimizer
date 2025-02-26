package search

import (
	"sort"
	"strings"

	"github.com/cwooper/schedule-optimizer/internal/cache"
	"github.com/cwooper/schedule-optimizer/internal/models"
	"github.com/cwooper/schedule-optimizer/internal/utils"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// This file should be used for searching for a course in the course's course
// string using fuzzy searching

type Match struct {
	Course models.Course
	Score  int
}

func SearchCourses(searchTerm string, term string) models.Response {
	searchTerm = strings.ToUpper(searchTerm)
	resp := models.Response{}

	courseManager := cache.GetInstance()
	courseList, err := courseManager.GetCourseList(term)
	if err != nil {
		resp.Errors = append(resp.Errors, "Term does not exist: "+term)
		return resp
	}

	// Pre-allocate slice for matches
	matches := make([]Match, 0, len(courseList))

	// Perform fuzzy search on each course
	for i := range courseList {
		score := fuzzy.RankMatch(searchTerm, courseList[i].CourseString)
		if score != -1 {
			matches = append(matches, Match{
				Course: courseList[i],
				Score:  score,
			})
		}
	}

	// Sort matches by score (lower is better)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score < matches[j].Score
	})

	// Determine number of courses to return (up to util maximum)
	numResults := min(len(matches), utils.MAX_OUTPUT_SEARCH_COURSES)

	// Pre-allocate resp.Courses
	resp.Courses = make([]models.Course, numResults)

	// Add top matches to resp.Courses
	for i := range numResults {
		resp.Courses[i] = matches[i].Course
	}

	return resp
}
