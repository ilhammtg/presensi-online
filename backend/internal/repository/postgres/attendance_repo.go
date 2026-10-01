package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilham/presensi-online/backend/internal/domain"
)

// AttendanceRepo is the PostgreSQL implementation of domain.AttendanceRepository.
type AttendanceRepo struct {
	db *pgxpool.Pool
}

// NewAttendanceRepo creates a new AttendanceRepo.
func NewAttendanceRepo(db *pgxpool.Pool) *AttendanceRepo {
	return &AttendanceRepo{db: db}
}

// Create inserts a new attendance record.
// The UNIQUE constraint (session_id, student_id) on the DB level prevents double-submission.
func (r *AttendanceRepo) Create(ctx context.Context, a *domain.Attendance) error {
	subSource := a.SubmissionSource
	if subSource == "" {
		subSource = domain.SubmissionSelfScan
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO attendances (
			id, session_id, student_id, status, scanned_at, device_id,
			latitude, longitude, distance_meters, attachment_url, notes,
			submission_source, updated_by, verified_by_lecturer
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`,
		a.ID, a.SessionID, a.StudentID, a.Status, a.ScannedAt, a.DeviceID,
		a.Latitude, a.Longitude, a.DistanceMeters, a.AttachmentURL, a.Notes,
		subSource, a.UpdatedBy, a.VerifiedByLecturer,
	)
	if err != nil {
		return fmt.Errorf("AttendanceRepo.Create: %w", err)
	}
	return nil
}

// FindBySessionAndStudent retrieves a specific attendance record.
// Used to check for duplicate submission (step 7 validation).
func (r *AttendanceRepo) FindBySessionAndStudent(ctx context.Context, sessionID, studentID uuid.UUID) (*domain.Attendance, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, session_id, student_id, status, scanned_at, device_id,
		       latitude, longitude, distance_meters, attachment_url, notes,
		       submission_source, updated_by, verified_by_lecturer, created_at, updated_at
		FROM attendances
		WHERE session_id = $1 AND student_id = $2
	`, sessionID, studentID)
	return scanAttendance(row)
}

// FindBySessionID retrieves all attendance records for a session.
// Used by the lecturer's realtime monitoring view.
func (r *AttendanceRepo) FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*domain.Attendance, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, session_id, student_id, status, scanned_at, device_id,
		       latitude, longitude, distance_meters, attachment_url, notes,
		       submission_source, updated_by, verified_by_lecturer, created_at, updated_at
		FROM attendances
		WHERE session_id = $1
		ORDER BY scanned_at ASC
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("AttendanceRepo.FindBySessionID: %w", err)
	}
	defer rows.Close()

	var result []*domain.Attendance
	for rows.Next() {
		a, err := scanAttendanceRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

// GetAttendeesBySessionID retrieves attendance records enriched with student profile.
func (r *AttendanceRepo) GetAttendeesBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*domain.AttendeeDetail, error) {
	rows, err := r.db.Query(ctx, `
		SELECT a.id, a.student_id, u.external_id, u.name,
		       a.status, a.scanned_at, a.distance_meters, a.notes,
		       a.submission_source, a.attachment_url, a.updated_by, a.verified_by_lecturer
		FROM attendances a
		JOIN users u ON a.student_id = u.id
		WHERE a.session_id = $1
		ORDER BY a.scanned_at ASC NULLS LAST
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("AttendanceRepo.GetAttendeesBySessionID: %w", err)
	}
	defer rows.Close()

	var result []*domain.AttendeeDetail
	for rows.Next() {
		detail := &domain.AttendeeDetail{}
		if err := rows.Scan(
			&detail.AttendanceID,
			&detail.StudentID,
			&detail.StudentNIM,
			&detail.StudentName,
			&detail.Status,
			&detail.ScannedAt,
			&detail.DistanceMeters,
			&detail.Notes,
			&detail.SubmissionSource,
			&detail.AttachmentURL,
			&detail.UpdatedBy,
			&detail.VerifiedByLecturer,
		); err != nil {
			return nil, fmt.Errorf("AttendanceRepo.GetAttendeesBySessionID scan: %w", err)
		}
		result = append(result, detail)
	}
	return result, rows.Err()
}

// Update performs a manual override update on an attendance record (lecturer action).
func (r *AttendanceRepo) Update(ctx context.Context, a *domain.Attendance) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE attendances
		SET status = $1, notes = $2, verified_by_lecturer = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`, a.Status, a.Notes, a.VerifiedByLecturer, a.ID)
	if err != nil {
		return fmt.Errorf("AttendanceRepo.Update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetSummaryByStudentAndSchedule returns attendance count per status for reporting.
func (r *AttendanceRepo) GetSummaryByStudentAndSchedule(ctx context.Context, studentID, scheduleID uuid.UUID) (map[domain.AttendanceStatus]int, error) {
	rows, err := r.db.Query(ctx, `
		SELECT a.status, COUNT(*) as cnt
		FROM attendances a
		JOIN class_sessions cs ON cs.id = a.session_id
		WHERE a.student_id = $1 AND cs.schedule_id = $2
		GROUP BY a.status
	`, studentID, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("AttendanceRepo.GetSummary: %w", err)
	}
	defer rows.Close()

	summary := make(map[domain.AttendanceStatus]int)
	for rows.Next() {
		var status domain.AttendanceStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		summary[status] = count
	}
	return summary, rows.Err()
}

// MarkPermission marks a student as izin or sakit for a session (with audit trail: manual_lecturer).
func (r *AttendanceRepo) MarkPermission(ctx context.Context, sessionID, studentID, lecturerID uuid.UUID, status domain.AttendanceStatus, notes string) error {
	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO attendances (
			id, session_id, student_id, status, notes, submission_source, updated_by, verified_by_lecturer, scanned_at, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, 'manual_lecturer', $5, TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		ON CONFLICT (session_id, student_id) DO UPDATE SET
			status = EXCLUDED.status,
			notes = EXCLUDED.notes,
			submission_source = 'manual_lecturer',
			updated_by = EXCLUDED.updated_by,
			verified_by_lecturer = TRUE,
			scanned_at = COALESCE(attendances.scanned_at, CURRENT_TIMESTAMP),
			updated_at = CURRENT_TIMESTAMP
	`, sessionID, studentID, status, notesPtr, lecturerID)
	if err != nil {
		return fmt.Errorf("AttendanceRepo.MarkPermission: %w", err)
	}
	return nil
}

// GetPermissionsByLecturer retrieves permission submissions for all classes taught by a lecturer.
func (r *AttendanceRepo) GetPermissionsByLecturer(ctx context.Context, lecturerID uuid.UUID, pendingOnly bool) ([]*domain.PermissionApprovalItem, error) {
	query := `
		SELECT a.id, a.session_id, cs.meeting_no, TO_CHAR(cs.session_date, 'YYYY-MM-DD'),
		       csc.id as schedule_id, csc.course_code, csc.course_name, 'Unit 01' as class_unit,
		       u.id as student_id, u.external_id as student_nim, u.name as student_name,
		       a.status, a.submission_source, a.attachment_url, a.notes,
		       a.verified_by_lecturer, a.updated_by, a.created_at
		FROM attendances a
		JOIN class_sessions cs ON cs.id = a.session_id
		JOIN class_schedules csc ON csc.id = cs.schedule_id
		JOIN users u ON u.id = a.student_id
		WHERE csc.lecturer_id = $1
		  AND (a.status IN ('izin', 'sakit') OR a.submission_source = 'app_request')
	`
	if pendingOnly {
		query += ` AND a.verified_by_lecturer = FALSE`
	}
	query += ` ORDER BY a.created_at DESC`

	rows, err := r.db.Query(ctx, query, lecturerID)
	if err != nil {
		return nil, fmt.Errorf("AttendanceRepo.GetPermissionsByLecturer: %w", err)
	}
	defer rows.Close()

	var items []*domain.PermissionApprovalItem
	for rows.Next() {
		item := &domain.PermissionApprovalItem{}
		if err := rows.Scan(
			&item.AttendanceID,
			&item.SessionID,
			&item.MeetingNo,
			&item.SessionDate,
			&item.ScheduleID,
			&item.CourseCode,
			&item.CourseName,
			&item.ClassUnit,
			&item.StudentID,
			&item.StudentNIM,
			&item.StudentName,
			&item.Status,
			&item.SubmissionSource,
			&item.AttachmentURL,
			&item.Notes,
			&item.VerifiedByLecturer,
			&item.UpdatedBy,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("AttendanceRepo.GetPermissionsByLecturer scan: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ApprovePermission approves or rejects a student's permission request.
func (r *AttendanceRepo) ApprovePermission(ctx context.Context, attendanceID, lecturerID uuid.UUID, approved bool, notes string) error {
	var tag pgconn.CommandTag
	var err error
	if approved {
		tag, err = r.db.Exec(ctx, `
			UPDATE attendances
			SET verified_by_lecturer = TRUE,
			    updated_by = $1,
			    notes = CASE WHEN $2 <> '' THEN $2 ELSE notes END,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $3
		`, lecturerID, notes, attendanceID)
	} else {
		tag, err = r.db.Exec(ctx, `
			UPDATE attendances
			SET status = 'alpa',
			    verified_by_lecturer = TRUE,
			    updated_by = $1,
			    notes = CASE WHEN $2 <> '' THEN $2 ELSE 'Permohonan izin ditolak dosen' END,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $3
		`, lecturerID, notes, attendanceID)
	}
	if err != nil {
		return fmt.Errorf("AttendanceRepo.ApprovePermission: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// SubmitPermissionRequest allows student to submit permission request with attachment.
func (r *AttendanceRepo) SubmitPermissionRequest(ctx context.Context, sessionID, studentID uuid.UUID, status domain.AttendanceStatus, notes string, attachmentURL *string) (*domain.Attendance, error) {
	id := uuid.New()
	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO attendances (
			id, session_id, student_id, status, notes, attachment_url,
			submission_source, verified_by_lecturer, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, 'app_request', FALSE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		ON CONFLICT (session_id, student_id) DO UPDATE SET
			status = EXCLUDED.status,
			notes = EXCLUDED.notes,
			attachment_url = EXCLUDED.attachment_url,
			submission_source = 'app_request',
			verified_by_lecturer = FALSE,
			updated_at = CURRENT_TIMESTAMP
	`, id, sessionID, studentID, status, notesPtr, attachmentURL)
	if err != nil {
		return nil, fmt.Errorf("AttendanceRepo.SubmitPermissionRequest: %w", err)
	}
	return r.FindBySessionAndStudent(ctx, sessionID, studentID)
}

// AutoMarkAlpa marks all enrolled students who have no attendance record for this session as alpa.
func (r *AttendanceRepo) AutoMarkAlpa(ctx context.Context, sessionID, scheduleID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO attendances (id, session_id, student_id, status, verified_by_lecturer, created_at, updated_at)
		SELECT gen_random_uuid(), $1, sp.student_id, 'alpa', TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		FROM study_plans sp
		WHERE sp.schedule_id = $2
		  AND sp.status = 'active'
		  AND NOT EXISTS (
			  SELECT 1 FROM attendances a WHERE a.session_id = $1 AND a.student_id = sp.student_id
		  )
	`, sessionID, scheduleID)
	if err != nil {
		return fmt.Errorf("AttendanceRepo.AutoMarkAlpa: %w", err)
	}
	return nil
}

// GetClassRecap computes complete matrix attendance recap for a schedule.
func (r *AttendanceRepo) GetClassRecap(ctx context.Context, scheduleID uuid.UUID) (*domain.ClassAttendanceRecap, error) {
	// 1. Get schedule info
	var courseName string
	err := r.db.QueryRow(ctx, `SELECT course_name FROM class_schedules WHERE id = $1`, scheduleID).Scan(&courseName)
	if err != nil {
		return nil, fmt.Errorf("GetClassRecap.schedule: %w", err)
	}

	// 2. Get all meetings conducted
	meetingRows, err := r.db.Query(ctx, `
		SELECT meeting_no FROM class_sessions WHERE schedule_id = $1 ORDER BY meeting_no ASC
	`, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("GetClassRecap.meetings: %w", err)
	}
	defer meetingRows.Close()

	var meetings []int
	for meetingRows.Next() {
		var m int
		if err := meetingRows.Scan(&m); err == nil {
			meetings = append(meetings, m)
		}
	}

	// 3. Get all enrolled students
	studRows, err := r.db.Query(ctx, `
		SELECT u.id, u.external_id, u.name
		FROM study_plans sp
		JOIN users u ON sp.student_id = u.id
		WHERE sp.schedule_id = $1 AND sp.status = 'active' AND u.deleted_at IS NULL
		ORDER BY u.external_id ASC
	`, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("GetClassRecap.students: %w", err)
	}
	defer studRows.Close()

	type studInfo struct {
		id   uuid.UUID
		nim  string
		name string
	}
	var enrolled []studInfo
	for studRows.Next() {
		var s studInfo
		if err := studRows.Scan(&s.id, &s.nim, &s.name); err == nil {
			enrolled = append(enrolled, s)
		}
	}

	// 4. Get all attendance records for this schedule's sessions
	attRows, err := r.db.Query(ctx, `
		SELECT a.student_id, cs.meeting_no, a.status
		FROM attendances a
		JOIN class_sessions cs ON a.session_id = cs.id
		WHERE cs.schedule_id = $1
	`, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("GetClassRecap.attendances: %w", err)
	}
	defer attRows.Close()

	studentAttMap := make(map[uuid.UUID]map[int]domain.AttendanceStatus)
	for attRows.Next() {
		var studID uuid.UUID
		var mNo int
		var st domain.AttendanceStatus
		if err := attRows.Scan(&studID, &mNo, &st); err == nil {
			if _, ok := studentAttMap[studID]; !ok {
				studentAttMap[studID] = make(map[int]domain.AttendanceStatus)
			}
			studentAttMap[studID][mNo] = st
		}
	}

	// 5. Build recap rows
	totalSessions := len(meetings)
	var studentRecaps []*domain.StudentAttendanceRecap
	for _, s := range enrolled {
		sr := &domain.StudentAttendanceRecap{
			StudentID:       s.id,
			StudentNIM:      s.nim,
			StudentName:     s.name,
			MeetingStatuses: make(map[int]domain.AttendanceStatus),
			TotalMeetings:   totalSessions,
		}

		stMap := studentAttMap[s.id]
		for _, m := range meetings {
			if status, ok := stMap[m]; ok {
				sr.MeetingStatuses[m] = status
				switch status {
				case domain.StatusHadir:
					sr.HadirCount++
					sr.TotalAttended++
				case domain.StatusTerlambat:
					sr.TerlambatCount++
					sr.TotalAttended++
				case domain.StatusIzin:
					sr.IzinCount++
				case domain.StatusSakit:
					sr.SakitCount++
				case domain.StatusAlpa:
					sr.AlpaCount++
				}
			} else {
				sr.MeetingStatuses[m] = domain.StatusAlpa
				sr.AlpaCount++
			}
		}

		if totalSessions > 0 {
			pct := float64(sr.TotalAttended) / float64(totalSessions) * 100.0
			sr.Percentage = math.Round(pct*10) / 10
		} else {
			sr.Percentage = 100.0
		}
		studentRecaps = append(studentRecaps, sr)
	}

	return &domain.ClassAttendanceRecap{
		ScheduleID:    scheduleID,
		CourseName:    courseName,
		TotalSessions: totalSessions,
		Meetings:      meetings,
		Students:      studentRecaps,
	}, nil
}

func scanAttendance(row pgx.Row) (*domain.Attendance, error) {
	a := &domain.Attendance{}
	err := row.Scan(
		&a.ID, &a.SessionID, &a.StudentID, &a.Status, &a.ScannedAt, &a.DeviceID,
		&a.Latitude, &a.Longitude, &a.DistanceMeters, &a.AttachmentURL, &a.Notes,
		&a.SubmissionSource, &a.UpdatedBy, &a.VerifiedByLecturer, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanAttendance: %w", err)
	}
	return a, nil
}

func scanAttendanceRow(rows pgx.Rows) (*domain.Attendance, error) {
	a := &domain.Attendance{}
	err := rows.Scan(
		&a.ID, &a.SessionID, &a.StudentID, &a.Status, &a.ScannedAt, &a.DeviceID,
		&a.Latitude, &a.Longitude, &a.DistanceMeters, &a.AttachmentURL, &a.Notes,
		&a.SubmissionSource, &a.UpdatedBy, &a.VerifiedByLecturer, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanAttendanceRow: %w", err)
	}
	return a, nil
}
