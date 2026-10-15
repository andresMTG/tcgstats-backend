package mocks

import (
	"context"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// UserRepositoryMock provides a testify mock for interfaces.UserRepository.
type UserRepositoryMock struct {
	mock.Mock
}

func (r *UserRepositoryMock) Create(ctx context.Context, user *models.User) error {
	return r.Called(ctx, user).Error(0)
}

func (r *UserRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := r.Called(ctx, id)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (r *UserRepositoryMock) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := r.Called(ctx, email)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

var _ interfaces.UserRepository = (*UserRepositoryMock)(nil)
