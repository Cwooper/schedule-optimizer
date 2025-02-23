package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/cwooper/schedule-optimizer/internal/models"
)

const (
	baseURL          = "https://registration.banner.wwu.edu/StudentRegistrationSsb/ssb"
	termSelectionURL = baseURL + "/term/termSelection?mode=search"
	termSearchURL    = baseURL + "/term/search"
	courseSearchURL  = baseURL + "/searchResults/searchResults"
	pageSize         = 500
)

// Client handles API requests and maintains session state
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new API client with cookie support
func NewClient() (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: time.Duration(5) * time.Minute,
			Jar:     jar,
		},
	}, nil
}

// initializeSession sets up the initial session and selects the term
func (c *Client) initializeSession(term string) error {
	// First, visit the initial page to get cookies
	resp, err := c.httpClient.Get(termSelectionURL)
	if err != nil {
		return fmt.Errorf("failed to initialize session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to initialize session, status: %d", resp.StatusCode)
	}

	// Then submit the term selection
	data := url.Values{
		"term":            {term},
		"studyPath":       {""},
		"studyPathText":   {""},
		"startDatepicker": {""},
		"endDatepicker":   {""},
	}

	resp, err = c.httpClient.PostForm(termSearchURL, data)
	if err != nil {
		return fmt.Errorf("failed to select term: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to select term, status: %d", resp.StatusCode)
	}

	// Let the server process the term selection
	time.Sleep(time.Second)
	return nil
}

// GetCourses retrieves all courses for a given term and subject
func (c *Client) GetCourses(term, subject, courseNum string) ([]models.Course, error) {
	// Initialize the session with the term
	if err := c.initializeSession(term); err != nil {
		return nil, err
	}

	// Escape special characters in the subject
	subject = strings.ReplaceAll(subject, "%", "%25")
	courseNum = strings.ReplaceAll(courseNum, "%", "%25")

	var allCourses []models.Course
	pageOffset := 0

	for {
		// Build the search URL with parameters
		params := url.Values{
			"txt_term":         {term},
			"txt_subject":      {subject},
			"txt_courseNumber": {courseNum},
			"pageOffset":       {fmt.Sprintf("%d", pageOffset)},
			"pageMaxSize":      {fmt.Sprintf("%d", pageSize)},
		}

		// Create request
		req, err := http.NewRequest("GET", courseSearchURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Add query parameters to URL
		req.URL.RawQuery = params.Encode()

		// Add headers
		req.Header.Set("Accept", "application/json")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		// Make the request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch courses: %w", err)
		}

		// Read and parse the response
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		var apiResp APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		// Check if we have any data
		if len(apiResp.Data) == 0 {
			break
		}

		// Convert API courses to our model
		for _, courseData := range apiResp.Data {
			course, err := apiToCourse(courseData)
			if err != nil {
				return nil, fmt.Errorf("failed to convert course: %w", err)
			}
			allCourses = append(allCourses, *course)
		}

		// If we received fewer courses than the page size, we're done
		if len(apiResp.Data) < pageSize {
			break
		}

		// Move to next page
		pageOffset += pageSize

		// Add a small delay to avoid overwhelming the server
		time.Sleep(500 * time.Millisecond)
	}

	return allCourses, nil
}
