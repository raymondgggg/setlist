package auth

import (
	"context"
	"fmt"
	"setlist/crypto"
	"setlist/db"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type fakeUserGetter struct {
	user      *db.User
	err       error
	createErr error
	created   *db.User
	gotEmail  string
}

func (f *fakeUserGetter) GetByEmail(ctx context.Context, email string) (*db.User, error) {
	f.gotEmail = email
	return f.user, f.err
}

func (f *fakeUserGetter) Create(ctx context.Context, u *db.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = u
	return nil
}

type fakeSessionRepository struct {
	session           *db.Session
	getByTokenHashErr error
	createErr         error
	updateErr         error
}

func (f *fakeSessionRepository) GetByTokenHash(ctx context.Context, hash string) (*db.Session, error) {
	return f.session, f.getByTokenHashErr
}

func (f *fakeSessionRepository) Create(ctx context.Context, cs *db.Session) error {
	f.session = cs
	return f.createErr
}

func (f *fakeSessionRepository) Update(ctx context.Context, us *db.Session) error {
	f.session = us
	return f.updateErr
}

func setupLoginTest(email, password string, err error) (*fakeUserGetter, *fakeSessionRepository, *Service) {
	userID := uuid.New()
	hash, _ := crypto.HashPassword(password)
	verifiedAt := time.Now()

	ug := &fakeUserGetter{
		user: &db.User{
			ID:              userID,
			Email:           email,
			PasswordHash:    &hash,
			EmailVerifiedAt: &verifiedAt,
		},
		err: nil,
	}

	sg := &fakeSessionRepository{session: &db.Session{}, createErr: err}

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
		ug, _, s := setupLoginTest(email, password, nil)

		res, err := s.Login(t.Context(), email, password)
		assert.NoError(t, err)

		// verify access token is usable
		claims, err := ParseAccessToken(res.AccessToken, []byte(s.jwtSecret))
		assert.NoError(t, err)
		assert.Equal(t, ug.user.ID, claims.UserID)
		assert.Equal(t, ug.user, res.User)
	})
	t.Run("session is created with correct data", func(t *testing.T) {
		email := "johndoe@gmail.com"
		password := "t0t4llys3curep4ssw0rd"
		ug, sg, s := setupLoginTest(email, password, nil)

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

func TestLoginEmailVerification(t *testing.T) {
	t.Run("unverified account with correct password is told to verify", func(t *testing.T) {
		email, password := "unverified@gmail.com", "t0t4llys3curep4ssw0rd"
		ug, sg, s := setupLoginTest(email, password, nil)
		ug.user.EmailVerifiedAt = nil

		res, err := s.Login(t.Context(), email, password)
		assert.Nil(t, res)
		assert.Equal(t, ErrEmailNotVerified, err)
		assert.Empty(t, sg.session.TokenHash, "no session should be created")
	})
	t.Run("unverified account with wrong password stays generic", func(t *testing.T) {
		email, password := "unverified@gmail.com", "t0t4llys3curep4ssw0rd"
		ug, _, s := setupLoginTest(email, password, nil)
		ug.user.EmailVerifiedAt = nil

		res, err := s.Login(t.Context(), email, "wrong-password")
		assert.Nil(t, res)
		assert.Equal(t, errLoginFailed, err)
	})
	t.Run("email lookup is case-insensitive", func(t *testing.T) {
		password := "t0t4llys3curep4ssw0rd"
		ug, _, s := setupLoginTest("johndoe@gmail.com", password, nil)

		_, err := s.Login(t.Context(), "  JohnDoe@Gmail.com ", password)
		assert.NoError(t, err)
		assert.Equal(t, "johndoe@gmail.com", ug.gotEmail)
	})
}

func setupLogoutTest(ctx context.Context, hashErr, updateErr error) (*fakeSessionRepository, *Service) {
	sg := &fakeSessionRepository{
		session:           &db.Session{},
		getByTokenHashErr: hashErr,
		updateErr:         updateErr,
	}

	s := &Service{
		session:   sg,
		logger:    zap.NewNop(),
		jwtSecret: "test-secret",
	}

	return sg, s
}

func TestLogout(t *testing.T) {
	t.Run("round trip: logout is successful", func(t *testing.T) {
		ctx := t.Context()
		sg, s := setupLogoutTest(ctx, nil, nil)

		success, err := s.Logout(ctx, sg.session.TokenHash)
		assert.NoError(t, err)
		assert.Equal(t, true, success)
	})
	t.Run("session not found but logout quietly succeeds as expected", func(t *testing.T) {
		ctx := t.Context()
		hashErr := fmt.Errorf("no rows returned")
		_, s := setupLogoutTest(ctx, hashErr, nil)

		hashedToken := crypto.HashToken("non-existing-hash")

		success, err := s.Logout(ctx, hashedToken)
		assert.NoError(t, err)
		assert.Equal(t, true, success)
	})
	t.Run("updating revoked at fails", func(t *testing.T) {
		ctx := t.Context()
		updateErr := fmt.Errorf("internal server error")
		sg, s := setupLogoutTest(ctx, nil, updateErr)

		success, err := s.Logout(ctx, sg.session.TokenHash)
		assert.Equal(t, updateErr, err)
		assert.Equal(t, false, success)
	})
}
