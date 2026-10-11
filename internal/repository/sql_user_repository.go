package repository

import (
	"context"
	"database/sql"

	"github.com/arezoohaghighi/todo-api/internal/model"
)

type SQLUserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *SQLUserRepository {
	return &SQLUserRepository{DB: db}
}

func (r *SQLUserRepository) Create(ctx context.Context, user *model.User) error {
	return r.DB.QueryRowContext(
		ctx,
		`INSERT INTO users (device_id) VALUES ($1) RETURNING id, created_at`,
		user.DeviceID,
	).Scan(&user.ID, &user.CreatedAt)
}

func (r *SQLUserRepository) GetByDeviceID(ctx context.Context, deviceID string) (*model.User, error) {
	var user model.User
	err := r.DB.QueryRowContext(
		ctx,
		`SELECT id, device_id, created_at FROM users WHERE device_id = $1`,
		deviceID,
	).Scan(&user.ID, &user.DeviceID, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
