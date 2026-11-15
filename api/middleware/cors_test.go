package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const testOrigin = "http://localhost:5173"

func newRouter(allowedOrigin string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS(allowedOrigin))
	r.POST("/graphql", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	return r
}

func do(r *gin.Engine, method, origin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/graphql", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCORS(t *testing.T) {
	t.Run("matching origin", func(t *testing.T) {
		w := do(newRouter(testOrigin), http.MethodPost, testOrigin)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, testOrigin, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "Origin", w.Header().Get("Vary"))
	})
	t.Run("non-matching origin", func(t *testing.T) {
		w := do(newRouter(testOrigin), http.MethodPost, "http://evil.example")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "Origin", w.Header().Get("Vary"))
	})
	t.Run("no origin", func(t *testing.T) {
		w := do(newRouter(testOrigin), http.MethodPost, "")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", w.Header().Get("Vary"))
	})
	t.Run("preflight from matching origin", func(t *testing.T) {
		w := do(newRouter(testOrigin), http.MethodOptions, testOrigin)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, testOrigin, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "POST, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
		assert.NotEmpty(t, w.Header().Get("Access-Control-Max-Age"))
		assert.Equal(t, "Origin", w.Header().Get("Vary"))
	})
	t.Run("preflight from non-matching origin", func(t *testing.T) {
		w := do(newRouter(testOrigin), http.MethodOptions, "http://evil.example")

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"))
	})
	t.Run("allowed origin with trailing slash and path", func(t *testing.T) {
		w := do(newRouter("http://localhost:5173/app/"), http.MethodPost, testOrigin)

		assert.Equal(t, testOrigin, w.Header().Get("Access-Control-Allow-Origin"))
	})
	t.Run("different port doesn't match", func(t *testing.T) {
		w := do(newRouter(testOrigin), http.MethodPost, "http://localhost:3000")

		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
	t.Run("unparsable allowed origin allows nothing", func(t *testing.T) {
		w := do(newRouter(""), http.MethodPost, testOrigin)

		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}
