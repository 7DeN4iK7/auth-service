package main

import (
	"log/slog"
	"net/http"
	"os"

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

	jwt := auth.NewJWT("secret")
	user_handler := users.NewHandler(users.NewService(rep, jwt))

	mux.HandleFunc("POST /auth/register", user_handler.CreateUser())
	mux.HandleFunc("POST /auth/login", user_handler.LoginUser())
	mux.HandleFunc("GET /my_info", middleware.AuthMiddleware(jwt, user_handler.UserInfo()))

	server := http.Server{
		Addr:              cfg.ServerCfg.Addr,
		ReadTimeout:       cfg.ServerCfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ServerCfg.ReadHeaderTimeout,
		IdleTimeout:       cfg.ServerCfg.ReadTimeout,
		WriteTimeout:      cfg.ServerCfg.WriteTimeout,
		Handler:           mux,
	}

	slog.Info("Starting server...")

	if err := server.ListenAndServe(); err != nil {
		slog.Error("Error running server", slog.Any("err", err))
		os.Exit(1)
	}

}
