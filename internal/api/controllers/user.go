package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/andresMTG/tcgstats-backend/internal/api/auth"
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/dto"
	"github.com/andresMTG/tcgstats-backend/internal/api/domain/models"
	"github.com/andresMTG/tcgstats-backend/internal/api/interfaces"
	"github.com/andresMTG/tcgstats-backend/internal/api/middleware"
	"github.com/andresMTG/tcgstats-backend/pkg/env"
	"github.com/google/uuid"
)

type UserController struct {
	service interfaces.UserService
	cookies *CookieSettings
}

type CookieSettings struct {
	Secure   bool   `env:"AUTH_COOKIE_SECURE" envDefault:"true"`
	SameSite string `env:"AUTH_COOKIE_SAME_SITE" envDefault:"lax"`
	BasePath string `env:"SERVER_BASE_PATH" envDefault:"/api/"`
}

func NewCookieSettings() (*CookieSettings, error) {
	settings, err := env.LoadEnvConfiguration[CookieSettings]()
	if err != nil {
		return nil, err
	}
	settings.BasePath = path.Clean("/" + settings.BasePath)
	switch strings.ToLower(settings.SameSite) {
	case "lax":
		settings.SameSite = "lax"
	case "strict":
		settings.SameSite = "strict"
	case "none":
		if !settings.Secure {
			return nil, errors.New("AUTH_COOKIE_SAME_SITE=none requires AUTH_COOKIE_SECURE=true")
		}
		settings.SameSite = "none"
	default:
		return nil, errors.New("AUTH_COOKIE_SAME_SITE must be lax, strict, or none")
	}
	return settings, nil
}

func NewUserController(service interfaces.UserService, cookies *CookieSettings) *UserController {
	return &UserController{service: service, cookies: cookies}
}

func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := c.service.CreateUser(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (c *UserController) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokens, err := c.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, interfaces.ErrInvalidCredentials) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(w, "unable to create session", http.StatusInternalServerError)
		return
	}

	c.setAuthCookies(w, tokens)
	w.WriteHeader(http.StatusNoContent)
}

func (c *UserController) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshCookie, err := r.Cookie(auth.RefreshCookieName)
	if err != nil || refreshCookie.Value == "" {
		c.clearAuthCookies(w)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tokens, err := c.service.Refresh(r.Context(), refreshCookie.Value)
	if err != nil {
		c.clearAuthCookies(w)
		if errors.Is(err, interfaces.ErrInvalidRefreshToken) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "unable to refresh session", http.StatusInternalServerError)
		return
	}

	c.setAuthCookies(w, tokens)
	w.WriteHeader(http.StatusNoContent)
}

func (c *UserController) Logout(w http.ResponseWriter, r *http.Request) {
	refreshCookie, err := r.Cookie(auth.RefreshCookieName)
	c.clearAuthCookies(w)
	if err == nil && refreshCookie.Value != "" {
		if err := c.service.Logout(r.Context(), refreshCookie.Value); err != nil &&
			!errors.Is(err, interfaces.ErrInvalidRefreshToken) {
			log.Printf("logout: failed to revoke refresh-token session: %v", err)
			http.Error(w, "unable to revoke session", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *UserController) setAuthCookies(w http.ResponseWriter, tokens *models.AuthTokens) {
	w.Header().Set("Cache-Control", "no-store")
	http.SetCookie(w, c.authCookie(auth.AccessCookieName, tokens.AccessToken,
		int(tokens.AccessTokenTTL.Seconds()), c.cookies.BasePath))
	http.SetCookie(w, c.authCookie(auth.RefreshCookieName, tokens.RefreshToken,
		int(tokens.RefreshTokenTTL.Seconds()), path.Join(c.cookies.BasePath, "users")))
}

func (c *UserController) clearAuthCookies(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	http.SetCookie(w, c.authCookie(auth.AccessCookieName, "", -1, c.cookies.BasePath))
	http.SetCookie(w, c.authCookie(auth.RefreshCookieName, "", -1, path.Join(c.cookies.BasePath, "users")))
}

func (c *UserController) authCookie(name, value string, maxAge int, cookiePath string) *http.Cookie {
	sameSite := http.SameSiteLaxMode
	switch c.cookies.SameSite {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	}
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     cookiePath,
		HttpOnly: true,
		Secure:   c.cookies.Secure,
		SameSite: sameSite,
		MaxAge:   maxAge,
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
	}
}

func (c *UserController) GetMe(w http.ResponseWriter, r *http.Request) {
	userIdStr, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(userIdStr)
	if err != nil {
		http.Error(w, "invalid user id in token", http.StatusUnauthorized)
		return
	}

	user, err := c.service.GetUserByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}
