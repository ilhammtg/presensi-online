package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/ilham/presensi-online/backend/internal/domain"
)

// SessionBroadcaster defines WebSocket broadcasting capabilities needed by SessionUseCase.
type SessionBroadcaster interface {
	WSPublisher
	BroadcastQRToken(roomID, sessionID uuid.UUID, token string, expiresIn int64)
	BroadcastSessionOpened(roomID, sessionID uuid.UUID, meetingNo int)
	BroadcastSessionClosed(roomID, sessionID uuid.UUID)
}

// SessionUseCase handles class session management (open/close by lecturer).
type SessionUseCase struct {
	sessionRepo    domain.ClassSessionRepository
	scheduleRepo   domain.ClassScheduleRepository
	studyPlanRepo  domain.StudyPlanRepository
	attendanceRepo domain.AttendanceRepository
	qrCache        domain.QRTokenCache
	sessionCache   domain.SessionCache
	broadcaster    SessionBroadcaster
	periodSeconds  int

	mu        sync.Mutex
	stopChans map[uuid.UUID]chan struct{}
}

// NewSessionUseCase creates a new SessionUseCase.
func NewSessionUseCase(
	sessionRepo domain.ClassSessionRepository,
	scheduleRepo domain.ClassScheduleRepository,
	studyPlanRepo domain.StudyPlanRepository,
	attendanceRepo domain.AttendanceRepository,
	qrCache domain.QRTokenCache,
	sessionCache domain.SessionCache,
	broadcaster SessionBroadcaster,
	periodSeconds int,
) *SessionUseCase {
	if periodSeconds <= 0 {
		periodSeconds = 15
	}
	return &SessionUseCase{
		sessionRepo:    sessionRepo,
		scheduleRepo:   scheduleRepo,
		studyPlanRepo:  studyPlanRepo,
		attendanceRepo: attendanceRepo,
		qrCache:        qrCache,
		sessionCache:   sessionCache,
		broadcaster:    broadcaster,
		periodSeconds:  periodSeconds,
		stopChans:      make(map[uuid.UUID]chan struct{}),
	}
}

// OpenSessionRequest contains the data needed to open a new attendance session.
type OpenSessionRequest struct {
	ScheduleID      uuid.UUID
	LecturerID      uuid.UUID
	MeetingNo       int
	DurationMinutes int
}

