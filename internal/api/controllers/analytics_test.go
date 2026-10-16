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

func TestAnalyticsController_DeckPerformance(t *testing.T) {
	t.Run("returns analytics for authenticated user and date range", func(t *testing.T) {
		service := &mocks.AnalyticsServiceMock{}
		userID := uuid.New()
		from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
		game := models.GamePokemon
		service.On("DeckPerformance", mock.Anything, userID, &from, &to, &game).Return([]models.DeckPerformance{{
			DeckUUID: uuid.New(),
			DeckName: "Tempo",
			MatchRecord: models.MatchRecord{
				Matches: 4, Wins: 2, Draws: 1, Losses: 1, WinRate: 50,
			},
		}}, nil).Once()
		request := withUserID(newRequest(http.MethodGet,
			"/analytics/decks?from=2026-10-01&to=2026-10-31&game=pokemon", ""), userID.String())
		response := recorder()

		controllers.NewAnalyticsController(service).DeckPerformance(response, request)

		require.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, "application/json", response.Header().Get("Content-Type"))
		var got []models.DeckPerformance
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		require.Len(t, got, 1)
		assert.Equal(t, "Tempo", got[0].DeckName)
		assert.Equal(t, int64(1), got[0].Draws)
		assert.Equal(t, 50.0, got[0].WinRate)
		service.AssertExpectations(t)
	})
	t.Run("rejects invalid date range", func(t *testing.T) {
		service := &mocks.AnalyticsServiceMock{}
		userID := uuid.New()
		request := withUserID(newRequest(http.MethodGet,
			"/analytics/decks?from=2026-11-01&to=2026-10-01", ""), userID.String())
		response := recorder()
		controllers.NewAnalyticsController(service).DeckPerformance(response, request)
		assert.Equal(t, http.StatusBadRequest, response.Code)
		service.AssertNotCalled(t, "DeckPerformance", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("rejects invalid date format", func(t *testing.T) {
		service := &mocks.AnalyticsServiceMock{}
		request := withUserID(newRequest(http.MethodGet, "/analytics/decks?from=10-01-2026", ""), uuid.New().String())
		response := recorder()
		controllers.NewAnalyticsController(service).DeckPerformance(response, request)
		assert.Equal(t, http.StatusBadRequest, response.Code)
		service.AssertNotCalled(t, "DeckPerformance", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("rejects unsupported game", func(t *testing.T) {
		service := &mocks.AnalyticsServiceMock{}
		request := withUserID(newRequest(http.MethodGet, "/analytics/decks?game=unknown", ""), uuid.New().String())
		response := recorder()
		controllers.NewAnalyticsController(service).DeckPerformance(response, request)
		assert.Equal(t, http.StatusBadRequest, response.Code)
		service.AssertNotCalled(t, "DeckPerformance", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("requires authenticated identity", func(t *testing.T) {
		service := &mocks.AnalyticsServiceMock{}
		response := recorder()
		controllers.NewAnalyticsController(service).DeckPerformance(response,
			newRequest(http.MethodGet, "/analytics/decks", ""))
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		service.AssertNotCalled(t, "DeckPerformance", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("reports service failure", func(t *testing.T) {
		service := &mocks.AnalyticsServiceMock{}
		userID := uuid.New()
		service.On("DeckPerformance", mock.Anything, userID, (*time.Time)(nil), (*time.Time)(nil), (*models.GameName)(nil)).
			Return([]models.DeckPerformance(nil), errors.New("database unavailable")).Once()
		request := withUserID(newRequest(http.MethodGet, "/analytics/decks", ""), userID.String())
		response := recorder()
		controllers.NewAnalyticsController(service).DeckPerformance(response, request)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		service.AssertExpectations(t)
	})
}

func TestAnalyticsController_OpponentPerformance(t *testing.T) {
	service := &mocks.AnalyticsServiceMock{}
	userID := uuid.New()
	service.On("OpponentPerformance", mock.Anything, userID, (*time.Time)(nil), (*time.Time)(nil), (*models.GameName)(nil)).
		Return([]models.OpponentPerformance{{OpponentName: "Rival"}}, nil).Once()
	request := withUserID(newRequest(http.MethodGet, "/analytics/opponents", ""), userID.String())
	response := recorder()
	controllers.NewAnalyticsController(service).OpponentPerformance(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var got []models.OpponentPerformance
	require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
	require.Len(t, got, 1)
	assert.Equal(t, "Rival", got[0].OpponentName)
	service.AssertExpectations(t)
}

func TestAnalyticsController_DeckMatchupPerformance(t *testing.T) {
	service := &mocks.AnalyticsServiceMock{}
	userID := uuid.New()
	game := models.GamePokemon
	service.On("DeckMatchupPerformance", mock.Anything, userID, (*time.Time)(nil), (*time.Time)(nil), &game).
		Return([]models.DeckMatchupPerformance{{
			DeckUUID: uuid.New(), DeckName: "Tempo", OpponentDeckArchetype: "Control",
			MatchRecord: models.MatchRecord{Matches: 2, Wins: 1, Losses: 1, WinRate: 50},
		}}, nil).Once()
	request := withUserID(newRequest(http.MethodGet, "/analytics/deck-matchups?game=pokemon", ""), userID.String())
	response := recorder()

	controllers.NewAnalyticsController(service).DeckMatchupPerformance(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	var got []models.DeckMatchupPerformance
	require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
	require.Len(t, got, 1)
	assert.Equal(t, "Tempo", got[0].DeckName)
	assert.Equal(t, "Control", got[0].OpponentDeckArchetype)
	assert.Equal(t, 50.0, got[0].WinRate)
	service.AssertExpectations(t)
}

func TestAnalyticsController_PhasePerformance(t *testing.T) {
	service := &mocks.AnalyticsServiceMock{}
	userID := uuid.New()
	service.On("PhasePerformance", mock.Anything, userID, (*time.Time)(nil), (*time.Time)(nil), (*models.GameName)(nil)).
		Return([]models.PhasePerformance{{Phase: models.PhaseFinals}}, nil).Once()
	request := withUserID(newRequest(http.MethodGet, "/analytics/phases", ""), userID.String())
	response := recorder()
	controllers.NewAnalyticsController(service).PhasePerformance(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var got []models.PhasePerformance
	require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
	require.Len(t, got, 1)
	assert.Equal(t, models.PhaseFinals, got[0].Phase)
	service.AssertExpectations(t)
}

func TestAnalyticsController_TournamentScopePerformance(t *testing.T) {
	service := &mocks.AnalyticsServiceMock{}
	userID := uuid.New()
	service.On("TournamentScopePerformance", mock.Anything, userID, (*time.Time)(nil), (*time.Time)(nil), (*models.GameName)(nil)).
		Return([]models.TournamentScopePerformance{{Scope: models.ScopePremier}}, nil).Once()
	request := withUserID(newRequest(http.MethodGet, "/analytics/tournament-scopes", ""), userID.String())
	response := recorder()
	controllers.NewAnalyticsController(service).TournamentScopePerformance(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var got []models.TournamentScopePerformance
	require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
	require.Len(t, got, 1)
	assert.Equal(t, models.ScopePremier, got[0].Scope)
	service.AssertExpectations(t)
}
