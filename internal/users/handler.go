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

func jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", slog.Any("err", err))
	}
}

func (h *Handler) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request RegisterRequest

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		id, err := h.Service.CreateUser(r.Context(), request)
		if err != nil {
			slog.Error("Create user error", slog.Any("error", err))
			switch {
			case errors.Is(err, ErrNoPassword),
				errors.Is(err, ErrNoUsername),
				errors.Is(err, ErrTooLongPassword),
				errors.Is(err, ErrTooLongUsername):
				http.Error(w, err.Error(), http.StatusBadRequest)
			case errors.Is(err, ErrUserAlreadyExists):
				http.Error(w, err.Error(), http.StatusConflict)
			default:
				http.Error(w, "unknown error", http.StatusInternalServerError)
			}

			return
		}

		slog.Info("created successful", slog.Any("id", id))

		w.WriteHeader(http.StatusCreated)
		jsonResponse(w, CreateResponse{ID: id})
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
			switch {
			case errors.Is(err, ErrWrongPassword),
				errors.Is(err, ErrUserNotFound):
				http.Error(w, err.Error(), http.StatusUnauthorized)
			default:
				http.Error(w, "internal server error", http.StatusInternalServerError)
				slog.Error("Login error", slog.Any("err", err))
			}
			return
		}

		jsonResponse(w, LoginResponse{Token: token})
	}
}

func (h *Handler) UserInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			http.Error(w, "invalid user ID", http.StatusUnauthorized)
			return
		}

		user, err := h.Service.GetUserByID(r.Context(), userID)
		if err != nil {
			switch {
			case errors.Is(err, ErrUserNotFound):
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		jsonResponse(w, UserResponse{
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		})
	}
}
