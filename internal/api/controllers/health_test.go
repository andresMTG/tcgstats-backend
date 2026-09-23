package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andresMTG/tcgstats-backend/internal/api/controllers"
	"github.com/andresMTG/tcgstats-backend/tests/utils"
	"github.com/stretchr/testify/assert"
)

func TestHealthController_Health(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		wantStatusCode int
	}{
		{
			name:           "Works correctly",
			wantStatusCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := utils.BuildJSONRequest(t, nil, "/health", http.MethodGet)

			c := controllers.NewHealth()

			c.Health(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
		})
	}
}