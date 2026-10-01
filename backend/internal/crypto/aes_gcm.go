// Package crypto provides AES-256-GCM authenticated encryption helpers
// for secure data exchange between the Campus API and the sync worker.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// ErrInvalidKeyLength is returned when the AES key is not 32 bytes.
var ErrInvalidKeyLength = errors.New("AES key must be exactly 32 bytes (256 bits)")

// AESGCM provides AES-256-GCM encrypt/decrypt operations.
type AESGCM struct {
	key []byte
}

// NewAESGCM creates a new AESGCM instance from a 64-character hex-encoded key.
func NewAESGCM(keyHex string) (*AESGCM, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid AES key hex: %w", err)
	}
	if len(key) != 32 {
		return nil, ErrInvalidKeyLength
	}
	return &AESGCM{key: key}, nil
}

// EncryptedPayload represents the wire format of an encrypted API payload,
// matching the "Format Amplop API" defined in the system specification.
type EncryptedPayload struct {
	IV         string `json:"iv"`         // Base64-encoded 96-bit nonce
	Ciphertext string `json:"ciphertext"` // Base64-encoded AES-GCM ciphertext
	Tag        string `json:"tag"`        // Base64-encoded 128-bit authentication tag (appended to ciphertext by Go's GCM)
}

// Encrypt encrypts plaintext using AES-256-GCM with a random 96-bit nonce.
// The authentication tag is automatically appended to the ciphertext by Go's
// cipher.AEAD.Seal implementation. The tag field in EncryptedPayload is
// extracted from the last 16 bytes of the combined ciphertext+tag slice.
func (a *AESGCM) Encrypt(plaintext []byte) (*EncryptedPayload, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate cryptographically secure random nonce (96-bit / 12 bytes)
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Seal appends the authentication tag to the ciphertext
	combined := gcm.Seal(nil, nonce, plaintext, nil)

	// Split ciphertext and tag (tag is always the last gcm.Overhead() bytes)
	tagSize := gcm.Overhead()
	ciphertext := combined[:len(combined)-tagSize]
	tag := combined[len(combined)-tagSize:]

	return &EncryptedPayload{
		IV:         base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		Tag:        base64.StdEncoding.EncodeToString(tag),
	}, nil
}

// Decrypt decrypts an EncryptedPayload using AES-256-GCM.
// It verifies the authentication tag automatically via GCM.Open.
func (a *AESGCM) Decrypt(payload *EncryptedPayload) ([]byte, error) {
	nonce, err := base64.StdEncoding.DecodeString(payload.IV)
	if err != nil {
		return nil, fmt.Errorf("invalid IV encoding: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(payload.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("invalid ciphertext encoding: %w", err)
	}

	tag, err := base64.StdEncoding.DecodeString(payload.Tag)
	if err != nil {
		return nil, fmt.Errorf("invalid tag encoding: %w", err)
	}

	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Recombine ciphertext + tag for GCM.Open
	combined := append(ciphertext, tag...)

	plaintext, err := gcm.Open(nil, nonce, combined, nil)
	if err != nil {
		// Authentication failure — data may be tampered
		return nil, fmt.Errorf("decryption failed (authentication tag mismatch): %w", err)
	}

	return plaintext, nil
}

// GenerateKeyHex generates a random 256-bit AES key and returns it as a hex string.
// Use this to generate the AES_KEY_HEX config value.
func GenerateKeyHex() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}
	return hex.EncodeToString(key), nil
}
