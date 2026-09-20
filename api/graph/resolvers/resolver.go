package graph

import (
	"context"
	"net/http"
	"setlist/auth"
	"setlist/config"
	"setlist/db"
	"time"

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
	Logout(ctx context.Context, refreshToken string) (bool, error)
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

func (r *Resolver) setRefreshCookie(value string, expires time.Time, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     auth.RefreshCookieName,
		Value:    value,
		HttpOnly: true,
		Secure:   r.env.IsProduction(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/graphql",
		Expires:  expires,
		MaxAge:   maxAge,
	}
}
