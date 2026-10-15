package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/andresMTG/tcgstats-backend/internal/api/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "userID"

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			tokenString, err := accessTokenFromRequest(r)
			if err != nil {
				http.Error(w, "unauthorized: missing or invalid token", http.StatusUnauthorized)
				return
			}

			claims, err := auth.ParseToken(tokenString, jwtSecret)
			if err != nil || claims.Type != "" && claims.Type != auth.AccessTokenType {
				http.Error(w, "unauthorized: invalid token", http.StatusUnauthorized)
				return
			}

			if _, err := uuid.Parse(claims.Subject); err != nil {
				http.Error(w, "unauthorized: invalid token payload", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func accessTokenFromRequest(r *http.Request) (string, error) {
	if cookie, err := r.Cookie(auth.AccessCookieName); err == nil {
		if cookie.Value == "" {
			return "", jwt.ErrTokenMalformed
		}
		return cookie.Value, nil
	}

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", jwt.ErrTokenMalformed
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		return "", jwt.ErrTokenMalformed
	}
	return tokenString, nil
}
