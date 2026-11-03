// Package calendar scrapes WWU's academic calendar from registrar.wwu.edu
// and answers when each term's registration, classes, and end fall, and
// which final exam slot a section takes (finals.go).
//
// Banner exposes sections but no dates, so the registrar site is the only
// source for term dates, registration phases, deadlines, holidays, and the
// finals schedule. Three pages feed it:
//   - /term-dates: start and end dates, published years ahead
//   - /calendars/finals: the exam slot for each class start time
//   - /important-dates-deadlines: registration phases, deadlines, and
//     holidays, one term per page through its term filter
//
// Sources and terms are independent: one failing does not block the others.
// A term that does not parse in full is skipped, keeping its stored rows,
// rather than replaced by a partial set. Finals and important dates are
// replaced per term in one transaction so a reader never sees a half-written
// term; term dates are upserted, so past terms keep theirs. Phase detection
// (schedule.go) falls back to estimated dates for any term the registrar has
// not published.
package calendar

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"

	"schedule-optimizer/internal/store"
)

type Service struct {
	db      *sql.DB
	queries *store.Queries
	client  *client
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db, queries: store.New(db), client: newClient(registrarBaseURL)}
}

// ScrapeAll refreshes every calendar source and returns all their errors
// joined. Rows are stored only for terms already in the terms table: the
// registrar publishes terms years before Banner does, and the foreign keys
// need the term to exist.
func (s *Service) ScrapeAll(ctx context.Context) error {
	known, err := s.knownTerms(ctx)
	if err != nil {
		return fmt.Errorf("loading known terms: %w", err)
	}
	return errors.Join(
		s.scrapeTermDates(ctx, known),
		s.scrapeFinals(ctx, known),
		s.scrapeImportantDates(ctx, known),
	)
}

func (s *Service) knownTerms(ctx context.Context) (map[string]bool, error) {
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

func (s *Service) scrapeTermDates(ctx context.Context, known map[string]bool) error {
	body, err := s.client.get(ctx, termDatesPath, nil)
	if err != nil {
		return err
	}
	dates, failed, err := parseTermDates(body)
	if err != nil {
		return err
	}
	dates = slices.DeleteFunc(dates, func(d termDates) bool { return !known[d.termCode] })
	if err := s.storeTermDates(ctx, dates); err != nil {
		return err
	}
	slog.Info("Scraped term dates", "terms", len(dates))
	return knownFailures("term dates", failed, known)
}

func (s *Service) scrapeFinals(ctx context.Context, known map[string]bool) error {
	body, err := s.client.get(ctx, finalsPath, nil)
	if err != nil {
		return err
	}
	mappings, failed, err := parseFinals(body)
	if err != nil {
		return err
	}
	errs := []error{knownFailures("finals", failed, known)}
	for term, rows := range groupByTerm(mappings, func(m finalMapping) string { return m.termCode }) {
		if !known[term] {
			continue
		}
		if err := s.replaceFinals(ctx, term, rows); err != nil {
			errs = append(errs, fmt.Errorf("finals for %s: %w", term, err))
			continue
		}
		slog.Info("Scraped finals schedule", "term", term, "slots", len(rows))
	}
	return errors.Join(errs...)
}

// scrapeImportantDates reads the term dropdown from the unfiltered page,
// then fetches the page filtered to each known term.
func (s *Service) scrapeImportantDates(ctx context.Context, known map[string]bool) error {
	body, err := s.client.get(ctx, importantDatesPath, nil)
	if err != nil {
		return err
	}
	options, err := parseTermSelect(body)
	if err != nil {
		return err
	}

	var errs []error
	for _, opt := range options {
		if !known[opt.termCode] {
			continue
		}
		dates, err := s.fetchImportantDates(ctx, opt)
		if err == nil {
			err = s.replaceImportantDates(ctx, opt.termCode, dates)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("important dates for %s: %w", opt.termCode, err))
			continue
		}
		slog.Info("Scraped important dates", "term", opt.termCode, "dates", len(dates))
	}
	return errors.Join(errs...)
}

func (s *Service) fetchImportantDates(ctx context.Context, opt termOption) ([]importantDate, error) {
	body, err := s.client.get(ctx, importantDatesPath, url.Values{"field_quarter_target_id": {opt.nodeID}})
	if err != nil {
		return nil, err
	}
	return parseImportantDates(body, opt.termCode)
}

// knownFailures reports the parse failures of terms Banner lists, and those
// tied to no term. A malformed row for a term years out would otherwise fail
// every daily run until the registrar fixes it.
func knownFailures(source string, failed []termError, known map[string]bool) error {
	var errs []error
	for _, f := range failed {
		if f.code == "" || known[f.code] {
			errs = append(errs, fmt.Errorf("%s for term %q: %w", source, f.code, f.err))
		}
	}
	return errors.Join(errs...)
}

func groupByTerm[T any](items []T, termCode func(T) string) map[string][]T {
	groups := make(map[string][]T)
	for _, item := range items {
		code := termCode(item)
		groups[code] = append(groups[code], item)
	}
	return groups
}
