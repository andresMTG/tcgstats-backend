package interfaces

import (
	"context"
	"errors"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type UserService interface {
	CreateUser(ctx context.Context, name, email, password string) (*models.User, error)
	Login(ctx context.Context, email, password string) (*models.AuthTokens, error)
	Refresh(ctx context.Context, token string) (*models.AuthTokens, error)
	Logout(ctx context.Context, token string) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type DeckService interface {
	CreateDeck(ctx context.Context, deck *models.Deck) error
	GetDeckByID(ctx context.Context, id uuid.UUID) (*models.Deck, error)
	ListUserDecks(ctx context.Context, userUUID uuid.UUID, gameFilter *models.GameName) ([]*models.Deck, error)
}

type TournamentService interface {
	CreateTournament(ctx context.Context, tournament *models.Tournament) error
	GetTournamentByID(ctx context.Context, id uuid.UUID) (*models.Tournament, error)
	ListUserTournaments(ctx context.Context, userUUID uuid.UUID) ([]*models.Tournament, error)
}

type OpponentService interface {
	CreateOpponent(ctx context.Context, opponent *models.Opponent) error
	GetOpponentByID(ctx context.Context, id uuid.UUID) (*models.Opponent, error)
	ListUserOpponents(ctx context.Context, userUUID uuid.UUID) ([]*models.Opponent, error)
}

type MatchService interface {
	CreateMatch(ctx context.Context, match *models.Match) error
	GetMatchByID(ctx context.Context, id uuid.UUID) (*models.Match, error)
	ListTournamentMatches(ctx context.Context, tournamentUUID uuid.UUID, resultFilter *models.ResultName) ([]*models.Match, error)
	ListDeckMatches(ctx context.Context, deckUUID uuid.UUID, resultFilter *models.ResultName) ([]*models.Match, error)
}

type AnalyticsService interface {
	DeckPerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.DeckPerformance, error)
	DeckMatchupPerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.DeckMatchupPerformance, error)
	OpponentPerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.OpponentPerformance, error)
	PhasePerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.PhasePerformance, error)
	TournamentScopePerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.TournamentScopePerformance, error)
}
