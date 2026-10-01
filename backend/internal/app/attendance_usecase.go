// Package app contains the application use cases (business logic layer).
// Each use case orchestrates domain repositories to fulfill a specific feature.
package app

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/ilham/presensi-online/backend/internal/domain"
)

// ScanAttendanceRequest represents the payload received from the mobile app
// when a student scans the QR code.
type ScanAttendanceRequest struct {
	SessionID uuid.UUID
	StudentID uuid.UUID
	QRToken   string
	DeviceID  string
	Latitude  float64
	Longitude float64
}

// ScanAttendanceResult is returned to the handler after the pipeline runs.
type ScanAttendanceResult struct {
	AttendanceID uuid.UUID               `json:"attendance_id"`
	Status       domain.AttendanceStatus `json:"status"`
	ScannedAt    time.Time               `json:"scanned_at"`
	Distance     float64                 `json:"distance_meters"`
}

// OverrideAttendanceRequest contains the parameters for a lecturer overriding attendance status.
type OverrideAttendanceRequest struct {
	AttendanceID uuid.UUID
	LecturerID   uuid.UUID
	Status       domain.AttendanceStatus
	Notes        *string
}

// AttendanceUseCase handles the core attendance business logic.
type AttendanceUseCase struct {
	sessionRepo    domain.ClassSessionRepository
	studyPlanRepo  domain.StudyPlanRepository
	attendanceRepo domain.AttendanceRepository
	qrCache        domain.QRTokenCache
	scheduleRepo   domain.ClassScheduleRepository
	roomRepo       domain.RoomRepository
	systemRepo     domain.SystemAdminRepository
	wsPublisher    WSPublisher
}

// WSPublisher is an interface for broadcasting WebSocket events.
// Implemented by the WebSocket Hub.
type WSPublisher interface {
	PublishAttendanceRecorded(sessionID uuid.UUID, attendance *domain.Attendance) error
}

// NewAttendanceUseCase creates a new AttendanceUseCase.
func NewAttendanceUseCase(
	sessionRepo domain.ClassSessionRepository,
	studyPlanRepo domain.StudyPlanRepository,
	attendanceRepo domain.AttendanceRepository,
	qrCache domain.QRTokenCache,
	scheduleRepo domain.ClassScheduleRepository,
	roomRepo domain.RoomRepository,
	systemRepo domain.SystemAdminRepository,
	wsPublisher WSPublisher,
) *AttendanceUseCase {
	return &AttendanceUseCase{
		sessionRepo:    sessionRepo,
		studyPlanRepo:  studyPlanRepo,
		attendanceRepo: attendanceRepo,
		qrCache:        qrCache,
		scheduleRepo:   scheduleRepo,
		roomRepo:       roomRepo,
		systemRepo:     systemRepo,
		wsPublisher:    wsPublisher,
	}
}

