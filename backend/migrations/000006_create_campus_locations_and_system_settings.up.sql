-- Migration: 000006_create_campus_locations_and_system_settings.up.sql
-- Purpose: Support multi-point geofence (campus_locations) and dynamic system settings

CREATE TABLE IF NOT EXISTS system_settings (
    key        VARCHAR(50) PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS campus_locations (
    id            UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(150)     NOT NULL,
    description   TEXT,
    latitude      DOUBLE PRECISION NOT NULL,
    longitude     DOUBLE PRECISION NOT NULL,
    radius_meters INT              NOT NULL DEFAULT 80,
    is_active     BOOLEAN          NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_campus_locations_active ON campus_locations (is_active);

-- Seed initial campus locations (Titik 1: Kampus Induk Umuslim)
INSERT INTO campus_locations (name, description, latitude, longitude, radius_meters, is_active)
VALUES 
    ('Kampus Induk Umuslim', 'Kampus Utama Matang Glumpang Dua, Bireuen', 5.193730, 96.787492, 80, TRUE),
    ('Kampus B (Kompleks Lab Terpadu)', 'Laboratorium Komputer & Studio Multimedia Terpadu', 5.194120, 96.788100, 70, TRUE)
ON CONFLICT DO NOTHING;

-- Seed system configuration defaults
INSERT INTO system_settings (key, value) VALUES
    ('campus_name', 'Universitas Almuslim'),
    ('campus_tagline', 'Fakultas Ilmu Komputer (FIKOM)'),
    ('primary_color', '#006633'),
    ('accent_color', '#D4AF37'),
    ('logo_url', ''),
    ('campus_api_url', 'http://mock-campus-api:9001'),
    ('campus_api_key', 'mock-api-key-change-me'),
    ('campus_sync_cron', '0 */6 * * *'),
    ('geofence_mode', 'multi_point'),
    ('max_tolerance_minutes', '15')
ON CONFLICT (key) DO NOTHING;
