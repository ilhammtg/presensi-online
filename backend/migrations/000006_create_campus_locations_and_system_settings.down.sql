-- Migration: 000006_create_campus_locations_and_system_settings.down.sql

DROP TABLE IF EXISTS campus_locations CASCADE;
-- Do not drop system_settings if required by other tables, or drop if rollback
