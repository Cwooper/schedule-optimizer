// Package scraper for scraping course data and saving it
package scraper

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/cwooper/schedule-optimizer/internal/gpa"
	"github.com/cwooper/schedule-optimizer/internal/models"
	"github.com/cwooper/schedule-optimizer/internal/utils"
	"github.com/cwooper/schedule-optimizer/pkg/protoutils"
)

// Enum for subjects and terms
const (
	Subject int = iota
	Terms
)

type Result struct {
	Term  string
	Count int
	Error error
}

// Gets the subjects/terms from url
func getSelectListFromURL(option string) ([]string, error) {
	resp, err := http.Get(utils.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error: status code %d accessing %s", resp.StatusCode, utils.URL)
	}

	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Find the select subject element and store them into a subjects slice
	selectOption := fmt.Sprintf("select#%s option", option)
	var subjects []string
	doc.Find(selectOption).Each(func(i int, s *goquery.Selection) {
		if value, exists := s.Attr("value"); exists {
			subjects = append(subjects, value)
		}
	})

	return subjects, nil
}

// Gets the given list from file or url dpeending on time since file creation
func getList(option int) ([]string, error) {
	// Create variables based upon type of get request
	var filename string
	var htmlID string
	var waitTime int
	if option == Subject {
		filename = "subjects.txt"
		htmlID = utils.SUBJECT_ID
		waitTime = utils.MAX_SUBJECT_WAIT
	} else {
		filename = "terms.txt"
		htmlID = utils.TERM_ID
		waitTime = utils.MAX_TERM_WAIT
	}

	var lines []string
	file := filepath.Join(utils.DataDirectory, filename)
	_, err := os.Stat(file)

	if err == nil {
		// Parse the file and get items from file
		lines, err = fileToLines(file)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %v to lines: %w", file, err)
		}

		fileTime, err := time.Parse(utils.TIME_FORMAT, lines[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse time in %v: %w", lines[0], err)
		}

		// If file creation to now is longer than MAX wait time defined in utils
		if time.Since(fileTime) < time.Duration(waitTime)*24*time.Hour {
			return lines[1:], nil
		}
	}

	// This will only run if all of the above fails
	// Get the items from url
	lines, err = getSelectListFromURL(htmlID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subjects from url: %w", err)
	}

	// Save the file for re-use
	linesToFile(lines, file)
	fmt.Printf("Longer than %d days, saved %v\n", waitTime, file)

	return lines, nil
}

// Filters the list of turns and returns the current year code
func filterTerms(terms []string) ([]string, string, error) {
	currentTerm := fmt.Sprintf("%d00", time.Now().Year())

	var filteredTerms []string
	for _, term := range terms {
		if term >= currentTerm && term != "All" {
			filteredTerms = append(filteredTerms, term)
		}
	}

	if len(filteredTerms) == 0 {
		return nil, "", fmt.Errorf("no valid terms found")
	}

	largestTerm := filteredTerms[0]
	for _, term := range filteredTerms {
		if term > largestTerm {
			largestTerm = term
		}
	}

	yearHigh, err := strconv.Atoi(largestTerm[2:4])
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse year: %w", err)
	}

	yearLow := yearHigh - 1
	year := fmt.Sprintf("%02d%02d", yearLow, yearHigh)
	return filteredTerms, year, nil
}

// Returns a list of all courses found for this term
func getCoursesFromURL(subjects []string, term, year string) ([]models.Course, error) {
	// About 1500 courses per term
	var courseList []models.Course

	for _, subject := range subjects {
		newCourses, err := getSubjectFromURL(subject, term, year)
		if err != nil {
			return nil, fmt.Errorf("failed to get courses: %w", err)
		}
		if len(newCourses) > 0 {
			courseList = append(courseList, newCourses...)
		}
	}

	return courseList, nil
}

// Gets the courses for the given term, returns the total courses found
func getCourses(subjects []string, term, year string) Result {
	termFile := filepath.Join(utils.DataDirectory, term+".pb")

	// Check if term protobuf already exists
	if _, err := os.Stat(termFile); err == nil {
		existingProto, err := utils.LoadCoursesProtobuf(termFile)
		if err != nil {
			return Result{
				Term:  term,
				Error: fmt.Errorf("failed to load existing protobuf: %w", err),
			}
		}

		pullTime := existingProto.PullTimestamp.AsTime()
		if time.Since(pullTime) < utils.MAX_COURSE_WAIT*24*time.Hour {
			courseList := protoutils.ProtoToCourses(existingProto)
			return Result{
				Term:  term,
				Count: len(courseList),
			}
		}
	}

	// Fetch new data
	courseList, err := getCoursesFromURL(subjects, term, year)
	if err != nil {
		return Result{Term: term, Error: fmt.Errorf("failed to get courses from url: %w", err)}
	}

	start := time.Now()
	err = gpa.GenerateGPA(&courseList)
	if err != nil {
		return Result{
			Term:  term,
			Error: fmt.Errorf("failed to generate gpa for course list: %w", err),
		}
	}
	duration := time.Since(start)
	log.Printf("	GenerateGPA for %s took %v to execute\n", term, duration)

	protobuf := protoutils.CoursesToProto(courseList)
	protobuf.PullTimestamp = timestamppb.Now()

	// Save the protobuf
	if err := utils.SaveCoursesProtobuf(protobuf, termFile); err != nil {
		return Result{
			Term:  term,
			Error: fmt.Errorf("failed to save protobuf for term %s: %w", term, err),
		}
	}

	log.Printf("%v: Found and saved %d courses to %v\n", term, len(courseList), termFile)

	return Result{
		Term:  term,
		Count: len(courseList),
	}
}

// Updates all courses in Term protobufs if deemed necessary
// It is up to the user to process the gpa of each course for each term
func UpdateCourses() error {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(utils.DataDirectory, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	subjects, err := getList(Subject)
	if err != nil {
		return fmt.Errorf("failed to get subjects list: %w", err)
	}

	terms, err := getList(Terms)
	if err != nil {
		return fmt.Errorf("failed to get terms list: %w", err)
	}

	terms, year, err := filterTerms(terms)
	if err != nil {
		return fmt.Errorf("failed to filter terms list: %w", err)
	}

	log.Printf("	Found %d subjects\n", len(subjects))
	log.Printf("	Current terms: %v\n", terms)

	var wg sync.WaitGroup
	results := make(chan Result, len(terms))

	for _, term := range terms {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			results <- getCourses(subjects, t, year)
		}(term)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	totalCount := 0
	for result := range results {
		if result.Error != nil {
			return result.Error
		}
		totalCount += result.Count
	}

	log.Printf("	Found %d total courses\n", totalCount)
	return nil
}
