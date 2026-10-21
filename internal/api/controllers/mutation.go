package controllers

import (
	"errors"
	"net/http"

	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
)

func writeMutationError(w http.ResponseWriter, err error) {
	if errors.Is(err, interfaces.ErrRecordNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
