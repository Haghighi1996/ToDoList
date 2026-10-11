package router

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/arezoohaghighi/todo-api/internal/handler"
	"github.com/arezoohaghighi/todo-api/internal/middleware"
	"github.com/arezoohaghighi/todo-api/internal/repository"
)

func New(db *sql.DB, sessionRepo repository.SessionRepository, authHandler *handler.AuthHandler, todoHandler *handler.TodoHandler, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.Health(db))
	mux.HandleFunc("POST /api/v1/auth/anonymous", authHandler.Anonymous)
	mux.Handle("POST /api/v1/auth/logout", middleware.Auth(sessionRepo, http.HandlerFunc(authHandler.Logout)))
	mux.Handle("GET /api/v1/todos", middleware.Auth(sessionRepo, http.HandlerFunc(todoHandler.List)))
	mux.Handle("POST /api/v1/todos", middleware.Auth(sessionRepo, http.HandlerFunc(todoHandler.Create)))
	mux.Handle("GET /api/v1/todos/{id}", middleware.Auth(sessionRepo, http.HandlerFunc(todoHandler.Get)))
	mux.Handle("PATCH /api/v1/todos/{id}", middleware.Auth(sessionRepo, http.HandlerFunc(todoHandler.Update)))
	mux.Handle("DELETE /api/v1/todos/{id}", middleware.Auth(sessionRepo, http.HandlerFunc(todoHandler.Delete)))
	return middleware.Recovery(middleware.RequestID(middleware.Logging(logger, mux)))
}
