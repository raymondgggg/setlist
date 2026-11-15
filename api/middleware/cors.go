package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	allowedMethods  = "POST, OPTIONS"
	allowedHeaders  = "Content-Type, Authorization"
	preflightMaxAge = "600"
)

// CORS returns middleware that allows cross-origin requests from allowedOrigin only.
// Only the scheme, host and port of allowedOrigin are used. If it cannot be parsed
// into an origin, no request is ever allowed.
func CORS(allowedOrigin string) gin.HandlerFunc {
	origin := normalizeOrigin(allowedOrigin)

	return func(c *gin.Context) {
		// Responses differ by Origin, so caches must key on it even for non-matching requests.
		c.Writer.Header().Add("Vary", "Origin")

		if origin != "" && c.Request.Header.Get("Origin") == origin {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")

			if c.Request.Method == http.MethodOptions {
				h.Set("Access-Control-Allow-Methods", allowedMethods)
				h.Set("Access-Control-Allow-Headers", allowedHeaders)
				h.Set("Access-Control-Max-Age", preflightMaxAge)
			}
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func normalizeOrigin(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return strings.ToLower(u.Scheme + "://" + u.Host)
}
