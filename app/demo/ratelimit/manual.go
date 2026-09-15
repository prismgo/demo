package ratelimitdemo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/prismgo/framework/cache"
	cachecontract "github.com/prismgo/framework/contracts/cache"
	"github.com/prismgo/framework/ratelimit"
)

// errControlled is returned by controlledRepository when failure injection is on.
var errControlled = errors.New("controlled cache failure")

// observationRepository records the limiter's cache writes and touch operations.
type observationRepository struct {
	cachecontract.Repository
	addKey   string
	addTTL   time.Duration
	addTTLs  map[string]time.Duration
	incKey   string
	touchKey string
}

// Add delegates and records the key and TTL of the timer write.
func (r *observationRepository) Add(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	r.addKey = key
	r.addTTL = ttl
	if r.addTTLs == nil {
		r.addTTLs = make(map[string]time.Duration)
	}
	r.addTTLs[key] = ttl
	return r.Repository.Add(ctx, key, value, ttl)
}

// Increment delegates and records the counter key.
func (r *observationRepository) Increment(ctx context.Context, key string, delta ...int64) (int64, error) {
	r.incKey = key
	return r.Repository.Increment(ctx, key, delta...)
}

// Touch delegates and records the touched counter key.
func (r *observationRepository) Touch(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	r.touchKey = key
	return r.Repository.Touch(ctx, key, ttl)
}

// controlledRepository injects deterministic timer state and cache failures.
type controlledRepository struct {
	cachecontract.Repository
	timerMissing bool
	timerPast    bool
	failAll      bool
}

// Add fails when failure injection is on, otherwise delegates.
func (r *controlledRepository) Add(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	if r.failAll {
		return false, errControlled
	}
	return r.Repository.Add(ctx, key, value, ttl)
}

// Increment fails when failure injection is on, otherwise delegates.
func (r *controlledRepository) Increment(ctx context.Context, key string, delta ...int64) (int64, error) {
	if r.failAll {
		return 0, errControlled
	}
	return r.Repository.Increment(ctx, key, delta...)
}

// Decrement fails when failure injection is on, otherwise delegates.
func (r *controlledRepository) Decrement(ctx context.Context, key string, delta ...int64) (int64, error) {
	if r.failAll {
		return 0, errControlled
	}
	return r.Repository.Decrement(ctx, key, delta...)
}

// Touch fails when failure injection is on, otherwise delegates.
func (r *controlledRepository) Touch(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if r.failAll {
		return false, errControlled
	}
	return r.Repository.Touch(ctx, key, ttl)
}

// Has reports a missing timer while timerMissing is set, otherwise delegates.
func (r *controlledRepository) Has(ctx context.Context, key string) (bool, error) {
	if r.failAll {
		return false, errControlled
	}
	if r.timerMissing && isTimerKey(key) {
		return false, nil
	}
	return r.Repository.Has(ctx, key)
}

// Get returns an elapsed timer while timerPast is set, otherwise delegates.
func (r *controlledRepository) Get(ctx context.Context, key string, fallback ...any) (any, error) {
	if r.failAll {
		return nil, errControlled
	}
	if r.timerPast && isTimerKey(key) {
		return time.Now().Add(-5 * time.Second).Unix(), nil
	}
	return r.Repository.Get(ctx, key, fallback...)
}

// Forget fails when failure injection is on, otherwise delegates.
func (r *controlledRepository) Forget(ctx context.Context, key string) error {
	if r.failAll {
		return errControlled
	}
	return r.Repository.Forget(ctx, key)
}

// isTimerKey reports whether a cache key is the limiter's window timer key.
func isTimerKey(key string) bool {
	return strings.HasPrefix(key, "timer:")
}

// hitScenario records two attempts and reads the counter back.
func hitScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	first, err := limiter.Hit(ctx, "hit", time.Minute)
	if err != nil {
		return "", fmt.Errorf("first hit: %w", err)
	}
	second, err := limiter.Hit(ctx, "hit", time.Minute)
	if err != nil {
		return "", fmt.Errorf("second hit: %w", err)
	}
	attempts, err := limiter.Attempts(ctx, "hit")
	if err != nil {
		return "", fmt.Errorf("read attempts: %w", err)
	}
	return fmt.Sprintf("first=%d second=%d attempts=%d", first, second, attempts), nil
}

