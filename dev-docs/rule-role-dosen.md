Berikut adalah dokumen spesifikasi teknis lengkap yang sudah disempurnakan dalam format Markdown (`.md`). Dokumen ini mencakup integrasi seluruh kebutuhan arsitektur, database terenkripsi, alur multi-kelas dosen, hingga mekanisme izin fleksibel (*Dual-Channel Permission* / izin via WhatsApp). Dokumen ini siap diberikan ke *agent AI* Anda sebagai panduan implementasi.

---

```markdown
# TECHNICAL SPECIFICATION DOCUMENT: SMART ATTENDANCE SYSTEM
**Target Deliverable:** Production-Ready Mobile & Web Geofence Attendance System  
**Version:** 1.2.0  
**Security Level:** AES-256-GCM Decoupled Ingestion & Anti-Fraud Mobile Protection  

---

## 1. PROJECT OVERVIEW & ARCHITECTURE

### 1.1 Executive Summary
Sistem presensi cerdas terintegrasi untuk perguruan tinggi yang memanfaatkan validasi **Dynamic Rolling QR Code** (TOTP via WebSocket) dan **Geofencing GPS** berbasis ruangan. 

Sistem ini dirancang dengan pendekatan *decoupled architecture* agar **sama sekali tidak membebani database utama akademik kampus (SIAKAD)** saat lonjakan trafik jam perkuliahan. Data disinkronisasi melalui background worker dengan enkripsi data simetris tingkat tinggi.

### 1.2 Technology Stack
* **Backend Service:** Golang (Fiber / Gin / Chi)
  * REST API untuk transaksi data & CRUD.
  * Gorilla WebSocket / Centrifugo untuk siaran Rolling QR dan pembaruan kehadiran realtime.
  * Ingestion Background Worker untuk sinkronisasi data dari Mock Kampus API.
* **Database & In-Memory Storage:**
  * **PostgreSQL 15+:** Penyimpanan relasional transaksional utama.
  * **Redis 7+:** Token state Rolling QR, TTL Sesi, Device Session, dan Rate Limiter API.
* **Frontend Web Dashboard:** Vue.js 3 (Composition API, Pinia, Vite, Tailwind CSS).
* **Mobile Application:** Flutter (Kotlin/Swift native bridges untuk proteksi Mock Location & Device Binding).

### 1.3 Topology & Data Pipeline


```

[ Mock Campus API / SIAKAD ]
│
▼ (AES-256-GCM Encrypted JSON Payload via HTTPS)
+──────────────────────────────────────────────────────────+
| Golang Sync Worker (Ingestion Engine)                   |
| 1. Buka Enkripsi (Verify Auth Tag + Decrypt Payload)    |
| 2. Bulk UPSERT data master ke PostgreSQL lokal           |
+──────────────────────────────────────────────────────────+
│
▼
+──────────────────────────────────────────────────────────+
| Local PostgreSQL & Redis Infrastructure                  |
+──────────────────────────────────────────────────────────+
▲                                       ▲
│ (REST & WebSocket)                    │ (REST & WebSocket)
▼                                       ▼
[ Flutter App ]                        [ Vue.js Dashboard ]

* Mahasiswa (Scan QR, GPS, E-KTM)      - Admin Prodi (Rekap, BAP)
* Dosen (Rolling QR, Manual Izin/WA)   - Pimpinan / Rektor (Monitoring)

