package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilham/presensi-online/backend/internal/domain"
)

// FacultyRepo implements domain.FacultyRepository.
type FacultyRepo struct{ db *pgxpool.Pool }
func NewFacultyRepo(db *pgxpool.Pool) *FacultyRepo { return &FacultyRepo{db: db} }

func (r *FacultyRepo) Upsert(ctx context.Context, f *domain.Faculty) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO faculties (id, code, name) VALUES ($1, $2, $3)
		ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, updated_at = CURRENT_TIMESTAMP
	`, f.ID, f.Code, f.Name)
	return err
}

func (r *FacultyRepo) FindByCode(ctx context.Context, code string) (*domain.Faculty, error) {
	f := &domain.Faculty{}
	err := r.db.QueryRow(ctx, `SELECT id, code, name, created_at, updated_at FROM faculties WHERE code = $1`, code).
		Scan(&f.ID, &f.Code, &f.Name, &f.CreatedAt, &f.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) { return nil, domain.ErrNotFound }
	if err != nil { return nil, err }
	return f, nil
}

func (r *FacultyRepo) FindAll(ctx context.Context) ([]*domain.Faculty, error) {
	rows, err := r.db.Query(ctx, `SELECT id, code, name, created_at, updated_at FROM faculties ORDER BY name`)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []*domain.Faculty
	for rows.Next() {
		f := &domain.Faculty{}
		if err := rows.Scan(&f.ID, &f.Code, &f.Name, &f.CreatedAt, &f.UpdatedAt); err != nil { return nil, err }
		out = append(out, f)
	}
	return out, rows.Err()
}

// StudyProgramRepo implements domain.StudyProgramRepository.
type StudyProgramRepo struct{ db *pgxpool.Pool }
func NewStudyProgramRepo(db *pgxpool.Pool) *StudyProgramRepo { return &StudyProgramRepo{db: db} }

func (r *StudyProgramRepo) Upsert(ctx context.Context, sp *domain.StudyProgram) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO study_programs (id, faculty_id, code, name) VALUES ($1, $2, $3, $4)
		ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, faculty_id = EXCLUDED.faculty_id, updated_at = CURRENT_TIMESTAMP
	`, sp.ID, sp.FacultyID, sp.Code, sp.Name)
	return err
}

func (r *StudyProgramRepo) FindByCode(ctx context.Context, code string) (*domain.StudyProgram, error) {
	sp := &domain.StudyProgram{}
	err := r.db.QueryRow(ctx, `SELECT id, faculty_id, code, name, created_at, updated_at FROM study_programs WHERE code = $1`, code).
		Scan(&sp.ID, &sp.FacultyID, &sp.Code, &sp.Name, &sp.CreatedAt, &sp.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) { return nil, domain.ErrNotFound }
	if err != nil { return nil, err }
	return sp, nil
}

func (r *StudyProgramRepo) FindByFacultyID(ctx context.Context, facultyID uuid.UUID) ([]*domain.StudyProgram, error) {
	rows, err := r.db.Query(ctx, `SELECT id, faculty_id, code, name, created_at, updated_at FROM study_programs WHERE faculty_id = $1`, facultyID)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []*domain.StudyProgram
	for rows.Next() {
		sp := &domain.StudyProgram{}
		if err := rows.Scan(&sp.ID, &sp.FacultyID, &sp.Code, &sp.Name, &sp.CreatedAt, &sp.UpdatedAt); err != nil { return nil, err }
		out = append(out, sp)
	}
	return out, rows.Err()
}

// BuildingRepo implements domain.BuildingRepository.
type BuildingRepo struct{ db *pgxpool.Pool }
func NewBuildingRepo(db *pgxpool.Pool) *BuildingRepo { return &BuildingRepo{db: db} }

