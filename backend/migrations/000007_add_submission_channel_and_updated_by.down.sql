-- Migration: 000007_add_submission_channel_and_updated_by.down.sql

DROP INDEX IF EXISTS idx_attendances_verified;
DROP INDEX IF EXISTS idx_attendances_submission_source;
ALTER TABLE attendances DROP COLUMN IF EXISTS updated_by;
ALTER TABLE attendances DROP COLUMN IF EXISTS submission_source;
DROP TYPE IF EXISTS submission_channel;
