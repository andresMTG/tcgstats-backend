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

func TestOpponentController_CreateOpponent(t *testing.T) {
	t.Run("requires user identity", func(t *testing.T) {
		service := &mocks.OpponentServiceMock{}
		response := recorder()
		controllers.NewOpponentController(service).CreateOpponent(response,
			newRequest(http.MethodPost, "/opponents", "{}"))
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		service.AssertNotCalled(t, "CreateOpponent", mock.Anything, mock.Anything)
	})
	t.Run("rejects malformed JSON", func(t *testing.T) {
		service := &mocks.OpponentServiceMock{}
		request := withUserID(newRequest(http.MethodPost, "/opponents", "{"), uuid.New().String())
		response := recorder()
		controllers.NewOpponentController(service).CreateOpponent(response, request)
		assert.Equal(t, http.StatusBadRequest, response.Code)
		service.AssertNotCalled(t, "CreateOpponent", mock.Anything, mock.Anything)
	})
	t.Run("creates opponent for authenticated user", func(t *testing.T) {
		service := &mocks.OpponentServiceMock{}
		userID := uuid.New()
		service.On("CreateOpponent", mock.Anything, mock.MatchedBy(func(opponent *models.Opponent) bool {
			return opponent.Name == "Opponent" && opponent.UserUUID == userID
		})).Return(nil).Once()
		request := withUserID(newRequest(http.MethodPost, "/opponents", `{"name":"Opponent"}`), userID.String())
		response := recorder()
		controllers.NewOpponentController(service).CreateOpponent(response, request)
		require.Equal(t, http.StatusCreated, response.Code)
		var got models.Opponent
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		assert.Equal(t, userID, got.UserUUID)
		service.AssertExpectations(t)
	})
	t.Run("reports service failure", func(t *testing.T) {
		service := &mocks.OpponentServiceMock{}
		service.On("CreateOpponent", mock.Anything, mock.Anything).Return(errors.New("write failed")).Once()
		request := withUserID(newRequest(http.MethodPost, "/opponents", `{"name":"Opponent"}`), uuid.New().String())
		response := recorder()
		controllers.NewOpponentController(service).CreateOpponent(response, request)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		service.AssertExpectations(t)
	})
}

func TestOpponentController_GetOpponent(t *testing.T) {
	t.Run("rejects invalid UUID", func(t *testing.T) {
		service := &mocks.OpponentServiceMock{}
		response := recorder()
		controllers.NewOpponentController(service).GetOpponent(response,
			withRouteParam(newRequest(http.MethodGet, "/opponents/invalid", ""), "id", "invalid"))
		assert.Equal(t, http.StatusBadRequest, response.Code)
		service.AssertNotCalled(t, "GetOpponentByID", mock.Anything, mock.Anything)
	})
	t.Run("returns opponent", func(t *testing.T) {
		service := &mocks.OpponentServiceMock{}
		id := uuid.New()
		want := &models.Opponent{UUID: id, Name: "Opponent"}
		service.On("GetOpponentByID", mock.Anything, id).Return(want, nil).Once()
		response := recorder()
		controllers.NewOpponentController(service).GetOpponent(response,
			withRouteParam(newRequest(http.MethodGet, "/opponents/"+id.String(), ""), "id", id.String()))
		require.Equal(t, http.StatusOK, response.Code)
		var got models.Opponent
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		assert.Equal(t, id, got.UUID)
		service.AssertExpectations(t)
	})
	t.Run("returns not found on service failure", func(t *testing.T) {
		service := &mocks.OpponentServiceMock{}
		id := uuid.New()
		service.On("GetOpponentByID", mock.Anything, id).
			Return((*models.Opponent)(nil), errors.New("missing")).Once()
		response := recorder()
		controllers.NewOpponentController(service).GetOpponent(response,
			withRouteParam(newRequest(http.MethodGet, "/opponents/"+id.String(), ""), "id", id.String()))
		assert.Equal(t, http.StatusNotFound, response.Code)
		service.AssertExpectations(t)
	})
}

func TestOpponentController_ListUserOpponents(t *testing.T) {
	t.Run("requires user identity", func(t *testing.T) {
		service := &mocks.OpponentServiceMock{}
		response := recorder()
		controllers.NewOpponentController(service).ListUserOpponents(response, newRequest(http.MethodGet, "/opponents", ""))
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		service.AssertNotCalled(t, "ListUserOpponents", mock.Anything, mock.Anything)
	})
	t.Run("returns user opponents", func(t *testing.T) {
		service := &mocks.OpponentServiceMock{}
		userID := uuid.New()
		want := []*models.Opponent{{Name: "Opponent"}}
		service.On("ListUserOpponents", mock.Anything, userID).Return(want, nil).Once()
		request := withUserID(newRequest(http.MethodGet, "/opponents", ""), userID.String())
		response := recorder()
		controllers.NewOpponentController(service).ListUserOpponents(response, request)
		require.Equal(t, http.StatusOK, response.Code)
		var got []*models.Opponent
		require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
		require.Len(t, got, 1)
		assert.Equal(t, "Opponent", got[0].Name)
		service.AssertExpectations(t)
	})
	t.Run("reports service failure", func(t *testing.T) {
		service := &mocks.OpponentServiceMock{}
		userID := uuid.New()
		service.On("ListUserOpponents", mock.Anything, userID).
			Return(([]*models.Opponent)(nil), errors.New("query failed")).Once()
		request := withUserID(newRequest(http.MethodGet, "/opponents", ""), userID.String())
		response := recorder()
		controllers.NewOpponentController(service).ListUserOpponents(response, request)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		service.AssertExpectations(t)
	})
}
