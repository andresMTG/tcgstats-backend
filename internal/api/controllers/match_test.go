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

func TestMatchController_CreateMatch(t *testing.T) {
	t.Run("rejects malformed JSON", func(t *testing.T) {
		service := &mocks.MatchServiceMock{}
		response := recorder()
		controllers.NewMatchController(service).CreateMatch(response, newRequest(http.MethodPost, "/matches", "{"))
		assert.Equal(t, http.StatusBadRequest, response.Code)
		service.AssertNotCalled(t, "CreateMatch", mock.Anything, mock.Anything)
	})
	t.Run("creates match", func(t *testing.T) {
		service := &mocks.MatchServiceMock{}
		tournamentID, opponentID := uuid.New(), uuid.New()
		service.On("CreateMatch", mock.Anything, mock.MatchedBy(func(match *models.Match) bool {
			return match.Result == models.ResultWin && match.RoundNumber == 4 &&
				match.Phase == models.PhaseTop8 && match.OpposingDeckArchetype == "Control" &&
				match.TournamentUUID == tournamentID && match.OpponentUUID == opponentID
		})).Return(nil).Once()
		body := `{"result":"win","round_number":4,"phase":"TOP_8","opposing_deck_archetype":"Control","tournament_uuid":"` +
			tournamentID.String() + `","opponent_uuid":"` + opponentID.String() + `"}`
		response := recorder()
		controllers.NewMatchController(service).CreateMatch(response, newRequest(http.MethodPost, "/matches", body))
		require.Equal(t, http.StatusCreated, response.Code)
		var got models.Match
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		assert.Equal(t, tournamentID, got.TournamentUUID)
		assert.Equal(t, opponentID, got.OpponentUUID)
		service.AssertExpectations(t)
	})
	t.Run("reports service failure", func(t *testing.T) {
		service := &mocks.MatchServiceMock{}
		service.On("CreateMatch", mock.Anything, mock.Anything).Return(errors.New("write failed")).Once()
		response := recorder()
		controllers.NewMatchController(service).CreateMatch(response, newRequest(http.MethodPost, "/matches", `{}`))
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		service.AssertExpectations(t)
	})
}

func TestMatchController_GetMatch(t *testing.T) {
	t.Run("rejects invalid UUID", func(t *testing.T) {
		service := &mocks.MatchServiceMock{}
		response := recorder()
		controllers.NewMatchController(service).GetMatch(response,
			withRouteParam(newRequest(http.MethodGet, "/matches/invalid", ""), "id", "invalid"))
		assert.Equal(t, http.StatusBadRequest, response.Code)
		service.AssertNotCalled(t, "GetMatchByID", mock.Anything, mock.Anything)
	})
	t.Run("returns match", func(t *testing.T) {
		service := &mocks.MatchServiceMock{}
		id := uuid.New()
		want := &models.Match{UUID: id, Result: models.ResultWin}
		service.On("GetMatchByID", mock.Anything, id).Return(want, nil).Once()
		response := recorder()
		controllers.NewMatchController(service).GetMatch(response,
			withRouteParam(newRequest(http.MethodGet, "/matches/"+id.String(), ""), "id", id.String()))
		require.Equal(t, http.StatusOK, response.Code)
		var got models.Match
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		assert.Equal(t, id, got.UUID)
		service.AssertExpectations(t)
	})
	t.Run("returns not found on service failure", func(t *testing.T) {
		service := &mocks.MatchServiceMock{}
		id := uuid.New()
		service.On("GetMatchByID", mock.Anything, id).Return((*models.Match)(nil), errors.New("missing")).Once()
		response := recorder()
		controllers.NewMatchController(service).GetMatch(response,
			withRouteParam(newRequest(http.MethodGet, "/matches/"+id.String(), ""), "id", id.String()))
		assert.Equal(t, http.StatusNotFound, response.Code)
		service.AssertExpectations(t)
	})
}

