# Software Requirement Specification (SRS): Smart Campus Geofence Attendance System

Sistem presensi perkuliahan modern berbasis lokasi (*geofencing*) dan *Dynamic Rolling QR Code* (TOTP), dirancang untuk meminimalkan kecurangan (titip absen, manipulasi lokasi/fake GPS) serta mencegah beban berlebih (*load spike*) pada pangkalan data inti universitas.

---

## 1. Arsitektur & Teknologi

* **Backend Service:** Golang (REST API, WebSocket Engine, Background Sync Worker)
* **Web Dashboard:** Vue.js 3 (Pinia, Tailwind CSS, Web-socket client)
* **Mobile Application:** Flutter (Android & iOS)
* **Database Transaksional:** PostgreSQL 15+ (Didukung Redis untuk sesi aktif & cache token rolling QR)
* **Protokol Keamanan Data:** AES-256-GCM (*Authenticated Encryption with Associated Data*) untuk pertukaran data API dari pangkalan data kampus.

---

## 2. Strategi Pengambilan & Sinkronisasi Data

Untuk mencegah latensi dan memutus ketergantungan langsung ke database akademik (SIAKAD), sistem menggunakan mekanisme *data ingestion decoupled*:

1. **Mock/Campus API Service** mengekspor payload data akademik (jadwal, ruangan, data mahasiswa/dosen, KRS).
2. Payload dienkripsi penuh menggunakan **AES-256-GCM** sebelum dikirimkan melalui jalur HTTPS.
3. **Golang Sync Worker** menerima payload, memverifikasi integritas data (*Auth Tag*), mendekripsi payload, dan menjalankan operasi `UPSERT` (`INSERT ... ON CONFLICT DO UPDATE`) ke PostgreSQL sistem presensi.
4. Aplikasi mobile dan web dashboard hanya melakukan kueri ke basis data PostgreSQL presensi, bukan ke server kampus.

---

## 3. Spesifikasi Enkripsi & Simulasi Payload API

Pertukaran data antara API kampus dan sistem presensi menggunakan format JSON dengan field terenkripsi berbasis AES-GCM (256-bit key, 96-bit nonce/IV).

### A. Format Amplop API (Encrypted Transit)

```json
{
  "service": "CAMPUS_ACADEMIC_FEED",
  "version": "1.0",
  "timestamp": 1790042400,
  "iv": "dGVzdG5vbmNlMTIzNA==",
  "ciphertext": "5a7b8e1f0c2a...[truncated_ciphertext]...",
  "tag": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
}

```

### B. Simulasi Data Mentah (Payload Sebelum Dienkripsi / Hasil Dekripsi)

```json
{
  "sync_timestamp": "2026-09-22T08:00:00Z",
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
      "code": "GEDUNG-B",
      "name": "Gedung Kuliah Terpadu",
      "rooms": [
        {
          "room_code": "LAB-KOM-1",
          "name": "Laboratorium Komputer 1",
          "latitude": 5.201452,
          "longitude": 96.702145,
          "radius_meters": 35
        }
      ]
    }
  ],
  "users": [
    {
      "external_id": "23552011001",
      "name": "Ahmad Fauzi",
      "email": "ahmad.fauzi@kampus.ac.id",
      "role": "mahasiswa",
      "prodi_code": "INF"
    },
    {
      "external_id": "198801102015041001",
      "name": "Dr. Irwan Setiawan, M.Kom.",
      "email": "irwan.s@kampus.ac.id",
      "role": "dosen",
      "prodi_code": "INF"
    }
  ],
  "schedules": [
    {
      "external_schedule_id": "SCH-2026-INF-001",
      "course_code": "INF301",
      "course_name": "Pemrograman Sistem Terdistribusi",
      "lecturer_id": "198801102015041001",
      "room_code": "LAB-KOM-1",
      "day_of_week": 2,
      "start_time": "08:00:00",
      "end_time": "10:30:00",
      "enrolled_students": [
        "23552011001"
      ]
    }
  ]
}

```

---

## 4. Rincian Fitur Aplikasi

### A. Mobile Application (Flutter)

Aplikasi difokuskan eksklusif untuk peran **Mahasiswa** dan **Dosen**.

#### 1. Peran: Mahasiswa

* **Beranda Perkuliahan:**
* Kartu jadwal kuliah hari ini (mata kuliah, jam, ruang, dosen pengampu).
* Indikator status sesi (apakah dosen sudah membuka presensi).


* **Modul Presensi (Core):**
* Pemindai QR berbasis kamera berkecepatan tinggi.
* *Geofencing Validation*: verifikasi jarak koordinat perangkat terhadap radius ruangan kelas.
* Status presensi instan (Hadir Tepat Waktu / Terlambat).


