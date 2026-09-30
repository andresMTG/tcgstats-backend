package interfaces

import (
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/dto"
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
)

// UsersService defines the operations the users controller needs.
type UsersService interface {
	Create(*dto.UserRequest) error
	Login(*dto.UserLoginRequest) (*dto.UserLoginResponse, error)
	ShowMe(string) (*dto.UserSimpleResponse, error)
}

// UserRepository defines the persistence operations required by UsersService.
type UserRepository interface {
	Create(*models.Users) error
	FindByEmail(string) (*models.Users, error)
	FindByUuid(string) (*models.Users, error)
}
