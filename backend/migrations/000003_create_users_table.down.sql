-- Migration: 000003_create_users_table.down.sql

DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_prodi_role;
DROP TABLE IF EXISTS users;
