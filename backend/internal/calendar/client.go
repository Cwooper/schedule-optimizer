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
	// Read one byte past the cap: a truncated page can still parse, and
	// would replace complete stored rows with a partial set.
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", path, err)
	}
	if len(body) > maxBodyBytes {
		return nil, fmt.Errorf("fetch %s: body over %d bytes", path, maxBodyBytes)
	}
	return body, nil
}
