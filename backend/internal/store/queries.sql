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

-- name: GetMeetingTimesBySectionIDs :many
SELECT
    section_id, start_time, end_time, building, room,
    sunday, monday, tuesday, wednesday, thursday, friday, saturday
FROM meeting_times
WHERE section_id IN (sqlc.slice('section_ids'))
ORDER BY section_id;

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

-- name: LogGeneration :one
INSERT INTO generation_logs (
    session_id, term, courses_count, schedules_generated,
    min_courses, max_courses, blocked_times_count, duration_ms
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: LogGenerationCourse :exec
INSERT INTO generation_log_courses (
    generation_log_id, subject, course_number, required
) VALUES (?, ?, ?, ?);

-- name: LogSearch :exec
INSERT INTO search_logs (
    session_id, term, scope, subject, course_number,
    title, instructor, open_seats, min_credits, max_credits,
    results_count, duration_ms
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

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

-- name: CourseExistsAnyTerm :one
SELECT EXISTS(
    SELECT 1 FROM sections
    WHERE subject = ? AND course_number = ?
) AS course_exists;

-- name: GetSubjectsWithDescriptionsByTerm :many
SELECT DISTINCT subject, subject_description
FROM sections
WHERE term = ?
ORDER BY subject;

-- name: GetSectionWithInstructorByTermAndCRN :one
SELECT
    s.id, s.term, s.crn, s.subject, s.subject_description,
    s.course_number, s.title, s.credit_hours_low,
    s.enrollment, s.max_enrollment, s.seats_available, s.wait_count, s.is_open,
    s.instructional_method,
    i.name AS instructor_name, i.email AS instructor_email
FROM sections s
LEFT JOIN instructors i ON s.id = i.section_id AND i.is_primary = 1
WHERE s.term = ? AND s.crn = ?;

-- name: GetSectionsWithInstructorByCourse :many
SELECT
    s.id, s.term, s.crn, s.subject, s.subject_description,
    s.course_number, s.title, s.credit_hours_low,
    s.enrollment, s.max_enrollment, s.seats_available, s.wait_count, s.is_open,
    s.instructional_method,
    i.name AS instructor_name, i.email AS instructor_email
FROM sections s
LEFT JOIN instructors i ON s.id = i.section_id AND i.is_primary = 1
WHERE s.term = ? AND s.subject = ? AND s.course_number = ?
ORDER BY s.crn;

-- name: ValidateCourseForTerm :one
SELECT
    COUNT(*) AS section_count,
    COALESCE(MAX(title), '') AS title
FROM sections
WHERE term = ? AND subject = ? AND course_number = ?;

-- name: CheckSchemaExists :one
SELECT COUNT(*) AS count FROM sections;

-- name: SearchSections :many
SELECT
    s.id, s.term, s.crn, s.subject, s.subject_description,
    s.course_number, s.title, s.credit_hours_low, s.credit_hours_high,
    s.enrollment, s.max_enrollment, s.seats_available, s.wait_count, s.is_open,
    s.instructional_method, s.schedule_type, s.campus,
    i.name AS instructor_name, i.email AS instructor_email
FROM sections s
LEFT JOIN instructors i ON s.id = i.section_id AND i.is_primary = 1
WHERE
    -- Term filter (NULL = all terms)
    (sqlc.narg('term') IS NULL OR s.term = sqlc.narg('term'))
    -- Subject filter (exact match)
    AND (sqlc.narg('subject') IS NULL OR s.subject = sqlc.narg('subject'))
    -- Course number filter (supports LIKE for wildcards)
    AND (sqlc.narg('course_number') IS NULL OR s.course_number LIKE sqlc.narg('course_number'))
    -- Title tokens (up to 3)
    AND (sqlc.narg('title_t1') IS NULL OR LOWER(s.title) LIKE '%' || LOWER(sqlc.narg('title_t1')) || '%')
    AND (sqlc.narg('title_t2') IS NULL OR LOWER(s.title) LIKE '%' || LOWER(sqlc.narg('title_t2')) || '%')
    AND (sqlc.narg('title_t3') IS NULL OR LOWER(s.title) LIKE '%' || LOWER(sqlc.narg('title_t3')) || '%')
    -- Instructor tokens (up to 3)
    AND (sqlc.narg('instr_t1') IS NULL OR LOWER(i.name) LIKE '%' || LOWER(sqlc.narg('instr_t1')) || '%')
    AND (sqlc.narg('instr_t2') IS NULL OR LOWER(i.name) LIKE '%' || LOWER(sqlc.narg('instr_t2')) || '%')
    AND (sqlc.narg('instr_t3') IS NULL OR LOWER(i.name) LIKE '%' || LOWER(sqlc.narg('instr_t3')) || '%')
    -- Open seats filter
    AND (sqlc.narg('open_seats') IS NULL OR sqlc.narg('open_seats') = 0 OR s.seats_available > 0)
    -- Credit range
    AND (sqlc.narg('min_credits') IS NULL OR s.credit_hours_low >= sqlc.narg('min_credits'))
    AND (sqlc.narg('max_credits') IS NULL OR s.credit_hours_low <= sqlc.narg('max_credits'))
ORDER BY s.term DESC, s.subject, s.course_number, s.crn
LIMIT sqlc.arg('result_limit');

-- name: GetActiveAnnouncement :one
SELECT id, title, body, type FROM announcements
WHERE active = 1 ORDER BY id DESC LIMIT 1;

-- name: InsertFeedback :exec
INSERT INTO feedback (session_id, message) VALUES (?, ?);

-- name: InsertGradeRow :exec
INSERT INTO grade_rows (term, crn, subject, course_number, title, professor,
    students_enrolled, grade_count,
    cnt_a, cnt_am, cnt_bp, cnt_b, cnt_bm, cnt_cp, cnt_c, cnt_cm,
    cnt_dp, cnt_d, cnt_dm, cnt_f, cnt_w, cnt_p, cnt_np, cnt_s, cnt_u)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetGradeRowCount :one
SELECT COUNT(*) FROM grade_rows;

-- name: GetAllGradeRows :many
SELECT * FROM grade_rows;

-- name: UpsertSubjectMapping :exec
INSERT INTO subject_mappings (banner_subject, grade_subject, match_count)
VALUES (?, ?, ?)
ON CONFLICT(banner_subject) DO UPDATE SET
    grade_subject = excluded.grade_subject,
    match_count = excluded.match_count;

-- name: GetSubjectMappings :many
SELECT * FROM subject_mappings;

-- name: GetSubjectMappingCount :one
SELECT COUNT(*) FROM subject_mappings;

-- name: UpsertInstructorMapping :exec
INSERT INTO instructor_mappings (banner_name, grade_name, match_count)
VALUES (?, ?, ?)
ON CONFLICT(banner_name) DO UPDATE SET
    grade_name = excluded.grade_name,
    match_count = excluded.match_count;

-- name: GetInstructorMappings :many
SELECT * FROM instructor_mappings;

-- name: InsertGradeAggregate :exec
INSERT INTO grade_aggregates (level, subject, course_number, instructor,
    sections_count, students_count,
    cnt_a, cnt_am, cnt_bp, cnt_b, cnt_bm, cnt_cp, cnt_c, cnt_cm,
    cnt_dp, cnt_d, cnt_dm, cnt_f, cnt_w, cnt_p, cnt_np, cnt_s, cnt_u,
    gpa, pass_rate)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: DeleteAllGradeAggregates :exec
DELETE FROM grade_aggregates;

-- name: GetGradeAggregateCount :one
SELECT COUNT(*) FROM grade_aggregates;

-- name: GetAllGradeAggregates :many
SELECT * FROM grade_aggregates;

-- name: GetGradeBannerJoinData :many
SELECT g.subject as grade_subject, g.course_number as grade_course_number,
       g.professor as grade_professor,
       s.subject as banner_subject, s.course_number as banner_course_number,
       i.name as banner_instructor
FROM grade_rows g
JOIN sections s ON g.term = s.term AND g.crn = s.crn
LEFT JOIN instructors i ON s.id = i.section_id AND i.is_primary = 1;
