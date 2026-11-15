package email

import (
	"os"
	"setlist/config"
	"testing"

	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// TestSendUserVerification_Live sends real emails through Resend, so it only
// runs when EMAIL_LIVE_TEST=1 (e.g. `EMAIL_LIVE_TEST=1 go test ./email -run Live -v`).
func TestSendUserVerification_Live(t *testing.T) {
	if os.Getenv("EMAIL_LIVE_TEST") != "1" {
		t.Skip("set EMAIL_LIVE_TEST=1 to send a real email via Resend")
	}

	// go test runs from the package dir; .env lives one level up. Existing
	// env vars win, so this is a no-op when the vars are already exported.
	_ = godotenv.Load("../.env")

	cfg, err := config.LoadEnv()
	require.NoError(t, err)

	rc := resend.NewClient(cfg.ResendAPIKey)
	c := NewClient(rc, zap.NewNop(), cfg.EmailFrom, cfg.FrontendBaseURL)

	t.Run("delivered", func(t *testing.T) {
		err := c.SendUserVerification(t.Context(), "delivered@resend.dev", "test-token")
		require.NoError(t, err)
	})
	t.Run("bounced address is still accepted by the API", func(t *testing.T) {
		err := c.SendUserVerification(t.Context(), "bounced@resend.dev", "test-token")
		require.NoError(t, err)
	})
}

func TestVerificationLink(t *testing.T) {
	t.Run("builds link from base url", func(t *testing.T) {
		link := verificationLink("http://localhost:5173", "abc123")
		require.Equal(t, "http://localhost:5173/account/verify-email?token=abc123", link)
	})
	t.Run("trailing slash on base url is tolerated", func(t *testing.T) {
		link := verificationLink("http://localhost:5173/", "abc123")
		require.Equal(t, "http://localhost:5173/account/verify-email?token=abc123", link)
	})
	t.Run("token is query escaped", func(t *testing.T) {
		link := verificationLink("http://localhost:5173", "a b&c=d")
		require.Equal(t, "http://localhost:5173/account/verify-email?token=a+b%26c%3Dd", link)
	})
}

func TestRenderVerification(t *testing.T) {
	link := verificationLink("http://localhost:5173", "abc123")

	htmlBody, textBody, err := renderVerification(link)
	require.NoError(t, err)

	require.Contains(t, htmlBody, `href="`+link+`"`)
	require.Contains(t, htmlBody, "expires in 24 hours")
	require.Contains(t, textBody, link)
	require.Contains(t, textBody, "expires in 24 hours")
}

func TestLogSender(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	s := NewLogSender(zap.New(core), "http://localhost:5173")

	err := s.SendUserVerification(t.Context(), "johndoe@gmail.com", "abc123")
	require.NoError(t, err)

	require.Equal(t, 1, logs.Len())
	fields := logs.All()[0].ContextMap()
	require.Equal(t, "johndoe@gmail.com", fields["to"])
	require.Equal(t, "http://localhost:5173/account/verify-email?token=abc123", fields["link"])
}