// OpenSession opens a new attendance session for a class, generates a TOTP seed,
// caches it in Redis, and begins broadcasting rolling QR tokens via WebSocket.
func (uc *SessionUseCase) OpenSession(ctx context.Context, req OpenSessionRequest) (*domain.ClassSession, error) {
	// Verify schedule exists and belongs to this lecturer
	schedule, err := uc.scheduleRepo.FindByID(ctx, req.ScheduleID)
	if err != nil {
		return nil, err
	}
	if schedule.LecturerID != req.LecturerID {
		return nil, domain.ErrForbidden
	}

	// Verify schedule is scheduled for today
	now := time.Now()
	todayDayOfWeek := int(now.Weekday())
	if todayDayOfWeek == 0 {
		todayDayOfWeek = 7
	}
	if schedule.DayOfWeek != todayDayOfWeek {
		return nil, domain.ErrScheduleNotToday
	}

	// Rule: Valid meeting range 1-16
	if req.MeetingNo < 1 || req.MeetingNo > 16 {
		return nil, &domain.DomainError{
			Code:    400,
			Message: "nomor pertemuan harus berada di antara pertemuan 1 sampai 16",
		}
	}

	// Check existing sessions for this schedule
	allSessions, err := uc.sessionRepo.FindByScheduleID(ctx, req.ScheduleID)
	if err == nil {
		for _, s := range allSessions {
			// Rule 1: No duplicate meeting number
			if s.MeetingNo == req.MeetingNo {
				return nil, &domain.DomainError{
					Code:    409,
					Message: fmt.Sprintf("pertemuan ke-%d sudah pernah dibuat untuk mata kuliah ini", req.MeetingNo),
				}
			}
			// Rule 2: Academic calendar constraint - only 1 meeting per schedule per day (cannot open future meetings on same day)
			if s.SessionDate.Year() == now.Year() && s.SessionDate.YearDay() == now.YearDay() {
				return nil, &domain.DomainError{
					Code:    400,
					Message: fmt.Sprintf("pertemuan perkuliahan untuk hari ini (Pertemuan %d) sudah pernah diselenggarakan. Sesuai kalender akademik, pertemuan selanjutnya hanya dapat dibuka pada jadwal minggu depan.", s.MeetingNo),
				}
			}
		}
	}

	// Ensure no active session already exists for this schedule
	existing, err := uc.sessionRepo.FindActiveByScheduleID(ctx, req.ScheduleID)
	if err == nil && existing != nil {
		return nil, &domain.DomainError{Code: 409, Message: "sesi presensi sudah aktif untuk jadwal ini"}
	}

	// Generate cryptographically secure TOTP seed (64 hex chars = 32 random bytes)
	seed, err := generateQRSeed()
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR seed: %w", err)
	}

	dur := req.DurationMinutes
	if dur <= 0 {
		dur = 30
	}
	expiresAt := now.Add(time.Duration(dur) * time.Minute)

	session := &domain.ClassSession{
		ID:              uuid.New(),
		ScheduleID:      req.ScheduleID,
		MeetingNo:       req.MeetingNo,
		SessionDate:     now,
		QRSeed:          seed,
		IsOpen:          true,
		OpenedAt:        now,
		DurationMinutes: dur,
		ExpiresAt:       &expiresAt,
		CreatedAt:       now,
	}

	// Persist session to database
	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Cache seed in Redis for fast TOTP validation
	if err := uc.qrCache.SetSeed(ctx, session.ID, seed); err != nil {
		return nil, fmt.Errorf("failed to cache QR seed: %w", err)
	}

	// Cache active session mapping: scheduleID → sessionID
	if err := uc.sessionCache.SetActiveSession(ctx, req.ScheduleID, session.ID); err != nil {
		fmt.Printf("warning: failed to cache active session: %v\n", err)
	}

	// Broadcast initial session opened & initial QR token
	if uc.broadcaster != nil {
		uc.broadcaster.BroadcastSessionOpened(req.ScheduleID, session.ID, req.MeetingNo)
		token, rem, err := uc.qrCache.GenerateCurrentToken(ctx, session.ID)
		if err == nil {
			uc.broadcaster.BroadcastQRToken(session.ID, session.ID, token, rem)
			uc.broadcaster.BroadcastQRToken(req.ScheduleID, session.ID, token, rem)
		}
	}

	// BR-SEC-02: Dynamic Rolling QR Code via WebSocket (every 10-15 seconds)
	uc.startRollingQR(session.ID, req.ScheduleID)

	return session, nil
}

func (uc *SessionUseCase) startRollingQR(sessionID, scheduleID uuid.UUID) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	// Stop any existing goroutine for this session
	if ch, ok := uc.stopChans[sessionID]; ok {
		close(ch)
	}

	stopChan := make(chan struct{})
	uc.stopChans[sessionID] = stopChan

	go func(sessID, schedID uuid.UUID, stop <-chan struct{}) {
		ticker := time.NewTicker(time.Duration(uc.periodSeconds) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				if uc.broadcaster == nil {
					continue
				}
				token, rem, err := uc.qrCache.GenerateCurrentToken(context.Background(), sessID)
				if err != nil {
					return
				}
				uc.broadcaster.BroadcastQRToken(sessID, sessID, token, rem)
				uc.broadcaster.BroadcastQRToken(schedID, sessID, token, rem)
			}
		}
	}(sessionID, scheduleID, stopChan)
}

// CloseSessionRequest contains the data needed to close an attendance session.
type CloseSessionRequest struct {
	SessionID  uuid.UUID
	LecturerID uuid.UUID
	BAPTopic   *string
}

