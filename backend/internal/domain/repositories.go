package domain

import (
	"context"

	"github.com/google/uuid"
)

// ─────────────────────────── REPOSITORY INTERFACES ───────────────────────────

// FacultyRepository defines persistence operations for Faculty.
type FacultyRepository interface {
	Upsert(ctx context.Context, f *Faculty) error
	FindByCode(ctx context.Context, code string) (*Faculty, error)
	FindAll(ctx context.Context) ([]*Faculty, error)
}

// StudyProgramRepository defines persistence operations for StudyProgram.
type StudyProgramRepository interface {
	Upsert(ctx context.Context, sp *StudyProgram) error
	FindByCode(ctx context.Context, code string) (*StudyProgram, error)
	FindByFacultyID(ctx context.Context, facultyID uuid.UUID) ([]*StudyProgram, error)
}

// BuildingRepository defines persistence operations for Building.
type BuildingRepository interface {
	Upsert(ctx context.Context, b *Building) error
	FindByCode(ctx context.Context, code string) (*Building, error)
}

// RoomRepository defines persistence operations for Room.
type RoomRepository interface {
	Upsert(ctx context.Context, r *Room) error
	FindByCode(ctx context.Context, roomCode string) (*Room, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Room, error)
}

// UserRepository defines persistence operations for User.
type UserRepository interface {
	Upsert(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByExternalID(ctx context.Context, externalID string) (*User, error)
	UpdateDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error
	GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
}

// ClassScheduleRepository defines persistence operations for ClassSchedule.
type ClassScheduleRepository interface {
	Upsert(ctx context.Context, cs *ClassSchedule) error
	FindByID(ctx context.Context, id uuid.UUID) (*ClassSchedule, error)
	FindByExternalID(ctx context.Context, externalID string) (*ClassSchedule, error)
	// FindTodayByLecturer returns schedules scheduled for today for a given lecturer.
	FindTodayByLecturer(ctx context.Context, lecturerID uuid.UUID, dayOfWeek int) ([]*ClassSchedule, error)
	// FindTodayByStudent returns schedules for today for a given student via study_plans.
	FindTodayByStudent(ctx context.Context, studentID uuid.UUID, dayOfWeek int) ([]*ClassSchedule, error)
	// FindByLecturerWithDetails returns all schedules taught by lecturer with room, building, student count.
	FindByLecturerWithDetails(ctx context.Context, lecturerID uuid.UUID) ([]*ClassScheduleDetail, error)
}

// StudyPlanRepository defines persistence operations for StudyPlan (KRS).
type StudyPlanRepository interface {
	Upsert(ctx context.Context, sp *StudyPlan) error
	IsEnrolled(ctx context.Context, studentID, scheduleID uuid.UUID) (bool, error)
	FindByStudentID(ctx context.Context, studentID uuid.UUID) ([]*StudyPlan, error)
	FindEnrolledStudentsByScheduleID(ctx context.Context, scheduleID uuid.UUID) ([]*EnrolledStudent, error)
}

// ClassSessionRepository defines persistence operations for ClassSession.
type ClassSessionRepository interface {
	Create(ctx context.Context, cs *ClassSession) error
	FindByID(ctx context.Context, id uuid.UUID) (*ClassSession, error)
	FindByScheduleID(ctx context.Context, scheduleID uuid.UUID) ([]*ClassSession, error)
	FindActiveByScheduleID(ctx context.Context, scheduleID uuid.UUID) (*ClassSession, error)
	Close(ctx context.Context, sessionID uuid.UUID, bapTopic *string) error
}

// AttendanceRepository defines persistence operations for Attendance.
type AttendanceRepository interface {
	Create(ctx context.Context, a *Attendance) error
	FindBySessionAndStudent(ctx context.Context, sessionID, studentID uuid.UUID) (*Attendance, error)
	FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*Attendance, error)
	GetAttendeesBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*AttendeeDetail, error)
	Update(ctx context.Context, a *Attendance) error
	// GetSummaryByStudentAndSchedule returns attendance count per status for a student in a schedule.
	GetSummaryByStudentAndSchedule(ctx context.Context, studentID, scheduleID uuid.UUID) (map[AttendanceStatus]int, error)
	// GetClassRecap computes complete matrix attendance recap for a schedule.
	GetClassRecap(ctx context.Context, scheduleID uuid.UUID) (*ClassAttendanceRecap, error)
	// MarkPermission marks a student as izin or sakit for a session (with audit trail: manual_lecturer).
	MarkPermission(ctx context.Context, sessionID, studentID, lecturerID uuid.UUID, status AttendanceStatus, notes string) error
	// GetPermissionsByLecturer retrieves permission submissions for all classes taught by a lecturer.
	GetPermissionsByLecturer(ctx context.Context, lecturerID uuid.UUID, pendingOnly bool) ([]*PermissionApprovalItem, error)
	// ApprovePermission approves or rejects a student's permission request.
	ApprovePermission(ctx context.Context, attendanceID, lecturerID uuid.UUID, approved bool, notes string) error
	// SubmitPermissionRequest allows student to submit permission request with attachment.
	SubmitPermissionRequest(ctx context.Context, sessionID, studentID uuid.UUID, status AttendanceStatus, notes string, attachmentURL *string) (*Attendance, error)
	// AutoMarkAlpa marks all absent enrolled students who didn't scan or get permission as alpa upon session close.
	AutoMarkAlpa(ctx context.Context, sessionID, scheduleID uuid.UUID) error
}

