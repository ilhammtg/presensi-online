package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ilham/presensi-online/backend/internal/app"
	"github.com/ilham/presensi-online/backend/internal/domain"
	authmw "github.com/ilham/presensi-online/backend/internal/middleware/auth"
)

// SessionHandler handles HTTP requests for class sessions.
type SessionHandler struct {
	sessionUC *app.SessionUseCase
}

// NewSessionHandler creates a new SessionHandler.
func NewSessionHandler(sessionUC *app.SessionUseCase) *SessionHandler {
	return &SessionHandler{
		sessionUC: sessionUC,
	}
}

// OpenSessionRequest represents the body payload to open a session.
type OpenSessionRequest struct {
	ScheduleID      string `json:"schedule_id" binding:"required,uuid"`
	MeetingNo       int    `json:"meeting_no" binding:"required,min=1,max=16"`
	DurationMinutes int    `json:"duration_minutes"`
}

// OpenSession handles POST /v1/sessions (Dosen only).
func (h *SessionHandler) OpenSession(c *gin.Context) {
	var req OpenSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	lecturerID, ok := authmw.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user ID in token"})
		return
	}

	schedUUID, err := uuid.Parse(req.ScheduleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule_id"})
		return
	}

	session, err := h.sessionUC.OpenSession(c.Request.Context(), app.OpenSessionRequest{
		ScheduleID:      schedUUID,
		LecturerID:      lecturerID,
		MeetingNo:       req.MeetingNo,
		DurationMinutes: req.DurationMinutes,
	})
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			c.JSON(domainErr.Code, gin.H{"error": domainErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, remaining, _ := h.sessionUC.GetCurrentQR(c.Request.Context(), session.ID)

	c.JSON(http.StatusCreated, gin.H{
		"session":            session,
		"current_qr_token":   token,
		"expires_in_seconds": remaining,
	})
}

// CloseSessionRequest represents the body payload to close a session.
type CloseSessionRequest struct {
	BAPTopic *string `json:"bap_topic"`
}

// CloseSession handles PATCH /v1/sessions/:id/close (Dosen only).
func (h *SessionHandler) CloseSession(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	lecturerID, ok := authmw.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user ID in token"})
		return
	}

	var req CloseSessionRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.sessionUC.CloseSession(c.Request.Context(), app.CloseSessionRequest{
		SessionID:  sessionID,
		LecturerID: lecturerID,
		BAPTopic:   req.BAPTopic,
	}); err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			c.JSON(domainErr.Code, gin.H{"error": domainErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "sesi perkuliahan berhasil ditutup"})
}

// GetSession handles GET /v1/sessions/:id.
func (h *SessionHandler) GetSession(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	session, err := h.sessionUC.GetSessionDetails(c.Request.Context(), sessionID)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			c.JSON(domainErr.Code, gin.H{"error": domainErr.Message})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// GetAttendees handles GET /v1/sessions/:id/attendees.
func (h *SessionHandler) GetAttendees(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	attendees, err := h.sessionUC.GetAttendees(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"count":      len(attendees),
		"attendees":  attendees,
	})
}

// GetQR handles GET /v1/sessions/:id/qr.
func (h *SessionHandler) GetQR(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	token, remaining, err := h.sessionUC.GetCurrentQR(c.Request.Context(), sessionID)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			c.JSON(domainErr.Code, gin.H{"error": domainErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id":         sessionID,
		"token":              token,
		"expires_in_seconds": remaining,
	})
}

// GetActiveSession handles GET /v1/sessions/active?schedule_id=...
func (h *SessionHandler) GetActiveSession(c *gin.Context) {
	schedStr := c.Query("schedule_id")
	if schedStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "schedule_id query parameter is required"})
		return
	}

	schedID, err := uuid.Parse(schedStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule_id"})
		return
	}

	session, token, remaining, err := h.sessionUC.GetActiveSessionBySchedule(c.Request.Context(), schedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active session found for this schedule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session":            session,
		"current_qr_token":   token,
		"expires_in_seconds": remaining,
	})
}

// GetDosenSchedules handles GET /v1/dosen/schedules.
func (h *SessionHandler) GetDosenSchedules(c *gin.Context) {
	lecturerID, ok := authmw.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	schedules, err := h.sessionUC.GetDosenSchedules(c.Request.Context(), lecturerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"schedules": schedules,
	})
}

