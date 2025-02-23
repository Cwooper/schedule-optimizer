package api

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/cwooper/schedule-optimizer/internal/utils"
)

// Quarter Codes
const (
	FALL   = 40
	WINTER = 10
	SUMMER = 30
	SPRING = 20
)

// TermInfo represents information about a term
type TermInfo struct {
	Code     string    // e.g., "202410"
	AcadYear string    // e.g., "2324"
	Quarter  int       // 40=Fall, 10=Winter, 20=Spring, 30=Summer
	LastPull time.Time // When we last successfully pulled this term
	HasProto bool      // Whether we have a protobuf for this term
}

// GetCurrentAcadYear returns the current academic year (e.g., "2324")
func GetCurrentAcadYear() string {
	now := time.Now()
	year := now.Year()
	month := now.Month()

	// If we're before September, we're in the previous academic year
	if month < time.September {
		year--
	}

	// Formats the aid year as "2324"
	return fmt.Sprintf("%02d%02d", year%100, (year+1)%100)
}

// ParseTermCode breaks down a term code into components
func ParseTermCode(code string) (TermInfo, error) {
	if len(code) != 6 {
		return TermInfo{}, fmt.Errorf("invalid term code length: %s", code)
	}

	year, err := strconv.Atoi(code[:4])
	if err != nil {
		return TermInfo{}, fmt.Errorf("invalid year in term code: %s", code)
	}

	quarter, err := strconv.Atoi(code[4:])
	if err != nil {
		return TermInfo{}, fmt.Errorf("invalid quarter in term code: %s", code)
	}

	// Calculate academic year based on quarter
	var acadYear string
	switch quarter {
	case SUMMER, WINTER, SPRING: // Previous aid year
		acadYear = fmt.Sprintf("%02d%02d", (year-1)%100, year%100)
	case FALL: // Aid year starts in fall
		acadYear = fmt.Sprintf("%02d%02d", year%100, (year+1)%100)
	default:
		return TermInfo{}, fmt.Errorf("invalid quarter code: %d", quarter)
	}

	// Check if we have a protobuf for this term
	protoPath := filepath.Join(utils.DataDirectory, code+".pb")
	hasProto := utils.FileExists(protoPath)

	var lastPull time.Time
	if hasProto {
		proto, err := utils.LoadCoursesProtobuf(protoPath)
		if err == nil && proto.PullTimestamp != nil {
			lastPull = proto.PullTimestamp.AsTime()
		}
	}

	return TermInfo{
		Code:     code,
		AcadYear: acadYear,
		Quarter:  quarter,
		LastPull: lastPull,
		HasProto: hasProto,
	}, nil
}

// filterTerms returns terms that should be included in our data
//
// Returns filteredTerms, year, error
func filterTerms(terms []string) ([]string, string, error) {
	currentAcadYear := GetCurrentAcadYear()

	// Track the highest year for the "year" return parameter
	var highestYear int

	// Map to track unique terms we want to keep
	validTerms := make(map[string]TermInfo)

	for _, term := range terms {
		// Skip non-term entries
		if term == "All" || len(term) != 6 {
			continue
		}

		termInfo, err := ParseTermCode(term)
		if err != nil {
			continue
		}

		yearNum, _ := strconv.Atoi(term[:4])
		if yearNum > highestYear {
			highestYear = yearNum
		}

		// Include terms if:
		// 1. They're in current or next academic year
		// 2. They have existing data and are within last 2 academic years
		if termInfo.AcadYear >= currentAcadYear ||
			(termInfo.HasProto && termInfo.AcadYear >= fmt.Sprintf("%02d%02d",
				(highestYear-2)%100, (highestYear-1)%100)) {
			validTerms[term] = termInfo
		}
	}

	if len(validTerms) == 0 {
		return nil, "", fmt.Errorf("no valid terms found")
	}

	// Convert map keys to sorted slice
	var filteredTerms []string
	for term := range validTerms {
		filteredTerms = append(filteredTerms, term)
	}
	sort.Strings(filteredTerms)

	// Generate year string for the API
	yearHigh := highestYear % 100
	yearLow := yearHigh - 1
	year := fmt.Sprintf("%02d%02d", yearLow, yearHigh)

	return filteredTerms, year, nil
}

// shouldUpdateTerm determines if a term needs updating based on:
// 1. How old our data is
// 2. Whether it's a current/future term (update more frequently)
// 3. Whether it's a past term (update rarely/never)
func shouldUpdateTerm(termInfo TermInfo) bool {
	if !termInfo.HasProto {
		return true // Always update if we don't have data
	}

	currentAcadYear := GetCurrentAcadYear()
	now := time.Now()

	// For current/future terms
	if termInfo.AcadYear >= currentAcadYear {
		// Update if data is older than util const
		return now.Sub(termInfo.LastPull) > utils.MAX_NEW_COURSE_WAIT*24*time.Hour
	}

	// For past terms
	// Update if data is older than util const
	return now.Sub(termInfo.LastPull) > utils.MAX_OLD_COURSE_WAIT*24*time.Hour
}
