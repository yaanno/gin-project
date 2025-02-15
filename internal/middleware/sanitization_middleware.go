package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
)

// SanitizationMiddleware sanitizes incoming request bodies
func SanitizationMiddleware() gin.HandlerFunc {
	// Strict HTML sanitization policy
	policy := bluemonday.StrictPolicy()

	return func(c *gin.Context) {
		// Only sanitize for specific content types
		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			c.Next()
			return
		}

		// Read the body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Unable to read request body",
			})
			return
		}

		// Sanitize the body
		sanitizedBody := policy.Sanitize(string(body))

		// Replace the request body with sanitized version
		c.Request.Body = io.NopCloser(bytes.NewBufferString(sanitizedBody))

		c.Next()
	}
}

// Ensure the middleware implements the gin.HandlerFunc interface
var _ gin.HandlerFunc = SanitizationMiddleware()
