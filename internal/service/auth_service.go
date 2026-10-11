package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/arezoohaghighi/todo-api/internal/auth"
	"github.com/arezoohaghighi/todo-api/internal/model"
	"github.com/arezoohaghighi/todo-api/internal/repository"
)

type AuthService struct {
	db          *sql.DB
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
}

func NewAuthService(db *sql.DB, userRepo repository.UserRepository, sessionRepo repository.SessionRepository) *AuthService {
	return &AuthService{
		db:          db,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (s *AuthService) CreateAnonymousUser(ctx context.Context, deviceID string) (string, string, time.Time, error) {
	deviceID = strings.TrimSpace(deviceID)
	if len(deviceID) == 0 || len(deviceID) > 64 {
		return "", "", time.Time{}, ErrInvalidRequest
	}
	user, err := s.userRepo.GetByDeviceID(ctx, deviceID)
	if err == nil {
		return s.createSession(ctx, user)
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return "", "", time.Time{}, err
	}

	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return "", "", time.Time{}, err
		}

		var newUser model.User
		err = tx.QueryRowContext(ctx,
			`INSERT INTO users (device_id) VALUES ($1) RETURNING id, device_id, created_at`,
			deviceID,
		).Scan(&newUser.ID, &newUser.DeviceID, &newUser.CreatedAt)
		if err != nil {
			tx.Rollback()
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				user, err := s.userRepo.GetByDeviceID(ctx, deviceID)
				if err != nil {
					return "", "", time.Time{}, err
				}
				return s.createSession(ctx, user)
			}
			return "", "", time.Time{}, err
		}

		token, err := auth.GenerateToken()
		if err != nil {
			tx.Rollback()
			return "", "", time.Time{}, err
		}

		tokenHash := auth.HashToken(token)
		expiresAt := time.Now().UTC().Add(24 * time.Hour)

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at) VALUES (gen_random_uuid(), $1, $2, $3, NOW())`,
			newUser.ID, tokenHash, expiresAt,
		); err != nil {
			tx.Rollback()
			return "", "", time.Time{}, err
		}

		if err := tx.Commit(); err != nil {
			return "", "", time.Time{}, err
		}

		return token, newUser.ID.String(), expiresAt, nil
	}

	return "", "", time.Time{}, fmt.Errorf("failed to create user after %d retries", maxRetries)
}

func (s *AuthService) createSession(ctx context.Context, user *model.User) (string, string, time.Time, error) {
	token, err := auth.GenerateToken()
	if err != nil {
		return "", "", time.Time{}, err
	}

	tokenHash := auth.HashToken(token)
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at) VALUES (gen_random_uuid(), $1, $2, $3, NOW())`,
		user.ID, tokenHash, expiresAt,
	)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return token, user.ID.String(), expiresAt, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	tokenHash := auth.HashToken(token)
	session, err := s.sessionRepo.FindActiveByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUnauthorized
		}
		return err
	}
	return s.sessionRepo.Revoke(ctx, session.ID.String())
}
