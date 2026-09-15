package ratelimitdemo

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prismgo/framework/ratelimit"
)

// facadeScenario drives the package-level facade against the running limiter.
func facadeScenario() (string, error) {
	ctx := context.Background()
	key := "demo-facade"
	if err := ratelimit.Clear(ctx, key); err != nil {
		return "", fmt.Errorf("clear facade key: %w", err)
	}
	hit, err := ratelimit.Hit(ctx, key, time.Minute)
	if err != nil {
		return "", fmt.Errorf("facade hit: %w", err)
	}
	incremented, err := ratelimit.Increment(ctx, key, time.Minute, 2)
	if err != nil {
		return "", fmt.Errorf("facade increment: %w", err)
	}
	attempts, err := ratelimit.Attempts(ctx, key)
	if err != nil {
		return "", fmt.Errorf("facade attempts: %w", err)
	}
	remaining, err := ratelimit.Remaining(ctx, key, 5)
	if err != nil {
		return "", fmt.Errorf("facade remaining: %w", err)
	}
	retries, err := ratelimit.RetriesLeft(ctx, key, 5)
	if err != nil {
		return "", fmt.Errorf("facade retries left: %w", err)
	}
	decremented, err := ratelimit.Decrement(ctx, key)
	if err != nil {
		return "", fmt.Errorf("facade decrement: %w", err)
	}
	blocked, err := ratelimit.TooManyAttempts(ctx, key, 2)
	if err != nil {
		return "", fmt.Errorf("facade too many attempts: %w", err)
	}
	return fmt.Sprintf("hit=%d increment=%d attempts=%d remaining=%d retries=%d decrement=%d blocked=%t",
		hit, incremented, attempts, remaining, retries, decremented, blocked), nil
}

// instanceScenario drives every documented method on an explicit limiter.
func instanceScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	limiter.For("instance", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(1)}
	})
	ctx := context.Background()
	hit, err := limiter.Hit(ctx, "instance-key", time.Minute)
	if err != nil {
		return "", fmt.Errorf("instance hit: %w", err)
	}
	incremented, err := limiter.Increment(ctx, "instance-key", time.Minute, 2)
	if err != nil {
		return "", fmt.Errorf("instance increment: %w", err)
	}
	decremented, err := limiter.Decrement(ctx, "instance-key", 1)
	if err != nil {
		return "", fmt.Errorf("instance decrement: %w", err)
	}
	remaining, err := limiter.Remaining(ctx, "instance-key", 5)
	if err != nil {
		return "", fmt.Errorf("instance remaining: %w", err)
	}
	if err := limiter.ResetAttempts(ctx, "instance-key"); err != nil {
		return "", fmt.Errorf("instance reset attempts: %w", err)
	}
	reset, err := limiter.Attempts(ctx, "instance-key")
	if err != nil {
		return "", fmt.Errorf("instance attempts after reset: %w", err)
	}
	return fmt.Sprintf("registered=%t hit=%d increment=%d decrement=%d remaining=%d reset=%d",
		limiter.Limiter("instance") != nil, hit, incremented, decremented, remaining, reset), nil
}
