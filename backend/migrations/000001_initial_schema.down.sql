-- Rollback initial schema

DROP INDEX IF EXISTS idx_section_attributes_section;
DROP INDEX IF EXISTS idx_meeting_times_section;
DROP INDEX IF EXISTS idx_instructors_section;
DROP INDEX IF EXISTS idx_sections_crn;
DROP INDEX IF EXISTS idx_sections_term_subject;
DROP INDEX IF EXISTS idx_sections_subject;
DROP INDEX IF EXISTS idx_sections_term;

DROP TABLE IF EXISTS generation_logs;
DROP TABLE IF EXISTS search_logs;
DROP TABLE IF EXISTS section_attributes;
DROP TABLE IF EXISTS meeting_times;
DROP TABLE IF EXISTS instructors;
DROP TABLE IF EXISTS sections;
DROP TABLE IF EXISTS terms;
