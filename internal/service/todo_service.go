package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/arezoohaghighi/todo-api/internal/model"
	"github.com/arezoohaghighi/todo-api/internal/repository"
)

type TodoService struct {
	todoRepo repository.TodoRepository
}

func NewTodoService(todoRepo repository.TodoRepository) *TodoService {
	return &TodoService{todoRepo: todoRepo}
}

func (s *TodoService) Create(ctx context.Context, userID string, text string, dueDate *time.Time) (*model.Todo, error) {
	if text == "" || len(text) > 500 {
		return nil, ErrInvalidRequest
	}
	trimmed := strings.TrimSpace(text)
	if len(trimmed) == 0 {
		return nil, ErrInvalidRequest
	}

	todo := model.Todo{
		UserID:    userID,
		Text:      trimmed,
		Completed: false,
		DueDate:   dueDate,
	}

	if err := s.todoRepo.Create(ctx, &todo); err != nil {
		return nil, err
	}
	return &todo, nil
}

func (s *TodoService) List(ctx context.Context, userID string, filters repository.ListFilter) ([]model.Todo, int64, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 || filters.Limit > 100 {
		filters.Limit = 20
	}
	return s.todoRepo.List(ctx, userID, filters)
}

func (s *TodoService) Get(ctx context.Context, userID string, id int64) (*model.Todo, error) {
	todo, err := s.todoRepo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return todo, nil
}

func (s *TodoService) Update(ctx context.Context, userID string, id int64, text *string, completed *bool, dueDate *time.Time, dueDateSet bool) (*model.Todo, error) {
	existing, err := s.todoRepo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	newText := existing.Text
	if text != nil {
		newText = strings.TrimSpace(*text)
		if newText == "" || len(newText) > 500 {
			return nil, ErrInvalidRequest
		}
	}

	todo := model.Todo{
		ID:        id,
		UserID:    userID,
		Text:      newText,
		Completed: existing.Completed,
		DueDate:   existing.DueDate,
		CreatedAt: existing.CreatedAt,
		UpdatedAt: existing.UpdatedAt,
	}

	if completed != nil {
		todo.Completed = *completed
	}
	if dueDateSet {
		todo.DueDate = dueDate
	}

	if err := s.todoRepo.Update(ctx, &todo); err != nil {
		return nil, err
	}
	return &todo, nil
}

func (s *TodoService) Delete(ctx context.Context, userID string, id int64) error {
	if err := s.todoRepo.Delete(ctx, id, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
