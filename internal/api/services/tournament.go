package services

import (
	"context"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/google/uuid"
)

type tournamentService struct {
	repo interfaces.TournamentRepository
}

func NewTournamentService(repo interfaces.TournamentRepository) interfaces.TournamentService {
	return &tournamentService{repo: repo}
}

func (s *tournamentService) CreateTournament(ctx context.Context, tournament *models.Tournament) error {
	tournament.UUID = uuid.New()
	return s.repo.Create(ctx, tournament)
}

func (s *tournamentService) UpdateTournament(ctx context.Context, tournament *models.Tournament) error {
	return s.repo.Update(ctx, tournament)
}

func (s *tournamentService) DeleteTournament(ctx context.Context, id, userUUID uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id, userUUID)
}

func (s *tournamentService) GetTournamentByID(ctx context.Context, id uuid.UUID) (*models.Tournament, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *tournamentService) ListUserTournaments(ctx context.Context, userUUID uuid.UUID) ([]*models.Tournament, error) {
	return s.repo.ListByUser(ctx, userUUID)
}
