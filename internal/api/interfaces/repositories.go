package interfaces

import (
	"context"
	"errors"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/google/uuid"
)

var ErrRefreshTokenNotFound = errors.New("refresh token session not found")

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type RefreshTokenStore interface {
	SaveRefreshToken(ctx context.Context, tokenID string, userUUID uuid.UUID, ttl time.Duration) error
	ConsumeRefreshToken(ctx context.Context, tokenID string) (uuid.UUID, error)
}

type DeckRepository interface {
	Create(ctx context.Context, deck *models.Deck) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Deck, error)
	ListByUser(ctx context.Context, userUUID uuid.UUID, gameFilter *models.GameName) ([]*models.Deck, error)
}

type TournamentRepository interface {
	Create(ctx context.Context, tournament *models.Tournament) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tournament, error)
	ListByUser(ctx context.Context, userUUID uuid.UUID) ([]*models.Tournament, error)
}

type OpponentRepository interface {
	Create(ctx context.Context, opponent *models.Opponent) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Opponent, error)
	ListByUser(ctx context.Context, userUUID uuid.UUID) ([]*models.Opponent, error)
}

type MatchRepository interface {
	Create(ctx context.Context, match *models.Match) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Match, error)
	ListByTournament(ctx context.Context, tournamentUUID uuid.UUID, resultFilter *models.ResultName) ([]*models.Match, error)
	ListByDeck(ctx context.Context, deckUUID uuid.UUID, resultFilter *models.ResultName) ([]*models.Match, error)
}

type AnalyticsRepository interface {
	DeckPerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time) ([]models.DeckPerformance, error)
	OpponentPerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time) ([]models.OpponentPerformance, error)
	PhasePerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time) ([]models.PhasePerformance, error)
	TournamentScopePerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time) ([]models.TournamentScopePerformance, error)
}
