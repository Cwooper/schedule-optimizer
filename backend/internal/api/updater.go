package api

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/cwooper/schedule-optimizer/internal/cache"
	"github.com/cwooper/schedule-optimizer/internal/gpa"
	"github.com/cwooper/schedule-optimizer/internal/models"
	"github.com/cwooper/schedule-optimizer/internal/utils"
	"github.com/cwooper/schedule-optimizer/pkg/protoutils"
)

// Reads lines from a file
func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return lines, nil
}

// Writes lines to a file with timestamp
func writeLines(lines []string, filename string) error {
	// Prepend current timestamp
	currentTime := time.Now().Format(utils.TIME_FORMAT)
	allLines := append([]string{currentTime}, lines...)

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range allLines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write line: %w", err)
		}
	}

	return writer.Flush()
}

// Gets subjects either from file or API
// This is no longer used/needed for the API client. However, it the subjects
// are still needed for the frontend, so it is left here for now.
func getSubjects(client *Client) ([]string, error) {
	subjectsFile := filepath.Join(utils.DataDirectory, "subjects.txt")

	// Try to read from file first
	if lines, err := readLines(subjectsFile); err == nil && len(lines) > 1 {
		fileTime, err := time.Parse(utils.TIME_FORMAT, lines[0])
		if err == nil && time.Since(fileTime) < utils.MAX_SUBJECT_WAIT*24*time.Hour {
			return lines[1:], nil
		}
	}

	// If file doesn't exist or is too old, fetch from API
	subjectResp, err := client.GetSubjects("202520") // Use current term
	if err != nil {
		return nil, fmt.Errorf("failed to get subjects from API: %w", err)
	}

	// Extract subject codes
	var subjects []string
	for _, subject := range subjectResp {
		subjects = append(subjects, subject.Code)
	}

	// Save to file
	if err := writeLines(subjects, subjectsFile); err != nil {
		log.Printf("Warning: Failed to save subjects to file: %v", err)
	}

	return subjects, nil
}

// Gets terms either from file or API
func getTerms(client *Client) ([]string, error) {
	termsFile := filepath.Join(utils.DataDirectory, "terms.txt")

	// Try to read from file first
	if lines, err := readLines(termsFile); err == nil && len(lines) > 1 {
		fileTime, err := time.Parse(utils.TIME_FORMAT, lines[0])
		if err == nil && time.Since(fileTime) < utils.MAX_TERM_WAIT*24*time.Hour {
			return lines[1:], nil
		}
	}

	// If file doesn't exist or is too old, fetch from API
	termResp, err := client.GetTerms()
	if err != nil {
		return nil, fmt.Errorf("failed to get terms from API: %w", err)
	}

	// Extract term codes
	var terms []string
	for _, term := range termResp {
		terms = append(terms, term.Code)
	}

	// Save to file
	if err := writeLines(terms, termsFile); err != nil {
		log.Printf("Warning: Failed to save terms to file: %v", err)
	}

	return terms, nil
}

// Gets all courses for a specific term
func getTermCourses(client *Client, term string) ([]models.Course, error) {
	termFile := filepath.Join(utils.DataDirectory, term+".pb")

	// Check if we have recent data
	if _, err := os.Stat(termFile); err == nil {
		existingProto, err := utils.LoadCoursesProtobuf(termFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load protobuf: %w", err)
		}

		pullTime := existingProto.PullTimestamp.AsTime()
		if time.Since(pullTime) < utils.MAX_NEW_COURSE_WAIT*24*time.Hour {
			return protoutils.ProtoToCourses(existingProto), nil
		}
	}

	// Fetch all courses for the term
	courses, err := client.GetCourses(term, "%", "%")
	if err != nil {
		return nil, fmt.Errorf("failed to get courses for term %s: %w", term, err)
	}

	// Generate GPA data
	if err := gpa.GenerateGPA(&courses); err != nil {
		return nil, fmt.Errorf("failed to generate GPA data: %w", err)
	}

	// Save to protobuf
	proto := protoutils.CoursesToProto(courses)
	proto.PullTimestamp = timestamppb.Now()
	if err := utils.SaveCoursesProtobuf(proto, termFile); err != nil {
		return nil, fmt.Errorf("failed to save protobuf: %w", err)
	}

	return courses, nil
}

// UpdateCourses updates all course data as needed
func UpdateCourses() error {
	// Create data directory if needed
	if err := os.MkdirAll(utils.DataDirectory, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create API client
	client, err := NewClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Get subjects
	// This is no longer used/needed for the API client. However, it the subjects
	// are still needed for the frontend, so it is left here for now.
	_, err = getSubjects(client)
	if err != nil {
		return fmt.Errorf("failed to get subjects: %w", err)
	}

	// Get terms
	terms, err := getTerms(client)
	if err != nil {
		return fmt.Errorf("failed to get terms: %w", err)
	}

	// Filter terms to only include relevant ones
	terms, _, err = filterTerms(terms)
	if err != nil {
		return fmt.Errorf("failed to filter terms: %w", err)
	}

	log.Printf("Processing terms: %v...\n", terms)
	start := time.Now()
	totalCourses := 0

	// Process each term
	for _, term := range terms {
		termInfo, err := ParseTermCode(term)
		if err != nil {
			log.Printf("Warning: Invalid term code %s: %v", term, err)
			continue
		}

		if !shouldUpdateTerm(termInfo) {
			log.Printf("Skipping term %s - data is current", term)
			continue
		}

		courses, err := getTermCourses(client, term)
		if err != nil {
			return fmt.Errorf("failed to get courses for term %s: %w", term, err)
		}

		totalCourses += len(courses)
		log.Printf("%s: Found and saved %d courses\n", term, len(courses))
	}

	duration := time.Since(start)
	log.Printf("Update completed in %v, processed %d courses", duration, totalCourses)

	// Preload cache with updated data
	if len(terms) > 0 {
		courseManager := cache.GetInstance()
		courseManager.PreloadCache(terms)
	}

	return nil
}
