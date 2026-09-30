package dto

import "time"

type UserRequest struct {
	Name     string
	Email    string
	Password string
}

type UserLoginRequest struct {
	Email    string
	Password string
}

type UserLoginResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type UserSimpleResponse struct {
	Name     string
	Email    string
}
