package domain

import (
	"time"

	"github.com/google/uuid"
)

// WIBLocation is Asia/Jakarta timezone (UTC+7) used across campus scheduling.
var WIBLocation = time.FixedZone("WIB", 7*3600)

// NowWIB returns current time in Asia/Jakarta timezone.
func NowWIB() time.Time {
	return time.Now().In(WIBLocation)
}

// CurrentDayOfWeek returns ISO-8601 day of week in WIB (1=Monday ... 7=Sunday).
func CurrentDayOfWeek() int {
	d := int(NowWIB().Weekday())
	if d == 0 {
		return 7
	}
	return d
}

// UserRole represents the role of a user in the system.
type UserRole string

const (
	RoleMahasiswa  UserRole = "mahasiswa"
	RoleDosen      UserRole = "dosen"
	RoleAdminProdi UserRole = "admin_prodi"
	RolePimpinan   UserRole = "pimpinan"
	RoleSuperadmin UserRole = "superadmin"
)

// AttendanceStatus represents the attendance status of a student.
type AttendanceStatus string

const (
	StatusHadir    AttendanceStatus = "hadir"
	StatusTerlambat AttendanceStatus = "terlambat"
	StatusIzin     AttendanceStatus = "izin"
	StatusSakit    AttendanceStatus = "sakit"
	StatusAlpa     AttendanceStatus = "alpa"
)

// EnrollmentStatus represents the KRS enrollment status of a student.
type EnrollmentStatus string

const (
	EnrollmentActive    EnrollmentStatus = "active"
	EnrollmentDropped   EnrollmentStatus = "dropped"
	EnrollmentWithdrawn EnrollmentStatus = "withdrawn"
)

// SubmissionChannel represents the channel through which attendance was submitted.
type SubmissionChannel string

const (
	SubmissionSelfScan       SubmissionChannel = "self_scan"
	SubmissionAppRequest     SubmissionChannel = "app_request"
	SubmissionManualLecturer SubmissionChannel = "manual_lecturer"
)

// ─────────────────────────── ENTITIES ───────────────────────────

