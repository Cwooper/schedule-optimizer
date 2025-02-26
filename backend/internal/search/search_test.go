package search

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cwooper/schedule-optimizer/internal/cache"
	"github.com/cwooper/schedule-optimizer/internal/models"
)

// FIXME tests are not runing

// MockCourseCache implements the CourseCache interface for testing
type MockCourseCache struct {
	courses []models.Course
	// Function fields that can be changed during tests
	getCourseListFunc    func(string) ([]models.Course, error)
	getCoursesFunc       func(string, string) ([]models.Course, error)
	getGlobalCoursesFunc func(string) ([]models.Course, error)
}

// NewMockCourseCache creates a new mock cache with default behaviors
func NewMockCourseCache(courses []models.Course) *MockCourseCache {
	mock := &MockCourseCache{
		courses: courses,
	}

	// Set default behaviors
	mock.getCourseListFunc = func(term string) ([]models.Course, error) {
		// Return a copy of the courses to prevent modifications affecting the original
		coursesCopy := make([]models.Course, len(mock.courses))
		copy(coursesCopy, mock.courses)
		return coursesCopy, nil
	}

	mock.getCoursesFunc = func(term, subject string) ([]models.Course, error) {
		var result []models.Course
		for _, course := range mock.courses {
			if strings.HasPrefix(course.Subject, subject) {
				result = append(result, course)
			}
		}
		return result, nil
	}

	mock.getGlobalCoursesFunc = func(subject string) ([]models.Course, error) {
		var result []models.Course
		for _, course := range mock.courses {
			if strings.HasPrefix(course.Subject, subject) {
				result = append(result, course)
			}
		}
		return result, nil
	}

	return mock
}

func (m *MockCourseCache) GetGlobalCourses(subject string) ([]models.Course, error) {
	return m.getGlobalCoursesFunc(subject)
}

func (m *MockCourseCache) GetCourses(term, subject string) ([]models.Course, error) {
	return m.getCoursesFunc(term, subject)
}

func (m *MockCourseCache) GetCourseList(term string) ([]models.Course, error) {
	return m.getCourseListFunc(term)
}

func (m *MockCourseCache) PreloadCache(terms []string) error {
	return nil
}

func (m *MockCourseCache) Invalidate() {
	// Do nothing for tests
}

// Helper function to create a course for testing
func createTestCourse(subject, title, instructor string, crn int) models.Course {
	// Create the course string as it would be in the real application
	// All fields should be uppercase to match how the real application does it
	courseString := strings.ToUpper(subject) + "|" + strings.ToUpper(title) + "|" + strings.ToUpper(instructor)

	return models.Course{
		Subject:      subject,
		Title:        title,
		Instructor:   instructor,
		CRN:          crn,
		CourseString: courseString,
		Sessions:     []models.Session{},
	}
}

// Test data setup
func setupTestCourses() []models.Course {
	return []models.Course{
		createTestCourse("CSCI 141", "Computer Programming I", "Smith, John", 12345),
		createTestCourse("CSCI 241", "Data Structures", "Doe, Jane", 12346),
		createTestCourse("CSCI 301", "Formal Languages", "Brown, Robert", 12347),
		createTestCourse("CSCI 305", "Analysis of Algorithms", "Williams, Mary", 12348),
		createTestCourse("CSCI 405", "Analysis of Algorithms II", "Johnson, David", 12349),
		createTestCourse("CSCI 497A", "Senior Project", "Miller, Sarah", 12350),
		createTestCourse("CSCI 497B", "Senior Project II", "Wilson, James", 12351),
		createTestCourse("MATH 124", "Calculus I", "Taylor, Elizabeth", 12352),
		createTestCourse("MATH 125", "Calculus II", "Anderson, Michael", 12353),
		createTestCourse("PHYS 161", "Physics with Calculus", "Martinez, Maria", 12354),
	}
}

