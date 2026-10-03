-- Migration: 000005_create_sessions_and_attendances.up.sql
-- Purpose: Core transactional tables: class_sessions and attendances

-- 1. Class Sessions (Sesi Absen yang Dibuka Dosen)
CREATE TABLE IF NOT EXISTS class_sessions (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id  UUID        NOT NULL REFERENCES class_schedules(id) ON DELETE RESTRICT,
    meeting_no   SMALLINT    NOT NULL,            -- Pertemuan ke- (1–16)
    session_date DATE        NOT NULL DEFAULT CURRENT_DATE,
    qr_seed      VARCHAR(64) NOT NULL,            -- Seed TOTP untuk QR bergulir
    is_open          BOOLEAN     NOT NULL DEFAULT TRUE,
    opened_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at        TIMESTAMPTZ,
    expires_at       TIMESTAMPTZ,
    duration_minutes INTEGER     DEFAULT 30,
    bap_topic        TEXT,                            -- Catatan topik BAP
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_schedule_meeting UNIQUE (schedule_id, meeting_no)
);

-- Index: pencarian sesi aktif (hot path saat ribuan mahasiswa scan)
CREATE INDEX IF NOT EXISTS idx_sessions_active_lookup
ON class_sessions (schedule_id)
WHERE is_open = TRUE;

-- 2. Attendances (Transaksi Presensi — Core Table)
CREATE TABLE IF NOT EXISTS attendances (
    id                   UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id           UUID             NOT NULL REFERENCES class_sessions(id) ON DELETE CASCADE,
    student_id           UUID             NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status               attendance_status NOT NULL,
    scanned_at           TIMESTAMPTZ,
    device_id            VARCHAR(120),             -- ID HP saat scan (untuk flagging titip absen)
    latitude             DOUBLE PRECISION,
    longitude            DOUBLE PRECISION,
    distance_meters      DOUBLE PRECISION,
    attachment_url       VARCHAR(255),             -- Bukti izin/sakit
    notes                TEXT,
    verified_by_lecturer BOOLEAN NOT NULL DEFAULT TRUE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- Anti-duplikasi presensi: satu mahasiswa satu kali per sesi (race condition ditangani di DB level)
    CONSTRAINT uq_session_student UNIQUE (session_id, student_id)
);

-- Index: validasi instan apakah mahasiswa sudah presensi (step 7 pipeline)
CREATE INDEX IF NOT EXISTS idx_attendances_session_student_lookup
ON attendances (session_id, student_id);

-- Index: rekap per mahasiswa per jadwal (riwayat kehadiran)
CREATE INDEX IF NOT EXISTS idx_attendances_student_session
ON attendances (student_id, session_id);
