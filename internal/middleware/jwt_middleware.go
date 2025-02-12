package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/yourusername/user-management-api/pkg/token"
)

func JWTAuthMiddleware(tokenManager *token.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractTokenFromHeader(c.GetHeader("Authorization"))
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		claims, err := tokenManager.ValidateToken(tokenString, token.AccessToken)
		if err != nil {
			switch err.(type) {
			case token.TokenExpiredError:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			case token.TokenBlacklistedError:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token blacklisted"})
			default:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			}
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func extractTokenFromHeader(header string) string {
	if header == "" {
		log.Info().Msg("Authorization header missing")
		return ""
	}
	parts := strings.Split(header, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		return parts[1]
	}
	return ""
}
