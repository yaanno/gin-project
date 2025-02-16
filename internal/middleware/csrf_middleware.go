package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/pkg/errors/apperrors"
)

type CSRFProtection struct {
	logger         *zerolog.Logger
	cookieName     string
	headerName     string
	cookieMaxAge   time.Duration
	cookiePath     string
	cookieDomain   string
	cookieSecure   bool
	cookieHttpOnly bool
	excludedRoutes []string
}

type CSRFOption func(*CSRFProtection)

func NewCSRFMiddleware(logger *zerolog.Logger, opts ...CSRFOption) *CSRFProtection {
	csrf := &CSRFProtection{
		logger:         logger,
		cookieName:     "X-CSRF-Token",
		headerName:     "X-CSRF-Token",
		cookieMaxAge:   1 * time.Hour,
		cookiePath:     "/",
		cookieSecure:   true,
		cookieHttpOnly: true,
	}

	// Apply custom options
	for _, opt := range opts {
		opt(csrf)
	}

	return csrf
}

func WithCookieName(name string) CSRFOption {
	return func(csrf *CSRFProtection) {
		csrf.cookieName = name
	}
}

func WithCookieDomain(domain string) CSRFOption {
	return func(csrf *CSRFProtection) {
		csrf.cookieDomain = domain
	}
}

func WithExcludedRoutes(routes []string) CSRFOption {
	return func(csrf *CSRFProtection) {
		csrf.excludedRoutes = routes
	}
}

// Secure time-constant string comparison
func (csrf *CSRFProtection) validateToken(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

func (csrf *CSRFProtection) isExcludedRoute(route string) bool {
	for _, excludedRoute := range csrf.excludedRoutes {
		if excludedRoute == route {
			return true
		}
	}
	return false
}

func (csrf *CSRFProtection) generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (csrf *CSRFProtection) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if csrf.isExcludedRoute(c.Request.URL.Path) {
			c.Next()
			return
		}

		csrfCookie, err := c.Request.Cookie(csrf.cookieName)
		var csrfToken string

		if err != nil || csrfCookie.Value == "" {
			csrfToken, err = csrf.generateToken()
			if err != nil {
				csrf.logger.Error().Err(err).Msg("failed to generate csrf token")
				c.Error(err)
			}

			c.Set(csrf.cookieName, csrfToken)
			c.SetSameSite(http.SameSiteStrictMode)
			c.SetCookie(csrf.cookieName,
				csrfToken,
				int(csrf.cookieMaxAge.Seconds()),
				csrf.cookiePath,
				csrf.cookieDomain,
				csrf.cookieSecure,
				csrf.cookieHttpOnly,
			)
		} else {
			csrfToken = csrfCookie.Value
		}

		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			token := c.GetHeader(csrf.headerName)

			if token == "" {
				csrf.logger.Error().Msg("missing csrf token")
				c.Error(apperrors.NewAuthenticationError(apperrors.ErrCodeInvalidCSRFToken, "missing csrf token", nil))
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			if !csrf.validateToken(token, csrfToken) {
				csrf.logger.Error().Msg("invalid csrf token")
				c.Error(apperrors.NewAuthenticationError(apperrors.ErrCodeInvalidCSRFToken, "invalid csrf token", nil))
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}
		c.Next()
	}
}
