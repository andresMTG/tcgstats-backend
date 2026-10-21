package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/auth"
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/andresMTG/tcgstats-backend/internal/api/services"
	"github.com/andresMTG/tcgstats-backend/tests/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_CreateUser(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	repo := &mocks.UserRepositoryMock{}
	ctx := context.Background()
	repo.On("Create", ctx, mock.MatchedBy(func(user *models.User) bool {
		return user.Name == "Player" && user.Email == "player@example.com"
	})).Return(nil).Once()

	service, err := services.NewUserService(repo, &mocks.RefreshTokenStoreMock{})
	require.NoError(t, err)

	user, err := service.CreateUser(ctx, "Player", "player@example.com", "password")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.NotEqual(t, uuid.Nil, user.UUID)
	assert.Equal(t, "Player", user.Name)
	assert.Equal(t, "player@example.com", user.Email)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password")))
	assert.NotEqual(t, "password", user.Password)
	repo.AssertExpectations(t)
}

func TestUserService_CreateUser_RepositoryError(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	repo := &mocks.UserRepositoryMock{}
	wantErr := errors.New("database unavailable")
	repo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(wantErr).Once()
	service, err := services.NewUserService(repo, &mocks.RefreshTokenStoreMock{})
	require.NoError(t, err)

	user, err := service.CreateUser(context.Background(), "Player", "player@example.com", "password")
	assert.Nil(t, user)
	assert.ErrorIs(t, err, wantErr)
	repo.AssertExpectations(t)
}

func TestUserService_CreateUser_PasswordTooLong(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	repo := &mocks.UserRepositoryMock{}
	service, err := services.NewUserService(repo, &mocks.RefreshTokenStoreMock{})
	require.NoError(t, err)

	user, err := service.CreateUser(context.Background(), "Player", "player@example.com", strings.Repeat("x", 73))
	assert.Nil(t, user)
	require.Error(t, err)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestUserService_Login(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_EXPIRATION_MINUTES", "15")
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &models.User{UUID: uuid.New(), Email: "player@example.com", Password: string(passwordHash)}
	repo := &mocks.UserRepositoryMock{}
	refreshStore := &mocks.RefreshTokenStoreMock{}
	repo.On("GetByEmail", mock.Anything, user.Email).Return(user, nil).Once()
	refreshStore.On("SaveRefreshToken", mock.Anything, mock.MatchedBy(func(tokenID string) bool {
		_, err := uuid.Parse(tokenID)
		return err == nil
	}), user.UUID, 30*24*time.Hour).Return(nil).Once()
	service, err := services.NewUserService(repo, refreshStore)
	require.NoError(t, err)

	tokens, err := service.Login(context.Background(), user.Email, "password")
	require.NoError(t, err)
	require.NotNil(t, tokens)
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEmpty(t, tokens.RefreshToken)

	claims, err := auth.ParseToken(tokens.AccessToken, "test-secret")
	require.NoError(t, err)
	assert.Equal(t, user.UUID.String(), claims.Subject)
	assert.Equal(t, auth.AccessTokenType, claims.Type)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
	assert.True(t, claims.ExpiresAt.Time.Before(time.Now().Add(16*time.Minute)))
	refreshClaims, err := auth.ParseToken(tokens.RefreshToken, "test-secret")
	require.NoError(t, err)
	assert.Equal(t, auth.RefreshTokenType, refreshClaims.Type)
	assert.Equal(t, user.UUID.String(), refreshClaims.Subject)
	assert.True(t, refreshClaims.ExpiresAt.Time.After(time.Now().Add(29*24*time.Hour)))
	repo.AssertExpectations(t)
	refreshStore.AssertExpectations(t)
}

func TestUserService_Login_InvalidCredentials(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	require.NoError(t, err)
	repo := &mocks.UserRepositoryMock{}
	repo.On("GetByEmail", mock.Anything, "player@example.com").
		Return(&models.User{Password: string(passwordHash)}, nil).Once()
	repo.On("GetByEmail", mock.Anything, "missing@example.com").
		Return((*models.User)(nil), errors.New("not found")).Once()
	service, err := services.NewUserService(repo, &mocks.RefreshTokenStoreMock{})
	require.NoError(t, err)

	token, err := service.Login(context.Background(), "player@example.com", "wrong")
	assert.Nil(t, token)
	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
	token, err = service.Login(context.Background(), "missing@example.com", "password")
	assert.Nil(t, token)
	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
	repo.AssertExpectations(t)
}

func TestUserService_Login_RefreshSessionStoreError(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &models.User{UUID: uuid.New(), Email: "player@example.com", Password: string(passwordHash)}
	repo := &mocks.UserRepositoryMock{}
	repo.On("GetByEmail", mock.Anything, user.Email).Return(user, nil).Once()
	refreshStore := &mocks.RefreshTokenStoreMock{}
	storeErr := errors.New("redis unavailable")
	refreshStore.On("SaveRefreshToken", mock.Anything, mock.Anything, user.UUID, 30*24*time.Hour).
		Return(storeErr).Once()
	service, err := services.NewUserService(repo, refreshStore)
	require.NoError(t, err)

	tokens, err := service.Login(context.Background(), user.Email, "password")
	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, storeErr)
	repo.AssertExpectations(t)
	refreshStore.AssertExpectations(t)
}

