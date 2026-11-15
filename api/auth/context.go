package auth

import (
	"context"

	"github.com/google/uuid"
)

type userIDKey struct{}

// WithUserID returns a copy of ctx carrying the authenticated user's ID.
func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}

// UserIDFromContext returns the authenticated user's ID, and false when the
// request is anonymous.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey{}).(uuid.UUID)
	return id, ok && id != uuid.Nil
}
