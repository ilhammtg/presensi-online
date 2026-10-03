package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilham/presensi-online/backend/internal/domain"
)

// ProdiRepo is the PostgreSQL implementation of domain.ProdiMonitoringRepository.
type ProdiRepo struct {
	db *pgxpool.Pool
}

// NewProdiRepo creates a new ProdiRepo.
func NewProdiRepo(db *pgxpool.Pool) *ProdiRepo {
	return &ProdiRepo{db: db}
}

// GetProdiOverview retrieves high-level statistical metrics for the given study program.
func (r *ProdiRepo) GetProdiOverview(ctx context.Context, prodiID uuid.UUID) (*domain.ProdiOverview, error) {
	// Current day of week in WIB (1=Senin ... 7=Minggu)
	weekday := domain.CurrentDayOfWeek()

	overview := &domain.ProdiOverview{ProdiID: prodiID}

	// 1. Prodi info and primary aggregates
	query := `
		SELECT 
			sp.name, sp.code, f.name,
			(SELECT COUNT(*) FROM users WHERE prodi_id = sp.id AND role = 'mahasiswa' AND is_active = true),
			(SELECT COUNT(*) FROM users WHERE prodi_id = sp.id AND role = 'dosen' AND is_active = true),
			(SELECT COUNT(DISTINCT cs.id) FROM class_schedules cs JOIN users u ON cs.lecturer_id = u.id WHERE u.prodi_id = sp.id AND cs.is_active = true),
			(SELECT COUNT(DISTINCT s.id) FROM class_sessions s JOIN class_schedules cs ON s.schedule_id = cs.id JOIN users u ON cs.lecturer_id = u.id WHERE u.prodi_id = sp.id),
			(SELECT COALESCE(ROUND((COUNT(CASE WHEN a.status IN ('hadir', 'terlambat') THEN 1 END)::numeric / NULLIF(COUNT(a.id), 0)) * 100, 1), 0.0) 
			 FROM attendances a 
			 JOIN class_sessions s ON a.session_id = s.id 
			 JOIN class_schedules cs ON s.schedule_id = cs.id 
			 JOIN users u ON cs.lecturer_id = u.id 
			 WHERE u.prodi_id = sp.id),
			(SELECT COUNT(*) FROM class_schedules cs JOIN users u ON cs.lecturer_id = u.id WHERE u.prodi_id = sp.id AND cs.day_of_week = $2 AND cs.is_active = true),
			(SELECT COUNT(DISTINCT s.schedule_id) FROM class_sessions s JOIN class_schedules cs ON s.schedule_id = cs.id JOIN users u ON cs.lecturer_id = u.id WHERE u.prodi_id = sp.id AND s.session_date = CURRENT_DATE)
		FROM study_programs sp 
		JOIN faculties f ON sp.faculty_id = f.id 
		WHERE sp.id = $1
	`

	err := r.db.QueryRow(ctx, query, prodiID, weekday).Scan(
		&overview.ProdiName,
		&overview.ProdiCode,
		&overview.FacultyName,
		&overview.TotalActiveStudents,
		&overview.TotalLecturers,
		&overview.TotalCoursesOffered,
		&overview.TotalSessionsHeld,
		&overview.AvgAttendanceRate,
		&overview.TodayClassesScheduled,
		&overview.TodayClassesCompleted,
	)
	if err != nil {
		return nil, fmt.Errorf("ProdiRepo.GetProdiOverview: %w", err)
	}

	// Calculate teaching compliance: expected target per course is roughly 6 meetings at week 6
	expectedTargetSessions := overview.TotalCoursesOffered * 6
	if expectedTargetSessions > 0 {
		compliance := (float64(overview.TotalSessionsHeld) / float64(expectedTargetSessions)) * 100
		if compliance > 100.0 {
			compliance = 100.0
		}
		overview.TeachingComplianceRate = float64(int(compliance*10)) / 10
	} else {
		overview.TeachingComplianceRate = 100.0
	}

	return overview, nil
}

