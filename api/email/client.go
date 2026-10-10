package email

import (
	"context"

	"github.com/resend/resend-go/v4"
	"go.uber.org/zap"
)

type Client struct {
	client    *resend.Client
	emailFrom string
	logger    *zap.Logger
}

func NewClient(c *resend.Client, l *zap.Logger, from string) *Client {
	return &Client{
		client:    c,
		emailFrom: from,
		logger:    l,
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

func (c *Client) SendUserVerification(ctx context.Context, to, token string) error {
	params := &resend.SendEmailRequest{
		From:    c.emailFrom,
		To:      []string{to},
		Subject: "Setlist: Please verify your email",
		Html:    "<h1>Hello world!</h1>",
	}
	return c.send(ctx, params)
}
