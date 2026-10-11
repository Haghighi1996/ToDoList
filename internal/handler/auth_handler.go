package handler

import (
	"log/slog"
	"net/http"

	"github.com/arezoohaghighi/todo-api/internal/dto"
	"github.com/arezoohaghighi/todo-api/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
	logger  *slog.Logger
}

func NewAuthHandler(service *service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{service: service, logger: logger}
}

func (h *AuthHandler) Anonymous(w http.ResponseWriter, r *http.Request) {
	var request dto.AnonymousAuthRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, h.logger, err)
		return
	}
	token, userID, expiresAt, err := h.service.CreateAnonymousUser(r.Context(), request.DeviceID)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.AnonymousAuthResponse{Token: token, UserID: userID, ExpiresAt: expiresAt.UTC().Format("2006-01-02T15:04:05Z07:00")})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	token := authHeader[len("Bearer "):]
	if err := h.service.Logout(r.Context(), token); err != nil {
		writeError(w, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
