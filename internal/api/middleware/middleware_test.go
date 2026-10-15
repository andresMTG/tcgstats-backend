package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/auth"
	"github.com/andresMTG/tcgstats-backend/internal/api/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_AcceptsAccessCookie(t *testing.T) {
	userID := uuid.New()
	token := signedToken(t, userID, auth.AccessTokenType)
	called := false
	handler := middleware.AuthMiddleware("test-secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		assert.Equal(t, userID.String(), r.Context().Value(middleware.UserIDKey))
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	request.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: token})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	assert.True(t, called)
	assert.Equal(t, http.StatusNoContent, response.Code)
}

func TestAuthMiddleware_AcceptsBearerFallback(t *testing.T) {
	token := signedToken(t, uuid.New(), auth.AccessTokenType)
	called := false
	handler := middleware.AuthMiddleware("test-secret")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	assert.True(t, called)
	assert.Equal(t, http.StatusNoContent, response.Code)
}

func TestAuthMiddleware_AcceptsLegacyBearerToken(t *testing.T) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject: uuid.NewString(), ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString([]byte("test-secret"))
	require.NoError(t, err)
	called := false
	handler := middleware.AuthMiddleware("test-secret")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	assert.True(t, called)
	assert.Equal(t, http.StatusNoContent, response.Code)
}

func TestAuthMiddleware_RejectsRefreshCookieAsAccessToken(t *testing.T) {
	token := signedToken(t, uuid.New(), auth.RefreshTokenType)
	called := false
	handler := middleware.AuthMiddleware("test-secret")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	request.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: token})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestAuthMiddleware_PrefersInvalidCookieOverBearerHeader(t *testing.T) {
	token := signedToken(t, uuid.New(), auth.AccessTokenType)
	called := false
	handler := middleware.AuthMiddleware("test-secret")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	request.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: "invalid"})
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func signedToken(t *testing.T, userID uuid.UUID, tokenType string) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID.String(), ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Type: tokenType,
	}).SignedString([]byte("test-secret"))
	require.NoError(t, err)
	return token
}
