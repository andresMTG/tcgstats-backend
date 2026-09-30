package mocks

import (
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/stretchr/testify/mock"
)

// UserRepositoryMock provides a testify mock for interfaces.UserRepository.
type UserRepositoryMock struct {
	mock.Mock
}

func (r *UserRepositoryMock) Create(user *models.Users) error {
	return r.Called(user).Error(0)
}

func (r *UserRepositoryMock) FindByEmail(email string) (*models.Users, error) {
	args := r.Called(email)
	user, _ := args.Get(0).(*models.Users)
	return user, args.Error(1)
}

func (r *UserRepositoryMock) FindByUuid(uuid string) (*models.Users, error) {
	args := r.Called(uuid)
	user, _ := args.Get(0).(*models.Users)
	return user, args.Error(1)
}

var _ interfaces.UserRepository = (*UserRepositoryMock)(nil)
