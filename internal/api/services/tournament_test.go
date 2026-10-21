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

func TestTournamentService_CreateTournament(t *testing.T) {
	tournament := &models.Tournament{Name: "Regional"}
	repo := &mocks.TournamentRepositoryMock{}
	repo.On("Create", mock.Anything, tournament).Return(nil).Once()

	require.NoError(t, services.NewTournamentService(repo).CreateTournament(context.Background(), tournament))
	assert.NotEqual(t, uuid.Nil, tournament.UUID)
	repo.AssertExpectations(t)
}

func TestTournamentService_UpdateTournament(t *testing.T) {
	t.Run("delegates updates", func(t *testing.T) {
		tournament := &models.Tournament{UUID: uuid.New(), Name: "Updated"}
		repo := &mocks.TournamentRepositoryMock{}
		repo.On("Update", mock.Anything, tournament).Return(nil).Once()

		require.NoError(t, services.NewTournamentService(repo).UpdateTournament(context.Background(), tournament))
		repo.AssertExpectations(t)
	})
	t.Run("propagates repository errors", func(t *testing.T) {
		tournament := &models.Tournament{UUID: uuid.New(), Name: "Updated"}
		wantErr := errors.New("update failed")
		repo := &mocks.TournamentRepositoryMock{}
		repo.On("Update", mock.Anything, tournament).Return(wantErr).Once()

		err := services.NewTournamentService(repo).UpdateTournament(context.Background(), tournament)

		assert.ErrorIs(t, err, wantErr)
		repo.AssertExpectations(t)
	})
}

func TestTournamentService_DeleteTournament(t *testing.T) {
	t.Run("delegates soft deletes with owner", func(t *testing.T) {
		id, userID := uuid.New(), uuid.New()
		repo := &mocks.TournamentRepositoryMock{}
		repo.On("SoftDelete", mock.Anything, id, userID).Return(nil).Once()

		require.NoError(t, services.NewTournamentService(repo).DeleteTournament(context.Background(), id, userID))
		repo.AssertExpectations(t)
	})
	t.Run("propagates repository errors", func(t *testing.T) {
		id, userID := uuid.New(), uuid.New()
		wantErr := errors.New("soft delete failed")
		repo := &mocks.TournamentRepositoryMock{}
		repo.On("SoftDelete", mock.Anything, id, userID).Return(wantErr).Once()

		err := services.NewTournamentService(repo).DeleteTournament(context.Background(), id, userID)

		assert.ErrorIs(t, err, wantErr)
		repo.AssertExpectations(t)
	})
}

func TestTournamentService_GetTournamentByID(t *testing.T) {
	id := uuid.New()
	want := &models.Tournament{UUID: id, Name: "Regional"}
	repo := &mocks.TournamentRepositoryMock{}
	repo.On("GetByID", mock.Anything, id).Return(want, nil).Once()

	got, err := services.NewTournamentService(repo).GetTournamentByID(context.Background(), id)

	require.NoError(t, err)
	assert.Same(t, want, got)
	repo.AssertExpectations(t)
}

func TestTournamentService_ListUserTournaments(t *testing.T) {
	userID := uuid.New()
	want := []*models.Tournament{{Name: "Regional"}}
	repo := &mocks.TournamentRepositoryMock{}
	repo.On("ListByUser", mock.Anything, userID).Return(want, nil).Once()

	got, err := services.NewTournamentService(repo).ListUserTournaments(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, want, got)
	repo.AssertExpectations(t)
}
