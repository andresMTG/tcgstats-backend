package controllers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/controllers"
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/andresMTG/tcgstats-backend/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeckController_UpdateAndDeleteDeck(t *testing.T) {
	service := &mocks.DeckServiceMock{}
	userID, deckID := uuid.New(), uuid.New()
	service.On("UpdateDeck", mock.Anything, mock.MatchedBy(func(deck *models.Deck) bool {
		return deck.UUID == deckID && deck.UserUUID == userID &&
			deck.Name == "Updated" && deck.Game == models.GamePokemon && deck.Decklist == "new list"
	})).Return(nil).Once()
	request := withUserID(withRouteParam(
		newRequest(http.MethodPut, "/decks/"+deckID.String(), `{"name":"Updated","game":"pokemon","decklist":"new list"}`),
		"id",
		deckID.String(),
	), userID.String())
	response := recorder()
	controllers.NewDeckController(service).UpdateDeck(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var updated models.Deck
	require.NoError(t, json.NewDecoder(response.Body).Decode(&updated))
	assert.Equal(t, deckID, updated.UUID)

	service.On("DeleteDeck", mock.Anything, deckID, userID).Return(nil).Once()
	request = withUserID(withRouteParam(
		newRequest(http.MethodDelete, "/decks/"+deckID.String(), ""),
		"id",
		deckID.String(),
	), userID.String())
	response = recorder()
	controllers.NewDeckController(service).DeleteDeck(response, request)
	assert.Equal(t, http.StatusNoContent, response.Code)
	service.AssertExpectations(t)
}

func TestTournamentController_UpdateAndDeleteTournament(t *testing.T) {
	service := &mocks.TournamentServiceMock{}
	userID, tournamentID, deckID := uuid.New(), uuid.New(), uuid.New()
	date := time.Date(2026, time.October, 15, 0, 0, 0, 0, time.UTC)
	service.On("UpdateTournament", mock.Anything, mock.MatchedBy(func(tournament *models.Tournament) bool {
		return tournament.UUID == tournamentID && tournament.UserUUID == userID &&
			tournament.Name == "Updated event" && tournament.Date.Equal(date) &&
			tournament.Game == models.GamePokemon && tournament.Scope == models.ScopeMajor &&
			tournament.IsOfficialCircuit && tournament.DeckUUID != nil && *tournament.DeckUUID == deckID
	})).Return(nil).Once()
	body := `{"name":"Updated event","date":"2026-10-15T00:00:00Z","game":"pokemon","scope":"MAJOR","is_official_circuit":true,"deck_uuid":"` + deckID.String() + `"}`
	request := withUserID(withRouteParam(
		newRequest(http.MethodPut, "/tournaments/"+tournamentID.String(), body),
		"id",
		tournamentID.String(),
	), userID.String())
	response := recorder()
	controllers.NewTournamentController(service).UpdateTournament(response, request)
	require.Equal(t, http.StatusOK, response.Code)

	service.On("DeleteTournament", mock.Anything, tournamentID, userID).Return(nil).Once()
	request = withUserID(withRouteParam(
		newRequest(http.MethodDelete, "/tournaments/"+tournamentID.String(), ""),
		"id",
		tournamentID.String(),
	), userID.String())
	response = recorder()
	controllers.NewTournamentController(service).DeleteTournament(response, request)
	assert.Equal(t, http.StatusNoContent, response.Code)
	service.AssertExpectations(t)
}

func TestOpponentController_UpdateAndDeleteOpponent(t *testing.T) {
	service := &mocks.OpponentServiceMock{}
	userID, opponentID := uuid.New(), uuid.New()
	service.On("UpdateOpponent", mock.Anything, mock.MatchedBy(func(opponent *models.Opponent) bool {
		return opponent.UUID == opponentID && opponent.UserUUID == userID && opponent.Name == "Updated opponent"
	})).Return(nil).Once()
	request := withUserID(withRouteParam(
		newRequest(http.MethodPut, "/opponents/"+opponentID.String(), `{"name":"Updated opponent"}`),
		"id",
		opponentID.String(),
	), userID.String())
	response := recorder()
	controllers.NewOpponentController(service).UpdateOpponent(response, request)
	require.Equal(t, http.StatusOK, response.Code)

	service.On("DeleteOpponent", mock.Anything, opponentID, userID).Return(nil).Once()
	request = withUserID(withRouteParam(
		newRequest(http.MethodDelete, "/opponents/"+opponentID.String(), ""),
		"id",
		opponentID.String(),
	), userID.String())
	response = recorder()
	controllers.NewOpponentController(service).DeleteOpponent(response, request)
	assert.Equal(t, http.StatusNoContent, response.Code)
	service.AssertExpectations(t)
}

func TestMatchController_UpdateAndDeleteMatch(t *testing.T) {
	service := &mocks.MatchServiceMock{}
	userID, matchID, tournamentID, opponentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	service.On("UpdateMatch", mock.Anything, mock.MatchedBy(func(match *models.Match) bool {
		return match.UUID == matchID && match.Result == models.ResultLose &&
			match.RoundNumber == 5 && match.Phase == models.PhaseTop4 &&
			match.OpposingDeckArchetype == "Control" &&
			match.TournamentUUID == tournamentID && match.OpponentUUID == opponentID
	}), userID).Return(nil).Once()
	body := `{"result":"lose","round_number":5,"phase":"TOP_4","opposing_deck_archetype":"Control","tournament_uuid":"` +
		tournamentID.String() + `","opponent_uuid":"` + opponentID.String() + `"}`
	request := withUserID(withRouteParam(
		newRequest(http.MethodPut, "/matches/"+matchID.String(), body),
		"id",
		matchID.String(),
	), userID.String())
	response := recorder()
	controllers.NewMatchController(service).UpdateMatch(response, request)
	require.Equal(t, http.StatusOK, response.Code)

	service.On("DeleteMatch", mock.Anything, matchID, userID).Return(nil).Once()
	request = withUserID(withRouteParam(
		newRequest(http.MethodDelete, "/matches/"+matchID.String(), ""),
		"id",
		matchID.String(),
	), userID.String())
	response = recorder()
	controllers.NewMatchController(service).DeleteMatch(response, request)
	assert.Equal(t, http.StatusNoContent, response.Code)
	service.AssertExpectations(t)
}

func TestDeckController_UpdateDeckRejectsUnownedRecord(t *testing.T) {
	service := &mocks.DeckServiceMock{}
	userID, deckID := uuid.New(), uuid.New()
	service.On("UpdateDeck", mock.Anything, mock.Anything).Return(interfaces.ErrRecordNotFound).Once()
	request := withUserID(withRouteParam(
		newRequest(http.MethodPut, "/decks/"+deckID.String(), `{"name":"Updated","game":"pokemon","decklist":"cards"}`),
		"id",
		deckID.String(),
	), userID.String())
	response := recorder()
	controllers.NewDeckController(service).UpdateDeck(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code)
	service.AssertExpectations(t)
}

func TestDeckController_DeleteDeckReportsInternalError(t *testing.T) {
	service := &mocks.DeckServiceMock{}
	userID, deckID := uuid.New(), uuid.New()
	service.On("DeleteDeck", mock.Anything, deckID, userID).Return(errors.New("database unavailable")).Once()
	request := withUserID(withRouteParam(
		newRequest(http.MethodDelete, "/decks/"+deckID.String(), ""),
		"id",
		deckID.String(),
	), userID.String())
	response := recorder()
	controllers.NewDeckController(service).DeleteDeck(response, request)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	service.AssertExpectations(t)
}
