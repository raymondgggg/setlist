package auth

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const secretLength = 32

func TestGenerateParseAccessToken(t *testing.T) {
	t.Run("round trip: access token generated successfully for user", func(t *testing.T) {
		userID := uuid.New()
		ttl := time.Minute * 10
		secret := make([]byte, secretLength)
		rand.Read(secret)

		token, err := GenerateAccessToken(userID, secret, ttl)
		assert.NoError(t, err)

		claims, err := ParseAccessToken(token, secret)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
	})
	t.Run("generating with one secret, parsing with another", func(t *testing.T) {
		userID := uuid.New()
		ttl := time.Minute * 10

		secret1 := make([]byte, secretLength)
		rand.Read(secret1)
		secret2 := make([]byte, secretLength)
		rand.Read(secret2)

		token, err := GenerateAccessToken(userID, secret1, ttl)
		assert.NoError(t, err)

		_, err = ParseAccessToken(token, secret2)
		assert.Error(t, err)
	})
	t.Run("expired token isn't able to be parsed", func(t *testing.T) {
		userID := uuid.New()
		ttl := time.Second * 1
		secret := make([]byte, secretLength)
		rand.Read(secret)

		token, err := GenerateAccessToken(userID, secret, ttl)
		assert.NoError(t, err)

		time.Sleep(ttl)

		_, err = ParseAccessToken(token, secret)
		assert.Error(t, err)
	})
}
