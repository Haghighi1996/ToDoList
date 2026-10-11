package model

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	RevokedAt time.Time `json:"revoked_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Session) IsValid(now time.Time) bool {
	if !s.RevokedAt.IsZero() {
		return false
	}
	return s.ExpiresAt.After(now)
}
