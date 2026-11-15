package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserIDFromContext(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		id := uuid.New()

		got, ok := UserIDFromContext(WithUserID(context.Background(), id))

		assert.True(t, ok)
		assert.Equal(t, id, got)
	})

	t.Run("anonymous", func(t *testing.T) {
		_, ok := UserIDFromContext(context.Background())

		assert.False(t, ok)
	})

	t.Run("nil UUID is not authenticated", func(t *testing.T) {
		_, ok := UserIDFromContext(WithUserID(context.Background(), uuid.Nil))

		assert.False(t, ok)
	})
}
