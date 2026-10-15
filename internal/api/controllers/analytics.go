package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/andresMTG/tcgstats-backend/internal/api/middleware"
	"github.com/google/uuid"
)

type AnalyticsController struct {
	service interfaces.AnalyticsService
}

func NewAnalyticsController(service interfaces.AnalyticsService) *AnalyticsController {
	return &AnalyticsController{service: service}
}

func (c *AnalyticsController) DeckPerformance(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := authenticatedUserUUID(w, r)
	if !ok {
		return
	}
	from, to, ok := parseAnalyticsDateRange(w, r)
	if !ok {
		return
	}
	performance, err := c.service.DeckPerformance(r.Context(), userUUID, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeAnalyticsJSON(w, performance)
}

func (c *AnalyticsController) OpponentPerformance(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := authenticatedUserUUID(w, r)
	if !ok {
		return
	}
	from, to, ok := parseAnalyticsDateRange(w, r)
	if !ok {
		return
	}
	performance, err := c.service.OpponentPerformance(r.Context(), userUUID, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeAnalyticsJSON(w, performance)
}

func (c *AnalyticsController) PhasePerformance(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := authenticatedUserUUID(w, r)
	if !ok {
		return
	}
	from, to, ok := parseAnalyticsDateRange(w, r)
	if !ok {
		return
	}
	performance, err := c.service.PhasePerformance(r.Context(), userUUID, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeAnalyticsJSON(w, performance)
}

func (c *AnalyticsController) TournamentScopePerformance(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := authenticatedUserUUID(w, r)
	if !ok {
		return
	}
	from, to, ok := parseAnalyticsDateRange(w, r)
	if !ok {
		return
	}
	performance, err := c.service.TournamentScopePerformance(r.Context(), userUUID, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeAnalyticsJSON(w, performance)
}

func authenticatedUserUUID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return uuid.Nil, false
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "invalid user id in token", http.StatusUnauthorized)
		return uuid.Nil, false
	}
	return userUUID, true
}

func parseAnalyticsDateRange(w http.ResponseWriter, r *http.Request) (*time.Time, *time.Time, bool) {
	from, err := parseAnalyticsDate(r.URL.Query().Get("from"))
	if err != nil {
		http.Error(w, "invalid from date: expected YYYY-MM-DD", http.StatusBadRequest)
		return nil, nil, false
	}
	to, err := parseAnalyticsDate(r.URL.Query().Get("to"))
	if err != nil {
		http.Error(w, "invalid to date: expected YYYY-MM-DD", http.StatusBadRequest)
		return nil, nil, false
	}
	if from != nil && to != nil && from.After(*to) {
		http.Error(w, "from date must be on or before to date", http.StatusBadRequest)
		return nil, nil, false
	}
	return from, to, true
}

func parseAnalyticsDate(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &date, nil
}

func writeAnalyticsJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