// GetLecturersCompliance computes progress and adherence for all lecturers in the prodi.
func (r *ProdiRepo) GetLecturersCompliance(ctx context.Context, prodiID uuid.UUID) ([]*domain.LecturerCompliance, error) {
	query := `
		SELECT 
			u.id,
			u.name,
			u.external_id,
			COUNT(DISTINCT cs.id),
			COALESCE(ARRAY_AGG(DISTINCT cs.course_name) FILTER (WHERE cs.course_name IS NOT NULL), '{}'),
			COUNT(DISTINCT s.id),
			MAX(s.session_date)
		FROM users u
		LEFT JOIN class_schedules cs ON u.id = cs.lecturer_id AND cs.is_active = true
		LEFT JOIN class_sessions s ON cs.id = s.schedule_id
		WHERE u.prodi_id = $1 AND u.role = 'dosen' AND u.is_active = true
		GROUP BY u.id, u.name, u.external_id
		ORDER BY u.name ASC
	`

	rows, err := r.db.Query(ctx, query, prodiID)
	if err != nil {
		return nil, fmt.Errorf("ProdiRepo.GetLecturersCompliance: %w", err)
	}
	defer rows.Close()

	var results []*domain.LecturerCompliance
	for rows.Next() {
		item := &domain.LecturerCompliance{}
		var lastDate *time.Time

		err := rows.Scan(
			&item.LecturerID,
			&item.LecturerName,
			&item.NIDN,
			&item.CourseCount,
			&item.CourseNames,
			&item.TotalSessionsHeld,
			&lastDate,
		)
		if err != nil {
			return nil, fmt.Errorf("ProdiRepo.GetLecturersCompliance scan: %w", err)
		}
		item.LastSessionDate = lastDate

		// 16 meetings per course is curriculum target
		item.TargetSessions = item.CourseCount * 16
		if item.CourseCount > 0 {
			// Expected meeting per course at this point in semester is ~6
			expectedMeetings := item.CourseCount * 6
			if expectedMeetings > 0 {
				ratio := (float64(item.TotalSessionsHeld) / float64(expectedMeetings)) * 100
				if ratio > 100 {
					ratio = 100
				}
				item.CompliancePercentage = float64(int(ratio*10)) / 10
			}
		}

		if item.CourseCount == 0 {
			item.Status = "lancar"
			item.CompliancePercentage = 100.0
		} else if item.TotalSessionsHeld >= item.CourseCount*5 {
			item.Status = "lancar"
		} else if item.TotalSessionsHeld >= item.CourseCount*3 {
			item.Status = "perlu_perhatian"
		} else {
			item.Status = "tertinggal"
		}

		results = append(results, item)
	}

	return results, nil
}

// GetStudentsAtRisk identifies students in the prodi whose attendance rate is below 75%.
func (r *ProdiRepo) GetStudentsAtRisk(ctx context.Context, prodiID uuid.UUID) ([]*domain.StudentAtRisk, error) {
	query := `
		SELECT 
			u.id,
			u.external_id,
			u.name,
			sp.name,
			cs.course_code,
			cs.course_name,
			l.name,
			COUNT(DISTINCT s.id),
			COUNT(CASE WHEN a.status IN ('hadir', 'terlambat') THEN 1 END),
			COUNT(CASE WHEN a.status = 'alpa' THEN 1 END),
			COUNT(CASE WHEN a.status IN ('izin', 'sakit') THEN 1 END),
			COALESCE(ROUND((COUNT(CASE WHEN a.status IN ('hadir', 'terlambat') THEN 1 END)::numeric / NULLIF(COUNT(DISTINCT s.id), 0)) * 100, 1), 0.0)
		FROM study_plans sp_plan
		JOIN users u ON sp_plan.student_id = u.id
		JOIN study_programs sp ON u.prodi_id = sp.id
		JOIN class_schedules cs ON sp_plan.schedule_id = cs.id
		JOIN users l ON cs.lecturer_id = l.id
		JOIN class_sessions s ON cs.id = s.schedule_id
		LEFT JOIN attendances a ON s.id = a.session_id AND u.id = a.student_id
		WHERE u.prodi_id = $1 AND sp_plan.status = 'active'
		GROUP BY u.id, u.external_id, u.name, sp.name, cs.course_code, cs.course_name, l.name
		HAVING COUNT(DISTINCT s.id) > 0 
		   AND (COUNT(CASE WHEN a.status IN ('hadir', 'terlambat') THEN 1 END)::numeric / COUNT(DISTINCT s.id)) < 0.75
		ORDER BY 12 ASC, u.name ASC
	`

	rows, err := r.db.Query(ctx, query, prodiID)
	if err != nil {
		return nil, fmt.Errorf("ProdiRepo.GetStudentsAtRisk: %w", err)
	}
	defer rows.Close()

	var results []*domain.StudentAtRisk
	for rows.Next() {
		item := &domain.StudentAtRisk{}
		err := rows.Scan(
			&item.StudentID,
			&item.StudentNIM,
			&item.StudentName,
			&item.ProdiName,
			&item.CourseCode,
			&item.CourseName,
			&item.LecturerName,
			&item.TotalSessionsHeld,
			&item.HadirCount,
			&item.AlpaCount,
			&item.IzinSakitCount,
			&item.AttendanceRate,
		)
		if err != nil {
			return nil, fmt.Errorf("ProdiRepo.GetStudentsAtRisk scan: %w", err)
		}

		if item.AttendanceRate < 50.0 {
			item.Recommendation = "Pemanggilan Kaprodi & Surat Peringatan II"
		} else {
			item.Recommendation = "Peringatan Akademik (Terancam Tidak Lulus/UAS)"
		}

		results = append(results, item)
	}

	return results, nil
}

