package services

import (
	"context"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/google/uuid"
)

type deckService struct {
	repo interfaces.DeckRepository
}

func NewDeckService(repo interfaces.DeckRepository) interfaces.DeckService {
	return &deckService{repo: repo}
}

func (s *deckService) CreateDeck(ctx context.Context, deck *models.Deck) error {
	deck.UUID = uuid.New()
	return s.repo.Create(ctx, deck)
}

func (s *deckService) GetDeckByID(ctx context.Context, id uuid.UUID) (*models.Deck, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *deckService) ListUserDecks(ctx context.Context, userUUID uuid.UUID, gameFilter *models.GameName) ([]*models.Deck, error) {
	return s.repo.ListByUser(ctx, userUUID, gameFilter)
}
