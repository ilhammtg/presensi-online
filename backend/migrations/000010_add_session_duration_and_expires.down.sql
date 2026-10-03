-- Migration: 000010_add_session_duration_and_expires.down.sql
ALTER TABLE class_sessions DROP COLUMN IF EXISTS expires_at;
ALTER TABLE class_sessions DROP COLUMN IF EXISTS duration_minutes;
