// Mock Campus API — simulates the university's SIAKAD system.
// Generates encrypted payloads using AES-256-GCM as specified in docs.md.
package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ilham/presensi-online/backend/internal/crypto"
)

//go:embed payload.json
var embeddedPayloadJSON []byte

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "9000"
	}
	keyHex := os.Getenv("AES_KEY_HEX")
	if keyHex == "" {
		// Default dev key — 32 zero bytes
		keyHex = "0000000000000000000000000000000000000000000000000000000000000000"
	}

	aes, err := crypto.NewAESGCM(keyHex)
	if err != nil {
		log.Fatalf("failed to init AES: %v", err)
	}

	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "mock-campus-api"})
	})

	// Main sync endpoint: returns encrypted campus data
	r.GET("/api/v1/sync/payload", func(c *gin.Context) {
		// Verify API key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "mock-api-key-change-me" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			return
		}

		var payload map[string]interface{}
		if len(embeddedPayloadJSON) > 0 {
			if err := json.Unmarshal(embeddedPayloadJSON, &payload); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse payload: " + err.Error()})
				return
			}
		} else {
			payload = buildFallbackPayload()
		}

		payload["sync_timestamp"] = time.Now().UTC().Format(time.RFC3339)

		plaintext, err := json.Marshal(payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal payload"})
			return
		}

		encrypted, err := aes.Encrypt(plaintext)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "encryption failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"service":    "CAMPUS_ACADEMIC_FEED",
			"version":    "1.0",
			"timestamp":  time.Now().Unix(),
			"iv":         encrypted.IV,
			"ciphertext": encrypted.Ciphertext,
			"tag":        encrypted.Tag,
		})
	})

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Mock Campus API running on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// buildFallbackPayload constructs a minimal fallback academic data payload.
func buildFallbackPayload() map[string]interface{} {
	return map[string]interface{}{
		"sync_timestamp": time.Now().UTC().Format(time.RFC3339),
		"academic_year":  "2026/2027",
		"semester_type":  1,
		"faculties": []map[string]interface{}{
			{
				"code": "FT",
				"name": "Fakultas Teknik",
				"study_programs": []map[string]interface{}{
					{"code": "INF", "name": "S1 Informatika"},
				},
			},
		},
		"buildings": []map[string]interface{}{
			{
				"code": "GEDUNG-B",
				"name": "Gedung Kuliah Terpadu",
				"rooms": []map[string]interface{}{
					{
						"room_code":      "LAB-KOM-1",
						"name":           "Laboratorium Komputer 1",
						"latitude":       5.201452,
						"longitude":      96.702145,
						"radius_meters":  35,
					},
				},
			},
		},
		"users": []map[string]interface{}{
			{
				"external_id": "23552011001",
				"name":        "Ahmad Fauzi",
				"email":       "ahmad.fauzi@kampus.ac.id",
				"role":        "mahasiswa",
				"prodi_code":  "INF",
			},
			{
				"external_id": "198801102015041001",
				"name":        "Dr. Irwan Setiawan, M.Kom.",
				"email":       "irwan.s@kampus.ac.id",
				"role":        "dosen",
				"prodi_code":  "INF",
			},
		},
		"schedules": []map[string]interface{}{
			{
				"external_schedule_id": "SCH-2026-INF-001",
				"course_code":          "INF301",
				"course_name":          "Pemrograman Sistem Terdistribusi",
				"lecturer_id":          "198801102015041001",
				"room_code":            "LAB-KOM-1",
				"day_of_week":          2,
				"start_time":           "08:00:00",
				"end_time":             "10:30:00",
				"enrolled_students":    []string{"23552011001"},
			},
		},
	}
}
