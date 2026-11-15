package middleware

import (
	"errors"
	"net/http"
	"strings"

	"setlist/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Error codes returned in the extensions of a 401 response so the client can
// decide whether to refresh the token or log in again.
const (
	CodeTokenExpired = "TOKEN_EXPIRED"
	CodeInvalidToken = "INVALID_TOKEN"
)

// Authenticate returns middleware that verifies the Bearer access token and
// stores the user ID in the request context. A request without an
// Authorization header continues as anonymous. A request with a malformed
// header or a token that fails verification is rejected with 401, so the
// client alone decides when to discard or refresh its token.
func Authenticate(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.Next()
			return
		}

		scheme, token, ok := strings.Cut(header, " ")
		token = strings.TrimSpace(token)
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			abortUnauthenticated(c, CodeInvalidToken, "Invalid authorization header.")
			return
		}

		claims, err := auth.ParseAccessToken(token, secret)
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			abortUnauthenticated(c, CodeTokenExpired, "Access token has expired.")
			return
		case err != nil || claims.UserID == uuid.Nil:
			abortUnauthenticated(c, CodeInvalidToken, "Invalid access token.")
			return
		}

		c.Request = c.Request.WithContext(auth.WithUserID(c.Request.Context(), claims.UserID))
		c.Next()
	}
}

// The body mirrors a GraphQL error response so Apollo clients can read the code.
func abortUnauthenticated(c *gin.Context, code, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"errors": []gin.H{{
			"message":    message,
			"extensions": gin.H{"code": code},
		}},
	})
}
