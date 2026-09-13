package graph

import (
	"context"
	"setlist/auth"
	"setlist/config"
	"setlist/db"

	"github.com/google/uuid"
	"go.uber.org/zap"
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
	Login(ctx context.Context, email, password string) (*auth.LoginResult, error)
}

type Resolver struct {
	users  userGetter
	auth   authService
	logger *zap.Logger
	env    config.Environment
}

func NewResolver(users userGetter, as *auth.Service, l *zap.Logger, e config.Environment) *Resolver {
	return &Resolver{
		users:  users,
		auth:   as,
		logger: l,
		env:    e,
	}
}

type Query struct{}

type Mutation struct{}
