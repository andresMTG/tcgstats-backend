package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Users struct {
	UUID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name     string
	Email    string `gorm:"uniqueIndex"`
	Password string
}

func (u *Users) BeforeCreate(tx *gorm.DB) (err error) {
	if u.UUID == uuid.Nil {
		uuidV7, err := uuid.NewV7()
		if err != nil {
			return err
		}
		u.UUID = uuidV7
	}
	return nil
}
