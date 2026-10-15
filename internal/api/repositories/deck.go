package repositories

import (
	"context"

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

func (r *deckRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Deck, error) {
	var deck models.Deck
	if err := r.db.WithContext(ctx).First(&deck, "uuid = ?", id).Error; err != nil {
		return nil, err
	}
	return &deck, nil
}

func (r *deckRepository) ListByUser(ctx context.Context, userUUID uuid.UUID, gameFilter *models.GameName) ([]*models.Deck, error) {
	var decks []*models.Deck
	query := r.db.WithContext(ctx).Where("user_uuid = ?", userUUID)
	
	if gameFilter != nil {
		query = query.Where("game = ?", *gameFilter)
	}

	if err := query.Find(&decks).Error; err != nil {
		return nil, err
	}
	return decks, nil
}