// Test parsing query components
func TestParseQuery(t *testing.T) {
	// Initialize subject codes
	subjectCodeCache = map[string]bool{
		"CSCI": true,
		"MATH": true,
		"PHYS": true,
	}

	tests := []struct {
		name     string
		query    string
		expected QueryComponents
	}{
		{
			name:  "Subject code only",
			query: "CSCI",
			expected: QueryComponents{
				OriginalQuery:   "CSCI",
				NormalizedQuery: "CSCI",
				SubjectCode:     "CSCI",
				Terms:           []string{"CSCI"},
				HasComma:        false,
			},
		},
		{
			name:  "Subject code with number",
			query: "CSCI 4",
			expected: QueryComponents{
				OriginalQuery:   "CSCI 4",
				NormalizedQuery: "CSCI 4",
				SubjectCode:     "CSCI",
				Numbers:         []string{"4"},
				Terms:           []string{"CSCI", "4"},
				HasComma:        false,
			},
		},
		{
			name:  "Instructor name with comma",
			query: "Smith, John",
			expected: QueryComponents{
				OriginalQuery:   "Smith, John",
				NormalizedQuery: "SMITH, JOHN",
				Terms:           []string{"SMITH,", "JOHN"},
				HasComma:        true,
			},
		},
		{
			name:  "Mixed case input",
			query: "cSci 497a",
			expected: QueryComponents{
				OriginalQuery:   "cSci 497a",
				NormalizedQuery: "CSCI 497A",
				SubjectCode:     "CSCI",
				Numbers:         []string{"497A"},
				Terms:           []string{"CSCI", "497A"},
				HasComma:        false,
			},
		},
		{
			name:  "Multiple numbers",
			query: "CSCI 4 5",
			expected: QueryComponents{
				OriginalQuery:   "CSCI 4 5",
				NormalizedQuery: "CSCI 4 5",
				SubjectCode:     "CSCI",
				Numbers:         []string{"4", "5"},
				Terms:           []string{"CSCI", "4", "5"},
				HasComma:        false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseQuery(tt.query)

			// Check subject code
			if result.SubjectCode != tt.expected.SubjectCode {
				t.Errorf("SubjectCode: got %v, want %v", result.SubjectCode, tt.expected.SubjectCode)
			}

			// Check numbers (ignoring order)
			if !sameStringSlices(result.Numbers, tt.expected.Numbers) {
				t.Errorf("Numbers: got %v, want %v", result.Numbers, tt.expected.Numbers)
			}

			// Check if comma detection is correct
			if result.HasComma != tt.expected.HasComma {
				t.Errorf("HasComma: got %v, want %v", result.HasComma, tt.expected.HasComma)
			}
		})
	}
}

// Helper to check if two string slices have the same elements (ignoring order)
func sameStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]int)
	for _, s := range a {
		aMap[s]++
	}

	bMap := make(map[string]int)
	for _, s := range b {
		bMap[s]++
	}

	return reflect.DeepEqual(aMap, bMap)
}