// hitDefaultDecayScenario shows non-positive windows fall back to one minute.
func hitDefaultDecayScenario() (string, error) {
	manager, err := newMemoryManager()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	repo := &observationRepository{Repository: manager.Default()}
	limiter := ratelimit.New(repo)
	ctx := context.Background()

	for _, key := range []string{"zero", "negative"} {
		if _, err := limiter.Hit(ctx, key, 0); err != nil {
			return "", fmt.Errorf("hit %s with zero window: %w", key, err)
		}
	}
	if _, err := limiter.Hit(ctx, "negative", -time.Second); err != nil {
		return "", fmt.Errorf("hit negative with negative window: %w", err)
	}
	return fmt.Sprintf("zero=%s negative=%s", repo.addTTLs["timer:zero"], repo.addTTLs["timer:negative"]), nil
}

// incrementScenario increments a counter by a custom amount.
func incrementScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	value, err := limiter.Increment(context.Background(), "increment", time.Minute, 5)
	if err != nil {
		return "", fmt.Errorf("increment by five: %w", err)
	}
	return fmt.Sprintf("value=%d", value), nil
}

// incrementDefaultScenario increments a counter with the default amount.
func incrementDefaultScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	first, err := limiter.Increment(ctx, "increment-default", time.Minute)
	if err != nil {
		return "", fmt.Errorf("first increment: %w", err)
	}
	second, err := limiter.Increment(ctx, "increment-default", time.Minute)
	if err != nil {
		return "", fmt.Errorf("second increment: %w", err)
	}
	return fmt.Sprintf("first=%d second=%d", first, second), nil
}

// decrementScenario decrements a counter by a custom amount.
func decrementScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	if _, err := limiter.Increment(ctx, "decrement", time.Minute, 5); err != nil {
		return "", fmt.Errorf("seed counter: %w", err)
	}
	after, err := limiter.Decrement(ctx, "decrement", 2)
	if err != nil {
		return "", fmt.Errorf("decrement by two: %w", err)
	}
	return fmt.Sprintf("before=5 after=%d", after), nil
}

// decrementDefaultScenario decrements a counter with the default amount.
func decrementDefaultScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	if _, err := limiter.Increment(ctx, "decrement-default", time.Minute, 3); err != nil {
		return "", fmt.Errorf("seed counter: %w", err)
	}
	after, err := limiter.Decrement(ctx, "decrement-default")
	if err != nil {
		return "", fmt.Errorf("default decrement: %w", err)
	}
	return fmt.Sprintf("before=3 after=%d", after), nil
}

// attemptsScenario reads a populated attempt counter.
func attemptsScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	for range 2 {
		if _, err := limiter.Hit(ctx, "attempts", time.Minute); err != nil {
			return "", fmt.Errorf("hit attempts counter: %w", err)
		}
	}
	attempts, err := limiter.Attempts(ctx, "attempts")
	if err != nil {
		return "", fmt.Errorf("read attempts: %w", err)
	}
	return fmt.Sprintf("attempts=%d", attempts), nil
}

// missingAttemptsScenario reads a counter that was never written.
func missingAttemptsScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	attempts, err := limiter.Attempts(context.Background(), "missing")
	if err != nil {
		return "", fmt.Errorf("read missing attempts: %w", err)
	}
	return fmt.Sprintf("attempts=%d", attempts), nil
}

// tooManyAttemptsScenario detects the exact upper bound of a window.
func tooManyAttemptsScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	before, err := limiter.TooManyAttempts(ctx, "bound", 2)
	if err != nil {
		return "", fmt.Errorf("check before hits: %w", err)
	}
	if _, err := limiter.Hit(ctx, "bound", time.Minute); err != nil {
		return "", fmt.Errorf("first hit: %w", err)
	}
	afterOne, err := limiter.TooManyAttempts(ctx, "bound", 2)
	if err != nil {
		return "", fmt.Errorf("check after one hit: %w", err)
	}
	if _, err := limiter.Hit(ctx, "bound", time.Minute); err != nil {
		return "", fmt.Errorf("second hit: %w", err)
	}
	afterTwo, err := limiter.TooManyAttempts(ctx, "bound", 2)
	if err != nil {
		return "", fmt.Errorf("check after two hits: %w", err)
	}
	return fmt.Sprintf("before=%t after-one=%t after-two=%t", before, afterOne, afterTwo), nil
}