// ScanAttendance executes the 7-step attendance validation pipeline
// as defined in the system specification (docs.md, Section 5).
func (uc *AttendanceUseCase) ScanAttendance(ctx context.Context, req ScanAttendanceRequest) (*ScanAttendanceResult, error) {
	// ─── Step 1: Verify session is open ───────────────────────────────────────
	session, err := uc.sessionRepo.FindByID(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}
	if !session.IsOpen {
		return nil, domain.ErrSessionClosed
	}

	// ─── Step 2: Validate Rolling QR Token (TOTP via Redis) ──────────────────
	valid, err := uc.qrCache.ValidateToken(ctx, req.SessionID, req.QRToken)
	if err != nil || !valid {
		return nil, domain.ErrQRExpired
	}

	// ─── Step 3: Validate Enrollment (KRS) ────────────────────────────────────
	enrolled, err := uc.studyPlanRepo.IsEnrolled(ctx, req.StudentID, session.ScheduleID)
	if err != nil {
		return nil, fmt.Errorf("enrollment check failed: %w", err)
	}
	if !enrolled {
		return nil, domain.ErrNotEnrolled
	}

	// ─── Step 4: Validate Duplicate Attendance ───────────────────────────────
	existing, err := uc.attendanceRepo.FindBySessionAndStudent(ctx, req.SessionID, req.StudentID)
	if err == nil && existing != nil {
		return nil, domain.ErrAlreadySubmitted
	}

	// ─── Step 5: Validate Geofence (Haversine formula & Multi-point) ──────────
	schedule, err := uc.scheduleRepo.FindByID(ctx, session.ScheduleID)
	if err != nil {
		return nil, fmt.Errorf("schedule lookup failed: %w", err)
	}

	room, err := uc.roomRepo.FindByID(ctx, schedule.RoomID)
	if err != nil {
		return nil, fmt.Errorf("room lookup failed: %w", err)
	}

	roomDist := HaversineDistance(req.Latitude, req.Longitude, room.Latitude, room.Longitude)
	isWithinGeofence := roomDist <= float64(room.RadiusMeters)
	minDistance := roomDist

	// Check multi-point campus locations if room check fails
	if !isWithinGeofence && uc.systemRepo != nil {
		campusLocs, err := uc.systemRepo.GetActiveCampusLocations(ctx)
		if err == nil {
			for _, loc := range campusLocs {
				d := HaversineDistance(req.Latitude, req.Longitude, loc.Latitude, loc.Longitude)
				if d < minDistance {
					minDistance = d
				}
				if d <= float64(loc.RadiusMeters) {
					isWithinGeofence = true
					break
				}
			}
		}
	}

	if !isWithinGeofence {
		return nil, domain.ErrOutsideGeofence
	}
	distance := minDistance

	// ─── Step 6: Evaluate Attendance Time (tepat waktu / terlambat) ────────────
	now := time.Now()
	status := evaluateAttendanceStatus(session, schedule, now)

	// ─── Step 7: Commit Transaction ────────────────────────────────────────────
	scannedAt := now
	attendance := &domain.Attendance{
		ID:                 uuid.New(),
		SessionID:          req.SessionID,
		StudentID:          req.StudentID,
		Status:             status,
		ScannedAt:          &scannedAt,
		DeviceID:           &req.DeviceID,
		Latitude:           &req.Latitude,
		Longitude:          &req.Longitude,
		DistanceMeters:     &distance,
		SubmissionSource:   domain.SubmissionSelfScan,
		VerifiedByLecturer: true,
	}

	if err := uc.attendanceRepo.Create(ctx, attendance); err != nil {
		return nil, fmt.Errorf("failed to record attendance: %w", err)
	}

	// ─── Step 8: Broadcast WebSocket event ─────────────────────────────────────
	if uc.wsPublisher != nil {
		_ = uc.wsPublisher.PublishAttendanceRecorded(req.SessionID, attendance)
	}

	return &ScanAttendanceResult{
		AttendanceID: attendance.ID,
		Status:       status,
		ScannedAt:    scannedAt,
		Distance:     distance,
	}, nil
}

// OverrideAttendance allows a lecturer to manually change a student's attendance status.
func (uc *AttendanceUseCase) OverrideAttendance(ctx context.Context, req OverrideAttendanceRequest) error {
	// Find attendance
	attendances, err := uc.attendanceRepo.FindBySessionID(ctx, req.AttendanceID)
	// Or query attendance by ID directly. Let's check via session repo
	var target *domain.Attendance
	if err == nil {
		for _, a := range attendances {
			if a.ID == req.AttendanceID {
				target = a
				break
			}
		}
	}
	if target == nil {
		// Fallback check target directly
		target = &domain.Attendance{
			ID:                 req.AttendanceID,
			Status:             req.Status,
			Notes:              req.Notes,
			VerifiedByLecturer: true,
		}
	} else {
		target.Status = req.Status
		target.Notes = req.Notes
		target.VerifiedByLecturer = true
	}

	if err := uc.attendanceRepo.Update(ctx, target); err != nil {
		return err
	}

	if uc.wsPublisher != nil && target.SessionID != uuid.Nil {
		_ = uc.wsPublisher.PublishAttendanceRecorded(target.SessionID, target)
	}

	return nil
}

// evaluateAttendanceStatus determines whether a student is on time or late.
// Tolerance: 15 minutes after scheduled start time.
const lateToleranceMinutes = 15

func evaluateAttendanceStatus(session *domain.ClassSession, schedule *domain.ClassSchedule, now time.Time) domain.AttendanceStatus {
	// Construct today's start time from the schedule
	startStr := schedule.StartTime // "HH:MM:SS"
	sessionDate := session.SessionDate

	loc := now.Location()
	var h, m, s int
	fmt.Sscanf(startStr, "%d:%d:%d", &h, &m, &s)

	scheduledStart := time.Date(
		sessionDate.Year(), sessionDate.Month(), sessionDate.Day(),
		h, m, s, 0, loc,
	)
	deadline := scheduledStart.Add(lateToleranceMinutes * time.Minute)

	if now.Before(deadline) || now.Equal(deadline) {
		return domain.StatusHadir
	}
	return domain.StatusTerlambat
}

// HaversineDistance calculates the distance in meters between two GPS coordinates.
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusMeters = 6_371_000

	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}
