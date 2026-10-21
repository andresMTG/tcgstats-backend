package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/dto"
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type MatchController struct {
	service interfaces.MatchService
}

func NewMatchController(service interfaces.MatchService) *MatchController {
	return &MatchController{service: service}
}

func (c *MatchController) RegisterRoutes(r chi.Router) {
	r.Post("/matches", c.CreateMatch)
	r.Get("/matches/{id}", c.GetMatch)
	r.Put("/matches/{id}", c.UpdateMatch)
	r.Delete("/matches/{id}", c.DeleteMatch)
	r.Get("/tournaments/{tournamentId}/matches", c.ListTournamentMatches)
	r.Get("/decks/{deckId}/matches", c.ListDeckMatches)
}

func (c *MatchController) CreateMatch(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	match := &models.Match{
		Result:                req.Result,
		RoundNumber:           req.RoundNumber,
		Phase:                 req.Phase,
		OpposingDeckArchetype: req.OpposingDeckArchetype,
		TournamentUUID:        req.TournamentUUID,
		OpponentUUID:          req.OpponentUUID,
	}

	if err := c.service.CreateMatch(r.Context(), match); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(match)
}

func (c *MatchController) UpdateMatch(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := authenticatedUserUUID(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	var req dto.UpdateMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	match := &models.Match{
		UUID:                  id,
		Result:                req.Result,
		RoundNumber:           req.RoundNumber,
		Phase:                 req.Phase,
		OpposingDeckArchetype: req.OpposingDeckArchetype,
		TournamentUUID:        req.TournamentUUID,
		OpponentUUID:          req.OpponentUUID,
	}
	if err := c.service.UpdateMatch(r.Context(), match, userUUID); err != nil {
		writeMutationError(w, err)
		return
	}
	json.NewEncoder(w).Encode(match)
}

func (c *MatchController) DeleteMatch(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := authenticatedUserUUID(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}
	if err := c.service.DeleteMatch(r.Context(), id, userUUID); err != nil {
		writeMutationError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *MatchController) GetMatch(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	match, err := c.service.GetMatchByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(match)
}

func (c *MatchController) ListTournamentMatches(w http.ResponseWriter, r *http.Request) {
	tournamentIdStr := chi.URLParam(r, "tournamentId")
	tournamentId, err := uuid.Parse(tournamentIdStr)
	if err != nil {
		http.Error(w, "invalid tournament uuid", http.StatusBadRequest)
		return
	}

	var resultFilter *models.ResultName
	if resultQuery := r.URL.Query().Get("result"); resultQuery != "" {
		result := models.ResultName(resultQuery)
		resultFilter = &result
	}

	matches, err := c.service.ListTournamentMatches(r.Context(), tournamentId, resultFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(matches)
}

func (c *MatchController) ListDeckMatches(w http.ResponseWriter, r *http.Request) {
	deckIdStr := chi.URLParam(r, "deckId")
	deckId, err := uuid.Parse(deckIdStr)
	if err != nil {
		http.Error(w, "invalid deck uuid", http.StatusBadRequest)
		return
	}

	var resultFilter *models.ResultName
	if resultQuery := r.URL.Query().Get("result"); resultQuery != "" {
		result := models.ResultName(resultQuery)
		resultFilter = &result
	}

	matches, err := c.service.ListDeckMatches(r.Context(), deckId, resultFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(matches)
}
