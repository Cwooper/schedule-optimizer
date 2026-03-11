package calendar

import (
	"context"
	"database/sql"
	"log/slog"

	"schedule-optimizer/internal/store"
)

func (s *Service) storeTermDates(ctx context.Context, dates []TermDates) error {
	for _, td := range dates {
		var finalsStart, finalsEnd sql.NullString
		if !td.FinalsStart.IsZero() {
			finalsStart = sql.NullString{String: td.FinalsStart.Format("2006-01-02"), Valid: true}
		}
		if !td.FinalsEnd.IsZero() {
			finalsEnd = sql.NullString{String: td.FinalsEnd.Format("2006-01-02"), Valid: true}
		}

		if err := s.queries.UpsertTermDates(ctx, store.UpsertTermDatesParams{
			TermCode:    td.TermCode,
			StartDate:   td.StartDate.Format("2006-01-02"),
			EndDate:     td.EndDate.Format("2006-01-02"),
			FinalsStart: finalsStart,
			FinalsEnd:   finalsEnd,
		}); err != nil {
			slog.Warn("Failed to upsert term dates", "term", td.TermCode, "error", err)
		}
	}
	return nil
}

func (s *Service) storeFinals(ctx context.Context, mappings []FinalMapping) error {
	// Delete existing mappings per term before reinserting
	seen := make(map[string]bool)
	for _, m := range mappings {
		if !seen[m.TermCode] {
			if err := s.queries.DeleteFinalMappingsByTerm(ctx, m.TermCode); err != nil {
				slog.Warn("Failed to delete finals mappings", "term", m.TermCode, "error", err)
			}
			seen[m.TermCode] = true
		}

		hasTuth := int64(0)
		if m.HasTuTh {
			hasTuth = 1
		}

		if err := s.queries.UpsertFinalMapping(ctx, store.UpsertFinalMappingParams{
			TermCode:       m.TermCode,
			TimeRangeStart: m.TimeRangeStart,
			TimeRangeEnd:   m.TimeRangeEnd,
			HasTuth:        hasTuth,
			ExamDate:       m.ExamDate.Format("2006-01-02"),
			ExamStartTime:  m.ExamStartTime,
			ExamEndTime:    m.ExamEndTime,
		}); err != nil {
			slog.Warn("Failed to upsert final mapping", "term", m.TermCode, "error", err)
		}
	}
	return nil
}

func (s *Service) storeHolidays(ctx context.Context, holidays []Holiday) error {
	for _, h := range holidays {
		if err := s.queries.UpsertHoliday(ctx, store.UpsertHolidayParams{
			TermCode:    h.TermCode,
			Date:        h.Date.Format("2006-01-02"),
			Description: h.Description,
		}); err != nil {
			slog.Warn("Failed to upsert holiday", "term", h.TermCode, "error", err)
		}
	}
	return nil
}

func (s *Service) storeImportantDates(ctx context.Context, dates []ImportantDate) error {
	for _, d := range dates {
		if err := s.queries.UpsertImportantDate(ctx, store.UpsertImportantDateParams{
			TermCode:    d.TermCode,
			Date:        d.Date.Format("2006-01-02"),
			Description: d.Description,
			Category:    d.Category,
		}); err != nil {
			slog.Warn("Failed to upsert important date", "term", d.TermCode, "error", err)
		}
	}
	return nil
}
