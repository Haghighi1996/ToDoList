package service

import (
	"net/http"
)

type AppError struct {
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrInvalidRequest = &AppError{Code: "invalid_request", Message: "Invalid request"}
	ErrUnauthorized   = &AppError{Code: "unauthorized", Message: "Unauthorized"}
	ErrNotFound       = &AppError{Code: "not_found", Message: "Not found"}
	ErrInternal       = &AppError{Code: "internal_server_error", Message: "Internal server error"}
)

func (e *AppError) Status() int {
	switch e.Code {
	case "invalid_request":
		return http.StatusBadRequest
	case "unauthorized":
		return http.StatusUnauthorized
	case "not_found":
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}
