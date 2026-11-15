package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"setlist/crypto"
	"setlist/db"
	"setlist/email"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	minPasswordLength = 8
	// bcrypt only uses the first 72 bytes of its input (and rejects longer ones).
	maxPasswordBytes = 72
)

var (
	ErrInvalidEmail        = errors.New("Please enter a valid email address.")
	ErrPasswordTooShort    = fmt.Errorf("Password must be at least %d characters.", minPasswordLength)
	ErrPasswordTooLong     = fmt.Errorf("Password must be at most %d bytes.", maxPasswordBytes)
	ErrEmailTaken          = errors.New("This email is already in use.")
	ErrInvalidVerification = errors.New("This verification link is invalid or has expired.")
	errInternal            = errors.New("internal server error")
)

type RegisterInput struct {
	FirstName *string
	LastName  *string
	Email     string
	Password  string
}

// normalizeEmail makes emails comparable: the database stores them lowercased.
func normalizeEmail(e string) string {
	return strings.ToLower(strings.TrimSpace(e))
}

func validateEmail(e string) error {
	addr, err := mail.ParseAddress(e)
	// ParseAddress also accepts "Name <a@b.com>"; we only want a bare address.
	if err != nil || addr.Address != e {
		return ErrInvalidEmail
	}
	return nil
}

func validatePassword(p string) error {
	if len(p) < minPasswordLength {
		return ErrPasswordTooShort
	}
	if len(p) > maxPasswordBytes {
		return ErrPasswordTooLong
	}
	return nil
}

// Register creates an unverified user and sends them a verification email.
// A failure to issue or send the verification email does not fail
// registration: the user exists and can have the email re-sent later.
func (s *Service) Register(ctx context.Context, in RegisterInput) (*db.User, error) {
	emailAddr := normalizeEmail(in.Email)
	if err := validateEmail(emailAddr); err != nil {
		return nil, err
	}
	if err := validatePassword(in.Password); err != nil {
		return nil, err
	}

	hash, err := crypto.HashPassword(in.Password)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return nil, errInternal
	}

	u := &db.User{
		ID:           uuid.New(),
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		Email:        emailAddr,
		PasswordHash: &hash,
	}
	if err := s.user.Create(ctx, u); err != nil {
		if errors.Is(err, db.ErrEmailTaken) {
			return nil, ErrEmailTaken
		}
		s.logger.Error("failed to create user", zap.Error(err))
		return nil, errInternal
	}

	if err := s.sendVerification(ctx, u); err != nil {
		s.logger.Error("user registered but verification email was not sent",
			zap.String("userID", u.ID.String()), zap.Error(err))
	}

	return u, nil
}

// sendVerification generates a token, stores its hash, and emails the raw token.
// The token is stored before sending so a link can never be sent that we can't honor.
func (s *Service) sendVerification(ctx context.Context, u *db.User) error {
	token, err := crypto.GenerateRandomToken()
	if err != nil {
		return fmt.Errorf("generate verification token: %w", err)
	}

	t := &db.EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    u.ID,
		TokenHash: crypto.HashToken(token),
		ExpiresAt: time.Now().Add(email.VerificationLinkTTL),
	}
	if err := s.verification.Create(ctx, t); err != nil {
		return fmt.Errorf("store verification token: %w", err)
	}

	if err := s.sender.SendUserVerification(ctx, u.Email, token); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}
	return nil
}

// VerifyEmail consumes a verification token and marks its user's email verified.
func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	_, err := s.verification.Consume(ctx, crypto.HashToken(token), time.Now())
	if err != nil {
		if errors.Is(err, db.ErrVerificationTokenInvalid) {
			s.logger.Warn("invalid verification token presented")
			return ErrInvalidVerification
		}
		s.logger.Error("failed to consume verification token", zap.Error(err))
		return errInternal
	}
	return nil
}
