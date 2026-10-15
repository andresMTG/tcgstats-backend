package repositories

import (
	"context"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type matchRepository struct {
	db *gorm.DB
}

func NewMatchRepository(db *gorm.DB) interfaces.MatchRepository {
	return &matchRepository{db: db}
}

func (r *matchRepository) Create(ctx context.Context, match *models.Match) error {
	return r.db.WithContext(ctx).Create(match).Error
}

func (r *matchRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Match, error) {
	var match models.Match
	if err := r.db.WithContext(ctx).First(&match, "uuid = ?", id).Error; err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *matchRepository) ListByTournament(ctx context.Context, tournamentUUID uuid.UUID, resultFilter *models.ResultName) ([]*models.Match, error) {
	var matches []*models.Match
	query := r.db.WithContext(ctx).Where("tournament_uuid = ?", tournamentUUID)

	if resultFilter != nil {
		query = query.Where("result = ?", *resultFilter)
	}

	if err := query.Find(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}

func (r *matchRepository) ListByDeck(ctx context.Context, deckUUID uuid.UUID, resultFilter *models.ResultName) ([]*models.Match, error) {
	var matches []*models.Match
	// Realiza um JOIN com tournaments para encontrar as matches do deck
	query := r.db.WithContext(ctx).
		Joins("JOIN tournaments ON tournaments.uuid = matches.tournament_uuid").
		Where("tournaments.deck_uuid = ?", deckUUID)

	if resultFilter != nil {
		query = query.Where("matches.result = ?", *resultFilter)
	}

	if err := query.Find(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}
