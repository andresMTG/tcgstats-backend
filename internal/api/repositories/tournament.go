package repositories

import (
	"context"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type tournamentRepository struct {
	db *gorm.DB
}

func NewTournamentRepository(db *gorm.DB) interfaces.TournamentRepository {
	return &tournamentRepository{db: db}
}

func (r *tournamentRepository) Create(ctx context.Context, tournament *models.Tournament) error {
	return r.db.WithContext(ctx).Create(tournament).Error
}

func (r *tournamentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tournament, error) {
	var tournament models.Tournament
	if err := r.db.WithContext(ctx).First(&tournament, "uuid = ?", id).Error; err != nil {
		return nil, err
	}
	return &tournament, nil
}

func (r *tournamentRepository) ListByUser(ctx context.Context, userUUID uuid.UUID) ([]*models.Tournament, error) {
	var tournaments []*models.Tournament
	if err := r.db.WithContext(ctx).Where("user_uuid = ?", userUUID).Find(&tournaments).Error; err != nil {
		return nil, err
	}
	return tournaments, nil
}
