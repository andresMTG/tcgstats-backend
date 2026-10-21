package services

import (
	"context"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/google/uuid"
)

type matchService struct {
	repo interfaces.MatchRepository
}

func NewMatchService(repo interfaces.MatchRepository) interfaces.MatchService {
	return &matchService{repo: repo}
}

func (s *matchService) CreateMatch(ctx context.Context, match *models.Match) error {
	match.UUID = uuid.New()
	return s.repo.Create(ctx, match)
}

func (s *matchService) UpdateMatch(ctx context.Context, match *models.Match, userUUID uuid.UUID) error {
	return s.repo.Update(ctx, match, userUUID)
}

func (s *matchService) DeleteMatch(ctx context.Context, id, userUUID uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id, userUUID)
}

func (s *matchService) GetMatchByID(ctx context.Context, id uuid.UUID) (*models.Match, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *matchService) ListTournamentMatches(ctx context.Context, tournamentUUID uuid.UUID, resultFilter *models.ResultName) ([]*models.Match, error) {
	return s.repo.ListByTournament(ctx, tournamentUUID, resultFilter)
}

func (s *matchService) ListDeckMatches(ctx context.Context, deckUUID uuid.UUID, resultFilter *models.ResultName) ([]*models.Match, error) {
	return s.repo.ListByDeck(ctx, deckUUID, resultFilter)
}
