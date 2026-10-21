package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andresMTG/tcgstats-backend/internal/api/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCORSMiddleware_AllowsConfiguredOriginWithCredentials(t *testing.T) {
	const origin = "https://tcgstats.example.com"
	cors, err := middleware.NewCORSMiddleware([]string{" " + origin + " "})
	require.NoError(t, err)

	called := false
	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/analytics/decks", nil)
	request.Header.Set("Origin", origin)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	assert.True(t, called)
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Equal(t, origin, response.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", response.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, response.Header().Get("Vary"), "Origin")
}

func TestCORSMiddleware_HandlesCredentialedPreflight(t *testing.T) {
	const origin = "https://tcgstats.example.com"
	cors, err := middleware.NewCORSMiddleware([]string{origin})
	require.NoError(t, err)

	called := false
	handler := cors(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	request := httptest.NewRequest(http.MethodOptions, "/api/users/login", nil)
	request.Header.Set("Origin", origin)
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "content-type")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	assert.False(t, called)
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Equal(t, origin, response.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", response.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, response.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, response.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
}

func TestCORSMiddleware_AllowsMutationPreflightMethods(t *testing.T) {
	cors, err := middleware.NewCORSMiddleware([]string{"https://tcgstats.example.com"})
	require.NoError(t, err)

	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodOptions, "/api/decks/id", nil)
			request.Header.Set("Origin", "https://tcgstats.example.com")
			request.Header.Set("Access-Control-Request-Method", method)
			request.Header.Set("Access-Control-Request-Headers", "content-type")
			response := httptest.NewRecorder()
			cors(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(response, request)

			assert.Equal(t, http.StatusNoContent, response.Code)
			assert.Contains(t, response.Header().Get("Access-Control-Allow-Methods"), method)
		})
	}
}

func TestCORSMiddleware_RejectsUnconfiguredOrigin(t *testing.T) {
	cors, err := middleware.NewCORSMiddleware([]string{"https://tcgstats.example.com"})
	require.NoError(t, err)

	called := false
	handler := cors(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	request := httptest.NewRequest(http.MethodPost, "/api/users/login", nil)
	request.Header.Set("Origin", "https://attacker.example")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, response.Code)
	assert.Empty(t, response.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_RejectsDisallowedPreflightMethodAndHeaders(t *testing.T) {
	cors, err := middleware.NewCORSMiddleware([]string{"https://tcgstats.example.com"})
	require.NoError(t, err)

	tests := []struct {
		name    string
		method  string
		headers string
	}{
		{name: "unsupported method", method: http.MethodTrace},
		{name: "unsupported header", method: http.MethodPost, headers: "x-custom-header"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodOptions, "/api/users/login", nil)
			request.Header.Set("Origin", "https://tcgstats.example.com")
			request.Header.Set("Access-Control-Request-Method", test.method)
			request.Header.Set("Access-Control-Request-Headers", test.headers)
			response := httptest.NewRecorder()
			cors(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(response, request)
			assert.Equal(t, http.StatusForbidden, response.Code)
		})
	}
}

func TestNewCORSMiddleware_RequiresExplicitOrigins(t *testing.T) {
	for _, origin := range []string{"*", "https://example.com/path", "ftp://example.com"} {
		t.Run(origin, func(t *testing.T) {
			_, err := middleware.NewCORSMiddleware([]string{origin})
			require.Error(t, err)
		})
	}
}

func TestCORSMiddleware_AllowsRequestsWithoutOrigin(t *testing.T) {
	cors, err := middleware.NewCORSMiddleware(nil)
	require.NoError(t, err)
	called := false
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	response := httptest.NewRecorder()
	cors(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(response, request)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, response.Code)
}
