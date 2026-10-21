package models

import (
	"time"

	"github.com/google/uuid"
)

type Deck struct {
	UUID      uuid.UUID  `gorm:"type:uuid;primaryKey" json:"uuid"`
	Name      string     `gorm:"type:varchar(255);not null" json:"name"`
	Game      GameName   `gorm:"type:game_name;not null" json:"game"`
	Decklist  string     `gorm:"type:varchar(500);not null" json:"decklist"`
	UserUUID  uuid.UUID  `gorm:"type:uuid" json:"user_uuid"`
	DeletedAt *time.Time `gorm:"type:timestamptz" json:"deleted_at"`

	User *User `gorm:"foreignKey:UserUUID" json:"user,omitempty"`
}
