package controllers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/andresMTG/tcgstats-backend/internal/api/controllers"
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeckController_CreateDeck(t *testing.T) {
	t.Run("requires user identity", func(t *testing.T) {
		service := &mocks.DeckServiceMock{}
		response := recorder()
		controllers.NewDeckController(service).CreateDeck(response, newRequest(http.MethodPost, "/decks", "{}"))
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		service.AssertNotCalled(t, "CreateDeck", mock.Anything, mock.Anything)
	})
	t.Run("rejects malformed JSON", func(t *testing.T) {
		service := &mocks.DeckServiceMock{}
		request := withUserID(newRequest(http.MethodPost, "/decks", "{"), uuid.New().String())
		response := recorder()
		controllers.NewDeckController(service).CreateDeck(response, request)
		assert.Equal(t, http.StatusBadRequest, response.Code)
		service.AssertNotCalled(t, "CreateDeck", mock.Anything, mock.Anything)
	})
	t.Run("creates deck for authenticated user", func(t *testing.T) {
		service := &mocks.DeckServiceMock{}
		userID := uuid.New()
		service.On("CreateDeck", mock.Anything, mock.MatchedBy(func(deck *models.Deck) bool {
			return deck.Name == "Tempo" && deck.Game == models.GamePokemon &&
				deck.Decklist == "cards" && deck.UserUUID == userID
		})).Return(nil).Once()
		request := withUserID(newRequest(http.MethodPost, "/decks",
			`{"name":"Tempo","game":"pokemon","decklist":"cards"}`), userID.String())
		response := recorder()
		controllers.NewDeckController(service).CreateDeck(response, request)
		require.Equal(t, http.StatusCreated, response.Code)
		var got models.Deck
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		assert.Equal(t, userID, got.UserUUID)
		service.AssertExpectations(t)
	})
	t.Run("reports service failure", func(t *testing.T) {
		service := &mocks.DeckServiceMock{}
		service.On("CreateDeck", mock.Anything, mock.Anything).Return(errors.New("write failed")).Once()
		request := withUserID(newRequest(http.MethodPost, "/decks", `{"name":"Tempo"}`), uuid.New().String())
		response := recorder()
		controllers.NewDeckController(service).CreateDeck(response, request)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		service.AssertExpectations(t)
	})
}

func TestDeckController_GetDeck(t *testing.T) {
	t.Run("rejects invalid UUID", func(t *testing.T) {
		service := &mocks.DeckServiceMock{}
		response := recorder()
		controllers.NewDeckController(service).GetDeck(response,
			withRouteParam(newRequest(http.MethodGet, "/decks/invalid", ""), "id", "invalid"))
		assert.Equal(t, http.StatusBadRequest, response.Code)
		service.AssertNotCalled(t, "GetDeckByID", mock.Anything, mock.Anything)
	})
	t.Run("returns deck", func(t *testing.T) {
		service := &mocks.DeckServiceMock{}
		id := uuid.New()
		want := &models.Deck{UUID: id, Name: "Tempo"}
		service.On("GetDeckByID", mock.Anything, id).Return(want, nil).Once()
		response := recorder()
		controllers.NewDeckController(service).GetDeck(response,
			withRouteParam(newRequest(http.MethodGet, "/decks/"+id.String(), ""), "id", id.String()))
		require.Equal(t, http.StatusOK, response.Code)
		var got models.Deck
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		assert.Equal(t, want, &got)
		service.AssertExpectations(t)
	})
	t.Run("returns not found on service failure", func(t *testing.T) {
		service := &mocks.DeckServiceMock{}
		id := uuid.New()
		service.On("GetDeckByID", mock.Anything, id).Return((*models.Deck)(nil), errors.New("missing")).Once()
		response := recorder()
		controllers.NewDeckController(service).GetDeck(response,
			withRouteParam(newRequest(http.MethodGet, "/decks/"+id.String(), ""), "id", id.String()))
		assert.Equal(t, http.StatusNotFound, response.Code)
		service.AssertExpectations(t)
	})
}

func TestDeckController_ListUserDecks(t *testing.T) {
	t.Run("requires user identity", func(t *testing.T) {
		service := &mocks.DeckServiceMock{}
		response := recorder()
		controllers.NewDeckController(service).ListUserDecks(response, newRequest(http.MethodGet, "/decks", ""))
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		service.AssertNotCalled(t, "ListUserDecks", mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("forwards optional game filter", func(t *testing.T) {
		service := &mocks.DeckServiceMock{}
		userID, game := uuid.New(), models.GamePokemon
		want := []*models.Deck{{Name: "Tempo", Game: game}}
		service.On("ListUserDecks", mock.Anything, userID, &game).Return(want, nil).Once()
		request := withUserID(newRequest(http.MethodGet, "/decks?game=pokemon", ""), userID.String())
		response := recorder()
		controllers.NewDeckController(service).ListUserDecks(response, request)
		require.Equal(t, http.StatusOK, response.Code)
		var got []*models.Deck
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		require.Len(t, got, 1)
		assert.Equal(t, "Tempo", got[0].Name)
		service.AssertExpectations(t)
	})
	t.Run("reports service failure", func(t *testing.T) {
		service := &mocks.DeckServiceMock{}
		userID := uuid.New()
		service.On("ListUserDecks", mock.Anything, userID, (*models.GameName)(nil)).
			Return(([]*models.Deck)(nil), errors.New("query failed")).Once()
		request := withUserID(newRequest(http.MethodGet, "/decks", ""), userID.String())
		response := recorder()
		controllers.NewDeckController(service).ListUserDecks(response, request)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		service.AssertExpectations(t)
	})
}
