package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
	"github.com/rs/zerolog"
)

type IPRateLimiter struct {
	mu          sync.Mutex
	limiters    map[string]*ratelimit.Bucket
	globalLimit int
	globalBurst int64
	duration    time.Duration
	logger      zerolog.Logger
}

func NewIPRateLimiter(globalLimit int, globalBurst int64, duration time.Duration, logger zerolog.Logger) *IPRateLimiter {
	return &IPRateLimiter{
		limiters:    make(map[string]*ratelimit.Bucket),
		globalLimit: globalLimit,
		globalBurst: globalBurst,
		duration:    duration,
		logger:      logger,
	}
}

func (l *IPRateLimiter) logRateLimitEvent(ip string, path string) {
	l.logger.Warn().
		Str("event", "rate_limit_triggered").
		Str("ip_address", ip).
		Str("request_path", path).
		Msg("IP rate limit exceeded")
}

func (l *IPRateLimiter) trackIPActivity(ip string, allowed bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Optional: Implement more advanced tracking if needed
	if !allowed {
		// You could implement a counter or more complex tracking mechanism
		l.logger.Info().
			Str("event", "ip_activity_tracked").
			Str("ip_address", ip).
			Bool("rate_limit_passed", allowed).
			Msg("IP activity monitored")
	}
}

func (l *IPRateLimiter) getIPLimiter(ip string) *ratelimit.Bucket {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[ip]
	if !exists {
		l.logger.Info().
			Str("event", "ip_limiter_created").
			Str("ip_address", ip).
			Msg("IP rate limiter created")
		limiter = ratelimit.NewBucket(l.duration, l.globalBurst)
		l.limiters[ip] = limiter
	}
	return limiter
}

func (l *IPRateLimiter) cleanupOldLimiters() {
	// Periodically clean up old IP limiters to prevent memory leaks
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			l.mu.Lock()
			for ip, limiter := range l.limiters {
				if limiter.Available() == l.globalBurst {
					delete(l.limiters, ip)
					l.logger.Info().
						Str("event", "ip_limiter_deleted").
						Str("ip_address", ip).
						Msg("IP rate limiter deleted")
				}
			}
			l.mu.Unlock()
		}
	}()
}

func IPRateLimitMiddleware(globalLimit int, globalBurst int64, duration time.Duration, logger zerolog.Logger) gin.HandlerFunc {
	ipRateLimiter := NewIPRateLimiter(globalLimit, globalBurst, duration, logger)
	ipRateLimiter.cleanupOldLimiters()

	return func(c *gin.Context) {
		// Get the real IP, handling potential proxy scenarios
		ip := getRealIP(c)

		limiter := ipRateLimiter.getIPLimiter(ip)

		// Take one token, check if available
		if limiter.TakeAvailable(1) == 0 {
			// Log the rate limit event
			ipRateLimiter.logRateLimitEvent(ip, c.Request.URL.Path)

			// Track IP activity
			ipRateLimiter.trackIPActivity(ip, false)

			// Respond with rate limit error
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{})
			return
		}

		c.Next()
	}
}

// Helper function to get the real client IP
func getRealIP(c *gin.Context) string {
	// Check for X-Forwarded-For header (common with proxies)
	ip := c.GetHeader("X-Forwarded-For")
	if ip != "" {
		// Take the first IP if multiple are present
		return strings.Split(ip, ",")[0]
	}

	// Fallback to RemoteIP
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		ip = c.Request.RemoteAddr
	}

	return ip
}
