// Package worker contains the campus data sync worker.
// It ingests encrypted payloads from the Mock Campus API and UPSERTs to PostgreSQL.
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/ilham/presensi-online/backend/internal/config"
	"github.com/ilham/presensi-online/backend/internal/crypto"
	"github.com/ilham/presensi-online/backend/internal/domain"
)

// CampusPayload is the raw (decrypted) data structure from the Campus API.
// This matches the "Simulasi Data Mentah" format in docs.md.
type CampusPayload struct {
	SyncTimestamp string          `json:"sync_timestamp"`
	AcademicYear  string          `json:"academic_year"`
	SemesterType  int             `json:"semester_type"`
	Faculties     []FacultyData   `json:"faculties"`
	Buildings     []BuildingData  `json:"buildings"`
	Users         []UserData      `json:"users"`
	Schedules     []ScheduleData  `json:"schedules"`
}

type FacultyData struct {
	Code          string              `json:"code"`
	Name          string              `json:"name"`
	StudyPrograms []StudyProgramData  `json:"study_programs"`
}

type StudyProgramData struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type BuildingData struct {
	Code  string     `json:"code"`
	Name  string     `json:"name"`
	Rooms []RoomData `json:"rooms"`
}

type RoomData struct {
	RoomCode      string  `json:"room_code"`
	Name          string  `json:"name"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	RadiusMeters  int     `json:"radius_meters"`
}

type UserData struct {
	ExternalID string `json:"external_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	ProdiCode  string `json:"prodi_code"`
	Password   string `json:"password,omitempty"`
}

type ScheduleData struct {
	ExternalScheduleID string   `json:"external_schedule_id"`
	CourseCode         string   `json:"course_code"`
	CourseName         string   `json:"course_name"`
	LecturerID         string   `json:"lecturer_id"` // external_id (NIDN)
	RoomCode           string   `json:"room_code"`
	DayOfWeek          int      `json:"day_of_week"`
	StartTime          string   `json:"start_time"`
	EndTime            string   `json:"end_time"`
	EnrolledStudents   []string `json:"enrolled_students"` // list of external_id (NIM)
}

// EncryptedAPIResponse matches the "Format Amplop API" in docs.md.
type EncryptedAPIResponse struct {
	Service    string `json:"service"`
	Version    string `json:"version"`
	Timestamp  int64  `json:"timestamp"`
	IV         string `json:"iv"`
	Ciphertext string `json:"ciphertext"`
	Tag        string `json:"tag"`
}

// Repositories needed by the sync worker.
type Repos struct {
	Faculty      domain.FacultyRepository
	StudyProgram domain.StudyProgramRepository
	Building     domain.BuildingRepository
	Room         domain.RoomRepository
	User         domain.UserRepository
	Schedule     domain.ClassScheduleRepository
	StudyPlan    domain.StudyPlanRepository
}

// SyncWorker fetches encrypted data from the Mock Campus API and syncs to PostgreSQL.
type SyncWorker struct {
	cfg    *config.Config
	aes    *crypto.AESGCM
	repos  Repos
	logger *zap.Logger
	client *http.Client
}

