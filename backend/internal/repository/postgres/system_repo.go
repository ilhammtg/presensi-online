package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilham/presensi-online/backend/internal/domain"
)

// SystemRepo is the PostgreSQL implementation of domain.SystemAdminRepository.
type SystemRepo struct {
	db *pgxpool.Pool
}

// NewSystemRepo creates a new SystemRepo.
func NewSystemRepo(db *pgxpool.Pool) *SystemRepo {
	return &SystemRepo{db: db}
}

// GetSystemStats computes system-wide global university statistics.
func (r *SystemRepo) GetSystemStats(ctx context.Context) (*domain.SystemStats, error) {
	stats := &domain.SystemStats{}

	query := `
		SELECT 
			(SELECT COUNT(*) FROM faculties),
			(SELECT COUNT(*) FROM study_programs),
			(SELECT COUNT(*) FROM users WHERE role = 'dosen' AND is_active = true),
			(SELECT COUNT(*) FROM users WHERE role = 'mahasiswa' AND is_active = true),
			(SELECT COUNT(*) FROM class_schedules WHERE is_active = true),
			(SELECT COUNT(*) FROM class_sessions),
			(SELECT COALESCE(ROUND((COUNT(CASE WHEN a.status IN ('hadir', 'terlambat') THEN 1 END)::numeric / NULLIF(COUNT(a.id), 0)) * 100, 1), 0.0) 
			 FROM attendances a)
	`

	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalFaculties,
		&stats.TotalStudyPrograms,
		&stats.TotalLecturers,
		&stats.TotalStudents,
		&stats.TotalSchedules,
		&stats.TotalSessionsHeld,
		&stats.GlobalAvgAttendance,
	)
	if err != nil {
		return nil, fmt.Errorf("SystemRepo.GetSystemStats: %w", err)
	}

	return stats, nil
}

