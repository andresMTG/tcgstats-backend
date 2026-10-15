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

type TournamentController struct {
	service interfaces.TournamentService
}

func NewTournamentController(service interfaces.TournamentService) *TournamentController {
	return &TournamentController{service: service}
}

func (c *TournamentController) CreateTournament(w http.ResponseWriter, r *http.Request) {
	userIdStr, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userUUID, _ := uuid.Parse(userIdStr)

	var req dto.CreateTournamentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tournament := &models.Tournament{
		Name:              req.Name,
		Date:              req.Date,
		Game:              req.Game,
		Scope:             req.Scope,
		IsOfficialCircuit: req.IsOfficialCircuit,
		UserUUID:          userUUID,
		DeckUUID:          req.DeckUUID,
	}

	if err := c.service.CreateTournament(r.Context(), tournament); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tournament)
}

func (c *TournamentController) GetTournament(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	tournament, err := c.service.GetTournamentByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(tournament)
}

func (c *TournamentController) ListUserTournaments(w http.ResponseWriter, r *http.Request) {
	userIdStr, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userUUID, _ := uuid.Parse(userIdStr)

	tournaments, err := c.service.ListUserTournaments(r.Context(), userUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tournaments)
}
