package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/arezoohaghighi/todo-api/internal/dto"
	"github.com/arezoohaghighi/todo-api/internal/repository"
	"github.com/arezoohaghighi/todo-api/internal/service"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, logger *slog.Logger, err error) {
	var appError *service.AppError
	if errors.As(err, &appError) {
		writeJSON(w, appError.Status(), dto.ErrorResponse{Error: appError.Code, Message: appError.Message})
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "Not found"})
		return
	}
	logger.Error("unexpected request error", "error", err)
	writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_server_error", Message: "Internal server error"})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return service.ErrInvalidRequest
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return service.ErrInvalidRequest
	}
	return nil
}
