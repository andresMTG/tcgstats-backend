package mocks

import (
	"context"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type DeckRepositoryMock struct{ mock.Mock }

func (r *DeckRepositoryMock) Create(ctx context.Context, deck *models.Deck) error {
	return r.Called(ctx, deck).Error(0)
}
func (r *DeckRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*models.Deck, error) {
	args := r.Called(ctx, id)
	deck, _ := args.Get(0).(*models.Deck)
	return deck, args.Error(1)
}
func (r *DeckRepositoryMock) ListByUser(ctx context.Context, userID uuid.UUID, filter *models.GameName) ([]*models.Deck, error) {
	args := r.Called(ctx, userID, filter)
	decks, _ := args.Get(0).([]*models.Deck)
	return decks, args.Error(1)
}

type TournamentRepositoryMock struct{ mock.Mock }

func (r *TournamentRepositoryMock) Create(ctx context.Context, tournament *models.Tournament) error {
	return r.Called(ctx, tournament).Error(0)
}
func (r *TournamentRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*models.Tournament, error) {
	args := r.Called(ctx, id)
	tournament, _ := args.Get(0).(*models.Tournament)
	return tournament, args.Error(1)
}
func (r *TournamentRepositoryMock) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Tournament, error) {
	args := r.Called(ctx, userID)
	tournaments, _ := args.Get(0).([]*models.Tournament)
	return tournaments, args.Error(1)
}

type OpponentRepositoryMock struct{ mock.Mock }

func (r *OpponentRepositoryMock) Create(ctx context.Context, opponent *models.Opponent) error {
	return r.Called(ctx, opponent).Error(0)
}
func (r *OpponentRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*models.Opponent, error) {
	args := r.Called(ctx, id)
	opponent, _ := args.Get(0).(*models.Opponent)
	return opponent, args.Error(1)
}
func (r *OpponentRepositoryMock) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Opponent, error) {
	args := r.Called(ctx, userID)
	opponents, _ := args.Get(0).([]*models.Opponent)
	return opponents, args.Error(1)
}

type MatchRepositoryMock struct{ mock.Mock }

func (r *MatchRepositoryMock) Create(ctx context.Context, match *models.Match) error {
	return r.Called(ctx, match).Error(0)
}
func (r *MatchRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*models.Match, error) {
	args := r.Called(ctx, id)
	match, _ := args.Get(0).(*models.Match)
	return match, args.Error(1)
}
func (r *MatchRepositoryMock) ListByTournament(ctx context.Context, tournamentID uuid.UUID, filter *models.ResultName) ([]*models.Match, error) {
	args := r.Called(ctx, tournamentID, filter)
	matches, _ := args.Get(0).([]*models.Match)
	return matches, args.Error(1)
}
func (r *MatchRepositoryMock) ListByDeck(ctx context.Context, deckID uuid.UUID, filter *models.ResultName) ([]*models.Match, error) {
	args := r.Called(ctx, deckID, filter)
	matches, _ := args.Get(0).([]*models.Match)
	return matches, args.Error(1)
}

var (
	_ interfaces.DeckRepository       = (*DeckRepositoryMock)(nil)
	_ interfaces.TournamentRepository = (*TournamentRepositoryMock)(nil)
	_ interfaces.OpponentRepository   = (*OpponentRepositoryMock)(nil)
	_ interfaces.MatchRepository      = (*MatchRepositoryMock)(nil)
)

type RefreshTokenStoreMock struct{ mock.Mock }

func (s *RefreshTokenStoreMock) SaveRefreshToken(ctx context.Context, tokenID string, userID uuid.UUID, ttl time.Duration) error {
	return s.Called(ctx, tokenID, userID, ttl).Error(0)
}
func (s *RefreshTokenStoreMock) ConsumeRefreshToken(ctx context.Context, tokenID string) (uuid.UUID, error) {
	args := s.Called(ctx, tokenID)
	userID, _ := args.Get(0).(uuid.UUID)
	return userID, args.Error(1)
}

var _ interfaces.RefreshTokenStore = (*RefreshTokenStoreMock)(nil)

type AnalyticsRepositoryMock struct{ mock.Mock }

func (r *AnalyticsRepositoryMock) DeckPerformance(ctx context.Context, userID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.DeckPerformance, error) {
	args := r.Called(ctx, userID, from, to, game)
	performance, _ := args.Get(0).([]models.DeckPerformance)
	return performance, args.Error(1)
}
func (r *AnalyticsRepositoryMock) DeckMatchupPerformance(ctx context.Context, userID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.DeckMatchupPerformance, error) {
	args := r.Called(ctx, userID, from, to, game)
	performance, _ := args.Get(0).([]models.DeckMatchupPerformance)
	return performance, args.Error(1)
}
func (r *AnalyticsRepositoryMock) OpponentPerformance(ctx context.Context, userID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.OpponentPerformance, error) {
	args := r.Called(ctx, userID, from, to, game)
	performance, _ := args.Get(0).([]models.OpponentPerformance)
	return performance, args.Error(1)
}
func (r *AnalyticsRepositoryMock) PhasePerformance(ctx context.Context, userID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.PhasePerformance, error) {
	args := r.Called(ctx, userID, from, to, game)
	performance, _ := args.Get(0).([]models.PhasePerformance)
	return performance, args.Error(1)
}
func (r *AnalyticsRepositoryMock) TournamentScopePerformance(ctx context.Context, userID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.TournamentScopePerformance, error) {
	args := r.Called(ctx, userID, from, to, game)
	performance, _ := args.Get(0).([]models.TournamentScopePerformance)
	return performance, args.Error(1)
}

var _ interfaces.AnalyticsRepository = (*AnalyticsRepositoryMock)(nil)