```

---

## 2. BUSINESS RULES & SYSTEM POLICIES

### 2.1 Presensi & Keamanan
* **BR-SEC-01 (Decoupled Database):** Seluruh query pembacaan dan pencatatan presensi hanya mengakses PostgreSQL lokal presensi. Kueri langsung ke database kampus tidak diizinkan.
* **BR-SEC-02 (Anti-Titip Absen - Rolling QR):** QR Code yang ditampilkan dosen berganti setiap 10–15 detik (*rolling hash* berbasis TOTP). Kode QR yang difoto atau dibagikan keluar ruangan akan kedaluwarsa seketika.
* **BR-SEC-03 (Geofencing Validation):** Koordinat perangkat mahasiswa wajib berada di dalam radius toleransi ruangan (`rooms.radius_meters`) yang dihitung menggunakan formula Haversine di layer Go backend.
* **BR-SEC-04 (Device Binding):** Akun mahasiswa terkunci pada 1 identitas perangkat (`device_id`). Jika terdeteksi berganti perangkat tanpa izin admin, presensi ditolak.
* **BR-SEC-05 (Anti-Mock GPS):** Aplikasi Flutter mendeteksi *mock provider/fake GPS* dan status *root/jailbreak*. Jika terdeteksi, pemindaian dibatalkan otomatis.

### 2.2 Role Dosen & Manajemen Kelas
* **BR-DSN-01 (Multi-Course & Multi-Class):** 1 dosen dapat mengampu banyak Mata Kuliah (MK), dan 1 MK dapat memiliki banyak kelas paralel (misal: *Pemrograman Web - Kelas A* dan *Kelas B*). Data presensi terisolasi per kelas per pertemuan.
* **BR-DSN-02 (Sesi Tunggal Pertemuan):** Dosen hanya dapat membuka sesi presensi sesuai hari perkuliahan aktif (`day_of_week`). Satu nomor pertemuan (`meeting_no` 1–16) hanya dapat dijalankan sekali per kelas.
* **BR-DSN-03 (Validasi BAP):** Sesi perkuliahan tidak dapat ditutup secara permanen jika dosen belum mengisi ringkasan materi/BAP (`bap_topic`) minimal 10 karakter.

### 2.3 Mekanisme Izin Fleksibel (Dual-Channel Permission)
* **BR-IZN-01 (Jalur Mandiri / App Request):** Mahasiswa mengajukan via aplikasi dengan mengunggah foto surat dokter/surat dispensasi $\rightarrow$ Berstatus `pending` $\rightarrow$ Dosen menyetujui/menolak via Approval Center.
* **BR-IZN-02 (Jalur Langsung / Chat WA / Fisik):** Jika mahasiswa mengabari via WhatsApp atau lisan:
  * Dosen dapat langsung mengubah status mahasiswa menjadi `izin` atau `sakit` secara manual di aplikasi mobile atau web dosen.
  * Status langsung sah (`verified_by_lecturer = true`) tanpa perlu mahasiswa mengunggah form/surat di aplikasi.
  * Sistem otomatis mencatat log audit: `submission_source = 'manual_lecturer'` dan `updated_by = UUID_DOSEN`.

---

## 3. DATA FLOW DIAGRAM (DFD)

### 3.1 DFD Level 0 (Context Diagram)


```

```
                +---------------------------+
                |  Sistem Akademik Kampus   |
                +---------------------------+
                    │                   ▲

```

Payload Terenkripsi  │                   │ Status Acknowledge
(AES-256-GCM Feed)   │                   │ Sinkronisasi
▼                   │
+───────────────────────────+
|                           |

* Kelola Sesi & QR|                           | - Scan Dynamic QR
* Input Izin / WA |                           | - Validasi GPS & Device
* Input BAP Kelas |      SISTEM PRESENSI      | - Riwayat Presensi
◄────────────────►|       GEOFENCE & QR       |◄───────────────────►
[ DOSEN ]      |                           |    [ MAHASISWA ]
|                           |
+───────────────────────────+
▲
│ Monitoring Realtime,
│ Rekap Presensi & Audit BAP
▼
+───────────────────+
|  ADMIN / PIMPINAN |
+───────────────────+

```

### 3.2 DFD Level 1 (Core Ingestion & Presensi Engine)


```

[ Kampus API Feed ]
│
▼ (Payload AES-256-GCM)
+─────────────────────────────────+
| 1.0 Decryption & Ingestion      | ──(Bulk Upsert)──► [(D1) Master Data Akademik]
+─────────────────────────────────+                           │
│ (Read Jadwal & KRS)
▼
[ Dosen ] ───► +─────────────────────────────────+ ◄──────────+
| 2.0 Sesi Perkuliahan & QR Gen   |
+─────────────────────────────────+
│
├──────(Simpan Sesi)─────► [(D2) Sesi Perkuliahan]
│                                  │
▼ (Broadcast TOTP Token via WS)    │ (Validasi Status Sesi)
[ Mahasiswa ] ► +─────────────────────────────────+ ◄─────────────+
| 3.0 Validasi Geofence & Absensi |
+─────────────────────────────────+ ◄──(Cek KRS)── [(D1) Master Data Akademik]
│
├──────(Catat Log Hadir)─► [(D3) Transaksi Presensi]
│                                  ▲
[ Dosen ] ────► +─────────────────────────────────+                │
(Izin via WA/   | 4.0 Manual Override & Izin WA   | ──(Update Izin)┘
Langsung)      +─────────────────────────────────+
│
▼ (Kueri Rekap & BAP)
+─────────────────────────────────+
| 5.0 Reporting & Web Monitoring  | ──► [ Admin Prodi / Rektor ]
+─────────────────────────────────+

