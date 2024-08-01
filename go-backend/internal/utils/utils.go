// Package utils for "modular" hard-coding and re-used functions
package utils

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/proto"

	pb "schedule-optimizer/internal/proto/generated"
)

const (
	ROUND       = 2    // Round to x decimal places
	MINS_IN_DAY = 1440 // Number of minutes in a day (60 * 24)

	MIN_COURSES = 1 // Minimum courses in a single schedule
	MAX_COURSES = 7 // Maximum courses in a single schedule

	MAX_SUBJECT_WAIT = 30 // Days
	MAX_COURSE_WAIT  = 2  // Days
	MAX_TERM_WAIT    = 10 // Days

	TIME_FORMAT = "2006-01-02 15:04:05"

	URL        = "https://web4u.banner.wwu.edu/pls/wwis/wwskcfnd.TimeTable"
	SUBJECT_ID = "subj" // HTML select id name
	TERM_ID    = "term" // HTML term id name
)

var DataDirectory string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	DataDirectory = filepath.Join(wd, "data")

	// Ensure that the directory exists
	if err := os.MkdirAll(DataDirectory, os.ModePerm); err != nil {
		panic(err)
	}
}

// Helper function to round float64 to n decimal places
func Round(x float64) float64 {
	pow := math.Pow(10, float64(ROUND))
	return math.Round(x*pow) / pow
}

// Helper function to load protobuf
func LoadProtobuf(filePath string) (*pb.CourseList, error) {
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
func SaveProtobuf(protobuf *pb.CourseList, filename string) error {
	data, err := proto.Marshal(protobuf)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write protobuf to file: %w", err)
	}

	return nil
}
