package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"setlist/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testSecret = []byte("test-secret")

func serveAuth(authorization string) (*httptest.ResponseRecorder, *uuid.UUID) {
	gin.SetMode(gin.TestMode)
	var got *uuid.UUID
	r := gin.New()
	r.Use(Authenticate(testSecret))
	r.POST("/graphql", func(c *gin.Context) {
		if id, ok := auth.UserIDFromContext(c.Request.Context()); ok {
			got = &id
		}
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w, got
}

func TestAuthenticate(t *testing.T) {
	id := uuid.New()
	valid, err := auth.GenerateAccessToken(id, testSecret, time.Minute)
	require.NoError(t, err)
	expired, err := auth.GenerateAccessToken(id, testSecret, -time.Minute)
	require.NoError(t, err)
	wrongSecret, err := auth.GenerateAccessToken(id, []byte("other"), time.Minute)
	require.NoError(t, err)
	nilUser, err := auth.GenerateAccessToken(uuid.Nil, testSecret, time.Minute)
	require.NoError(t, err)

	t.Run("no header continues as anonymous", func(t *testing.T) {
		w, got := serveAuth("")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Nil(t, got)
	})

	t.Run("valid token sets the user ID", func(t *testing.T) {
		w, got := serveAuth("Bearer " + valid)

		assert.Equal(t, http.StatusOK, w.Code)
		require.NotNil(t, got)
		assert.Equal(t, id, *got)
	})

	t.Run("scheme is case insensitive", func(t *testing.T) {
		w, got := serveAuth("bearer " + valid)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotNil(t, got)
	})

	t.Run("expired token", func(t *testing.T) {
		w, got := serveAuth("Bearer " + expired)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), CodeTokenExpired)
		assert.Nil(t, got)
	})

	for name, header := range map[string]string{
		"wrong secret":         "Bearer " + wrongSecret,
		"garbage token":        "Bearer not-a-jwt",
		"nil user ID":          "Bearer " + nilUser,
		"wrong scheme":         "Basic " + valid,
		"missing token":        "Bearer",
		"empty token":          "Bearer  ",
		"token without scheme": valid,
	} {
		t.Run("invalid: "+name, func(t *testing.T) {
			w, got := serveAuth(header)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), CodeInvalidToken)
			assert.Contains(t, w.Body.String(), `"errors"`)
			assert.Nil(t, got)
		})
	}
}
