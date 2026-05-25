package calendar

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"schedule-optimizer/internal/store"
)

// Service scrapes and stores academic calendar data from the WWU registrar website.
type Service struct {
	queries *store.Queries
	client  *Client
}

// NewService creates a new calendar service.
func NewService(queries *store.Queries) *Service {
	return newServiceWithClient(queries, NewClient())
}

func newServiceWithClient(queries *store.Queries, client *Client) *Service {
	return &Service{queries: queries, client: client}
}

// ScrapeAll fetches and stores all calendar data.
// Partial success is acceptable — each data source is independent.
// Only stores data for terms that already exist in the terms table
// (populated by the Banner scraper), skipping future terms the registrar
// knows about but Banner hasn't published yet.
func (s *Service) ScrapeAll(ctx context.Context) error {
	known, err := s.loadKnownTerms(ctx)
	if err != nil {
		return fmt.Errorf("loading known terms: %w", err)
	}

	var errs []error

	if err := s.scrapeTermDates(ctx, known); err != nil {
		slog.Error("Failed to scrape term dates", "error", err)
		errs = append(errs, err)
	}

	if err := s.scrapeFinals(ctx, known); err != nil {
		slog.Error("Failed to scrape finals schedule", "error", err)
		errs = append(errs, err)
	}

	if err := s.scrapeImportantDates(ctx, known); err != nil {
		slog.Error("Failed to scrape important dates", "error", err)
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// loadKnownTerms returns the set of term codes that exist in the terms table.
func (s *Service) loadKnownTerms(ctx context.Context) (map[string]bool, error) {
	terms, err := s.queries.GetTerms(ctx)
	if err != nil {
		return nil, err
	}
	known := make(map[string]bool, len(terms))
	for _, t := range terms {
		known[t.Code] = true
	}
	return known, nil
}

// scrapeTermDates fetches and stores term start/end dates.
func (s *Service) scrapeTermDates(ctx context.Context, known map[string]bool) error {
	body, err := s.client.fetch("/term-dates")
	if err != nil {
		return err
	}

	dates, err := parseTermDates(body)
	if err != nil {
		return err
	}

	filtered := filterSlice(dates, known, func(d TermDates) string { return d.TermCode })
	slog.Info("Parsed term dates", "total", len(dates), "known", len(filtered))
	return s.storeTermDates(ctx, filtered)
}

// scrapeFinals fetches and stores finals schedule mappings.
func (s *Service) scrapeFinals(ctx context.Context, known map[string]bool) error {
	body, err := s.client.fetch("/calendars/finals")
	if err != nil {
		return err
	}

	mappings, err := parseFinals(body)
	if err != nil {
		return err
	}

	filtered := filterSlice(mappings, known, func(m FinalMapping) string { return m.TermCode })
	slog.Info("Parsed finals mappings", "total", len(mappings), "known", len(filtered))
	return s.storeFinals(ctx, filtered)
}

// scrapeImportantDates fetches the term selector, then scrapes each term's important dates.
func (s *Service) scrapeImportantDates(ctx context.Context, known map[string]bool) error {
	body, err := s.client.fetch("/important-dates-deadlines")
	if err != nil {
		return err
	}

	options, err := parseDrupalTermSelect(body)
	if err != nil {
		return err
	}

	slog.Info("Found term options for important dates", "count", len(options))

	for _, opt := range options {
		if !known[opt.TermCode] {
			slog.Debug("Skipping unknown term for important dates", "term", opt.TermCode)
			continue
		}

		ajaxBody, err := s.client.fetchDrupalAjax(opt.NodeID)
		if err != nil {
			slog.Warn("Failed to fetch important dates", "term", opt.TermCode, "error", err)
			continue
		}

		holidays, importantDates, err := parseImportantDatesAjax(ajaxBody, opt.TermCode)
		if err != nil {
			slog.Warn("Failed to parse important dates", "term", opt.TermCode, "error", err)
			continue
		}

		if err := s.storeHolidays(ctx, holidays); err != nil {
			slog.Warn("Failed to store holidays", "term", opt.TermCode, "error", err)
		}
		if err := s.storeImportantDates(ctx, importantDates); err != nil {
			slog.Warn("Failed to store important dates", "term", opt.TermCode, "error", err)
		}

		slog.Info("Scraped important dates",
			"term", opt.TermCode,
			"holidays", len(holidays),
			"dates", len(importantDates),
		)
	}

	return nil
}

// GetFinalForSection resolves the final exam slot for a section given its
// meeting start time and whether it includes Tuesday or Thursday meetings.
func (s *Service) GetFinalForSection(ctx context.Context, termCode string, startTime string, hasTuTh bool) (*FinalMapping, error) {
	mappings, err := s.queries.GetFinalMappingsByTerm(ctx, termCode)
	if err != nil {
		return nil, err
	}

	hasTuThInt := int64(0)
	if hasTuTh {
		hasTuThInt = 1
	}

	for _, m := range mappings {
		if m.HasTuth != hasTuThInt {
			continue
		}
		// Check if startTime falls within this mapping's range
		if startTime >= m.TimeRangeStart && startTime <= m.TimeRangeEnd {
			return &FinalMapping{
				TermCode:       m.TermCode,
				TimeRangeStart: m.TimeRangeStart,
				TimeRangeEnd:   m.TimeRangeEnd,
				HasTuTh:        hasTuTh,
				ExamDate:       parseDate(m.ExamDate),
				ExamStartTime:  m.ExamStartTime,
				ExamEndTime:    m.ExamEndTime,
			}, nil
		}
	}

	return nil, nil // no mapping found (e.g., evening class)
}

func parseDate(s string) (t time.Time) {
	t, _ = time.Parse("2006-01-02", s)
	return
}

// filterSlice returns only elements whose term code exists in known.
func filterSlice[T any](items []T, known map[string]bool, termCode func(T) string) []T {
	result := make([]T, 0, len(items))
	for _, item := range items {
		if known[termCode(item)] {
			result = append(result, item)
		}
	}
	return result
}