```

---

## 4. COMPLETE PRODUCTION-READY DATABASE SCHEMA (POSTGRESQL)

```sql
-- Aktivasi ekstensi UUID generator
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Definisi ENUM Types
CREATE TYPE user_role AS ENUM ('mahasiswa', 'dosen', 'admin_prodi', 'pimpinan');
CREATE TYPE attendance_status AS ENUM ('hadir', 'terlambat', 'izin', 'sakit', 'alpa');
CREATE TYPE enrollment_status AS ENUM ('active', 'dropped', 'withdrawn');
CREATE TYPE submission_channel AS ENUM ('self_scan', 'app_request', 'manual_lecturer');

-- 1. TABEL FAKULTAS
CREATE TABLE faculties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. TABEL PROGRAM STUDI
CREATE TABLE study_programs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    faculty_id UUID NOT NULL REFERENCES faculties(id) ON DELETE RESTRICT,
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. TABEL GEDUNG
CREATE TABLE buildings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 4. TABEL RUANGAN (TITIK PUSAT GEOFENCE)
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

-- 5. TABEL PENGGUNA (USERS)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id VARCHAR(50) NOT NULL UNIQUE, -- NIM atau NIDN
    name VARCHAR(150) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role user_role NOT NULL,
    prodi_id UUID REFERENCES study_programs(id) ON DELETE SET NULL,
    device_id VARCHAR(120),                  -- Penguncian Device Mahasiswa
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- 6. TABEL JADWAL KULIAH (MULTI-MK & MULTI-KELAS)
CREATE TABLE class_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_schedule_id VARCHAR(50) UNIQUE, -- ID Jadwal dari Kampus
    course_code VARCHAR(30) NOT NULL,
    course_name VARCHAR(150) NOT NULL,
    class_group VARCHAR(10) NOT NULL,        -- Label Kelas: 'A', 'B', 'Pagi', dll
    academic_year VARCHAR(10) NOT NULL,      -- Format: '2026/2027'
    semester_type SMALLINT NOT NULL,         -- 1: Ganjil, 2: Genap, 3: Pendek
    lecturer_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE RESTRICT,
    day_of_week SMALLINT NOT NULL,           -- 1 (Senin) s.d 7 (Minggu)
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- 7. TABEL KRS MAHASISWA (STUDY PLANS)
CREATE TABLE study_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    schedule_id UUID NOT NULL REFERENCES class_schedules(id) ON DELETE CASCADE,
    status enrollment_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_student_schedule UNIQUE (student_id, schedule_id)
);

-- 8. TABEL SESI PERTEMUAN KELAS
CREATE TABLE class_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES class_schedules(id) ON DELETE RESTRICT,
    meeting_no SMALLINT NOT NULL,             -- Pertemuan ke- (1-16)
    session_date DATE NOT NULL DEFAULT CURRENT_DATE,
    qr_seed VARCHAR(64) NOT NULL,             -- Secret Key untuk Generator TOTP
    is_open BOOLEAN NOT NULL DEFAULT TRUE,
    opened_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMPTZ,
    bap_topic TEXT,                           -- Catatan BAP / Materi Dosen
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_schedule_meeting UNIQUE (schedule_id, meeting_no)
);