func (r *BuildingRepo) Upsert(ctx context.Context, b *domain.Building) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO buildings (id, code, name) VALUES ($1, $2, $3)
		ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
	`, b.ID, b.Code, b.Name)
	return err
}

func (r *BuildingRepo) FindByCode(ctx context.Context, code string) (*domain.Building, error) {
	b := &domain.Building{}
	err := r.db.QueryRow(ctx, `SELECT id, code, name, created_at FROM buildings WHERE code = $1`, code).
		Scan(&b.ID, &b.Code, &b.Name, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) { return nil, domain.ErrNotFound }
	if err != nil { return nil, err }
	return b, nil
}

// RoomRepo implements domain.RoomRepository.
type RoomRepo struct{ db *pgxpool.Pool }
func NewRoomRepo(db *pgxpool.Pool) *RoomRepo { return &RoomRepo{db: db} }

func (r *RoomRepo) Upsert(ctx context.Context, rm *domain.Room) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO rooms (id, building_id, room_code, name, latitude, longitude, radius_meters, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (room_code) DO UPDATE SET
			name = EXCLUDED.name, latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude, radius_meters = EXCLUDED.radius_meters
	`, rm.ID, rm.BuildingID, rm.RoomCode, rm.Name, rm.Latitude, rm.Longitude, rm.RadiusMeters, rm.IsActive)
	return err
}

func (r *RoomRepo) FindByCode(ctx context.Context, roomCode string) (*domain.Room, error) {
	rm := &domain.Room{}
	err := r.db.QueryRow(ctx, `
		SELECT id, building_id, room_code, name, latitude, longitude, radius_meters, is_active, created_at
		FROM rooms WHERE room_code = $1
	`, roomCode).Scan(&rm.ID, &rm.BuildingID, &rm.RoomCode, &rm.Name, &rm.Latitude, &rm.Longitude, &rm.RadiusMeters, &rm.IsActive, &rm.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) { return nil, domain.ErrNotFound }
	if err != nil { return nil, err }
	return rm, nil
}

func (r *RoomRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	rm := &domain.Room{}
	err := r.db.QueryRow(ctx, `
		SELECT id, building_id, room_code, name, latitude, longitude, radius_meters, is_active, created_at
		FROM rooms WHERE id = $1
	`, id).Scan(&rm.ID, &rm.BuildingID, &rm.RoomCode, &rm.Name, &rm.Latitude, &rm.Longitude, &rm.RadiusMeters, &rm.IsActive, &rm.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) { return nil, domain.ErrNotFound }
	if err != nil { return nil, err }
	return rm, nil
}

// ClassScheduleRepo implements domain.ClassScheduleRepository.
type ClassScheduleRepo struct{ db *pgxpool.Pool }
func NewClassScheduleRepo(db *pgxpool.Pool) *ClassScheduleRepo { return &ClassScheduleRepo{db: db} }

