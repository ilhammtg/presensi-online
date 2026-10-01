// Package auth provides JWT token generation and validation.
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/ilham/presensi-online/backend/internal/domain"
)

// Claims represents the JWT claims payload.
type Claims struct {
	UserID   uuid.UUID       `json:"uid"`
	Role     domain.UserRole `json:"role"`
	DeviceID string          `json:"did,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair holds an access token and a refresh token.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"` // Unix timestamp
}

// JWTService handles JWT token operations.
type JWTService struct {
	secret         []byte
	accessTTL      time.Duration
	refreshTTL     time.Duration
}

// NewJWTService creates a new JWTService.
func NewJWTService(secret string, accessTTLMins, refreshTTLDays int) *JWTService {
	return &JWTService{
		secret:     []byte(secret),
		accessTTL:  time.Duration(accessTTLMins) * time.Minute,
		refreshTTL: time.Duration(refreshTTLDays) * 24 * time.Hour,
	}
}

// GenerateTokenPair creates a new access + refresh token pair for a user.
func (s *JWTService) GenerateTokenPair(userID uuid.UUID, role domain.UserRole, deviceID string) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(s.accessTTL)

	accessClaims := &Claims{
		UserID:   userID,
		Role:     role,
		DeviceID: deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			Subject:   userID.String(),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("JWTService: failed to sign access token: %w", err)
	}

	refreshClaims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTTL)),
			Subject:   userID.String(),
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("JWTService: failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry.Unix(),
	}, nil
}

// ValidateToken parses and validates a JWT token string.
func (s *JWTService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
