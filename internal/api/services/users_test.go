package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/dto"
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/services"
	"github.com/andresMTG/tcgstats-backend/tests/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const testJWTSecret = "test-secret-with-at-least-32-bytes!"

func TestNewUsersService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		secret    string
		tokenTTL  time.Duration
		wantError bool
	}{
		{name: "Valid configuration", secret: testJWTSecret, tokenTTL: time.Minute},
		{name: "JWT secret is too short", secret: "short", tokenTTL: time.Minute, wantError: true},
		{name: "JWT expiration is zero", secret: testJWTSecret, tokenTTL: 0, wantError: true},
		{name: "JWT expiration is negative", secret: testJWTSecret, tokenTTL: -time.Second, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, err := services.NewUsersService(&mocks.UserRepositoryMock{}, tt.secret, tt.tokenTTL)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, service)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, service)
		})
	}
}

func TestUsersService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		request   *dto.UserRequest
		repoError error
		wantError bool
	}{
		{
			name:    "Creates user with hashed password",
			request: &dto.UserRequest{Name: "Ada", Email: "ada@example.com", Password: "correct horse battery staple"},
		},
		{
			name:      "Repository error",
			request:   &dto.UserRequest{Name: "Ada", Email: "ada@example.com", Password: "secret"},
			repoError: errors.New("database unavailable"),
			wantError: true,
		},
		{
			name:      "Password exceeds bcrypt limit",
			request:   &dto.UserRequest{Name: "Ada", Email: "ada@example.com", Password: string(make([]byte, 73))},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mocks.UserRepositoryMock{}
			if tt.name != "Password exceeds bcrypt limit" {
				repo.On("Create", mock.MatchedBy(func(user *models.Users) bool {
					return user != nil && user.Name == tt.request.Name && user.Email == tt.request.Email &&
						bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(tt.request.Password)) == nil
				})).Return(tt.repoError).Once()
			}

			service, err := services.NewUsersService(repo, testJWTSecret, time.Minute)
			assert.NoError(t, err)
			err = service.Create(tt.request)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestUsersService_Login(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("018f47b2-7c9a-7abc-8def-0123456789ab")
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		user        *models.Users
		repoError   error
		password    string
		wantErr     error
		wantRepoErr bool
	}{
		{
			name:     "Valid credentials",
			user:     &models.Users{UUID: userID, Email: "ada@example.com", Password: string(passwordHash)},
			password: "secret",
		},
		{
			name:      "User does not exist",
			repoError: gorm.ErrRecordNotFound,
			password:  "secret",
			wantErr:   services.ErrInvalidCredentials,
		},
		{
			name:     "Incorrect password",
			user:     &models.Users{UUID: userID, Email: "ada@example.com", Password: string(passwordHash)},
			password: "wrong",
			wantErr:  services.ErrInvalidCredentials,
		},
		{
			name:        "Repository error",
			repoError:   errors.New("database unavailable"),
			password:    "secret",
			wantRepoErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mocks.UserRepositoryMock{}
			repo.On("FindByEmail", "ada@example.com").Return(tt.user, tt.repoError).Once()
			service, err := services.NewUsersService(repo, testJWTSecret, 15*time.Minute)
			assert.NoError(t, err)

			response, err := service.Login(&dto.UserLoginRequest{Email: "ada@example.com", Password: tt.password})

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, response)
			} else if tt.wantRepoErr {
				assert.EqualError(t, err, "database unavailable")
				assert.Nil(t, response)
			} else {
				if !assert.NoError(t, err) || !assert.NotNil(t, response) {
					return
				}
				assert.Equal(t, "Bearer", response.TokenType)
				assert.True(t, response.ExpiresAt.After(time.Now()))

				claims := &jwt.RegisteredClaims{}
				token, parseErr := jwt.ParseWithClaims(response.AccessToken, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte(testJWTSecret), nil
				})
				assert.NoError(t, parseErr)
				assert.True(t, token.Valid)
				assert.Equal(t, userID.String(), claims.Subject)
				assert.WithinDuration(t, response.ExpiresAt, claims.ExpiresAt.Time, time.Second)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestUsersService_ShowMe(t *testing.T) {
	t.Parallel()

	userID := "018f47b2-7c9a-7abc-8def-0123456789ab"
	tests := []struct {
		name        string
		user        *models.Users
		repoError   error
		wantErr     error
		wantRepoErr bool
	}{
		{
			name: "User found",
			user: &models.Users{Name: "Ada", Email: "ada@example.com"},
		},
		{
			name:      "User does not exist",
			repoError: gorm.ErrRecordNotFound,
			wantErr:   services.ErrInvalidCredentials,
		},
		{
			name:        "Repository error",
			repoError:   errors.New("database unavailable"),
			wantRepoErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mocks.UserRepositoryMock{}
			repo.On("FindByUuid", userID).Return(tt.user, tt.repoError).Once()
			service, err := services.NewUsersService(repo, testJWTSecret, time.Minute)
			assert.NoError(t, err)

			response, err := service.ShowMe(userID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, response)
			} else if tt.wantRepoErr {
				assert.EqualError(t, err, "database unavailable")
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, &dto.UserSimpleResponse{Name: "Ada", Email: "ada@example.com"}, response)
			}
			repo.AssertExpectations(t)
		})
	}
}
