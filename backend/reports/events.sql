-- Usage reports over the events table, read-only. From the repo root:
--
--   ssh hetz "sqlite3 -readonly -box ~/schedule-optimizer/backend/data/schedule.db" < backend/reports/events.sql
--
-- Each query covers today and the 30 whole days before it; add
-- -cmd '.parameter set @days 90' to the sqlite3 command for another window.
-- Days are Pacific standard time (SQLite has no time zones), so during
-- daylight time events between midnight and 1 a.m. land on the previous day. A session is one browser's
-- localStorage, not a person: clearing storage or switching browsers starts
-- a new one, and so does every render by a crawler that runs JavaScript.

.print 'Visits per day (distinct sessions)'
SELECT date(created_at, '-8 hours') AS day,
       COUNT(DISTINCT session_id) AS sessions,
       COUNT(DISTINCT CASE WHEN target = 'desktop' THEN session_id END) AS desktop,
       COUNT(DISTINCT CASE WHEN target = 'mobile' THEN session_id END) AS mobile
FROM events
WHERE name = 'visit'
  AND created_at >= datetime(date('now', '-8 hours', '-' || coalesce(@days, 30) || ' days'), '+8 hours')
GROUP BY day
ORDER BY day;

-- A session is new on the day of its first recorded visit, so sessions
-- older than the events table count as new on their first visit after it.
.print ''
.print 'New vs. returning sessions per day'
WITH first_visit AS (
    SELECT session_id, date(MIN(created_at), '-8 hours') AS first_day
    FROM events
    WHERE name = 'visit'
    GROUP BY session_id
)
SELECT date(e.created_at, '-8 hours') AS day,
       COUNT(DISTINCT CASE WHEN f.first_day = date(e.created_at, '-8 hours') THEN e.session_id END) AS new_sessions,
       COUNT(DISTINCT CASE WHEN f.first_day < date(e.created_at, '-8 hours') THEN e.session_id END) AS returning_sessions
FROM events e
JOIN first_visit f USING (session_id)
WHERE e.name = 'visit'
  AND e.created_at >= datetime(date('now', '-8 hours', '-' || coalesce(@days, 30) || ' days'), '+8 hours')
GROUP BY day
ORDER BY day;

.print ''
.print 'Events by name and target'
SELECT name,
       coalesce(target, '') AS target,
       COUNT(*) AS events,
       COUNT(DISTINCT session_id) AS sessions
FROM events
WHERE created_at >= datetime(date('now', '-8 hours', '-' || coalesce(@days, 30) || ' days'), '+8 hours')
GROUP BY name, target
ORDER BY name, events DESC;

.print ''
.print '.ics exports including each option'
WITH exports AS (
    SELECT ',' || target || ',' AS options
    FROM events
    WHERE name = 'ics_export'
      AND created_at >= datetime(date('now', '-8 hours', '-' || coalesce(@days, 30) || ' days'), '+8 hours')
),
ics_options(ord, option) AS (
    VALUES (1, 'classes'), (2, 'finals'), (3, 'holidays'),
           (4, 'reg1'), (5, 'reg2'), (6, 'reg3'), (7, 'deadlines')
)
SELECT o.option,
       SUM(instr(e.options, ',' || o.option || ',') > 0) AS exports,
       round(100.0 * SUM(instr(e.options, ',' || o.option || ',') > 0) / COUNT(*), 1) AS pct
FROM ics_options o
CROSS JOIN exports e
GROUP BY o.ord
ORDER BY o.ord;
