-- Migration: 000010_add_session_duration_and_expires.up.sql
ALTER TABLE class_sessions ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;
ALTER TABLE class_sessions ADD COLUMN IF NOT EXISTS duration_minutes INTEGER DEFAULT 30;
