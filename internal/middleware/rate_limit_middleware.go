package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
)

func RateLimitMiddleware(limit int, burst int64, duration time.Duration) gin.HandlerFunc {
	rateLimiter := ratelimit.NewBucket(duration, burst) // e.g., 10 requests per second, burst of 20

	return func(c *gin.Context) {
		if rateLimiter.TakeAvailable(burst) == 0 { // Check if a token is available
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too Many Requests"})
			return
		}

		c.Next() // Continue to the next handler
	}
}
