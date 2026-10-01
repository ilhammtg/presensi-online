-- Migration: 000002_create_academic_location_tables.up.sql
-- Purpose: Create faculties, study_programs, buildings, rooms

-- 1. Faculties
CREATE TABLE IF NOT EXISTS faculties (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code       VARCHAR(20) NOT NULL UNIQUE,
    name       VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. Study Programs (Program Studi)
CREATE TABLE IF NOT EXISTS study_programs (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    faculty_id UUID        NOT NULL REFERENCES faculties(id) ON DELETE RESTRICT,
    code       VARCHAR(20) NOT NULL UNIQUE,
    name       VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. Buildings
CREATE TABLE IF NOT EXISTS buildings (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code       VARCHAR(20) NOT NULL UNIQUE,
    name       VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 4. Rooms (with geofence coordinates)
CREATE TABLE IF NOT EXISTS rooms (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    building_id    UUID         NOT NULL REFERENCES buildings(id) ON DELETE RESTRICT,
    room_code      VARCHAR(30)  NOT NULL UNIQUE,
    name           VARCHAR(100) NOT NULL,
    latitude       DOUBLE PRECISION NOT NULL,
    longitude      DOUBLE PRECISION NOT NULL,
    radius_meters  INT          NOT NULL DEFAULT 35,
    is_active      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);
