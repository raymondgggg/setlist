package auth

import (
	"context"
	"setlist/crypto"
	"setlist/db"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type fakeUserGetter struct {
	user *db.User
	err  error
}

func (f *fakeUserGetter) GetByEmail(ctx context.Context, email string) (*db.User, error) {
	return f.user, f.err
}

type fakeSessionCreator struct {
	session *db.Session
	err     error
}

func (f *fakeSessionCreator) Create(ctx context.Context, s *db.Session) error {
	f.session = s
	return f.err
}

func setupLoginTest(email, password string) (*fakeUserGetter, *fakeSessionCreator, *Service) {
	userID := uuid.New()
	hash, _ := crypto.HashPassword(password)

	ug := &fakeUserGetter{
		user: &db.User{
			ID:           userID,
			Email:        email,
			PasswordHash: &hash,
		},
		err: nil,
	}

	sg := &fakeSessionCreator{session: &db.Session{}, err: nil}

	s := &Service{
		user:      ug,
		session:   sg,
		logger:    zap.NewNop(),
		jwtSecret: "test-secret",
	}

	return ug, sg, s
}

func TestLogin(t *testing.T) {
	t.Run("round trip: login is successful", func(t *testing.T) {
		email := "testuser@gmail.com"
		password := "test123"
		ug, _, s := setupLoginTest(email, password)

		res, err := s.Login(t.Context(), email, password)
		assert.NoError(t, err)

		// verify access token is usable
		claims, err := ParseAccessToken(res.AccessToken, []byte(s.jwtSecret))
		assert.NoError(t, err)
		assert.Equal(t, ug.user.ID, claims.UserID)
	})
	t.Run("session is created with correct data", func(t *testing.T) {
		email := "johndoe@gmail.com"
		password := "t0t4llys3curep4ssw0rd"
		ug, sg, s := setupLoginTest(email, password)

		res, err := s.Login(t.Context(), email, password)
		assert.NoError(t, err)

		// 1. check the user ID on the session matches the one login created
		assert.Equal(t, ug.user.ID, sg.session.UserID)

		// 2. session's hashed refresh token matches the one returned in the login result hashed
		assert.Equal(t, sg.session.TokenHash, crypto.HashToken(res.RefreshToken))

		// 3. token expires in about now + RefreshTokenTTL (1s delta)
		assert.WithinDuration(t, time.Now().Add(RefreshTokenTTL), sg.session.ExpiresAt, 1*time.Second)
	})
}
