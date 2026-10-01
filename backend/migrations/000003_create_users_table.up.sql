-- Migration: 000003_create_users_table.up.sql
-- Purpose: Create users table with device binding support

CREATE TABLE IF NOT EXISTS users (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id   VARCHAR(50)  NOT NULL UNIQUE,  -- NIM (mahasiswa) / NIDN (dosen)
    name          VARCHAR(150) NOT NULL,
    email         VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role          user_role    NOT NULL,
    prodi_id      UUID         REFERENCES study_programs(id) ON DELETE SET NULL,
    device_id     VARCHAR(120),                  -- Device Binding: UUID perangkat aktif
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMPTZ                    -- Soft delete
);

-- Index for role + prodi monitoring (realtime dashboard prodi)
CREATE INDEX IF NOT EXISTS idx_users_prodi_role ON users (prodi_id, role) WHERE is_active = TRUE;
-- Index for fast email lookup (login)
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email) WHERE deleted_at IS NULL;
