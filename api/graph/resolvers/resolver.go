package graph

import (
	"context"
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

type Resolver struct {
	users userGetter
}

func NewResolver(users userGetter) *Resolver {
	return &Resolver{users: users}
}

type Query struct{}

type Mutation struct{}
