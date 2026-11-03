-- Academic calendar data scraped from registrar.wwu.edu. Dates are
-- YYYY-MM-DD local dates in Pacific time; times are HHMM.

CREATE TABLE term_dates (
    term_code TEXT PRIMARY KEY REFERENCES terms(code) ON DELETE CASCADE,
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL -- last day of the term, inclusive
);

-- Final exam slot for each class start-time range. The registrar keys slots
-- on whether a class meets Tuesday or Thursday, not on its exact days.
CREATE TABLE finals_mappings (
    id INTEGER PRIMARY KEY,
    term_code TEXT NOT NULL REFERENCES terms(code) ON DELETE CASCADE,
    time_range_start TEXT NOT NULL,
    time_range_end TEXT NOT NULL, -- inclusive
    has_tuth BOOLEAN NOT NULL,
    exam_date TEXT NOT NULL,
    exam_start_time TEXT NOT NULL,
    exam_end_time TEXT NOT NULL,
    UNIQUE(term_code, time_range_start, has_tuth)
);

-- Every dated row of the registrar's important dates page, holidays included.
CREATE TABLE important_dates (
    id INTEGER PRIMARY KEY,
    term_code TEXT NOT NULL REFERENCES terms(code) ON DELETE CASCADE,
    date TEXT NOT NULL,
    end_date TEXT, -- set when the row is a range, e.g. a registration phase or break
    description TEXT NOT NULL,
    category TEXT NOT NULL, -- academic, registration, deadline, or holiday
    opens_registration BOOLEAN NOT NULL DEFAULT 0,
    UNIQUE(term_code, date, description)
);

-- Term phase detection reads one registration-open date per term.
CREATE UNIQUE INDEX idx_important_dates_registration_open
    ON important_dates(term_code) WHERE opens_registration;
