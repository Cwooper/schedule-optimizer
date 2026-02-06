-- Initial schema for Schedule Optimizer
-- Based on Banner API response structure

-- Available academic terms
CREATE TABLE terms (
    code TEXT PRIMARY KEY,
    description TEXT NOT NULL,
    last_scraped_at TIMESTAMP
);

-- Main course sections table
CREATE TABLE sections (
    id INTEGER PRIMARY KEY,
    term TEXT NOT NULL,
    crn TEXT NOT NULL,
    subject TEXT NOT NULL,
    subject_description TEXT,
    course_number TEXT NOT NULL,
    sequence_number TEXT,
    title TEXT NOT NULL,
    campus TEXT,
    schedule_type TEXT,
    instructional_method TEXT,
    instructional_method_desc TEXT,
    credit_hours_low INTEGER,
    credit_hours_high INTEGER,
    enrollment INTEGER DEFAULT 0,
    max_enrollment INTEGER DEFAULT 0,
    seats_available INTEGER DEFAULT 0,
    wait_capacity INTEGER DEFAULT 0,
    wait_count INTEGER DEFAULT 0,
    is_open INTEGER DEFAULT 1,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(term, crn)
);

-- Instructors (faculty) for each section
CREATE TABLE instructors (
    id INTEGER PRIMARY KEY,
    section_id INTEGER NOT NULL,
    banner_id TEXT,
    name TEXT NOT NULL,
    email TEXT,
    is_primary INTEGER DEFAULT 1,
    FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE
);

-- Meeting times for each section
CREATE TABLE meeting_times (
    id INTEGER PRIMARY KEY,
    section_id INTEGER NOT NULL,
    start_time TEXT,
    end_time TEXT,
    start_date TEXT,
    end_date TEXT,
    building TEXT,
    building_description TEXT,
    room TEXT,
    monday INTEGER DEFAULT 0,
    tuesday INTEGER DEFAULT 0,
    wednesday INTEGER DEFAULT 0,
    thursday INTEGER DEFAULT 0,
    friday INTEGER DEFAULT 0,
    saturday INTEGER DEFAULT 0,
    sunday INTEGER DEFAULT 0,
    schedule_type TEXT,
    meeting_type TEXT,
    credit_hours REAL,
    hours_per_week REAL,
    FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE
);

-- Section attributes (GUR designations, delivery mode, etc.)
CREATE TABLE section_attributes (
    id INTEGER PRIMARY KEY,
    section_id INTEGER NOT NULL,
    code TEXT NOT NULL,
    description TEXT,
    FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_sections_term ON sections(term);
CREATE INDEX idx_sections_subject ON sections(subject);
CREATE INDEX idx_sections_term_subject ON sections(term, subject);
CREATE INDEX idx_sections_term_subject_course ON sections(term, subject, course_number);
CREATE INDEX idx_sections_crn ON sections(term, crn);
CREATE INDEX idx_instructors_section ON instructors(section_id);
CREATE INDEX idx_instructors_primary ON instructors(section_id, is_primary);
CREATE INDEX idx_meeting_times_section ON meeting_times(section_id);
CREATE INDEX idx_section_attributes_section ON section_attributes(section_id);

-- Analytics logs
CREATE TABLE search_logs (
    id INTEGER PRIMARY KEY,
    query TEXT NOT NULL,
    term TEXT,
    results_count INTEGER,
    session_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE generation_logs (
    id INTEGER PRIMARY KEY,
    term TEXT,
    courses_requested TEXT,
    schedules_generated INTEGER,
    session_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
