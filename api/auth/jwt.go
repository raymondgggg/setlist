package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateAccessToken constructs a new JWT access token for the provided userID.
func GenerateAccessToken(userID uuid.UUID, secret []byte, ttl time.Duration) (string, error) {
	c := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return t.SignedString(secret)
}

// ParseAccessToken takes the provided tokenString and parses the original Claims object out of it.
func ParseAccessToken(tokenString string, secret []byte) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("token was not signed using HS256 alg")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	c, ok := t.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("unable to extract claims")
	}
	return c, nil
}
