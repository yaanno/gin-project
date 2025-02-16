package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rs/zerolog"
)

// SanitizationMiddleware sanitizes incoming request bodies
func SanitizationMiddleware(log *zerolog.Logger) gin.HandlerFunc {
	// Strict HTML sanitization policy
	policy := bluemonday.StrictPolicy()

	return func(c *gin.Context) {
		// Only sanitize for specific content types
		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			// log.Error().
			// 	Str("content_type", contentType).
			// 	Msg("Skipping body sanitization")
			c.Next()
			return
		}

		// Read the body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Unable to read request body")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// Sanitize the body
		sanitizedBody := policy.Sanitize(string(body))

		// Replace the request body with sanitized version
		c.Request.Body = io.NopCloser(bytes.NewBufferString(sanitizedBody))

		c.Next()
	}
}

var logger zerolog.Logger

// Ensure the middleware implements the gin.HandlerFunc interface
var _ gin.HandlerFunc = SanitizationMiddleware(&logger)
