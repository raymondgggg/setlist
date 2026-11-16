package graph

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"setlist/auth"
	"setlist/db"
	"setlist/graph"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type fakeUsers struct {
	user  *db.User
	err   error
	gotID uuid.UUID
}

func (f *fakeUsers) GetByID(_ context.Context, id uuid.UUID) (*db.User, error) {
	f.gotID = id
	return f.user, f.err
}

type fakeAuth struct {
	login    *auth.LoginResult
	loginErr error
	register *db.User
	regErr   error
}

func (f *fakeAuth) Login(context.Context, string, string) (*auth.LoginResult, error) {
	return f.login, f.loginErr
}
func (f *fakeAuth) Logout(context.Context, string) (bool, error) { return true, nil }
func (f *fakeAuth) Register(context.Context, auth.RegisterInput) (*db.User, error) {
	return f.register, f.regErr
}
func (f *fakeAuth) VerifyEmail(context.Context, string) error { return nil }

func newTestResolver(u *fakeUsers, a *fakeAuth) *Resolver {
	return &Resolver{users: u, auth: a, logger: zap.NewNop()}
}

func TestSessionQuery(t *testing.T) {
	id := uuid.New()
	du := &db.User{ID: id, Email: "a@example.com"}

	t.Run("returns the caller's user", func(t *testing.T) {
		users := &fakeUsers{user: du}
		r := newTestResolver(users, &fakeAuth{})

		s, err := r.Query().Session(auth.WithUserID(t.Context(), id))
		require.NoError(t, err)
		assert.Equal(t, id, users.gotID)
		assert.Equal(t, graph.ToGraphQLUser(du), s.User)
	})

	t.Run("anonymous caller is unauthenticated", func(t *testing.T) {
		users := &fakeUsers{user: du}
		r := newTestResolver(users, &fakeAuth{})

		s, err := r.Query().Session(t.Context())
		assert.Nil(t, s)
		assert.EqualError(t, err, graph.UnauthenticatedError().Error())
		assert.Equal(t, uuid.Nil, users.gotID)
	})

	t.Run("deleted user is unauthenticated", func(t *testing.T) {
		r := newTestResolver(&fakeUsers{err: gorm.ErrRecordNotFound}, &fakeAuth{})

		s, err := r.Query().Session(auth.WithUserID(t.Context(), id))
		assert.Nil(t, s)
		assert.EqualError(t, err, graph.UnauthenticatedError().Error())
	})

	t.Run("db error is not leaked", func(t *testing.T) {
		r := newTestResolver(&fakeUsers{err: errors.New("pq: connection refused")}, &fakeAuth{})

		s, err := r.Query().Session(auth.WithUserID(t.Context(), id))
		assert.Nil(t, s)
		assert.EqualError(t, err, "internal server error")
	})
}

func TestLogin(t *testing.T) {
	du := &db.User{ID: uuid.New(), Email: "a@example.com"}
	res := &auth.LoginResult{User: du, AccessToken: "access", RefreshToken: "refresh"}

	t.Run("returns access token and session user, sets cookie", func(t *testing.T) {
		r := newTestResolver(&fakeUsers{}, &fakeAuth{login: res})
		w := httptest.NewRecorder()
		ctx := graph.WithResponseWriter(t.Context(), w)

		got, err := r.Mutation().Login(ctx, "a@example.com", "pw")
		require.NoError(t, err)
		assert.Equal(t, "access", got.AccessToken)
		assert.Equal(t, graph.ToGraphQLUser(du), got.Session.User)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, auth.RefreshCookieName, cookies[0].Name)
		assert.Equal(t, "refresh", cookies[0].Value)
	})

	t.Run("auth failure is returned without a cookie", func(t *testing.T) {
		loginErr := errors.New("invalid email or password")
		r := newTestResolver(&fakeUsers{}, &fakeAuth{loginErr: loginErr})
		w := httptest.NewRecorder()
		ctx := graph.WithResponseWriter(t.Context(), w)

		got, err := r.Mutation().Login(ctx, "a@example.com", "pw")
		assert.Nil(t, got)
		assert.Equal(t, loginErr, err)
		assert.Empty(t, w.Result().Cookies())
	})
}

func TestCreateUser(t *testing.T) {
	t.Run("returns only the new user's ID", func(t *testing.T) {
		id := uuid.New()
		r := newTestResolver(&fakeUsers{}, &fakeAuth{register: &db.User{ID: id}})

		got, err := r.Mutation().CreateUser(t.Context(), graph.NewUser{})
		require.NoError(t, err)
		assert.Equal(t, id, got)
	})

	t.Run("registration error returns a nil ID", func(t *testing.T) {
		regErr := errors.New("This email is already in use.")
		r := newTestResolver(&fakeUsers{}, &fakeAuth{regErr: regErr})

		got, err := r.Mutation().CreateUser(t.Context(), graph.NewUser{})
		assert.Equal(t, uuid.Nil, got)
		assert.Equal(t, regErr, err)
	})
}