// nonPositiveLimitScenario shows non-positive limits disable counting entirely.
func nonPositiveLimitScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	for range 3 {
		if _, err := limiter.Hit(ctx, "disabled", time.Minute); err != nil {
			return "", fmt.Errorf("seed hits: %w", err)
		}
	}
	zero, err := limiter.TooManyAttempts(ctx, "disabled", 0)
	if err != nil {
		return "", fmt.Errorf("check zero limit: %w", err)
	}
	negative, err := limiter.TooManyAttempts(ctx, "disabled", -1)
	if err != nil {
		return "", fmt.Errorf("check negative limit: %w", err)
	}
	return fmt.Sprintf("zero-limit=%t negative-limit=%t", zero, negative), nil
}

// expiredWindowScenario shows an expired window resets the attempt counter.
func expiredWindowScenario() (string, error) {
	manager, err := newMemoryManager()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	repo := &controlledRepository{Repository: manager.Default(), timerMissing: true}
	limiter := ratelimit.New(repo)
	ctx := context.Background()
	for range 2 {
		if _, err := limiter.Hit(ctx, "expired", time.Minute); err != nil {
			return "", fmt.Errorf("hit expired window: %w", err)
		}
	}
	blocked, err := limiter.TooManyAttempts(ctx, "expired", 2)
	if err != nil {
		return "", fmt.Errorf("check expired window: %w", err)
	}
	attempts, err := limiter.Attempts(ctx, "expired")
	if err != nil {
		return "", fmt.Errorf("read attempts after expiry: %w", err)
	}
	return fmt.Sprintf("allowed=%t attempts=%d", !blocked, attempts), nil
}

// remainingScenario reports the remaining attempts for a partially used window.
func remainingScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	for range 2 {
		if _, err := limiter.Hit(ctx, "remaining", time.Minute); err != nil {
			return "", fmt.Errorf("seed hits: %w", err)
		}
	}
	remaining, err := limiter.Remaining(ctx, "remaining", 5)
	if err != nil {
		return "", fmt.Errorf("read remaining: %w", err)
	}
	retries, err := limiter.RetriesLeft(ctx, "remaining", 5)
	if err != nil {
		return "", fmt.Errorf("read retries left: %w", err)
	}
	return fmt.Sprintf("remaining=%d retries=%d", remaining, retries), nil
}

// remainingFloorScenario clamps remaining attempts at zero.
func remainingFloorScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	for range 5 {
		if _, err := limiter.Hit(ctx, "over", time.Minute); err != nil {
			return "", fmt.Errorf("seed hits: %w", err)
		}
	}
	remaining, err := limiter.Remaining(ctx, "over", 2)
	if err != nil {
		return "", fmt.Errorf("read floored remaining: %w", err)
	}
	return fmt.Sprintf("remaining=%d", remaining), nil
}

// availableInScenario reports the window length of a fresh limiter key.
func availableInScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	if _, err := limiter.Hit(ctx, "available", time.Minute); err != nil {
		return "", fmt.Errorf("hit available key: %w", err)
	}
	wait, err := limiter.AvailableIn(ctx, "available")
	if err != nil {
		return "", fmt.Errorf("read available in: %w", err)
	}
	if wait < 58 || wait > 60 {
		return "", fmt.Errorf("available in = %d, want 58..60 for a one minute window", wait)
	}
	return "window=1m0s", nil
}

// availableInElapsedScenario clamps an already elapsed window to zero.
func availableInElapsedScenario() (string, error) {
	manager, err := newMemoryManager()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	repo := &controlledRepository{Repository: manager.Default(), timerPast: true}
	limiter := ratelimit.New(repo)
	wait, err := limiter.AvailableIn(context.Background(), "elapsed")
	if err != nil {
		return "", fmt.Errorf("read elapsed available in: %w", err)
	}
	return fmt.Sprintf("elapsed=%d", wait), nil
}

