package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenType   = "access"
	RefreshTokenType  = "refresh"
	AccessCookieName  = "access_token"
	RefreshCookieName = "refresh_token"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	jwt.RegisteredClaims
	Type string `json:"type,omitempty"`
}

func ParseToken(tokenString, secret string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !token.Valid || claims.Subject == "" || claims.ExpiresAt == nil {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
