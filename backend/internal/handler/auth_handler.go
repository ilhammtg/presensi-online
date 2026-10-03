package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/ilham/presensi-online/backend/internal/domain"
	authmw "github.com/ilham/presensi-online/backend/internal/middleware/auth"
)

// AuthHandler handles authentication requests.
type AuthHandler struct {
	userRepo domain.UserRepository
	jwtSvc   *authmw.JWTService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(userRepo domain.UserRepository, jwtSvc *authmw.JWTService) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, jwtSvc: jwtSvc}
}

// LoginRequest is the JSON body for the login endpoint.
type LoginRequest struct {
	Username string `json:"username"` // NIM, NIDN, or Email
	Email    string `json:"email"`    // Backward compatibility with email-only clients
	Password string `json:"password" binding:"required"`
	DeviceID string `json:"device_id"` // Optional for mobile device binding
}

// Login handles POST /v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	identifier := strings.TrimSpace(req.Username)
	if identifier == "" {
		identifier = strings.TrimSpace(req.Email)
	}
	if identifier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username atau email wajib diisi"})
		return
	}

	ctx := c.Request.Context()

	var user *domain.User
	var err error

	// If identifier contains '@', try finding by email first; otherwise by external_id (NIDN / NIM)
	if strings.Contains(identifier, "@") {
		user, err = h.userRepo.FindByEmail(ctx, identifier)
		if err != nil {
			user, err = h.userRepo.FindByExternalID(ctx, identifier)
		}
	} else {
		user, err = h.userRepo.FindByExternalID(ctx, identifier)
		if err != nil {
			user, err = h.userRepo.FindByEmail(ctx, identifier)
		}
	}

	if err != nil || user == nil {
		// Return generic error to prevent user enumeration
		c.JSON(http.StatusUnauthorized, gin.H{"error": "username/email atau password salah"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "akun tidak aktif"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "email atau password salah"})
		return
	}

	// Device binding: bind device on first login or verify on subsequent logins
	deviceID := req.DeviceID
	if deviceID != "" {
		if user.DeviceID == nil {
			// First login from this device — bind it
			if err := h.userRepo.UpdateDeviceID(ctx, user.ID, deviceID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengikat perangkat"})
				return
			}
		} else if *user.DeviceID != deviceID {
			// Different device — reject (potential account sharing / titip absen)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "perangkat tidak dikenal. Gunakan perangkat yang terdaftar atau hubungi admin.",
			})
			return
		}
	}

	tokens, err := h.jwtSvc.GenerateTokenPair(user.ID, user.Role, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membuat token"})
		return
	}

	profile, err := h.userRepo.GetProfile(ctx, user.ID)
	var userPayload gin.H
	if err == nil && profile != nil {
		userPayload = gin.H{
			"id":           profile.ID,
			"external_id":  profile.ExternalID,
			"name":         profile.Name,
			"email":        profile.Email,
			"role":         profile.Role,
			"prodi_id":     profile.ProdiID,
			"prodi_name":   profile.ProdiName,
			"faculty_name": profile.FacultyName,
			"avatar_url":   profile.AvatarURL,
		}
	} else {
		userPayload = gin.H{
			"id":          user.ID,
			"external_id": user.ExternalID,
			"name":        user.Name,
			"email":       user.Email,
			"role":        user.Role,
			"prodi_id":    user.ProdiID,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_at":    tokens.ExpiresAt,
		"user":          userPayload,
	})
}

// Me handles GET /v1/auth/me — returns the current user's profile with academic details.
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := authmw.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	profile, err := h.userRepo.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// ChangePasswordRequest represents body for changing password.
type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// ChangePassword handles POST /v1/auth/change-password.
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := authmw.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input tidak valid: " + err.Error()})
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "konfirmasi kata sandi tidak cocok"})
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kata sandi lama salah"})
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memproses kata sandi baru"})
		return
	}

	if err := h.userRepo.UpdatePassword(c.Request.Context(), userID, string(newHash)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menyimpan kata sandi baru"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "kata sandi berhasil diubah"})
}

// UpdateProfilePhotoRequest represents body for updating profile photo.
type UpdateProfilePhotoRequest struct {
	AvatarURL string `json:"avatar_url" binding:"required"`
}

// UpdateProfilePhoto handles POST /v1/auth/profile-photo.
func (h *AuthHandler) UpdateProfilePhoto(c *gin.Context) {
	userID, ok := authmw.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	var req UpdateProfilePhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data foto tidak valid"})
		return
	}

	if err := h.userRepo.UpdateAvatar(c.Request.Context(), userID, req.AvatarURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menyimpan foto profil"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "foto profil berhasil diperbarui",
		"avatar_url": req.AvatarURL,
	})
}
