package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/arezoohaghighi/todo-api/internal/model"
)

type SQLTodoRepository struct {
	DB *sql.DB
}

func NewTodoRepository(db *sql.DB) *SQLTodoRepository {
	return &SQLTodoRepository{DB: db}
}

func (r *SQLTodoRepository) Create(ctx context.Context, todo *model.Todo) error {
	return r.DB.QueryRowContext(
		ctx,
		`INSERT INTO todos (user_id, text, completed, due_date, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id, created_at, updated_at`,
		todo.UserID, todo.Text, todo.Completed, todo.DueDate,
	).Scan(&todo.ID, &todo.CreatedAt, &todo.UpdatedAt)
}

func (r *SQLTodoRepository) GetByID(ctx context.Context, id int64, userID string) (*model.Todo, error) {
	var todo model.Todo
	var dueDate sql.NullTime
	err := r.DB.QueryRowContext(
		ctx,
		`SELECT id, user_id, text, completed, due_date, created_at, updated_at FROM todos WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&todo.ID, &todo.UserID, &todo.Text, &todo.Completed, &dueDate, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if dueDate.Valid {
		t := dueDate.Time
		todo.DueDate = &t
	}
	return &todo, nil
}

func (r *SQLTodoRepository) List(ctx context.Context, userID string, filters ListFilter) ([]model.Todo, int64, error) {
	whereClause, args := buildTodoWhere(filters, userID)
	argIdx := len(args) + 1

	orderBy := "created_at DESC"
	switch filters.Sort {
	case "due_date_asc":
		orderBy = "due_date ASC NULLS LAST"
	case "due_date_desc":
		orderBy = "due_date DESC NULLS LAST"
	}

	limit := filters.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (filters.Page - 1) * limit

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM todos WHERE %s", whereClause)
	var total int64
	err := r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	listQuery := fmt.Sprintf(
		"SELECT id, user_id, text, completed, due_date, created_at, updated_at FROM todos WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d",
		whereClause, orderBy, argIdx, argIdx+1,
	)
	args = append(args, limit, offset)

	rows, err := r.DB.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var todos []model.Todo
	for rows.Next() {
		var todo model.Todo
		var dueDate sql.NullTime
		if err := rows.Scan(&todo.ID, &todo.UserID, &todo.Text, &todo.Completed, &dueDate, &todo.CreatedAt, &todo.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if dueDate.Valid {
			t := dueDate.Time
			todo.DueDate = &t
		}
		todos = append(todos, todo)
	}

	return todos, total, nil
}

func buildTodoWhere(filters ListFilter, userID string) (string, []interface{}) {
	where := "user_id = $1"
	args := []interface{}{userID}
	argIdx := 2

	if filters.Completed != nil {
		where += fmt.Sprintf(" AND completed = $%d", argIdx)
		args = append(args, *filters.Completed)
		argIdx++
	}

	if filters.Overdue != nil {
		if *filters.Overdue {
			where += " AND completed = false AND due_date IS NOT NULL AND due_date < NOW()"
		} else {
			where += " AND (completed = true OR due_date IS NULL OR due_date >= NOW())"
		}
	}

	return where, args
}

func (r *SQLTodoRepository) Update(ctx context.Context, todo *model.Todo) error {
	err := r.DB.QueryRowContext(
		ctx,
		`UPDATE todos SET text = $1, completed = $2, due_date = $3, updated_at = NOW() WHERE id = $4 AND user_id = $5 RETURNING updated_at`,
		todo.Text, todo.Completed, todo.DueDate, todo.ID, todo.UserID,
	).Scan(&todo.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (r *SQLTodoRepository) Delete(ctx context.Context, id int64, userID string) error {
	result, err := r.DB.ExecContext(ctx, `DELETE FROM todos WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
