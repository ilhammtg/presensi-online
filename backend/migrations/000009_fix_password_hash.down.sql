-- Migration: 000009_fix_password_hash.down.sql
-- Rollback: tidak ada rollback yang aman untuk password hash
-- (tidak bisa kembalikan ke hash yang sudah invalid)
SELECT 1; -- no-op
