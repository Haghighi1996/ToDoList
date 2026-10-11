package model

import (
	"time"
)

type Todo struct {
	ID        int64      `json:"id"`
	UserID    string     `json:"user_id"`
	Text      string     `json:"text"`
	Completed bool       `json:"completed"`
	DueDate   *time.Time `json:"due_date,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (t *Todo) IsOverdue(now time.Time) bool {
	if t.Completed {
		return false
	}
	if t.DueDate == nil {
		return false
	}
	return t.DueDate.Before(now)
}
