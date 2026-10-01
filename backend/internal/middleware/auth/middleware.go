package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ilham/presensi-online/backend/internal/domain"
)

const (
	ContextKeyUserID   = "user_id"
	ContextKeyUserRole = "user_role"
	ContextKeyDeviceID = "device_id"
	ContextKeyClaims   = "jwt_claims"
)

// Middleware returns a Gin middleware that validates the JWT Bearer token.
func Middleware(jwtSvc *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtSvc.ValidateToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Store claims in context for downstream handlers
		c.Set(ContextKeyClaims, claims)
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUserRole, claims.Role)
		c.Set(ContextKeyDeviceID, claims.DeviceID)

		c.Next()
	}
}

// RequireRoles returns a Gin middleware that enforces role-based access control.
// Usage: router.Use(RequireRoles(domain.RoleDosen, domain.RoleAdminProdi))
func RequireRoles(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	allowed := make(map[domain.UserRole]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyUserRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}

		userRole, ok := role.(domain.UserRole)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid role type"})
			return
		}

		if _, permitted := allowed[userRole]; !permitted {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied: insufficient role"})
			return
		}

		c.Next()
	}
}

// GetUserID extracts the user UUID from the Gin context (set by Middleware).
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextKeyUserID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// GetUserRole extracts the user role from the Gin context.
func GetUserRole(c *gin.Context) (domain.UserRole, bool) {
	val, exists := c.Get(ContextKeyUserRole)
	if !exists {
		return "", false
	}
	role, ok := val.(domain.UserRole)
	return role, ok
}
