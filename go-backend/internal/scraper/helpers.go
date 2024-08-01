package scraper

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"google.golang.org/protobuf/proto"
	pb "schedule-optimizer/internal/proto/generated"

	"schedule-optimizer/internal/utils"
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

// Helper function to load protobuf
func loadProtobuf(filePath string) (*pb.CourseList, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read protobuf file: %w", err)
	}

	var courseList pb.CourseList
	if err := proto.Unmarshal(data, &courseList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}

	return &courseList, nil
}

// Helper to save a protobuf
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
