package models

import (
	"time"

	"github.com/google/uuid"
)

type Tournament struct {
	UUID              uuid.UUID       `gorm:"type:uuid;primaryKey" json:"uuid"`
	Name              string          `gorm:"type:varchar(255);not null" json:"name"`
	Date              time.Time       `gorm:"type:date;not null" json:"date"`
	Game              GameName        `gorm:"type:game_name;not null" json:"game"`
	Scope             TournamentScope `gorm:"type:tournament_scope;not null" json:"scope"`
	IsOfficialCircuit bool            `gorm:"type:boolean;default:false" json:"is_official_circuit"`
	UserUUID          uuid.UUID       `gorm:"type:uuid" json:"user_uuid"`
	DeckUUID          *uuid.UUID      `gorm:"type:uuid" json:"deck_uuid"`
	DeletedAt         *time.Time      `gorm:"type:timestamptz" json:"deleted_at"`

	User *User `gorm:"foreignKey:UserUUID" json:"user,omitempty"`
	Deck *Deck `gorm:"foreignKey:DeckUUID" json:"deck,omitempty"`
}
