package scraper

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"schedule-optimizer/internal/models"
	"schedule-optimizer/internal/utils"
)

// Helper function for printing a list of courses.
func printCourseList(courseList []models.Course) {
	for _, course := range courseList {
		fmt.Printf("%v\n", course)
	}
	fmt.Println()
}

// Normalizes the spaces inside of a string by replacing multiple ws with one
func normalizeSpaces(s string) string {
	// First, trim leading and trailing spaces
	s = strings.TrimSpace(s)

	// Then, replace multiple spaces with a single space
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(s, " ")
}

// convertMultiple converts multiple strings to ints
// It returns the converted ints and an error if any conversion failed
func convertMultiple(strs ...string) ([]int, error) {
	result := make([]int, len(strs))
	for i, s := range strs {
		if s == "" {
			continue // Skip empty strings, leaving 0 in the result
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("failed to convert '%s' to int: %w", s, err)
		}
		result[i] = n
	}
	return result, nil
}

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
