package dto

import (
	"bytes"
	"encoding/json"
	"time"
)

// OptionalTime distinguishes an omitted JSON field from a null value.
// Set=true and Value=nil means the client explicitly cleared the due date.
type OptionalTime struct {
	Set   bool
	Value *time.Time
}

func (o *OptionalTime) UnmarshalJSON(data []byte) error {
	o.Set = true
	if bytes.Equal(data, []byte("null")) {
		o.Value = nil
		return nil
	}
	var value time.Time
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	o.Value = &value
	return nil
}

type CreateTodoRequest struct {
	Text    string     `json:"text"`
	DueDate *time.Time `json:"due_date"`
}

type UpdateTodoRequest struct {
	Text      *string      `json:"text"`
	Completed *bool        `json:"completed"`
	DueDate   OptionalTime `json:"due_date"`
}

type TodoResponse struct {
	ID        int64      `json:"id"`
	UserID    string     `json:"user_id"`
	Text      string     `json:"text"`
	Completed bool       `json:"completed"`
	DueDate   *time.Time `json:"due_date"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type TodoListResponse struct {
	Items      []*TodoResponse `json:"items"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	Total      int64           `json:"total"`
	TotalPages int             `json:"total_pages"`
}