// ProdiMonitoringRepository defines queries for Program Study monitoring.
type ProdiMonitoringRepository interface {
	GetProdiOverview(ctx context.Context, prodiID uuid.UUID) (*ProdiOverview, error)
	GetLecturersCompliance(ctx context.Context, prodiID uuid.UUID) ([]*LecturerCompliance, error)
	GetStudentsAtRisk(ctx context.Context, prodiID uuid.UUID) ([]*StudentAtRisk, error)
	GetProdiClasses(ctx context.Context, prodiID uuid.UUID) ([]*ProdiClassSummary, error)
	GetLiveTodayClasses(ctx context.Context, prodiID uuid.UUID) ([]*ProdiLiveClassItem, error)
}

// SystemAdminRepository defines queries for Superadmin / System Admin.
type SystemAdminRepository interface {
	GetSystemStats(ctx context.Context) (*SystemStats, error)
	GetAllStudyPrograms(ctx context.Context) ([]*StudyProgramDetail, error)
	GetUsers(ctx context.Context, role *UserRole, prodiID *uuid.UUID) ([]*SystemUserItem, error)
	GetThemeSettings(ctx context.Context) (*SystemThemeSettings, error)
	SaveThemeSettings(ctx context.Context, s *SystemThemeSettings) error
	GetCampusApiConfig(ctx context.Context) (*CampusApiConfig, error)
	SaveCampusApiConfig(ctx context.Context, cfg *CampusApiConfig) error
	GetAllCampusLocations(ctx context.Context) ([]*CampusLocation, error)
	GetActiveCampusLocations(ctx context.Context) ([]*CampusLocation, error)
	CreateCampusLocation(ctx context.Context, loc *CampusLocation) error
	UpdateCampusLocation(ctx context.Context, loc *CampusLocation) error
	DeleteCampusLocation(ctx context.Context, id uuid.UUID) error
}

// ─────────────────────────── CACHE INTERFACES ───────────────────────────

// QRTokenCache defines the Redis interface for rolling QR token management.
type QRTokenCache interface {
	// SetSeed stores the TOTP seed for a session.
	SetSeed(ctx context.Context, sessionID uuid.UUID, seed string) error
	// GetSeed retrieves the TOTP seed for a session.
	GetSeed(ctx context.Context, sessionID uuid.UUID) (string, error)
	// DeleteSeed removes the TOTP seed when a session is closed.
	DeleteSeed(ctx context.Context, sessionID uuid.UUID) error
	// ValidateToken checks if the given TOTP token is valid for the session.
	ValidateToken(ctx context.Context, sessionID uuid.UUID, token string) (bool, error)
	// GenerateCurrentToken generates the current TOTP token and remaining seconds for a session.
	GenerateCurrentToken(ctx context.Context, sessionID uuid.UUID) (string, int64, error)
}

// SessionCache defines the Redis interface for active session tracking.
type SessionCache interface {
	// SetActiveSession marks a schedule as having an open session.
	SetActiveSession(ctx context.Context, scheduleID, sessionID uuid.UUID) error
	// GetActiveSession retrieves the active session ID for a schedule.
	GetActiveSession(ctx context.Context, scheduleID uuid.UUID) (uuid.UUID, error)
	// DeleteActiveSession removes the active session marker when closed.
	DeleteActiveSession(ctx context.Context, scheduleID uuid.UUID) error
}