-- 9. TABEL TRANSAKSI PRESENSI (ATTENDANCES)
CREATE TABLE attendances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES class_sessions(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status attendance_status NOT NULL,
    scanned_at TIMESTAMPTZ,
    device_id VARCHAR(120),                   -- ID Device saat scan absensi
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    distance_meters DOUBLE PRECISION,
    attachment_url VARCHAR(255),              -- Bukti izin/sakit (opsional jika via WA)
    notes TEXT,                               -- Catatan alasan izin/sakit / keterangan WA
    submission_source submission_channel NOT NULL DEFAULT 'self_scan',
    updated_by UUID REFERENCES users(id),     -- ID Dosen (jika input izin langsung)
    verified_by_lecturer BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_session_student UNIQUE (session_id, student_id)
);

-- Compound Indexes untuk Skalabilitas Jam Sibuk
CREATE INDEX idx_sessions_active ON class_sessions (schedule_id) WHERE is_open = TRUE;
CREATE INDEX idx_attendances_session_student ON attendances (session_id, student_id);
CREATE INDEX idx_study_plans_student ON study_plans (student_id) WHERE status = 'active';
CREATE INDEX idx_schedules_lecturer ON class_schedules (lecturer_id) WHERE is_active = TRUE;
CREATE INDEX idx_schedules_day_time ON class_schedules (day_of_week, start_time, end_time);

```

---

## 5. EXAMPLE RECORDS SETIAP ENTITAS (SAMPLE DATA)

### 5.1 `faculties`

| id | code | name |
| --- | --- | --- |
| `a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11` | FT | Fakultas Teknik |

### 5.2 `study_programs`

| id | faculty_id | code | name |
| --- | --- | --- | --- |
| `c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33` | `a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11` | INF | S1 Informatika |

### 5.3 `buildings`

| id | code | name |
| --- | --- | --- |
| `e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55` | GD-TI | Gedung Kuliah Terpadu TI |

### 5.4 `rooms`

| id | building_id | room_code | name | latitude | longitude | radius_meters |
| --- | --- | --- | --- | --- | --- | --- |
| `10eebc99-9c0b-4ef8-bb6d-6bb9bd380a77` | `e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55` | LAB-R01 | Lab Jaringan & Komputasi | 5.201452 | 96.702145 | 35 |

### 5.5 `users`

| id | external_id | name | role | prodi_id | device_id |
| --- | --- | --- | --- | --- | --- |
| `30eebc99-9c0b-4ef8-bb6d-6bb9bd380a99` | 198801102015041001 | Dr. Irwan Setiawan, M.Kom. | dosen | `c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33` | `DEV-DOSEN-IRWAN` |
| `40eebc99-9c0b-4ef8-bb6d-6bb9bd380b11` | 23552011001 | Ahmad Fauzi | mahasiswa | `c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33` | `DEV-ANDR-FAUZI` |
| `50eebc99-9c0b-4ef8-bb6d-6bb9bd380b22` | 23552011002 | Siti Rahmah | mahasiswa | `c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33` | `DEV-IOS-SITI` |

### 5.6 `class_schedules` (1 Dosen Mengajar Banyak MK & Kelas Beda)

| id | external_schedule_id | course_code | course_name | class_group | lecturer_id | room_id | day_of_week | start_time | end_time |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `70eebc99-9c0b-4ef8-bb6d-6bb9bd380b44` | SCH-2026-INF-01 | INF301 | Pemrograman Sistem Terdistribusi | A | `30eebc99-9c0b-4ef8-bb6d-6bb9bd380a99` | `10eebc99-9c0b-4ef8-bb6d-6bb9bd380a77` | 2 | 08:00:00 | 10:30:00 |
| `71eebc99-9c0b-4ef8-bb6d-6bb9bd380b45` | SCH-2026-INF-02 | INF301 | Pemrograman Sistem Terdistribusi | B | `30eebc99-9c0b-4ef8-bb6d-6bb9bd380a99` | `10eebc99-9c0b-4ef8-bb6d-6bb9bd380a77` | 2 | 10:45:00 | 13:15:00 |
| `72eebc99-9c0b-4ef8-bb6d-6bb9bd380b46` | SCH-2026-INF-03 | INF202 | Arsitektur & Organisasi Komputer | A | `30eebc99-9c0b-4ef8-bb6d-6bb9bd380a99` | `10eebc99-9c0b-4ef8-bb6d-6bb9bd380a77` | 4 | 08:00:00 | 10:30:00 |

### 5.7 `study_plans` (KRS)

| id | student_id | schedule_id | status |
| --- | --- | --- | --- |
| `80eebc99-9c0b-4ef8-bb6d-6bb9bd380b55` | `40eebc99-9c0b-4ef8-bb6d-6bb9bd380b11` | `70eebc99-9c0b-4ef8-bb6d-6bb9bd380b44` | active |
| `81eebc99-9c0b-4ef8-bb6d-6bb9bd380b56` | `50eebc99-9c0b-4ef8-bb6d-6bb9bd380b22` | `70eebc99-9c0b-4ef8-bb6d-6bb9bd380b44` | active |

### 5.8 `class_sessions`

| id | schedule_id | meeting_no | session_date | qr_seed | is_open | bap_topic |
| --- | --- | --- | --- | --- | --- | --- |
| `a1eebc99-9c0b-4ef8-bb6d-6bb9bd380b77` | `70eebc99-9c0b-4ef8-bb6d-6bb9bd380b44` | 3 | 2026-09-22 | `K9xZ8#vL1$mP2!qR` | TRUE | Sinkronisasi Jam & Distributed Consensus |

