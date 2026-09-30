package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/andresMTG/tcgstats-backend/internal/api/domain/dto"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/andresMTG/tcgstats-backend/internal/api/middleware"
	"github.com/andresMTG/tcgstats-backend/internal/api/services"
)

// Users represents the users controller.
type Users struct {
	services interfaces.UsersService
}

// NewUsers creates a new users controller.
func NewUsers(s *services.UsersService) *Users {
	return &Users{services: s}
}

// Create returns 201 Created status to indicate that the a user is created.
// Create returns 400 Bad Request if the data to create is wrong.
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.UserRequest

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := u.services.Create(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	loginResponse, err := u.services.Login(&req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			http.Error(w, "invalid email or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(loginResponse)
}

func (u *Users) ShowMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	simpleUserResponse, err := u.services.ShowMe(userID)
	if err != nil {
		http.Error(w, "User error", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("Olá, o seu nome é: " + simpleUserResponse.Name + " e o seu email é: " + simpleUserResponse.Email))
}
