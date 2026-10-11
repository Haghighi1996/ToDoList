package dto

import (
	"encoding/json"
	"testing"
)

func TestOptionalTimeDistinguishesOmittedAndNull(t *testing.T) {
	var omitted UpdateTodoRequest
	if err := json.Unmarshal([]byte(`{}`), &omitted); err != nil {
		t.Fatal(err)
	}
	if omitted.DueDate.Set {
		t.Fatal("omitted due_date is marked set")
	}
	var cleared UpdateTodoRequest
	if err := json.Unmarshal([]byte(`{"due_date":null}`), &cleared); err != nil {
		t.Fatal(err)
	}
	if !cleared.DueDate.Set || cleared.DueDate.Value != nil {
		t.Fatal("null due_date must explicitly clear the value")
	}
}
