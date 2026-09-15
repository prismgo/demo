package ratelimitdemo

import (
	"context"
	"fmt"
	"time"

	"github.com/prismgo/framework/cache"
	cachecontract "github.com/prismgo/framework/contracts/cache"
	"github.com/prismgo/framework/ratelimit"
)

// storeResolutionScenario proves a dedicated limiter store stays separate from the default store.
func storeResolutionScenario() (string, error) {
	manager, err := cache.NewManager(cache.Config{
		Default: "shared",
		Prefix:  "demo",
		Stores: map[string]cache.StoreConfig{
			"shared":  {Driver: "memory", Prefix: "shared"},
			"limiter": {Driver: "memory", Prefix: "limiter"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("create two-store cache manager: %w", err)
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	dedicated := ratelimit.New(manager.Store("limiter"))
	if _, err := dedicated.Hit(ctx, "dedicated", time.Minute); err != nil {
		return "", fmt.Errorf("hit dedicated limiter: %w", err)
	}
	dedicatedOnly, err := storeHas(manager.Store("limiter"), ctx, "dedicated")
	if err != nil {
		return "", err
	}
	sharedHasDedicated, err := storeHas(manager.Store("shared"), ctx, "dedicated")
	if err != nil {
		return "", err
	}

	fallback := ratelimit.New(manager.Store(""))
	if _, err := fallback.Hit(ctx, "fallback", time.Minute); err != nil {
		return "", fmt.Errorf("hit fallback limiter: %w", err)
	}
	fallbackDefault, err := storeHas(manager.Store("shared"), ctx, "fallback")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("dedicated-only=%t fallback-default=%t",
		dedicatedOnly && !sharedHasDedicated, fallbackDefault), nil
}

// storeHas reports whether a repository stores the named key.
func storeHas(repo cachecontract.Repository, ctx context.Context, key string) (bool, error) {
	has, err := repo.Has(ctx, key)
	if err != nil {
		return false, fmt.Errorf("check store key %q: %w", key, err)
	}
	return has, nil
}

// memoryStoreScenario shows in-process memory state is shared by limiters and isolated per manager.
func memoryStoreScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	ctx := context.Background()
	first, err := limiter.Hit(ctx, "state", time.Minute)
	if err != nil {
		return "", fmt.Errorf("first hit: %w", err)
	}
	second, err := limiter.Hit(ctx, "state", time.Minute)
	if err != nil {
		return "", fmt.Errorf("second hit: %w", err)
	}
	reread, err := ratelimit.New(manager.Default()).Attempts(ctx, "state")
	if err != nil {
		return "", fmt.Errorf("reread attempts: %w", err)
	}

	otherManager, otherLimiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = otherManager.Close() }()
	isolated, err := otherLimiter.Attempts(ctx, "state")
	if err != nil {
		return "", fmt.Errorf("isolated attempts: %w", err)
	}

	return fmt.Sprintf("hits=%d,%d reread=%d isolated=%d", first, second, reread, isolated), nil
}