func TestMatchController_ListTournamentMatches(t *testing.T) {
	t.Run("rejects invalid tournament UUID", func(t *testing.T) {
		service := &mocks.MatchServiceMock{}
		response := recorder()
		controllers.NewMatchController(service).ListTournamentMatches(response,
			withRouteParam(newRequest(http.MethodGet, "/tournaments/invalid/matches", ""), "tournamentId", "invalid"))
		assert.Equal(t, http.StatusBadRequest, response.Code)
		service.AssertNotCalled(t, "ListTournamentMatches", mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("forwards result filter", func(t *testing.T) {
		service := &mocks.MatchServiceMock{}
		tournamentID, result := uuid.New(), models.ResultDraw
		want := []*models.Match{{Result: result}}
		service.On("ListTournamentMatches", mock.Anything, tournamentID, &result).Return(want, nil).Once()
		request := withRouteParam(newRequest(http.MethodGet,
			"/tournaments/"+tournamentID.String()+"/matches?result=draw", ""), "tournamentId", tournamentID.String())
		response := recorder()
		controllers.NewMatchController(service).ListTournamentMatches(response, request)
		require.Equal(t, http.StatusOK, response.Code)
		var got []*models.Match
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		require.Len(t, got, 1)
		assert.Equal(t, result, got[0].Result)
		service.AssertExpectations(t)
	})
	t.Run("reports service failure", func(t *testing.T) {
		service := &mocks.MatchServiceMock{}
		tournamentID := uuid.New()
		service.On("ListTournamentMatches", mock.Anything, tournamentID, (*models.ResultName)(nil)).
			Return(([]*models.Match)(nil), errors.New("query failed")).Once()
		request := withRouteParam(newRequest(http.MethodGet,
			"/tournaments/"+tournamentID.String()+"/matches", ""), "tournamentId", tournamentID.String())
		response := recorder()
		controllers.NewMatchController(service).ListTournamentMatches(response, request)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		service.AssertExpectations(t)
	})
}

func TestMatchController_ListDeckMatches(t *testing.T) {
	t.Run("rejects invalid deck UUID", func(t *testing.T) {
		service := &mocks.MatchServiceMock{}
		response := recorder()
		controllers.NewMatchController(service).ListDeckMatches(response,
			withRouteParam(newRequest(http.MethodGet, "/decks/invalid/matches", ""), "deckId", "invalid"))
		assert.Equal(t, http.StatusBadRequest, response.Code)
		service.AssertNotCalled(t, "ListDeckMatches", mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("forwards result filter", func(t *testing.T) {
		service := &mocks.MatchServiceMock{}
		deckID, result := uuid.New(), models.ResultWin
		want := []*models.Match{{Result: result}}
		service.On("ListDeckMatches", mock.Anything, deckID, &result).Return(want, nil).Once()
		request := withRouteParam(newRequest(http.MethodGet, "/decks/"+deckID.String()+"/matches?result=win", ""),
			"deckId", deckID.String())
		response := recorder()
		controllers.NewMatchController(service).ListDeckMatches(response, request)
		require.Equal(t, http.StatusOK, response.Code)
		var got []*models.Match
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		require.Len(t, got, 1)
		assert.Equal(t, result, got[0].Result)
		service.AssertExpectations(t)
	})
	t.Run("reports service failure", func(t *testing.T) {
		service := &mocks.MatchServiceMock{}
		deckID := uuid.New()
		service.On("ListDeckMatches", mock.Anything, deckID, (*models.ResultName)(nil)).
			Return(([]*models.Match)(nil), errors.New("query failed")).Once()
		request := withRouteParam(newRequest(http.MethodGet, "/decks/"+deckID.String()+"/matches", ""),
			"deckId", deckID.String())
		response := recorder()
		controllers.NewMatchController(service).ListDeckMatches(response, request)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		service.AssertExpectations(t)
	})
}
