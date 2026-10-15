package dto

import (
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/google/uuid"
)

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateDeckRequest struct {
	Name     string          `json:"name"`
	Game     models.GameName `json:"game"`
	Decklist string          `json:"decklist"`
}

type CreateTournamentRequest struct {
	Name              string                 `json:"name"`
	Date              time.Time              `json:"date"`
	Game              models.GameName        `json:"game"`
	Scope             models.TournamentScope `json:"scope"`
	IsOfficialCircuit bool                   `json:"is_official_circuit"`
	DeckUUID          *uuid.UUID             `json:"deck_uuid,omitempty"`
}

type CreateOpponentRequest struct {
	Name string `json:"name"`
}

type CreateMatchRequest struct {
	Result                models.ResultName `json:"result"`
	RoundNumber           int               `json:"round_number"`
	Phase                 models.MatchPhase `json:"phase"`
	OpposingDeckArchetype string            `json:"opposing_deck_archetype"`
	TournamentUUID        uuid.UUID         `json:"tournament_uuid"`
	OpponentUUID          uuid.UUID         `json:"opponent_uuid"`
}
