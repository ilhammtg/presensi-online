package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ilham/presensi-online/backend/internal/domain"
)

// SystemHandler handles requests for Superadmin system administration.
type SystemHandler struct {
	systemRepo domain.SystemAdminRepository
}

// NewSystemHandler creates a new SystemHandler.
func NewSystemHandler(systemRepo domain.SystemAdminRepository) *SystemHandler {
	return &SystemHandler{systemRepo: systemRepo}
}

// GetSystemStats handles GET /v1/system/stats.
func (h *SystemHandler) GetSystemStats(c *gin.Context) {
	stats, err := h.systemRepo.GetSystemStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetStudyPrograms handles GET /v1/system/study-programs.
func (h *SystemHandler) GetStudyPrograms(c *gin.Context) {
	programs, err := h.systemRepo.GetAllStudyPrograms(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"study_programs": programs})
}

// GetUsers handles GET /v1/system/users.
func (h *SystemHandler) GetUsers(c *gin.Context) {
	var roleFilter *domain.UserRole
	if r := c.Query("role"); r != "" {
		ur := domain.UserRole(r)
		roleFilter = &ur
	}

	var prodiFilter *uuid.UUID
	if p := c.Query("prodi_id"); p != "" {
		if parsed, err := uuid.Parse(p); err == nil {
			prodiFilter = &parsed
		}
	}

	users, err := h.systemRepo.GetUsers(c.Request.Context(), roleFilter, prodiFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// GetThemeSettings handles GET /v1/system/theme (Public or authenticated for theme application).
func (h *SystemHandler) GetThemeSettings(c *gin.Context) {
	theme, err := h.systemRepo.GetThemeSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, theme)
}

// SaveThemeSettings handles POST /v1/system/theme (Superadmin only).
func (h *SystemHandler) SaveThemeSettings(c *gin.Context) {
	var req domain.SystemThemeSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format data tema tidak valid"})
		return
	}

	// Basic validation
	if req.PrimaryColor == "" {
		req.PrimaryColor = "#006633"
	}
	if req.AccentColor == "" {
		req.AccentColor = "#D4AF37"
	}
	if req.CampusName == "" {
		req.CampusName = "Universitas Almuslim"
	}

	if err := h.systemRepo.SaveThemeSettings(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Pengaturan identitas dan tema kampus berhasil disimpan",
		"settings": req,
	})
}

// GetCampusApiConfig handles GET /v1/system/config
func (h *SystemHandler) GetCampusApiConfig(c *gin.Context) {
	cfg, err := h.systemRepo.GetCampusApiConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// SaveCampusApiConfig handles POST /v1/system/config
func (h *SystemHandler) SaveCampusApiConfig(c *gin.Context) {
	var cfg domain.CampusApiConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format data konfigurasi tidak valid: " + err.Error()})
		return
	}
	if err := h.systemRepo.SaveCampusApiConfig(c.Request.Context(), &cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Konfigurasi koneksi SIAKAD kampus berhasil disimpan",
		"config":  cfg,
	})
}

// GetAllCampusLocations handles GET /v1/system/locations
func (h *SystemHandler) GetAllCampusLocations(c *gin.Context) {
	locations, err := h.systemRepo.GetAllCampusLocations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"locations": locations})
}

// GetActiveCampusLocations handles GET /v1/system/locations/active
func (h *SystemHandler) GetActiveCampusLocations(c *gin.Context) {
	locations, err := h.systemRepo.GetActiveCampusLocations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"locations": locations})
}

// CreateCampusLocation handles POST /v1/system/locations
func (h *SystemHandler) CreateCampusLocation(c *gin.Context) {
	var loc domain.CampusLocation
	if err := c.ShouldBindJSON(&loc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data lokasi tidak valid: " + err.Error()})
		return
	}
	if loc.ID == uuid.Nil {
		loc.ID = uuid.New()
	}
	if loc.RadiusMeters <= 0 {
		loc.RadiusMeters = 80
	}
	if err := h.systemRepo.CreateCampusLocation(c.Request.Context(), &loc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message":  "Titik geofence kampus berhasil ditambahkan",
		"location": loc,
	})
}

// UpdateCampusLocation handles PUT /v1/system/locations/:id
func (h *SystemHandler) UpdateCampusLocation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}

	var loc domain.CampusLocation
	if err := c.ShouldBindJSON(&loc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data lokasi tidak valid: " + err.Error()})
		return
	}
	loc.ID = id

	if err := h.systemRepo.UpdateCampusLocation(c.Request.Context(), &loc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "Titik geofence kampus berhasil diperbarui",
		"location": loc,
	})
}

// DeleteCampusLocation handles DELETE /v1/system/locations/:id
func (h *SystemHandler) DeleteCampusLocation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}

	if err := h.systemRepo.DeleteCampusLocation(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Titik geofence berhasil dihapus"})
}

// TriggerSync handles POST /v1/system/sync
func (h *SystemHandler) TriggerSync(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Sinkronisasi SIAKAD Kampus berhasil dipicu.",
		"status":  "SYNC_TRIGGERED",
	})
}

