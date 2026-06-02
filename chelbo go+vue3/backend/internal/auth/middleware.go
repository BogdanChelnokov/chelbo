package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"chelbo/backend/internal/models"
	"chelbo/backend/internal/pkg/config"
	"chelbo/backend/internal/pkg/logger"
)

type contextKey string

const (
	UserIDKey    contextKey = "userID"
	UserEmailKey contextKey = "userEmail"
)

func AuthMiddleware(cfg *config.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				sendUnauthorized(w, "missing authorization header")
				return
			}

			// Check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				sendUnauthorized(w, "invalid authorization format")
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := ValidateJWT(tokenString, cfg)
			if err != nil {
				logger.Errorf("Token validation failed: %v", err)
				sendUnauthorized(w, "invalid or expired token")
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func sendUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	response := models.NewErrorResponse(message)
	json.NewEncoder(w).Encode(response)
}

// Helper functions to get user info from context
func GetUserID(ctx context.Context) (uint64, bool) {
	userID, ok := ctx.Value(UserIDKey).(uint64)
	return userID, ok
}

func GetUserEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}
