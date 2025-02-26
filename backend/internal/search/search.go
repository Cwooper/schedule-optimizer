package search

import (
	"regexp"
	"sort"
	"strings"

	"github.com/cwooper/schedule-optimizer/internal/cache"
	"github.com/cwooper/schedule-optimizer/internal/models"
	"github.com/cwooper/schedule-optimizer/internal/utils"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Match represents a course match with its score
type Match struct {
	Course models.Course
	Score  int // Higher is better
}

// QueryComponents represents parsed components of a search query
type QueryComponents struct {
	OriginalQuery   string   // The original search query
	NormalizedQuery string   // Uppercase, trimmed query
	SubjectCode     string   // Potential subject code (e.g., "CSCI")
	Numbers         []string // Potential course numbers (e.g., ["4"] for "CSCI 4")
	Terms           []string // All terms in the query
	HasComma        bool     // Indicates potential instructor name
}

// Regular expressions for parsing query components
var (
	numberRegex      = regexp.MustCompile(`\b\d+[A-Za-z]?\b`) // Matches course numbers, including those with letter suffixes like "497A"
	subjectCodeCache map[string]bool                          // Cache of known subject codes
)

// initializeSubjectCodes builds a map of all subject codes from the course list
func initializeSubjectCodes(courses []models.Course) {
	if subjectCodeCache != nil {
		return // Already initialized
	}

	subjectCodeCache = make(map[string]bool)
	for _, course := range courses {
		parts := strings.Fields(course.Subject)
		if len(parts) > 0 {
			subjectCodeCache[parts[0]] = true
		}
	}
}

// parseQuery analyzes a search query and extracts its components
func parseQuery(query string) QueryComponents {
	// Normalize the query
	originalQuery := query
	query = strings.TrimSpace(strings.ToUpper(query))

	components := QueryComponents{
		OriginalQuery:   originalQuery,
		NormalizedQuery: query,
		HasComma:        strings.Contains(query, ","),
		Terms:           strings.Fields(query),
	}

	// Extract potential course numbers
	components.Numbers = numberRegex.FindAllString(query, -1)

	// Identify potential subject code (but don't set for instructor searches with comma)
	if !components.HasComma {
		for _, term := range components.Terms {
			// If this term is in our subject code cache, it's likely a subject code
			if subjectCodeCache[term] {
				components.SubjectCode = term
				break
			}

			// If term matches the subject pattern (2-4 uppercase letters), consider it a potential subject
			if match, _ := regexp.MatchString(`^[A-Z]{2,4}$`, term); match {
				components.SubjectCode = term
				break
			}
		}
	}

	return components
}

// scoreExactSubjectMatch scores a course based on exact subject code match
// Returns high score (1000) if the subject code matches exactly
func scoreExactSubjectMatch(course models.Course, components QueryComponents) int {
	if components.SubjectCode == "" {
		return 0
	}

	// Extract subject code from course
	parts := strings.Fields(course.Subject)
	if len(parts) == 0 {
		return 0
	}

	subjectCode := parts[0]

	// Direct comparison with case insensitivity
	if strings.EqualFold(subjectCode, components.SubjectCode) {
		return 1000
	}

	// Partial subject match (lower score)
	if strings.HasPrefix(strings.ToUpper(subjectCode), components.SubjectCode) {
		return 800
	}

	// If the query is the full subject (e.g., "CSCI 141"), give a high score
	if strings.HasPrefix(strings.ToUpper(course.Subject), components.NormalizedQuery) {
		return 950
	}

	return 0
}

// scoreNumberMatch scores a course based on course number matches
// Returns high score (500-700) if any number in the query matches the course number
func scoreNumberMatch(course models.Course, components QueryComponents) int {
	if len(components.Numbers) == 0 {
		return 0
	}

	// Extract course number
	parts := strings.Fields(course.Subject)
	if len(parts) < 2 {
		return 0
	}

	courseNumber := parts[1]
	highestScore := 0

	// Check each number in the query
	for _, num := range components.Numbers {
		// Exact number match
		if courseNumber == num {
			return 700 // Highest score for exact match
		}

		// Prefix match (e.g., "4" matches "497")
		if strings.HasPrefix(courseNumber, num) {
			score := 600
			// Longer prefix matches get higher scores
			if len(num) > 1 {
				score = 650
			}
			if score > highestScore {
				highestScore = score
			}
		}
	}

	return highestScore
}

// scoreInstructorMatch scores a course based on instructor name match
// Uses fuzzy matching with higher weight if comma is present (likely instructor search)
func scoreInstructorMatch(course models.Course, components QueryComponents) int {
	if course.Instructor == "" {
		return 0
	}

	// Join all terms for a more comprehensive match
	queryText := components.NormalizedQuery
	instructorUpper := strings.ToUpper(course.Instructor)

	// Direct contains check first (faster than fuzzy)
	if strings.Contains(instructorUpper, queryText) {
		// Higher score if the query has a comma (likely an instructor search)
		if components.HasComma {
			return 500
		}
		return 400
	}

	// Fuzzy match for instructor
	fuzzScore := fuzzy.RankMatch(queryText, instructorUpper)
	if fuzzScore == -1 {
		return 0
	}

	// Convert fuzzy score (lower is better) to a positive score
	// Scale based on whether the query has a comma
	if components.HasComma {
		return max(0, 450-fuzzScore*3)
	}
	return max(0, 350-fuzzScore*3)
}

// scoreTitleMatch scores a course based on course title match
func scoreTitleMatch(course models.Course, components QueryComponents) int {
	if course.Title == "" {
		return 0
	}

	titleUpper := strings.ToUpper(course.Title)
	queryText := components.NormalizedQuery

	// Direct contains check first
	if strings.Contains(titleUpper, queryText) {
		return 300
	}

	// Check for individual terms matches
	if len(components.Terms) > 0 {
		for _, term := range components.Terms {
			if strings.Contains(titleUpper, term) {
				return 250
			}
		}
	}

	// Fuzzy match for title
	fuzzScore := fuzzy.RankMatch(queryText, titleUpper)
	if fuzzScore == -1 {
		return 0
	}

	// Convert fuzzy score to positive score
	return max(0, 250-fuzzScore*2)
}

// scoreFallbackMatch uses the original course string for a fallback match
// Only applies if no other matches were found
func scoreFallbackMatch(course models.Course, components QueryComponents) int {
	// Use the course string for a comprehensive fuzzy match
	fuzzScore := fuzzy.RankMatch(components.NormalizedQuery, course.CourseString)
	if fuzzScore == -1 {
		return 0
	}

	// For short queries, a direct contains check might be more relevant
	if len(components.NormalizedQuery) > 2 && strings.Contains(course.CourseString, components.NormalizedQuery) {
		return 100
	}

	// Convert fuzzy score to positive score (lower weight than specific matches)
	return max(0, 100-fuzzScore)
}

// scoreCourse calculates a total score for how well a course matches the query
func scoreCourse(course models.Course, components QueryComponents) int {
	// Start with scores from specific strategies
	subjectScore := scoreExactSubjectMatch(course, components)
	numberScore := scoreNumberMatch(course, components)
	instructorScore := scoreInstructorMatch(course, components)
	titleScore := scoreTitleMatch(course, components)

	// Only use fallback if no other matches were found
	fallbackScore := 0
	if subjectScore == 0 && numberScore == 0 && instructorScore == 0 && titleScore == 0 {
		fallbackScore = scoreFallbackMatch(course, components)
	}

	// Total score is the sum of all individual scores
	return subjectScore + numberScore + instructorScore + titleScore + fallbackScore
}

// SearchCourses is the main search function that finds relevant courses
// based on a search term and term identifier
func SearchCourses(searchTerm string, term string) models.Response {
	resp := models.Response{}

	// Get courses from the cache
	courseManager := cache.GetInstance()
	courseList, err := courseManager.GetCourseList(term)
	if err != nil {
		resp.Errors = append(resp.Errors, "Term does not exist: "+term)
		return resp
	}

	// Initialize subject codes if needed
	initializeSubjectCodes(courseList)

	// Parse the search query
	components := parseQuery(searchTerm)

	// Score and collect matches
	matches := make([]Match, 0, len(courseList))
	for i := range courseList {
		score := scoreCourse(courseList[i], components)
		if score > 0 {
			matches = append(matches, Match{
				Course: courseList[i],
				Score:  score,
			})
		}
	}

	// Sort matches by score (higher is better)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// If we have enough matches already, return them
	maxResults := utils.MAX_OUTPUT_SEARCH_COURSES
	if len(matches) >= maxResults {
		resp.Courses = make([]models.Course, maxResults)
		for i := range maxResults {
			resp.Courses[i] = matches[i].Course
		}
		return resp
	}

	// If we don't have enough matches with the component-based approach,
	// supplement with the original fuzzy search
	if len(matches) < maxResults {
		// Convert existing matches to a map for quick lookup
		existingMatches := make(map[int]bool)
		for _, match := range matches {
			existingMatches[match.Course.CRN] = true
		}

		// Add additional matches using the original fuzzy search approach
		additionalMatches := fuzzySearchCourses(searchTerm, courseList, existingMatches, maxResults-len(matches))
		matches = append(matches, additionalMatches...)

		// Re-sort all matches
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Score > matches[j].Score
		})
	}

	// Determine final number of results to return
	numResults := min(len(matches), maxResults)

	// Add matches to response
	resp.Courses = make([]models.Course, numResults)
	for i := range numResults {
		resp.Courses[i] = matches[i].Course
	}

	return resp
}

