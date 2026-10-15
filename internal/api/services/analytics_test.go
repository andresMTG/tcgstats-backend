package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/services"
	"github.com/andresMTG/tcgstats-backend/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyticsService_DeckPerformance(t *testing.T) {
	ctx, userID := context.Background(), uuid.New()
	from, to := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	repo := &mocks.AnalyticsRepositoryMock{}
	repo.On("DeckPerformance", ctx, userID, &from, &to).Return([]models.DeckPerformance{
		{DeckName: "A", MatchRecord: models.MatchRecord{Matches: 4, Wins: 2, Draws: 1, Losses: 1}},
		{DeckName: "B", MatchRecord: models.MatchRecord{Matches: 0}},
	}, nil).Once()

	got, err := services.NewAnalyticsService(repo).DeckPerformance(ctx, userID, &from, &to)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, 50.0, got[0].WinRate)
	assert.Equal(t, int64(1), got[0].Draws)
	assert.Zero(t, got[1].WinRate)
	repo.AssertExpectations(t)
}

func TestAnalyticsService_OpponentPerformance(t *testing.T) {
	ctx, userID := context.Background(), uuid.New()
	repo := &mocks.AnalyticsRepositoryMock{}
	repo.On("OpponentPerformance", ctx, userID, (*time.Time)(nil), (*time.Time)(nil)).
		Return([]models.OpponentPerformance{{
			OpponentName: "Rival",
			MatchRecord:  models.MatchRecord{Matches: 3, Wins: 1, Draws: 1, Losses: 1},
		}}, nil).Once()

	got, err := services.NewAnalyticsService(repo).OpponentPerformance(ctx, userID, nil, nil)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.InDelta(t, 100.0/3, got[0].WinRate, 0.001)
	assert.Equal(t, int64(1), got[0].Draws)
	repo.AssertExpectations(t)
}

func TestAnalyticsService_PhasePerformance(t *testing.T) {
	ctx, userID := context.Background(), uuid.New()
	wantErr := errors.New("query failed")
	repo := &mocks.AnalyticsRepositoryMock{}
	repo.On("PhasePerformance", ctx, userID, (*time.Time)(nil), (*time.Time)(nil)).
		Return([]models.PhasePerformance(nil), wantErr).Once()

	got, err := services.NewAnalyticsService(repo).PhasePerformance(ctx, userID, nil, nil)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, wantErr)
	repo.AssertExpectations(t)
}

func TestAnalyticsService_TournamentScopePerformance(t *testing.T) {
	ctx, userID := context.Background(), uuid.New()
	repo := &mocks.AnalyticsRepositoryMock{}
	repo.On("TournamentScopePerformance", ctx, userID, (*time.Time)(nil), (*time.Time)(nil)).
		Return([]models.TournamentScopePerformance{{
			Scope:       models.ScopeMajor,
			MatchRecord: models.MatchRecord{Matches: 5, Wins: 4, Losses: 1},
		}}, nil).Once()

	got, err := services.NewAnalyticsService(repo).TournamentScopePerformance(ctx, userID, nil, nil)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 80.0, got[0].WinRate)
	repo.AssertExpectations(t)
}
