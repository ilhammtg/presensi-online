package crypto_test

import (
	"encoding/json"
	"testing"

	"github.com/ilham/presensi-online/backend/internal/crypto"
)

const testKeyHex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	aes, err := crypto.NewAESGCM(testKeyHex)
	if err != nil {
		t.Fatalf("NewAESGCM error: %v", err)
	}

	plaintext := []byte(`{"sync_timestamp":"2026-09-22T08:00:00Z","academic_year":"2026/2027"}`)

	payload, err := aes.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt error: %v", err)
	}

	if payload.IV == "" || payload.Ciphertext == "" || payload.Tag == "" {
		t.Fatal("encrypted payload fields must not be empty")
	}

	decrypted, err := aes.Decrypt(payload)
	if err != nil {
		t.Fatalf("Decrypt error: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("round-trip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestDecrypt_TamperedTag_Fails(t *testing.T) {
	aes, err := crypto.NewAESGCM(testKeyHex)
	if err != nil {
		t.Fatalf("NewAESGCM error: %v", err)
	}

	payload, _ := aes.Encrypt([]byte("sensitive data"))
	payload.Tag = "AAAAAAAAAAAAAAAAAAAAAA==" // tampered tag

	_, err = aes.Decrypt(payload)
	if err == nil {
		t.Error("expected decryption to fail with tampered tag, but it succeeded")
	}
}

func TestEncryptDecrypt_LargePayload(t *testing.T) {
	aes, err := crypto.NewAESGCM(testKeyHex)
	if err != nil {
		t.Fatalf("NewAESGCM error: %v", err)
	}

	// Simulate a large campus sync payload
	data := map[string]interface{}{
		"sync_timestamp": "2026-09-22T08:00:00Z",
		"academic_year":  "2026/2027",
		"users":          make([]map[string]string, 1000),
	}
	plaintext, _ := json.Marshal(data)

	payload, err := aes.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt error: %v", err)
	}

	decrypted, err := aes.Decrypt(payload)
	if err != nil {
		t.Fatalf("Decrypt error: %v", err)
	}

	if len(decrypted) != len(plaintext) {
		t.Errorf("length mismatch: got %d, want %d", len(decrypted), len(plaintext))
	}
}

func TestNewAESGCM_InvalidKey(t *testing.T) {
	_, err := crypto.NewAESGCM("tooshort")
	if err == nil {
		t.Error("expected error for invalid key length")
	}
}

func TestGenerateKeyHex(t *testing.T) {
	key, err := crypto.GenerateKeyHex()
	if err != nil {
		t.Fatalf("GenerateKeyHex error: %v", err)
	}
	if len(key) != 64 {
		t.Errorf("expected 64-char hex key, got %d chars", len(key))
	}
}
