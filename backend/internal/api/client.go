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
	BASE_URL  = "https://registration.banner.wwu.edu/StudentRegistrationSsb/ssb"
	TERMS_URL = "https://registration.banner.wwu.edu/StudentRegistrationSsb/ssb/classSearch/getTerms"
	SUBJ_URL  = "https://registration.banner.wwu.edu/StudentRegistrationSsb/ssb/classSearch/get_subject"

	termSelectionURL = BASE_URL + "/term/termSelection?mode=search"
	termSearchURL    = BASE_URL + "/term/search"
	courseSearchURL  = BASE_URL + "/searchResults/searchResults"
	pageSize         = 500
)

// Client handles API requests and maintains session state
type Client struct {
	httpClient *http.Client
	baseURL    string
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
		baseURL: BASE_URL,
	}, nil
}

// initializeSession sets up the initial session and selects the term
func (c *Client) initializeSession(term string) error {
	// First, visit the initial page to get cookies
	termSelectionURL := c.baseURL + "/term/termSelection?mode=search"
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

	termSearchURL := c.baseURL + "/term/search"
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

// GetTerms retrieves available terms from the API
func (c *Client) GetTerms() ([]TermResponse, error) {
	// Create request
	req, err := http.NewRequest("GET", TERMS_URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("searchTerm", "")
	q.Add("offset", "1")
	q.Add("max", "10")
	req.URL.RawQuery = q.Encode()

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get terms: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse response
	var terms []TermResponse
	if err := json.NewDecoder(resp.Body).Decode(&terms); err != nil {
		return nil, fmt.Errorf("failed to parse terms response: %w", err)
	}

	return terms, nil
}

// GetSubjects retrieves available subjects for a term from the API
func (c *Client) GetSubjects(term string) ([]SubjectResponse, error) {
	// Create request
	req, err := http.NewRequest("GET", SUBJ_URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("searchTerm", "")
	q.Add("term", term)
	q.Add("offset", "1")
	q.Add("max", "500")
	req.URL.RawQuery = q.Encode()

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get subjects: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse response
	var subjects []SubjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&subjects); err != nil {
		return nil, fmt.Errorf("failed to parse subjects response: %w", err)
	}

	return subjects, nil
}

// GetCourses retrieves all courses for a given term and subject
func (c *Client) GetCourses(term, subject, courseNum string) ([]models.Course, error) {
	// Initialize the session with the term
	if err := c.initializeSession(term); err != nil {
		return nil, err
	}

	// Trim space (just in case) and escape special characters
	term = strings.TrimSpace(term) // term does not need escaping
	subject = strings.TrimSpace(strings.ReplaceAll(subject, "%", "%25"))
	courseNum = strings.TrimSpace(strings.ReplaceAll(courseNum, "%", "%25"))

	var allCourses []models.Course
	pageOffset := 0

	for {
		// Build the search URL with parameters
		params := url.Values{
			"txt_term":      {term},
			"pageOffset":    {fmt.Sprintf("%d", pageOffset)},
			"pageMaxSize":   {fmt.Sprintf("%d", pageSize)},
			"sortColumn":    {"subjectDescription"},
			"sortDirection": {"asc"},
		}

		// Simply exclude the parameters if they are empty
		if subject != "" && subject != "%" {
			params.Add("txt_subject", subject)
		}
		if courseNum != "" && courseNum != "%" {
			params.Add("txt_courseNumber", courseNum)
		}

		// Create request
		searchURL := c.baseURL + "/searchResults/searchResults"
		req, err := http.NewRequest("GET", searchURL, nil)
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
	}

	return allCourses, nil
}
