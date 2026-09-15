package ratelimitdemo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/prismgo/framework/cache"
	configpkg "github.com/prismgo/framework/config"
	"github.com/prismgo/framework/ratelimit"
)

// redisStoreScenario verifies the limiter stores its counters in a shared Redis store.
func redisStoreScenario() (string, error) {
	driver := strings.TrimSpace(configpkg.GetString("cache.limiter.driver", ""))
	if driver == "" {
		driver = cache.DefaultName()
	}
	if driver != "redis" {
		return "", fmt.Errorf("redis-store scenario requires CACHE_LIMITER_DRIVER=redis, got %q", driver)
	}
	rawURL := strings.TrimSpace(os.Getenv("PRISMGO_REDIS_TEST_URL"))
	if rawURL == "" {
		return "", errors.New("PRISMGO_REDIS_TEST_URL is required for the redis-store scenario")
	}
	options, err := goredis.ParseURL(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse PRISMGO_REDIS_TEST_URL: %w", err)
	}
	client := goredis.NewClient(options)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return "", fmt.Errorf("ping redis: %w", err)
	}

	repo := cache.Store(driver)
	limiter := ratelimit.Resolve()
	key := fmt.Sprintf("prismgo_demo_ratelimit_store_%d", time.Now().UnixNano())
	if _, err := limiter.Hit(ctx, key, time.Minute); err != nil {
		return "", fmt.Errorf("hit shared limiter: %w", err)
	}
	attempts, err := limiter.Attempts(ctx, key)
	if err != nil {
		return "", fmt.Errorf("read shared attempts: %w", err)
	}

	prefix := repo.GetStore().Prefix()
	counterKey := prefixedKey(prefix, key)
	timerKey := prefixedKey(prefix, "timer:"+key)
	exists := client.Exists(ctx, counterKey).Val() == 1
	_ = client.Del(ctx, counterKey, timerKey).Err()

	return fmt.Sprintf("store=%s exists=%t attempts=%d", driver, exists, attempts), nil
}

// prefixedKey joins a cache prefix and a limiter key the way Repository does.
func prefixedKey(prefix, key string) string {
	if strings.TrimSpace(prefix) == "" {
		return key
	}
	return prefix + ":" + key
}