// GetScheduleStudents handles GET /v1/schedules/:id/students.
func (h *SessionHandler) GetScheduleStudents(c *gin.Context) {
	idStr := c.Param("id")
	schedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return
	}

	students, err := h.sessionUC.GetEnrolledStudents(c.Request.Context(), schedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"schedule_id": schedID,
		"students":    students,
		"count":       len(students),
	})
}

// GetScheduleSessions handles GET /v1/schedules/:id/sessions.
func (h *SessionHandler) GetScheduleSessions(c *gin.Context) {
	idStr := c.Param("id")
	schedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return
	}

	sessions, nextMeeting, err := h.sessionUC.GetScheduleSessions(c.Request.Context(), schedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"schedule_id":     schedID,
		"sessions":        sessions,
		"next_meeting_no": nextMeeting,
	})
}

// GetClassRecap handles GET /v1/schedules/:id/recap.
func (h *SessionHandler) GetClassRecap(c *gin.Context) {
	idStr := c.Param("id")
	schedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return
	}

	recap, err := h.sessionUC.GetClassRecap(c.Request.Context(), schedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, recap)
}

// PermissionRequest represents body for marking permission.
type PermissionRequest struct {
	StudentID string `json:"student_id" binding:"required,uuid"`
	Status    string `json:"status" binding:"required,oneof=izin sakit hadir alpa"`
	Notes     string `json:"notes"`
}

// MarkPermission handles POST /v1/sessions/:id/permission.
func (h *SessionHandler) MarkPermission(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	lecturerID, ok := authmw.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req PermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	studID, err := uuid.Parse(req.StudentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}

	if err := h.sessionUC.MarkPermission(c.Request.Context(), sessionID, studID, lecturerID, domain.AttendanceStatus(req.Status), req.Notes); err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			c.JSON(domainErr.Code, gin.H{"error": domainErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "status kehadiran berhasil dicatat",
	})
}

// GetDosenPermissions handles GET /v1/dosen/permissions
func (h *SessionHandler) GetDosenPermissions(c *gin.Context) {
	lecturerID, ok := authmw.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	pendingOnly := c.Query("status") == "pending"
	permissions, err := h.sessionUC.GetPermissionsByLecturer(c.Request.Context(), lecturerID, pendingOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if permissions == nil {
		permissions = []*domain.PermissionApprovalItem{}
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// ApprovePermissionRequest represents body to approve or reject permission.
type ApprovePermissionRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Notes  string `json:"notes"`
}

// ApproveDosenPermission handles POST /v1/dosen/permissions/:id/approve
func (h *SessionHandler) ApproveDosenPermission(c *gin.Context) {
	idStr := c.Param("id")
	attID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attendance id"})
		return
	}

	lecturerID, ok := authmw.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req ApprovePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	approved := req.Action == "approve"
	if err := h.sessionUC.ApprovePermission(c.Request.Context(), attID, lecturerID, approved, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	msg := "Permohonan izin berhasil disetujui."
	if !approved {
		msg = "Permohonan izin ditolak (dicatat sebagai alpa)."
	}
	c.JSON(http.StatusOK, gin.H{"message": msg})
}

// StudentPermissionRequest represents body for student app request.
type StudentPermissionRequest struct {
	SessionID     string  `json:"session_id" binding:"required,uuid"`
	Status        string  `json:"status" binding:"required,oneof=izin sakit"`
	Notes         string  `json:"notes"`
	AttachmentURL *string `json:"attachment_url"`
}

// SubmitStudentPermission handles POST /v1/attendance/permission-request
func (h *SessionHandler) SubmitStudentPermission(c *gin.Context) {
	studentID, ok := authmw.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req StudentPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessID, err := uuid.Parse(req.SessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	att, err := h.sessionUC.SubmitStudentPermission(
		c.Request.Context(),
		sessID,
		studentID,
		domain.AttendanceStatus(req.Status),
		req.Notes,
		req.AttachmentURL,
	)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			c.JSON(domainErr.Code, gin.H{"error": domainErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Permohonan izin berhasil diajukan, menunggu persetujuan dosen.",
		"attendance": att,
	})
}

