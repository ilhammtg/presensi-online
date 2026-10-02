-- Migration: 000008_seed_initial_data.up.sql
-- Purpose: Initial seed data for faculties, study programs, buildings, rooms, users, and schedules

-- 1. Faculties
INSERT INTO faculties (id, code, name) VALUES
  ('11111111-0000-0000-0000-000000000001', 'FIKOM', 'Fakultas Ilmu Komputer'),
  ('11111111-0000-0000-0000-000000000002', 'FKIP', 'Fakultas Keguruan dan Ilmu Pendidikan'),
  ('11111111-0000-0000-0000-000000000003', 'FT', 'Fakultas Teknik')
ON CONFLICT (code) DO NOTHING;

-- 2. Study Programs
INSERT INTO study_programs (id, faculty_id, code, name) VALUES
  ('22222222-0000-0000-0000-000000000001', COALESCE((SELECT id FROM faculties WHERE code = 'FIKOM' LIMIT 1), '11111111-0000-0000-0000-000000000001'), 'INF', 'S1 Informatika'),
  ('22222222-0000-0000-0000-000000000002', COALESCE((SELECT id FROM faculties WHERE code = 'FIKOM' LIMIT 1), '11111111-0000-0000-0000-000000000001'), 'SI', 'S1 Sistem Informasi'),
  ('22222222-0000-0000-0000-000000000003', COALESCE((SELECT id FROM faculties WHERE code = 'FIKOM' LIMIT 1), '11111111-0000-0000-0000-000000000001'), 'IM', 'S1 Informatika Medis')
ON CONFLICT (code) DO NOTHING;

-- 3. Buildings
INSERT INTO buildings (id, code, name) VALUES
  ('33333333-0000-0000-0000-000000000001', 'GEDUNG-A', 'Gedung Rektorat & Kuliah A'),
  ('33333333-0000-0000-0000-000000000002', 'GEDUNG-FIKOM', 'Gedung Fakultas Ilmu Komputer')
ON CONFLICT (code) DO NOTHING;

-- 4. Rooms (Gunakan subquery untuk building_id agar aman jika building sudah ada dengan UUID berbeda)
INSERT INTO rooms (id, building_id, room_code, name, latitude, longitude, radius_meters) VALUES
  ('44444444-0000-0000-0000-000000000001', COALESCE((SELECT id FROM buildings WHERE code = 'GEDUNG-FIKOM' LIMIT 1), '33333333-0000-0000-0000-000000000002'), 'LAB-K1', 'Lab Komputer 1', 5.193730, 96.787492, 40),
  ('44444444-0000-0000-0000-000000000002', COALESCE((SELECT id FROM buildings WHERE code = 'GEDUNG-FIKOM' LIMIT 1), '33333333-0000-0000-0000-000000000002'), 'LAB-K2', 'Lab Komputer 2', 5.194000, 96.787700, 40),
  ('44444444-0000-0000-0000-000000000003', COALESCE((SELECT id FROM buildings WHERE code = 'GEDUNG-A' LIMIT 1),     '33333333-0000-0000-0000-000000000001'), 'R-101',  'Ruang Teori 101', 5.193500, 96.787200, 35),
  ('44444444-0000-0000-0000-000000000004', COALESCE((SELECT id FROM buildings WHERE code = 'GEDUNG-A' LIMIT 1),     '33333333-0000-0000-0000-000000000001'), 'R-102',  'Ruang Teori 102', 5.193600, 96.787300, 35),
  ('44444444-0000-0000-0000-000000000005', COALESCE((SELECT id FROM buildings WHERE code = 'GEDUNG-A' LIMIT 1),     '33333333-0000-0000-0000-000000000001'), 'R-204',  'Ruang Teori 204', 5.193800, 96.787500, 35),
  ('44444444-0000-0000-0000-000000000006', COALESCE((SELECT id FROM buildings WHERE code = 'GEDUNG-FIKOM' LIMIT 1), '33333333-0000-0000-0000-000000000002'), 'LAB-DB', 'Lab Database',    5.194200, 96.788000, 40),
  ('44444444-0000-0000-0000-000000000007', COALESCE((SELECT id FROM buildings WHERE code = 'GEDUNG-FIKOM' LIMIT 1), '33333333-0000-0000-0000-000000000002'), 'LAB-JRG','Lab Jaringan',    5.194100, 96.787900, 40),
  ('44444444-0000-0000-0000-000000000008', COALESCE((SELECT id FROM buildings WHERE code = 'GEDUNG-FIKOM' LIMIT 1), '33333333-0000-0000-0000-000000000002'), 'LAB-MM', 'Lab Multimedia',  5.194300, 96.788100, 40)
ON CONFLICT (room_code) DO NOTHING;

-- 5. System Admin & Superadmin Users (Default password: "password" -> bcrypt hash $2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi)
INSERT INTO users (id, external_id, name, email, password_hash, role, prodi_id) VALUES
  ('aaaaaaaa-0000-0000-0000-000000000099', 'SUPERADMIN01', 'Administrator Sistem',
   'superadmin@kampus.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
   'superadmin', NULL),
  ('aaaaaaaa-0000-0000-0000-000000000010', 'ADM-001', 'Admin Prodi Informatika',
   'admin.ti@kampus.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
   'admin_prodi', COALESCE((SELECT id FROM study_programs WHERE code = 'INF' LIMIT 1), '22222222-0000-0000-0000-000000000001')),
  ('aaaaaaaa-0000-0000-0000-000000000011', 'ADM-IM-01', 'Admin Prodi Informatika Medis',
   'admin.im@kampus.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
   'admin_prodi', COALESCE((SELECT id FROM study_programs WHERE code = 'IM' LIMIT 1), '22222222-0000-0000-0000-000000000003'))