func TestUserService_Refresh_RotatesRefreshToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	ctx, userID := context.Background(), uuid.New()
	oldTokenID := uuid.NewString()
	refreshStore := &mocks.RefreshTokenStoreMock{}
	refreshStore.On("ConsumeRefreshToken", ctx, oldTokenID).Return(userID, nil).Once()
	refreshStore.On("SaveRefreshToken", ctx, mock.MatchedBy(func(tokenID string) bool {
		return tokenID != oldTokenID
	}), userID, 30*24*time.Hour).Return(nil).Once()
	service, err := services.NewUserService(&mocks.UserRepositoryMock{}, refreshStore)
	require.NoError(t, err)
	oldRefresh := signedRefreshToken(t, userID, oldTokenID)

	tokens, err := service.Refresh(ctx, oldRefresh)
	require.NoError(t, err)
	require.NotNil(t, tokens)
	accessClaims, err := auth.ParseToken(tokens.AccessToken, "test-secret")
	require.NoError(t, err)
	assert.Equal(t, auth.AccessTokenType, accessClaims.Type)
	assert.Equal(t, userID.String(), accessClaims.Subject)
	refreshClaims, err := auth.ParseToken(tokens.RefreshToken, "test-secret")
	require.NoError(t, err)
	assert.Equal(t, auth.RefreshTokenType, refreshClaims.Type)
	assert.NotEqual(t, oldTokenID, refreshClaims.ID)
	refreshStore.AssertExpectations(t)
}

func TestUserService_Refresh_RejectsConsumedToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	ctx, userID := context.Background(), uuid.New()
	tokenID := uuid.NewString()
	refreshStore := &mocks.RefreshTokenStoreMock{}
	refreshStore.On("ConsumeRefreshToken", ctx, tokenID).
		Return(uuid.Nil, interfaces.ErrRefreshTokenNotFound).Once()
	service, err := services.NewUserService(&mocks.UserRepositoryMock{}, refreshStore)
	require.NoError(t, err)

	tokens, err := service.Refresh(ctx, signedRefreshToken(t, userID, tokenID))
	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, services.ErrInvalidRefreshToken)
	refreshStore.AssertExpectations(t)
}

func TestUserService_Logout_RevokesRefreshToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	ctx, userID := context.Background(), uuid.New()
	tokenID := uuid.NewString()
	refreshStore := &mocks.RefreshTokenStoreMock{}
	refreshStore.On("ConsumeRefreshToken", ctx, tokenID).Return(userID, nil).Once()
	service, err := services.NewUserService(&mocks.UserRepositoryMock{}, refreshStore)
	require.NoError(t, err)

	require.NoError(t, service.Logout(ctx, signedRefreshToken(t, userID, tokenID)))
	refreshStore.AssertExpectations(t)
}

func TestUserService_RejectsAccessTokenForRefresh(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	userID := uuid.New()
	refreshStore := &mocks.RefreshTokenStoreMock{}
	service, err := services.NewUserService(&mocks.UserRepositoryMock{}, refreshStore)
	require.NoError(t, err)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID.String(), ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Type: auth.AccessTokenType,
	}).SignedString([]byte("test-secret"))
	require.NoError(t, err)

	tokens, err := service.Refresh(context.Background(), token)
	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, services.ErrInvalidRefreshToken)
	refreshStore.AssertNotCalled(t, "ConsumeRefreshToken", mock.Anything, mock.Anything)
}

func signedRefreshToken(t *testing.T, userID uuid.UUID, tokenID string) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID: tokenID, Subject: userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
		},
		Type: auth.RefreshTokenType,
	}).SignedString([]byte("test-secret"))
	require.NoError(t, err)
	return token
}

func TestUserService_GetUserByID(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	ctx := context.Background()
	id := uuid.New()
	want := &models.User{UUID: id}
	repo := &mocks.UserRepositoryMock{}
	repo.On("GetByID", ctx, id).Return(want, nil).Once()
	service, err := services.NewUserService(repo, &mocks.RefreshTokenStoreMock{})
	require.NoError(t, err)

	got, err := service.GetUserByID(ctx, id)
	require.NoError(t, err)
	assert.Same(t, want, got)
	repo.AssertExpectations(t)
}
