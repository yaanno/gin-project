package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/pkg/authentication"
	"github.com/yourusername/user-management-api/pkg/token"
)

func AuthMiddleware(authManager *authentication.AuthenticationManagerImpl, logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new logger with the request URI, method, and middleware name
		logger = logger.With().Str("uri", c.Request.URL.Path).Str("method", c.Request.Method).Str("middleware", "JWTAuthMiddleware").Logger()

		tokenString := extractTokenFromHeader(c.GetHeader("Authorization"))
		if tokenString == "" {
			logger.
				Info().
				Msg("Authorization header missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		claims, err := authManager.ValidateToken(tokenString, token.AccessToken)
		if err != nil {
			switch err.(type) {
			case token.TokenExpiredError:
				logger.Error().Msg("Token expired")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			case token.TokenBlacklistedError:
				logger.Error().Msg("Token blacklisted")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token blacklisted"})
			default:
				logger.Error().Msg("Invalid token")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			}
			c.Abort()
			return
		}

		// Additional user status check
		user, err := authManager.FindUserByUsername(claims.Username)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "User not found",
			})
			return
		}

		// Check user status
		if err := authManager.CheckUserStatus(user); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
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
