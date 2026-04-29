package testutils

import (
	"context"
	"encoding/json"
	"go-url-shortener/internal/auth"
	"go-url-shortener/internal/http/middleware"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

const TestSecretKey = "test_secret_key_for_jwt"

func GenerateTestToken(t *testing.T, userID int) string {
	token, err := auth.GenerateToken(userID, TestSecretKey)
	require.NoError(t, err)
	return token
}

func CreateAuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "missing auth header"})
				return
			}

			if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid auth header"})
				return
			}

			tokenStr := authHeader[7:]

			userID, err := auth.ValidateToken(tokenStr, TestSecretKey)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid token"})
				return
			}

			ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