### 5.9 `attendances` (Demonstrasi Transaksi Scan Mandiri vs Izin WA)

| id | session_id | student_id | status | scanned_at | submission_source | updated_by | notes |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b88` | `a1eebc99-9c0b-4ef8-bb6d-6bb9bd380b77` | `40eebc99-9c0b-4ef8-bb6d-6bb9bd380b11` | hadir | 2026-09-22 08:04:12+07 | self_scan | NULL | Scan mandiri GPS valid |
| `b2eebc99-9c0b-4ef8-bb6d-6bb9bd380b89` | `a1eebc99-9c0b-4ef8-bb6d-6bb9bd380b77` | `50eebc99-9c0b-4ef8-bb6d-6bb9bd380b22` | izin | NULL | manual_lecturer | `30eebc99...` (Dosen) | Izin via chat WA - ada kemalangan keluarga |

---

## 6. ROLE FEATURE BREAKDOWN & UI SPECIFICATION

### 6.1 Role: Mahasiswa (Mobile Flutter)

1. **Beranda & Jadwal:**
* Kartu daftar mata kuliah semester berjalan berdasarkan KRS.
* Setiap kartu memuat: Nama MK, Kode Kelas, Ruangan, Jam, Nama Dosen Pengampu.
* Kartu dinamis hari ini: Label status apakah sesi presensi sudah dibuka atau belum oleh dosen.


2. **Pemindai Presensi (Core):**
* Pemindai kamera QR instan.
* Auto-validasi koordinat GPS ke titik geofence ruangan.
* Notifikasi status instan: *Hadir Tepat Waktu*, *Terlambat*, atau *Ditolak (Di luar radius / QR expired)*.


3. **Pengajuan Izin Mandiri (App Channel):**
* Form pengajuan: Pilih Sakit / Izin.
* Upload surat bukti (kamera/dokumen).
* Status tracking: Menunggu Konfirmasi Dosen / Disetujui / Ditolak.


4. **Riwayat & Profil:**
* Rekap absensi per mata kuliah selama 1 semester (Persentase Hadir, Sakit, Izin, Alpa).
* *Early Warning Bar*: Peringatan jika alpa mendekati batas toleransi maksimal kehadiran (<75%).
* Menu Profil & E-KTM digital dengan barcode identitas mahasiswa.



---

### 6.2 Role: Dosen (Mobile Flutter & Web Dashboard)

1. **Beranda & Jadwal Mengajar Hari Ini:**
* Ringkasan kelas yang harus diajar hari ini (MK, Kelas Paralel, Jam, Ruangan).
* Tombol aksi: **"Buka Sesi Perkuliahan"**.


2. **Halaman Sesi Berjalan (Live Presensi):**
* **Dynamic QR Code:** Tampilan kode QR berganti otomatis tiap 10–15 detik via WebSocket (dapat dihubungkan ke proyektor atau via layar HP).
* **Live Stream Kehadiran:** Daftar nama mahasiswa kelas tersebut masuk otomatis secara realtime saat memindai.
* **Aksi Cepat Izin Langsung / WA:** Dosen bisa tap nama mahasiswa $\rightarrow$ Pilih **"Set Izin"** atau **"Set Sakit"** $\rightarrow$ Isi catatan singkat (misal: *Izin via WA*) $\rightarrow$ Data kehadiran langsung tersimpan sah.
* **BAP & Tutup Sesi:** Form pengisian materi perkuliahan dan tombol **"Tutup Sesi Perkuliahan"**.


3. **Menu "Mata Kuliah Saya" (Katalog MK & Kelas):**
* Menampilkan semua daftar MK yang diampu.
* Pemisahan kelas paralel yang jelas (misal: *Sistem Operasi - Kelas A* dan *Kelas B*).
* **Rekapitulasi Presensi per MK & Kelas:** Matriks kehadiran Pertemuan 1–16, persentase kehadiran seluruh mahasiswa kelas tersebut, dan tombol ekspor ke Excel/PDF.


4. **Persetujuan Izin (Approval Center):**
* Meninjau mahasiswa yang mengajukan izin via form aplikasi mandiri.
* Tombol: **"Setujui"** atau **"Tolak"**.
* Tombol cepat: **"+ Beri Izin Manual"** untuk menginput izin mahasiswa yang mengabari via WhatsApp di luar jam kelas.


5. **Pengaturan Akun:**
* Tampilan biodata diri & NIDN.
* Form **Ganti Password**.



---

### 6.3 Role: Admin Program Studi (Web Dashboard Vue.js)

* **Monitoring Realtime:** Dashboard status kehadiran kelas per prodi yang sedang aktif detik demi detik.
* **Manajemen & Evaluasi:** Melihat rekapitulasi kehadiran per semester, BAP dosen, dan daftar mahasiswa dengan kehadiran di bawah ambang batas (terancam tidak lulus/tidak bisa ikut ujian).
* **Pusat Ekspor:** Download laporan presensi prodi format BAP resmi (PDF) dan rekapitulasi nilai kehadiran (Excel).

---

### 6.4 Role: Pimpinan / Dekan / Rektor (Web Dashboard Vue.js)

* **Executive Dashboard:** Grafik metrik tingkat kehadiran dosen dan mahasiswa per fakultas/program studi secara realtime.
* **Audit Kinerja Mengajar:** Analisis rasio kesesuaian jam mengajar terhadap jadwal resmi, kelas kosong, atau keterlambatan buka sesi oleh pengajar.

---

## 7. MOCK CAMPUS API & ENKRIPSI AES-256-GCM

Data master dikirim dari API kampus dalam bentuk payload terenkripsi **AES-256-GCM** (*Authenticated Encryption with Associated Data*) untuk menjamin *confidentiality* dan *anti-tampering*.

### 7.1 Format Payload Terenkripsi (Transit JSON)

```json
{
  "service": "CAMPUS_ACADEMIC_FEED",
  "version": "1.0",
  "timestamp": 1790442400,
  "iv": "dGVzdG5vbmNlMTIzNA==",
  "ciphertext": "5a7b8e1f0c2a8839d012...[encrypted_hex_string]...",
  "tag": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
}

