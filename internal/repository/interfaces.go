package repository

import (
	"context"

	"github.com/arezoohaghighi/todo-api/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByDeviceID(ctx context.Context, deviceID string) (*model.User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	FindActiveByTokenHash(ctx context.Context, tokenHash string) (*model.Session, error)
	Revoke(ctx context.Context, sessionID string) error
}

type TodoRepository interface {
	Create(ctx context.Context, todo *model.Todo) error
	GetByID(ctx context.Context, id int64, userID string) (*model.Todo, error)
	List(ctx context.Context, userID string, filters ListFilter) ([]model.Todo, int64, error)
	Update(ctx context.Context, todo *model.Todo) error
	Delete(ctx context.Context, id int64, userID string) error
}

type ListFilter struct {
	Completed *bool
	Overdue   *bool
	Sort      string
	Page      int
	Limit     int
}
