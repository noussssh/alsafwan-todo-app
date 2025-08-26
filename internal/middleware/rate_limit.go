package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func RateLimitMiddleware(rate limiter.Rate) gin.HandlerFunc {
	store := memory.NewStore()
	instance := limiter.New(store, rate)
	
	return func(c *gin.Context) {
		key := c.ClientIP()
		
		ctx := context.Background()
		res, err := instance.Get(ctx, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limiting error"})
			c.Abort()
			return
		}
		
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", res.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", res.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", res.Reset))
		
		if res.Reached {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"retry_after": res.Reset,
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

func LoginRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(limiter.Rate{
		Period: 3 * time.Minute,
		Limit:  10,
	})
}