-- Academic calendar data scraped from registrar.wwu.edu
-- Supports ICS export, finals lookup, and replaces hardcoded term dates in jobs

-- Term start/end dates and finals week dates
CREATE TABLE term_dates (
    term_code TEXT PRIMARY KEY REFERENCES terms(code) ON DELETE CASCADE,
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    finals_start TEXT,
    finals_end TEXT,
    scraped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Finals schedule mapping: class start time + day pattern -> exam slot
CREATE TABLE finals_mappings (
    id INTEGER PRIMARY KEY,
    term_code TEXT NOT NULL REFERENCES terms(code) ON DELETE CASCADE,
    time_range_start TEXT NOT NULL,
    time_range_end TEXT NOT NULL,
    has_tuth INTEGER NOT NULL,
    exam_date TEXT NOT NULL,
    exam_start_time TEXT NOT NULL,
    exam_end_time TEXT NOT NULL,
    UNIQUE(term_code, time_range_start, has_tuth)
);

-- Holidays (no-class days) per term
CREATE TABLE holidays (
    id INTEGER PRIMARY KEY,
    term_code TEXT NOT NULL REFERENCES terms(code) ON DELETE CASCADE,
    date TEXT NOT NULL,
    description TEXT NOT NULL,
    UNIQUE(term_code, date)
);

-- Catch-all for important dates and deadlines
CREATE TABLE important_dates (
    id INTEGER PRIMARY KEY,
    term_code TEXT NOT NULL REFERENCES terms(code) ON DELETE CASCADE,
    date TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'deadline',
    UNIQUE(term_code, date, description)
);

CREATE INDEX idx_finals_mappings_term ON finals_mappings(term_code);
CREATE INDEX idx_holidays_term ON holidays(term_code);
CREATE INDEX idx_important_dates_term ON important_dates(term_code);
