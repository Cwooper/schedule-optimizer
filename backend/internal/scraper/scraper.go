// Package scraper fetches course data from WWU's Banner API and stores it in SQLite.
//
// # Banner API Quirks
//
// The Banner API requires a specific request sequence:
//  1. GET /classSearch/getTerms - Returns available terms AND sets session cookies
//  2. POST /term/search - Selects a term for the session (server needs ~1s to process)
//  3. GET /searchResults/searchResults - Returns paginated course data (500 per page)
//
// Without this sequence, course fetches return empty results or errors.
//
// # Design Decisions
//
// Concurrent pagination: Banner can be slow (sometimes >1 minute per request under load),
// so we fetch pages concurrently with a configurable worker pool (default 4). This
// significantly reduces total scrape time for large terms (~2500+ courses).
//
// Partial success: If some pages fail, we store what we got and continue. The scraper
// is idempotent - failed pages will be refetched on the next run. This is preferable
// to failing the entire scrape due to one transient error.
//
// 2-minute HTTP timeout: Banner servers can be extremely slow under load. The previous
// implementation used 5 minutes; we use 2 minutes as a compromise.
//
// 1-second delay after session init: The previous implementation included this delay
// to let the server process term selection. Without it, the first page fetch sometimes
// fails.
//
// # Usage
//
// The Job Service (future) will call ScrapeTerm for each active term on a schedule.
// This package only handles the fetch-and-store logic, not scheduling.
package scraper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"schedule-optimizer/internal/store"
)

// Scraper handles fetching course data from Banner and storing it in the database.
type Scraper struct {
	queries     *store.Queries
	client      *Client
	concurrency int
}

// NewScraper creates a new Scraper with the given database queries and concurrency level.
func NewScraper(queries *store.Queries, concurrency int) (*Scraper, error) {
	return newScraperWithBaseURL(queries, concurrency, baseURL)
}

func newScraperWithBaseURL(queries *store.Queries, concurrency int, base string) (*Scraper, error) {
	if concurrency < 1 {
		concurrency = 4
	}

	client, err := newClientWithBaseURL(base)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	return &Scraper{
		queries:     queries,
		client:      client,
		concurrency: concurrency,
	}, nil
}

// ScrapeTerm fetches all courses for the given term and stores them in the database.
// Returns the number of sections stored. Partial success is possible if some pages fail.
func (s *Scraper) ScrapeTerm(ctx context.Context, term string) (int, error) {
	slog.Info("Starting term scrape", "term", term, "concurrency", s.concurrency)

	// Fetch terms to initialize cookies and upsert term list
	terms, err := s.client.FetchTerms()
	if err != nil {
		return 0, fmt.Errorf("fetch terms: %w", err)
	}

	// Upsert all terms
	for _, t := range terms {
		if err := s.queries.UpsertTerm(ctx, store.UpsertTermParams{
			Code:        t.Code,
			Description: t.Description,
		}); err != nil {
			slog.Warn("Failed to upsert term", "code", t.Code, "error", err)
		}
	}

	// Initialize session for the target term
	if err := s.client.InitializeSession(term); err != nil {
		return 0, fmt.Errorf("initialize session: %w", err)
	}

	// Brief delay to let server process term selection
	time.Sleep(time.Second)

	// Fetch first page to get total count
	firstPage, err := s.client.FetchPage(term, 0)
	if err != nil {
		return 0, fmt.Errorf("fetch first page: %w", err)
	}
	if firstPage.Error != nil {
		return 0, fmt.Errorf("first page error: %w", firstPage.Error)
	}

	totalCount := firstPage.TotalCount
	slog.Info("Term has courses", "term", term, "total", totalCount)

	if totalCount == 0 {
		return 0, nil
	}

	// Calculate offsets for remaining pages
	var offsets []int
	for offset := PageSize(); offset < totalCount; offset += PageSize() {
		offsets = append(offsets, offset)
	}

	// Channel for page results
	results := make(chan *PageResult, len(offsets)+1)

	// Send first page result
	results <- firstPage

	// Start workers for remaining pages
	var wg sync.WaitGroup
	offsetChan := make(chan int, len(offsets))

	for range s.concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for offset := range offsetChan {
				select {
				case <-ctx.Done():
					results <- &PageResult{Offset: offset, Error: ctx.Err()}
					return
				default:
					result, _ := s.client.FetchPage(term, offset)
					results <- result
				}
			}
		}()
	}

	// Send offsets to workers
	for _, offset := range offsets {
		offsetChan <- offset
	}
	close(offsetChan)

	// Close results channel when all workers done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results
	var (
		stored     int
		pageErrors int
		saveErrors int
	)

	for result := range results {
		if result.Error != nil {
			slog.Warn("Page fetch failed", "offset", result.Offset, "error", result.Error)
			pageErrors++
			continue
		}

		for _, course := range result.Courses {
			if err := saveCourse(ctx, s.queries, course); err != nil {
				slog.Warn("Failed to save course",
					"crn", course.CourseReferenceNumber,
					"error", err,
				)
				saveErrors++
				continue
			}
			stored++
		}
	}

	// Check if we got any data at all
	if stored == 0 && pageErrors > 0 {
		return 0, errors.New("all pages failed, no data stored")
	}

	// Update last_scraped_at timestamp
	if err := s.queries.UpdateTermScrapedAt(ctx, term); err != nil {
		slog.Warn("Failed to update term scraped timestamp", "term", term, "error", err)
	}

	slog.Info("Term scrape complete",
		"term", term,
		"stored", stored,
		"page_errors", pageErrors,
		"save_errors", saveErrors,
		"expected", totalCount,
	)

	return stored, nil
}

// ScrapeTerms fetches all available terms from Banner.
// Useful for populating the terms table without scraping course data.
func (s *Scraper) ScrapeTerms(ctx context.Context) ([]TermResponse, error) {
	terms, err := s.client.FetchTerms()
	if err != nil {
		return nil, err
	}

	for _, t := range terms {
		if err := s.queries.UpsertTerm(ctx, store.UpsertTermParams{
			Code:        t.Code,
			Description: t.Description,
		}); err != nil {
			slog.Warn("Failed to upsert term", "code", t.Code, "error", err)
		}
	}

	return terms, nil
}