// Test scoring individual components
func TestScoringComponents(t *testing.T) {
	// Sample courses for testing
	csCourse := createTestCourse("CSCI 141", "Computer Programming I", "Smith, John", 12345)
	advCsCourse := createTestCourse("CSCI 497A", "Senior Project", "Miller, Sarah", 12350)
	mathCourse := createTestCourse("MATH 124", "Calculus I", "Taylor, Elizabeth", 12352)

	// Initialize subject codes
	subjectCodeCache = map[string]bool{
		"CSCI": true,
		"MATH": true,
	}

	// Test subject matching
	t.Run("Test Subject Matching", func(t *testing.T) {
		components := parseQuery("CSCI")

		if score := scoreExactSubjectMatch(csCourse, components); score != 1000 {
			t.Errorf("Expected score 1000 for exact subject match, got %d", score)
		}

		if score := scoreExactSubjectMatch(mathCourse, components); score != 0 {
			t.Errorf("Expected score 0 for non-matching subject, got %d", score)
		}
	})

	// Test number matching
	t.Run("Test Number Matching", func(t *testing.T) {
		components := parseQuery("4")

		if score := scoreNumberMatch(csCourse, components); score != 0 {
			t.Errorf("Expected score 0 for non-matching number, got %d", score)
		}

		if score := scoreNumberMatch(advCsCourse, components); score != 600 {
			t.Errorf("Expected score 600 for prefix number match, got %d", score)
		}

		// Test exact number match
		exactComponents := parseQuery("497A")
		if score := scoreNumberMatch(advCsCourse, exactComponents); score != 700 {
			t.Errorf("Expected score 700 for exact number match, got %d", score)
		}
	})

	// Test instructor matching
	t.Run("Test Instructor Matching", func(t *testing.T) {
		// Test with comma (instructor search)
		commaComponents := parseQuery("Smith, J")
		score := scoreInstructorMatch(csCourse, commaComponents)
		if score <= 0 {
			t.Errorf("Expected positive score for instructor match with comma, got %d", score)
		}

		// Test without comma
		noCommaComponents := parseQuery("Smith")
		score = scoreInstructorMatch(csCourse, noCommaComponents)
		if score <= 0 {
			t.Errorf("Expected positive score for instructor match without comma, got %d", score)
		}

		// Should score higher with comma
		if score >= scoreInstructorMatch(csCourse, commaComponents) {
			t.Errorf("Expected higher score with comma than without")
		}
	})

	// Test title matching
	t.Run("Test Title Matching", func(t *testing.T) {
		components := parseQuery("Programming")

		if score := scoreTitleMatch(csCourse, components); score <= 0 {
			t.Errorf("Expected positive score for title match, got %d", score)
		}

		if score := scoreTitleMatch(mathCourse, components); score > 0 {
			t.Errorf("Expected zero score for non-matching title, got %d", score)
		}
	})
}

// Test the overall search function
func TestSearchCourses(t *testing.T) {
	// Set up test courses
	courses := setupTestCourses()

	// Create and set mock cache
	mockCache := NewMockCourseCache(courses)
	cache.SetInstance(mockCache)
	defer cache.ResetInstance()

	// Initialize subject codes
	initializeSubjectCodes(courses)

	// Add debug output
	t.Logf("Test setup: %d courses in mock cache", len(courses))
	for i, course := range courses {
		t.Logf("Course %d: Subject=%s, Title=%s, Instructor=%s, CourseString=%s",
			i, course.Subject, course.Title, course.Instructor, course.CourseString)
	}

	// Force the mock to return the exact courses we set up
	mockCache.getCourseListFunc = func(term string) ([]models.Course, error) {
		return courses, nil
	}

	tests := []struct {
		name       string
		searchTerm string
		wantCRNs   []int // Expected CRNs in the result, in order
	}{
		{
			name:       "Subject code search",
			searchTerm: "CSCI",
			wantCRNs:   []int{12345, 12346, 12347, 12348, 12349, 12350, 12351}, // All CSCI courses
		},
		{
			name:       "Course number prefix search",
			searchTerm: "CSCI 4",
			wantCRNs:   []int{12349, 12350, 12351}, // CSCI 405, 497A, 497B
		},
		{
			name:       "Instructor search with comma",
			searchTerm: "Smith, John",
			wantCRNs:   []int{12345}, // CSCI 141 with Smith, John
		},
		{
			name:       "Title search",
			searchTerm: "Algorithms",
			wantCRNs:   []int{12348, 12349}, // Courses with "Algorithms" in title
		},
		{
			name:       "Mixed subject and number",
			searchTerm: "MATH 12",
			wantCRNs:   []int{12352, 12353}, // MATH 124, 125
		},
		{
			name:       "Lowercase input",
			searchTerm: "csci 497",
			wantCRNs:   []int{12350, 12351}, // CSCI 497A, 497B
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SearchCourses(tt.searchTerm, "202520") // Term doesn't matter with mock

			// Debug output to diagnose issues
			t.Logf("Search term: %s", tt.searchTerm)
			components := parseQuery(tt.searchTerm)
			t.Logf("Parsed components: SubjectCode=%s, Numbers=%v, HasComma=%v, Terms=%v",
				components.SubjectCode, components.Numbers, components.HasComma, components.Terms)

			// Debug scoring for first few courses
			for i, course := range courses {
				if i < 3 || course.CRN == tt.wantCRNs[0] {
					score := scoreCourse(course, components)
					t.Logf("Course %s score: %d", course.Subject, score)

					// Break down scoring
					subScore := scoreExactSubjectMatch(course, components)
					numScore := scoreNumberMatch(course, components)
					instScore := scoreInstructorMatch(course, components)
					titleScore := scoreTitleMatch(course, components)
					t.Logf("  Breakdown: subject=%d, number=%d, instructor=%d, title=%d",
						subScore, numScore, instScore, titleScore)
				}
			}

			// Check number of results
			if len(result.Courses) != len(tt.wantCRNs) {
				t.Errorf("Got %d results, want %d", len(result.Courses), len(tt.wantCRNs))

				// Show what results we did get
				if len(result.Courses) > 0 {
					t.Logf("Actual results:")
					for i, c := range result.Courses {
						t.Logf("  %d: %s (CRN: %d)", i, c.Subject, c.CRN)
					}
				}
			}

			// Check CRNs in order
			for i := range min(len(result.Courses), len(tt.wantCRNs)) {
				if result.Courses[i].CRN != tt.wantCRNs[i] {
					t.Errorf("Result %d: got CRN %d, want %d", i, result.Courses[i].CRN, tt.wantCRNs[i])
				}
			}
		})
	}
}

