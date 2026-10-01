-- Migration: 000005_create_sessions_and_attendances.down.sql

DROP INDEX IF EXISTS idx_attendances_student_session;
DROP INDEX IF EXISTS idx_attendances_session_student_lookup;
DROP INDEX IF EXISTS idx_sessions_active_lookup;
DROP TABLE IF EXISTS attendances;
DROP TABLE IF EXISTS class_sessions;