// resetAttemptsScenario clears attempts while preserving the window timer.
func resetAttemptsScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	for range 3 {
		if _, err := limiter.Hit(ctx, "reset", time.Minute); err != nil {
			return "", fmt.Errorf("seed hits: %w", err)
		}
	}
	if err := limiter.ResetAttempts(ctx, "reset"); err != nil {
		return "", fmt.Errorf("reset attempts: %w", err)
	}
	attempts, err := limiter.Attempts(ctx, "reset")
	if err != nil {
		return "", fmt.Errorf("read attempts after reset: %w", err)
	}
	wait, err := limiter.AvailableIn(ctx, "reset")
	if err != nil {
		return "", fmt.Errorf("read timer after reset: %w", err)
	}
	return fmt.Sprintf("before=3 after=%d timer=%t", attempts, wait > 0), nil
}

// clearScenario clears both attempts and the window timer.
func clearScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	for range 3 {
		if _, err := limiter.Hit(ctx, "clear", time.Minute); err != nil {
			return "", fmt.Errorf("seed hits: %w", err)
		}
	}
	if err := limiter.Clear(ctx, "clear"); err != nil {
		return "", fmt.Errorf("clear key: %w", err)
	}
	attempts, err := limiter.Attempts(ctx, "clear")
	if err != nil {
		return "", fmt.Errorf("read attempts after clear: %w", err)
	}
	wait, err := limiter.AvailableIn(ctx, "clear")
	if err != nil {
		return "", fmt.Errorf("read timer after clear: %w", err)
	}
	return fmt.Sprintf("before=3 after=%d timer=%d", attempts, wait), nil
}

// attemptSuccessScenario runs a successful atomic attempt.
func attemptSuccessScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	value, allowed, err := limiter.Attempt(ctx, "attempt", 2, time.Minute, func(context.Context) (any, error) {
		return "ok", nil
	})
	if err != nil {
		return "", fmt.Errorf("run successful attempt: %w", err)
	}
	attempts, err := limiter.Attempts(ctx, "attempt")
	if err != nil {
		return "", fmt.Errorf("read attempts: %w", err)
	}
	return fmt.Sprintf("value=%v allowed=%t attempts=%d", value, allowed, attempts), nil
}

// attemptBlockedScenario skips the callback once the limit is reached.
func attemptBlockedScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	if _, err := limiter.Hit(ctx, "blocked", time.Minute); err != nil {
		return "", fmt.Errorf("seed blocked key: %w", err)
	}
	called := false
	value, allowed, err := limiter.Attempt(ctx, "blocked", 1, time.Minute, func(context.Context) (any, error) {
		called = true
		return "never", nil
	})
	if err != nil {
		return "", fmt.Errorf("run blocked attempt: %w", err)
	}
	attempts, err := limiter.Attempts(ctx, "blocked")
	if err != nil {
		return "", fmt.Errorf("read attempts: %w", err)
	}
	return fmt.Sprintf("value=%v allowed=%t called=%t attempts=%d", value, allowed, called, attempts), nil
}

// attemptErrorScenario shows a failed callback is not counted.
func attemptErrorScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	callbackErr := errors.New("callback failed")
	_, allowed, err := limiter.Attempt(context.Background(), "attempt-error", 2, time.Minute, func(context.Context) (any, error) {
		return nil, callbackErr
	})
	attempts, readErr := limiter.Attempts(context.Background(), "attempt-error")
	if readErr != nil {
		return "", fmt.Errorf("read attempts after callback error: %w", readErr)
	}
	return fmt.Sprintf("allowed=%t error=%t attempts=%d", allowed, errors.Is(err, callbackErr), attempts), nil
}

// attemptNilScenario rejects a nil atomic-attempt callback.
func attemptNilScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	_, allowed, err := limiter.Attempt(context.Background(), "attempt-nil", 2, time.Minute, nil)
	return fmt.Sprintf("allowed=%t error=%t", allowed, err != nil), nil
}

// newMemoryManager builds an isolated memory cache manager without an application.
func newMemoryManager() (*cache.Manager, error) {
	manager, err := cache.NewManager(cache.Config{
		Default: "memory",
		Stores:  map[string]cache.StoreConfig{"memory": {Driver: "memory"}},
	})
	if err != nil {
		return nil, fmt.Errorf("create memory cache manager: %w", err)
	}
	return manager, nil
}
