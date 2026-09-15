package redisdemo

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/prismgo/framework/cache"
)

// openCacheLive boots an isolated application whose default cache store is Redis,
// using a unique global prefix so scenarios never collide.
func openCacheLive() (prefix string, cleanup func(), err error) {
	prefix = fmt.Sprintf("prismgo_demo_redis_cache_%d", time.Now().UnixNano())
	_, base, err := openLiveWith(liveOptions{env: map[string]string{
		"CACHE_STORE":        "redis",
		"CACHE_PREFIX":       prefix,
		"CACHE_REDIS_PREFIX": "redis",
	}})
	if err != nil {
		return "", nil, err
	}
	cleanup = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = cache.Flush(ctx)
		base()
	}
	return prefix, cleanup, nil
}

// cacheDriverScenario proves the cache Redis driver shares the configured connection.
func cacheDriverScenario() (string, error) {
	prefix, cleanup, err := openCacheLive()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	repo := cache.Store("redis")
	if err := repo.Put(ctx, "driver", "hello", time.Minute); err != nil {
		return "", fmt.Errorf("put via redis driver: %w", err)
	}
	value, err := repo.Get(ctx, "driver")
	if err != nil {
		return "", fmt.Errorf("get via redis driver: %w", err)
	}
	client, err := redisConnectionClient("cache")
	if err != nil {
		return "", err
	}
	exists, err := client.Exists(ctx, prefix+":redis:driver").Result()
	if err != nil {
		return "", fmt.Errorf("check redis key: %w", err)
	}
	if exists != 1 {
		return "", fmt.Errorf("redis key exists = %d, want 1", exists)
	}
	return fmt.Sprintf("store=%s value=%v shared-key=true", repo.Name(), value), nil
}

// cacheBasicScenario exercises the documented basic cache reads and writes.
func cacheBasicScenario() (string, error) {
	_, cleanup, err := openCacheLive()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := cache.Put(ctx, "basic", "value", time.Minute); err != nil {
		return "", fmt.Errorf("put basic: %w", err)
	}
	value, err := cache.String(ctx, "basic")
	if err != nil {
		return "", fmt.Errorf("get basic: %w", err)
	}
	has, err := cache.Has(ctx, "basic")
	if err != nil {
		return "", fmt.Errorf("has basic: %w", err)
	}
	if err := cache.Forget(ctx, "basic"); err != nil {
		return "", fmt.Errorf("forget basic: %w", err)
	}
	missing, err := cache.Missing(ctx, "basic")
	if err != nil {
		return "", fmt.Errorf("missing basic: %w", err)
	}
	return fmt.Sprintf("get=%s has=%t missing=%t", value, has, missing), nil
}

// cacheTTLScenario exercises Redis-backed TTL management through cache Touch.
func cacheTTLScenario() (string, error) {
	prefix, cleanup, err := openCacheLive()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := cache.Put(ctx, "ttl", "value", 2*time.Minute); err != nil {
		return "", fmt.Errorf("put ttl: %w", err)
	}
	client, err := redisConnectionClient("cache")
	if err != nil {
		return "", err
	}
	key := prefix + ":redis:ttl"
	initial, err := client.Do(ctx, "ttl", key).Int64()
	if err != nil {
		return "", fmt.Errorf("read initial ttl: %w", err)
	}
	if initial <= 0 {
		return "", fmt.Errorf("initial ttl = %d, want positive", initial)
	}
	if _, err := cache.Touch(ctx, "ttl", 0); err != nil {
		return "", fmt.Errorf("persist ttl key: %w", err)
	}
	persisted, err := client.Do(ctx, "ttl", key).Int64()
	if err != nil {
		return "", fmt.Errorf("read persisted ttl: %w", err)
	}
	if _, err := cache.Touch(ctx, "ttl", time.Minute); err != nil {
		return "", fmt.Errorf("re-touch ttl key: %w", err)
	}
	retouched, err := client.Do(ctx, "ttl", key).Int64()
	if err != nil {
		return "", fmt.Errorf("read retouched ttl: %w", err)
	}
	return fmt.Sprintf("initial-positive=%t persisted=%t retouched-positive=%t",
		initial > 0, persisted == -1, retouched > 0), nil
}

