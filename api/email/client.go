package email

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"net/url"
	"strings"
	"time"

	"github.com/resend/resend-go/v4"
	"go.uber.org/zap"
)

//go:embed templates/verify_email.html
var verifyEmailTemplateSrc string

var verifyEmailTemplate = template.Must(template.New("verify_email").Parse(verifyEmailTemplateSrc))

// VerificationLinkTTL is how long a verification link is valid. The token
// stored server-side must expire on the same schedule.
const VerificationLinkTTL = 24 * time.Hour

type Client struct {
	client          *resend.Client
	emailFrom       string
	frontendBaseURL string
	logger          *zap.Logger
}

func NewClient(c *resend.Client, l *zap.Logger, from, frontendBaseURL string) *Client {
	return &Client{
		client:          c,
		emailFrom:       from,
		frontendBaseURL: frontendBaseURL,
		logger:          l,
	}
}

func (c *Client) send(ctx context.Context, params *resend.SendEmailRequest) error {
	_, err := c.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		c.logger.Error("failed to send email",
			zap.Error(err),
			zap.Strings("to", params.To),
			zap.String("subject", params.Subject))
		return err
	}
	return nil
}

// verificationLink builds the frontend URL the user clicks to verify their email.
func verificationLink(baseURL, token string) string {
	return strings.TrimRight(baseURL, "/") + "/account/verify-email?token=" + url.QueryEscape(token)
}

// renderVerification returns the HTML and plain-text bodies for the verification email.
func renderVerification(link string) (htmlBody, textBody string, err error) {
	var buf bytes.Buffer
	if err := verifyEmailTemplate.Execute(&buf, struct {
		Link      string
		ExpiresIn string
	}{Link: link, ExpiresIn: fmt.Sprintf("%d hours", int(VerificationLinkTTL.Hours()))}); err != nil {
		return "", "", fmt.Errorf("render verification email: %w", err)
	}
	textBody = "Welcome to Setlist! Confirm your email address to finish setting up your account:\n\n" +
		link + "\n\n" +
		fmt.Sprintf("This link expires in %d hours.\n\n", int(VerificationLinkTTL.Hours())) +
		"If you didn't create a Setlist account, you can safely ignore this email."
	return buf.String(), textBody, nil
}

func (c *Client) SendUserVerification(ctx context.Context, to, token string) error {
	htmlBody, textBody, err := renderVerification(verificationLink(c.frontendBaseURL, token))
	if err != nil {
		c.logger.Error("failed to render verification email", zap.Error(err))
		return err
	}

	params := &resend.SendEmailRequest{
		From:    c.emailFrom,
		To:      []string{to},
		Subject: "Setlist: Please verify your email",
		Html:    htmlBody,
		Text:    textBody,
	}
	return c.send(ctx, params)
}
