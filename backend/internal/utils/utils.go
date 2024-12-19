// Package utils for "modular" hard-coding and re-used functions
package utils

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"

	"google.golang.org/protobuf/proto"

	pb "github.com/cwooper/schedule-optimizer/internal/proto/generated"
)

const (
	ROUND       = 2    // Round to x decimal places
	MINS_IN_DAY = 1440 // Number of minutes in a day (60 * 24)

	TIME_FORMAT = "2006-01-02 15:04:05"

	URL        = "https://web4u.banner.wwu.edu/pls/wwis/wwskcfnd.TimeTable"
	SUBJECT_ID = "subj" // HTML select id name
	TERM_ID    = "term" // HTML term id name
)

// Returns the location of the data directory
var DataDirectory string

// Finds and initializes the data directory
func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	DataDirectory = filepath.Join(wd, "../data")

	// Ensure that the directory exists
	if err := os.MkdirAll(DataDirectory, os.ModePerm); err != nil {
		panic(err)
	}
}

var (
	// Returns first and last name with a comma (e.g. "Smith, John")
	CommaNameRegexp = regexp.MustCompile(`^.*?(\S+)\s*,\s*(\S+).*$`)
	// Returns first and last name listed normally (e.g. "John Smith")
	PlainNameRegexp = regexp.MustCompile(`^(\S+)(?:.*\s)?(\S+)$`)
	// Returns the Course Subject and Number Associated (e.g. "CSCI 497A")
	SubjectRegexp = regexp.MustCompile(`^(\S+)\s+(\d+[A-Z]?)$`)
)

// Helper function to round float64 to n decimal places
func Round(x float64) float64 {
	pow := math.Pow(10, float64(ROUND))
	return math.Round(x*pow) / pow
}

// Helper function to load protobuf
func LoadCoursesProtobuf(filePath string) (*pb.CourseList, error) {
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
func SaveCoursesProtobuf(protobuf *pb.CourseList, filename string) error {
	data, err := proto.Marshal(protobuf)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write protobuf to file: %w", err)
	}

	return nil
}

// Self explanatory
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Writes the pid to a tmp file
func WritePidFile() error {
	// Get the process ID
	pid := os.Getpid()

	// Try writing to /tmp first
	pidFile := filepath.Join(os.TempDir(), "schedule-optimizer.pid")
	err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
	if err != nil {
		// Fallback to current directory if /tmp fails
		pidFile = "schedule-optimizer.pid"
		err = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
		if err != nil {
			return fmt.Errorf("failed to write PID file: %w", err)
		}
	}

	// Log where we wrote the PID
	log.Printf("PID %d written to %s", pid, pidFile)

	// Clean up PID file on exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if err := os.Remove(pidFile); err != nil {
			log.Printf("Failed to remove PID file: %v", err)
		}
		os.Exit(0)
	}()

	return nil
}

// The below constants and variables are used for the GPA spreadsheet
const (
	CODE_COL = iota
	TERM_COl
	CRN_COL
	TITLE_COL
	PROFESSOR_COL
	ENROLLED_COL
	GRADE_COUNT_COL
	CNT_A_COL // Grade columns start here
	CNT_AM_COL
	CNT_BP_COL
	CNT_B_COL
	CNT_BM_COL
	CNT_CP_COL
	CNT_C_COL
	CNT_CM_COL
	CNT_DP_COL
	CNT_D_COL
	CNT_DM_COL
	CNT_F_COL
	CNT_W_COL // We don't include courses dropped
)

var (
	CourseSubjectMapping = map[string]string{
		"AHE":  "CFPA",
		"ASLC": "SPED",
		"ARAB": "LANG",
		"ASTR": "PHYS",
		"BNS":  "PSY",
		"BUS":  "ACCT",
		"CHIN": "LANG",
		"CLST": "LANG",
		"CD":   "HHD",
		"CSEC": "CSE",
		"C2C":  "HCS",
		"CISS": "CSCI",
		"DNC":  "THTR",
		"DATA": "CSCI",
		"DIAD": "EDUC",
		"ECE":  "ELED",
		"EDAD": "SPED",
		"ESJ":  "SEC",
		"EECE": "ENGD",
		"ENGR": "ENGD",
		"EUS":  "LANG",
		"FIN":  "FMKT",
		"FREN": "LANG",
		"GERM": "LANG",
		"GLBL": "ESCI",
		"GREK": "LANG",
		"HLED": "HHD",
		"HRM":  "MGMT",
		"HSP":  "HCS",
		"HUMA": "GHR",
		"ID":   "ENGD",
		"I T":  "ECEM",
		"IEP":  "LANG",
		"IBUS": "MGMT",
		"ITAL": "LANG",
		"JAPN": "LANG",
		"KIN":  "HHD",
		"LAT":  "LANG",
		"MIS":  "DSCI",
		"MFGE": "ENGD",
		"MKTG": "FMKT",
		"MPAC": "ACCT",
		"M/CS": "MATH",
		"MLE":  "ECEM",
		"NURS": "HCS",
		"OPS":  "MBA",
		"PA":   "HHD",
		"PE":   "HHD",
		"PEH":  "HHD",
		"PME":  "ENGD",
		"PORT": "LANG",
		"RECR": "HHD",
		"RC":   "HCS",
		"REL":  "LBRL",
		"RUSS": "LANG",
		"SPAN": "LANG",
		"SUST": "UEPP",
		"TEOP": "ELIT",
		"TESL": "ELED",
	}

	GpaMap = map[int]float64{
		CNT_A_COL:  4.0,
		CNT_AM_COL: 3.7,
		CNT_BP_COL: 3.3,
		CNT_B_COL:  3.0,
		CNT_BM_COL: 2.7,
		CNT_CP_COL: 2.3,
		CNT_C_COL:  2.0,
		CNT_CM_COL: 1.7,
		CNT_DP_COL: 1.3,
		CNT_D_COL:  1.0,
		CNT_DM_COL: 0.7,
		CNT_F_COL:  0.0,
	}
)
