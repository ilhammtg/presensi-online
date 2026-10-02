-- Migration: 000009_fix_password_hash.up.sql
-- Purpose: Fix broken bcrypt hashes from migration 000008 seed data.
--          The previous hash ($2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy)
--          was invalid and did not match any password.
--          New hash is for password "password": $2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi
--
-- Also aligns external_id and email with payload.json used by sync worker.
-- Also ensure avatar_url column exists on users table.
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT;

-- Fix superadmin: SUP-01 → SUPERADMIN01, nama dan email sesuai payload.json
UPDATE users
SET 
    external_id   = 'SUPERADMIN01',
    name          = 'Administrator Sistem',
    password_hash = '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    updated_at    = CURRENT_TIMESTAMP
WHERE external_id = 'SUP-01' OR external_id = 'SUPERADMIN01';

-- Fix admin prodi INF: ADM-INF-01 → ADM-001, email sesuai payload.json
UPDATE users
SET 
    external_id   = 'ADM-001',
    email         = 'admin.ti@kampus.ac.id',
    name          = 'Admin Prodi Informatika',
    password_hash = '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    updated_at    = CURRENT_TIMESTAMP
WHERE external_id IN ('ADM-INF-01', 'ADM-001') OR email IN ('admin.inf@kampus.ac.id', 'admin.ti@kampus.ac.id');

-- Fix admin prodi IM
UPDATE users
SET 
    password_hash = '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    updated_at    = CURRENT_TIMESTAMP
WHERE external_id = 'ADM-IM-01' OR email = 'admin.im@kampus.ac.id';

-- Fix all other users (dosen, mahasiswa) yang hash nya dari seed salah
-- Hanya update jika hash masih yang lama (invalid)
UPDATE users
SET 
    password_hash = '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    updated_at    = CURRENT_TIMESTAMP
WHERE password_hash = '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy';
