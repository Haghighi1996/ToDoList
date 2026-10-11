package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	DeviceID  string    `json:"device_id"`
	CreatedAt time.Time `json:"created_at"`
}
