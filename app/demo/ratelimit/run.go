// Package ratelimitdemo contains runnable examples of the Laravel-style rate limiter.
package ratelimitdemo

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prismgo/framework/cache"
	"github.com/prismgo/framework/ratelimit"
)

// Result records one observable rate limiter scenario.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

// Run executes a rate limiter catalog scenario.
func Run(name string) (Result, error) {
	value, err := run(name)
	if err != nil {
		return Result{}, fmt.Errorf("ratelimit demo %s: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

// run isolates Gin's process-wide mode and writer while a scenario executes.
func run(name string) (string, error) {
	mode := gin.Mode()
	writer := gin.DefaultWriter
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	defer func() {
		gin.SetMode(mode)
		gin.DefaultWriter = writer
	}()

	switch name {
	case "architecture":
		return architectureScenario()
	case "config":
		return configScenario()
	case "store-resolution":
		return storeResolutionScenario()
	case "memory-store":
		return memoryStoreScenario()
	case "redis-store":
		return redisStoreScenario()
	case "config-registration":
		return configRegistrationScenario()
	case "auto-initialization":
		return autoInitializationScenario()
	case "quick-start":
		return quickStartScenario()
	case "explicit-limiter":
		return explicitLimiterScenario()
	case "named-registration":
		return namedRegistrationScenario()
	case "named-lookup":
		return namedLookupScenario()
	case "unregistered-pass-through":
		return unregisteredPassThroughScenario()
	case "multi-rule-counters":
		return multiRuleCountersScenario()
	case "multi-rule-block":
		return multiRuleBlockScenario()
	case "limit-contract":
		return limitContractScenario()
	case "every":
		return builderScenario(ratelimit.Every(30*time.Second, 10))
	case "per-second":
		return builderScenario(ratelimit.PerSecond(2))
	case "per-minute":
		return builderScenario(ratelimit.PerMinute(3))
	case "per-minutes":
		return builderScenario(ratelimit.PerMinutes(10, 20))
	case "per-hour":
		return builderScenario(ratelimit.PerHour(5))
	case "per-day":
		return builderScenario(ratelimit.PerDay(7))
	case "none":
		return builderScenario(ratelimit.None())
	case "by":
		return byScenario()
	case "fallback-key":
		return fallbackKeyScenario()
	case "after-count":
		return afterCountScenario()
	case "after-skip":
		return afterSkipScenario()
	case "after-headers":
		return afterHeadersScenario()
	case "custom-response":
		return customResponseScenario()
	case "result-contract":
		return resultContractScenario()
	case "throttle":
		return throttleScenario()
	case "throttle-for":
		return throttleForScenario()
	case "disabled-rule":
		return disabledRuleScenario()
	case "default-response":
		return defaultResponseScenario()
	case "success-headers":
		return successHeadersScenario()
	case "over-limit-headers":
		return overLimitHeadersScenario()
	case "tightest-headers":
		return tightestHeadersScenario()
	case "route-group":
		return routeGroupScenario()
	case "route-compatibility":
		return routeCompatibilityScenario()
	case "facade":
		return facadeScenario()
	case "instance":
		return instanceScenario()
	case "hit":
		return hitScenario()
	case "hit-default-decay":
		return hitDefaultDecayScenario()
	case "increment":
		return incrementScenario()
	case "increment-default":
		return incrementDefaultScenario()
	case "decrement":
		return decrementScenario()
	case "decrement-default":
		return decrementDefaultScenario()
	case "attempts":
		return attemptsScenario()
	case "attempts-missing":
		return missingAttemptsScenario()
	case "too-many-attempts":
		return tooManyAttemptsScenario()
	case "non-positive-limit":
		return nonPositiveLimitScenario()
	case "expired-window":
		return expiredWindowScenario()
	case "remaining":
		return remainingScenario()
	case "remaining-floor":
		return remainingFloorScenario()
	case "available-in":
		return availableInScenario()
	case "available-in-elapsed":
		return availableInElapsedScenario()
	case "reset-attempts":
		return resetAttemptsScenario()
	case "clear":
		return clearScenario()
	case "attempt-success":
		return attemptSuccessScenario()
	case "attempt-blocked":
		return attemptBlockedScenario()
	case "attempt-error":
		return attemptErrorScenario()
	case "attempt-nil":
		return attemptNilScenario()
	case "clean-key":
		return cleanKeyScenario()
	case "hashed-key":
		return hashedKeyScenario()
	case "unhashed-key":
		return unhashedKeyScenario()
	case "manual-key":
		return manualKeyScenario()
	case "key-design":
		return keyDesignScenario()
	case "cache-layout":
		return cacheLayoutScenario()
	case "fixed-window":
		return fixedWindowScenario()
	case "middleware-key":
		return middlewareKeyScenario()
	case "cache-errors":
		return cacheErrorsScenario()
	case "redis-errors":
		return redisErrorsScenario()
	case "counter-type-error":
		return counterTypeErrorScenario()
	case "error-policy":
		return errorPolicyScenario()
	case "laravel-compatibility":
		return laravelCompatibilityScenario()
	default:
		return "", fmt.Errorf("unknown scenario %q", name)
	}
}

// newTestLimiter builds an isolated in-process limiter and its owning manager.
func newTestLimiter() (*cache.Manager, *ratelimit.RateLimiter, error) {
	manager, err := newMemoryManager()
	if err != nil {
		return nil, nil, err
	}
	return manager, ratelimit.New(manager.Default()), nil
}

// perform runs one request through the engine and returns the recorded response.
func perform(engine *gin.Engine, method, target string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	return recorder
}

// testContext returns a Gin context suitable for invoking a limiter function.
func testContext() *gin.Context {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return ctx
}
