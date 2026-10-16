package models

import (
	"github.com/google/uuid"
)

type MatchRecord struct {
	Matches int64   `json:"matches" gorm:"column:matches"`
	Wins    int64   `json:"wins" gorm:"column:wins"`
	Draws   int64   `json:"draws" gorm:"column:draws"`
	Losses  int64   `json:"losses" gorm:"column:losses"`
	WinRate float64 `json:"win_rate" gorm:"-"`
}

type DeckPerformance struct {
	DeckUUID uuid.UUID `json:"deck_uuid" gorm:"column:deck_uuid"`
	DeckName string    `json:"deck_name" gorm:"column:deck_name"`
	MatchRecord
}

type DeckMatchupPerformance struct {
	DeckUUID              uuid.UUID `json:"deck_uuid" gorm:"column:deck_uuid"`
	DeckName              string    `json:"deck_name" gorm:"column:deck_name"`
	OpponentDeckArchetype string    `json:"opponent_deck_archetype" gorm:"column:opponent_deck_archetype"`
	MatchRecord
}

type OpponentPerformance struct {
	OpponentUUID uuid.UUID `json:"opponent_uuid" gorm:"column:opponent_uuid"`
	OpponentName string    `json:"opponent_name" gorm:"column:opponent_name"`
	MatchRecord
}

type PhasePerformance struct {
	Phase MatchPhase `json:"phase" gorm:"column:phase"`
	MatchRecord
}

type TournamentScopePerformance struct {
	Scope TournamentScope `json:"scope" gorm:"column:scope"`
	MatchRecord
}
