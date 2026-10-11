package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/arezoohaghighi/todo-api/internal/model"
	"github.com/arezoohaghighi/todo-api/internal/repository"
)

type fakeTodoRepository struct {
	todo *model.Todo
}

func (r *fakeTodoRepository) Create(_ context.Context, todo *model.Todo) error {
	todo.ID = 1
	r.todo = todo
	return nil
}
func (r *fakeTodoRepository) GetByID(_ context.Context, _ int64, _ string) (*model.Todo, error) {
	if r.todo == nil {
		return nil, repository.ErrNotFound
	}
	copy := *r.todo
	return &copy, nil
}
func (r *fakeTodoRepository) List(_ context.Context, _ string, _ repository.ListFilter) ([]model.Todo, int64, error) {
	return nil, 0, nil
}
func (r *fakeTodoRepository) Update(_ context.Context, todo *model.Todo) error {
	r.todo = todo
	return nil
}
func (r *fakeTodoRepository) Delete(_ context.Context, _ int64, _ string) error {
	return errors.New("database unavailable")
}

func TestTodoServiceCreateTrimsText(t *testing.T) {
	repo := &fakeTodoRepository{}
	service := NewTodoService(repo)
	todo, err := service.Create(context.Background(), "user", "  write tests  ", nil)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if todo.Text != "write tests" {
		t.Fatalf("text = %q, want trimmed text", todo.Text)
	}
}

func TestTodoServiceCreateRejectsWhitespace(t *testing.T) {
	service := NewTodoService(&fakeTodoRepository{})
	_, err := service.Create(context.Background(), "user", "   ", nil)
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("error = %v, want invalid request", err)
	}
}

func TestTodoServiceUpdateClearsDueDate(t *testing.T) {
	dueDate := time.Now().Add(24 * time.Hour)
	repo := &fakeTodoRepository{todo: &model.Todo{ID: 1, UserID: "user", Text: "task", DueDate: &dueDate}}
	service := NewTodoService(repo)
	todo, err := service.Update(context.Background(), "user", 1, nil, nil, nil, true)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if todo.DueDate != nil {
		t.Fatal("due date was not cleared")
	}
}

func TestTodoServiceDeletePreservesUnexpectedError(t *testing.T) {
	service := NewTodoService(&fakeTodoRepository{})
	err := service.Delete(context.Background(), "user", 1)
	if errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, unexpected errors must not become not found", err)
	}
}
