package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ilham/presensi-online/backend/internal/domain"
	authmw "github.com/ilham/presensi-online/backend/internal/middleware/auth"
)

// ProdiHandler handles requests for Program Study monitoring.
type ProdiHandler struct {
	prodiRepo  domain.ProdiMonitoringRepository
	userRepo   domain.UserRepository
	systemRepo domain.SystemAdminRepository
}

// NewProdiHandler creates a new ProdiHandler.
func NewProdiHandler(
	prodiRepo domain.ProdiMonitoringRepository,
	userRepo domain.UserRepository,
	systemRepo domain.SystemAdminRepository,
) *ProdiHandler {
	return &ProdiHandler{
		prodiRepo:  prodiRepo,
		userRepo:   userRepo,
		systemRepo: systemRepo,
	}
}

// resolveProdiID enforces strict isolation: admin_prodi is locked to their own prodi_id.
func (h *ProdiHandler) resolveProdiID(c *gin.Context) (uuid.UUID, error) {
	userID, ok := authmw.GetUserID(c)
	if !ok {
		return uuid.Nil, errors.New("unauthenticated")
	}

	role, _ := authmw.GetUserRole(c)

	if role == domain.RoleAdminProdi {
		user, err := h.userRepo.FindByID(c.Request.Context(), userID)
		if err != nil {
			return uuid.Nil, err
		}
		if user.ProdiID == nil {
			return uuid.Nil, errors.New("admin prodi belum ditugaskan ke program studi manapun")
		}
		return *user.ProdiID, nil
	}

	if role == domain.RoleSuperadmin {
		reqProdi := c.Query("prodi_id")
		if reqProdi != "" {
			parsed, err := uuid.Parse(reqProdi)
			if err == nil {
				return parsed, nil
			}
		}
		// Fallback to first available prodi for superadmin
		programs, err := h.systemRepo.GetAllStudyPrograms(c.Request.Context())
		if err == nil && len(programs) > 0 {
			return programs[0].ID, nil
		}
	}

	return uuid.Nil, errors.New("akses program studi tidak diizinkan")
}

// GetProdiOverview handles GET /v1/prodi/overview.
func (h *ProdiHandler) GetProdiOverview(c *gin.Context) {
	prodiID, err := h.resolveProdiID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	overview, err := h.prodiRepo.GetProdiOverview(c.Request.Context(), prodiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// GetLecturersCompliance handles GET /v1/prodi/lecturers-compliance.
func (h *ProdiHandler) GetLecturersCompliance(c *gin.Context) {
	prodiID, err := h.resolveProdiID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	results, err := h.prodiRepo.GetLecturersCompliance(c.Request.Context(), prodiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"prodi_id":  prodiID,
		"lecturers": results,
	})
}

// GetStudentsAtRisk handles GET /v1/prodi/students-at-risk.
func (h *ProdiHandler) GetStudentsAtRisk(c *gin.Context) {
	prodiID, err := h.resolveProdiID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	results, err := h.prodiRepo.GetStudentsAtRisk(c.Request.Context(), prodiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"prodi_id": prodiID,
		"students": results,
	})
}

// GetProdiClasses handles GET /v1/prodi/classes.
func (h *ProdiHandler) GetProdiClasses(c *gin.Context) {
	prodiID, err := h.resolveProdiID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	classes, err := h.prodiRepo.GetProdiClasses(c.Request.Context(), prodiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"prodi_id": prodiID,
		"classes":  classes,
	})
}

// GetLiveTodayClasses handles GET /v1/prodi/live-today.
func (h *ProdiHandler) GetLiveTodayClasses(c *gin.Context) {
	prodiID, err := h.resolveProdiID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	classes, err := h.prodiRepo.GetLiveTodayClasses(c.Request.Context(), prodiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"prodi_id": prodiID,
		"classes":  classes,
	})
}

