// Package scraper for scraping course data and saving it
package scraper

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/protobuf/proto"

	"schedule-optimizer/internal/models"
	pb "schedule-optimizer/internal/proto/generated"
	"schedule-optimizer/internal/utils"
	"schedule-optimizer/pkg/protoutils"
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
func arrayToFile(lines []string, file string) error {
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


func fetchSubjectsFromURL() ([]string, error) {

}

func fetchSubjectList() ([]string, error) {
	subjectsFile := filepath.Join(utils.DataDirectory, "subjects.txt")
	_, err := os.Stat(subjectsFile)
	var subjects []string

	if os.IsExist(err) {
		// Parse the file and get subjects from file
		subjects, err = fileToLines(subjectsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %v to lines: %w", subjectsFile, err)
		}

		fileTime, err := time.Parse(utils.TIME_FORMAT, subjects[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse time in %v: %w", subjects[0], err)
		}

		currentTime := time.Now()
		duration := currentTime.Sub(fileTime)

		// If file creation to now is longer than MAX wait time defined in utils
		if duration < utils.MAX_SUBJECT_WAIT*24*time.Hour {
			return subjects[1:], nil
		}
	}

	// This will only run if all of the above fails
	// Fetch the subjects from url
	subjects, err = fetchSubjectsFromURL()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subjects from url: %w", err)
	}

	arrayToFile(subjects, subjectsFile)
	fmt.Printf("Longer than %d days, saved %v\n", utils.MAX_SUBJECT_WAIT, subjectsFile)

	return subjects, nil
}

func fetchTermsList() ([]string, error) {

}

func filterTerms(terms []string) ([]string, string, error) {

}

func getCourses(subject, term, year string) ([]models.Course, error) {

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

	subjects, err := fetchSubjectList()
	if err != nil {
		return fmt.Errorf("failed to fetch subjects list: %w")
	}

	terms, err := fetchTermsList()
	if err != nil {
		return fmt.Errorf("failed to fetch terms list: %w")
	}

	terms, year, err := filterTerms(terms)
	if err != nil {
		return fmt.Errorf("failed to filter terms list: %w")
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
