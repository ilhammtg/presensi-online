// Package redisrepo provides Redis-backed implementations for QR token and session caching.
package redisrepo

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	keyPrefixQRSeed    = "qr:seed:"
	keyPrefixActiveSession = "session:active:"
	defaultTOTPPeriod  = 15 // seconds
)

// QRTokenCacheRepo implements domain.QRTokenCache using Redis.
type QRTokenCacheRepo struct {
	client *redis.Client
	period int // TOTP window in seconds
}

// NewQRTokenCacheRepo creates a new QRTokenCacheRepo.
func NewQRTokenCacheRepo(client *redis.Client, periodSeconds int) *QRTokenCacheRepo {
	if periodSeconds <= 0 {
		periodSeconds = defaultTOTPPeriod
	}
	return &QRTokenCacheRepo{client: client, period: periodSeconds}
}

// SetSeed stores the TOTP seed for a session in Redis.
// TTL is set to 24 hours (maximum session duration).
func (r *QRTokenCacheRepo) SetSeed(ctx context.Context, sessionID uuid.UUID, seed string) error {
	key := keyPrefixQRSeed + sessionID.String()
	return r.client.Set(ctx, key, seed, 24*time.Hour).Err()
}

// GetSeed retrieves the TOTP seed for a session.
func (r *QRTokenCacheRepo) GetSeed(ctx context.Context, sessionID uuid.UUID) (string, error) {
	key := keyPrefixQRSeed + sessionID.String()
	seed, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("QR seed not found for session %s", sessionID)
		}
		return "", fmt.Errorf("QRTokenCache.GetSeed: %w", err)
	}
	return seed, nil
}

// DeleteSeed removes the TOTP seed when a session is closed.
func (r *QRTokenCacheRepo) DeleteSeed(ctx context.Context, sessionID uuid.UUID) error {
	key := keyPrefixQRSeed + sessionID.String()
	return r.client.Del(ctx, key).Err()
}

// ValidateToken checks if the given token is valid for the session.
// It checks direct match with the static session token (seed), with TOTP fallback.
func (r *QRTokenCacheRepo) ValidateToken(ctx context.Context, sessionID uuid.UUID, token string) (bool, error) {
	seed, err := r.GetSeed(ctx, sessionID)
	if err != nil {
		return false, err
	}

	// 1. Direct match with static session token (1x QR token for the entire session)
	if strings.EqualFold(strings.TrimSpace(token), strings.TrimSpace(seed)) {
		return true, nil
	}

	// 2. Fallback TOTP check for backward compatibility
	now := time.Now().Unix()
	period := int64(r.period)
	for _, counter := range []int64{now / period, (now / period) - 1} {
		expected := generateTOTP(seed, counter, 6)
		if expected == token {
			return true, nil
		}
	}
	return false, nil
}

// GenerateCurrentToken returns the static token for a session.
// With static QR, the token remains valid and fixed until the session is closed.
func (r *QRTokenCacheRepo) GenerateCurrentToken(ctx context.Context, sessionID uuid.UUID) (string, int64, error) {
	seed, err := r.GetSeed(ctx, sessionID)
	if err != nil {
		return "", 0, err
	}

	// Static token has no 15s expiration; it lives until session closure
	return seed, 0, nil
}

// generateTOTP generates a TOTP token using HMAC-SHA1 (RFC 6238 compatible).
// seed is the base32-encoded secret; counter is the time step.
func generateTOTP(seed string, counter int64, digits int) string {
	// Pad seed to valid base32 length
	seed = strings.ToUpper(strings.ReplaceAll(seed, " ", ""))
	if n := len(seed) % 8; n != 0 {
		seed += strings.Repeat("=", 8-n)
	}

	key, err := base32.StdEncoding.DecodeString(seed)
	if err != nil {
		// Fallback: use seed as raw bytes
		key = []byte(seed)
	}

	// HOTP: HMAC-SHA1(key, counter)
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(counter))

	mac := hmac.New(sha1.New, key)
	mac.Write(msg)
	h := mac.Sum(nil)

	// Dynamic truncation (RFC 4226)
	offset := h[len(h)-1] & 0x0f
	code := binary.BigEndian.Uint32(h[offset:offset+4]) & 0x7fffffff

	mod := uint32(math.Pow10(digits))
	return fmt.Sprintf("%0*d", digits, code%mod)
}

// ─────────────────────────── SESSION CACHE ───────────────────────────

// SessionCacheRepo implements domain.SessionCache using Redis.
type SessionCacheRepo struct {
	client *redis.Client
}

// NewSessionCacheRepo creates a new SessionCacheRepo.
func NewSessionCacheRepo(client *redis.Client) *SessionCacheRepo {
	return &SessionCacheRepo{client: client}
}

// SetActiveSession marks a schedule as having an open session.
func (r *SessionCacheRepo) SetActiveSession(ctx context.Context, scheduleID, sessionID uuid.UUID) error {
	key := keyPrefixActiveSession + scheduleID.String()
	return r.client.Set(ctx, key, sessionID.String(), 24*time.Hour).Err()
}

// GetActiveSession retrieves the active session ID for a schedule.
func (r *SessionCacheRepo) GetActiveSession(ctx context.Context, scheduleID uuid.UUID) (uuid.UUID, error) {
	key := keyPrefixActiveSession + scheduleID.String()
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return uuid.Nil, fmt.Errorf("no active session for schedule %s", scheduleID)
		}
		return uuid.Nil, fmt.Errorf("SessionCache.GetActiveSession: %w", err)
	}
	return uuid.Parse(val)
}

// DeleteActiveSession removes the active session marker when closed.
func (r *SessionCacheRepo) DeleteActiveSession(ctx context.Context, scheduleID uuid.UUID) error {
	key := keyPrefixActiveSession + scheduleID.String()
	return r.client.Del(ctx, key).Err()
}