// GetProdiClasses lists all course classes within the prodi with aggregated stats.
func (r *ProdiRepo) GetProdiClasses(ctx context.Context, prodiID uuid.UUID) ([]*domain.ProdiClassSummary, error) {
	query := `
		SELECT 
			cs.id,
			cs.course_code,
			cs.course_name,
			cs.day_of_week,
			TO_CHAR(cs.start_time, 'HH24:MI') || ' - ' || TO_CHAR(cs.end_time, 'HH24:MI') || ' WIB',
			rm.name || ' (' || bld.name || ')',
			l.name,
			l.external_id,
			(SELECT COUNT(*) FROM study_plans sp_sub WHERE sp_sub.schedule_id = cs.id AND sp_sub.status = 'active'),
			(SELECT COUNT(*) FROM class_sessions s_sub WHERE s_sub.schedule_id = cs.id),
			(SELECT COALESCE(ROUND((COUNT(CASE WHEN a.status IN ('hadir', 'terlambat') THEN 1 END)::numeric / NULLIF(COUNT(a.id), 0)) * 100, 1), 0.0)
			 FROM attendances a 
			 JOIN class_sessions s_sub2 ON a.session_id = s_sub2.id 
			 WHERE s_sub2.schedule_id = cs.id)
		FROM class_schedules cs
		JOIN users l ON cs.lecturer_id = l.id
		JOIN rooms rm ON cs.room_id = rm.id
		JOIN buildings bld ON rm.building_id = bld.id
		WHERE l.prodi_id = $1 AND cs.is_active = true
		ORDER BY cs.day_of_week ASC, cs.start_time ASC
	`

	rows, err := r.db.Query(ctx, query, prodiID)
	if err != nil {
		return nil, fmt.Errorf("ProdiRepo.GetProdiClasses: %w", err)
	}
	defer rows.Close()

	dayNames := map[int]string{
		1: "Senin", 2: "Selasa", 3: "Rabu", 4: "Kamis", 5: "Jumat", 6: "Sabtu", 7: "Minggu",
	}

	var results []*domain.ProdiClassSummary
	for rows.Next() {
		item := &domain.ProdiClassSummary{}
		err := rows.Scan(
			&item.ScheduleID,
			&item.CourseCode,
			&item.CourseName,
			&item.DayOfWeek,
			&item.TimeRange,
			&item.RoomName,
			&item.LecturerName,
			&item.LecturerNIDN,
			&item.TotalEnrolled,
			&item.CompletedSessions,
			&item.AvgAttendanceRate,
		)
		if err != nil {
			return nil, fmt.Errorf("ProdiRepo.GetProdiClasses scan: %w", err)
		}
		item.DayName = dayNames[item.DayOfWeek]
		results = append(results, item)
	}

	return results, nil
}