```

### 7.2 Data Mentah Asli (Payload Sebelum Dienkripsi / Hasil Dekripsi)

```json
{
  "sync_timestamp": "2026-09-27T08:00:00Z",
  "academic_year": "2026/2027",
  "semester_type": 1,
  "faculties": [
    {
      "code": "FT",
      "name": "Fakultas Teknik",
      "study_programs": [
        {
          "code": "INF",
          "name": "S1 Informatika"
        }
      ]
    }
  ],
  "buildings": [
    {
      "code": "GD-TI",
      "name": "Gedung Kuliah Terpadu TI",
      "rooms": [
        {
          "room_code": "LAB-R01",
          "name": "Lab Jaringan & Komputasi",
          "latitude": 5.201452,
          "longitude": 96.702145,
          "radius_meters": 35
        }
      ]
    }
  ],
  "users": [
    {
      "external_id": "198801102015041001",
      "name": "Dr. Irwan Setiawan, M.Kom.",
      "email": "irwan.s@kampus.ac.id",
      "role": "dosen",
      "prodi_code": "INF"
    },
    {
      "external_id": "23552011001",
      "name": "Ahmad Fauzi",
      "email": "ahmad.fauzi@kampus.ac.id",
      "role": "mahasiswa",
      "prodi_code": "INF"
    },
    {
      "external_id": "23552011002",
      "name": "Siti Rahmah",
      "email": "siti.rahmah@kampus.ac.id",
      "role": "mahasiswa",
      "prodi_code": "INF"
    }
  ],
  "schedules": [
    {
      "external_schedule_id": "SCH-2026-INF-01",
      "course_code": "INF301",
      "course_name": "Pemrograman Sistem Terdistribusi",
      "class_group": "A",
      "lecturer_id": "198801102015041001",
      "room_code": "LAB-R01",
      "day_of_week": 2,
      "start_time": "08:00:00",
      "end_time": "10:30:00",
      "enrolled_students": [
        "23552011001",
        "23552011002"
      ]
    },
    {
      "external_schedule_id": "SCH-2026-INF-02",
      "course_code": "INF301",
      "course_name": "Pemrograman Sistem Terdistribusi",
      "class_group": "B",
      "lecturer_id": "198801102015041001",
      "room_code": "LAB-R01",
      "day_of_week": 2,
      "start_time": "10:45:00",
      "end_time": "13:15:00",
      "enrolled_students": []
    }
  ]
}

