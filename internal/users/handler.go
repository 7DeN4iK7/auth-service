package users

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/7DeN4iK7/auth-service/internal/middleware"
)

type UserService interface {
	CreateUser(ctx context.Context, r RegisterRequest) (int, error)
	LoginUser(ctx context.Context, r LoginRequest) (string, error)
	GetUserByID(ctx context.Context, id int) (User, error)
}

type Handler struct {
	Service UserService
}

func NewHandler(s UserService) *Handler {
	return &Handler{Service: s}
}

func (h *Handler) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)

		var request RegisterRequest

		decoder.Decode(&request)

		id, err := h.Service.CreateUser(r.Context(), request)
		if err != nil {
			slog.Error("Create user error", slog.Any("error", err))
			switch {
			case errors.Is(err, ErrNoPassword):
				http.Error(w, "No password entered", http.StatusBadRequest)
			case errors.Is(err, ErrNoUsername):
				http.Error(w, "No username entered", http.StatusBadRequest)
			case errors.Is(err, ErrUserAlreadyExists):
				http.Error(w, "User already exists", http.StatusConflict)
			default:
				http.Error(w, "Unknown error", http.StatusInternalServerError)
			}

			return
		}

		slog.Info("Created successful", slog.Any("id", id))

		w.Write([]byte("Created successful"))
	}
}

func (h *Handler) LoginUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request LoginRequest

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		token, err := h.Service.LoginUser(r.Context(), request)
		if err != nil {
			slog.Error("Login error", slog.Any("err", err))
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(LoginResponse{
			Token: token,
		})
	}
}

func (h *Handler) UserInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			http.Error(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}

		user, err := h.Service.GetUserByID(r.Context(), userID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UserResponse{
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		})
	}
}
