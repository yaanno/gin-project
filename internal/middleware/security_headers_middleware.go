package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type SecurityHeadersMiddleware struct {
	// Optional configuration options
	allowCSP  bool
	cspPolicy string
}

type SecurityHeadersOption func(*SecurityHeadersMiddleware)

// NewSecurityHeadersMiddleware creates a new middleware for adding security headers
func NewSecurityHeadersMiddleware(opts ...SecurityHeadersOption) *SecurityHeadersMiddleware {
	middleware := &SecurityHeadersMiddleware{
		allowCSP:  true,
		cspPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'",
	}

	// Apply any custom options
	for _, opt := range opts {
		opt(middleware)
	}

	return middleware
}

// WithCustomCSP allows customizing the Content Security Policy
func WithCustomCSP(policy string) SecurityHeadersOption {
	return func(m *SecurityHeadersMiddleware) {
		m.cspPolicy = policy
	}
}

// Handler implements the middleware logic
func (m *SecurityHeadersMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS protection in browsers
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict transport security (HSTS)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy
		if m.allowCSP {
			c.Header("Content-Security-Policy", m.cspPolicy)
		}

		// Permissions Policy (formerly Feature Policy)
		c.Header("Permissions-Policy",
			"geolocation=(), "+
				"camera=(), "+
				"microphone=(), "+
				"payment=()",
		)

		// Add server timing headers for performance insights
		start := time.Now()
		defer func() {
			duration := time.Since(start)
			c.Header("Server-Timing",
				"middleware;desc=\"Security Headers\";dur="+strconv.FormatInt(duration.Milliseconds(), 10),
			)
		}()

		c.Next()
	}
}
