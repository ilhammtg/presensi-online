package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authmw "github.com/ilham/presensi-online/backend/internal/middleware/auth"
	"github.com/ilham/presensi-online/backend/internal/ws"
)

// WSHandler handles WebSocket connection requests.
type WSHandler struct {
	hub    *ws.Hub
	jwtSvc *authmw.JWTService
}

// NewWSHandler creates a new WSHandler.
func NewWSHandler(hub *ws.Hub, jwtSvc *authmw.JWTService) *WSHandler {
	return &WSHandler{
		hub:    hub,
		jwtSvc: jwtSvc,
	}
}

// ServeWS handles GET /v1/ws.
// Supports auth via query param "?token=..." or "Authorization: Bearer ...".
// Room identification via "?session_id=..." or "?schedule_id=...".
func (h *WSHandler) ServeWS(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authentication token"})
		return
	}

	claims, err := h.jwtSvc.ValidateToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	roomIDStr := c.Query("session_id")
	if roomIDStr == "" {
		roomIDStr = c.Query("schedule_id")
	}
	if roomIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing session_id or schedule_id"})
		return
	}

	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room UUID"})
		return
	}

	if err := h.hub.ServeWS(c.Writer, c.Request, roomID, claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upgrade websocket: " + err.Error()})
		return
	}
}
