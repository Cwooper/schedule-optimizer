package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"schedule-optimizer/internal/models"
	"schedule-optimizer/internal/utils"
)

var (
	httpClient  *http.Client
	timePattern = regexp.MustCompile(`(\d{1,2}):(\d{2})\s*(AM|PM)?`)
)

func init() {
	httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
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

// convertTimes converts times format from "09:00-01:50 pm" to 900, 1350
func convertTimes(times string) (int, int, error) {
	timeParts := strings.Split(times, " ")
	if len(timeParts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format: %s", times)
	}

	amPm := timeParts[1]
	timeStartEnd := strings.Split(timeParts[0], "-")
	if len(timeStartEnd) != 2 {
		return 0, 0, fmt.Errorf("invalid time range format: %s", timeParts[0])
	}

	startTime := strings.ReplaceAll(timeStartEnd[0], ":", "")
	endTime := strings.ReplaceAll(timeStartEnd[1], ":", "")

	startTimeInt, err := strconv.Atoi(startTime)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start time: %s", startTime)
	}

	endTimeInt, err := strconv.Atoi(endTime)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end time: %s", endTime)
	}

	// Convert times to military time
	if strings.ToLower(amPm) == "pm" {
		if endTimeInt < 1200 {
			endTimeInt += 1200
		}
		// If a class starts before 8, it must be a pm, unless
		// it's bigger than the end_time when making it a pm
		if startTimeInt < 800 && startTimeInt+1200 < endTimeInt {
			startTimeInt += 1200
		}
	}

	return startTimeInt, endTimeInt, nil
}

// Extracts the header from the string parts and the given cell
func extractHeader(parts []string, cells *goquery.Selection) (*models.Course, error) {
	subject := strings.TrimSpace(parts[0])
	title := strings.TrimSpace(parts[0])
	credits := strings.TrimSpace(strings.ReplaceAll(parts[2], "cr", ""))
	creditsInt, err := strconv.Atoi(credits)
	if err != nil {
		return nil, fmt.Errorf("failed to parse string to int: %w", err)
	}
	prerequisites := ""
	if cells.Length() > 1 {
		prerequisites = strings.TrimSpace(cells.Eq(1).Text())
	}

	return &models.Course{
		Subject:       subject,
		Title:         title,
		Credits:       creditsInt,
		Prerequisites: prerequisites,
	}, nil
}

// Extracts a course from the cells and fills them into the provided course
func extractCourse(cells *goquery.Selection, course *models.Course) error {
	// quarter := strings.TrimSpace(cells.Eq(0).Text()) // Unused
	crn := strings.TrimSpace(cells.Eq(1).Text())
	days := strings.TrimSpace(cells.Eq(2).Text())
	times := strings.TrimSpace(cells.Eq(3).Text())
	instructor := strings.TrimSpace(cells.Eq(4).Text())
	location := strings.TrimSpace(cells.Eq(5).Text())
	additionalFees := strings.TrimSpace(cells.Eq(6).Text())
	capacity := strings.TrimSpace(cells.Eq(7).Text())
	enrolled := strings.TrimSpace(cells.Eq(8).Text())
	available := strings.TrimSpace(cells.Eq(9).Text())
	waitlist := strings.TrimSpace(cells.Eq(10).Text())
	restrictions := strings.TrimSpace(cells.Eq(11).Text())
	attributes := strings.TrimSpace(cells.Eq(12).Text())

	var startTime, endTime int
	isAsync := false
	isTimeTBD := false
	if times != "" && timePattern.MatchString(times) {
		// manually declare error to not overwrite startTime and endTime
		var err error
		startTime, endTime, err = convertTimes(times)
		if err != nil {
			return fmt.Errorf("failed to parse %v: %w", times, err)
		}
	} else if times == "N/A" {
		isAsync = true
	} else if times == "TBD" {
		isTimeTBD = true
	}

	// Fill the session block
	session := models.Session{
		Days:       days,
		StartTime:  startTime,
		EndTime:    endTime,
		Instructor: instructor,
		Location:   location,
		IsAsync:    isAsync,
		IsTimeTBD:  isTimeTBD,
	}

	// Convert all of the needed number strings to ints
	nums, err := convertMultiple(crn, capacity, enrolled, available, waitlist)
	if err != nil {
		return fmt.Errorf("error converting course numbers: %w", err)
	}

	course.CRN = nums[0]
	course.Sessions = append(course.Sessions, session)
	course.AdditionalFees = additionalFees
	course.Capacity = nums[1]
	course.Capacity = nums[1]
	course.Enrolled = nums[2]
	course.AvailableSeats = nums[3]
	course.WaitlistCount = nums[4]
	course.Restrictions = restrictions
	course.Attributes = attributes

	return nil
}

