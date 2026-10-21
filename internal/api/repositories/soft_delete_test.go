package repositories

import (
	"context"
	"strings"
	"testing"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestActiveMatchQueriesExcludeSoftDeletedRecords(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=localhost user=postgres dbname=tcgstats sslmode=disable",
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	require.NoError(t, err)

	matches := make([]models.Match, 0)
	result := activeMatchQuery(db).Find(&matches)
	require.NoError(t, result.Error)
	statement := strings.ToLower(result.Statement.SQL.String())
	for _, predicate := range []string{
		"matches.deleted_at is null",
		"t.deleted_at is null",
		"o.deleted_at is null",
		"d.deleted_at is null",
		"t.deck_uuid is null",
	} {
		assert.Contains(t, statement, predicate)
	}
}

func TestAnalyticsActiveMatchesAreScopedAndExcludeDeletedRelatedRecords(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=localhost user=postgres dbname=tcgstats sslmode=disable",
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	require.NoError(t, err)

	matches := make([]models.Match, 0)
	result := (&analyticsRepository{db: db}).
		activeMatches(context.Background(), uuid.New()).
		Find(&matches)
	require.NoError(t, result.Error)
	statement := strings.ToLower(result.Statement.SQL.String())
	for _, predicate := range []string{
		"m.deleted_at is null",
		"t.deleted_at is null",
		"o.deleted_at is null",
		"d.deleted_at is null",
		"t.user_uuid",
		"o.user_uuid",
		"d.user_uuid",
	} {
		assert.Contains(t, statement, predicate)
	}
}
