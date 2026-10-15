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

func TestDeckService(t *testing.T) {
	t.Run("CreateDeck assigns UUID and propagates errors", func(t *testing.T) {
		deck := &models.Deck{Name: "Control"}
		repo := &mocks.DeckRepositoryMock{}
		repo.On("Create", mock.Anything, deck).Return(nil).Once()
		service := services.NewDeckService(repo)

		err := service.CreateDeck(context.Background(), deck)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, deck.UUID)
		repo.AssertExpectations(t)
	})
	t.Run("CreateDeck returns repository error", func(t *testing.T) {
		deck := &models.Deck{Name: "Control"}
		wantErr := errors.New("write failed")
		repo := &mocks.DeckRepositoryMock{}
		repo.On("Create", mock.Anything, deck).Return(wantErr).Once()
		assert.ErrorIs(t, services.NewDeckService(repo).CreateDeck(context.Background(), deck), wantErr)
		repo.AssertExpectations(t)
	})
	t.Run("GetDeckByID delegates", func(t *testing.T) {
		id, wantErr := uuid.New(), errors.New("not found")
		repo := &mocks.DeckRepositoryMock{}
		repo.On("GetByID", mock.Anything, id).Return((*models.Deck)(nil), wantErr).Once()
		got, err := services.NewDeckService(repo).GetDeckByID(context.Background(), id)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, wantErr)
		repo.AssertExpectations(t)
	})
	t.Run("ListUserDecks forwards filter", func(t *testing.T) {
		userID := uuid.New()
		game := models.GamePokemon
		want := []*models.Deck{{Name: "Tempo"}}
		repo := &mocks.DeckRepositoryMock{}
		repo.On("ListByUser", mock.Anything, userID, &game).Return(want, nil).Once()
		got, err := services.NewDeckService(repo).ListUserDecks(context.Background(), userID, &game)
		require.NoError(t, err)
		assert.Same(t, want[0], got[0])
		repo.AssertExpectations(t)
	})
}

func TestTournamentService(t *testing.T) {
	t.Run("CreateTournament assigns UUID", func(t *testing.T) {
		tournament := &models.Tournament{Name: "Regional"}
		repo := &mocks.TournamentRepositoryMock{}
		repo.On("Create", mock.Anything, tournament).Return(nil).Once()
		require.NoError(t, services.NewTournamentService(repo).CreateTournament(context.Background(), tournament))
		assert.NotEqual(t, uuid.Nil, tournament.UUID)
		repo.AssertExpectations(t)
	})
	t.Run("CreateTournament returns repository error", func(t *testing.T) {
		tournament := &models.Tournament{Name: "Regional"}
		wantErr := errors.New("write failed")
		repo := &mocks.TournamentRepositoryMock{}
		repo.On("Create", mock.Anything, tournament).Return(wantErr).Once()
		assert.ErrorIs(t, services.NewTournamentService(repo).CreateTournament(context.Background(), tournament), wantErr)
		repo.AssertExpectations(t)
	})
	t.Run("GetTournamentByID delegates", func(t *testing.T) {
		id, want := uuid.New(), &models.Tournament{Name: "Regional"}
		repo := &mocks.TournamentRepositoryMock{}
		repo.On("GetByID", mock.Anything, id).Return(want, nil).Once()
		got, err := services.NewTournamentService(repo).GetTournamentByID(context.Background(), id)
		require.NoError(t, err)
		assert.Same(t, want, got)
		repo.AssertExpectations(t)
	})
	t.Run("ListUserTournaments delegates", func(t *testing.T) {
		userID := uuid.New()
		want := []*models.Tournament{{Name: "Regional"}}
		repo := &mocks.TournamentRepositoryMock{}
		repo.On("ListByUser", mock.Anything, userID).Return(want, nil).Once()
		got, err := services.NewTournamentService(repo).ListUserTournaments(context.Background(), userID)
		require.NoError(t, err)
		assert.Same(t, want[0], got[0])
		repo.AssertExpectations(t)
	})
}

