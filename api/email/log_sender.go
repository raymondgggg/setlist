package email

import (
	"context"

	"go.uber.org/zap"
)

// LogSender is a local-development stand-in for Client that logs the
// verification link instead of sending anything. The link contains the raw
// verification token, so this must never be used outside local development
// (config refuses EMAIL_LOG_ONLY in production).
type LogSender struct {
	logger          *zap.Logger
	frontendBaseURL string
}

func NewLogSender(l *zap.Logger, frontendBaseURL string) *LogSender {
	return &LogSender{logger: l, frontendBaseURL: frontendBaseURL}
}

func (s *LogSender) SendUserVerification(ctx context.Context, to, token string) error {
	s.logger.Warn("DEV ONLY: verification email not sent, use this link instead",
		zap.String("to", to),
		zap.String("link", verificationLink(s.frontendBaseURL, token)))
	return nil
}