// CloseSession closes an open session, saves the BAP topic, and cleans up Redis cache.
func (uc *SessionUseCase) CloseSession(ctx context.Context, req CloseSessionRequest) error {
	session, err := uc.sessionRepo.FindByID(ctx, req.SessionID)
	if err != nil {
		return err
	}

	// Verify ownership via schedule
	schedule, err := uc.scheduleRepo.FindByID(ctx, session.ScheduleID)
	if err != nil {
		return err
	}
	if schedule.LecturerID != req.LecturerID {
		return domain.ErrForbidden
	}

	// BR-DSN-03 (Validasi BAP): Sesi perkuliahan tidak dapat ditutup secara permanen
	// jika dosen belum mengisi ringkasan materi/BAP (bap_topic) minimal 10 karakter.
	if req.BAPTopic == nil || len(strings.TrimSpace(*req.BAPTopic)) < 10 {
		return &domain.DomainError{
			Code:    400,
			Message: "Ringkasan materi perkuliahan (BAP) wajib diisi minimal 10 karakter sebelum sesi dapat ditutup.",
		}
	}

	// Stop rolling QR ticker
	uc.mu.Lock()
	if ch, ok := uc.stopChans[req.SessionID]; ok {
		close(ch)
		delete(uc.stopChans, req.SessionID)
	}
	uc.mu.Unlock()

	// Close the session in DB
	if err := uc.sessionRepo.Close(ctx, req.SessionID, req.BAPTopic); err != nil {
		return fmt.Errorf("failed to close session: %w", err)
	}

	// Automatically mark all enrolled students who didn't attend or get permission as alpa
	_ = uc.attendanceRepo.AutoMarkAlpa(ctx, req.SessionID, session.ScheduleID)

	// Clean up Redis cache
	_ = uc.qrCache.DeleteSeed(ctx, req.SessionID)
	_ = uc.sessionCache.DeleteActiveSession(ctx, session.ScheduleID)

	// Broadcast session closed
	if uc.broadcaster != nil {
		uc.broadcaster.BroadcastSessionClosed(req.SessionID, req.SessionID)
		uc.broadcaster.BroadcastSessionClosed(session.ScheduleID, req.SessionID)
	}

	return nil
}

// GetDosenSchedules returns all schedules taught by a lecturer with details.
func (uc *SessionUseCase) GetDosenSchedules(ctx context.Context, lecturerID uuid.UUID) ([]*domain.ClassScheduleDetail, error) {
	return uc.scheduleRepo.FindByLecturerWithDetails(ctx, lecturerID)
}

// GetMahasiswaSchedules returns all enrolled schedules for a mahasiswa (student) with details.
func (uc *SessionUseCase) GetMahasiswaSchedules(ctx context.Context, studentID uuid.UUID) ([]*domain.ClassScheduleDetail, error) {
	return uc.scheduleRepo.FindByStudentWithDetails(ctx, studentID)
}

// GetEnrolledStudents returns all students enrolled in a class schedule.
func (uc *SessionUseCase) GetEnrolledStudents(ctx context.Context, scheduleID uuid.UUID) ([]*domain.EnrolledStudent, error) {
	return uc.studyPlanRepo.FindEnrolledStudentsByScheduleID(ctx, scheduleID)
}

// GetScheduleSessions returns all sessions for a schedule and next recommended meeting number.
func (uc *SessionUseCase) GetScheduleSessions(ctx context.Context, scheduleID uuid.UUID) ([]*domain.MeetingSessionItem, int, error) {
	sessions, err := uc.sessionRepo.FindByScheduleID(ctx, scheduleID)
	if err != nil {
		return nil, 1, err
	}

	maxMeetingNo := 0
	var items []*domain.MeetingSessionItem
	for _, s := range sessions {
		if s.MeetingNo > maxMeetingNo {
			maxMeetingNo = s.MeetingNo
		}
		attendees, _ := uc.attendanceRepo.GetAttendeesBySessionID(ctx, s.ID)
		items = append(items, &domain.MeetingSessionItem{
			SessionID:      s.ID,
			MeetingNo:      s.MeetingNo,
			SessionDate:    s.SessionDate,
			IsOpen:         s.IsOpen,
			OpenedAt:       s.OpenedAt,
			ClosedAt:       s.ClosedAt,
			BAPTopic:       s.BAPTopic,
			AttendeesCount: len(attendees),
		})
	}
	nextMeeting := maxMeetingNo + 1
	if nextMeeting > 16 {
		nextMeeting = 16
	}

	return items, nextMeeting, nil
}

// GetClassRecap returns full attendance recap for a schedule.
func (uc *SessionUseCase) GetClassRecap(ctx context.Context, scheduleID uuid.UUID) (*domain.ClassAttendanceRecap, error) {
	return uc.attendanceRepo.GetClassRecap(ctx, scheduleID)
}

