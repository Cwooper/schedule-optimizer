package calendar

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	registrarBaseURL   = "https://registrar.wwu.edu"
	termDatesPath      = "/term-dates"
	finalsPath         = "/calendars/finals"
	importantDatesPath = "/important-dates-deadlines"
	drupalAjaxPath     = "/views/ajax"

	httpTimeout = 30 * time.Second
	// Pages run 25-50 KB; the cap only guards against a runaway response.
	maxBodyBytes = 8 << 20
)

type client struct {
	http    *http.Client
	baseURL string
}

func newClient(baseURL string) *client {
	return &client{http: &http.Client{Timeout: httpTimeout}, baseURL: baseURL}
}

func (c *client) get(ctx context.Context, path string, query url.Values) ([]byte, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: status %d", path, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
}

// importantDatesQuery asks Drupal's views AJAX endpoint to render one term
// of the important dates view, as the page's term dropdown does. The view
// name and display come from the page's drupalSettings and change only if
// the registrar rebuilds the view. Drupal answers 500 unless
// ajax_page_state[libraries] is present, even empty.
func importantDatesQuery(nodeID string) url.Values {
	return url.Values{
		"_wrapper_format":            {"drupal_ajax"},
		"view_name":                  {"important_dates_and_deadlines"},
		"view_display_id":            {"block_3"},
		"field_quarter_target_id":    {nodeID},
		"ajax_page_state[libraries]": {""},
	}
}
