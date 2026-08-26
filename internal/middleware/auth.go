package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/7DeN4iK7/auth-service/internal/auth"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func AuthMiddleware(jwtService *auth.JWTService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Token not found", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwtService.ParseToken(tokenString)
		if err != nil || !token.Valid {
			slog.Error("Error parsing token", slog.Any("err", err))
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		userID, err := jwtService.GetUserID(token)
		if err != nil {
			slog.Error("Error getting user_id", slog.Any("err", err))
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		next(w, r.WithContext(ctx))
	}
}