// Extracts a session from the cells and fills them into the provided course
func extractSession(cells *goquery.Selection, course *models.Course) error {
	days := strings.TrimSpace(cells.Eq(2).Text())
	times := strings.TrimSpace(cells.Eq(3).Text())
	instructor := strings.TrimSpace(cells.Eq(4).Text())
	location := strings.TrimSpace(cells.Eq(5).Text())
	// All other data does not exist for labs
	
	var startTime, endTime int
	isAsync := false
	isTimeTBD := false
	if times != "" && timePattern.MatchString(times) {
		// manually declare error to not overwrite startTime and endTime
		var err error
		startTime, endTime, err = convertTimes(times)
		if err != nil {
			return fmt.Errorf("failed to parse %v: %w", times, err)
		}
	} else if times == "N/A" {
		isAsync = true
	} else if times == "TBD" {
		isTimeTBD = true
	}

	session := models.Session{
		Days: days,
		StartTime: startTime,
		EndTime: endTime,
		Instructor: instructor,
		Location: location,
		IsAsync: isAsync,
		IsTimeTBD: isTimeTBD,
	}

	course.Sessions = append(course.Sessions, session)

	return nil
}

// Gets the courses from the given table
func getCoursesFromTable(table *goquery.Selection) ([]models.Course, error) {
	var courses []models.Course
	var lastCourse *models.Course
	var err error

	rows := table.Find("tr")
	for i := 0; i < rows.Length(); i++ {
		row := rows.Eq(i)
		cells := row.Find("td")
		cellCount := cells.Length()

		switch cellCount {
		case 1:
			// Header row
			headerText := cells.Eq(0).Text()
			parts := strings.Split(headerText, "|")
			if len(parts) != 3 {
				continue
			}
			lastCourse, err = extractHeader(parts, cells)
			if err != nil {
				return nil, fmt.Errorf("error extracting header: %w", err)
			}
		case 13:
			// Main course row
			if lastCourse != nil {
				courses = append(courses, *lastCourse)
			}
			err = extractCourse(cells, lastCourse)
			if err != nil {
				return nil, fmt.Errorf("error extracting course: %w", err)
			}
		case 12:
			// Extra session row
			if lastCourse != nil {
				err = extractSession(cells, lastCourse)
				if err != nil {
					return nil, fmt.Errorf("error extracting lab: %w", err)
				}
			}
		}
	}

	if lastCourse != nil {
		courses = append(courses, *lastCourse)
	}

	return courses, nil
}

// GetCourseTables finds all course-like tables and returns them
func getCourseTables(doc *goquery.Document) []*goquery.Selection {
	var courseTables []*goquery.Selection
	tables := doc.Find("table")

	tables.Each(func(i int, table *goquery.Selection) {
		// Find all table cells as defined by "fieldformatboldtext"
		rows := table.Find("tr")
		rows.Each(func(j int, row *goquery.Selection) {
			cells := row.Find("td.fieldformatboldtext")

			// If we found such cells, this table contains course info
			if cells.Length() > 0 {
				courseTables = append(courseTables, table)
				return
			}
		})
	})

	return courseTables
}

// Returns all of the courses found at the given url specified
// by the subject, term and year
func getSubjectFromURL(subject, term, year string) ([]models.Course, error) {
	payload := map[string]string{
		"term":    term,
		"curr_yr": year,
		"subj":    subject,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", utils.URL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: status code %d accessing %s", resp.StatusCode, utils.URL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	allCourses := make([]models.Course, 0)

	// Get the course tables from the response body
	tables := getCourseTables(doc)
	for _, table := range tables {
		courses, err := getCoursesFromTable(table)
		if err != nil {
			return nil, fmt.Errorf("failed to get course from table: %w", err)
		}

		allCourses = append(allCourses, courses...)
	}

	return allCourses, nil
}
