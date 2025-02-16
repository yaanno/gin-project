package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/pkg/authentication"
	"github.com/yourusername/user-management-api/pkg/errors/apperrors"
	"github.com/yourusername/user-management-api/pkg/token"
)

func AuthMiddleware(authManager *authentication.AuthenticationManagerImpl, logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new logger with the request URI, method, and middleware name
		logger = logger.With().Str("uri", c.Request.URL.Path).Str("method", c.Request.Method).Str("middleware", "AuthMiddleware").Logger()

		tokenString := extractTokenFromHeader(c.GetHeader("Authorization"))
		if tokenString == "" {
			logger.
				Error().
				Str("header", "Authorization").
				Msg("Authorization header missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		// Validate token
		// TODO: Add support for refresh tokens
		// TODO: Add general error messages without exposing internal error details
		claims, err := authManager.ValidateToken(tokenString, token.AccessToken)
		if err != nil {
			logger.Error().Err(err).Str("token", tokenString).Msg("Token validation failed")

			switch err.Code() {
			case apperrors.ErrCodeTokenExpired:
				logger.Error().
					Str("token", tokenString).
					Msg("Token expired")
				c.JSON(http.StatusUnauthorized, gin.H{})
			case apperrors.ErrCodeTokenBlacklisted:
				logger.Error().
					Str("token", tokenString).
					Msg("Token blacklisted")
				c.JSON(http.StatusUnauthorized, gin.H{})
			case apperrors.ErrCodeTokenInvalidType,
				apperrors.ErrCodeInvalidTokenSignature,
				apperrors.ErrCodeTokenMalformed,
				apperrors.ErrCodeTokenInvalidClaim:
				logger.Error().
					Str("token", tokenString).
					Msg("Invalid token")
				c.JSON(http.StatusBadRequest, gin.H{})
			default:
				logger.Error().
					Str("token", tokenString).
					Msg("Invalid token")
				c.JSON(http.StatusUnauthorized, gin.H{})
				// TODO: this should be internal error
			}
			c.Abort()
			return
		}

		// Additional user status check
		user, err := authManager.FindUserByUsername(claims.Username)
		if err != nil {
			logger.Error().
				Str("username", claims.Username).
				Msg("User not found")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
			return
		}

		// Check user status
		if err := authManager.CheckUserStatus(user); err != nil {
			logger.Error().
				Err(err).
				Str("username", claims.Username).
				Msg("User status check failed")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func extractTokenFromHeader(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.Split(header, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		return parts[1]
	}
	return ""
}
