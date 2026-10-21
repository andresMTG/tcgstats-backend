package models

import (
	"time"

	"github.com/google/uuid"
)

type Opponent struct {
	UUID      uuid.UUID  `gorm:"type:uuid;primaryKey" json:"uuid"`
	Name      string     `gorm:"type:varchar(255);not null" json:"name"`
	UserUUID  uuid.UUID  `gorm:"type:uuid" json:"user_uuid"`
	DeletedAt *time.Time `gorm:"type:timestamptz" json:"deleted_at"`

	User *User `gorm:"foreignKey:UserUUID" json:"user,omitempty"`
}
