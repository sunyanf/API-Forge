package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/response"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rate     int           // requests per window
	window   time.Duration // window size
}

type visitor struct {
	tokens    int
	lastReset time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}
	// Start cleanup goroutine to remove expired entries
	go rl.cleanup()
	return rl
}

// cleanup removes expired entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, v := range rl.visitors {
			if now.Sub(v.lastReset) > rl.window {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request is allowed for the given key
func (rl *RateLimiter) Allow(key string) (bool, int, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[key]

	if !exists {
		rl.visitors[key] = &visitor{
			tokens:    rl.rate - 1,
			lastReset: now,
		}
		return true, rl.rate - 1, now.Add(rl.window)
	}

	// Reset if window has passed
	if now.Sub(v.lastReset) > rl.window {
		v.tokens = rl.rate - 1
		v.lastReset = now
		return true, rl.rate - 1, now.Add(rl.window)
	}

	// Deny if no tokens left
	if v.tokens <= 0 {
		return false, 0, v.lastReset.Add(rl.window)
	}

	v.tokens--
	return true, v.tokens, v.lastReset.Add(rl.window)
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(rate int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, window)

	return func(c *gin.Context) {
		// Use IP or user_id as key
		key := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(uint); ok {
				key = strconv.FormatUint(uint64(uid), 10)
			} else {
				// 使用客户端IP作为key
				key = c.ClientIP()
			}
		}

		allowed, remaining, resetTime := limiter.Allow(key)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(rate))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			response.Error(c, http.StatusTooManyRequests, "rate limit exceeded, please try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}
