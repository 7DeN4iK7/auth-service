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

	"github.com/7DeN4iK7/auth-service/internal/auth"
	"github.com/7DeN4iK7/auth-service/internal/config"
	"github.com/7DeN4iK7/auth-service/internal/middleware"
	"github.com/7DeN4iK7/auth-service/internal/users"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		slog.Error("Config loading error", slog.Any("err", err))
		os.Exit(1)
	}

	mux := http.NewServeMux()

	rep, err := users.NewRepository(cfg.PostgresCfg)
	if err != nil {
		slog.Error("Error creating repository: ", slog.Any("error", err))
		os.Exit(1)
	}
	defer rep.Close()

	jwt := auth.NewJWT(cfg.JwtCfg)
	user_handler := users.NewHandler(users.NewService(rep, jwt))

	mux.HandleFunc("POST /auth/register", user_handler.CreateUser())
	mux.HandleFunc("POST /auth/login", user_handler.LoginUser())
	mux.HandleFunc("GET /my_info", middleware.AuthMiddleware(jwt, user_handler.UserInfo()))

	server := http.Server{
		Addr:              cfg.ServerCfg.Addr,
		ReadTimeout:       cfg.ServerCfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ServerCfg.ReadHeaderTimeout,
		IdleTimeout:       cfg.ServerCfg.IdleTimeout,
		WriteTimeout:      cfg.ServerCfg.WriteTimeout,
		Handler:           mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("Starting server...")

	serverErr := make(chan error)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", slog.Any("err", err))
		}
	case <-ctx.Done():
		slog.Info("Shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown error", slog.Any("err", err))
	}

	slog.Info("Server stopped")

}
