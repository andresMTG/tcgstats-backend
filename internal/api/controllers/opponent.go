package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/dto"
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/andresMTG/tcgstats-backend/internal/api/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OpponentController struct {
	service interfaces.OpponentService
}

func NewOpponentController(service interfaces.OpponentService) *OpponentController {
	return &OpponentController{service: service}
}

func (c *OpponentController) CreateOpponent(w http.ResponseWriter, r *http.Request) {
	userIdStr, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userUUID, _ := uuid.Parse(userIdStr)

	var req dto.CreateOpponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	opponent := &models.Opponent{
		Name:     req.Name,
		UserUUID: userUUID,
	}

	if err := c.service.CreateOpponent(r.Context(), opponent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(opponent)
}

func (c *OpponentController) UpdateOpponent(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := authenticatedUserUUID(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	var req dto.UpdateOpponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	opponent := &models.Opponent{UUID: id, Name: req.Name, UserUUID: userUUID}
	if err := c.service.UpdateOpponent(r.Context(), opponent); err != nil {
		writeMutationError(w, err)
		return
	}
	json.NewEncoder(w).Encode(opponent)
}

func (c *OpponentController) DeleteOpponent(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := authenticatedUserUUID(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}
	if err := c.service.DeleteOpponent(r.Context(), id, userUUID); err != nil {
		writeMutationError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *OpponentController) GetOpponent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	opponent, err := c.service.GetOpponentByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(opponent)
}

func (c *OpponentController) ListUserOpponents(w http.ResponseWriter, r *http.Request) {
	userIdStr, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userUUID, _ := uuid.Parse(userIdStr)

	opponents, err := c.service.ListUserOpponents(r.Context(), userUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(opponents)
}
