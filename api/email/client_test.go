package email

import (
	"os"
	"setlist/config"
	"testing"

	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
	c := NewClient(rc, zap.NewNop(), cfg.EmailFrom)

	t.Run("delivered", func(t *testing.T) {
		err := c.SendUserVerification(t.Context(), "delivered@resend.dev", "test-token")
		require.NoError(t, err)
	})
	t.Run("bounced address is still accepted by the API", func(t *testing.T) {
		err := c.SendUserVerification(t.Context(), "bounced@resend.dev", "test-token")
		require.NoError(t, err)
	})
}
