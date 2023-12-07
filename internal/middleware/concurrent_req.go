package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/vearne/consul-cache/internal/resource"
	"net/http"
	"sync/atomic"
)

func ConcurrentReq() gin.HandlerFunc {
	return func(c *gin.Context) {
		atomic.AddInt64(&resource.ConcurrentReq, 1)
		defer atomic.AddInt64(&resource.ConcurrentReq, -1)

		if atomic.LoadInt64(&resource.ConcurrentReq) > resource.MaxConcurrent.Value() {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
}
