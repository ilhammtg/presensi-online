-- Migration: 000004_create_schedules_and_plans.up.sql
-- Purpose: Create class_schedules (master kelas) and study_plans (KRS enrollment)

-- 1. Class Schedules (Master Jadwal)
CREATE TABLE IF NOT EXISTS class_schedules (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id   VARCHAR(50) UNIQUE,            -- ID dari SIAKAD kampus
    course_code   VARCHAR(30) NOT NULL,
    course_name   VARCHAR(150) NOT NULL,
    academic_year VARCHAR(10) NOT NULL,           -- Format: "2026/2027"
    semester_type SMALLINT    NOT NULL,           -- 1=Ganjil, 2=Genap, 3=Pendek
    lecturer_id   UUID        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    room_id       UUID        NOT NULL REFERENCES rooms(id) ON DELETE RESTRICT,
    day_of_week   SMALLINT    NOT NULL,           -- 1 (Senin) - 7 (Minggu)
    start_time    TIME        NOT NULL,
    end_time      TIME        NOT NULL,
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMPTZ                     -- Soft delete
);

-- Index: query jadwal hari ini (beranda mahasiswa/dosen)
CREATE INDEX IF NOT EXISTS idx_schedules_day_time
ON class_schedules (day_of_week, start_time, end_time)
WHERE is_active = TRUE AND deleted_at IS NULL;

-- 2. Study Plans (KRS / Enrollment)
CREATE TABLE IF NOT EXISTS study_plans (
    id          UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id  UUID             NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    schedule_id UUID             NOT NULL REFERENCES class_schedules(id) ON DELETE CASCADE,
    status      enrollment_status NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_student_schedule UNIQUE (student_id, schedule_id)
);

-- Index: load mata kuliah aktif mahasiswa
CREATE INDEX IF NOT EXISTS idx_study_plans_active_student
ON study_plans (student_id)
WHERE status = 'active';
