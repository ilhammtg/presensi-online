-- ============================================================
-- SUPABASE MIGRATION: Presensi Online - Universitas Almuslim
-- Run this in Supabase SQL Editor (Dashboard > SQL Editor)
-- ============================================================

-- 1. Extensions & Enums
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

DO $$ BEGIN
  CREATE TYPE user_role AS ENUM ('mahasiswa','dosen','admin_prodi','pimpinan','superadmin');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE attendance_status AS ENUM ('hadir','terlambat','izin','sakit','alpa');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE enrollment_status AS ENUM ('active','dropped','withdrawn');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- submission_channel: hapus 'app_request' (mahasiswa tidak submit izin via app)
DO $$ BEGIN
  CREATE TYPE submission_channel AS ENUM ('self_scan','manual_lecturer');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- 2. Faculties
CREATE TABLE IF NOT EXISTS faculties (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code       VARCHAR(20) NOT NULL UNIQUE,
  name       VARCHAR(150) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Study Programs
CREATE TABLE IF NOT EXISTS study_programs (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  faculty_id UUID NOT NULL REFERENCES faculties(id) ON DELETE RESTRICT,
  code       VARCHAR(20) NOT NULL UNIQUE,
  name       VARCHAR(150) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Buildings
CREATE TABLE IF NOT EXISTS buildings (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code       VARCHAR(20) NOT NULL UNIQUE,
  name       VARCHAR(150) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. Rooms
CREATE TABLE IF NOT EXISTS rooms (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  building_id   UUID NOT NULL REFERENCES buildings(id) ON DELETE RESTRICT,
  room_code     VARCHAR(30) NOT NULL UNIQUE,
  name          VARCHAR(100) NOT NULL,
  latitude      DOUBLE PRECISION NOT NULL DEFAULT 5.193730,
  longitude     DOUBLE PRECISION NOT NULL DEFAULT 96.787492,
  radius_meters INT NOT NULL DEFAULT 35,
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. Users
CREATE TABLE IF NOT EXISTS users (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  external_id   VARCHAR(50) NOT NULL UNIQUE,
  name          VARCHAR(150) NOT NULL,
  email         VARCHAR(100) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role          user_role NOT NULL,
  prodi_id      UUID REFERENCES study_programs(id) ON DELETE SET NULL,
  device_id     VARCHAR(120),
  avatar_url    VARCHAR(500),
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_prodi_role ON users (prodi_id, role) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email) WHERE deleted_at IS NULL;

-- 7. Class Schedules
CREATE TABLE IF NOT EXISTS class_schedules (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  external_id   VARCHAR(50) UNIQUE,
  course_code   VARCHAR(30) NOT NULL,
  course_name   VARCHAR(150) NOT NULL,
  academic_year VARCHAR(10) NOT NULL DEFAULT '2026/2027',
  semester_type SMALLINT NOT NULL DEFAULT 1,
  lecturer_id   UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  room_id       UUID NOT NULL REFERENCES rooms(id) ON DELETE RESTRICT,
  day_of_week   SMALLINT NOT NULL,  -- 1=Senin .. 7=Minggu
  start_time    TIME NOT NULL,
  end_time      TIME NOT NULL,
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_schedules_day_time
  ON class_schedules (day_of_week, start_time, end_time)
  WHERE is_active = TRUE AND deleted_at IS NULL;

-- 8. Study Plans (KRS)
CREATE TABLE IF NOT EXISTS study_plans (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  student_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  schedule_id UUID NOT NULL REFERENCES class_schedules(id) ON DELETE CASCADE,
  status      enrollment_status NOT NULL DEFAULT 'active',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_student_schedule UNIQUE (student_id, schedule_id)
);

CREATE INDEX IF NOT EXISTS idx_study_plans_active_student
  ON study_plans (student_id) WHERE status = 'active';

-- 9. Class Sessions
CREATE TABLE IF NOT EXISTS class_sessions (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  schedule_id  UUID NOT NULL REFERENCES class_schedules(id) ON DELETE RESTRICT,
  meeting_no   SMALLINT NOT NULL,
  session_date DATE NOT NULL DEFAULT CURRENT_DATE,
  qr_seed      VARCHAR(64) NOT NULL,
  is_open      BOOLEAN NOT NULL DEFAULT TRUE,
  opened_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  closed_at    TIMESTAMPTZ,
  bap_topic    TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_schedule_meeting UNIQUE (schedule_id, meeting_no)
);

CREATE INDEX IF NOT EXISTS idx_sessions_active_lookup
  ON class_sessions (schedule_id) WHERE is_open = TRUE;

-- 10. Attendances
CREATE TABLE IF NOT EXISTS attendances (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id           UUID NOT NULL REFERENCES class_sessions(id) ON DELETE CASCADE,
  student_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status               attendance_status NOT NULL,
  submission_source    submission_channel NOT NULL DEFAULT 'self_scan',
  scanned_at           TIMESTAMPTZ,
  device_id            VARCHAR(120),
  latitude             DOUBLE PRECISION,
  longitude            DOUBLE PRECISION,
  distance_meters      DOUBLE PRECISION,
  attachment_url       VARCHAR(255),
  notes                TEXT,
  verified_by_lecturer BOOLEAN NOT NULL DEFAULT TRUE,
  updated_by           UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_session_student UNIQUE (session_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_attendances_session_student ON attendances (session_id, student_id);
CREATE INDEX IF NOT EXISTS idx_attendances_student_session ON attendances (student_id, session_id);
CREATE INDEX IF NOT EXISTS idx_attendances_submission_source ON attendances (submission_source);

-- 11. System Settings
CREATE TABLE IF NOT EXISTS system_settings (
  key        VARCHAR(50) PRIMARY KEY,
  value      TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 12. Campus Locations
CREATE TABLE IF NOT EXISTS campus_locations (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name          VARCHAR(150) NOT NULL,
  description   TEXT,
  latitude      DOUBLE PRECISION NOT NULL,
  longitude     DOUBLE PRECISION NOT NULL,
  radius_meters INT NOT NULL DEFAULT 80,
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_campus_locations_active ON campus_locations (is_active);

-- ============================================================
-- SEED DATA
-- ============================================================

-- Seed: System Settings
INSERT INTO system_settings (key, value) VALUES
  ('campus_name', 'Universitas Almuslim'),
  ('campus_tagline', 'Fakultas Ilmu Komputer (FIKOM)'),
  ('primary_color', '#006633'),
  ('accent_color', '#D4AF37'),
  ('logo_url', ''),
  ('geofence_mode', 'multi_point'),
  ('max_tolerance_minutes', '15'),
  ('qr_totp_period_seconds', '15'),
  ('app_version', '1.0.0')
ON CONFLICT (key) DO NOTHING;

-- Seed: Campus Locations
INSERT INTO campus_locations (name, description, latitude, longitude, radius_meters, is_active) VALUES
  ('Kampus Induk Umuslim', 'Kampus Utama Matang Glumpang Dua, Bireuen', 5.193730, 96.787492, 80, TRUE),
  ('Kampus B (Lab Terpadu)', 'Laboratorium Komputer & Studio Multimedia Terpadu', 5.194120, 96.788100, 70, TRUE)
ON CONFLICT DO NOTHING;

-- Seed: Faculty
INSERT INTO faculties (id, code, name) VALUES
  ('11111111-0000-0000-0000-000000000001', 'FIKOM', 'Fakultas Ilmu Komputer')
ON CONFLICT (code) DO NOTHING;

-- Seed: Study Programs
INSERT INTO study_programs (id, faculty_id, code, name) VALUES
  ('22222222-0000-0000-0000-000000000001', '11111111-0000-0000-0000-000000000001', 'INF', 'Informatika'),
  ('22222222-0000-0000-0000-000000000002', '11111111-0000-0000-0000-000000000001', 'SK',  'Sistem Komputer'),
  ('22222222-0000-0000-0000-000000000003', '11111111-0000-0000-0000-000000000001', 'IM',  'Informatika Medis'),
  ('22222222-0000-0000-0000-000000000004', '11111111-0000-0000-0000-000000000001', 'BD',  'Bisnis Digital')
ON CONFLICT (code) DO NOTHING;

-- Seed: Buildings
INSERT INTO buildings (id, code, name) VALUES
  ('33333333-0000-0000-0000-000000000001', 'GD-A', 'Gedung A (FIKOM)'),
  ('33333333-0000-0000-0000-000000000002', 'LAB',  'Gedung Lab Komputer')
ON CONFLICT (code) DO NOTHING;

-- Seed: Rooms
INSERT INTO rooms (id, building_id, room_code, name, latitude, longitude, radius_meters) VALUES
  ('44444444-0000-0000-0000-000000000001', '33333333-0000-0000-0000-000000000002', 'LAB-K1', 'Lab Komputer 1', 5.193730, 96.787492, 40),
  ('44444444-0000-0000-0000-000000000002', '33333333-0000-0000-0000-000000000002', 'LAB-K2', 'Lab Komputer 2', 5.194000, 96.787700, 40),
  ('44444444-0000-0000-0000-000000000003', '33333333-0000-0000-0000-000000000001', 'R-101',  'Ruang Teori 101', 5.193500, 96.787200, 35),
  ('44444444-0000-0000-0000-000000000004', '33333333-0000-0000-0000-000000000001', 'R-102',  'Ruang Teori 102', 5.193600, 96.787300, 35),
  ('44444444-0000-0000-0000-000000000005', '33333333-0000-0000-0000-000000000001', 'R-204',  'Ruang Teori 204', 5.193800, 96.787500, 35),
  ('44444444-0000-0000-0000-000000000006', '33333333-0000-0000-0000-000000000002', 'LAB-DB', 'Lab Database',    5.194200, 96.788000, 40),
  ('44444444-0000-0000-0000-000000000007', '33333333-0000-0000-0000-000000000002', 'LAB-JRG','Lab Jaringan',    5.194100, 96.787900, 40),
  ('44444444-0000-0000-0000-000000000008', '33333333-0000-0000-0000-000000000002', 'LAB-MM', 'Lab Multimedia',  5.194300, 96.788100, 40)
ON CONFLICT (room_code) DO NOTHING;

-- Seed: Lecturers (password = bcrypt of 'password', cost 10)
-- Bcrypt hash for 'password': $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy
INSERT INTO users (id, external_id, name, email, password_hash, role, prodi_id) VALUES
  ('aaaaaaaa-0000-0000-0000-000000000001', '0001018501', 'Afriana, SE., MM',
   'afriana.mm@kampus.ac.id', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
   'dosen', '22222222-0000-0000-0000-000000000001'),
  ('aaaaaaaa-0000-0000-0000-000000000002', '0002018502', 'Al Khaidar, M.Kom',
   'al.mkom@kampus.ac.id', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
   'dosen', '22222222-0000-0000-0000-000000000001'),
  ('aaaaaaaa-0000-0000-0000-000000000003', '0012018502', 'Fitri Rizani, M.Kom',
   'fitri.mkom@kampus.ac.id', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
   'dosen', '22222222-0000-0000-0000-000000000001'),
  ('aaaaaaaa-0000-0000-0000-000000000004', '0013018503', 'Hannan Asrawi, M.Kom',
   'hannan.mkom@kampus.ac.id', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
   'dosen', '22222222-0000-0000-0000-000000000001'),
  ('aaaaaaaa-0000-0000-0000-000000000005', '0016018506', 'Ikramullah, M.Kom',
   'ikramullah.mkom@kampus.ac.id', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
   'dosen', '22222222-0000-0000-0000-000000000001')
ON CONFLICT (external_id) DO NOTHING;

-- Admin Prodi
INSERT INTO users (id, external_id, name, email, password_hash, role, prodi_id) VALUES
  ('aaaaaaaa-0000-0000-0000-000000000010', 'ADM-INF-01', 'Admin Prodi Informatika',
   'admin.inf@kampus.ac.id', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
   'admin_prodi', '22222222-0000-0000-0000-000000000001'),
  ('aaaaaaaa-0000-0000-0000-000000000011', 'ADM-IM-01', 'Admin Prodi Informatika Medis',
   'admin.im@kampus.ac.id', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
   'admin_prodi', '22222222-0000-0000-0000-000000000003')
ON CONFLICT (external_id) DO NOTHING;

-- Seed: Students (20 mahasiswa INF)
INSERT INTO users (external_id, name, email, password_hash, role, prodi_id) VALUES
  ('26552020001','Rizki Aulia','26552020001@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020002','Diki Maulida','26552020002@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020003','Siti Fitriani','26552020003@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020004','Putri Wardani','26552020004@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020005','Muhammad Farhan','26552020005@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020006','Cut Nurhaliza','26552020006@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020007','Teuku Raihan','26552020007@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020008','Zulfa Zahira','26552020008@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020009','Irfan Maulana','26552020009@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020010','Nurul Aini','26552020010@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020011','Azhar Ramadhan','26552020011@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020012','Nadia Safira','26552020012@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020013','Hafizuddin','26552020013@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020014','Rahmi Fitria','26552020014@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020015','Zikri Fauzan','26552020015@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020016','Maisarah','26552020016@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020017','Syarifuddin','26552020017@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020018','Khairina Putri','26552020018@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020019','Mukhlis Fadli','26552020019@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001'),
  ('26552020020','Alya Nazira','26552020020@mhs.kampus.ac.id','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','mahasiswa','22222222-0000-0000-0000-000000000001')
ON CONFLICT (external_id) DO NOTHING;

-- Seed: Class Schedules
-- day_of_week: 1=Senin, 2=Selasa, 3=Rabu, 4=Kamis, 5=Jumat
INSERT INTO class_schedules (id, external_id, course_code, course_name, academic_year, semester_type, lecturer_id, room_id, day_of_week, start_time, end_time) VALUES
  ('55555555-0000-0000-0000-000000000001', 'SCH-IF203-A', 'IF203', 'Algoritma dan Pemrograman', '2026/2027', 1,
   'aaaaaaaa-0000-0000-0000-000000000001', '44444444-0000-0000-0000-000000000002', 1, '08:00', '09:40'),
  ('55555555-0000-0000-0000-000000000002', 'SCH-IF203-B', 'IF203-B', 'Algoritma dan Pemrograman Kelas B', '2026/2027', 1,
   'aaaaaaaa-0000-0000-0000-000000000001', '44444444-0000-0000-0000-000000000001', 1, '13:30', '15:10'),
  ('55555555-0000-0000-0000-000000000003', 'SCH-IF204', 'IF204', 'Struktur Data & Algoritma', '2026/2027', 1,
   'aaaaaaaa-0000-0000-0000-000000000002', '44444444-0000-0000-0000-000000000004', 1, '10:00', '11:40'),
  ('55555555-0000-0000-0000-000000000004', 'SCH-IF205', 'IF205', 'Basis Data Terdistribusi', '2026/2027', 1,
   'aaaaaaaa-0000-0000-0000-000000000003', '44444444-0000-0000-0000-000000000006', 2, '08:00', '10:30'),
  ('55555555-0000-0000-0000-000000000005', 'SCH-IF206', 'IF206', 'Jaringan Komputer', '2026/2027', 1,
   'aaaaaaaa-0000-0000-0000-000000000004', '44444444-0000-0000-0000-000000000007', 3, '13:00', '15:30'),
  ('55555555-0000-0000-0000-000000000006', 'SCH-IF207', 'IF207', 'Pemrograman Web & Mobile', '2026/2027', 1,
   'aaaaaaaa-0000-0000-0000-000000000005', '44444444-0000-0000-0000-000000000008', 4, '08:00', '10:30'),
  ('55555555-0000-0000-0000-000000000007', 'SCH-IF305', 'IF305', 'Sistem Informasi Manajemen', '2026/2027', 1,
   'aaaaaaaa-0000-0000-0000-000000000001', '44444444-0000-0000-0000-000000000005', 4, '10:00', '11:40')
ON CONFLICT (external_id) DO NOTHING;

-- Seed: Study Plans (Enroll all 20 students to the main courses)
DO $$
DECLARE
  student RECORD;
  sched_ids UUID[] := ARRAY[
    '55555555-0000-0000-0000-000000000001'::UUID,
    '55555555-0000-0000-0000-000000000003'::UUID,
    '55555555-0000-0000-0000-000000000004'::UUID,
    '55555555-0000-0000-0000-000000000005'::UUID,
    '55555555-0000-0000-0000-000000000006'::UUID,
    '55555555-0000-0000-0000-000000000007'::UUID
  ];
  sid UUID;
BEGIN
  FOR student IN SELECT id FROM users WHERE role = 'mahasiswa' LOOP
    FOREACH sid IN ARRAY sched_ids LOOP
      INSERT INTO study_plans (student_id, schedule_id, status)
      VALUES (student.id, sid, 'active')
      ON CONFLICT (student_id, schedule_id) DO NOTHING;
    END LOOP;
  END LOOP;
END $$;

-- Enable Row Level Security (RLS) - disabled for backend service access
-- The Go backend uses the postgres connection directly with service role,
-- so RLS is not needed for backend. Enable only if using Supabase client directly.
-- ALTER TABLE users ENABLE ROW LEVEL SECURITY;

SELECT 'Migration completed successfully!' as status;
SELECT 'Users: ' || COUNT(*)::text FROM users;
SELECT 'Schedules: ' || COUNT(*)::text FROM class_schedules;
SELECT 'Study Plans: ' || COUNT(*)::text FROM study_plans;
