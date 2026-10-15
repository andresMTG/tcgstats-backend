package controllers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/auth"
	"github.com/andresMTG/tcgstats-backend/internal/api/controllers"
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/andresMTG/tcgstats-backend/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserController_CreateUser(t *testing.T) {
	t.Run("creates user", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		want := &models.User{UUID: uuid.New(), Name: "Player", Email: "player@example.com"}
		service.On("CreateUser", mock.Anything, "Player", "player@example.com", "password").Return(want, nil).Once()
		response := recorder()

		newUserController(service).CreateUser(response, newRequest(http.MethodPost, "/users",
			`{"name":"Player","email":"player@example.com","password":"password"}`))

		require.Equal(t, http.StatusCreated, response.Code)
		var got models.User
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		assert.Equal(t, want.UUID, got.UUID)
		assert.Empty(t, got.Password)
		service.AssertExpectations(t)
	})
	t.Run("rejects malformed JSON", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		response := recorder()
		newUserController(service).CreateUser(response, newRequest(http.MethodPost, "/users", "{"))
		assert.Equal(t, http.StatusBadRequest, response.Code)
		service.AssertNotCalled(t, "CreateUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("reports service failure", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		service.On("CreateUser", mock.Anything, "Player", "player@example.com", "password").
			Return((*models.User)(nil), errors.New("conflict")).Once()
		response := recorder()
		newUserController(service).CreateUser(response, newRequest(http.MethodPost, "/users",
			`{"name":"Player","email":"player@example.com","password":"password"}`))
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		service.AssertExpectations(t)
	})
}

func TestUserController_Login(t *testing.T) {
	t.Run("sets HttpOnly auth cookies without returning tokens in JSON", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		tokens := &models.AuthTokens{
			AccessToken: "access-token", RefreshToken: "refresh-token",
			AccessTokenTTL: 15 * time.Minute, RefreshTokenTTL: 30 * 24 * time.Hour,
		}
		service.On("Login", mock.Anything, "player@example.com", "password").Return(tokens, nil).Once()
		response := recorder()
		newUserController(service).Login(response, newRequest(http.MethodPost, "/users/login",
			`{"email":"player@example.com","password":"password"}`))
		require.Equal(t, http.StatusNoContent, response.Code)
		assert.Empty(t, response.Body.String())
		assert.Equal(t, "no-store", response.Header().Get("Cache-Control"))
		cookies := response.Result().Cookies()
		require.Len(t, cookies, 2)
		assert.Equal(t, auth.AccessCookieName, cookies[0].Name)
		assert.Equal(t, "access-token", cookies[0].Value)
		assert.True(t, cookies[0].HttpOnly)
		assert.True(t, cookies[0].Secure)
		assert.Equal(t, "/api", cookies[0].Path)
		assert.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)
		assert.Equal(t, 15*60, cookies[0].MaxAge)
		assert.Equal(t, auth.RefreshCookieName, cookies[1].Name)
		assert.Equal(t, "/api/users", cookies[1].Path)
		assert.True(t, cookies[1].HttpOnly)
		assert.True(t, cookies[1].Secure)
		assert.Equal(t, 30*24*60*60, cookies[1].MaxAge)
		service.AssertExpectations(t)
	})
	t.Run("rejects malformed JSON", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		response := recorder()
		newUserController(service).Login(response, newRequest(http.MethodPost, "/users/login", "{"))
		assert.Equal(t, http.StatusBadRequest, response.Code)
		service.AssertNotCalled(t, "Login", mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("rejects service failure", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		service.On("Login", mock.Anything, "player@example.com", "wrong").
			Return((*models.AuthTokens)(nil), interfaces.ErrInvalidCredentials).Once()
		response := recorder()
		newUserController(service).Login(response, newRequest(http.MethodPost, "/users/login",
			`{"email":"player@example.com","password":"wrong"}`))
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		service.AssertExpectations(t)
	})
	t.Run("reports session storage failure", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		service.On("Login", mock.Anything, "player@example.com", "password").
			Return((*models.AuthTokens)(nil), errors.New("redis unavailable")).Once()
		response := recorder()
		newUserController(service).Login(response, newRequest(http.MethodPost, "/users/login",
			`{"email":"player@example.com","password":"password"}`))
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		service.AssertExpectations(t)
	})
}