// cacheAtomicScenario exercises Redis-backed atomic cache operations.
func cacheAtomicScenario() (string, error) {
	_, cleanup, err := openCacheLive()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	first, err := cache.Increment(ctx, "counter")
	if err != nil {
		return "", fmt.Errorf("increment: %w", err)
	}
	second, err := cache.Increment(ctx, "counter", 4)
	if err != nil {
		return "", fmt.Errorf("increment by 4: %w", err)
	}
	third, err := cache.Decrement(ctx, "counter", 2)
	if err != nil {
		return "", fmt.Errorf("decrement by 2: %w", err)
	}
	added, err := cache.Add(ctx, "atomic-add", "1", time.Minute)
	if err != nil {
		return "", fmt.Errorf("atomic add: %w", err)
	}
	existing, err := cache.Add(ctx, "atomic-add", "2", time.Minute)
	if err != nil {
		return "", fmt.Errorf("atomic add existing: %w", err)
	}
	pulled, err := cache.Pull[int](ctx, "counter")
	if err != nil {
		return "", fmt.Errorf("pull counter: %w", err)
	}
	return fmt.Sprintf("increment=%d increment-by=%d decrement=%d add-new=%t add-existing=%t pulled=%d",
		first, second, third, added, existing, pulled), nil
}

// cacheBulkScenario exercises Redis-backed bulk cache operations.
func cacheBulkScenario() (string, error) {
	_, cleanup, err := openCacheLive()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := cache.PutMany(ctx, map[string]string{"a": "1", "b": "2", "c": "3"}, time.Minute); err != nil {
		return "", fmt.Errorf("put many: %w", err)
	}
	values, err := cache.Many[string](ctx, []string{"a", "b", "c"})
	if err != nil {
		return "", fmt.Errorf("get many: %w", err)
	}
	pairs := make([]string, 0, len(values))
	for key, value := range values {
		pairs = append(pairs, key+"="+value)
	}
	sort.Strings(pairs)
	if err := cache.ForgetMany(ctx, []string{"a", "b"}); err != nil {
		return "", fmt.Errorf("forget many: %w", err)
	}
	remaining, err := cache.Many[string](ctx, []string{"a", "b", "c"})
	if err != nil {
		return "", fmt.Errorf("get many after forget: %w", err)
	}
	return fmt.Sprintf("bulk=%s forgotten=%t/%t remaining=%s",
		strings.Join(pairs, ","), remaining["a"] == "", remaining["b"] == "", remaining["c"]), nil
}

// cacheTagsScenario exercises Redis-backed tagged cache operations.
func cacheTagsScenario() (string, error) {
	_, cleanup, err := openCacheLive()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tagged := cache.Tags("orders")
	if err := tagged.Put(ctx, "tagged", "value", time.Minute); err != nil {
		return "", fmt.Errorf("put tagged: %w", err)
	}
	value, err := tagged.Get(ctx, "tagged")
	if err != nil {
		return "", fmt.Errorf("get tagged: %w", err)
	}
	if err := tagged.Flush(ctx); err != nil {
		return "", fmt.Errorf("flush tagged: %w", err)
	}
	has, err := tagged.Has(ctx, "tagged")
	if err != nil {
		return "", fmt.Errorf("has tagged after flush: %w", err)
	}
	return fmt.Sprintf("get=%v flushed=%t", value, !has), nil
}

// cacheFlushScenario verifies prefix-aware Redis flush removes only cached keys.
func cacheFlushScenario() (string, error) {
	prefix, cleanup, err := openCacheLive()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := cache.Put(ctx, "flush-a", "1", time.Minute); err != nil {
		return "", fmt.Errorf("put flush-a: %w", err)
	}
	if err := cache.Put(ctx, "flush-b", "2", time.Minute); err != nil {
		return "", fmt.Errorf("put flush-b: %w", err)
	}
	client, err := redisConnectionClient("cache")
	if err != nil {
		return "", err
	}
	before, err := client.Keys(ctx, prefix+":*").Result()
	if err != nil {
		return "", fmt.Errorf("count keys before flush: %w", err)
	}
	if err := cache.Flush(ctx); err != nil {
		return "", fmt.Errorf("flush cache: %w", err)
	}
	after, err := client.Keys(ctx, prefix+":*").Result()
	if err != nil {
		return "", fmt.Errorf("count keys after flush: %w", err)
	}
	return fmt.Sprintf("before=%d after=%d", len(before), len(after)), nil
}
