package middleware

import (
	"fmt"
	"net/http"
	"time"

	"agileos-backend/logger"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	Period  time.Duration
	Limit   int64
	Message string
}

// Global rate limiter instances
var (
	globalLimiter *limiter.Limiter
	authLimiter   *limiter.Limiter
)

// InitRateLimiters initializes rate limiters
func InitRateLimiters() {
	// Global rate limiter: 100 requests per minute per IP
	globalStore := memory.NewStore()
	globalRate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  100,
	}
	globalLimiter = limiter.New(globalStore, globalRate)

	// Auth rate limiter: 5 login attempts per minute per IP (brute force protection)
	authStore := memory.NewStore()
	authRate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  5,
	}
	authLimiter = limiter.New(authStore, authRate)

	logger.Log.Info().Msg("✓ Rate limiters initialized")
}

// GlobalRateLimit applies global rate limiting to all endpoints
func GlobalRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		ip := c.ClientIP()

		// Check rate limit
		context, err := globalLimiter.Get(c.Request.Context(), ip)
		if err != nil {
			logger.LogError("Rate limiter error", err, map[string]interface{}{
				"ip": ip,
			})
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", context.Reset))

		// Check if limit exceeded
		if context.Reached {
			logger.LogSecurity("rate_limit_exceeded", "", ip, map[string]interface{}{
				"limit":     context.Limit,
				"endpoint":  c.Request.URL.Path,
				"method":    c.Request.Method,
				"user_agent": c.Request.UserAgent(),
			})

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
				"retry_after": context.Reset,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthRateLimit applies strict rate limiting to authentication endpoints
func AuthRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		ip := c.ClientIP()

		// Check rate limit
		context, err := authLimiter.Get(c.Request.Context(), ip)
		if err != nil {
			logger.LogError("Auth rate limiter error", err, map[string]interface{}{
				"ip": ip,
			})
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", context.Reset))

		// Check if limit exceeded
		if context.Reached {
			logger.LogSecurity("auth_rate_limit_exceeded", "", ip, map[string]interface{}{
				"limit":      context.Limit,
				"endpoint":   c.Request.URL.Path,
				"method":     c.Request.Method,
				"user_agent": c.Request.UserAgent(),
				"severity":   "HIGH", // Potential brute force attack
			})

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many login attempts",
				"message": "You have exceeded the maximum number of login attempts. Please try again later.",
				"retry_after": context.Reset,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CustomRateLimit creates a custom rate limiter with specified config
func CustomRateLimit(config RateLimitConfig) gin.HandlerFunc {
	store := memory.NewStore()
	rate := limiter.Rate{
		Period: config.Period,
		Limit:  config.Limit,
	}
	customLimiter := limiter.New(store, rate)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		context, err := customLimiter.Get(c.Request.Context(), ip)
		if err != nil {
			logger.LogError("Custom rate limiter error", err, map[string]interface{}{
				"ip": ip,
			})
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", context.Reset))

		if context.Reached {
			logger.LogSecurity("custom_rate_limit_exceeded", "", ip, map[string]interface{}{
				"limit":    context.Limit,
				"period":   config.Period.String(),
				"endpoint": c.Request.URL.Path,
			})

			message := config.Message
			if message == "" {
				message = "Rate limit exceeded. Please try again later."
			}

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     message,
				"retry_after": context.Reset,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
