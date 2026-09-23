package controllers


import (
	"net/http"
)

// Health represents the health controller.
type Health struct{}

// NewHealth creates a new health controller.
func NewHealth() *Health {
	return &Health{}
}

// Health returns 200 OK status to indicate that the service is healthy.
func (h *Health) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}