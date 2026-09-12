// Package cachedemo contains runnable examples of cache configuration and operations.
package cachedemo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prismgo/framework/cache"
	"github.com/prismgo/framework/config"
	cachecontract "github.com/prismgo/framework/contracts/cache"
)

// Result records an observable cache demo outcome.
type Result struct {
	Case    string   `json:"case"`
	Key     string   `json:"key,omitempty"`
	Value   string   `json:"value"`
	Details []string `json:"details,omitempty"`
}

// Run executes one documented cache scenario in the current application.
func Run(ctx context.Context, name string) (Result, error) {
	result := Result{Case: name}
	if value, ok, err := configurationResult(name); ok || err != nil {
		result.Value = value
		return result, err
	}
	result.Key = "demo:cache:" + name
	value, details, err := operationResult(ctx, name, result.Key)
	if err != nil {
		return Result{}, fmt.Errorf("cache demo %s: %w", name, err)
	}
	result.Value, result.Details = value, details
	return result, nil
}

func configurationResult(name string) (string, bool, error) {
	cfg := config.Resolve()
	if cfg == nil {
		return "", false, fmt.Errorf("cache demo: config is nil")
	}
	switch name {
	case "architecture":
		var factory cachecontract.Factory = cache.Resolve()
		return fmt.Sprintf("%T -> %s", factory, factory.Default().Name()), true, nil
	case "config":
		return fmt.Sprintf("cache.default=%s; stores=%d", cfg.GetString("cache.default"), len(cfg.GetStringMap("cache.stores"))), true, nil
	case "driver-prerequisites":
		connection := cfg.GetString("cache.stores.redis.connection")
		if connection == "" {
			return "", true, fmt.Errorf("cache demo: Redis store has no named connection")
		}
		return fmt.Sprintf("memory and file: local; redis: named connection %s; failover: ordered stores", connection), true, nil
	case "top-level-config":
		encoding := cfg.GetString("cache.encoding")
		if encoding == "" {
			encoding = "inherited"
		}
		return fmt.Sprintf("default=%s; encoding=%s; prefix=%s", cfg.GetString("cache.default"), encoding, cfg.GetString("cache.prefix")), true, nil
	case "memory-config", "redis-config", "file-config", "failover-config":
		store := map[string]string{"memory-config": "memory", "redis-config": "redis", "file-config": "file", "failover-config": "failover"}[name]
		fields := cfg.GetStringMap("cache.stores." + store)
		if len(fields) == 0 {
			return "", true, fmt.Errorf("cache demo: store %q is not configured", store)
		}
		return fmt.Sprintf("%s: driver=%v; prefix=%v; connection=%v; path=%v; children=%v", store, fields["driver"], fields["prefix"], fields["connection"], fields["path"], fields["stores"]), true, nil
	case "lock-config":
		return fmt.Sprintf("prefix=%s; retry_sleep_ms=%d", cfg.GetString("cache.lock.prefix"), cfg.GetInt("cache.lock.retry_sleep_ms")), true, nil
	case "flexible-config":
		return fmt.Sprintf("refresh_timeout=%d seconds", cfg.GetInt("cache.flexible.refresh_timeout")), true, nil
	default:
		return "", false, nil
	}
}

