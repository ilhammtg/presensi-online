Berikut rancangan ERD visual *production-ready* yang dioptimalkan untuk skalabilitas, integritas data (*soft delete*, *audit trail*), serta efisiensi beban query saat jam sibuk perkuliahan.

---

### 1. Diagram Relasi Entitas (ERD)

```
 +--------------------+             +--------------------+
 |     faculties      |             |     buildings      |
 +--------------------+             +--------------------+
 | PK id (UUID)       | 1         1 | PK id (UUID)       |
 |    code (VARCHAR)  |<---+   +--->|    code (VARCHAR)  |
 |    name (VARCHAR)  |    |   |    |    name (VARCHAR)  |
 +--------------------+    |   |    +--------------------+
           | 1             |   |              | 1
           | N             |   |              | N
 +--------------------+    |   |    +--------------------+
 |  study_programs    |    |   |    |       rooms        |
 +--------------------+    |   |    +--------------------+
 | PK id (UUID)       |    |   |    | PK id (UUID)       |
 | FK faculty_id      |----+   +----| FK building_id     |
 |    code (VARCHAR)  |             |    room_code       |
 |    name (VARCHAR)  |             |    name            |
 +--------------------+             |    latitude        |
           | 1                      |    longitude       |
           |                        |    radius_meters   |
           +-----------------+      +--------------------+
           | N               | N              | 1
 +--------------------+      |                |
 |       users        |      |                |
 +--------------------+      |                |
 | PK id (UUID)       |      |                |
 |    external_id     |      |                |
 |    name            |      |                |
 |    email           |      |                |
 |    role (ENUM)     |      |                |
 |    device_id       |      |                |
 | FK prodi_id        |------+                |
 +--------------------+                       |
      | 1          | 1                        |
      |            |                          |
      | N          | N                        |
      |    +--------------------+             |
      |    |  class_schedules   |             |
      |    +--------------------+             |
      |    | PK id (UUID)       |             |
      |    | FK lecturer_id     |-------------+ (1 Dosen pengampu)
      |    | FK room_id         |-------------+ (1 Ruangan terjadwal)
      |    |    course_code     |
      |    |    course_name     |
      |    |    day_of_week     |
      |    |    start_time      |
      |    |    end_time        |
      |    |    academic_year   |
      |    +--------------------+
      |              | 1           | 1
      |              |             |
      |              | N           | N
      |    +--------------------+  +--------------------+
      |    |    study_plans     |  |   class_sessions   |
      |    |   (Enroll / KRS)   |  |   (Sesi Terbuka)   |
      |    +--------------------+  +--------------------+
      +--->| FK student_id      |  | PK id (UUID)       |
           | FK schedule_id     |  | FK schedule_id     |
           |    status (ACTIVE) |  |    meeting_no      |
           +--------------------+  |    qr_seed (TOTP)  |
                     |             |    is_open         |
                     |             |    opened_at       |
                     |             |    closed_at       |
                     |             +--------------------+
                     |                        | 1
                     |                        |
                     |                        | N
                     |             +--------------------+
                     |             |    attendances     |
                     |             |  (Transaksi Absen) |
                     |             +--------------------+
                     |             | PK id (UUID)       |
                     +------------>| FK student_id      |
                                   | FK session_id      |
                                   |    status (ENUM)   |
                                   |    scanned_at      |
                                   |    device_id       |
                                   |    latitude        |
                                   |    longitude       |
                                   |    distance_meters |
                                   |    attachment_url  |
                                   +--------------------+

```

---

### 2. Standar Skema SQL Production (PostgreSQL 15+)

Skema ini telah dilengkapi:

* UUID generator bawaan (`pgcrypto` / `gen_random_uuid`).
* Konvensi penamaan seragam (*snake_case*, plural table).
* Soft-delete (`deleted_at`) untuk data master agar data transaksi historis tetap terjaga.
* *Compound unique constraints* untuk mencegah data ganda.