// MarkPermission marks a student as izin or sakit during an active attendance session with audit trail.
func (uc *SessionUseCase) MarkPermission(ctx context.Context, sessionID, studentID, lecturerID uuid.UUID, status domain.AttendanceStatus, notes string) error {
	if status != domain.StatusIzin && status != domain.StatusSakit && status != domain.StatusAlpa && status != domain.StatusHadir {
		return fmt.Errorf("invalid attendance status: %s", status)
	}

	session, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("sesi perkuliahan tidak ditemukan: %w", err)
	}
	if !session.IsOpen {
		return &domain.DomainError{Code: 400, Message: "izin hanya dapat diberikan saat proses absensi/sesi perkuliahan sedang berlangsung"}
	}

	if err := uc.attendanceRepo.MarkPermission(ctx, sessionID, studentID, lecturerID, status, notes); err != nil {
		return err
	}

	if uc.broadcaster != nil {
		notesCopy := notes
		now := time.Now()
		_ = uc.broadcaster.PublishAttendanceRecorded(session.ID, &domain.Attendance{
			SessionID: session.ID,
			StudentID: studentID,
			Status:    status,
			Notes:     &notesCopy,
			ScannedAt: &now,
		})
	}

	return nil
}

// GetPermissionsByLecturer retrieves permission submissions for a lecturer's classes.
func (uc *SessionUseCase) GetPermissionsByLecturer(ctx context.Context, lecturerID uuid.UUID, pendingOnly bool) ([]*domain.PermissionApprovalItem, error) {
	return uc.attendanceRepo.GetPermissionsByLecturer(ctx, lecturerID, pendingOnly)
}

// ApprovePermission approves or rejects a student's permission request.
func (uc *SessionUseCase) ApprovePermission(ctx context.Context, attendanceID, lecturerID uuid.UUID, approved bool, notes string) error {
	return uc.attendanceRepo.ApprovePermission(ctx, attendanceID, lecturerID, approved, notes)
}

// SubmitStudentPermission handles permission application submitted by student with doctor/dispensation attachment.
func (uc *SessionUseCase) SubmitStudentPermission(ctx context.Context, sessionID, studentID uuid.UUID, status domain.AttendanceStatus, notes string, attachmentURL *string) (*domain.Attendance, error) {
	if status != domain.StatusIzin && status != domain.StatusSakit {
		return nil, &domain.DomainError{Code: 400, Message: "status permohonan harus izin atau sakit"}
	}
	return uc.attendanceRepo.SubmitPermissionRequest(ctx, sessionID, studentID, status, notes, attachmentURL)
}

// GetSessionDetails retrieves session details.
func (uc *SessionUseCase) GetSessionDetails(ctx context.Context, sessionID uuid.UUID) (*domain.ClassSession, error) {
	return uc.sessionRepo.FindByID(ctx, sessionID)
}

// GetAttendees retrieves the enriched list of attendees for a session.
func (uc *SessionUseCase) GetAttendees(ctx context.Context, sessionID uuid.UUID) ([]*domain.AttendeeDetail, error) {
	return uc.attendanceRepo.GetAttendeesBySessionID(ctx, sessionID)
}

// GetCurrentQR generates the current rolling TOTP token on-demand for a session.
func (uc *SessionUseCase) GetCurrentQR(ctx context.Context, sessionID uuid.UUID) (string, int64, error) {
	session, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return "", 0, err
	}
	if !session.IsOpen {
		return "", 0, domain.ErrSessionClosed
	}

	// Ensure rolling QR ticker is running for open session
	uc.mu.Lock()
	if _, ok := uc.stopChans[sessionID]; !ok {
		uc.mu.Unlock()
		uc.startRollingQR(session.ID, session.ScheduleID)
	} else {
		uc.mu.Unlock()
	}

	return uc.qrCache.GenerateCurrentToken(ctx, sessionID)
}

// GetActiveSessionBySchedule retrieves the currently open session for a schedule, if any.
func (uc *SessionUseCase) GetActiveSessionBySchedule(ctx context.Context, scheduleID uuid.UUID) (*domain.ClassSession, string, int64, error) {
	session, err := uc.sessionRepo.FindActiveByScheduleID(ctx, scheduleID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, "", 0, nil
		}
		return nil, "", 0, err
	}
	if session == nil {
		return nil, "", 0, nil
	}

	token, remaining, _ := uc.qrCache.GenerateCurrentToken(ctx, session.ID)
	return session, token, remaining, nil
}

// generateQRSeed creates a cryptographically secure 32-byte random seed encoded as hex.
func generateQRSeed() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
