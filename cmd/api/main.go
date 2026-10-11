package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arezoohaghighi/todo-api/internal/config"
	"github.com/arezoohaghighi/todo-api/internal/db"
	"github.com/arezoohaghighi/todo-api/internal/handler"
	"github.com/arezoohaghighi/todo-api/internal/repository"
	"github.com/arezoohaghighi/todo-api/internal/router"
	"github.com/arezoohaghighi/todo-api/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	config, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.New(ctx, config.DatabaseURL)
	if err != nil {
		return err
	}
	defer database.Close()

	userRepo := repository.NewUserRepository(database.DB)
	sessionRepo := repository.NewSessionRepository(database.DB)
	todoRepo := repository.NewTodoRepository(database.DB)
	authService := service.NewAuthService(database.DB, userRepo, sessionRepo)
	todoService := service.NewTodoService(todoRepo)
	authHandler := handler.NewAuthHandler(authService, logger)
	todoHandler := handler.NewTodoHandler(todoService, logger)

	server := &http.Server{
		Addr:              ":" + config.Port,
		Handler:           router.New(database.DB, sessionRepo, authHandler, todoHandler, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server started", "port", config.Port)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
