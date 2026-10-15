package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/auth"
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/andresMTG/tcgstats-backend/pkg/env"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type JWTConfig struct {
	Secret                string `env:"JWT_SECRET,required"`
	Expiration            int    `env:"JWT_EXPIRATION_MINUTES" envDefault:"15"`
	RefreshExpirationDays int    `env:"JWT_REFRESH_EXPIRATION_DAYS" envDefault:"30"`
}

var (
	ErrInvalidCredentials  = interfaces.ErrInvalidCredentials
	ErrInvalidRefreshToken = interfaces.ErrInvalidRefreshToken
)

type userService struct {
	repo         interfaces.UserRepository
	refreshStore interfaces.RefreshTokenStore
	config       *JWTConfig
}

func NewUserService(repo interfaces.UserRepository, refreshStore interfaces.RefreshTokenStore) (interfaces.UserService, error) {
	config, err := env.LoadEnvConfiguration[JWTConfig]()
	if err != nil {
		return nil, err
	}
	if config.Expiration <= 0 || config.RefreshExpirationDays <= 0 {
		return nil, errors.New("JWT expirations must be positive")
	}
	return &userService{repo: repo, refreshStore: refreshStore, config: config}, nil
}

func (s *userService) CreateUser(ctx context.Context, name, email, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		UUID:     uuid.New(),
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (*models.AuthTokens, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	tokens, refreshTokenID, err := s.createTokens(user.UUID)
	if err != nil {
		return nil, err
	}
	if err := s.refreshStore.SaveRefreshToken(ctx, refreshTokenID, user.UUID, tokens.RefreshTokenTTL); err != nil {
		return nil, fmt.Errorf("save refresh token session: %w", err)
	}
	return tokens, nil
}

func (s *userService) Refresh(ctx context.Context, tokenString string) (*models.AuthTokens, error) {
	claims, err := auth.ParseToken(tokenString, s.config.Secret)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}
	if claims.Type != auth.RefreshTokenType || claims.ID == "" {
		return nil, ErrInvalidRefreshToken
	}

	userUUID, err := s.refreshStore.ConsumeRefreshToken(ctx, claims.ID)
	if errors.Is(err, interfaces.ErrRefreshTokenNotFound) || err == nil && claims.Subject != userUUID.String() {
		return nil, ErrInvalidRefreshToken
	}
	if err != nil {
		return nil, fmt.Errorf("consume refresh token session: %w", err)
	}

	tokens, refreshTokenID, err := s.createTokens(userUUID)
	if err != nil {
		return nil, err
	}
	if err := s.refreshStore.SaveRefreshToken(ctx, refreshTokenID, userUUID, tokens.RefreshTokenTTL); err != nil {
		return nil, fmt.Errorf("save rotated refresh token session: %w", err)
	}
	return tokens, nil
}

func (s *userService) Logout(ctx context.Context, tokenString string) error {
	claims, err := auth.ParseToken(tokenString, s.config.Secret)
	if err != nil || claims.Type != auth.RefreshTokenType || claims.ID == "" {
		return ErrInvalidRefreshToken
	}

	userUUID, err := s.refreshStore.ConsumeRefreshToken(ctx, claims.ID)
	if errors.Is(err, interfaces.ErrRefreshTokenNotFound) || err == nil && claims.Subject != userUUID.String() {
		return ErrInvalidRefreshToken
	}
	if err != nil {
		return fmt.Errorf("revoke refresh token session: %w", err)
	}
	return nil
}

func (s *userService) createTokens(userUUID uuid.UUID) (*models.AuthTokens, string, error) {
	now := time.Now()
	accessTTL := time.Duration(s.config.Expiration) * time.Minute
	refreshTTL := time.Duration(s.config.RefreshExpirationDays) * 24 * time.Hour
	refreshTokenID := uuid.NewString()

	accessToken, err := signToken(&auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userUUID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
		},
		Type: auth.AccessTokenType,
	}, s.config.Secret)
	if err != nil {
		return nil, "", err
	}
	refreshToken, err := signToken(&auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshTokenID,
			Subject:   userUUID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTTL)),
		},
		Type: auth.RefreshTokenType,
	}, s.config.Secret)
	if err != nil {
		return nil, "", err
	}
	return &models.AuthTokens{
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
	}, refreshTokenID, nil
}

func signToken(claims *auth.Claims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return tokenString, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.repo.GetByID(ctx, id)
}
