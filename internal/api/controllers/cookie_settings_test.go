package controllers_test

import (
	"testing"

	"github.com/andresMTG/tcgstats-backend/internal/api/controllers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCookieSettings_NormalizesBasePath(t *testing.T) {
	t.Setenv("AUTH_COOKIE_SECURE", "true")
	t.Setenv("AUTH_COOKIE_SAME_SITE", "lax")
	t.Setenv("SERVER_BASE_PATH", "/api/")

	settings, err := controllers.NewCookieSettings()
	require.NoError(t, err)
	assert.Equal(t, "/api", settings.BasePath)
	assert.True(t, settings.Secure)
	assert.Equal(t, "lax", settings.SameSite)
}

func TestNewCookieSettings_RejectsSameSiteNoneWithoutSecure(t *testing.T) {
	t.Setenv("AUTH_COOKIE_SECURE", "false")
	t.Setenv("AUTH_COOKIE_SAME_SITE", "none")

	_, err := controllers.NewCookieSettings()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires AUTH_COOKIE_SECURE=true")
}
