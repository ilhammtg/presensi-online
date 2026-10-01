# Smart Campus Geofence Attendance System

Sistem presensi kampus modern berbasis **Geofencing** dan **Dynamic Rolling QR Code (TOTP)**, dirancang untuk meminimalkan kecurangan (titip absen, fake GPS) dan mencegah beban berlebih pada pangkalan data kampus.

## Arsitektur

```
Campus API (Mock SIAKAD)
        │ AES-256-GCM HTTPS
        ▼
Golang Backend (REST API + WebSocket + Sync Worker)
        │
   ┌────┴────┐
   ▼         ▼
PostgreSQL  Redis (QR Token Cache + Session)
   │
   ├── Flutter Mobile App (Mahasiswa + Dosen)
   └── Vue.js Web Dashboard (Admin, Pimpinan, Superadmin)
```

## Struktur Proyek

```
presensi-online/
├── backend/              # Golang monorepo
│   ├── cmd/
│   │   ├── api/          # REST API + WebSocket server
│   │   ├── worker/       # Campus data sync worker
│   │   └── mock-campus-api/  # SIAKAD simulator (AES-256-GCM)
│   ├── internal/
│   │   ├── domain/       # Entities + Repository interfaces
│   │   ├── app/          # Use cases (business logic)
│   │   ├── handler/      # HTTP handlers (Gin)
│   │   ├── ws/           # WebSocket hub
│   │   ├── worker/       # Sync worker logic
│   │   ├── crypto/       # AES-256-GCM
│   │   ├── middleware/   # JWT Auth + RBAC
│   │   └── repository/   # PostgreSQL + Redis implementations
│   └── migrations/       # SQL migration files
├── web-dashboard/        # Vue.js 3 (Fase 4)
├── mobile/               # Flutter (Fase 3)
├── docker-compose.yml
└── Makefile
```

## Quick Start

### 1. Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Flutter 3.x (untuk mobile)
- Node.js 20+ (untuk web dashboard)

### 2. Menjalankan Semua Service dengan Docker (Recommended)

Cukup satu perintah untuk menjalankan seluruh stack (Database, Migrations, Backend API, Worker, Mock SIAKAD, dan Web Dashboard):

```bash
# Start all containers
make up
# atau: docker compose up -d --build
```

Setelah container berjalan:
- 🌐 **Web Dashboard (Vue 3)**: [http://localhost:5173](http://localhost:5173) (atau [http://localhost:3000](http://localhost:3000))
- 🚀 **REST API & WebSocket**: [http://localhost:8080](http://localhost:8080)
- 🏫 **Mock Campus API (SIAKAD)**: [http://localhost:9001](http://localhost:9001)
- 🐘 **PostgreSQL**: `localhost:5434`
- ⚡ **Redis**: `localhost:6380`

Untuk melihat logs atau mematikan service:
```bash
make logs   # Follow log semua container
make ps     # Cek status container
make down   # Hentikan semua container
```

### 3. Setup Dev Lokal Manual (Opsional)

Jika ingin menjalankan service backend dan frontend secara native/lokal di mesin host:

```bash
# 1. Start PostgreSQL + Redis saja
make dev

# 2. Apply DB migrations
make migrate-up

# 3. Jalankan masing-masing service di terminal terpisah:
make run-mock     # Mock Campus API (:9001)
make run-worker   # Sync Worker
make run-api      # REST API (:8080)
make run-web      # Vue Dashboard (:5173)
```

### 5. Run Tests

```bash
make test          # Semua tests
make test-crypto   # Crypto module saja
make test-cover    # Dengan coverage report
```

## API Endpoints

| Method | Endpoint | Auth | Deskripsi |
|--------|----------|------|-----------|
| `GET` | `/health` | - | Health check |
| `POST` | `/v1/auth/login` | - | Login (email + password + device_id) |
| `GET` | `/v1/auth/me` | JWT | Profile user |
| `POST` | `/v1/attendance/scan` | JWT (Mahasiswa) | Scan QR presensi |
| `POST` | `/v1/sessions` | JWT (Dosen) | Buka sesi presensi |
| `PATCH` | `/v1/sessions/:id/close` | JWT (Dosen) | Tutup sesi + input BAP |
| `GET` | `/v1/sessions/:id/attendees` | JWT (Dosen) | Daftar hadir realtime |
| `GET` | `/ws` | JWT (query param) | WebSocket connection |

## Validasi Presensi (7 Langkah)

```
POST /v1/attendance/scan
  1. Sesi masih buka? (is_open = true)
  2. QR Token valid? (TOTP via Redis, window 15 detik)
  3. Mahasiswa enrolled di kelas? (KRS/study_plans)
  4. Device ID cocok? (device binding anti titip absen)
  5. Di dalam radius ruangan? (Haversine formula)
  6. Tepat waktu atau terlambat? (15 menit toleransi)
  7. Commit ke DB + Broadcast WebSocket
```

## Database Migrations

| File | Isi |
|------|-----|
| `000001` | pgcrypto extension + ENUM types |
| `000002` | faculties, study_programs, buildings, rooms |
| `000003` | users (dengan device binding) |
| `000004` | class_schedules, study_plans (KRS) |
| `000005` | class_sessions, attendances + indexes |

## Tech Stack

| Layer | Teknologi |
|-------|-----------|
| Backend | Go 1.21, Gin, pgx/v5, go-redis |
| Auth | JWT (golang-jwt/jwt/v5), bcrypt |
| QR | TOTP (HMAC-SHA1, RFC 6238), Redis |
| Crypto | AES-256-GCM (crypto/cipher) |
| Database | PostgreSQL 15, Redis 7 |
| Mobile | Flutter (Fase 3) |
| Web | Vue.js 3, Vite, Pinia, Tailwind (Fase 4) |
| DevOps | Docker Compose, multi-stage Dockerfile |

## Fase Pengembangan

- [x] **Fase 1**: Backend Foundation & Database
- [ ] **Fase 2**: Core Attendance Engine (WebSocket + QR rolling)
- [ ] **Fase 3**: Flutter Mobile App
- [ ] **Fase 4**: Vue.js Web Dashboard  
- [ ] **Fase 5**: Security Hardening & Testing
