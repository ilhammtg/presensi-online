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

// ClassSessionRepo is the PostgreSQL implementation of domain.ClassSessionRepository.
type ClassSessionRepo struct {
	db *pgxpool.Pool
}

// NewClassSessionRepo creates a new ClassSessionRepo.
func NewClassSessionRepo(db *pgxpool.Pool) *ClassSessionRepo {
	return &ClassSessionRepo{db: db}
}

// Create inserts a new class session (called when lecturer opens attendance).
func (r *ClassSessionRepo) Create(ctx context.Context, cs *domain.ClassSession) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO class_sessions (id, schedule_id, meeting_no, session_date, qr_seed, is_open, opened_at, duration_minutes, expires_at)
		VALUES ($1, $2, $3, $4, $5, TRUE, CURRENT_TIMESTAMP, $6, $7)
	`, cs.ID, cs.ScheduleID, cs.MeetingNo, cs.SessionDate, cs.QRSeed, cs.DurationMinutes, cs.ExpiresAt)
	if err != nil {
		return fmt.Errorf("ClassSessionRepo.Create: %w", err)
	}
	return nil
}

// FindByID retrieves a class session by its UUID.
func (r *ClassSessionRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, schedule_id, meeting_no, session_date, qr_seed, is_open,
		       opened_at, closed_at, duration_minutes, expires_at, bap_topic, created_at
		FROM class_sessions WHERE id = $1
	`, id)
	return scanSession(row)
}

// FindActiveByScheduleID retrieves the currently open session for a given schedule.
// Uses the partial index idx_sessions_active_lookup for high performance.
func (r *ClassSessionRepo) FindActiveByScheduleID(ctx context.Context, scheduleID uuid.UUID) (*domain.ClassSession, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, schedule_id, meeting_no, session_date, qr_seed, is_open,
		       opened_at, closed_at, duration_minutes, expires_at, bap_topic, created_at
		FROM class_sessions
		WHERE schedule_id = $1 AND is_open = TRUE
		LIMIT 1
	`, scheduleID)
	return scanSession(row)
}

// FindByScheduleID retrieves all sessions for a schedule ordered by meeting_no ASC.
func (r *ClassSessionRepo) FindByScheduleID(ctx context.Context, scheduleID uuid.UUID) ([]*domain.ClassSession, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, schedule_id, meeting_no, session_date, qr_seed, is_open,
		       opened_at, closed_at, duration_minutes, expires_at, bap_topic, created_at
		FROM class_sessions
		WHERE schedule_id = $1
		ORDER BY meeting_no ASC
	`, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("ClassSessionRepo.FindByScheduleID: %w", err)
	}
	defer rows.Close()

	var out []*domain.ClassSession
	for rows.Next() {
		cs := &domain.ClassSession{}
		if err := rows.Scan(
			&cs.ID, &cs.ScheduleID, &cs.MeetingNo, &cs.SessionDate, &cs.QRSeed,
			&cs.IsOpen, &cs.OpenedAt, &cs.ClosedAt, &cs.DurationMinutes, &cs.ExpiresAt,
			&cs.BAPTopic, &cs.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, cs)
	}
	return out, rows.Err()
}

// Close marks a session as closed and optionally stores the BAP topic.
func (r *ClassSessionRepo) Close(ctx context.Context, sessionID uuid.UUID, bapTopic *string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE class_sessions
		SET is_open = FALSE, closed_at = CURRENT_TIMESTAMP, bap_topic = $1
		WHERE id = $2 AND is_open = TRUE
	`, bapTopic, sessionID)
	if err != nil {
		return fmt.Errorf("ClassSessionRepo.Close: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanSession(row pgx.Row) (*domain.ClassSession, error) {
	cs := &domain.ClassSession{}
	err := row.Scan(
		&cs.ID, &cs.ScheduleID, &cs.MeetingNo, &cs.SessionDate, &cs.QRSeed,
		&cs.IsOpen, &cs.OpenedAt, &cs.ClosedAt, &cs.DurationMinutes, &cs.ExpiresAt,
		&cs.BAPTopic, &cs.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanSession: %w", err)
	}
	return cs, nil
}
