package calendar

import (
	"context"
	"database/sql"
	"time"

	"schedule-optimizer/internal/store"
)

func (s *Service) withTx(ctx context.Context, fn func(*store.Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(s.queries.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit()
}

// storeTermDates upserts rather than replaces: past terms drop off the page
// but keep their dates.
func (s *Service) storeTermDates(ctx context.Context, dates []termDates) error {
	return s.withTx(ctx, func(q *store.Queries) error {
		for _, d := range dates {
			if err := q.UpsertTermDates(ctx, store.UpsertTermDatesParams{
				TermCode:  d.termCode,
				StartDate: d.start.Format(dateLayout),
				EndDate:   d.end.Format(dateLayout),
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) replaceFinals(ctx context.Context, term string, mappings []finalMapping) error {
	return s.withTx(ctx, func(q *store.Queries) error {
		if err := q.DeleteFinalMappingsByTerm(ctx, term); err != nil {
			return err
		}
		for _, m := range mappings {
			if err := q.InsertFinalMapping(ctx, store.InsertFinalMappingParams{
				TermCode:       term,
				TimeRangeStart: m.rangeStart,
				TimeRangeEnd:   m.rangeEnd,
				HasTuth:        m.hasTuTh,
				ExamDate:       m.examDate.Format(dateLayout),
				ExamStartTime:  m.examStart,
				ExamEndTime:    m.examEnd,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

// replaceImportantDates replaces the term's rows wholesale, so a row the
// registrar renamed or removed does not linger beside its replacement.
func (s *Service) replaceImportantDates(ctx context.Context, term string, dates []importantDate) error {
	return s.withTx(ctx, func(q *store.Queries) error {
		if err := q.DeleteImportantDatesByTerm(ctx, term); err != nil {
			return err
		}
		for _, d := range dates {
			if err := q.InsertImportantDate(ctx, store.InsertImportantDateParams{
				TermCode:          term,
				Date:              d.date.Format(dateLayout),
				EndDate:           nullDate(d.end),
				Description:       d.description,
				Category:          string(d.category),
				OpensRegistration: d.opensRegistration,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func nullDate(t time.Time) sql.NullString {
	if t.IsZero() {
		return sql.NullString{}
	}
	return sql.NullString{String: t.Format(dateLayout), Valid: true}
}