// Faculty represents a university faculty (e.g., Fakultas Teknik).
type Faculty struct {
	ID        uuid.UUID
	Code      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// StudyProgram represents a study program (program studi) within a faculty.
type StudyProgram struct {
	ID        uuid.UUID
	FacultyID uuid.UUID
	Code      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Building represents a physical campus building.
type Building struct {
	ID        uuid.UUID
	Code      string
	Name      string
	CreatedAt time.Time
}

// Room represents a classroom or lab with geofence coordinates.
type Room struct {
	ID            uuid.UUID
	BuildingID    uuid.UUID
	RoomCode      string
	Name          string
	Latitude      float64
	Longitude     float64
	RadiusMeters  int
	IsActive      bool
	CreatedAt     time.Time
}

// User represents a system user (mahasiswa, dosen, admin, etc.).
type User struct {
	ID           uuid.UUID  `json:"id"`
	ExternalID   string     `json:"external_id"` // NIM or NIDN
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         UserRole   `json:"role"`
	ProdiID      *uuid.UUID `json:"prodi_id"`
	DeviceID     *string    `json:"device_id"` // Device binding — nil if not yet bound
	AvatarURL    *string    `json:"avatar_url"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// UserProfile represents enriched academic profile for a user.
type UserProfile struct {
	ID          uuid.UUID  `json:"id"`
	ExternalID  string     `json:"external_id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Role        UserRole   `json:"role"`
	ProdiID     *uuid.UUID `json:"prodi_id"`
	ProdiName   string     `json:"prodi_name"`
	FacultyName string     `json:"faculty_name"`
	AvatarURL   *string    `json:"avatar_url"`
	IsActive    bool       `json:"is_active"`
}

// ClassSchedule represents a recurring class schedule (master kelas).
type ClassSchedule struct {
	ID           uuid.UUID
	ExternalID   *string   // ID from campus database (SIAKAD)
	CourseCode   string
	CourseName   string
	AcademicYear string    // e.g., "2026/2027"
	SemesterType int       // 1=Ganjil, 2=Genap, 3=Pendek
	LecturerID   uuid.UUID
	RoomID       uuid.UUID
	DayOfWeek    int       // 1=Monday … 7=Sunday
	StartTime    string    // "HH:MM:SS"
	EndTime      string    // "HH:MM:SS"
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// StudyPlan represents a student's KRS enrollment in a class schedule.
type StudyPlan struct {
	ID         uuid.UUID
	StudentID  uuid.UUID
	ScheduleID uuid.UUID
	Status     EnrollmentStatus
	CreatedAt  time.Time
}

// ClassSession represents a single opened attendance session by a lecturer.
type ClassSession struct {
	ID              uuid.UUID  `json:"id"`
	ScheduleID      uuid.UUID  `json:"schedule_id"`
	MeetingNo       int        `json:"meeting_no"`
	SessionDate     time.Time  `json:"session_date"`
	QRSeed          string     `json:"qr_seed"`
	IsOpen          bool       `json:"is_open"`
	OpenedAt        time.Time  `json:"opened_at"`
	ClosedAt        *time.Time `json:"closed_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	DurationMinutes int        `json:"duration_minutes"`
	BAPTopic        *string    `json:"bap_topic"`
	CreatedAt       time.Time  `json:"created_at"`
}

// Attendance represents a single attendance transaction.
type Attendance struct {
	ID                  uuid.UUID
	SessionID           uuid.UUID
	StudentID           uuid.UUID
	Status              AttendanceStatus
	ScannedAt           *time.Time
	DeviceID            *string
	Latitude            *float64
	Longitude           *float64
	DistanceMeters      *float64
	AttachmentURL       *string // Bukti izin/sakit
	Notes               *string
	SubmissionSource    SubmissionChannel
	UpdatedBy           *uuid.UUID
	VerifiedByLecturer  bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// AttendeeDetail represents an attendance record enriched with student user info.
type AttendeeDetail struct {
	AttendanceID       uuid.UUID         `json:"attendance_id"`
	StudentID          uuid.UUID         `json:"student_id"`
	StudentNIM         string            `json:"student_nim"`
	StudentName        string            `json:"student_name"`
	Status             AttendanceStatus  `json:"status"`
	ScannedAt          *time.Time        `json:"scanned_at"`
	DistanceMeters     *float64          `json:"distance_meters"`
	Notes              *string           `json:"notes"`
	SubmissionSource   SubmissionChannel `json:"submission_source"`
	AttachmentURL      *string           `json:"attachment_url,omitempty"`
	UpdatedBy          *uuid.UUID        `json:"updated_by,omitempty"`
	VerifiedByLecturer bool              `json:"verified_by_lecturer"`
}

// PermissionApprovalItem represents a student permission request for the lecturer's Approval Center.
type PermissionApprovalItem struct {
	AttendanceID       uuid.UUID         `json:"attendance_id"`
	SessionID          uuid.UUID         `json:"session_id"`
	MeetingNo          int               `json:"meeting_no"`
	SessionDate        string            `json:"session_date"`
	ScheduleID         uuid.UUID         `json:"schedule_id"`
	CourseCode         string            `json:"course_code"`
	CourseName         string            `json:"course_name"`
	ClassUnit          string            `json:"class_unit"`
	StudentID          uuid.UUID         `json:"student_id"`
	StudentNIM         string            `json:"student_nim"`
	StudentName        string            `json:"student_name"`
	Status             AttendanceStatus  `json:"status"`
	SubmissionSource   SubmissionChannel `json:"submission_source"`
	AttachmentURL      *string           `json:"attachment_url,omitempty"`
	Notes              *string           `json:"notes,omitempty"`
	VerifiedByLecturer bool              `json:"verified_by_lecturer"`
	UpdatedBy          *uuid.UUID        `json:"updated_by,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
}

// ClassScheduleDetail represents a schedule enriched with room, building, student count, and active session status.
type ClassScheduleDetail struct {
	ID               uuid.UUID  `json:"id"`
	CourseCode       string     `json:"course_code"`
	CourseName       string     `json:"course_name"`
	ClassUnit        string     `json:"class_unit"`
	AcademicYear     string     `json:"academic_year"`
	SemesterType     int        `json:"semester_type"`
	LecturerID       uuid.UUID  `json:"lecturer_id"`
	LecturerName     string     `json:"lecturer_name"`
	RoomID           uuid.UUID  `json:"room_id"`
	RoomName         string     `json:"room_name"`
	BuildingName     string     `json:"building_name"`
	DayOfWeek        int        `json:"day_of_week"`
	DayName          string     `json:"day_name"`
	StartTime        string     `json:"start_time"`
	EndTime          string     `json:"end_time"`
	IsToday          bool       `json:"is_today"`
	EnrolledCount    int        `json:"enrolled_count"`
	HasActiveSession bool       `json:"has_active_session"`
	ActiveSessionID  *uuid.UUID `json:"active_session_id,omitempty"`
	NextMeetingNo    int        `json:"next_meeting_no"`
}

// EnrolledStudent represents a student taking this class schedule (KRS).
type EnrolledStudent struct {
	StudentID   uuid.UUID `json:"student_id"`
	StudentNIM  string    `json:"student_nim"`
	StudentName string    `json:"student_name"`
	Email       string    `json:"email"`
	Status      string    `json:"status"`
}

// MeetingSessionItem represents a conducted meeting session for a schedule.
type MeetingSessionItem struct {
	SessionID      uuid.UUID  `json:"session_id"`
	MeetingNo      int        `json:"meeting_no"`
	SessionDate    time.Time  `json:"session_date"`
	IsOpen         bool       `json:"is_open"`
	OpenedAt       time.Time  `json:"opened_at"`
	ClosedAt       *time.Time `json:"closed_at"`
	BAPTopic       *string    `json:"bap_topic"`
	AttendeesCount int        `json:"attendees_count"`
}

// StudentAttendanceRecap represents attendance breakdown for a student.
type StudentAttendanceRecap struct {
	StudentID       uuid.UUID                `json:"student_id"`
	StudentNIM      string                   `json:"student_nim"`
	StudentName     string                   `json:"student_name"`
	MeetingStatuses map[int]AttendanceStatus `json:"meeting_statuses"`
	HadirCount      int                      `json:"hadir_count"`
	TerlambatCount  int                      `json:"terlambat_count"`
	IzinCount       int                      `json:"izin_count"`
	SakitCount      int                      `json:"sakit_count"`
	AlpaCount       int                      `json:"alpa_count"`
	TotalAttended   int                      `json:"total_attended"`
	TotalMeetings   int                      `json:"total_meetings"`
	Percentage      float64                  `json:"percentage"`
}

// ClassAttendanceRecap represents overall class attendance recap.
type ClassAttendanceRecap struct {
	ScheduleID    uuid.UUID                 `json:"schedule_id"`
	CourseName    string                    `json:"course_name"`
	TotalSessions int                       `json:"total_sessions"`
	Meetings      []int                     `json:"meetings"`
	Students      []*StudentAttendanceRecap `json:"students"`
}

// ─────────────────────────── PRODI & SYSTEM ADMIN ENTITIES ───────────────────────────

// ProdiOverview represents aggregate statistics for a study program.
type ProdiOverview struct {
	ProdiID                uuid.UUID `json:"prodi_id"`
	ProdiName              string    `json:"prodi_name"`
	ProdiCode              string    `json:"prodi_code"`
	FacultyName            string    `json:"faculty_name"`
	TotalActiveStudents    int       `json:"total_active_students"`
	TotalLecturers         int       `json:"total_lecturers"`
	TotalCoursesOffered    int       `json:"total_courses_offered"`
	TotalSessionsHeld      int       `json:"total_sessions_held"`
	AvgAttendanceRate      float64   `json:"avg_attendance_rate"`
	TeachingComplianceRate float64   `json:"teaching_compliance_rate"`
	TodayClassesScheduled  int       `json:"today_classes_scheduled"`
	TodayClassesCompleted  int       `json:"today_classes_completed"`
}

// LecturerCompliance represents teaching adherence for a lecturer in the prodi.
type LecturerCompliance struct {
	LecturerID           uuid.UUID  `json:"lecturer_id"`
	LecturerName         string     `json:"lecturer_name"`
	NIDN                 string     `json:"nidn"`
	CourseCount          int        `json:"course_count"`
	CourseNames          []string   `json:"course_names"`
	TotalSessionsHeld    int        `json:"total_sessions_held"`
	TargetSessions       int        `json:"target_sessions"`
	CompliancePercentage float64    `json:"compliance_percentage"`
	Status               string     `json:"status"` // "lancar", "perlu_perhatian", "tertinggal"
	LastSessionDate      *time.Time `json:"last_session_date"`
}

// StudentAtRisk represents a student whose attendance rate is below 75%.
type StudentAtRisk struct {
	StudentID         uuid.UUID `json:"student_id"`
	StudentNIM        string    `json:"student_nim"`
	StudentName       string    `json:"student_name"`
	ProdiName         string    `json:"prodi_name"`
	CourseCode        string    `json:"course_code"`
	CourseName        string    `json:"course_name"`
	LecturerName      string    `json:"lecturer_name"`
	TotalSessionsHeld int       `json:"total_sessions_held"`
	HadirCount        int       `json:"hadir_count"`
	AlpaCount         int       `json:"alpa_count"`
	IzinSakitCount    int       `json:"izin_sakit_count"`
	AttendanceRate    float64   `json:"attendance_rate"`
	Recommendation    string    `json:"recommendation"`
}

// ProdiClassSummary represents summary of a class in the prodi.
type ProdiClassSummary struct {
	ScheduleID        uuid.UUID `json:"schedule_id"`
	CourseCode        string    `json:"course_code"`
	CourseName        string    `json:"course_name"`
	DayOfWeek         int       `json:"day_of_week"`
	DayName           string    `json:"day_name"`
	TimeRange         string    `json:"time_range"`
	RoomName          string    `json:"room_name"`
	LecturerName      string    `json:"lecturer_name"`
	LecturerNIDN      string    `json:"lecturer_nidn"`
	TotalEnrolled     int       `json:"total_enrolled"`
	CompletedSessions int       `json:"completed_sessions"`
	AvgAttendanceRate float64   `json:"avg_attendance_rate"`
}

// StudyProgramDetail represents study program with faculty info.
type StudyProgramDetail struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	FacultyID   uuid.UUID `json:"faculty_id"`
	FacultyName string    `json:"faculty_name"`
}

// SystemStats represents university-wide system metrics.
type SystemStats struct {
	TotalFaculties      int     `json:"total_faculties"`
	TotalStudyPrograms  int     `json:"total_study_programs"`
	TotalLecturers      int     `json:"total_lecturers"`
	TotalStudents       int     `json:"total_students"`
	TotalSchedules      int     `json:"total_schedules"`
	TotalSessionsHeld   int     `json:"total_sessions_held"`
	GlobalAvgAttendance float64 `json:"global_avg_attendance"`
}

// SystemUserItem represents a user in system user management.
type SystemUserItem struct {
	ID         uuid.UUID `json:"id"`
	ExternalID string    `json:"external_id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Role       UserRole  `json:"role"`
	ProdiName  *string   `json:"prodi_name"`
	DeviceID   *string   `json:"device_id"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

// SystemThemeSettings represents campus branding and identity config.
type SystemThemeSettings struct {
	PrimaryColor  string `json:"primary_color"`
	AccentColor   string `json:"accent_color"`
	CampusName    string `json:"campus_name"`
	CampusTagline string `json:"campus_tagline"`
	LogoURL       string `json:"logo_url"`
}

// CampusLocation represents a campus geofence point (multi-point geofence).
type CampusLocation struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	RadiusMeters int       `json:"radius_meters"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CampusApiConfig represents the dynamic API connection configuration for SIAKAD sync.
type CampusApiConfig struct {
	ApiURL        string `json:"campus_api_url"`
	ApiKey        string `json:"campus_api_key"`
	SyncCron      string `json:"campus_sync_cron"`
	CampusName    string `json:"campus_name"`
	CampusTagline string `json:"campus_tagline"`
	GeofenceMode  string `json:"geofence_mode"`
	MaxTolerance  int    `json:"max_tolerance_minutes"`
}

// ProdiLiveClassItem represents real-time live monitoring of today's class schedule for Admin Prodi.
type ProdiLiveClassItem struct {
	ScheduleID       uuid.UUID  `json:"schedule_id"`
	SessionID        *uuid.UUID `json:"session_id,omitempty"`
	CourseCode       string     `json:"course_code"`
	CourseName       string     `json:"course_name"`
	LecturerName     string     `json:"lecturer_name"`
	LecturerNIDN     string     `json:"lecturer_nidn"`
	RoomName         string     `json:"room_name"`
	DayOfWeek        int        `json:"day_of_week"`
	StartTime        string     `json:"start_time"`
	EndTime          string     `json:"end_time"`
	Status           string     `json:"status"` // SEDANG_BERLANGSUNG, SELESAI, BELUM_DIMULAI, TIDAK_MASUK
	TotalEnrolled    int        `json:"total_enrolled"`
	TotalHadir       int        `json:"total_hadir"`
	TotalIzinSakit   int        `json:"total_izin_sakit"`
	OpenedAt         *time.Time `json:"opened_at,omitempty"`
	ClosedAt         *time.Time `json:"closed_at,omitempty"`
}

// ─────────────────────────── ERRORS ───────────────────────────

// DomainError represents a business logic error with an HTTP-friendly code.
type DomainError struct {
	Code    int
	Message string
}

func (e *DomainError) Error() string { return e.Message }

var (
	ErrSessionClosed    = &DomainError{Code: 400, Message: "sesi presensi telah berakhir"}
	ErrQRExpired        = &DomainError{Code: 400, Message: "QR code telah kadaluwarsa"}
	ErrNotEnrolled      = &DomainError{Code: 403, Message: "mahasiswa tidak terdaftar di kelas ini"}
	ErrDeviceMismatch   = &DomainError{Code: 403, Message: "perangkat tidak sesuai, terindikasi titip absen"}
	ErrOutsideGeofence  = &DomainError{Code: 422, Message: "anda berada di luar radius presensi"}
	ErrAlreadySubmitted = &DomainError{Code: 409, Message: "presensi sudah tercatat untuk sesi ini"}
	ErrScheduleNotToday = &DomainError{Code: 422, Message: "sesi presensi hanya dapat dibuka pada hari jadwal perkuliahan"}
	ErrNotFound         = &DomainError{Code: 404, Message: "data tidak ditemukan"}
	ErrUnauthorized     = &DomainError{Code: 401, Message: "akses tidak diizinkan"}
	ErrForbidden        = &DomainError{Code: 403, Message: "akses ditolak"}
)
