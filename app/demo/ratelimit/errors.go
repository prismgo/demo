package ratelimitdemo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/prismgo/framework/cache"
	configpkg "github.com/prismgo/framework/config"
	"github.com/prismgo/framework/ratelimit"
	frameworkredis "github.com/prismgo/framework/redis"
)

// cacheErrorsScenario propagates cache operation failures to the caller.
func cacheErrorsScenario() (string, error) {
	manager, err := newMemoryManager()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	limiter := ratelimit.New(&controlledRepository{Repository: manager.Default(), failAll: true})
	ctx := context.Background()

	_, hitErr := limiter.Hit(ctx, "failure", time.Minute)
	_, attemptsErr := limiter.Attempts(ctx, "failure")
	_, tooManyErr := limiter.TooManyAttempts(ctx, "failure", 3)
	resetErr := limiter.ResetAttempts(ctx, "failure")
	_, availableErr := limiter.AvailableIn(ctx, "failure")

	return fmt.Sprintf("hit=%t attempts=%t too-many=%t reset=%t available=%t",
		hitErr != nil, attemptsErr != nil, tooManyErr != nil, resetErr != nil, availableErr != nil), nil
}

// counterTypeErrorScenario propagates a non-integer counter value error.
func counterTypeErrorScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	if err := manager.Default().Put(ctx, "bad-counter", "not-a-number", 0); err != nil {
		return "", fmt.Errorf("seed invalid counter: %w", err)
	}
	_, err = limiter.Attempts(ctx, "bad-counter")
	return fmt.Sprintf("invalid=%t", errors.Is(err, cache.ErrInvalidCounter)), nil
}

// redisErrorsScenario propagates a live Redis connection failure.
func redisErrorsScenario() (string, error) {
	driver := strings.TrimSpace(configpkg.GetString("cache.limiter.driver", ""))
	if driver == "" {
		driver = cache.DefaultName()
	}
	if driver != "redis" {
		return "", fmt.Errorf("redis-errors scenario requires CACHE_LIMITER_DRIVER=redis, got %q", driver)
	}
	if strings.TrimSpace(os.Getenv("PRISMGO_REDIS_TEST_URL")) == "" {
		return "", errors.New("PRISMGO_REDIS_TEST_URL is required for the redis-errors scenario")
	}

	limiter := ratelimit.Resolve()
	if limiter == nil {
		return "", errors.New("global ratelimit limiter was not initialized")
	}
	// Resolve the store before breaking the connection so the limiter keeps the
	// client that Purge closes instead of silently reconnecting on first use.
	if cache.Store("redis") == nil {
		return "", errors.New("redis limiter store is not available")
	}
	manager := frameworkredis.ManagerInstance()
	if manager == nil {
		return "", errors.New("redis manager is not initialized")
	}
	if _, err := frameworkredis.Connection("cache"); err != nil {
		return "", fmt.Errorf("resolve redis cache connection: %w", err)
	}
	if err := manager.Purge("cache"); err != nil {
		return "", fmt.Errorf("purge redis cache connection: %w", err)
	}

	_, err := limiter.Hit(context.Background(), "redis-errors", time.Minute)
	if err == nil {
		return "", errors.New("expected a Redis connection error, got nil")
	}
	return fmt.Sprintf("error=true closed=%t", strings.Contains(strings.ToLower(err.Error()), "closed")), nil
}