// GetLiveTodayClasses retrieves real-time status of classes for today in the given study program.
func (r *ProdiRepo) GetLiveTodayClasses(ctx context.Context, prodiID uuid.UUID) ([]*domain.ProdiLiveClassItem, error) {
	weekday := domain.CurrentDayOfWeek()

	query := `
		SELECT 
			cs.id AS schedule_id,
			sess.id AS session_id,
			cs.course_code,
			cs.course_name,
			l.name AS lecturer_name,
			l.external_id AS lecturer_nidn,
			rm.name || ' (' || bld.name || ')' AS room_name,
			cs.day_of_week,
			to_char(cs.start_time, 'HH24:MI') AS start_time,
			to_char(cs.end_time, 'HH24:MI') AS end_time,
			(SELECT COUNT(*) FROM study_plans sp WHERE sp.schedule_id = cs.id AND sp.status = 'active') AS total_enrolled,
			COALESCE((SELECT COUNT(*) FROM attendances a WHERE a.session_id = sess.id AND a.status IN ('hadir', 'terlambat')), 0) AS total_hadir,
			COALESCE((SELECT COUNT(*) FROM attendances a WHERE a.session_id = sess.id AND a.status IN ('izin', 'sakit')), 0) AS total_izin_sakit,
			sess.is_open,
			sess.opened_at,
			sess.closed_at,
			CURRENT_TIME > (cs.start_time + INTERVAL '30 minutes') AS is_past_start
		FROM class_schedules cs
		JOIN users l ON cs.lecturer_id = l.id
		JOIN rooms rm ON cs.room_id = rm.id
		JOIN buildings bld ON rm.building_id = bld.id
		LEFT JOIN class_sessions sess ON sess.schedule_id = cs.id AND sess.session_date = CURRENT_DATE
		WHERE l.prodi_id = $1 AND cs.is_active = TRUE 
		  AND (cs.day_of_week = $2 OR NOT EXISTS (
		      SELECT 1 FROM class_schedules sub_cs 
		      JOIN users sub_l ON sub_cs.lecturer_id = sub_l.id 
		      WHERE sub_l.prodi_id = $1 AND sub_cs.day_of_week = $2 AND sub_cs.is_active = TRUE
		  ))
		ORDER BY cs.start_time ASC
	`

	rows, err := r.db.Query(ctx, query, prodiID, weekday)
	if err != nil {
		return nil, fmt.Errorf("ProdiRepo.GetLiveTodayClasses: %w", err)
	}
	defer rows.Close()

	var results []*domain.ProdiLiveClassItem
	for rows.Next() {
		item := &domain.ProdiLiveClassItem{}
		var sessionID *uuid.UUID
		var isOpen *bool
		var openedAt, closedAt *time.Time
		var isPastStart bool

		err := rows.Scan(
			&item.ScheduleID,
			&sessionID,
			&item.CourseCode,
			&item.CourseName,
			&item.LecturerName,
			&item.LecturerNIDN,
			&item.RoomName,
			&item.DayOfWeek,
			&item.StartTime,
			&item.EndTime,
			&item.TotalEnrolled,
			&item.TotalHadir,
			&item.TotalIzinSakit,
			&isOpen,
			&openedAt,
			&closedAt,
			&isPastStart,
		)
		if err != nil {
			return nil, fmt.Errorf("ProdiRepo.GetLiveTodayClasses scan: %w", err)
		}

		item.SessionID = sessionID
		item.OpenedAt = openedAt
		item.ClosedAt = closedAt

		// Calculate status according to requirements:
		// SEDANG_BERLANGSUNG (sesi aktif), SELESAI (sesi sudah ditutup), TIDAK_MASUK (terlewat tanpa sesi), BELUM_DIMULAI
		if isOpen != nil && *isOpen {
			item.Status = "SEDANG_BERLANGSUNG"
		} else if sessionID != nil && closedAt != nil {
			item.Status = "SELESAI"
		} else if isPastStart {
			item.Status = "TIDAK_MASUK"
		} else {
			item.Status = "BELUM_DIMULAI"
		}

		results = append(results, item)
	}

	return results, nil
}
