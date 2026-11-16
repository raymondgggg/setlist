package auth

import (
	"context"
	"fmt"
	"setlist/crypto"
	"setlist/db"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type userRepository interface {
	GetByEmail(ctx context.Context, email string) (*db.User, error)
	Create(ctx context.Context, u *db.User) error
}

type verificationRepository interface {
	Create(ctx context.Context, t *db.EmailVerificationToken) error
	Consume(ctx context.Context, tokenHash string, now time.Time) (uuid.UUID, error)
}

type verificationSender interface {
	SendUserVerification(ctx context.Context, to, token string) error
}

type sessionRepository interface {
	GetByTokenHash(ctx context.Context, hash string) (*db.Session, error)
	Create(ctx context.Context, s *db.Session) error
	Update(ctx context.Context, s *db.Session) error
}

type Service struct {
	user         userRepository
	session      sessionRepository
	verification verificationRepository
	sender       verificationSender
	logger       *zap.Logger
	jwtSecret    string
}

func NewService(
	ur *db.UserRepository,
	sr *db.SessionRepository,
	vr *db.EmailVerificationRepository,
	sender verificationSender,
	l *zap.Logger,
	secret string,
) *Service {
	return &Service{
		user:         ur,
		session:      sr,
		verification: vr,
		sender:       sender,
		logger:       l,
		jwtSecret:    secret,
	}
}

// Refresh tokens have a 30 day lifespan.
const RefreshTokenTTL = (time.Hour * 24) * 30
const RefreshCookieName = "setlist-refresh-token"

var errLoginFailed = fmt.Errorf("The username or password provided is incorrect.")

// ErrEmailNotVerified is only returned after the password has been checked,
// so it reveals nothing to someone who doesn't already know the credentials.
var ErrEmailNotVerified = fmt.Errorf("Please verify your email address before logging in.")

type LoginResult struct {
	User         *db.User
	AccessToken  string
	RefreshToken string
}

func (s *Service) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	res := &LoginResult{}
	u, err := s.user.GetByEmail(ctx, normalizeEmail(email))
	if err != nil || u == nil {
		s.logger.Warn("failed to fetch user by email", zap.Error(err))
		return nil, errLoginFailed
	}
	res.User = u

	if u.PasswordHash == nil {
		// TODO: implement OAuth only flow
		s.logger.Warn("login attempt against account with no password set", zap.String("email", u.Email))
		return nil, errLoginFailed
	}
	if err = crypto.ComparePassword(*u.PasswordHash, password); err != nil {
		s.logger.Warn("passwords don't match", zap.Error(err))
		return nil, errLoginFailed
	}

	if u.EmailVerifiedAt == nil {
		s.logger.Warn("login attempt against unverified account", zap.String("userID", u.ID.String()))
		return nil, ErrEmailNotVerified
	}

	refreshToken, err := crypto.GenerateRandomToken()
	if err != nil {
		s.logger.Error("failed to generate refresh token", zap.Error(err))
		return nil, errLoginFailed
	}
	res.RefreshToken = refreshToken

	session := &db.Session{
		ID:        uuid.New(),
		UserID:    u.ID,
		TokenHash: crypto.HashToken(refreshToken),
		ExpiresAt: time.Now().Add(RefreshTokenTTL),
	}

	if err = s.session.Create(ctx, session); err != nil {
		s.logger.Error("failed to create session", zap.Error(err))
		return nil, errLoginFailed
	}

	secretBytes := []byte(s.jwtSecret)
	// Access tokens have a 10 minute lifespan.
	accessTokenTTL := time.Minute * 10
	accessToken, err := GenerateAccessToken(u.ID, secretBytes, accessTokenTTL)
	if err != nil {
		s.logger.Error("failed to generate access token", zap.Error(err))
		return nil, errLoginFailed
	}
	res.AccessToken = accessToken

	return res, nil
}

func (s *Service) Logout(ctx context.Context, hashedToken string) (bool, error) {
	session, err := s.session.GetByTokenHash(ctx, hashedToken)
	if err != nil {
		s.logger.Error("failed to fetch session", zap.Error(err))
		return true, nil
	}

	now := time.Now()
	session.RevokedAt = &now

	if err := s.session.Update(ctx, session); err != nil {
		s.logger.Error("failed to revoke session", zap.Error(err))
		return false, fmt.Errorf("internal server error")
	}

	return true, nil
}
