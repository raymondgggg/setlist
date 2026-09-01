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
