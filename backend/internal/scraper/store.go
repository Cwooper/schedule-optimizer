package scraper

import (
	"context"
	"database/sql"
	"fmt"

	"schedule-optimizer/internal/store"
)

// saveCourse persists a single course to the database.
// Handles upsert of section, then replaces all child records.
func saveCourse(ctx context.Context, queries *store.Queries, course CourseData) error {
	// Determine credit hours (use CreditHourLow, fallback to CreditHours)
	var creditLow, creditHigh sql.NullInt64
	if course.CreditHourLow != nil {
		creditLow = toNullInt64(int64(*course.CreditHourLow))
	}
	if course.CreditHourHigh != nil {
		creditHigh = toNullInt64(int64(*course.CreditHourHigh))
	}

	// Upsert section
	sectionID, err := queries.UpsertSection(ctx, store.UpsertSectionParams{
		Term:                    course.Term,
		Crn:                     course.CourseReferenceNumber,
		Subject:                 course.Subject,
		SubjectDescription:      toNullString(course.SubjectDescription),
		CourseNumber:            course.CourseNumber,
		SequenceNumber:          toNullString(course.SequenceNumber),
		Title:                   course.CourseTitle,
		Campus:                  toNullString(course.CampusDescription),
		ScheduleType:            toNullString(course.ScheduleTypeDescription),
		InstructionalMethod:     toNullString(course.InstructionalMethod),
		InstructionalMethodDesc: toNullString(course.InstructionalMethodDescr),
		CreditHoursLow:          creditLow,
		CreditHoursHigh:         creditHigh,
		Enrollment:              toNullInt64(int64(course.Enrollment)),
		MaxEnrollment:           toNullInt64(int64(course.MaximumEnrollment)),
		SeatsAvailable:          toNullInt64(int64(course.SeatsAvailable)),
		WaitCapacity:            toNullInt64(int64(course.WaitCapacity)),
		WaitCount:               toNullInt64(int64(course.WaitCount)),
		IsOpen:                  toNullInt64(boolToInt64(course.OpenSection)),
	})
	if err != nil {
		return fmt.Errorf("upsert section %s: %w", course.CourseReferenceNumber, err)
	}

	// Delete existing children before inserting new ones
	if err := queries.DeleteInstructorsBySection(ctx, sectionID); err != nil {
		return fmt.Errorf("delete instructors for section %d: %w", sectionID, err)
	}
	if err := queries.DeleteMeetingTimesBySection(ctx, sectionID); err != nil {
		return fmt.Errorf("delete meeting times for section %d: %w", sectionID, err)
	}
	if err := queries.DeleteSectionAttributesBySection(ctx, sectionID); err != nil {
		return fmt.Errorf("delete attributes for section %d: %w", sectionID, err)
	}

	// Insert instructors
	for _, faculty := range course.Faculty {
		if err := queries.InsertInstructor(ctx, store.InsertInstructorParams{
			SectionID: sectionID,
			BannerID:  toNullString(faculty.BannerID),
			Name:      faculty.DisplayName,
			Email:     toNullString(faculty.EmailAddress),
			IsPrimary: sql.NullInt64{Int64: boolToInt64(faculty.PrimaryIndicator), Valid: true},
		}); err != nil {
			return fmt.Errorf("insert instructor for section %d: %w", sectionID, err)
		}
	}

	// Insert meeting times
	for _, mf := range course.MeetingsFaculty {
		mt := mf.MeetingTime
		if err := queries.InsertMeetingTime(ctx, store.InsertMeetingTimeParams{
			SectionID:           sectionID,
			StartTime:           toNullString(mt.BeginTime),
			EndTime:             toNullString(mt.EndTime),
			StartDate:           toNullString(mt.StartDate),
			EndDate:             toNullString(mt.EndDate),
			Building:            toNullString(mt.Building),
			BuildingDescription: toNullString(mt.BuildingDescription),
			Room:                toNullString(mt.Room),
			Monday:              toNullInt64(boolToInt64(mt.Monday)),
			Tuesday:             toNullInt64(boolToInt64(mt.Tuesday)),
			Wednesday:           toNullInt64(boolToInt64(mt.Wednesday)),
			Thursday:            toNullInt64(boolToInt64(mt.Thursday)),
			Friday:              toNullInt64(boolToInt64(mt.Friday)),
			Saturday:            toNullInt64(boolToInt64(mt.Saturday)),
			Sunday:              toNullInt64(boolToInt64(mt.Sunday)),
			ScheduleType:        toNullString(mt.MeetingScheduleType),
			MeetingType:         toNullString(mt.MeetingType),
			CreditHours:         toNullFloat64(mt.CreditHourSession),
			HoursPerWeek:        toNullFloat64(mt.HoursWeek),
		}); err != nil {
			return fmt.Errorf("insert meeting time for section %d: %w", sectionID, err)
		}
	}

	// Insert section attributes
	for _, attr := range course.SectionAttributes {
		if err := queries.InsertSectionAttribute(ctx, store.InsertSectionAttributeParams{
			SectionID:   sectionID,
			Code:        attr.Code,
			Description: toNullString(attr.Description),
		}); err != nil {
			return fmt.Errorf("insert attribute for section %d: %w", sectionID, err)
		}
	}

	return nil
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func toNullFloat64(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

func toNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: true}
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
