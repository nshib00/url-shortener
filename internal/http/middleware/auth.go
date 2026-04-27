package middleware

import (
	"context"
	auth "go-url-shortener/internal/auth"
	resp "go-url-shortener/internal/utils/api/response"
	"net/http"
	"strings"

	"github.com/go-chi/render"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware(secretKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				render.JSON(w, r, resp.Error("missing auth header"))
				return
			}

			headerParts := strings.Split(authHeader, " ")
			if headerParts[0] != "Bearer" || len(headerParts) != 2 {
				render.JSON(w, r, resp.Error("invalid auth header"))
				return
			}
			tokenStr := headerParts[1]

			userID, err := auth.ValidateToken(tokenStr, secretKey)
			if err != nil {
				render.JSON(w, r, resp.Error("invalid token"))
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
