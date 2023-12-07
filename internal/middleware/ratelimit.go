package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/vearne/consul-cache/internal/resource"
	"golang.org/x/time/rate"
	"net/http"
)

func GlobalRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		gLimiter := resource.GlobalRatelimiter.Load().(*rate.Limiter)
		if !gLimiter.Allow() {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
}