ON CONFLICT (external_id) DO UPDATE SET
  name = EXCLUDED.name,
  email = EXCLUDED.email,
  password_hash = EXCLUDED.password_hash;

-- 6. Dosen (Lecturers) — Password default: "password"
INSERT INTO users (id, external_id, name, email, password_hash, role, prodi_id) VALUES
  ('aaaaaaaa-0000-0000-0000-000000000001', '0001018501', 'Afriana, SE., MM',
   'afriana.mm@kampus.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
   'dosen', COALESCE((SELECT id FROM study_programs WHERE code = 'INF' LIMIT 1), '22222222-0000-0000-0000-000000000001')),
  ('aaaaaaaa-0000-0000-0000-000000000002', '0002018502', 'Al Khaidar, M.Kom',
   'al.mkom@kampus.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
   'dosen', COALESCE((SELECT id FROM study_programs WHERE code = 'INF' LIMIT 1), '22222222-0000-0000-0000-000000000001')),
  ('aaaaaaaa-0000-0000-0000-000000000003', '0012018502', 'Fitri Rizani, M.Kom',
   'fitri.mkom@kampus.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
   'dosen', COALESCE((SELECT id FROM study_programs WHERE code = 'INF' LIMIT 1), '22222222-0000-0000-0000-000000000001')),
  ('aaaaaaaa-0000-0000-0000-000000000004', '0013018503', 'Hannan Asrawi, M.Kom',
   'hannan.mkom@kampus.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
   'dosen', COALESCE((SELECT id FROM study_programs WHERE code = 'INF' LIMIT 1), '22222222-0000-0000-0000-000000000001')),
  ('aaaaaaaa-0000-0000-0000-000000000005', '0016018506', 'Ikramullah, M.Kom',
   'ikramullah.mkom@kampus.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
   'dosen', COALESCE((SELECT id FROM study_programs WHERE code = 'INF' LIMIT 1), '22222222-0000-0000-0000-000000000001'))
ON CONFLICT (external_id) DO UPDATE SET
  name = EXCLUDED.name,
  email = EXCLUDED.email,
  password_hash = EXCLUDED.password_hash;

-- 7. Mahasiswa (Students) — Password default: "password"
INSERT INTO users (external_id, name, email, password_hash, role, prodi_id) VALUES
  ('26552020001','Rizki Aulia','26552020001@mhs.kampus.ac.id','$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi','mahasiswa', COALESCE((SELECT id FROM study_programs WHERE code = 'INF' LIMIT 1), '22222222-0000-0000-0000-000000000001')),
  ('26552020002','Diki Maulida','26552020002@mhs.kampus.ac.id','$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi','mahasiswa', COALESCE((SELECT id FROM study_programs WHERE code = 'INF' LIMIT 1), '22222222-0000-0000-0000-000000000001')),
  ('26552020003','Siti Fitriani','26552020003@mhs.kampus.ac.id','$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi','mahasiswa', COALESCE((SELECT id FROM study_programs WHERE code = 'INF' LIMIT 1), '22222222-0000-0000-0000-000000000001')),
  ('26552020004','Putri Wardani','26552020004@mhs.kampus.ac.id','$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi','mahasiswa', COALESCE((SELECT id FROM study_programs WHERE code = 'INF' LIMIT 1), '22222222-0000-0000-0000-000000000001')),
  ('26552020005','Muhammad Farhan','26552020005@mhs.kampus.ac.id','$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi','mahasiswa', COALESCE((SELECT id FROM study_programs WHERE code = 'INF' LIMIT 1), '22222222-0000-0000-0000-000000000001'))
ON CONFLICT (external_id) DO UPDATE SET
  name = EXCLUDED.name,
  email = EXCLUDED.email,
  password_hash = EXCLUDED.password_hash;

-- 8. Class Schedules
INSERT INTO class_schedules (id, external_id, course_code, course_name, academic_year, semester_type, lecturer_id, room_id, day_of_week, start_time, end_time) VALUES
  ('55555555-0000-0000-0000-000000000001', 'SCH-IF203-A', 'IF203', 'Algoritma dan Pemrograman', '2026/2027', 1,
   COALESCE((SELECT id FROM users WHERE external_id = '0001018501' LIMIT 1), 'aaaaaaaa-0000-0000-0000-000000000001'),
   COALESCE((SELECT id FROM rooms WHERE room_code = 'LAB-K2' LIMIT 1), '44444444-0000-0000-0000-000000000002'),
   1, '08:00', '09:40'),
  ('55555555-0000-0000-0000-000000000002', 'SCH-IF203-B', 'IF203-B', 'Algoritma dan Pemrograman Kelas B', '2026/2027', 1,
   COALESCE((SELECT id FROM users WHERE external_id = '0001018501' LIMIT 1), 'aaaaaaaa-0000-0000-0000-000000000001'),
   COALESCE((SELECT id FROM rooms WHERE room_code = 'LAB-K1' LIMIT 1), '44444444-0000-0000-0000-000000000001'),
   1, '13:30', '15:10'),
  ('55555555-0000-0000-0000-000000000003', 'SCH-IF204', 'IF204', 'Struktur Data & Algoritma', '2026/2027', 1,
   COALESCE((SELECT id FROM users WHERE external_id = '0002018502' LIMIT 1), 'aaaaaaaa-0000-0000-0000-000000000002'),
   COALESCE((SELECT id FROM rooms WHERE room_code = 'R-102' LIMIT 1), '44444444-0000-0000-0000-000000000004'),
   1, '10:00', '11:40')
ON CONFLICT (external_id) DO NOTHING;