func (r *ClassScheduleRepo) Upsert(ctx context.Context, cs *domain.ClassSchedule) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO class_schedules (id, external_id, course_code, course_name, academic_year, semester_type,
			lecturer_id, room_id, day_of_week, start_time, end_time, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (external_id) DO UPDATE SET
			course_name = EXCLUDED.course_name, room_id = EXCLUDED.room_id,
			day_of_week = EXCLUDED.day_of_week, start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time, updated_at = CURRENT_TIMESTAMP
	`, cs.ID, cs.ExternalID, cs.CourseCode, cs.CourseName, cs.AcademicYear, cs.SemesterType,
		cs.LecturerID, cs.RoomID, cs.DayOfWeek, cs.StartTime, cs.EndTime, cs.IsActive)
	return err
}

func (r *ClassScheduleRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.ClassSchedule, error) {
	return r.scanOne(ctx, `WHERE id = $1`, id)
}

func (r *ClassScheduleRepo) FindByExternalID(ctx context.Context, extID string) (*domain.ClassSchedule, error) {
	return r.scanOne(ctx, `WHERE external_id = $1`, extID)
}

func (r *ClassScheduleRepo) FindTodayByLecturer(ctx context.Context, lecturerID uuid.UUID, dayOfWeek int) ([]*domain.ClassSchedule, error) {
	return r.scanMany(ctx, `WHERE lecturer_id = $1 AND day_of_week = $2 AND is_active = TRUE AND deleted_at IS NULL`, lecturerID, dayOfWeek)
}

func (r *ClassScheduleRepo) FindTodayByStudent(ctx context.Context, studentID uuid.UUID, dayOfWeek int) ([]*domain.ClassSchedule, error) {
	rows, err := r.db.Query(ctx, `
		SELECT cs.id, cs.external_id, cs.course_code, cs.course_name, cs.academic_year, cs.semester_type,
		       cs.lecturer_id, cs.room_id, cs.day_of_week, cs.start_time, cs.end_time, cs.is_active,
		       cs.created_at, cs.updated_at, cs.deleted_at
		FROM class_schedules cs
		JOIN study_plans sp ON sp.schedule_id = cs.id
		WHERE sp.student_id = $1 AND cs.day_of_week = $2 AND sp.status = 'active'
		  AND cs.is_active = TRUE AND cs.deleted_at IS NULL
		ORDER BY cs.start_time
	`, studentID, dayOfWeek)
	if err != nil { return nil, err }
	defer rows.Close()
	return scanScheduleRows(rows)
}

// FindByLecturerWithDetails retrieves all class schedules for a lecturer with room, building, student count, and active session status.
func (r *ClassScheduleRepo) FindByLecturerWithDetails(ctx context.Context, lecturerID uuid.UUID) ([]*domain.ClassScheduleDetail, error) {
	today := time.Now()
	todayDayOfWeek := int(today.Weekday())
	if todayDayOfWeek == 0 {
		todayDayOfWeek = 7
	}

	rows, err := r.db.Query(ctx, `
		SELECT 
			cs.id, cs.course_code, cs.course_name, cs.academic_year, cs.semester_type,
			cs.lecturer_id, cs.room_id, r.name AS room_name, b.name AS building_name,
			cs.day_of_week, 
			to_char(cs.start_time, 'HH24:MI') AS start_time_str,
			to_char(cs.end_time, 'HH24:MI') AS end_time_str,
			(SELECT COUNT(*) FROM study_plans sp WHERE sp.schedule_id = cs.id AND sp.status = 'active') AS enrolled_count,
			sess.id AS active_session_id,
			COALESCE((SELECT MAX(csess.meeting_no) FROM class_sessions csess WHERE csess.schedule_id = cs.id), 0) + 1 AS next_meeting_no
		FROM class_schedules cs
		JOIN rooms r ON cs.room_id = r.id
		JOIN buildings b ON r.building_id = b.id
		LEFT JOIN class_sessions sess ON sess.schedule_id = cs.id AND sess.is_open = TRUE
		WHERE cs.lecturer_id = $1 AND cs.is_active = TRUE AND cs.deleted_at IS NULL
		ORDER BY cs.day_of_week ASC, cs.start_time ASC
	`, lecturerID)
	if err != nil {
		return nil, fmt.Errorf("ClassScheduleRepo.FindByLecturerWithDetails: %w", err)
	}
	defer rows.Close()

	dayNames := map[int]string{
		1: "Senin", 2: "Selasa", 3: "Rabu", 4: "Kamis", 5: "Jumat", 6: "Sabtu", 7: "Minggu",
	}

	var list []*domain.ClassScheduleDetail
	for rows.Next() {
		d := &domain.ClassScheduleDetail{}
		var activeSessID *uuid.UUID
		err := rows.Scan(
			&d.ID, &d.CourseCode, &d.CourseName, &d.AcademicYear, &d.SemesterType,
			&d.LecturerID, &d.RoomID, &d.RoomName, &d.BuildingName,
			&d.DayOfWeek, &d.StartTime, &d.EndTime,
			&d.EnrolledCount, &activeSessID, &d.NextMeetingNo,
		)
		if err != nil {
			return nil, err
		}

		d.ClassUnit = "Unit 01"
		d.DayName = dayNames[d.DayOfWeek]
		d.IsToday = (d.DayOfWeek == todayDayOfWeek)
		if activeSessID != nil {
			d.HasActiveSession = true
			d.ActiveSessionID = activeSessID
		}

		list = append(list, d)
	}
	return list, rows.Err()
}

// FindByStudentWithDetails retrieves all enrolled class schedules for a student via study_plans,
// enriched with room, building, lecturer name, student count, and active session status.
func (r *ClassScheduleRepo) FindByStudentWithDetails(ctx context.Context, studentID uuid.UUID) ([]*domain.ClassScheduleDetail, error) {
	today := time.Now()
	todayDayOfWeek := int(today.Weekday())
	if todayDayOfWeek == 0 {
		todayDayOfWeek = 7
	}

	rows, err := r.db.Query(ctx, `
		SELECT 
			cs.id, cs.course_code, cs.course_name, cs.academic_year, cs.semester_type,
			cs.lecturer_id, cs.room_id, r.name AS room_name, b.name AS building_name,
			cs.day_of_week,
			to_char(cs.start_time, 'HH24:MI') AS start_time_str,
			to_char(cs.end_time, 'HH24:MI') AS end_time_str,
			(SELECT COUNT(*) FROM study_plans sp2 WHERE sp2.schedule_id = cs.id AND sp2.status = 'active') AS enrolled_count,
			sess.id AS active_session_id,
			COALESCE((SELECT MAX(csess.meeting_no) FROM class_sessions csess WHERE csess.schedule_id = cs.id), 0) + 1 AS next_meeting_no
		FROM study_plans sp
		JOIN class_schedules cs ON sp.schedule_id = cs.id
		JOIN rooms r ON cs.room_id = r.id
		JOIN buildings b ON r.building_id = b.id
		LEFT JOIN class_sessions sess ON sess.schedule_id = cs.id AND sess.is_open = TRUE
		WHERE sp.student_id = $1 AND sp.status = 'active'
		  AND cs.is_active = TRUE AND cs.deleted_at IS NULL
		ORDER BY cs.day_of_week ASC, cs.start_time ASC
	`, studentID)
	if err != nil {
		return nil, fmt.Errorf("ClassScheduleRepo.FindByStudentWithDetails: %w", err)
	}
	defer rows.Close()

	dayNames := map[int]string{
		1: "Senin", 2: "Selasa", 3: "Rabu", 4: "Kamis", 5: "Jumat", 6: "Sabtu", 7: "Minggu",
	}

	var list []*domain.ClassScheduleDetail
	for rows.Next() {
		d := &domain.ClassScheduleDetail{}
		var activeSessID *uuid.UUID
		err := rows.Scan(
			&d.ID, &d.CourseCode, &d.CourseName, &d.AcademicYear, &d.SemesterType,
			&d.LecturerID, &d.RoomID, &d.RoomName, &d.BuildingName,
			&d.DayOfWeek, &d.StartTime, &d.EndTime,
			&d.EnrolledCount, &activeSessID, &d.NextMeetingNo,
		)
		if err != nil {
			return nil, err
		}

		d.ClassUnit = "Unit 01"
		d.DayName = dayNames[d.DayOfWeek]
		d.IsToday = (d.DayOfWeek == todayDayOfWeek)
		if activeSessID != nil {
			d.HasActiveSession = true
			d.ActiveSessionID = activeSessID
		}

		list = append(list, d)
	}
	return list, rows.Err()
}

func (r *ClassScheduleRepo) scanOne(ctx context.Context, where string, args ...interface{}) (*domain.ClassSchedule, error) {
	q := `SELECT id, external_id, course_code, course_name, academic_year, semester_type,
		         lecturer_id, room_id, day_of_week, start_time, end_time, is_active,
		         created_at, updated_at, deleted_at
		  FROM class_schedules ` + where
	cs := &domain.ClassSchedule{}
	err := r.db.QueryRow(ctx, q, args...).Scan(
		&cs.ID, &cs.ExternalID, &cs.CourseCode, &cs.CourseName, &cs.AcademicYear, &cs.SemesterType,
		&cs.LecturerID, &cs.RoomID, &cs.DayOfWeek, &cs.StartTime, &cs.EndTime, &cs.IsActive,
		&cs.CreatedAt, &cs.UpdatedAt, &cs.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) { return nil, domain.ErrNotFound }
	if err != nil { return nil, fmt.Errorf("ClassScheduleRepo.scanOne: %w", err) }
	return cs, nil
}

func (r *ClassScheduleRepo) scanMany(ctx context.Context, where string, args ...interface{}) ([]*domain.ClassSchedule, error) {
	q := `SELECT id, external_id, course_code, course_name, academic_year, semester_type,
		         lecturer_id, room_id, day_of_week, start_time, end_time, is_active,
		         created_at, updated_at, deleted_at
		  FROM class_schedules ` + where + ` ORDER BY start_time`
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil { return nil, err }
	defer rows.Close()
	return scanScheduleRows(rows)
}

func scanScheduleRows(rows pgx.Rows) ([]*domain.ClassSchedule, error) {
	var out []*domain.ClassSchedule
	for rows.Next() {
		cs := &domain.ClassSchedule{}
		if err := rows.Scan(
			&cs.ID, &cs.ExternalID, &cs.CourseCode, &cs.CourseName, &cs.AcademicYear, &cs.SemesterType,
			&cs.LecturerID, &cs.RoomID, &cs.DayOfWeek, &cs.StartTime, &cs.EndTime, &cs.IsActive,
			&cs.CreatedAt, &cs.UpdatedAt, &cs.DeletedAt,
		); err != nil { return nil, err }
		out = append(out, cs)
	}
	return out, rows.Err()
}

// StudyPlanRepo implements domain.StudyPlanRepository.
type StudyPlanRepo struct{ db *pgxpool.Pool }
func NewStudyPlanRepo(db *pgxpool.Pool) *StudyPlanRepo { return &StudyPlanRepo{db: db} }

func (r *StudyPlanRepo) Upsert(ctx context.Context, sp *domain.StudyPlan) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO study_plans (id, student_id, schedule_id, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (student_id, schedule_id) DO UPDATE SET status = EXCLUDED.status
	`, sp.ID, sp.StudentID, sp.ScheduleID, sp.Status)
	return err
}

