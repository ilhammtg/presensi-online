-- Migration: 000007_add_submission_channel_and_updated_by.up.sql
-- Purpose: Support dual-channel permission (app_request vs manual_lecturer) and audit logging as per technical spec

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'submission_channel') THEN
        CREATE TYPE submission_channel AS ENUM ('self_scan', 'app_request', 'manual_lecturer');
    END IF;
END $$;

ALTER TABLE attendances
    ADD COLUMN IF NOT EXISTS submission_source submission_channel NOT NULL DEFAULT 'self_scan',
    ADD COLUMN IF NOT EXISTS updated_by UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_attendances_submission_source
    ON attendances (submission_source);

CREATE INDEX IF NOT EXISTS idx_attendances_verified
    ON attendances (verified_by_lecturer);
