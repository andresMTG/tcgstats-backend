package repositories

import (
	"context"
	"time"

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

func (r *tournamentRepository) Update(ctx context.Context, tournament *models.Tournament) error {
	query := r.db.WithContext(ctx).Model(&models.Tournament{}).
		Where("uuid = ? AND user_uuid = ? AND deleted_at IS NULL", tournament.UUID, tournament.UserUUID)
	if tournament.DeckUUID != nil {
		query = query.Where(
			"EXISTS (SELECT 1 FROM decks WHERE decks.uuid = ? AND decks.user_uuid = ? AND decks.deleted_at IS NULL)",
			*tournament.DeckUUID,
			tournament.UserUUID,
		)
	}
	result := query.Updates(map[string]interface{}{
		"name":                tournament.Name,
		"date":                tournament.Date,
		"game":                tournament.Game,
		"scope":               tournament.Scope,
		"is_official_circuit": tournament.IsOfficialCircuit,
		"deck_uuid":           tournament.DeckUUID,
	})
	return affectedRecord(result)
}

func (r *tournamentRepository) SoftDelete(ctx context.Context, id, userUUID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&models.Tournament{}).
		Where("uuid = ? AND user_uuid = ? AND deleted_at IS NULL", id, userUUID).
		Update("deleted_at", time.Now())
	return affectedRecord(result)
}

func (r *tournamentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tournament, error) {
	var tournament models.Tournament
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&tournament, "uuid = ?", id).Error; err != nil {
		return nil, err
	}
	return &tournament, nil
}

func (r *tournamentRepository) ListByUser(ctx context.Context, userUUID uuid.UUID) ([]*models.Tournament, error) {
	var tournaments []*models.Tournament
	if err := r.db.WithContext(ctx).Where("user_uuid = ? AND deleted_at IS NULL", userUUID).Find(&tournaments).Error; err != nil {
		return nil, err
	}
	return tournaments, nil
}
