package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/dto"
	"github.com/andresMTG/tcgstats-backend/internal/api/middleware"
	"github.com/andresMTG/tcgstats-backend/internal/api/services"
	"github.com/andresMTG/tcgstats-backend/tests/mocks"
	"github.com/andresMTG/tcgstats-backend/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUsersController_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		body           []byte
		serviceErr     error
		wantStatusCode int
		wantRequest    *dto.UserRequest
	}{
		{
			name:           "Works correctly",
			body:           []byte(`{"Name":"Ada","Email":"ada@example.com","Password":"secret"}`),
			wantStatusCode: http.StatusCreated,
			wantRequest:    &dto.UserRequest{Name: "Ada", Email: "ada@example.com", Password: "secret"},
		},
		{
			name:           "Invalid JSON",
			body:           []byte(`{"Name":`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "Service error",
			body:           []byte(`{"Name":"Ada","Email":"ada@example.com","Password":"secret"}`),
			serviceErr:     errors.New("create failed"),
			wantStatusCode: http.StatusBadRequest,
			wantRequest:    &dto.UserRequest{Name: "Ada", Email: "ada@example.com", Password: "secret"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := &mocks.UsersServiceMock{}
			if tt.wantRequest != nil {
				service.On("Create", mock.MatchedBy(func(req *dto.UserRequest) bool {
					return req != nil && *req == *tt.wantRequest
				})).Return(tt.serviceErr).Once()
			}
			controller := &Users{services: service}
			rec := httptest.NewRecorder()
			req := utils.BuildJSONRequest(t, tt.body, "/users", http.MethodPost)

			controller.Create(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
			service.AssertExpectations(t)
		})
	}
}

func TestUsersController_Login(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2026, time.September, 30, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name           string
		body           []byte
		loginResponse  *dto.UserLoginResponse
		serviceErr     error
		wantStatusCode int
		wantRequest    *dto.UserLoginRequest
	}{
		{
			name:           "Works correctly",
			body:           []byte(`{"Email":"ada@example.com","Password":"secret"}`),
			loginResponse:  &dto.UserLoginResponse{AccessToken: "signed-token", TokenType: "Bearer", ExpiresAt: expiresAt},
			wantStatusCode: http.StatusOK,
			wantRequest:    &dto.UserLoginRequest{Email: "ada@example.com", Password: "secret"},
		},
		{
			name:           "Invalid JSON",
			body:           []byte(`{"Email":`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "Missing credentials",
			body:           []byte(`{"Email":"ada@example.com"}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "Invalid credentials",
			body:           []byte(`{"Email":"ada@example.com","Password":"wrong"}`),
			serviceErr:     services.ErrInvalidCredentials,
			wantStatusCode: http.StatusUnauthorized,
			wantRequest:    &dto.UserLoginRequest{Email: "ada@example.com", Password: "wrong"},
		},
		{
			name:           "Unexpected service error",
			body:           []byte(`{"Email":"ada@example.com","Password":"secret"}`),
			serviceErr:     errors.New("database unavailable"),
			wantStatusCode: http.StatusInternalServerError,
			wantRequest:    &dto.UserLoginRequest{Email: "ada@example.com", Password: "secret"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := &mocks.UsersServiceMock{}
			if tt.wantRequest != nil {
				service.On("Login", mock.MatchedBy(func(req *dto.UserLoginRequest) bool {
					return req != nil && *req == *tt.wantRequest
				})).Return(tt.loginResponse, tt.serviceErr).Once()
			}
			controller := &Users{services: service}
			rec := httptest.NewRecorder()
			req := utils.BuildJSONRequest(t, tt.body, "/users/login", http.MethodPost)

			controller.Login(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
			service.AssertExpectations(t)
			if tt.wantStatusCode == http.StatusOK {
				assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
				var got dto.UserLoginResponse
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
				assert.Equal(t, *tt.loginResponse, got)
			}
		})
	}
}

func TestUsersController_ShowMe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		response       *dto.UserSimpleResponse
		serviceErr     error
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "Works correctly",
			response:       &dto.UserSimpleResponse{Name: "Ada", Email: "ada@example.com"},
			wantStatusCode: http.StatusOK,
			wantBody:       "Olá, o seu nome é: Ada e o seu email é: ada@example.com",
		},
		{
			name:           "Service error",
			serviceErr:     services.ErrInvalidCredentials,
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "User error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := &mocks.UsersServiceMock{}
			service.On("ShowMe", "user-123").Return(tt.response, tt.serviceErr).Once()
			controller := &Users{services: service}
			rec := httptest.NewRecorder()
			req := utils.BuildJSONRequest(t, nil, "/users/me", http.MethodGet)
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user-123"))

			controller.ShowMe(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
			service.AssertExpectations(t)
		})
	}
}
