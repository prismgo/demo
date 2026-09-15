package ratelimitdemo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	cachecontract "github.com/prismgo/framework/contracts/cache"
	"github.com/prismgo/framework/ratelimit"
)

// failurePolicy selects how a rate limiter call treats a cache error.
type failurePolicy int

const (
	// failOpen allows the request when the cache is unavailable.
	failOpen failurePolicy = iota
	// failClosed blocks the request when the cache is unavailable.
	failClosed
)

// allow reports whether the policy admits a request that saw err.
func (p failurePolicy) allow(err error) bool {
	return p == failOpen || err == nil
}

// architectureScenario keeps every documented rate limiter contract referenced at compile time.
func architectureScenario() (string, error) {
	var (
		_ func(cachecontract.Repository) *ratelimit.RateLimiter                                       = ratelimit.New
		_ func(string, ratelimit.LimiterFunc)                                                         = ratelimit.For
		_ func(string) ratelimit.LimiterFunc                                                          = ratelimit.Limiter
		_ func(time.Duration, int) ratelimit.Limit                                                    = ratelimit.Every
		_ func(int) ratelimit.Limit                                                                   = ratelimit.PerSecond
		_ func(int) ratelimit.Limit                                                                   = ratelimit.PerMinute
		_ func(int, int) ratelimit.Limit                                                              = ratelimit.PerMinutes
		_ func(int) ratelimit.Limit                                                                   = ratelimit.PerHour
		_ func(int) ratelimit.Limit                                                                   = ratelimit.PerDay
		_ func() ratelimit.Limit                                                                      = ratelimit.None
		_ func(context.Context, string, int) (bool, error)                                            = ratelimit.TooManyAttempts
		_ func(context.Context, string, time.Duration) (int64, error)                                 = ratelimit.Hit
		_ func(context.Context, string, int, time.Duration, ratelimit.AttemptFunc) (any, bool, error) = ratelimit.Attempt
		_ ratelimit.Limit
		_ ratelimit.Result
		_ ratelimit.AfterFunc
		_ ratelimit.ResponseFunc
		_ ratelimit.AttemptFunc
		_ ratelimit.LimiterFunc
	)
	return "limiter=RateLimiter limit=Limit result=Result store=cache.Repository", nil
}

// limitContractScenario inspects every documented Limit field and builder method.
func limitContractScenario() (string, error) {
	limit := ratelimit.Limit{
		MaxAttempts:  5,
		Decay:        time.Minute,
		Key:          "user:1",
		Fallback:     "user:fallback",
		AfterFunc:    func(*gin.Context) bool { return true },
		ResponseFunc: func(*gin.Context, ratelimit.Result) {},
	}
	return fmt.Sprintf("max=%d decay=%s key=%s fallback=%s after=%t response=%t",
		limit.MaxAttempts, limit.Decay, limit.Key, limit.Fallback,
		limit.AfterFunc != nil, limit.ResponseFunc != nil), nil
}

// resultContractScenario builds a Result and exposes its documented fields.
func resultContractScenario() (string, error) {
	result := ratelimit.Result{
		Limit:       ratelimit.PerMinute(5),
		Key:         "user:1",
		MaxAttempts: 5,
		Attempts:    2,
		Remaining:   3,
		RetryAfter:  42,
		ResetAt:     1700000000,
	}
	return fmt.Sprintf("limit=%d key=%s max=%d attempts=%d remaining=%d retry=%d reset=%d",
		result.Limit.MaxAttempts, result.Key, result.MaxAttempts, result.Attempts,
		result.Remaining, result.RetryAfter, result.ResetAt), nil
}

// builderScenario reports the window a constructor produces.
func builderScenario(limit ratelimit.Limit) (string, error) {
	return fmt.Sprintf("max=%d decay=%s", limit.MaxAttempts, limit.Decay), nil
}

// keyDesignScenario keeps the documented key layers referenced at compile time.
func keyDesignScenario() (string, error) {
	var (
		_ func(*ratelimit.RateLimiter, string, string) string = (*ratelimit.RateLimiter).MiddlewareKey
		_ func(string) string                                 = ratelimit.CleanRateLimiterKey
	)
	return "business=orders:42 dimension=user:42 isolation=ratelimit:api:user:42", nil
}

// errorPolicyScenario contrasts fail-open and fail-closed cache error handling.
func errorPolicyScenario() (string, error) {
	cacheErr := errors.New("cache unavailable")
	return fmt.Sprintf("fail-open=%t fail-closed=%t", failOpen.allow(cacheErr), failClosed.allow(cacheErr)), nil
}

// laravelCompatibilityScenario keeps the documented Laravel parity surface referenced.
func laravelCompatibilityScenario() (string, error) {
	var (
		_ func(int) ratelimit.Limit                                                                   = ratelimit.PerMinute
		_ func(int, int) ratelimit.Limit                                                              = ratelimit.PerMinutes
		_ func(context.Context, string, int, time.Duration, ratelimit.AttemptFunc) (any, bool, error) = ratelimit.Attempt
		_ func(context.Context, string, int) (bool, error)                                            = ratelimit.TooManyAttempts
	)
	return "compatible=for,limit-builders,attempt,counters,headers,after,response boundary=sha256-hash,timer-prefix,no-artisan", nil
}
