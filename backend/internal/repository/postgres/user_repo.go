package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilham/presensi-online/backend/internal/domain"
)

// UserRepo is the PostgreSQL implementation of domain.UserRepository.
type UserRepo struct {
	db *pgxpool.Pool
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

// Upsert inserts or updates a user by external_id.
// Used by the sync worker to ingest data from the Campus API.
func (r *UserRepo) Upsert(ctx context.Context, u *domain.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, external_id, name, email, password_hash, role, prodi_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (external_id) DO UPDATE SET
			name          = EXCLUDED.name,
			email         = EXCLUDED.email,
			password_hash = EXCLUDED.password_hash,
			role          = EXCLUDED.role,
			prodi_id      = EXCLUDED.prodi_id,
			is_active     = EXCLUDED.is_active,
			updated_at    = CURRENT_TIMESTAMP
	`, u.ID, u.ExternalID, u.Name, u.Email, u.PasswordHash, u.Role, u.ProdiID, u.IsActive)
	if err != nil {
		// Fallback: If email conflicts with an existing record, update that record by email
		tag, err2 := r.db.Exec(ctx, `
			UPDATE users SET
				external_id   = $1,
				name          = $2,
				password_hash = $3,
				role          = $4,
				prodi_id      = $5,
				is_active     = $6,
				updated_at    = CURRENT_TIMESTAMP
			WHERE email = $7
		`, u.ExternalID, u.Name, u.PasswordHash, u.Role, u.ProdiID, u.IsActive, u.Email)
		if err2 == nil && tag.RowsAffected() > 0 {
			return nil
		}
		return fmt.Errorf("UserRepo.Upsert: %w", err)
	}
	return nil
}

// FindByID retrieves a user by their UUID.
func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, external_id, name, email, password_hash, role, prodi_id,
		       device_id, avatar_url, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, id)

	return scanUser(row)
}

// FindByEmail retrieves a user by email address (used for login).
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, external_id, name, email, password_hash, role, prodi_id,
		       device_id, avatar_url, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`, email)

	return scanUser(row)
}

// FindByExternalID retrieves a user by their NIM or NIDN.
func (r *UserRepo) FindByExternalID(ctx context.Context, externalID string) (*domain.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, external_id, name, email, password_hash, role, prodi_id,
		       device_id, avatar_url, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE external_id = $1 AND deleted_at IS NULL
	`, externalID)

	return scanUser(row)
}

// UpdateDeviceID binds a device UUID to a user account (device binding).
func (r *UserRepo) UpdateDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE users SET device_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND deleted_at IS NULL
	`, deviceID, userID)
	if err != nil {
		return fmt.Errorf("UserRepo.UpdateDeviceID: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdatePassword updates user password hash.
func (r *UserRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND deleted_at IS NULL
	`, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("UserRepo.UpdatePassword: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateAvatar updates user avatar photo URL or base64 data URI.
func (r *UserRepo) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE users SET avatar_url = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND deleted_at IS NULL
	`, avatarURL, userID)
	if err != nil {
		return fmt.Errorf("UserRepo.UpdateAvatar: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetProfile retrieves user profile enriched with study program and faculty names.
func (r *UserRepo) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error) {
	row := r.db.QueryRow(ctx, `
		SELECT u.id, u.external_id, u.name, u.email, u.role, u.prodi_id,
		       COALESCE(sp.name, '') AS prodi_name,
		       COALESCE(f.name, '') AS faculty_name,
		       u.avatar_url, u.is_active
		FROM users u
		LEFT JOIN study_programs sp ON u.prodi_id = sp.id
		LEFT JOIN faculties f ON sp.faculty_id = f.id
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`, userID)

	p := &domain.UserProfile{}
	err := row.Scan(
		&p.ID, &p.ExternalID, &p.Name, &p.Email, &p.Role, &p.ProdiID,
		&p.ProdiName, &p.FacultyName, &p.AvatarURL, &p.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("UserRepo.GetProfile: %w", err)
	}
	return p, nil
}

// scanUser scans a pgx.Row into a domain.User.
func scanUser(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}
	err := row.Scan(
		&u.ID, &u.ExternalID, &u.Name, &u.Email, &u.PasswordHash, &u.Role,
		&u.ProdiID, &u.DeviceID, &u.AvatarURL, &u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanUser: %w", err)
	}
	return u, nil
}
