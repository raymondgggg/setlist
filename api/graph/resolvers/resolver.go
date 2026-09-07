package graph

import (
	"context"
	"setlist/auth"
	"setlist/db"

	"github.com/google/uuid"
)

//go:generate go tool gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type userGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*db.User, error)
}

type authService interface {
	Login(ctx context.Context, email, password string) (string, string, error)
}

type Resolver struct {
	users userGetter
	auth  authService
}

func NewResolver(users userGetter, as *auth.Service) *Resolver {
	return &Resolver{
		users: users,
		auth:  as,
	}
}

type Query struct{}

type Mutation struct{}
