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

func TestOpponentService_CreateOpponent(t *testing.T) {
	opponent := &models.Opponent{Name: "Opponent"}
	repo := &mocks.OpponentRepositoryMock{}
	repo.On("Create", mock.Anything, opponent).Return(nil).Once()

	require.NoError(t, services.NewOpponentService(repo).CreateOpponent(context.Background(), opponent))
	assert.NotEqual(t, uuid.Nil, opponent.UUID)
	repo.AssertExpectations(t)
}

func TestOpponentService_UpdateOpponent(t *testing.T) {
	t.Run("delegates updates", func(t *testing.T) {
		opponent := &models.Opponent{UUID: uuid.New(), Name: "Updated"}
		repo := &mocks.OpponentRepositoryMock{}
		repo.On("Update", mock.Anything, opponent).Return(nil).Once()

		require.NoError(t, services.NewOpponentService(repo).UpdateOpponent(context.Background(), opponent))
		repo.AssertExpectations(t)
	})
	t.Run("propagates repository errors", func(t *testing.T) {
		opponent := &models.Opponent{UUID: uuid.New(), Name: "Updated"}
		wantErr := errors.New("update failed")
		repo := &mocks.OpponentRepositoryMock{}
		repo.On("Update", mock.Anything, opponent).Return(wantErr).Once()

		err := services.NewOpponentService(repo).UpdateOpponent(context.Background(), opponent)

		assert.ErrorIs(t, err, wantErr)
		repo.AssertExpectations(t)
	})
}

func TestOpponentService_DeleteOpponent(t *testing.T) {
	t.Run("delegates soft deletes with owner", func(t *testing.T) {
		id, userID := uuid.New(), uuid.New()
		repo := &mocks.OpponentRepositoryMock{}
		repo.On("SoftDelete", mock.Anything, id, userID).Return(nil).Once()

		require.NoError(t, services.NewOpponentService(repo).DeleteOpponent(context.Background(), id, userID))
		repo.AssertExpectations(t)
	})
	t.Run("propagates repository errors", func(t *testing.T) {
		id, userID := uuid.New(), uuid.New()
		wantErr := errors.New("soft delete failed")
		repo := &mocks.OpponentRepositoryMock{}
		repo.On("SoftDelete", mock.Anything, id, userID).Return(wantErr).Once()

		err := services.NewOpponentService(repo).DeleteOpponent(context.Background(), id, userID)

		assert.ErrorIs(t, err, wantErr)
		repo.AssertExpectations(t)
	})
}

func TestOpponentService_GetOpponentByID(t *testing.T) {
	id := uuid.New()
	want := &models.Opponent{UUID: id, Name: "Opponent"}
	repo := &mocks.OpponentRepositoryMock{}
	repo.On("GetByID", mock.Anything, id).Return(want, nil).Once()

	got, err := services.NewOpponentService(repo).GetOpponentByID(context.Background(), id)

	require.NoError(t, err)
	assert.Same(t, want, got)
	repo.AssertExpectations(t)
}

func TestOpponentService_ListUserOpponents(t *testing.T) {
	userID := uuid.New()
	want := []*models.Opponent{{Name: "Opponent"}}
	repo := &mocks.OpponentRepositoryMock{}
	repo.On("ListByUser", mock.Anything, userID).Return(want, nil).Once()

	got, err := services.NewOpponentService(repo).ListUserOpponents(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, want, got)
	repo.AssertExpectations(t)
}
