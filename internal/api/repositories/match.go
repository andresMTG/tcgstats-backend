package repositories

import (
	"context"
	"time"

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

func (r *matchRepository) Update(ctx context.Context, match *models.Match, userUUID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&models.Match{}).
		Where(
			"uuid = ? AND deleted_at IS NULL AND EXISTS (SELECT 1 FROM tournaments WHERE tournaments.uuid = matches.tournament_uuid AND tournaments.user_uuid = ? AND tournaments.deleted_at IS NULL)",
			match.UUID,
			userUUID,
		).
		Where(
			"EXISTS (SELECT 1 FROM tournaments WHERE tournaments.uuid = ? AND tournaments.user_uuid = ? AND tournaments.deleted_at IS NULL)",
			match.TournamentUUID,
			userUUID,
		).
		Where(
			"EXISTS (SELECT 1 FROM opponents WHERE opponents.uuid = ? AND opponents.user_uuid = ? AND opponents.deleted_at IS NULL)",
			match.OpponentUUID,
			userUUID,
		).
		Updates(map[string]interface{}{
			"result":                  match.Result,
			"round_number":            match.RoundNumber,
			"phase":                   match.Phase,
			"opposing_deck_archetype": match.OpposingDeckArchetype,
			"tournament_uuid":         match.TournamentUUID,
			"opponent_uuid":           match.OpponentUUID,
		})
	return affectedRecord(result)
}

func (r *matchRepository) SoftDelete(ctx context.Context, id, userUUID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&models.Match{}).
		Where(
			"uuid = ? AND deleted_at IS NULL AND EXISTS (SELECT 1 FROM tournaments WHERE tournaments.uuid = matches.tournament_uuid AND tournaments.user_uuid = ? AND tournaments.deleted_at IS NULL)",
			id,
			userUUID,
		).
		Update("deleted_at", time.Now())
	return affectedRecord(result)
}

func (r *matchRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Match, error) {
	var match models.Match
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&match, "uuid = ?", id).Error; err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *matchRepository) ListByTournament(ctx context.Context, tournamentUUID uuid.UUID, resultFilter *models.ResultName) ([]*models.Match, error) {
	var matches []*models.Match
	query := activeMatchQuery(r.db.WithContext(ctx)).Where("matches.tournament_uuid = ?", tournamentUUID)

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
	query := activeMatchQuery(r.db.WithContext(ctx)).
		Where("t.deck_uuid = ?", deckUUID)

	if resultFilter != nil {
		query = query.Where("matches.result = ?", *resultFilter)
	}

	if err := query.Find(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}

func activeMatchQuery(query *gorm.DB) *gorm.DB {
	return query.
		Joins("JOIN tournaments AS t ON t.uuid = matches.tournament_uuid").
		Joins("JOIN opponents AS o ON o.uuid = matches.opponent_uuid AND o.user_uuid = t.user_uuid").
		Joins("LEFT JOIN decks AS d ON d.uuid = t.deck_uuid AND d.user_uuid = t.user_uuid").
		Where("matches.deleted_at IS NULL AND t.deleted_at IS NULL AND o.deleted_at IS NULL").
		Where("(t.deck_uuid IS NULL OR (d.uuid IS NOT NULL AND d.deleted_at IS NULL))")
}
