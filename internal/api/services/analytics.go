package services

import (
	"context"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/google/uuid"
)

type analyticsService struct {
	repo interfaces.AnalyticsRepository
}

func NewAnalyticsService(repo interfaces.AnalyticsRepository) interfaces.AnalyticsService {
	return &analyticsService{repo: repo}
}

func (s *analyticsService) DeckPerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.DeckPerformance, error) {
	performance, err := s.repo.DeckPerformance(ctx, userUUID, from, to, game)
	if err != nil {
		return nil, err
	}
	for i := range performance {
		performance[i].WinRate = calculateWinRate(performance[i].Wins, performance[i].Matches)
	}
	return performance, nil
}

func (s *analyticsService) DeckMatchupPerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.DeckMatchupPerformance, error) {
	performance, err := s.repo.DeckMatchupPerformance(ctx, userUUID, from, to, game)
	if err != nil {
		return nil, err
	}
	for i := range performance {
		performance[i].WinRate = calculateWinRate(performance[i].Wins, performance[i].Matches)
	}
	return performance, nil
}

func (s *analyticsService) OpponentPerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.OpponentPerformance, error) {
	performance, err := s.repo.OpponentPerformance(ctx, userUUID, from, to, game)
	if err != nil {
		return nil, err
	}
	for i := range performance {
		performance[i].WinRate = calculateWinRate(performance[i].Wins, performance[i].Matches)
	}
	return performance, nil
}

func (s *analyticsService) PhasePerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.PhasePerformance, error) {
	performance, err := s.repo.PhasePerformance(ctx, userUUID, from, to, game)
	if err != nil {
		return nil, err
	}
	for i := range performance {
		performance[i].WinRate = calculateWinRate(performance[i].Wins, performance[i].Matches)
	}
	return performance, nil
}

func (s *analyticsService) TournamentScopePerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.TournamentScopePerformance, error) {
	performance, err := s.repo.TournamentScopePerformance(ctx, userUUID, from, to, game)
	if err != nil {
		return nil, err
	}
	for i := range performance {
		performance[i].WinRate = calculateWinRate(performance[i].Wins, performance[i].Matches)
	}
	return performance, nil
}

func calculateWinRate(wins, matches int64) float64 {
	if matches == 0 {
		return 0
	}
	return float64(wins) * 100 / float64(matches)
}

var _ interfaces.AnalyticsService = (*analyticsService)(nil)
