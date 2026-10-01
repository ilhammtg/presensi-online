// Handler for attendance scanning and management
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ilham/presensi-online/backend/internal/app"
	"github.com/ilham/presensi-online/backend/internal/domain"
	authmw "github.com/ilham/presensi-online/backend/internal/middleware/auth"
)

// AttendanceHandler handles HTTP requests related to attendance.
type AttendanceHandler struct {
	attendanceUC *app.AttendanceUseCase
}

// NewAttendanceHandler creates a new AttendanceHandler.
func NewAttendanceHandler(attendanceUC *app.AttendanceUseCase) *AttendanceHandler {
	return &AttendanceHandler{
		attendanceUC: attendanceUC,
	}
}

// ScanRequest is the JSON body for the scan endpoint.
type ScanRequest struct {
	SessionID string  `json:"session_id" binding:"required,uuid"`
	QRToken   string  `json:"qr_token" binding:"required"`
	DeviceID  string  `json:"device_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

// ScanAttendance handles POST /v1/attendance/scan
// Hot path: executed by every student scanning the dynamic QR code.
func (h *AttendanceHandler) ScanAttendance(c *gin.Context) {
	var req ScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	studentID, ok := authmw.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user ID in token"})
		return
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	ucReq := app.ScanAttendanceRequest{
		SessionID: sessionID,
		StudentID: studentID,
		QRToken:   req.QRToken,
		DeviceID:  req.DeviceID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}

	result, err := h.attendanceUC.ScanAttendance(c.Request.Context(), ucReq)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			c.JSON(domainErr.Code, gin.H{"error": domainErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "presensi berhasil dicatat",
		"attendance_id":   result.AttendanceID,
		"status":          result.Status,
		"scanned_at":      result.ScannedAt,
		"distance_meters": result.Distance,
	})
}

// OverrideRequest is the JSON body for lecturer manual override.
type OverrideRequest struct {
	Status domain.AttendanceStatus `json:"status" binding:"required"`
	Notes  *string                 `json:"notes"`
}

// OverrideAttendance handles PATCH /v1/attendances/:id (Dosen only).
func (h *AttendanceHandler) OverrideAttendance(c *gin.Context) {
	idStr := c.Param("id")
	attID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attendance id"})
		return
	}

	lecturerID, ok := authmw.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user ID in token"})
		return
	}

	var req OverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	if err := h.attendanceUC.OverrideAttendance(c.Request.Context(), app.OverrideAttendanceRequest{
		AttendanceID: attID,
		LecturerID:   lecturerID,
		Status:       req.Status,
		Notes:        req.Notes,
	}); err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			c.JSON(domainErr.Code, gin.H{"error": domainErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status presensi berhasil diperbarui"})
}
