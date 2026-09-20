package crypto

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomToken(t *testing.T) {
	t.Run("round trip: generates token successfully", func(t *testing.T) {
		token, err := GenerateRandomToken()
		assert.NoError(t, err)

		decodedToken, err := base64.RawURLEncoding.DecodeString(token)
		assert.NoError(t, err)

		assert.Len(t, decodedToken, byteLength)
	})
	t.Run("two generated tokens don't match", func(t *testing.T) {
		t1, err := GenerateRandomToken()
		assert.NoError(t, err)

		t2, err := GenerateRandomToken()
		assert.NoError(t, err)

		assert.NotEqual(t, t1, t2)
	})
}

func TestHashToken(t *testing.T) {
	t.Run("hashed token does not match original", func(t *testing.T) {
		token, err := GenerateRandomToken()
		assert.NoError(t, err)

		hashed := HashToken(token)
		assert.NotEqual(t, token, hashed)
	})
	t.Run("hashed token is the same value every time", func(t *testing.T) {
		token, err := GenerateRandomToken()
		assert.NoError(t, err)

		hash1 := HashToken(token)
		assert.NotEqual(t, token, hash1)
		hash2 := HashToken(token)
		assert.NotEqual(t, token, hash2)

		assert.Equal(t, hash2, hash1)
	})
	t.Run("two generated tokens have different hashes", func(t *testing.T) {
		token1, err := GenerateRandomToken()
		assert.NoError(t, err)

		token2, err := GenerateRandomToken()
		assert.NoError(t, err)

		hash1 := HashToken(token1)
		assert.NotEqual(t, token1, hash1)
		hash2 := HashToken(token2)
		assert.NotEqual(t, token2, hash2)

		assert.NotEqual(t, hash1, hash2)
	})
	t.Run("known input produces expected hash", func(t *testing.T) {
		known := "test123"
		funcHash := HashToken(known)

		alg := sha256.New()
		alg.Write([]byte(known))
		sha256Hash := alg.Sum(nil)
		sha256Str := base64.RawURLEncoding.EncodeToString(sha256Hash)

		assert.Equal(t, funcHash, sha256Str)
	})
}
