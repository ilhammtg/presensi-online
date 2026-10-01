-- Migration: 000004_create_schedules_and_plans.down.sql

DROP INDEX IF EXISTS idx_study_plans_active_student;
DROP INDEX IF EXISTS idx_schedules_day_time;
DROP TABLE IF EXISTS study_plans;
DROP TABLE IF EXISTS class_schedules;
