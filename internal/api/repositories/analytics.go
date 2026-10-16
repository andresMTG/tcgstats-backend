package repositories

import (
	"context"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type analyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) interfaces.AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) DeckPerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.DeckPerformance, error) {
	performance := make([]models.DeckPerformance, 0)
	query := r.db.WithContext(ctx).
		Table("matches AS m").
		Select("d.uuid AS deck_uuid, d.name AS deck_name, "+matchRecordSelect).
		Joins("JOIN tournaments AS t ON t.uuid = m.tournament_uuid").
		Joins("JOIN decks AS d ON d.uuid = t.deck_uuid AND d.user_uuid = ?", userUUID).
		Where("t.user_uuid = ?", userUUID)
	query = applyTournamentDateRange(query, from, to)
	query = applyTournamentGameFilter(query, game)

	if err := query.Group("d.uuid, d.name").Order("d.name, d.uuid").Scan(&performance).Error; err != nil {
		return nil, err
	}
	return performance, nil
}

func (r *analyticsRepository) DeckMatchupPerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.DeckMatchupPerformance, error) {
	performance := make([]models.DeckMatchupPerformance, 0)
	query := r.db.WithContext(ctx).
		Table("matches AS m").
		Select("d.uuid AS deck_uuid, d.name AS deck_name, m.opposing_deck_archetype AS opponent_deck_archetype, "+matchRecordSelect).
		Joins("JOIN tournaments AS t ON t.uuid = m.tournament_uuid").
		Joins("JOIN decks AS d ON d.uuid = t.deck_uuid AND d.user_uuid = ?", userUUID).
		Where("t.user_uuid = ?", userUUID)
	query = applyTournamentDateRange(query, from, to)
	query = applyTournamentGameFilter(query, game)

	if err := query.Group("d.uuid, d.name, m.opposing_deck_archetype").
		Order("d.name, m.opposing_deck_archetype").
		Scan(&performance).Error; err != nil {
		return nil, err
	}
	return performance, nil
}

func (r *analyticsRepository) OpponentPerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.OpponentPerformance, error) {
	performance := make([]models.OpponentPerformance, 0)
	query := r.db.WithContext(ctx).
		Table("matches AS m").
		Select("o.uuid AS opponent_uuid, o.name AS opponent_name, "+matchRecordSelect).
		Joins("JOIN tournaments AS t ON t.uuid = m.tournament_uuid").
		Joins("JOIN opponents AS o ON o.uuid = m.opponent_uuid AND o.user_uuid = ?", userUUID).
		Where("t.user_uuid = ?", userUUID)
	query = applyTournamentDateRange(query, from, to)
	query = applyTournamentGameFilter(query, game)

	if err := query.Group("o.uuid, o.name").Order("o.name, o.uuid").Scan(&performance).Error; err != nil {
		return nil, err
	}
	return performance, nil
}

func (r *analyticsRepository) PhasePerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.PhasePerformance, error) {
	performance := make([]models.PhasePerformance, 0)
	query := r.db.WithContext(ctx).
		Table("matches AS m").
		Select("m.phase AS phase, "+matchRecordSelect).
		Joins("JOIN tournaments AS t ON t.uuid = m.tournament_uuid").
		Where("t.user_uuid = ?", userUUID)
	query = applyTournamentDateRange(query, from, to)
	query = applyTournamentGameFilter(query, game)

	if err := query.Group("m.phase").Order("m.phase").Scan(&performance).Error; err != nil {
		return nil, err
	}
	return performance, nil
}

func (r *analyticsRepository) TournamentScopePerformance(ctx context.Context, userUUID uuid.UUID, from, to *time.Time, game *models.GameName) ([]models.TournamentScopePerformance, error) {
	performance := make([]models.TournamentScopePerformance, 0)
	query := r.db.WithContext(ctx).
		Table("matches AS m").
		Select("t.scope AS scope, "+matchRecordSelect).
		Joins("JOIN tournaments AS t ON t.uuid = m.tournament_uuid").
		Where("t.user_uuid = ?", userUUID)
	query = applyTournamentDateRange(query, from, to)
	query = applyTournamentGameFilter(query, game)

	if err := query.Group("t.scope").Order("t.scope").Scan(&performance).Error; err != nil {
		return nil, err
	}
	return performance, nil
}

const matchRecordSelect = `
	COUNT(*) AS matches,
	COUNT(*) FILTER (WHERE m.result = 'win') AS wins,
	COUNT(*) FILTER (WHERE m.result = 'draw') AS draws,
	COUNT(*) FILTER (WHERE m.result = 'lose') AS losses`

func applyTournamentDateRange(query *gorm.DB, from, to *time.Time) *gorm.DB {
	if from != nil {
		query = query.Where("t.date >= ?", from.Format("2006-01-02"))
	}
	if to != nil {
		query = query.Where("t.date <= ?", to.Format("2006-01-02"))
	}
	return query
}

func applyTournamentGameFilter(query *gorm.DB, game *models.GameName) *gorm.DB {
	if game != nil {
		query = query.Where("t.game = ?", *game)
	}
	return query
}

var _ interfaces.AnalyticsRepository = (*analyticsRepository)(nil)
