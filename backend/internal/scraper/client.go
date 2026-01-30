package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const (
	baseURL     = "https://registration.banner.wwu.edu/StudentRegistrationSsb/ssb"
	pageSize    = 500
	httpTimeout = 2 * time.Minute // Banner servers can be slow under load
)

// Client handles HTTP requests to the Banner API.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Banner API client with a cookie jar.
func NewClient() (*Client, error) {
	return newClientWithBaseURL(baseURL)
}

func newClientWithBaseURL(base string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}

	return &Client{
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: httpTimeout,
		},
		baseURL: base,
	}, nil
}

// FetchTerms retrieves available terms from the Banner API.
// This also initializes cookies needed for subsequent requests.
func (c *Client) FetchTerms() ([]TermResponse, error) {
	reqURL := c.baseURL + "/classSearch/getTerms?searchTerm=&offset=1&max=100"

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("fetch terms: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch terms: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read terms response: %w", err)
	}

	var terms []TermResponse
	if err := json.Unmarshal(body, &terms); err != nil {
		return nil, fmt.Errorf("decode terms: %w", err)
	}

	slog.Debug("Fetched terms", "count", len(terms))
	return terms, nil
}

// InitializeSession sets the term context for subsequent course fetches.
// Must be called after FetchTerms and before FetchPage.
func (c *Client) InitializeSession(term string) error {
	reqURL := c.baseURL + "/term/search?mode=search"

	data := url.Values{}
	data.Set("term", term)
	data.Set("studyPath", "")
	data.Set("studyPathText", "")
	data.Set("startDatepicker", "")
	data.Set("endDatepicker", "")

	resp, err := c.httpClient.PostForm(reqURL, data)
	if err != nil {
		return fmt.Errorf("initialize session: %w", err)
	}
	defer resp.Body.Close()

	// Read and discard body to ensure connection can be reused
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("initialize session: unexpected status %d", resp.StatusCode)
	}

	slog.Debug("Session initialized", "term", term)
	return nil
}

// FetchPage fetches a single page of course results.
func (c *Client) FetchPage(term string, offset int) (*PageResult, error) {
	reqURL := c.baseURL + "/searchResults/searchResults"

	params := url.Values{}
	params.Set("txt_term", term)
	params.Set("pageOffset", fmt.Sprintf("%d", offset))
	params.Set("pageMaxSize", fmt.Sprintf("%d", pageSize))
	params.Set("sortColumn", "subjectDescription")
	params.Set("sortDirection", "asc")

	fullURL := reqURL + "?" + params.Encode()

	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		return &PageResult{Offset: offset, Error: fmt.Errorf("fetch page %d: %w", offset, err)}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &PageResult{
			Offset: offset,
			Error:  fmt.Errorf("fetch page %d: unexpected status %d", offset, resp.StatusCode),
		}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &PageResult{Offset: offset, Error: fmt.Errorf("read page %d: %w", offset, err)}, nil
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		// Try to get a snippet of the response for debugging
		snippet := string(body)
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		snippet = strings.ReplaceAll(snippet, "\n", " ")
		return &PageResult{
			Offset: offset,
			Error:  fmt.Errorf("decode page %d: %w (response: %s)", offset, err, snippet),
		}, nil
	}

	if !apiResp.Success {
		return &PageResult{
			Offset: offset,
			Error:  fmt.Errorf("page %d: API returned success=false", offset),
		}, nil
	}

	slog.Debug("Fetched page", "offset", offset, "courses", len(apiResp.Data), "total", apiResp.TotalCount)

	return &PageResult{
		Courses:    apiResp.Data,
		TotalCount: apiResp.TotalCount,
		Offset:     offset,
	}, nil
}

// PageSize returns the number of courses per page.
func PageSize() int {
	return pageSize
}
