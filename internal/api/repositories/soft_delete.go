package repositories

import (
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"gorm.io/gorm"
)

func affectedRecord(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return interfaces.ErrRecordNotFound
	}
	return nil
}
