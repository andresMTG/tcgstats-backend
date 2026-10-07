package postgres

import (
	"context"
	"embed"
	"errors"
	"time"

	"github.com/andresMTG/tcgstats-backend/pkg/env"

	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/golang-migrate/migrate/v4"
	pg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var fs embed.FS

type PostgresConfiguration struct {
	DSN             string        `env:"DSN,required"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS"    envDefault:"25"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS"    envDefault:"5"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`
	ConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"1m"`
}

func Module() fx.Option {
	return fx.Module("postgres", fx.Options(
		fx.Provide(
			NewDB,
			env.LoadEnvConfiguration[PostgresConfiguration],
		),
		fx.Invoke(RunMigrations),
	))
}

func NewDB(lc fx.Lifecycle, config *PostgresConfiguration) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(config.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

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
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}

	driver, err := pg.WithInstance(sqlDB, &pg.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
