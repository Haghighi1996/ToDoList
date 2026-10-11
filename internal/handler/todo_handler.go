package handler

import (
	"log/slog"
	"math"
	"net/http"
	"strconv"

	"github.com/arezoohaghighi/todo-api/internal/dto"
	"github.com/arezoohaghighi/todo-api/internal/middleware"
	"github.com/arezoohaghighi/todo-api/internal/model"
	"github.com/arezoohaghighi/todo-api/internal/repository"
	"github.com/arezoohaghighi/todo-api/internal/service"
)

type TodoHandler struct {
	service *service.TodoService
	logger  *slog.Logger
}

func NewTodoHandler(service *service.TodoService, logger *slog.Logger) *TodoHandler {
	return &TodoHandler{service: service, logger: logger}
}

func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request dto.CreateTodoRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, h.logger, err)
		return
	}
	todo, err := h.service.Create(r.Context(), middleware.UserIDFromContext(r.Context()), request.Text, request.DueDate)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}
	writeJSON(w, http.StatusCreated, todoResponse(todo))
}

func (h *TodoHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := todoID(r)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}
	todo, err := h.service.Get(r.Context(), middleware.UserIDFromContext(r.Context()), id)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, todoResponse(todo))
}

func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	filters, err := parseListFilter(r)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}
	todos, total, err := h.service.List(r.Context(), middleware.UserIDFromContext(r.Context()), filters)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}
	items := make([]*dto.TodoResponse, 0, len(todos))
	for i := range todos {
		items = append(items, todoResponse(&todos[i]))
	}
	writeJSON(w, http.StatusOK, dto.TodoListResponse{
		Items: items, Page: filters.Page, Limit: filters.Limit, Total: total,
		TotalPages: int(math.Ceil(float64(total) / float64(filters.Limit))),
	})
}

func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := todoID(r)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}
	var request dto.UpdateTodoRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, h.logger, err)
		return
	}
	if request.Text == nil && request.Completed == nil && !request.DueDate.Set {
		writeError(w, h.logger, service.ErrInvalidRequest)
		return
	}
	// Set=true and Value=nil represents an explicit JSON null and clears the date.
	todo, err := h.service.Update(r.Context(), middleware.UserIDFromContext(r.Context()), id, request.Text, request.Completed, request.DueDate.Value, request.DueDate.Set)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, todoResponse(todo))
}

func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := todoID(r)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}
	if err := h.service.Delete(r.Context(), middleware.UserIDFromContext(r.Context()), id); err != nil {
		writeError(w, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func todoID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, service.ErrInvalidRequest
	}
	return id, nil
}

func parseListFilter(r *http.Request) (repository.ListFilter, error) {
	query := r.URL.Query()
	filter := repository.ListFilter{Page: 1, Limit: 20, Sort: query.Get("sort")}
	if filter.Sort != "" && filter.Sort != "due_date_asc" && filter.Sort != "due_date_desc" {
		return repository.ListFilter{}, service.ErrInvalidRequest
	}
	var err error
	if value := query.Get("page"); value != "" {
		filter.Page, err = strconv.Atoi(value)
		if err != nil || filter.Page < 1 {
			return repository.ListFilter{}, service.ErrInvalidRequest
		}
	}
	if value := query.Get("limit"); value != "" {
		filter.Limit, err = strconv.Atoi(value)
		if err != nil || filter.Limit < 1 || filter.Limit > 100 {
			return repository.ListFilter{}, service.ErrInvalidRequest
		}
	}
	for key, target := range map[string]**bool{"completed": &filter.Completed, "overdue": &filter.Overdue} {
		if value := query.Get(key); value != "" {
			parsed, parseErr := strconv.ParseBool(value)
			if parseErr != nil {
				return repository.ListFilter{}, service.ErrInvalidRequest
			}
			*target = &parsed
		}
	}
	return filter, nil
}

func todoResponse(todo *model.Todo) *dto.TodoResponse {
	return &dto.TodoResponse{ID: todo.ID, UserID: todo.UserID, Text: todo.Text, Completed: todo.Completed, DueDate: todo.DueDate, CreatedAt: todo.CreatedAt, UpdatedAt: todo.UpdatedAt}
}
