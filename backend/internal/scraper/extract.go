package scraper

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/cwooper/schedule-optimizer/internal/models"
	"github.com/cwooper/schedule-optimizer/internal/utils"
)

var (
	httpClient  *http.Client
	timePattern = regexp.MustCompile(`(\d{1,2}):(\d{2})\s*(AM|PM)?`)
)

func init() {
	httpClient = &http.Client{
		Timeout: time.Minute * 5,
	}
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

// Extracts the header from the string parts and the given rows
func extractHeader(parts []string, rows *goquery.Selection) (*models.Course, error) {
	subject := strings.TrimSpace(parts[0])
	title := strings.TrimSpace(parts[1])
	credits := strings.TrimSpace(strings.ReplaceAll(parts[2], "cr", ""))

	prerequisites := ""
	for i := 1; i < rows.Length(); i++ {
		prerequisites += strings.TrimSpace(rows.Eq(i).Text())
		if i != rows.Length()-1 { // Add spaces between rows
			prerequisites += " "
		}
	}

	return &models.Course{
		Subject:       subject,
		Title:         title,
		Credits:       credits,
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
	// waitlist := strings.TrimSpace(cells.Eq(10).Text()) // Unused
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
	nums, err := convertMultiple(crn, capacity, enrolled, available)
	if err != nil {
		return fmt.Errorf("error converting course numbers: %w", err)
	}

	course.CRN = nums[0]
	course.Sessions = append(course.Sessions, session)
	course.AdditionalFees = normalizeSpaces(additionalFees)
	course.Capacity = nums[1]
	course.Enrolled = nums[2]
	course.AvailableSeats = nums[3]
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
		Days:       days,
		StartTime:  startTime,
		EndTime:    endTime,
		Instructor: instructor,
		Location:   location,
		IsAsync:    isAsync,
		IsTimeTBD:  isTimeTBD,
	}

	course.Sessions = append(course.Sessions, session)

	return nil
}

// Gets the courses from the given table
//
//		Format of a course table:
//		<table> Header Information </table> (if fieldformatboldtext in <td>)
//		<div></div>
//		<table> Course Information <table> (if 13 <td> elements in <tr>)
//		<table> Session Information <table> (if 12 <td> elements in <tr>)
//		The above two can repeat in any order
//	 The above then repeats any number of times
func getCoursesFromTables(doc *goquery.Document) ([]models.Course, error) {
	lastCourse := models.NewCourse()
	var courses []models.Course
	var err error

	// Iterate over every table
	doc.Find("table").Each(func(i int, tableSelection *goquery.Selection) {
		rows := tableSelection.Find("tr") // Used in the header only
		if strings.Contains(rows.Find("td").First().Text(), "Term") {
			return // This is an info row
		}

		boldTextCells := tableSelection.Find("td.fieldformatboldtext")
		if boldTextCells.Length() > 0 {
			headerText := strings.TrimSpace(boldTextCells.Text())

			// Extract the parts of the header's td.fieldformatboldtext
			title := boldTextCells.Find("a").Text()
			if title != "" {
				beforeAfter := strings.SplitN(headerText, title, 2)
				// This is a new header, process the previous course if exists
				if lastCourse.Sessions != nil {
					courses = append(courses, *lastCourse)
				}
				// Extract new header

				parts := []string{beforeAfter[0], title, beforeAfter[1]}
				lastCourse, err = extractHeader(parts, rows)
				if err != nil {
					fmt.Printf("error extracting header: %v\n", err)
					return
				}
			}
		} else {
			// This table might contain course or session information
			tableSelection.Find("tr").Each(func(j int, row *goquery.Selection) {
				cells := row.Find("td")
				cellCount := cells.Length()

				if cellCount == 13 {
					// This is a course row
					if lastCourse.Sessions != nil {
						// Append the course if it's a main session
						courses = append(courses, *lastCourse)
					}

					// Reset the last course for the next one
					if lastCourse.Sessions != nil {
						lastCourse = &models.Course{
							Subject:       lastCourse.Subject,
							Title:         lastCourse.Title,
							Credits:       lastCourse.Credits,
							Prerequisites: lastCourse.Prerequisites,
						}
					}

					err = extractCourse(cells, lastCourse)
					if err != nil {
						fmt.Printf("error extracting course %s: %v\n", lastCourse.Subject, err)
						return
					}
				} else if cellCount == 12 && lastCourse.Sessions != nil {
					err = extractSession(cells, lastCourse)
					if err != nil {
						fmt.Printf("error extracting session: %v\n", err)
						return
					}
				}
			})
		}
	})

	// Append the last course if it exists
	if lastCourse.Sessions != nil {
		courses = append(courses, *lastCourse)
	}

	return courses, nil
}

// Formatted as "Subject|Title|Professor(s)|..."
func courseToCourseString(course models.Course) string {
	var instructors []string

	for _, session := range course.Sessions {
		instructors = append(instructors, session.Instructor)
	}

	instructorString := strings.Join(instructors, "|")
	allStrings := []string{course.Subject, course.Title, instructorString}

	result := strings.Join(allStrings, "|")

	return strings.ToUpper(result)
}

// Returns all of the courses found at the given url specified
// by the subject, term and year
func getSubjectFromURL(subject, term, year string) ([]models.Course, error) {
	// Create url.Values to hold the payload
	data := url.Values{}
	data.Set("term", term)
	data.Set("curr_yr", year)
	data.Set("subj", subject)

	// Create a new request with the encoded form data
	encodedData := data.Encode()
	req, err := http.NewRequest("POST", utils.URL, strings.NewReader(encodedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the correct Content-Type for form data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: status code %d accessing %s with payload %v",
			resp.StatusCode, utils.URL, encodedData)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Get the course tables from the response body
	courses, err := getCoursesFromTables(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses from tables: %w", err)
	}

	for i := range courses {
		courses[i].CourseString = courseToCourseString(courses[i])
	}

	return courses, nil
}
