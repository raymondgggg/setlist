package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequireEnv(t *testing.T) {
	t.Run("env variable exists", func(t *testing.T) {
		// Setup test environment variable
		t.Setenv("TEST_ENV", "hello")

		v, err := requireEnv("TEST_ENV")
		assert.NoError(t, err)
		assert.Equal(t, "hello", v)
	})
	t.Run("env variable doesn't exist", func(t *testing.T) {
		v, err := requireEnv("TEST_ENV")
		assert.Error(t, err)
		assert.Equal(t, "", v)
	})
}

func TestLoadEnv(t *testing.T) {
	t.Run("missing .env file with vars set in environment", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "secret")
		t.Setenv("RESEND_API_KEY", "key")
		t.Setenv("EMAIL_FROM", "Setlist <noreply@example.com>")
		t.Setenv("FRONTEND_BASE_URL", "http://localhost:5173")

		c, err := LoadEnv()
		assert.NoError(t, err)
		assert.Equal(t, "secret", c.JWTSecret)
		assert.Equal(t, Environment(Development), c.Env)
	})
	t.Run("missing .env file and missing required var", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "")

		c, err := LoadEnv()
		assert.ErrorContains(t, err, "JWT_SECRET is required")
		assert.Nil(t, c)
	})
}
