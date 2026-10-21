package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/services"
	"github.com/andresMTG/tcgstats-backend/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMatchService_CreateMatch(t *testing.T) {
	match := &models.Match{Result: models.ResultWin}
	repo := &mocks.MatchRepositoryMock{}
	repo.On("Create", mock.Anything, match).Return(nil).Once()

	require.NoError(t, services.NewMatchService(repo).CreateMatch(context.Background(), match))
	assert.NotEqual(t, uuid.Nil, match.UUID)
	repo.AssertExpectations(t)
}

func TestMatchService_UpdateMatch(t *testing.T) {
	t.Run("delegates updates with owner", func(t *testing.T) {
		match, userID := &models.Match{UUID: uuid.New(), Result: models.ResultLose}, uuid.New()
		repo := &mocks.MatchRepositoryMock{}
		repo.On("Update", mock.Anything, match, userID).Return(nil).Once()

		require.NoError(t, services.NewMatchService(repo).UpdateMatch(context.Background(), match, userID))
		repo.AssertExpectations(t)
	})
	t.Run("propagates repository errors", func(t *testing.T) {
		match := &models.Match{UUID: uuid.New(), Result: models.ResultLose}
		userID := uuid.New()
		wantErr := errors.New("update failed")
		repo := &mocks.MatchRepositoryMock{}
		repo.On("Update", mock.Anything, match, userID).Return(wantErr).Once()

		err := services.NewMatchService(repo).UpdateMatch(context.Background(), match, userID)

		assert.ErrorIs(t, err, wantErr)
		repo.AssertExpectations(t)
	})
}

func TestMatchService_DeleteMatch(t *testing.T) {
	t.Run("delegates soft deletes with owner", func(t *testing.T) {
		id, userID := uuid.New(), uuid.New()
		repo := &mocks.MatchRepositoryMock{}
		repo.On("SoftDelete", mock.Anything, id, userID).Return(nil).Once()

		require.NoError(t, services.NewMatchService(repo).DeleteMatch(context.Background(), id, userID))
		repo.AssertExpectations(t)
	})
	t.Run("propagates repository errors", func(t *testing.T) {
		id, userID := uuid.New(), uuid.New()
		wantErr := errors.New("soft delete failed")
		repo := &mocks.MatchRepositoryMock{}
		repo.On("SoftDelete", mock.Anything, id, userID).Return(wantErr).Once()

		err := services.NewMatchService(repo).DeleteMatch(context.Background(), id, userID)

		assert.ErrorIs(t, err, wantErr)
		repo.AssertExpectations(t)
	})
}

func TestMatchService_GetMatchByID(t *testing.T) {
	id := uuid.New()
	want := &models.Match{UUID: id, Result: models.ResultWin}
	repo := &mocks.MatchRepositoryMock{}
	repo.On("GetByID", mock.Anything, id).Return(want, nil).Once()

	got, err := services.NewMatchService(repo).GetMatchByID(context.Background(), id)

	require.NoError(t, err)
	assert.Same(t, want, got)
	repo.AssertExpectations(t)
}

func TestMatchService_ListTournamentMatches(t *testing.T) {
	tournamentID := uuid.New()
	result := models.ResultWin
	want := []*models.Match{{Result: result}}
	repo := &mocks.MatchRepositoryMock{}
	repo.On("ListByTournament", mock.Anything, tournamentID, &result).Return(want, nil).Once()

	got, err := services.NewMatchService(repo).ListTournamentMatches(context.Background(), tournamentID, &result)

	require.NoError(t, err)
	assert.Equal(t, want, got)
	repo.AssertExpectations(t)
}

func TestMatchService_ListDeckMatches(t *testing.T) {
	deckID := uuid.New()
	result := models.ResultLose
	want := []*models.Match{{Result: result}}
	repo := &mocks.MatchRepositoryMock{}
	repo.On("ListByDeck", mock.Anything, deckID, &result).Return(want, nil).Once()

	got, err := services.NewMatchService(repo).ListDeckMatches(context.Background(), deckID, &result)

	require.NoError(t, err)
	assert.Equal(t, want, got)
	repo.AssertExpectations(t)
}
