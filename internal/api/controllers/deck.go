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

type DeckController struct {
	service interfaces.DeckService
}

func NewDeckController(service interfaces.DeckService) *DeckController {
	return &DeckController{service: service}
}

func (c *DeckController) CreateDeck(w http.ResponseWriter, r *http.Request) {
	userIdStr, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userUUID, _ := uuid.Parse(userIdStr)

	var req dto.CreateDeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	deck := &models.Deck{
		Name:     req.Name,
		Game:     req.Game,
		Decklist: req.Decklist,
		UserUUID: userUUID,
	}

	if err := c.service.CreateDeck(r.Context(), deck); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deck)
}

func (c *DeckController) GetDeck(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	deck, err := c.service.GetDeckByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(deck)
}

func (c *DeckController) ListUserDecks(w http.ResponseWriter, r *http.Request) {
	userIdStr, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userUUID, _ := uuid.Parse(userIdStr)

	var gameFilter *models.GameName
	if gameQuery := r.URL.Query().Get("game"); gameQuery != "" {
		game := models.GameName(gameQuery)
		gameFilter = &game
	}

	decks, err := c.service.ListUserDecks(r.Context(), userUUID, gameFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(decks)
}
