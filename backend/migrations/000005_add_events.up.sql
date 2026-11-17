-- Client-side events the server can't observe from API calls, such as
-- exports and view toggles. The vocabulary is fixed in internal/api/events.go.
CREATE TABLE events (
    id INTEGER PRIMARY KEY,
    session_id TEXT NOT NULL,
    name TEXT NOT NULL,
    target TEXT,
    term TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_created_at ON events(created_at);
CREATE INDEX idx_events_name_created_at ON events(name, created_at);
