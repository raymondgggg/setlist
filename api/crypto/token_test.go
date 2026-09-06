package crypto

import (
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
