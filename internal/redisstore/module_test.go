package redisstore_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/andresMTG/tcgstats-backend/internal/redisstore"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokenStore_ConsumesSessionOnlyOnce(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	store := redisstore.NewRefreshTokenStore(client)
	ctx, tokenID, userID := context.Background(), uuid.NewString(), uuid.New()

	require.NoError(t, store.SaveRefreshToken(ctx, tokenID, userID, time.Minute))
	got, err := store.ConsumeRefreshToken(ctx, tokenID)
	require.NoError(t, err)
	assert.Equal(t, userID, got)

	got, err = store.ConsumeRefreshToken(ctx, tokenID)
	assert.Equal(t, uuid.Nil, got)
	assert.ErrorIs(t, err, interfaces.ErrRefreshTokenNotFound)
}

func TestRefreshTokenStore_ExpiresSession(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	store := redisstore.NewRefreshTokenStore(client)
	ctx, tokenID, userID := context.Background(), uuid.NewString(), uuid.New()

	require.NoError(t, store.SaveRefreshToken(ctx, tokenID, userID, time.Second))
	server.FastForward(2 * time.Second)

	got, err := store.ConsumeRefreshToken(ctx, tokenID)
	assert.Equal(t, uuid.Nil, got)
	assert.True(t, errors.Is(err, interfaces.ErrRefreshTokenNotFound))
}