func TestUserController_Refresh(t *testing.T) {
	t.Run("rotates auth cookies", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		tokens := &models.AuthTokens{
			AccessToken: "new-access", RefreshToken: "new-refresh",
			AccessTokenTTL: 15 * time.Minute, RefreshTokenTTL: 30 * 24 * time.Hour,
		}
		service.On("Refresh", mock.Anything, "old-refresh").Return(tokens, nil).Once()
		request := newRequest(http.MethodPost, "/users/refresh", "")
		request.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: "old-refresh"})
		response := recorder()

		newUserController(service).Refresh(response, request)

		require.Equal(t, http.StatusNoContent, response.Code)
		assert.Equal(t, "no-store", response.Header().Get("Cache-Control"))
		cookies := response.Result().Cookies()
		require.Len(t, cookies, 2)
		assert.Equal(t, "new-access", cookies[0].Value)
		assert.Equal(t, "new-refresh", cookies[1].Value)
		assert.True(t, cookies[0].HttpOnly)
		assert.True(t, cookies[1].HttpOnly)
		service.AssertExpectations(t)
	})
	t.Run("rejects missing refresh cookie and clears auth cookies", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		response := recorder()
		newUserController(service).Refresh(response, newRequest(http.MethodPost, "/users/refresh", ""))
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		cookies := response.Result().Cookies()
		require.Len(t, cookies, 2)
		assert.Equal(t, -1, cookies[0].MaxAge)
		assert.Equal(t, -1, cookies[1].MaxAge)
		service.AssertNotCalled(t, "Refresh", mock.Anything, mock.Anything)
	})
	t.Run("rejects invalid refresh token", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		service.On("Refresh", mock.Anything, "expired").
			Return((*models.AuthTokens)(nil), interfaces.ErrInvalidRefreshToken).Once()
		request := newRequest(http.MethodPost, "/users/refresh", "")
		request.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: "expired"})
		response := recorder()
		newUserController(service).Refresh(response, request)
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		service.AssertExpectations(t)
	})
}

func TestUserController_Logout(t *testing.T) {
	t.Run("revokes refresh session and expires cookies", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		service.On("Logout", mock.Anything, "refresh-token").Return(nil).Once()
		request := newRequest(http.MethodPost, "/users/logout", "")
		request.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: "refresh-token"})
		response := recorder()

		newUserController(service).Logout(response, request)

		require.Equal(t, http.StatusNoContent, response.Code)
		cookies := response.Result().Cookies()
		require.Len(t, cookies, 2)
		assert.Equal(t, -1, cookies[0].MaxAge)
		assert.Equal(t, -1, cookies[1].MaxAge)
		assert.True(t, cookies[0].HttpOnly)
		assert.True(t, cookies[1].HttpOnly)
		service.AssertExpectations(t)
	})
	t.Run("logout without refresh cookie is idempotent", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		response := recorder()
		newUserController(service).Logout(response, newRequest(http.MethodPost, "/users/logout", ""))
		assert.Equal(t, http.StatusNoContent, response.Code)
		service.AssertNotCalled(t, "Logout", mock.Anything, mock.Anything)
	})
	t.Run("logs revocation errors without changing the response", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		service.On("Logout", mock.Anything, "jwt-secret-value").
			Return(errors.New("redis unavailable")).Once()
		request := newRequest(http.MethodPost, "/users/logout", "")
		request.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: "jwt-secret-value"})
		response := recorder()
		var logs bytes.Buffer
		originalLogOutput := log.Writer()
		log.SetOutput(&logs)
		t.Cleanup(func() {
			log.SetOutput(originalLogOutput)
		})

		newUserController(service).Logout(response, request)

		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Equal(t, "unable to revoke session\n", response.Body.String())
		assert.Contains(t, logs.String(), "logout: failed to revoke refresh-token session: redis unavailable")
		assert.NotContains(t, logs.String(), "jwt-secret-value")
		service.AssertExpectations(t)
	})
}

func TestUserController_GetMe(t *testing.T) {
	t.Run("requires user identity", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		response := recorder()
		newUserController(service).GetMe(response, newRequest(http.MethodGet, "/users/me", ""))
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		service.AssertNotCalled(t, "GetUserByID", mock.Anything, mock.Anything)
	})
	t.Run("rejects invalid identity", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		request := newRequest(http.MethodGet, "/users/me", "")
		request = withUserID(request, "not-a-uuid")
		response := recorder()
		newUserController(service).GetMe(response, request)
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		service.AssertNotCalled(t, "GetUserByID", mock.Anything, mock.Anything)
	})
	t.Run("returns current user", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		id := uuid.New()
		want := &models.User{UUID: id, Name: "Player"}
		service.On("GetUserByID", mock.Anything, id).Return(want, nil).Once()
		request := withUserID(newRequest(http.MethodGet, "/users/me", ""), id.String())
		response := recorder()
		newUserController(service).GetMe(response, request)
		require.Equal(t, http.StatusOK, response.Code)
		var got models.User
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		assert.Equal(t, id, got.UUID)
		service.AssertExpectations(t)
	})
	t.Run("returns not found on service failure", func(t *testing.T) {
		service := &mocks.UserServiceMock{}
		id := uuid.New()
		service.On("GetUserByID", mock.Anything, id).Return((*models.User)(nil), errors.New("missing")).Once()
		request := withUserID(newRequest(http.MethodGet, "/users/me", ""), id.String())
		response := recorder()
		newUserController(service).GetMe(response, request)
		assert.Equal(t, http.StatusNotFound, response.Code)
		service.AssertExpectations(t)
	})
}

func newUserController(service *mocks.UserServiceMock) *controllers.UserController {
	return controllers.NewUserController(service, &controllers.CookieSettings{
		Secure: true, SameSite: "lax", BasePath: "/api",
	})
}
