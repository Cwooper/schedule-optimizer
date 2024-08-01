// Package scraper for scraping course data and saving it
package scraper

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"google.golang.org/protobuf/proto"

	"schedule-optimizer/internal/models"
	pb "schedule-optimizer/internal/proto/generated"
	"schedule-optimizer/internal/utils"
	"schedule-optimizer/pkg/protoutils"
)

// Enum for subjects and terms
const (
	Subject int = iota
	Terms
)

// Parses the given file to an array strings split by lines
func fileToLines(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Create a scanner to read the file
	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return lines, nil
}

// Saves the given lines into a file specified by file
// Prepends the time to the top (format specified utils)
func linesToFile(lines []string, file string) error {
	// Prepend the current line to the list
	currentTime := time.Now().Format(utils.TIME_FORMAT)
	allLines := append([]string{currentTime}, lines...)

	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	// Write each line to the file
	writer := bufio.NewWriter(f)
	for _, line := range allLines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("failed to write line to file: %w", err)
		}
	}

	// Flush the writer to ensure all data is written to the file
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

// Fetches the subjects/terms from url
func fetchFromURL(option string) ([]string, error) {
	resp, err := http.Get(utils.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
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

// Fetches the given list from file or url dpeending on time since file creation
func fetchList(option int) ([]string, error) {
	// Create variables based upon type of fetch request
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

		currentTime := time.Now()
		duration := currentTime.Sub(fileTime)

		// If file creation to now is longer than MAX wait time defined in utils
		if duration < time.Duration(waitTime)*24*time.Hour {
			return lines[1:], nil
		}
	}

	// This will only run if all of the above fails
	// Fetch the items from url
	lines, err = fetchFromURL(htmlID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subjects from url: %w", err)
	}

	// Save the file for re-use
	linesToFile(lines, file)
	fmt.Printf("Longer than %d days, saved %v\n", waitTime, file)

	return lines, nil
}

// Filters the list of turns and returns the current year code
func filterTerms(terms []string) ([]string, string, error) {
	currentTerm := fmt.Sprintf("%d00", time.Now().Year())

	// At most there will ever be 8 terms above this one
	filteredTerms := make([]string, 8)
	for _, term := range terms {
		if term >= currentTerm && term != "ALL" {
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

func getCourses(subject, term, year string) ([]models.Course, error) {
	return nil, nil
}

func saveProtobuf(protobuf *pb.CourseList, filename string) error {
	data, err := proto.Marshal(protobuf)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write protobuf to file: %w", err)
	}

	return nil
}

func UpdateCourses() error {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(utils.DataDirectory, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	subjects, err := fetchList(Subject)
	if err != nil {
		return fmt.Errorf("failed to fetch subjects list: %w", err)
	}

	terms, err := fetchList(Terms)
	if err != nil {
		return fmt.Errorf("failed to fetch terms list: %w", err)
	}

	terms, year, err := filterTerms(terms)
	if err != nil {
		return fmt.Errorf("failed to filter terms list: %w", err)
	}

	fmt.Printf("Found %d subjects\n", len(subjects))
	fmt.Printf("Current terms: %v\n", terms)

	for _, term := range terms {
		// About 1500 courses per term, 2000 to be efficient
		courseList := make([]models.Course, 2000)
		for _, subject := range subjects {
			newCourses, err := getCourses(subject, term, year)
			if err != nil {
				return fmt.Errorf("failed to get courses: %w", err)
			}
			if len(newCourses) > 0 {
				courseList = append(courseList, newCourses...)
			}
		}

		protobuf := protoutils.CoursesToProto(courseList)

		// Save the protobuf
		termFile := filepath.Join(utils.DataDirectory, term, ".pb")
		if err := saveProtobuf(protobuf, termFile); err != nil {
			return fmt.Errorf("failed to save protobuf for term %s: %w", term, err)
		}

		fmt.Printf("%v: Found and saved %d courses to %v\n", term, len(courseList), termFile)
	}

	return nil
}
