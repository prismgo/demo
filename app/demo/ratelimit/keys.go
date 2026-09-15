package ratelimitdemo

import (
	"context"
	"fmt"
	"time"

	"github.com/prismgo/framework/ratelimit"
)

// cleanKeyScenario strips control characters and surrounding whitespace.
func cleanKeyScenario() (string, error) {
	cleaned := ratelimit.CleanRateLimiterKey(" user:\x01\x1f42\x7f\t ")
	return fmt.Sprintf("clean=%s", cleaned), nil
}

// hashedKeyScenario shows middleware dimension keys can be SHA-256 hashed.
func hashedKeyScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	limiter.ShouldHashKeys(true)
	return fmt.Sprintf("hashed=%s", limiter.MiddlewareKey("api", "user:1")), nil
}

// unhashedKeyScenario shows middleware dimension keys stay readable by default.
func unhashedKeyScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	return fmt.Sprintf("plain=%s", limiter.MiddlewareKey("api", "user:1")), nil
}

// manualKeyScenario shows manual counter keys bypass middleware namespacing.
func manualKeyScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	if _, err := limiter.Hit(ctx, "orders:42", time.Minute); err != nil {
		return "", fmt.Errorf("hit manual key: %w", err)
	}
	manual, err := limiter.Attempts(ctx, "orders:42")
	if err != nil {
		return "", fmt.Errorf("read manual attempts: %w", err)
	}
	scoped, err := limiter.Attempts(ctx, limiter.MiddlewareKey("api", "orders:42"))
	if err != nil {
		return "", fmt.Errorf("read scoped attempts: %w", err)
	}
	return fmt.Sprintf("manual=%d scoped=%d", manual, scoped), nil
}

// cacheLayoutScenario exposes the counter and timer keys the limiter writes.
func cacheLayoutScenario() (string, error) {
	manager, err := newMemoryManager()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	repo := &observationRepository{Repository: manager.Default()}
	limiter := ratelimit.New(repo)
	if _, err := limiter.Hit(context.Background(), "orders", 2*time.Minute); err != nil {
		return "", fmt.Errorf("hit layout key: %w", err)
	}
	return fmt.Sprintf("add=%s increment=%s touch=%s ttl=%s",
		repo.addKey, repo.incKey, repo.touchKey, repo.addTTL), nil
}

// fixedWindowScenario walks a window from fresh to limit to replenished.
func fixedWindowScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	fresh, err := limiter.TooManyAttempts(ctx, "window", 2)
	if err != nil {
		return "", fmt.Errorf("check fresh window: %w", err)
	}
	for range 2 {
		if _, err := limiter.Hit(ctx, "window", time.Minute); err != nil {
			return "", fmt.Errorf("seed window hits: %w", err)
		}
	}
	atLimit, err := limiter.TooManyAttempts(ctx, "window", 2)
	if err != nil {
		return "", fmt.Errorf("check exhausted window: %w", err)
	}
	if err := limiter.Clear(ctx, "window"); err != nil {
		return "", fmt.Errorf("clear window: %w", err)
	}
	afterClear, err := limiter.TooManyAttempts(ctx, "window", 2)
	if err != nil {
		return "", fmt.Errorf("check replenished window: %w", err)
	}
	return fmt.Sprintf("initial-blocked=%t at-limit-blocked=%t after-clear-blocked=%t", fresh, atLimit, afterClear), nil
}

// middlewareKeyScenario shows the middleware key namespace and normalization.
func middlewareKeyScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	plain := limiter.MiddlewareKey("api", "user:1")
	normalized := limiter.MiddlewareKey(" api ", " user:1 ")
	return fmt.Sprintf("plain=%s normalized=%s", plain, normalized), nil
}
