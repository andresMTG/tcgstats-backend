package mocks

import (
	"context"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// UserServiceMock provides a testify mock for interfaces.UserService.
type UserServiceMock struct {
	mock.Mock
}

func (s *UserServiceMock) CreateUser(ctx context.Context, name, email, password string) (*models.User, error) {
	args := s.Called(ctx, name, email, password)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (s *UserServiceMock) Login(ctx context.Context, email, password string) (*models.AuthTokens, error) {
	args := s.Called(ctx, email, password)
	tokens, _ := args.Get(0).(*models.AuthTokens)
	return tokens, args.Error(1)
}

func (s *UserServiceMock) Refresh(ctx context.Context, token string) (*models.AuthTokens, error) {
	args := s.Called(ctx, token)
	tokens, _ := args.Get(0).(*models.AuthTokens)
	return tokens, args.Error(1)
}

func (s *UserServiceMock) Logout(ctx context.Context, token string) error {
	return s.Called(ctx, token).Error(0)
}

func (s *UserServiceMock) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := s.Called(ctx, id)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

var _ interfaces.UserService = (*UserServiceMock)(nil)

type DeckServiceMock struct{ mock.Mock }

func (s *DeckServiceMock) CreateDeck(ctx context.Context, deck *models.Deck) error {
	return s.Called(ctx, deck).Error(0)
}
func (s *DeckServiceMock) GetDeckByID(ctx context.Context, id uuid.UUID) (*models.Deck, error) {
	args := s.Called(ctx, id)
	deck, _ := args.Get(0).(*models.Deck)
	return deck, args.Error(1)
}
func (s *DeckServiceMock) ListUserDecks(ctx context.Context, userID uuid.UUID, filter *models.GameName) ([]*models.Deck, error) {
	args := s.Called(ctx, userID, filter)
	decks, _ := args.Get(0).([]*models.Deck)
	return decks, args.Error(1)
}

type TournamentServiceMock struct{ mock.Mock }

func (s *TournamentServiceMock) CreateTournament(ctx context.Context, tournament *models.Tournament) error {
	return s.Called(ctx, tournament).Error(0)
}
func (s *TournamentServiceMock) GetTournamentByID(ctx context.Context, id uuid.UUID) (*models.Tournament, error) {
	args := s.Called(ctx, id)
	tournament, _ := args.Get(0).(*models.Tournament)
	return tournament, args.Error(1)
}
func (s *TournamentServiceMock) ListUserTournaments(ctx context.Context, userID uuid.UUID) ([]*models.Tournament, error) {
	args := s.Called(ctx, userID)
	tournaments, _ := args.Get(0).([]*models.Tournament)
	return tournaments, args.Error(1)
}

type OpponentServiceMock struct{ mock.Mock }

func (s *OpponentServiceMock) CreateOpponent(ctx context.Context, opponent *models.Opponent) error {
	return s.Called(ctx, opponent).Error(0)
}
func (s *OpponentServiceMock) GetOpponentByID(ctx context.Context, id uuid.UUID) (*models.Opponent, error) {
	args := s.Called(ctx, id)
	opponent, _ := args.Get(0).(*models.Opponent)
	return opponent, args.Error(1)
}
func (s *OpponentServiceMock) ListUserOpponents(ctx context.Context, userID uuid.UUID) ([]*models.Opponent, error) {
	args := s.Called(ctx, userID)
	opponents, _ := args.Get(0).([]*models.Opponent)
	return opponents, args.Error(1)
}

type MatchServiceMock struct{ mock.Mock }

func (s *MatchServiceMock) CreateMatch(ctx context.Context, match *models.Match) error {
	return s.Called(ctx, match).Error(0)
}
func (s *MatchServiceMock) GetMatchByID(ctx context.Context, id uuid.UUID) (*models.Match, error) {
	args := s.Called(ctx, id)
	match, _ := args.Get(0).(*models.Match)
	return match, args.Error(1)
}
func (s *MatchServiceMock) ListTournamentMatches(ctx context.Context, tournamentID uuid.UUID, filter *models.ResultName) ([]*models.Match, error) {
	args := s.Called(ctx, tournamentID, filter)
	matches, _ := args.Get(0).([]*models.Match)
	return matches, args.Error(1)
}
func (s *MatchServiceMock) ListDeckMatches(ctx context.Context, deckID uuid.UUID, filter *models.ResultName) ([]*models.Match, error) {
	args := s.Called(ctx, deckID, filter)
	matches, _ := args.Get(0).([]*models.Match)
	return matches, args.Error(1)
}

var (
	_ interfaces.DeckService       = (*DeckServiceMock)(nil)
	_ interfaces.TournamentService = (*TournamentServiceMock)(nil)
	_ interfaces.OpponentService   = (*OpponentServiceMock)(nil)
	_ interfaces.MatchService      = (*MatchServiceMock)(nil)
)

type AnalyticsServiceMock struct{ mock.Mock }

func (s *AnalyticsServiceMock) DeckPerformance(ctx context.Context, userID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.DeckPerformance, error) {
	args := s.Called(ctx, userID, from, to, game)
	performance, _ := args.Get(0).([]models.DeckPerformance)
	return performance, args.Error(1)
}
func (s *AnalyticsServiceMock) DeckMatchupPerformance(ctx context.Context, userID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.DeckMatchupPerformance, error) {
	args := s.Called(ctx, userID, from, to, game)
	performance, _ := args.Get(0).([]models.DeckMatchupPerformance)
	return performance, args.Error(1)
}
func (s *AnalyticsServiceMock) OpponentPerformance(ctx context.Context, userID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.OpponentPerformance, error) {
	args := s.Called(ctx, userID, from, to, game)
	performance, _ := args.Get(0).([]models.OpponentPerformance)
	return performance, args.Error(1)
}
func (s *AnalyticsServiceMock) PhasePerformance(ctx context.Context, userID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.PhasePerformance, error) {
	args := s.Called(ctx, userID, from, to, game)
	performance, _ := args.Get(0).([]models.PhasePerformance)
	return performance, args.Error(1)
}
func (s *AnalyticsServiceMock) TournamentScopePerformance(ctx context.Context, userID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.TournamentScopePerformance, error) {
	args := s.Called(ctx, userID, from, to, game)
	performance, _ := args.Get(0).([]models.TournamentScopePerformance)
	return performance, args.Error(1)
}

var _ interfaces.AnalyticsService = (*AnalyticsServiceMock)(nil)
