package api_test

import (
	"testing"

	"github.com/andresMTG/tcgstats-backend/internal/api"
	"github.com/andresMTG/tcgstats-backend/pkg/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfiguration_LoadsAllowedCORSOrigins(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com,http://localhost:3000")
	config, err := env.LoadEnvConfiguration[api.Configuration]()
	require.NoError(t, err)
	assert.Equal(t, []string{"https://app.example.com", "http://localhost:3000"}, config.CORSAllowedOrigins)
}
