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

type userGetter interface {
	GetByEmail(ctx context.Context, email string) (*db.User, error)
}

type sessionCreator interface {
	Create(ctx context.Context, s *db.Session) error
}

type Service struct {
	user      userGetter
	session   sessionCreator
	logger    *zap.Logger
	jwtSecret string
}

func NewService(ur *db.UserRepository, sr *db.SessionRepository, l *zap.Logger, secret string) *Service {
	return &Service{
		user:      ur,
		session:   sr,
		logger:    l,
		jwtSecret: secret,
	}
}

// Refresh tokens have a 30 day lifespan.
const RefreshTokenTTL = (time.Hour * 24) * 30

var errLoginFailed = fmt.Errorf("The username or password provided is incorrect.")

type LoginResult struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
}

func (s *Service) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	res := &LoginResult{}
	u, err := s.user.GetByEmail(ctx, email)
	if err != nil || u == nil {
		s.logger.Warn("failed to fetch user", zap.Error(err))
		return nil, errLoginFailed
	}
	res.UserID = u.ID

	if u.PasswordHash == nil {
		// TODO: implement OAuth only flow
		s.logger.Warn("login attempt against account with no password set", zap.String("email", u.Email))
		return nil, errLoginFailed
	}
	if err = crypto.ComparePassword(*u.PasswordHash, password); err != nil {
		s.logger.Warn("passwords don't match", zap.Error(err))
		return nil, errLoginFailed
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
		TokenHash: refreshToken,
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