// NewSyncWorker creates a new SyncWorker.
func NewSyncWorker(cfg *config.Config, repos Repos, logger *zap.Logger) (*SyncWorker, error) {
	aes, err := crypto.NewAESGCM(cfg.AES.KeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to init AES: %w", err)
	}

	return &SyncWorker{
		cfg:    cfg,
		aes:    aes,
		repos:  repos,
		logger: logger,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Run executes a single sync cycle: fetch → decrypt → validate → upsert.
func (w *SyncWorker) Run(ctx context.Context) error {
	w.logger.Info("sync worker: starting sync cycle")
	start := time.Now()

	// 1. Fetch encrypted payload from Mock Campus API
	raw, err := w.fetchPayload(ctx)
	if err != nil {
		return fmt.Errorf("fetch failed: %w", err)
	}

	// 2. Decrypt using AES-256-GCM
	plaintext, err := w.aes.Decrypt(&crypto.EncryptedPayload{
		IV:         raw.IV,
		Ciphertext: raw.Ciphertext,
		Tag:        raw.Tag,
	})
	if err != nil {
		return fmt.Errorf("decryption failed (auth tag mismatch?): %w", err)
	}

	// 3. Parse JSON payload
	var payload CampusPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return fmt.Errorf("failed to parse payload: %w", err)
	}

	// 4. UPSERT all data
	if err := w.upsertAll(ctx, payload); err != nil {
		return fmt.Errorf("upsert failed: %w", err)
	}

	w.logger.Info("sync worker: cycle complete",
		zap.Duration("duration", time.Since(start)),
		zap.String("sync_timestamp", payload.SyncTimestamp),
	)
	return nil
}

func (w *SyncWorker) fetchPayload(ctx context.Context) (*EncryptedAPIResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		w.cfg.MockAPI.URL+"/api/v1/sync/payload", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", w.cfg.MockAPI.APIKey)

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mock API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result EncryptedAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (w *SyncWorker) upsertAll(ctx context.Context, payload CampusPayload) error {
	// Map prodi_code → UUID for schedule linking
	prodiCodeToID := make(map[string]uuid.UUID)
	roomCodeToID := make(map[string]uuid.UUID)
	userExtIDToID := make(map[string]uuid.UUID)

	// 1. Faculties + Study Programs
	for _, fData := range payload.Faculties {
		f := &domain.Faculty{
			ID:   uuid.New(),
			Code: fData.Code,
			Name: fData.Name,
		}
		if err := w.repos.Faculty.Upsert(ctx, f); err != nil {
			return fmt.Errorf("faculty upsert [%s]: %w", f.Code, err)
		}

		// Re-fetch to get actual UUID (in case it was an existing record)
		existing, _ := w.repos.Faculty.FindByCode(ctx, fData.Code)
		if existing != nil {
			f = existing
		}

		for _, spData := range fData.StudyPrograms {
			sp := &domain.StudyProgram{
				ID:        uuid.New(),
				FacultyID: f.ID,
				Code:      spData.Code,
				Name:      spData.Name,
			}
			if err := w.repos.StudyProgram.Upsert(ctx, sp); err != nil {
				return fmt.Errorf("study program upsert [%s]: %w", sp.Code, err)
			}
			existing, _ := w.repos.StudyProgram.FindByCode(ctx, spData.Code)
			if existing != nil {
				prodiCodeToID[spData.Code] = existing.ID
			}
		}
	}

	// 2. Buildings + Rooms
	for _, bData := range payload.Buildings {
		b := &domain.Building{
			ID:   uuid.New(),
			Code: bData.Code,
			Name: bData.Name,
		}
		if err := w.repos.Building.Upsert(ctx, b); err != nil {
			return fmt.Errorf("building upsert [%s]: %w", b.Code, err)
		}

		existing, _ := w.repos.Building.FindByCode(ctx, bData.Code)
		if existing != nil {
			b = existing
		}

		for _, rData := range bData.Rooms {
			r := &domain.Room{
				ID:           uuid.New(),
				BuildingID:   b.ID,
				RoomCode:     rData.RoomCode,
				Name:         rData.Name,
				Latitude:     rData.Latitude,
				Longitude:    rData.Longitude,
				RadiusMeters: rData.RadiusMeters,
				IsActive:     true,
			}
			if err := w.repos.Room.Upsert(ctx, r); err != nil {
				return fmt.Errorf("room upsert [%s]: %w", r.RoomCode, err)
			}
			existingRoom, _ := w.repos.Room.FindByCode(ctx, rData.RoomCode)
			if existingRoom != nil {
				roomCodeToID[rData.RoomCode] = existingRoom.ID
			}
		}
	}

	// 3. Users
	for _, uData := range payload.Users {
		prodiID, _ := prodiCodeToID[uData.ProdiCode]

		// Default password: "password" (or specified in payload)
		plainPass := uData.Password
		if plainPass == "" {
			plainPass = "password"
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(plainPass), bcrypt.DefaultCost)

		u := &domain.User{
			ID:           uuid.New(),
			ExternalID:   uData.ExternalID,
			Name:         uData.Name,
			Email:        uData.Email,
			PasswordHash: string(hash),
			Role:         domain.UserRole(uData.Role),
			ProdiID:      &prodiID,
			IsActive:     true,
		}
		if err := w.repos.User.Upsert(ctx, u); err != nil {
			return fmt.Errorf("user upsert [%s]: %w", u.ExternalID, err)
		}
		existingUser, _ := w.repos.User.FindByExternalID(ctx, uData.ExternalID)
		if existingUser != nil {
			userExtIDToID[uData.ExternalID] = existingUser.ID
		}
	}

	// 4. Schedules + Study Plans (KRS enrollment)
	for _, sData := range payload.Schedules {
		lecturerID, ok := userExtIDToID[sData.LecturerID]
		if !ok {
			w.logger.Warn("lecturer not found, skipping schedule", zap.String("lecturer_id", sData.LecturerID))
			continue
		}
		roomID, ok := roomCodeToID[sData.RoomCode]
		if !ok {
			w.logger.Warn("room not found, skipping schedule", zap.String("room_code", sData.RoomCode))
			continue
		}

		extID := sData.ExternalScheduleID
		cs := &domain.ClassSchedule{
			ID:           uuid.New(),
			ExternalID:   &extID,
			CourseCode:   sData.CourseCode,
			CourseName:   sData.CourseName,
			AcademicYear: payload.AcademicYear,
			SemesterType: payload.SemesterType,
			LecturerID:   lecturerID,
			RoomID:       roomID,
			DayOfWeek:    sData.DayOfWeek,
			StartTime:    sData.StartTime,
			EndTime:      sData.EndTime,
			IsActive:     true,
		}
		if err := w.repos.Schedule.Upsert(ctx, cs); err != nil {
			return fmt.Errorf("schedule upsert [%s]: %w", *cs.ExternalID, err)
		}

		existingSchedule, _ := w.repos.Schedule.FindByExternalID(ctx, sData.ExternalScheduleID)
		if existingSchedule == nil {
			continue
		}

		// Upsert KRS enrollments
		for _, studentExtID := range sData.EnrolledStudents {
			studentID, ok := userExtIDToID[studentExtID]
			if !ok {
				w.logger.Warn("student not found for enrollment", zap.String("student_id", studentExtID))
				continue
			}
			sp := &domain.StudyPlan{
				ID:         uuid.New(),
				StudentID:  studentID,
				ScheduleID: existingSchedule.ID,
				Status:     domain.EnrollmentActive,
			}
			if err := w.repos.StudyPlan.Upsert(ctx, sp); err != nil {
				return fmt.Errorf("study plan upsert [%s→%s]: %w", studentExtID, sData.ExternalScheduleID, err)
			}
		}
	}

	return nil
}