func operationResult(ctx context.Context, name, key string) (string, []string, error) {
	repo := cache.Default()
	switch name {
	case "facade":
		if err := cache.Put(ctx, key, "facade-value", time.Minute); err != nil {
			return "", nil, err
		}
		value, err := cache.Get[string](ctx, key)
		return value, []string{"store:" + cache.DefaultName()}, err
	case "named-store":
		if err := cache.PutFrom(ctx, "file", key, "file-value", time.Minute); err != nil {
			return "", nil, err
		}
		value, err := cache.GetFrom[string](ctx, "file", key)
		return value, []string{"store:" + cache.Store("file").Name()}, err
	case "repository":
		return repositoryExample(ctx, repo, key)
	case "missing-store":
		_, err := cache.Store("not-configured").Get(ctx, key)
		if !errors.Is(err, cache.ErrStoreNotFound) {
			return "", nil, fmt.Errorf("missing store error = %v, want ErrStoreNotFound", err)
		}
		return "ErrStoreNotFound", nil, nil
	case "get":
		if err := cache.Forget(ctx, key); err != nil {
			return "", nil, err
		}
		_, err := cache.Get[string](ctx, key)
		if !errors.Is(err, cache.ErrCacheMiss) {
			return "", nil, fmt.Errorf("cache miss error = %v, want ErrCacheMiss", err)
		}
		if err := cache.Put(ctx, key, "stored", time.Minute); err != nil {
			return "", nil, err
		}
		value, err := cache.Get[string](ctx, key)
		return value, []string{"initial:ErrCacheMiss"}, err
	case "fallbacks":
		if err := cache.Forget(ctx, key); err != nil {
			return "", nil, err
		}
		calls := 0
		fallback := cache.Lazy(func(context.Context) (string, error) {
			calls++
			return "lazy-value", nil
		})
		first, err := cache.Get[string](ctx, key, cache.Value("default-value"))
		if err != nil {
			return "", nil, err
		}
		second, err := cache.Get[string](ctx, key, fallback)
		if err != nil {
			return "", nil, err
		}
		wasCached, err := cache.Has(ctx, key)
		if err != nil {
			return "", nil, err
		}
		if err := cache.Put(ctx, key, "stored", time.Minute); err != nil {
			return "", nil, err
		}
		third, err := cache.Get[string](ctx, key, fallback)
		return fmt.Sprintf("%s -> %s -> %s", first, second, third), []string{fmt.Sprintf("lazy_calls:%d", calls), fmt.Sprintf("cached_after_fallback:%t", wasCached)}, err
	case "typed-retrieval":
		return typedRetrieval(ctx, key)
	case "existence":
		if err := cache.Forget(ctx, key); err != nil {
			return "", nil, err
		}
		before, err := cache.Missing(ctx, key)
		if err != nil {
			return "", nil, err
		}
		if err := cache.Put(ctx, key, "present", time.Minute); err != nil {
			return "", nil, err
		}
		after, err := cache.Has(ctx, key)
		return fmt.Sprintf("missing_before=%t; has_after=%t", before, after), nil, err
	case "put":
		if err := cache.Put(ctx, key, "first", time.Minute); err != nil {
			return "", nil, err
		}
		if err := cache.Set(ctx, key, "second", time.Minute); err != nil {
			return "", nil, err
		}
		value, err := cache.Get[string](ctx, key)
		return value, []string{"alias:Set"}, err
	case "forever":
		if err := cache.Forever(ctx, key, "permanent"); err != nil {
			return "", nil, err
		}
		value, err := cache.Get[string](ctx, key)
		return value, []string{"ttl:no-expiry"}, err
	default:
		return advancedOperationResult(ctx, name, key)
	}
}

func repositoryExample(ctx context.Context, repo cachecontract.Repository, key string) (string, []string, error) {
	if err := repo.Put(ctx, key, "injected", time.Minute); err != nil {
		return "", nil, err
	}
	value, err := repo.Get(ctx, key)
	return fmt.Sprint(value), []string{"store:" + repo.Name()}, err
}

func typedRetrieval(ctx context.Context, key string) (string, []string, error) {
	values := []struct {
		key   string
		value any
	}{{key + ":string", "hello"}, {key + ":int", 7}, {key + ":float", 1.5}, {key + ":bool", true}}
	for _, item := range values {
		if err := cache.Put(ctx, item.key, item.value, time.Minute); err != nil {
			return "", nil, err
		}
	}
	word, err := cache.String(ctx, key+":string")
	if err != nil {
		return "", nil, err
	}
	count, err := cache.Integer(ctx, key+":int")
	if err != nil {
		return "", nil, err
	}
	ratio, err := cache.Float(ctx, key+":float")
	if err != nil {
		return "", nil, err
	}
	active, err := cache.Boolean(ctx, key+":bool")
	return fmt.Sprintf("%s/%d/%.1f/%t", word, count, ratio, active), nil, err
}