```

---

## 8. BACKEND IMPLEMENTATION: AES-GCM & HAVERSINE (GOLANG)

Berikut referensi modul kriptografi dan kalkulasi jarak geofence di Golang:

### 8.1 Modul Enkripsi & Dekripsi AES-256-GCM

```go
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

// DecryptPayload membuka enkripsi ciphertext menggunakan AES-256-GCM
func DecryptPayload(cipherTextHex, keyHex, nonceHex, tagHex string) ([]byte, error) {
	key, _ := hex.DecodeString(keyHex)
	nonce, _ := hex.DecodeString(nonceHex)
	tag, _ := hex.DecodeString(tagHex)
	cipherText, _ := hex.DecodeString(cipherTextHex)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Gabungkan ciphertext dan auth tag (standar Go GCM mengharapkan tag di akhir payload)
	fullCipher := append(cipherText, tag...)

	plainText, err := aesGCM.Open(nil, nonce, fullCipher, nil)
	if err != nil {
		return nil, errors.New("autentikasi tag gagal: payload data rusak atau termanipulasi")
	}

	return plainText, nil
}

```

### 8.2 Modul Geofence Radius (Formula Haversine di Backend)

```go
package location

import (
	"math"
)

// CalculateHaversineDistance menghitung jarak antara dua koordinat GPS dalam satuan meter
func CalculateHaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // Radius bumi dalam satuan meter

	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	radLat1 := lat1 * (math.Pi / 180.0)
	radLat2 := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(radLat1)*math.Cos(radLat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// IsWithinRadius mengecek apakah mahasiswa berada di dalam toleransi radius ruangan
func IsWithinRadius(studentLat, studentLon, roomLat, roomLon float64, radiusMeters int) (bool, float64) {
	dist := CalculateHaversineDistance(studentLat, studentLon, roomLat, roomLon)
	return dist <= float64(radiusMeters), dist
}

```

---

## 9. INSTRUCTION FOR AI DEVELOPMENT AGENT

Saat mengimplementasikan sistem berdasarkan dokumen ini:

1. Mulai dengan membuat skrip migrasi database pada PostgreSQL berdasarkan skema di **Bab 4**. Pastikan semua *foreign key constraints* dan *compound indexes* terpasang sempurna.
2. Buat unit test mandiri untuk fungsi dekripsi **AES-256-GCM** dan kalkulasi jarak **Haversine** di Golang.
3. Pastikan penanganan *race condition* saat presensi diselesaikan dengan mengandalkan `CONSTRAINT uq_session_student UNIQUE (session_id, student_id)` di PostgreSQL, bukan sekadar pengecekan `SELECT` di kode backend.
4. Buat antarmuka dosen di Flutter dengan menyertakan opsi *Dual-Channel Permission*: tombol persetujuan izin aplikasi dan tombol aksi cepat untuk input izin mahasiswa yang mengabari via WhatsApp.

```

```