func TestOpponentService(t *testing.T) {
	t.Run("CreateOpponent assigns UUID", func(t *testing.T) {
		opponent := &models.Opponent{Name: "Opponent"}
		repo := &mocks.OpponentRepositoryMock{}
		repo.On("Create", mock.Anything, opponent).Return(nil).Once()
		require.NoError(t, services.NewOpponentService(repo).CreateOpponent(context.Background(), opponent))
		assert.NotEqual(t, uuid.Nil, opponent.UUID)
		repo.AssertExpectations(t)
	})
	t.Run("CreateOpponent returns repository error", func(t *testing.T) {
		opponent := &models.Opponent{Name: "Opponent"}
		wantErr := errors.New("write failed")
		repo := &mocks.OpponentRepositoryMock{}
		repo.On("Create", mock.Anything, opponent).Return(wantErr).Once()
		assert.ErrorIs(t, services.NewOpponentService(repo).CreateOpponent(context.Background(), opponent), wantErr)
		repo.AssertExpectations(t)
	})
	t.Run("GetOpponentByID delegates", func(t *testing.T) {
		id, want := uuid.New(), &models.Opponent{Name: "Opponent"}
		repo := &mocks.OpponentRepositoryMock{}
		repo.On("GetByID", mock.Anything, id).Return(want, nil).Once()
		got, err := services.NewOpponentService(repo).GetOpponentByID(context.Background(), id)
		require.NoError(t, err)
		assert.Same(t, want, got)
		repo.AssertExpectations(t)
	})
	t.Run("ListUserOpponents delegates", func(t *testing.T) {
		userID := uuid.New()
		want := []*models.Opponent{{Name: "Opponent"}}
		repo := &mocks.OpponentRepositoryMock{}
		repo.On("ListByUser", mock.Anything, userID).Return(want, nil).Once()
		got, err := services.NewOpponentService(repo).ListUserOpponents(context.Background(), userID)
		require.NoError(t, err)
		assert.Same(t, want[0], got[0])
		repo.AssertExpectations(t)
	})
}

func TestMatchService(t *testing.T) {
	t.Run("CreateMatch assigns UUID", func(t *testing.T) {
		match := &models.Match{Result: models.ResultWin}
		repo := &mocks.MatchRepositoryMock{}
		repo.On("Create", mock.Anything, match).Return(nil).Once()
		require.NoError(t, services.NewMatchService(repo).CreateMatch(context.Background(), match))
		assert.NotEqual(t, uuid.Nil, match.UUID)
		repo.AssertExpectations(t)
	})
	t.Run("CreateMatch returns repository error", func(t *testing.T) {
		match := &models.Match{Result: models.ResultWin}
		wantErr := errors.New("write failed")
		repo := &mocks.MatchRepositoryMock{}
		repo.On("Create", mock.Anything, match).Return(wantErr).Once()
		assert.ErrorIs(t, services.NewMatchService(repo).CreateMatch(context.Background(), match), wantErr)
		repo.AssertExpectations(t)
	})
	t.Run("GetMatchByID delegates", func(t *testing.T) {
		id, want := uuid.New(), &models.Match{Result: models.ResultWin}
		repo := &mocks.MatchRepositoryMock{}
		repo.On("GetByID", mock.Anything, id).Return(want, nil).Once()
		got, err := services.NewMatchService(repo).GetMatchByID(context.Background(), id)
		require.NoError(t, err)
		assert.Same(t, want, got)
		repo.AssertExpectations(t)
	})
	t.Run("ListTournamentMatches forwards filter", func(t *testing.T) {
		tournamentID := uuid.New()
		result := models.ResultWin
		want := []*models.Match{{Result: result}}
		repo := &mocks.MatchRepositoryMock{}
		repo.On("ListByTournament", mock.Anything, tournamentID, &result).Return(want, nil).Once()
		got, err := services.NewMatchService(repo).ListTournamentMatches(context.Background(), tournamentID, &result)
		require.NoError(t, err)
		assert.Same(t, want[0], got[0])
		repo.AssertExpectations(t)
	})
	t.Run("ListDeckMatches forwards filter", func(t *testing.T) {
		deckID := uuid.New()
		result := models.ResultLose
		want := []*models.Match{{Result: result}}
		repo := &mocks.MatchRepositoryMock{}
		repo.On("ListByDeck", mock.Anything, deckID, &result).Return(want, nil).Once()
		got, err := services.NewMatchService(repo).ListDeckMatches(context.Background(), deckID, &result)
		require.NoError(t, err)
		assert.Same(t, want[0], got[0])
		repo.AssertExpectations(t)
	})
}