// Test that fuzzy fallback works when component matching fails
func TestFuzzyFallback(t *testing.T) {
	// Set up test courses
	courses := setupTestCourses()

	// Create and set mock cache
	mockCache := &MockCourseCache{courses: courses}
	cache.SetInstance(mockCache)
	defer cache.ResetInstance()

	// Initialize subject codes
	initializeSubjectCodes(courses)

	// This search should use fuzzy fallback as the components won't match strongly
	result := SearchCourses("projectsenior", "202520")

	// Should find courses with "Senior Project" in the title
	found := false
	for _, course := range result.Courses {
		if course.CRN == 12350 || course.CRN == 12351 { // Senior Project courses
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Fuzzy fallback failed to find relevant courses")
	}
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	// Set up test courses
	courses := setupTestCourses()

	// Create and set mock cache
	mockCache := &MockCourseCache{courses: courses}
	cache.SetInstance(mockCache)
	defer cache.ResetInstance()

	// Initialize subject codes
	initializeSubjectCodes(courses)

	// Empty search term
	t.Run("Empty search term", func(t *testing.T) {
		result := SearchCourses("", "202520")
		if len(result.Courses) > 0 {
			t.Errorf("Expected no results for empty search term, got %d", len(result.Courses))
		}
	})

	// Nonsensical search term
	t.Run("Nonsensical search term", func(t *testing.T) {
		result := SearchCourses("xyzxyzxyz", "202520")
		if len(result.Courses) > 0 {
			t.Errorf("Expected no results for nonsensical search term, got %d", len(result.Courses))
		}
	})

	// Invalid term
	t.Run("Invalid term", func(t *testing.T) {
		// Override mock to return error for this test
		originalFunc := mockCache.getCourseListFunc
		mockCache.getCourseListFunc = func(term string) ([]models.Course, error) {
			return nil, fmt.Errorf("term not found")
		}
		defer func() { mockCache.getCourseListFunc = originalFunc }()

		result := SearchCourses("CSCI", "invalid")
		if len(result.Errors) == 0 {
			t.Errorf("Expected error for invalid term, got none")
		}
	})
}
