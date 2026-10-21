package repositories

import (
	"context"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type deckRepository struct {
	db *gorm.DB
}

func NewDeckRepository(db *gorm.DB) interfaces.DeckRepository {
	return &deckRepository{db: db}
}

func (r *deckRepository) Create(ctx context.Context, deck *models.Deck) error {
	return r.db.WithContext(ctx).Create(deck).Error
}

func (r *deckRepository) Update(ctx context.Context, deck *models.Deck) error {
	result := r.db.WithContext(ctx).Model(&models.Deck{}).
		Where("uuid = ? AND user_uuid = ? AND deleted_at IS NULL", deck.UUID, deck.UserUUID).
		Updates(map[string]interface{}{
			"name":     deck.Name,
			"game":     deck.Game,
			"decklist": deck.Decklist,
		})
	return affectedRecord(result)
}

func (r *deckRepository) SoftDelete(ctx context.Context, id, userUUID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&models.Deck{}).
		Where("uuid = ? AND user_uuid = ? AND deleted_at IS NULL", id, userUUID).
		Update("deleted_at", time.Now())
	return affectedRecord(result)
}

func (r *deckRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Deck, error) {
	var deck models.Deck
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&deck, "uuid = ?", id).Error; err != nil {
		return nil, err
	}
	return &deck, nil
}

func (r *deckRepository) ListByUser(ctx context.Context, userUUID uuid.UUID, gameFilter *models.GameName) ([]*models.Deck, error) {
	var decks []*models.Deck
	query := r.db.WithContext(ctx).Where("user_uuid = ? AND deleted_at IS NULL", userUUID)

	if gameFilter != nil {
		query = query.Where("game = ?", *gameFilter)
	}

	if err := query.Find(&decks).Error; err != nil {
		return nil, err
	}
	return decks, nil
}
