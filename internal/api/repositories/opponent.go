package repositories

import (
	"context"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type opponentRepository struct {
	db *gorm.DB
}

func NewOpponentRepository(db *gorm.DB) interfaces.OpponentRepository {
	return &opponentRepository{db: db}
}

func (r *opponentRepository) Create(ctx context.Context, opponent *models.Opponent) error {
	return r.db.WithContext(ctx).Create(opponent).Error
}

func (r *opponentRepository) Update(ctx context.Context, opponent *models.Opponent) error {
	result := r.db.WithContext(ctx).Model(&models.Opponent{}).
		Where("uuid = ? AND user_uuid = ? AND deleted_at IS NULL", opponent.UUID, opponent.UserUUID).
		Update("name", opponent.Name)
	return affectedRecord(result)
}

func (r *opponentRepository) SoftDelete(ctx context.Context, id, userUUID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&models.Opponent{}).
		Where("uuid = ? AND user_uuid = ? AND deleted_at IS NULL", id, userUUID).
		Update("deleted_at", time.Now())
	return affectedRecord(result)
}

func (r *opponentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Opponent, error) {
	var opponent models.Opponent
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&opponent, "uuid = ?", id).Error; err != nil {
		return nil, err
	}
	return &opponent, nil
}

func (r *opponentRepository) ListByUser(ctx context.Context, userUUID uuid.UUID) ([]*models.Opponent, error) {
	var opponents []*models.Opponent
	if err := r.db.WithContext(ctx).Where("user_uuid = ? AND deleted_at IS NULL", userUUID).Find(&opponents).Error; err != nil {
		return nil, err
	}
	return opponents, nil
}