// GetAllStudyPrograms returns all study programs across all faculties.
func (r *SystemRepo) GetAllStudyPrograms(ctx context.Context) ([]*domain.StudyProgramDetail, error) {
	query := `
		SELECT sp.id, sp.code, sp.name, f.id, f.name
		FROM study_programs sp
		JOIN faculties f ON sp.faculty_id = f.id
		ORDER BY f.name ASC, sp.name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("SystemRepo.GetAllStudyPrograms: %w", err)
	}
	defer rows.Close()

	var results []*domain.StudyProgramDetail
	for rows.Next() {
		sp := &domain.StudyProgramDetail{}
		if err := rows.Scan(&sp.ID, &sp.Code, &sp.Name, &sp.FacultyID, &sp.FacultyName); err != nil {
			return nil, fmt.Errorf("SystemRepo.GetAllStudyPrograms scan: %w", err)
		}
		results = append(results, sp)
	}

	return results, nil
}

// GetUsers retrieves users with optional role and prodi filters.
func (r *SystemRepo) GetUsers(ctx context.Context, role *domain.UserRole, prodiID *uuid.UUID) ([]*domain.SystemUserItem, error) {
	query := `
		SELECT 
			u.id, u.external_id, u.name, u.email, u.role, sp.name, u.device_id, u.is_active, u.created_at
		FROM users u
		LEFT JOIN study_programs sp ON u.prodi_id = sp.id
		WHERE u.deleted_at IS NULL
	`

	var args []interface{}
	idx := 1

	if role != nil {
		query += fmt.Sprintf(" AND u.role = $%d", idx)
		args = append(args, *role)
		idx++
	}

	if prodiID != nil {
		query += fmt.Sprintf(" AND u.prodi_id = $%d", idx)
		args = append(args, *prodiID)
		idx++
	}

	query += " ORDER BY u.role ASC, u.name ASC LIMIT 100"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("SystemRepo.GetUsers: %w", err)
	}
	defer rows.Close()

	var results []*domain.SystemUserItem
	for rows.Next() {
		user := &domain.SystemUserItem{}
		var prodiName *string
		var deviceID *string

		if err := rows.Scan(
			&user.ID,
			&user.ExternalID,
			&user.Name,
			&user.Email,
			&user.Role,
			&prodiName,
			&deviceID,
			&user.IsActive,
			&user.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("SystemRepo.GetUsers scan: %w", err)
		}
		user.ProdiName = prodiName
		user.DeviceID = deviceID
		results = append(results, user)
	}

	return results, nil
}

// GetThemeSettings loads the branding and theme configuration from system_settings.
func (r *SystemRepo) GetThemeSettings(ctx context.Context) (*domain.SystemThemeSettings, error) {
	settings := &domain.SystemThemeSettings{
		PrimaryColor:  "#006633",
		AccentColor:   "#D4AF37",
		CampusName:    "Universitas Almuslim",
		CampusTagline: "Sistem Informasi Presensi & QR Dinamis",
		LogoURL:       "",
	}

	rows, err := r.db.Query(ctx, `SELECT key, value FROM system_settings`)
	if err != nil {
		return settings, nil // Return defaults if table or rows not yet ready
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err == nil {
			switch k {
			case "primary_color":
				settings.PrimaryColor = v
			case "accent_color":
				settings.AccentColor = v
			case "campus_name":
				settings.CampusName = v
			case "campus_tagline":
				settings.CampusTagline = v
			case "logo_url":
				settings.LogoURL = v
			}
		}
	}

	return settings, nil
}

// SaveThemeSettings persists the branding and theme configuration.
func (r *SystemRepo) SaveThemeSettings(ctx context.Context, s *domain.SystemThemeSettings) error {
	queries := map[string]string{
		"primary_color":  s.PrimaryColor,
		"accent_color":   s.AccentColor,
		"campus_name":    s.CampusName,
		"campus_tagline": s.CampusTagline,
		"logo_url":       s.LogoURL,
	}

	for k, v := range queries {
		_, err := r.db.Exec(ctx, `
			INSERT INTO system_settings (key, value, updated_at)
			VALUES ($1, $2, CURRENT_TIMESTAMP)
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP
		`, k, v)
		if err != nil {
			return fmt.Errorf("SystemRepo.SaveThemeSettings (%s): %w", k, err)
		}
	}

	return nil
}

// GetCampusApiConfig loads the dynamic SIAKAD API connection settings.
func (r *SystemRepo) GetCampusApiConfig(ctx context.Context) (*domain.CampusApiConfig, error) {
	cfg := &domain.CampusApiConfig{
		ApiURL:        "http://mock-campus-api:9001",
		ApiKey:        "mock-api-key-change-me",
		SyncCron:      "0 */6 * * *",
		CampusName:    "Universitas Almuslim",
		CampusTagline: "Fakultas Ilmu Komputer (FIKOM)",
		GeofenceMode:  "multi_point",
		MaxTolerance:  15,
	}

	rows, err := r.db.Query(ctx, `SELECT key, value FROM system_settings`)
	if err != nil {
		return cfg, nil
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err == nil {
			switch k {
			case "campus_api_url":
				cfg.ApiURL = v
			case "campus_api_key":
				cfg.ApiKey = v
			case "campus_sync_cron":
				cfg.SyncCron = v
			case "campus_name":
				cfg.CampusName = v
			case "campus_tagline":
				cfg.CampusTagline = v
			case "geofence_mode":
				cfg.GeofenceMode = v
			case "max_tolerance_minutes":
				var tol int
				if _, err := fmt.Sscanf(v, "%d", &tol); err == nil && tol > 0 {
					cfg.MaxTolerance = tol
				}
			}
		}
	}

	return cfg, nil
}

// SaveCampusApiConfig persists the dynamic SIAKAD API connection settings.
func (r *SystemRepo) SaveCampusApiConfig(ctx context.Context, cfg *domain.CampusApiConfig) error {
	queries := map[string]string{
		"campus_api_url":        cfg.ApiURL,
		"campus_api_key":        cfg.ApiKey,
		"campus_sync_cron":      cfg.SyncCron,
		"campus_name":           cfg.CampusName,
		"campus_tagline":        cfg.CampusTagline,
		"geofence_mode":         cfg.GeofenceMode,
		"max_tolerance_minutes": fmt.Sprintf("%d", cfg.MaxTolerance),
	}

	for k, v := range queries {
		_, err := r.db.Exec(ctx, `
			INSERT INTO system_settings (key, value, updated_at)
			VALUES ($1, $2, CURRENT_TIMESTAMP)
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP
		`, k, v)
		if err != nil {
			return fmt.Errorf("SystemRepo.SaveCampusApiConfig (%s): %w", k, err)
		}
	}

	return nil
}

// GetAllCampusLocations returns all campus geofence points.
func (r *SystemRepo) GetAllCampusLocations(ctx context.Context) ([]*domain.CampusLocation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, COALESCE(description, ''), latitude, longitude, radius_meters, is_active, created_at, updated_at
		FROM campus_locations
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("SystemRepo.GetAllCampusLocations: %w", err)
	}
	defer rows.Close()

	var list []*domain.CampusLocation
	for rows.Next() {
		loc := &domain.CampusLocation{}
		if err := rows.Scan(
			&loc.ID, &loc.Name, &loc.Description, &loc.Latitude, &loc.Longitude,
			&loc.RadiusMeters, &loc.IsActive, &loc.CreatedAt, &loc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("SystemRepo.GetAllCampusLocations scan: %w", err)
		}
		list = append(list, loc)
	}
	return list, nil
}

// GetActiveCampusLocations returns only active campus geofence points.
func (r *SystemRepo) GetActiveCampusLocations(ctx context.Context) ([]*domain.CampusLocation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, COALESCE(description, ''), latitude, longitude, radius_meters, is_active, created_at, updated_at
		FROM campus_locations
		WHERE is_active = TRUE
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("SystemRepo.GetActiveCampusLocations: %w", err)
	}
	defer rows.Close()

	var list []*domain.CampusLocation
	for rows.Next() {
		loc := &domain.CampusLocation{}
		if err := rows.Scan(
			&loc.ID, &loc.Name, &loc.Description, &loc.Latitude, &loc.Longitude,
			&loc.RadiusMeters, &loc.IsActive, &loc.CreatedAt, &loc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("SystemRepo.GetActiveCampusLocations scan: %w", err)
		}
		list = append(list, loc)
	}
	return list, nil
}

// CreateCampusLocation inserts a new campus geofence point.
func (r *SystemRepo) CreateCampusLocation(ctx context.Context, loc *domain.CampusLocation) error {
	if loc.ID == uuid.Nil {
		loc.ID = uuid.New()
	}
	if loc.RadiusMeters <= 0 {
		loc.RadiusMeters = 80
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO campus_locations (id, name, description, latitude, longitude, radius_meters, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, loc.ID, loc.Name, loc.Description, loc.Latitude, loc.Longitude, loc.RadiusMeters, loc.IsActive)
	if err != nil {
		return fmt.Errorf("SystemRepo.CreateCampusLocation: %w", err)
	}
	return nil
}

// UpdateCampusLocation updates an existing campus geofence point.
func (r *SystemRepo) UpdateCampusLocation(ctx context.Context, loc *domain.CampusLocation) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE campus_locations
		SET name = $1, description = $2, latitude = $3, longitude = $4, radius_meters = $5, is_active = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
	`, loc.Name, loc.Description, loc.Latitude, loc.Longitude, loc.RadiusMeters, loc.IsActive, loc.ID)
	if err != nil {
		return fmt.Errorf("SystemRepo.UpdateCampusLocation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// DeleteCampusLocation deletes a campus geofence point.
func (r *SystemRepo) DeleteCampusLocation(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM campus_locations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("SystemRepo.DeleteCampusLocation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
