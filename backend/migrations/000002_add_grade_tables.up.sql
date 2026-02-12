CREATE TABLE grade_rows (
    id INTEGER PRIMARY KEY,
    term TEXT NOT NULL,
    crn TEXT NOT NULL,
    subject TEXT NOT NULL,
    course_number TEXT NOT NULL,
    title TEXT NOT NULL,
    professor TEXT NOT NULL DEFAULT '',
    students_enrolled INTEGER NOT NULL DEFAULT 0,
    grade_count INTEGER NOT NULL DEFAULT 0,
    cnt_a INTEGER NOT NULL DEFAULT 0,
    cnt_am INTEGER NOT NULL DEFAULT 0,
    cnt_bp INTEGER NOT NULL DEFAULT 0,
    cnt_b INTEGER NOT NULL DEFAULT 0,
    cnt_bm INTEGER NOT NULL DEFAULT 0,
    cnt_cp INTEGER NOT NULL DEFAULT 0,
    cnt_c INTEGER NOT NULL DEFAULT 0,
    cnt_cm INTEGER NOT NULL DEFAULT 0,
    cnt_dp INTEGER NOT NULL DEFAULT 0,
    cnt_d INTEGER NOT NULL DEFAULT 0,
    cnt_dm INTEGER NOT NULL DEFAULT 0,
    cnt_f INTEGER NOT NULL DEFAULT 0,
    cnt_w INTEGER NOT NULL DEFAULT 0,
    cnt_p INTEGER NOT NULL DEFAULT 0,
    cnt_np INTEGER NOT NULL DEFAULT 0,
    cnt_s INTEGER NOT NULL DEFAULT 0,
    cnt_u INTEGER NOT NULL DEFAULT 0,
    UNIQUE(term, crn)
);
CREATE INDEX idx_grade_rows_subject ON grade_rows(subject, course_number);

CREATE TABLE subject_mappings (
    banner_subject TEXT PRIMARY KEY,
    grade_subject TEXT NOT NULL,
    match_count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE instructor_mappings (
    banner_name TEXT PRIMARY KEY,
    grade_name TEXT NOT NULL,
    match_count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE grade_aggregates (
    id INTEGER PRIMARY KEY,
    level TEXT NOT NULL,
    subject TEXT NOT NULL,
    course_number TEXT NOT NULL DEFAULT '',
    instructor TEXT NOT NULL DEFAULT '',
    sections_count INTEGER NOT NULL DEFAULT 0,
    students_count INTEGER NOT NULL DEFAULT 0,
    cnt_a INTEGER NOT NULL DEFAULT 0,
    cnt_am INTEGER NOT NULL DEFAULT 0,
    cnt_bp INTEGER NOT NULL DEFAULT 0,
    cnt_b INTEGER NOT NULL DEFAULT 0,
    cnt_bm INTEGER NOT NULL DEFAULT 0,
    cnt_cp INTEGER NOT NULL DEFAULT 0,
    cnt_c INTEGER NOT NULL DEFAULT 0,
    cnt_cm INTEGER NOT NULL DEFAULT 0,
    cnt_dp INTEGER NOT NULL DEFAULT 0,
    cnt_d INTEGER NOT NULL DEFAULT 0,
    cnt_dm INTEGER NOT NULL DEFAULT 0,
    cnt_f INTEGER NOT NULL DEFAULT 0,
    cnt_w INTEGER NOT NULL DEFAULT 0,
    cnt_p INTEGER NOT NULL DEFAULT 0,
    cnt_np INTEGER NOT NULL DEFAULT 0,
    cnt_s INTEGER NOT NULL DEFAULT 0,
    cnt_u INTEGER NOT NULL DEFAULT 0,
    gpa REAL NOT NULL DEFAULT 0.0,
    pass_rate REAL,
    UNIQUE(level, subject, course_number, instructor)
);