func (r *StudyPlanRepo) IsEnrolled(ctx context.Context, studentID, scheduleID uuid.UUID) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM study_plans
		WHERE student_id = $1 AND schedule_id = $2 AND status = 'active'
	`, studentID, scheduleID).Scan(&count)
	return count > 0, err
}

func (r *StudyPlanRepo) FindByStudentID(ctx context.Context, studentID uuid.UUID) ([]*domain.StudyPlan, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, student_id, schedule_id, status, created_at
		FROM study_plans WHERE student_id = $1 AND status = 'active'
	`, studentID)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []*domain.StudyPlan
	for rows.Next() {
		sp := &domain.StudyPlan{}
		if err := rows.Scan(&sp.ID, &sp.StudentID, &sp.ScheduleID, &sp.Status, &sp.CreatedAt); err != nil { return nil, err }
		out = append(out, sp)
	}
	return out, rows.Err()
}

// FindEnrolledStudentsByScheduleID returns all active enrolled students in a schedule.
func (r *StudyPlanRepo) FindEnrolledStudentsByScheduleID(ctx context.Context, scheduleID uuid.UUID) ([]*domain.EnrolledStudent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.external_id, u.name, u.email, sp.status
		FROM study_plans sp
		JOIN users u ON sp.student_id = u.id
		WHERE sp.schedule_id = $1 AND sp.status = 'active' AND u.deleted_at IS NULL
		ORDER BY u.external_id ASC
	`, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("StudyPlanRepo.FindEnrolledStudentsByScheduleID: %w", err)
	}
	defer rows.Close()

	var students []*domain.EnrolledStudent
	for rows.Next() {
		s := &domain.EnrolledStudent{}
		if err := rows.Scan(&s.StudentID, &s.StudentNIM, &s.StudentName, &s.Email, &s.Status); err != nil {
			return nil, err
		}
		students = append(students, s)
	}
	return students, rows.Err()
}
