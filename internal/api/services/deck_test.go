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

func TestDeckService_CreateDeck(t *testing.T) {
	deck := &models.Deck{Name: "Control"}
	repo := &mocks.DeckRepositoryMock{}
	repo.On("Create", mock.Anything, deck).Return(nil).Once()

	require.NoError(t, services.NewDeckService(repo).CreateDeck(context.Background(), deck))
	assert.NotEqual(t, uuid.Nil, deck.UUID)
	repo.AssertExpectations(t)
}

func TestDeckService_UpdateDeck(t *testing.T) {
	t.Run("delegates updates", func(t *testing.T) {
		deck := &models.Deck{UUID: uuid.New(), Name: "Updated"}
		repo := &mocks.DeckRepositoryMock{}
		repo.On("Update", mock.Anything, deck).Return(nil).Once()

		require.NoError(t, services.NewDeckService(repo).UpdateDeck(context.Background(), deck))
		repo.AssertExpectations(t)
	})
	t.Run("propagates repository errors", func(t *testing.T) {
		deck := &models.Deck{UUID: uuid.New(), Name: "Updated"}
		wantErr := errors.New("update failed")
		repo := &mocks.DeckRepositoryMock{}
		repo.On("Update", mock.Anything, deck).Return(wantErr).Once()

		err := services.NewDeckService(repo).UpdateDeck(context.Background(), deck)

		assert.ErrorIs(t, err, wantErr)
		repo.AssertExpectations(t)
	})
}

func TestDeckService_DeleteDeck(t *testing.T) {
	t.Run("delegates soft deletes with owner", func(t *testing.T) {
		id, userID := uuid.New(), uuid.New()
		repo := &mocks.DeckRepositoryMock{}
		repo.On("SoftDelete", mock.Anything, id, userID).Return(nil).Once()

		require.NoError(t, services.NewDeckService(repo).DeleteDeck(context.Background(), id, userID))
		repo.AssertExpectations(t)
	})
	t.Run("propagates repository errors", func(t *testing.T) {
		id, userID := uuid.New(), uuid.New()
		wantErr := errors.New("soft delete failed")
		repo := &mocks.DeckRepositoryMock{}
		repo.On("SoftDelete", mock.Anything, id, userID).Return(wantErr).Once()

		err := services.NewDeckService(repo).DeleteDeck(context.Background(), id, userID)

		assert.ErrorIs(t, err, wantErr)
		repo.AssertExpectations(t)
	})
}

func TestDeckService_GetDeckByID(t *testing.T) {
	id := uuid.New()
	want := &models.Deck{UUID: id}
	repo := &mocks.DeckRepositoryMock{}
	repo.On("GetByID", mock.Anything, id).Return(want, nil).Once()

	got, err := services.NewDeckService(repo).GetDeckByID(context.Background(), id)

	require.NoError(t, err)
	assert.Same(t, want, got)
	repo.AssertExpectations(t)
}

func TestDeckService_ListUserDecks(t *testing.T) {
	userID, game := uuid.New(), models.GamePokemon
	want := []*models.Deck{{Name: "Tempo"}}
	repo := &mocks.DeckRepositoryMock{}
	repo.On("ListByUser", mock.Anything, userID, &game).Return(want, nil).Once()

	got, err := services.NewDeckService(repo).ListUserDecks(context.Background(), userID, &game)

	require.NoError(t, err)
	assert.Equal(t, want, got)
	repo.AssertExpectations(t)
}
