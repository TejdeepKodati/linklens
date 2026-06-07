package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter implements a sliding-window counter in Redis.
type RateLimiter struct {
	rdb *redis.Client
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

// Limit returns a Gin middleware that allows at most `limit` requests per `window`
// per unique key (IP or user_id). The `prefix` differentiates different limiters.
func (rl *RateLimiter) Limit(prefix string, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use user_id if authenticated, otherwise fall back to IP
		key := c.GetString("user_id")
		if key == "" {
			key = c.ClientIP()
		}
		redisKey := fmt.Sprintf("rl:%s:%s", prefix, key)

		ctx := context.Background()
		pipe := rl.rdb.Pipeline()
		incr := pipe.Incr(ctx, redisKey)
		pipe.Expire(ctx, redisKey, window)

		if _, err := pipe.Exec(ctx); err != nil {
			// On Redis error, allow request (fail open)
			c.Next()
			return
		}

		count := incr.Val()
		remaining := limit - count
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Window", window.String())

		if count > limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			return
		}

		c.Next()
	}
}
