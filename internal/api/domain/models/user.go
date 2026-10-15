package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UUID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"uuid"`
	Name     string    `gorm:"type:varchar(255);not null" json:"name"`
	Email    string    `gorm:"type:varchar(255);unique;not null" json:"email"`
	Password string    `gorm:"type:varchar(255);not null" json:"-"`
}

type AuthTokens struct {
	AccessToken     string        `json:"-"`
	RefreshToken    string        `json:"-"`
	AccessTokenTTL  time.Duration `json:"-"`
	RefreshTokenTTL time.Duration `json:"-"`
}
