package postgres

import (
	"context"
	"os"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"

	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Module() fx.Option {
	return fx.Module("postgres", fx.Options(
		fx.Provide(NewDB),
		fx.Invoke(RunMigrations),
	))
}

func NewDB(lc fx.Lifecycle) (*gorm.DB, error) {
	dsn := os.Getenv("DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})

	return db, nil
}

func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(&models.Users{})
}
