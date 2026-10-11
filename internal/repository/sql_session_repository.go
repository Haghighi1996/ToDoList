package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/arezoohaghighi/todo-api/internal/model"
)

type SQLSessionRepository struct {
	DB *sql.DB
}

func NewSessionRepository(db *sql.DB) *SQLSessionRepository {
	return &SQLSessionRepository{DB: db}
}

func (r *SQLSessionRepository) Create(ctx context.Context, session *model.Session) error {
	return r.DB.QueryRowContext(
		ctx,
		`INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at) VALUES ($1, $2, $3, $4, NOW()) RETURNING created_at`,
		session.ID, session.UserID, session.TokenHash, session.ExpiresAt,
	).Scan(&session.CreatedAt)
}

func (r *SQLSessionRepository) FindActiveByTokenHash(ctx context.Context, tokenHash string) (*model.Session, error) {
	var session model.Session
	var expiresAt time.Time
	var revokedAt sql.NullTime
	err := r.DB.QueryRowContext(
		ctx,
		`SELECT id, user_id, token_hash, expires_at, revoked_at, created_at FROM sessions WHERE token_hash = $1 AND expires_at > NOW() AND revoked_at IS NULL`,
		tokenHash,
	).Scan(&session.ID, &session.UserID, &session.TokenHash, &expiresAt, &revokedAt, &session.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	session.ExpiresAt = expiresAt
	if revokedAt.Valid {
		session.RevokedAt = revokedAt.Time
	}
	return &session, nil
}

func (r *SQLSessionRepository) Revoke(ctx context.Context, sessionID string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE sessions SET revoked_at = NOW() WHERE id = $1`, sessionID)
	return err
}
