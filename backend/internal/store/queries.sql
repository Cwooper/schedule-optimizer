-- name: GetTerms :many
SELECT code, description, last_scraped_at FROM terms ORDER BY code DESC;

-- name: GetTermByCode :one
SELECT code, description, last_scraped_at FROM terms WHERE code = ?;

-- name: UpsertTerm :exec
INSERT INTO terms (code, description)
VALUES (?, ?)
ON CONFLICT(code) DO UPDATE SET description = excluded.description;

-- name: UpdateTermScrapedAt :exec
UPDATE terms SET last_scraped_at = CURRENT_TIMESTAMP WHERE code = ?;

-- name: GetSectionsByTerm :many
SELECT * FROM sections WHERE term = ? ORDER BY subject, course_number;

-- name: GetSectionByTermAndCRN :one
SELECT * FROM sections WHERE term = ? AND crn = ?;

-- name: GetSectionsBySubject :many
SELECT * FROM sections WHERE term = ? AND subject = ? ORDER BY course_number;

-- name: UpsertSection :one
INSERT INTO sections (
    term, crn, subject, subject_description, course_number, sequence_number,
    title, campus, schedule_type, instructional_method, instructional_method_desc,
    credit_hours_low, credit_hours_high, enrollment, max_enrollment, seats_available,
    wait_capacity, wait_count, is_open, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(term, crn) DO UPDATE SET
    subject = excluded.subject,
    subject_description = excluded.subject_description,
    course_number = excluded.course_number,
    sequence_number = excluded.sequence_number,
    title = excluded.title,
    campus = excluded.campus,
    schedule_type = excluded.schedule_type,
    instructional_method = excluded.instructional_method,
    instructional_method_desc = excluded.instructional_method_desc,
    credit_hours_low = excluded.credit_hours_low,
    credit_hours_high = excluded.credit_hours_high,
    enrollment = excluded.enrollment,
    max_enrollment = excluded.max_enrollment,
    seats_available = excluded.seats_available,
    wait_capacity = excluded.wait_capacity,
    wait_count = excluded.wait_count,
    is_open = excluded.is_open,
    updated_at = CURRENT_TIMESTAMP
RETURNING id;

-- name: DeleteSectionsByTerm :exec
DELETE FROM sections WHERE term = ?;

-- name: GetInstructorsBySection :many
SELECT * FROM instructors WHERE section_id = ?;

-- name: GetPrimaryInstructorBySection :one
SELECT * FROM instructors WHERE section_id = ? AND is_primary = 1;

-- name: InsertInstructor :exec
INSERT INTO instructors (section_id, banner_id, name, email, is_primary)
VALUES (?, ?, ?, ?, ?);

-- name: DeleteInstructorsBySection :exec
DELETE FROM instructors WHERE section_id = ?;

-- name: GetMeetingTimesBySection :many
SELECT * FROM meeting_times WHERE section_id = ?;

-- name: InsertMeetingTime :exec
INSERT INTO meeting_times (
    section_id, start_time, end_time, start_date, end_date,
    building, building_description, room,
    monday, tuesday, wednesday, thursday, friday, saturday, sunday,
    schedule_type, meeting_type, credit_hours, hours_per_week
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: DeleteMeetingTimesBySection :exec
DELETE FROM meeting_times WHERE section_id = ?;

-- name: GetSectionAttributesBySection :many
SELECT * FROM section_attributes WHERE section_id = ?;

-- name: InsertSectionAttribute :exec
INSERT INTO section_attributes (section_id, code, description)
VALUES (?, ?, ?);

-- name: DeleteSectionAttributesBySection :exec
DELETE FROM section_attributes WHERE section_id = ?;

-- name: LogSearch :exec
INSERT INTO search_logs (query, term, results_count, session_id)
VALUES (?, ?, ?, ?);

-- name: LogGeneration :exec
INSERT INTO generation_logs (term, courses_requested, schedules_generated, session_id)
VALUES (?, ?, ?, ?);

-- name: GetSectionCount :one
SELECT COUNT(*) FROM sections WHERE term = ?;

-- name: GetDistinctSubjects :many
SELECT DISTINCT subject FROM sections ORDER BY subject;

-- name: GetDistinctSubjectsByTerm :many
SELECT DISTINCT subject FROM sections WHERE term = ? ORDER BY subject;

-- name: GetDistinctTerms :many
SELECT DISTINCT term FROM sections ORDER BY term DESC;

-- name: GetSectionsWithInstructorByTerm :many
SELECT
    s.id, s.term, s.crn, s.subject, s.subject_description,
    s.course_number, s.title, s.credit_hours_low,
    s.enrollment, s.max_enrollment, s.seats_available, s.wait_count, s.is_open,
    s.instructional_method,
    i.name AS instructor_name, i.email AS instructor_email
FROM sections s
LEFT JOIN instructors i ON s.id = i.section_id AND i.is_primary = 1
WHERE s.term = ?
ORDER BY s.id;

-- name: GetMeetingTimesByTerm :many
SELECT
    m.section_id, m.start_time, m.end_time, m.building, m.room,
    m.sunday, m.monday, m.tuesday, m.wednesday, m.thursday, m.friday, m.saturday
FROM meeting_times m
JOIN sections s ON m.section_id = s.id
WHERE s.term = ?
ORDER BY m.section_id;

-- name: GetTermsNeverScraped :many
SELECT code, description, last_scraped_at FROM terms
WHERE last_scraped_at IS NULL ORDER BY code DESC;

