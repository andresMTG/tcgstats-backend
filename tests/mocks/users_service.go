package mocks

import (
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/dto"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/stretchr/testify/mock"
)

// UsersServiceMock provides a testify mock for interfaces.UsersService.
type UsersServiceMock struct {
	mock.Mock
}

func (s *UsersServiceMock) Create(req *dto.UserRequest) error {
	return s.Called(req).Error(0)
}

func (s *UsersServiceMock) Login(req *dto.UserLoginRequest) (*dto.UserLoginResponse, error) {
	args := s.Called(req)
	response, _ := args.Get(0).(*dto.UserLoginResponse)
	return response, args.Error(1)
}

func (s *UsersServiceMock) ShowMe(userID string) (*dto.UserSimpleResponse, error) {
	args := s.Called(userID)
	response, _ := args.Get(0).(*dto.UserSimpleResponse)
	return response, args.Error(1)
}

var _ interfaces.UsersService = (*UsersServiceMock)(nil)
