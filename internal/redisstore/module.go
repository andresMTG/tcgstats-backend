package redisstore

import (
	"context"
	"fmt"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/andresMTG/tcgstats-backend/pkg/env"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

type Configuration struct {
	Address  string `env:"REDIS_ADDR" envDefault:"127.0.0.1:6379"`
	Password string `env:"REDIS_PASSWORD"`
	Database int    `env:"REDIS_DB" envDefault:"0"`
}

func Module() fx.Option {
	return fx.Module("redis", fx.Options(
		fx.Provide(
			env.LoadEnvConfiguration[Configuration],
			NewClient,
			NewRefreshTokenStore,
		),
	))
}

func NewClient(lifecycle fx.Lifecycle, config *Configuration) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Password: config.Password,
		DB:       config.Database,
	})

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			pingContext, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := client.Ping(pingContext).Err(); err != nil {
				return fmt.Errorf("connect to redis: %w", err)
			}
			return nil
		},
		OnStop: func(context.Context) error {
			return client.Close()
		},
	})

	return client, nil
}

func NewRefreshTokenStore(client *redis.Client) interfaces.RefreshTokenStore {
	return &refreshTokenStore{client: client}
}

type refreshTokenStore struct {
	client *redis.Client
}

func (s *refreshTokenStore) SaveRefreshToken(ctx context.Context, tokenID string, userUUID uuid.UUID, ttl time.Duration) error {
	return s.client.Set(ctx, refreshTokenKey(tokenID), userUUID.String(), ttl).Err()
}

func (s *refreshTokenStore) ConsumeRefreshToken(ctx context.Context, tokenID string) (uuid.UUID, error) {
	value, err := s.client.GetDel(ctx, refreshTokenKey(tokenID)).Result()
	if err == redis.Nil {
		return uuid.Nil, interfaces.ErrRefreshTokenNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("consume refresh token session: %w", err)
	}

	userUUID, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse refresh token owner: %w", err)
	}
	return userUUID, nil
}

func refreshTokenKey(tokenID string) string {
	return "tcgstats:refresh:" + tokenID
}

var _ interfaces.RefreshTokenStore = (*refreshTokenStore)(nil)
