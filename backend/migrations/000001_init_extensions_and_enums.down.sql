-- Migration: 000001_init_extensions_and_enums.down.sql

DROP TYPE IF EXISTS enrollment_status;
DROP TYPE IF EXISTS attendance_status;
DROP TYPE IF EXISTS user_role;
DROP EXTENSION IF EXISTS "pgcrypto";
