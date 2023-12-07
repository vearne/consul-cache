package resource

import (
	"expvar"
	"golang.org/x/time/rate"
	"sync/atomic"
)

var (
	// 	*rate.Limiter
	GlobalRatelimiter atomic.Value
	GlobalRate        *expvar.Int

	MaxConcurrent *expvar.Int
)

const (
	DefaultGlobalRate    = 5000
	DefaultMaxConcurrent = 50000
)

func initRateLimiter() {
	GlobalRatelimiter.Store(rate.NewLimiter(rate.Limit(DefaultGlobalRate), DefaultGlobalRate))
	GlobalRate = expvar.NewInt("GlobalRate")
	GlobalRate.Set(DefaultGlobalRate)

	MaxConcurrent = expvar.NewInt("MaxConcurrentNum")
	MaxConcurrent.Set(DefaultMaxConcurrent)
}