// fuzzySearchCourses implements the original fuzzy search as a fallback
// It excludes courses that are already in existingMatches
func fuzzySearchCourses(searchTerm string, courseList []models.Course, existingMatches map[int]bool, limit int) []Match {
	searchTerm = strings.ToUpper(searchTerm)
	matches := make([]Match, 0)

	// Direct contains check first (faster than fuzzy)
	for i := range courseList {
		// Skip if this course is already matched
		if existingMatches[courseList[i].CRN] {
			continue
		}

		// If the course string contains the search term directly, that's a high-quality match
		if strings.Contains(courseList[i].CourseString, searchTerm) {
			matches = append(matches, Match{
				Course: courseList[i],
				Score:  75, // Higher score for direct contains
			})
			continue
		}
	}

	// If we didn't find enough direct matches, try fuzzy matching
	if len(matches) < limit {
		for i := range courseList {
			// Skip if this course is already matched
			if existingMatches[courseList[i].CRN] || containsCourse(matches, courseList[i].CRN) {
				continue
			}

			// Calculate fuzzy match score
			score := fuzzy.RankMatch(searchTerm, courseList[i].CourseString)
			if score != -1 {
				// Convert to positive score (original implementation used negative ranking)
				positiveScore := 50 - min(score, 50) // Lower weight than component matches
				matches = append(matches, Match{
					Course: courseList[i],
					Score:  positiveScore,
				})
			}
		}
	}

	// Sort by score
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Limit results
	if len(matches) > limit {
		matches = matches[:limit]
	}

	return matches
}

// Helper to check if a course with a specific CRN is already in the matches
func containsCourse(matches []Match, crn int) bool {
	for _, match := range matches {
		if match.Course.CRN == crn {
			return true
		}
	}
	return false
}