* **Izin & Dispensasi:**
* Pengajuan ketidakhadiran (Sakit / Izin).
* Pengunggahan berkas bukti (surat dokter / surat dispensasi).
* Pelacak status persetujuan oleh dosen.


* **Akademik & Profil:**
* Riwayat persentase kehadiran per mata kuliah dalam satu semester.
* Indikator peringatan batas alpa (maksimal toleransi absensi).
* E-KTM digital dengan informasi identitas dan kode barcode.



#### 2. Peran: Dosen

* **Manajemen Kelas Berjalan:**
* Buka dan tutup sesi perkuliahan sesuai jadwal.
* Penampil *Dynamic Rolling QR Code* (berganti setiap 10–15 detik via WebSocket) untuk ditampilkan di layar HP atau proyektor.
* Pengaturan batas toleransi keterlambatan.


* **Monitoring Kelas Realtime:**
* Daftar mahasiswa yang berhasil absen masuk secara instan via WebSocket.
* *Manual Override*: opsi menandai hadir manual jika terjadi kendala sinyal atau anomali sensor GPS mahasiswa.


* **Verifikasi Izin:**
* Tinjau surat bukti izin/sakit mahasiswa dan aksi persetujuan/penolakan.


* **Pelaporan:**
* Ringkasan rekap kelas harian dan pengisian Berita Acara Perkuliahan (BAP) singkat.



#### 3. Keamanan Perangkat (Device-Level Security)

* Deteksi *Mock Location / Fake GPS* (pemindaian dibatalkan jika aplikasi mendeteksi modul lokasi tiruan).
* Proteksi *Rooted / Jailbroken device*.
* *Device Binding*: akun mahasiswa dikunci pada satu UUID perangkat aktif.

---

### B. Web Dashboard (Vue.js + Golang RBAC)

Aplikasi web dirancang untuk peran manajerial dan pemantauan menyeluruh.

#### 1. Peran: Admin Program Studi (Prodi)

* **Monitoring Presensi Realtime:** Memantau sesi perkuliahan aktif di tingkat prodi detik demi detik.
* **Pengelolaan Master & Jadwal:** Tinjauan jadwal hasil sinkronisasi API, pemetaan ulang ruangan darurat jika terjadi perpindahan kelas.
* **Verifikasi Lanjutan:** Akses pengesahan dispensasi mahasiswa lintas kelas.
* **Pusat Rekapitulasi & Ekspor:** Mengunduh berkas BAP dan rekapitulasi kehadiran (format Excel dan PDF) untuk keperluan audit nilai akhir semester.

#### 2. Peran: Pimpinan (Dekan / Rektorat)

* **Executive Summary Dashboard:**
* Grafik metrik tingkat kehadiran dosen dan mahasiswa per fakultas/jurusan secara *realtime*.
* Rasio kepatuhan jam mulai perkuliahan terhadap jadwal resmi.
* Notifikasi kelas kosong atau dosen yang belum membuka sesi melewati batas toleransi.



#### 3. Peran: Superadmin / Tim IT

* Pemantauan *health check* sinkronisasi API kampus.
* Log rotasi kunci enkripsi AES-256-GCM.
* Manajemen hak akses (*Role-Based Access Control*).

---

## 5. Alur Validasi Presensi (Pipeline Logika Backend)

```
[Mahasiswa Scan QR] 
        |
        v
(1) Cek Sesi: Apakah sesi terbuka (is_open = true)?
        |-- [Tidak] --> Return HTTP 400 (Sesi Telah Berakhir)
        v
(2) Validasi Rolling QR Token (TOTP via Redis)
        |-- [Invalid/Expired] --> Return HTTP 400 (QR Code Kadaluwarsa)
        v
(3) Validasi Enrollment: Apakah mahasiswa ada di KRS jadwal ini?
        |-- [Tidak] --> Return HTTP 403 (Mahasiswa Tidak Terdaftar di Kelas)
        v
(4) Validasi Device ID: Apakah device_id cocok dengan profil user?
        |-- [Beda] --> Return HTTP 403 (Perangkat Tidak Sesuai / Terindikasi Titip Absen)
        v
(5) Validasi Radius Lokasi (Rumus Haversine: Jarak User < Radius Ruangan)
        |-- [Di luar radius] --> Return HTTP 422 (Di Luar Radius Presensi)
        v
(6) Evaluasi Waktu: Masuk toleransi tepat waktu atau terlambat?
        v
(7) Commit Transaction ke DB (INSERT ke attendances dengan Unique Constraint)
        v
(8) Broadcast WebSocket Event ke Mobile Dosen & Web Dashboard