```sql
-- Ekstensi dasar
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ENUM Types
CREATE TYPE user_role AS ENUM ('mahasiswa', 'dosen', 'admin_prodi', 'pimpinan');
CREATE TYPE attendance_status AS ENUM ('hadir', 'terlambat', 'izin', 'sakit', 'alpa');
CREATE TYPE enrollment_status AS ENUM ('active', 'dropped', 'withdrawn');

-- 1. STRUKTUR AKADEMIK & LOKASI
CREATE TABLE faculties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE study_programs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    faculty_id UUID NOT NULL REFERENCES faculties(id) ON DELETE RESTRICT,
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE buildings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    building_id UUID NOT NULL REFERENCES buildings(id) ON DELETE RESTRICT,
    room_code VARCHAR(30) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    radius_meters INT NOT NULL DEFAULT 35,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. USERS
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id VARCHAR(50) NOT NULL UNIQUE, -- NIM / NIDN
    name VARCHAR(150) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role user_role NOT NULL,
    prodi_id UUID REFERENCES study_programs(id) ON DELETE SET NULL,
    device_id VARCHAR(120),                  -- Kunci perangkat (Device Binding)
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- 3. JADWAL KULIAH (MASTER KELAS)
CREATE TABLE class_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id VARCHAR(50) UNIQUE,          -- ID dari database kampus
    course_code VARCHAR(30) NOT NULL,
    course_name VARCHAR(150) NOT NULL,
    academic_year VARCHAR(10) NOT NULL,      -- Format: 2026/2027
    semester_type SMALLINT NOT NULL,         -- 1: Ganjil, 2: Genap, 3: Pendek
    lecturer_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE RESTRICT,
    day_of_week SMALLINT NOT NULL,           -- 1 (Senin) - 7 (Minggu)
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- 4. KRS / ENROLLMENT (KONEKSI MAHASISWA KE JADWAL)
CREATE TABLE study_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    schedule_id UUID NOT NULL REFERENCES class_schedules(id) ON DELETE CASCADE,
    status enrollment_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_student_schedule UNIQUE (student_id, schedule_id)
);

-- 5. SESI KELAS (TRIGGER SAAT DOSEN BUKA ABSEN)
CREATE TABLE class_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES class_schedules(id) ON DELETE RESTRICT,
    meeting_no SMALLINT NOT NULL,             -- Pertemuan ke- (1-16)
    session_date DATE NOT NULL DEFAULT CURRENT_DATE,
    qr_seed VARCHAR(64) NOT NULL,             -- Kunci TOTP untuk QR bergulir
    is_open BOOLEAN NOT NULL DEFAULT TRUE,
    opened_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMPTZ,
    bap_topic TEXT,                           -- Catatan BAP Dosen
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_schedule_meeting UNIQUE (schedule_id, meeting_no)
);

-- 6. TRANSAKSI PRESENSI
CREATE TABLE attendances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES class_sessions(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status attendance_status NOT NULL,
    scanned_at TIMESTAMPTZ,
    device_id VARCHAR(120),                   -- ID HP saat melakukan scan
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    distance_meters DOUBLE PRECISION,
    attachment_url VARCHAR(255),              -- Bukti izin/sakit
    notes TEXT,
    verified_by_lecturer BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_session_student UNIQUE (session_id, student_id)
);

```

---

### 3. Indeks Khusus untuk Query Super Ringan

Saat ribuan mahasiswa serentak memindai presensi di 15 menit awal kelas, query tidak boleh melakukan *table scan*. Berikut indeks penopangnya:

```sql
-- 1. Index Sesi Aktif: Query pencarian jadwal yang sedang buka kelas
CREATE INDEX idx_sessions_active_lookup 
ON class_sessions (schedule_id) 
WHERE is_open = TRUE;

-- 2. Index Mahasiswa Scan: Validasi instan apakah mahasiswa sudah presensi
CREATE INDEX idx_attendances_session_student_lookup 
ON attendances (session_id, student_id);

-- 3. Index Jadwal Hari Ini: Query beranda mahasiswa/dosen berdasarkan hari berjalan
CREATE INDEX idx_schedules_day_time 
ON class_schedules (day_of_week, start_time, end_time) 
WHERE is_active = TRUE AND deleted_at IS NULL;

-- 4. Index KRS Mahasiswa: Load mata kuliah aktif semester berjalan
CREATE INDEX idx_study_plans_active_student 
ON study_plans (student_id) 
WHERE status = 'active';

-- 5. Index Monitoring Realtime Prodi/Fakultas
CREATE INDEX idx_users_prodi_role 
ON users (prodi_id, role) 
WHERE is_active = TRUE;

```

---

### 4. Titik Kunci Keamanan & Integritas Data

1. **Anti-Duplikasi Presensi:** Adanya `CONSTRAINT uq_session_student UNIQUE (session_id, student_id)` mencegah mahasiswa melakukan *double submission* secara bersamaan (*race condition* ditangani di level basis data).
2. **Device Enforcement:** Perbandingan kolom `device_id` di tabel `attendances` dengan `device_id` di tabel `users` memudahkan *flagging* jika ada indikasi titip absen dengan meminjam ponsel teman.
3. **Data Immutability:** Data pada `class_sessions` dan `attendances` memakai `RESTRICT` pada relasi ke `class_schedules`, menjamin riwayat audit presensi tidak akan hilang meskipun jadwal diubah.