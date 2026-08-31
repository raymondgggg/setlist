package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashAndComparePassword(t *testing.T) {
	t.Run("round trip success: password hashed and hash matches", func(t *testing.T) {
		plain := "plain"

		hashed, err := HashPassword(plain)
		assert.NoError(t, err)

		err = ComparePassword(hashed, plain)
		assert.NoError(t, err)
	})
	t.Run("wrong password fails", func(t *testing.T) {
		correctPlain := "correctpass"
		wrongPlain := "wrongpass"

		correctHashed, err := HashPassword(correctPlain)
		assert.NoError(t, err)

		err = ComparePassword(correctHashed, wrongPlain)
		assert.Error(t, err)
	})
	t.Run("hash string doesn't equal original password", func(t *testing.T) {
		originalPlain := "originalpass"

		hashed, err := HashPassword(originalPlain)
		assert.NoError(t, err)
		assert.NotEqual(t, originalPlain, hashed)
	})
	t.Run("two hashes of the same password differ", func(t *testing.T) {
		plain := "plain"

		hashedOne, err := HashPassword(plain)
		assert.NoError(t, err)

		hashedTwo, err := HashPassword(plain)
		assert.NoError(t, err)

		assert.NotEqual(t, hashedOne, hashedTwo)
	})
	t.Run("empty password", func(t *testing.T) {
		empty := ""

		emptyHash, err := HashPassword(empty)
		assert.NoError(t, err)
		assert.NotEqual(t, empty, emptyHash)
	})
}
