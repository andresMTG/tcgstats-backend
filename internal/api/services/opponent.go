package services

import (
	"context"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/google/uuid"
)

type opponentService struct {
	repo interfaces.OpponentRepository
}

func NewOpponentService(repo interfaces.OpponentRepository) interfaces.OpponentService {
	return &opponentService{repo: repo}
}

func (s *opponentService) CreateOpponent(ctx context.Context, opponent *models.Opponent) error {
	opponent.UUID = uuid.New()
	return s.repo.Create(ctx, opponent)
}

func (s *opponentService) UpdateOpponent(ctx context.Context, opponent *models.Opponent) error {
	return s.repo.Update(ctx, opponent)
}

func (s *opponentService) DeleteOpponent(ctx context.Context, id, userUUID uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id, userUUID)
}

func (s *opponentService) GetOpponentByID(ctx context.Context, id uuid.UUID) (*models.Opponent, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *opponentService) ListUserOpponents(ctx context.Context, userUUID uuid.UUID) ([]*models.Opponent, error) {
	return s.repo.ListByUser(ctx, userUUID)
}
