package auth

import (
	"context"
	"errors"
	"setlist/crypto"
	"setlist/db"
	"setlist/email"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeVerificationRepository struct {
	created    *db.EmailVerificationToken
	createErr  error
	consumeErr error
	consumedBy uuid.UUID
	gotHash    string
}

func (f *fakeVerificationRepository) Create(ctx context.Context, t *db.EmailVerificationToken) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = t
	return nil
}

func (f *fakeVerificationRepository) Consume(ctx context.Context, hash string, now time.Time) (uuid.UUID, error) {
	f.gotHash = hash
	return f.consumedBy, f.consumeErr
}

type fakeSender struct {
	err     error
	to      string
	token   string
	numSent int
}

func (f *fakeSender) SendUserVerification(ctx context.Context, to, token string) error {
	f.numSent++
	f.to, f.token = to, token
	return f.err
}

func setupRegisterTest() (*fakeUserGetter, *fakeVerificationRepository, *fakeSender, *Service) {
	ug := &fakeUserGetter{}
	vr := &fakeVerificationRepository{}
	sender := &fakeSender{}
	s := &Service{
		user:         ug,
		verification: vr,
		sender:       sender,
		logger:       zap.NewNop(),
		jwtSecret:    "test-secret",
	}
	return ug, vr, sender, s
}

func validInput() RegisterInput {
	return RegisterInput{Email: "  JohnDoe@Gmail.com ", Password: "t0t4llys3curep4ssw0rd"}
}

func TestRegister(t *testing.T) {
	t.Run("creates an unverified user and emails a verification token", func(t *testing.T) {
		ug, vr, sender, s := setupRegisterTest()

		u, err := s.Register(t.Context(), validInput())
		require.NoError(t, err)

		// user: lowercased email, hashed password, not verified
		assert.Same(t, ug.created, u)
		assert.Equal(t, "johndoe@gmail.com", u.Email)
		assert.NotEqual(t, uuid.Nil, u.ID)
		assert.Nil(t, u.EmailVerifiedAt)
		require.NotNil(t, u.PasswordHash)
		assert.NotEqual(t, validInput().Password, *u.PasswordHash)
		assert.NoError(t, crypto.ComparePassword(*u.PasswordHash, validInput().Password))

		// token: stored hashed, for this user, expiring on schedule; raw token is what's emailed
		require.NotNil(t, vr.created)
		assert.Equal(t, u.ID, vr.created.UserID)
		assert.Equal(t, crypto.HashToken(sender.token), vr.created.TokenHash)
		assert.NotEqual(t, sender.token, vr.created.TokenHash)
		assert.WithinDuration(t, time.Now().Add(email.VerificationLinkTTL), vr.created.ExpiresAt, time.Second)
		assert.Equal(t, "johndoe@gmail.com", sender.to)
	})
	t.Run("duplicate email", func(t *testing.T) {
		ug, _, sender, s := setupRegisterTest()
		ug.createErr = db.ErrEmailTaken

		u, err := s.Register(t.Context(), validInput())
		assert.Nil(t, u)
		assert.Equal(t, ErrEmailTaken, err)
		assert.Zero(t, sender.numSent)
	})
	t.Run("unexpected create failure is not leaked", func(t *testing.T) {
		ug, _, _, s := setupRegisterTest()
		ug.createErr = errors.New("connection refused")

		u, err := s.Register(t.Context(), validInput())
		assert.Nil(t, u)
		assert.Equal(t, errInternal, err)
	})
	t.Run("send failure still registers the user", func(t *testing.T) {
		ug, _, sender, s := setupRegisterTest()
		sender.err = errors.New("resend down")

		u, err := s.Register(t.Context(), validInput())
		assert.NoError(t, err)
		assert.Same(t, ug.created, u)
	})
	t.Run("token storage failure still registers the user and sends nothing", func(t *testing.T) {
		ug, vr, sender, s := setupRegisterTest()
		vr.createErr = errors.New("db down")

		u, err := s.Register(t.Context(), validInput())
		assert.NoError(t, err)
		assert.Same(t, ug.created, u)
		assert.Zero(t, sender.numSent)
	})
	t.Run("validation", func(t *testing.T) {
		tests := []struct {
			name   string
			mutate func(*RegisterInput)
			err    error
		}{
			{"empty email", func(in *RegisterInput) { in.Email = "" }, ErrInvalidEmail},
			{"not an email", func(in *RegisterInput) { in.Email = "nope" }, ErrInvalidEmail},
			{"display-name form", func(in *RegisterInput) { in.Email = "John <john@gmail.com>" }, ErrInvalidEmail},
			{"short password", func(in *RegisterInput) { in.Password = "short" }, ErrPasswordTooShort},
			{"password over bcrypt limit", func(in *RegisterInput) {
				in.Password = string(make([]byte, maxPasswordBytes+1))
			}, ErrPasswordTooLong},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ug, _, _, s := setupRegisterTest()
				in := validInput()
				tt.mutate(&in)

				u, err := s.Register(t.Context(), in)
				assert.Nil(t, u)
				assert.Equal(t, tt.err, err)
				assert.Nil(t, ug.created)
			})
		}
	})
}

func TestVerifyEmail(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		_, vr, _, s := setupRegisterTest()

		err := s.VerifyEmail(t.Context(), "raw-token")
		assert.NoError(t, err)
		assert.Equal(t, crypto.HashToken("raw-token"), vr.gotHash)
	})
	t.Run("invalid, expired or used token", func(t *testing.T) {
		_, vr, _, s := setupRegisterTest()
		vr.consumeErr = db.ErrVerificationTokenInvalid

		assert.Equal(t, ErrInvalidVerification, s.VerifyEmail(t.Context(), "raw-token"))
	})
	t.Run("unexpected failure is not leaked", func(t *testing.T) {
		_, vr, _, s := setupRegisterTest()
		vr.consumeErr = errors.New("db down")

		assert.Equal(t, errInternal, s.VerifyEmail(t.Context(), "raw-token"))
	})
}
