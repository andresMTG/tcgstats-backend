package models

import (
	"time"

	"github.com/google/uuid"
)

type Match struct {
	UUID                  uuid.UUID  `gorm:"type:uuid;primaryKey" json:"uuid"`
	Result                ResultName `gorm:"type:result_name;not null" json:"result"`
	RoundNumber           int        `gorm:"type:int;not null" json:"round_number"`
	Phase                 MatchPhase `gorm:"type:match_phase;not null" json:"phase"`
	OpposingDeckArchetype string     `gorm:"type:varchar(255);not null" json:"opposing_deck_archetype"`
	TournamentUUID        uuid.UUID  `gorm:"type:uuid" json:"tournament_uuid"`
	OpponentUUID          uuid.UUID  `gorm:"type:uuid" json:"opponent_uuid"`
	DeletedAt             *time.Time `gorm:"type:timestamptz" json:"deleted_at"`

	Tournament *Tournament `gorm:"foreignKey:TournamentUUID" json:"tournament,omitempty"`
	Opponent   *Opponent   `gorm:"foreignKey:OpponentUUID" json:"opponent,omitempty"`
}
