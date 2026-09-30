package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/dto"
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

const minimumJWTSecretLength = 32

type UsersService struct {
	repo      interfaces.UserRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewUsersService(repo interfaces.UserRepository, jwtSecret string, tokenTTL time.Duration) (*UsersService, error) {
	if len(jwtSecret) < minimumJWTSecretLength {
		return nil, fmt.Errorf("JWT_SECRET must be at least %d bytes", minimumJWTSecretLength)
	}
	if tokenTTL <= 0 {
		return nil, errors.New("JWT expiration must be greater than zero")
	}

	return &UsersService{repo: repo, jwtSecret: []byte(jwtSecret), tokenTTL: tokenTTL}, nil
}

func (s UsersService) Create(u *dto.UserRequest) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.Users{
		Name:     u.Name,
		Email:    u.Email,
		Password: string(passwordHash),
	}

	return s.repo.Create(user)
}

func (s UsersService) Login(u *dto.UserLoginRequest) (*dto.UserLoginResponse, error) {
	user, err := s.repo.FindByEmail(u.Email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	expiresAt := time.Now().UTC().Add(s.tokenTTL)
	claims := jwt.RegisteredClaims{
		Subject:   user.UUID.String(),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &dto.UserLoginResponse{
		AccessToken: signedToken,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
	}, nil
}

func (s UsersService) ShowMe(uuid string) (*dto.UserSimpleResponse, error) {
	user, err := s.repo.FindByUuid(uuid)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	return &dto.UserSimpleResponse{
		Name:  user.Name,
		Email: user.Email,
	}, nil
}
