package calendar

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	registrarBaseURL = "https://registrar.wwu.edu"
	httpTimeout      = 30 * time.Second
)

// Client handles HTTP requests to the WWU registrar website.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new registrar website client.
func NewClient() *Client {
	return newClientWithBaseURL(registrarBaseURL)
}

func newClientWithBaseURL(base string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: httpTimeout},
		baseURL:    base,
	}
}

func (c *Client) fetch(path string) ([]byte, error) {
	resp, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: status %d", path, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// fetchDrupalAjax fetches the important dates AJAX endpoint for a given Drupal node ID.
// The Drupal views/ajax endpoint requires the full query string as a GET request.
// NOTE: view_path and view_name are tied to WWU's Drupal config and may change if the site is restructured.
func (c *Client) fetchDrupalAjax(nodeID string) ([]byte, error) {
	url := c.baseURL + "/views/ajax?" +
		"_wrapper_format=drupal_ajax" +
		"&field_quarter_target_id=" + nodeID +
		"&view_name=important_dates_and_deadlines" +
		"&view_display_id=block_3" +
		"&view_args=" +
		"&view_path=%2Fnode%2F1172" +
		"&view_base_path=tuition-fees-calendar" +
		"&view_dom_id=placeholder" +
		"&pager_element=0" +
		"&_drupal_ajax=1" +
		"&ajax_page_state%5Btheme%5D=ashlar" +
		"&ajax_page_state%5Btheme_token%5D=" +
		"&ajax_page_state%5Blibraries%5D="

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch drupal ajax (node %s): %w", nodeID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch drupal ajax (node %s): status %d", nodeID, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
