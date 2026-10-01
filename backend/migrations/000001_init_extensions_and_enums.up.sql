-- Migration: 000001_init_extensions_and_enums.up.sql
-- Purpose: Create PostgreSQL extensions and ENUM types

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM (
            'mahasiswa',
            'dosen',
            'admin_prodi',
            'pimpinan',
            'superadmin'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'attendance_status') THEN
        CREATE TYPE attendance_status AS ENUM (
            'hadir',
            'terlambat',
            'izin',
            'sakit',
            'alpa'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'enrollment_status') THEN
        CREATE TYPE enrollment_status AS ENUM (
            'active',
            'dropped',
            'withdrawn'
        );
    END IF;
END $$;
