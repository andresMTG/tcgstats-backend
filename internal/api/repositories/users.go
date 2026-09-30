package repositories

import (
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"gorm.io/gorm"
)

type User struct {
	db *gorm.DB
}


func NewUser(db *gorm.DB) *User {
	return &User{db: db}
}

func (u *User) Create(user *models.Users) error {
	if err := user.BeforeCreate(u.db); err != nil {
		return err
	}
	return u.db.Create(user).Error
}

func (u *User) FindByEmail(email string) (*models.Users, error) {
	var user models.Users
	if err := u.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *User) FindByUuid(uuid string) (*models.Users, error) {
	var user models.Users
	if err := u.db.Where("uuid = ?", uuid).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}