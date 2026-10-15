package controllers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/controllers"
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTournamentController_CreateTournament(t *testing.T) {
	t.Run("requires user identity", func(t *testing.T) {
		service := &mocks.TournamentServiceMock{}
		response := recorder()
		controllers.NewTournamentController(service).CreateTournament(response,
			newRequest(http.MethodPost, "/tournaments", "{}"))
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		service.AssertNotCalled(t, "CreateTournament", mock.Anything, mock.Anything)
	})
	t.Run("rejects malformed JSON", func(t *testing.T) {
		service := &mocks.TournamentServiceMock{}
		request := withUserID(newRequest(http.MethodPost, "/tournaments", "{"), uuid.New().String())
		response := recorder()
		controllers.NewTournamentController(service).CreateTournament(response, request)
		assert.Equal(t, http.StatusBadRequest, response.Code)
		service.AssertNotCalled(t, "CreateTournament", mock.Anything, mock.Anything)
	})
	t.Run("creates tournament with request fields", func(t *testing.T) {
		service := &mocks.TournamentServiceMock{}
		userID, deckID := uuid.New(), uuid.New()
		date := time.Date(2026, time.October, 15, 0, 0, 0, 0, time.UTC)
		service.On("CreateTournament", mock.Anything, mock.MatchedBy(func(tournament *models.Tournament) bool {
			return tournament.Name == "Regional" && tournament.Date.Equal(date) &&
				tournament.Game == models.GamePokemon && tournament.Scope == models.ScopeMajor &&
				tournament.IsOfficialCircuit && tournament.UserUUID == userID && tournament.DeckUUID != nil &&
				*tournament.DeckUUID == deckID
		})).Return(nil).Once()
		body := `{"name":"Regional","date":"2026-10-15T00:00:00Z","game":"pokemon","scope":"MAJOR","is_official_circuit":true,"deck_uuid":"` + deckID.String() + `"}`
		request := withUserID(newRequest(http.MethodPost, "/tournaments", body), userID.String())
		response := recorder()
		controllers.NewTournamentController(service).CreateTournament(response, request)
		require.Equal(t, http.StatusCreated, response.Code)
		var got models.Tournament
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		assert.Equal(t, userID, got.UserUUID)
		assert.Equal(t, &deckID, got.DeckUUID)
		service.AssertExpectations(t)
	})
	t.Run("reports service failure", func(t *testing.T) {
		service := &mocks.TournamentServiceMock{}
		service.On("CreateTournament", mock.Anything, mock.Anything).Return(errors.New("write failed")).Once()
		request := withUserID(newRequest(http.MethodPost, "/tournaments", `{"name":"Regional"}`), uuid.New().String())
		response := recorder()
		controllers.NewTournamentController(service).CreateTournament(response, request)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		service.AssertExpectations(t)
	})
}

func TestTournamentController_GetTournament(t *testing.T) {
	t.Run("rejects invalid UUID", func(t *testing.T) {
		service := &mocks.TournamentServiceMock{}
		response := recorder()
		controllers.NewTournamentController(service).GetTournament(response,
			withRouteParam(newRequest(http.MethodGet, "/tournaments/invalid", ""), "id", "invalid"))
		assert.Equal(t, http.StatusBadRequest, response.Code)
		service.AssertNotCalled(t, "GetTournamentByID", mock.Anything, mock.Anything)
	})
	t.Run("returns tournament", func(t *testing.T) {
		service := &mocks.TournamentServiceMock{}
		id := uuid.New()
		want := &models.Tournament{UUID: id, Name: "Regional"}
		service.On("GetTournamentByID", mock.Anything, id).Return(want, nil).Once()
		response := recorder()
		controllers.NewTournamentController(service).GetTournament(response,
			withRouteParam(newRequest(http.MethodGet, "/tournaments/"+id.String(), ""), "id", id.String()))
		require.Equal(t, http.StatusOK, response.Code)
		var got models.Tournament
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		assert.Equal(t, id, got.UUID)
		service.AssertExpectations(t)
	})
	t.Run("returns not found on service failure", func(t *testing.T) {
		service := &mocks.TournamentServiceMock{}
		id := uuid.New()
		service.On("GetTournamentByID", mock.Anything, id).
			Return((*models.Tournament)(nil), errors.New("missing")).Once()
		response := recorder()
		controllers.NewTournamentController(service).GetTournament(response,
			withRouteParam(newRequest(http.MethodGet, "/tournaments/"+id.String(), ""), "id", id.String()))
		assert.Equal(t, http.StatusNotFound, response.Code)
		service.AssertExpectations(t)
	})
}

func TestTournamentController_ListUserTournaments(t *testing.T) {
	t.Run("requires user identity", func(t *testing.T) {
		service := &mocks.TournamentServiceMock{}
		response := recorder()
		controllers.NewTournamentController(service).ListUserTournaments(response,
			newRequest(http.MethodGet, "/tournaments", ""))
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		service.AssertNotCalled(t, "ListUserTournaments", mock.Anything, mock.Anything)
	})
	t.Run("returns user tournaments", func(t *testing.T) {
		service := &mocks.TournamentServiceMock{}
		userID := uuid.New()
		want := []*models.Tournament{{Name: "Regional"}}
		service.On("ListUserTournaments", mock.Anything, userID).Return(want, nil).Once()
		request := withUserID(newRequest(http.MethodGet, "/tournaments", ""), userID.String())
		response := recorder()
		controllers.NewTournamentController(service).ListUserTournaments(response, request)
		require.Equal(t, http.StatusOK, response.Code)
		var got []*models.Tournament
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		require.Len(t, got, 1)
		assert.Equal(t, "Regional", got[0].Name)
		service.AssertExpectations(t)
	})
	t.Run("reports service failure", func(t *testing.T) {
		service := &mocks.TournamentServiceMock{}
		userID := uuid.New()
		service.On("ListUserTournaments", mock.Anything, userID).
			Return(([]*models.Tournament)(nil), errors.New("query failed")).Once()
		request := withUserID(newRequest(http.MethodGet, "/tournaments", ""), userID.String())
		response := recorder()
		controllers.NewTournamentController(service).ListUserTournaments(response, request)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		service.AssertExpectations(t)
	})
}
