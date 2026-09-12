package cachedemo_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/prismgo/framework/cache"
	"github.com/redis/go-redis/v9"

	cachedemo "prismgo-demo/app/demo/cache"
	"prismgo-demo/bootstrap"

	// Register the demo cache and Redis connection configuration for the application.
	_ "prismgo-demo/config"
)

func TestCacheDemoRealRedisIntegration(t *testing.T) {
	redisURL := os.Getenv("PRISMGO_REDIS_TEST_URL")
	if redisURL == "" {
		t.Skip("PRISMGO_REDIS_TEST_URL is required for real Redis integration")
	}
	parsed, err := url.Parse(redisURL)
	if err != nil || parsed.Hostname() == "" || parsed.Port() == "" {
		t.Fatalf("parse PRISMGO_REDIS_TEST_URL = %q: parsed=%v, error=%v, want host and port", redisURL, parsed, err)
	}
	t.Setenv("REDIS_CACHE_URL", redisURL)
	t.Setenv("REDIS_HOST", parsed.Hostname())
	t.Setenv("REDIS_PORT", parsed.Port())
	t.Setenv("CACHE_STORE", "redis")
	t.Setenv("CACHE_PREFIX", fmt.Sprintf("prismgo_demo_cache_%d", time.Now().UnixNano()))
	t.Setenv("APP_ENV", "testing")
	t.Setenv("APP_KEY", "base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=")
	t.Setenv("DB_CONNECTION", "sqlite")
	t.Setenv("QUEUE_CONNECTION", "sync")
	t.Setenv("SESSION_DRIVER", "file")
	t.Setenv("CACHE_FILE_PATH", t.TempDir())
	t.Setenv("CACHE_FILE_LOCK_PATH", t.TempDir())

	app := bootstrap.NewApplication(t.TempDir())
	if err := app.Boot(); err != nil {
		t.Fatalf("boot real Redis cache demo application error = %v, want nil", err)
	}
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Errorf("close real Redis cache demo application error = %v, want nil", err)
		}
	})
	if got := cache.DefaultName(); got != "redis" {
		t.Fatalf("default cache store = %q, want redis for live integration", got)
	}
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("parse real Redis URL = %q: %v", redisURL, err)
	}
	client := redis.NewClient(options)
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close real Redis client error = %v, want nil", err)
		}
	})
	outsideKey := fmt.Sprintf("prismgo_demo_cache_outside_%d", time.Now().UnixNano())
	if err := client.Set(context.Background(), outsideKey, "preserve", time.Minute).Err(); err != nil {
		t.Fatalf("seed outside-prefix Redis key error = %v, want nil", err)
	}
	t.Cleanup(func() {
		if err := client.Del(context.Background(), outsideKey).Err(); err != nil {
			t.Errorf("delete outside-prefix Redis key error = %v, want nil", err)
		}
	})
	t.Cleanup(func() {
		if err := cache.Flush(context.Background()); err != nil {
			t.Errorf("flush isolated real Redis cache prefix error = %v, want nil", err)
		}
	})

	wants := map[string]string{
		"add":                   "first=true; second=false; value=first",
		"put-many":              "alpha/beta",
		"remember":              "loaded/loaded; loader_calls=1",
		"remember-forever":      "permanent/permanent; loader_calls=1",
		"flexible":              "first=version-1; stale=version-1; refreshed=version-2",
		"touch":                 "updated=true; value=retained; missing=false",
		"many":                  "present=one; missing=fallback",
		"pull":                  "value=taken; remains=false",
		"forget":                "after_forget=false; after_delete=false",
		"forget-many":           `a=""; b=""`,
		"flush":                 "after_flush=false; after_clear=false",
		"counters":              "1 -> 5 -> 3",
		"lock":                  "acquired=true; contender=false; released=true",
		"lock-callback":         "acquired=true; callback=true; reacquired=true",
		"lock-block":            "timeout=true; acquired=true; callback=true",
		"lock-restore":          "restored=true; reacquired=true; force_released=true",
		"lock-flush":            "flushed=true; reacquired=true",
		"funnel":                "failure_callback=true; entered_after_release=true",
		"without-overlapping":   "overlap_blocked=true; entered=true; callback=true",
		"tags-memory":           "before=tagged; after_flush=false",
		"tags-redis":            "before=tagged; after_flush=false",
		"tags-unsupported":      "ErrTagsUnsupported",
		"memo":                  "first=first; memo=first; after_write=third",
		"failover":              "value=survived; fallback=survived",
		"custom-driver":         "value=custom-value; prefix=demo:tenant",
		"resource-lifecycle":    "manager_closed=true",
		"events":                "cache.writing,cache.written,cache.retrieving,cache.hit",
		"event-contract":        "event=cache.written; store=memory",
		"deferred":              "stale=refreshed; after_deferred=updated",
		"key-prefixes":          "cache=",
		"encoding":              "number=42; name=Ada",
		"errors":                "ErrCacheMiss,ErrStoreNotFound,ErrTagsUnsupported",
		"memory-capabilities":   "store=memory; touch=true; atomic=true; bulk=true; lock=true; tags=true",
		"file-capabilities":     "store=file; touch=true; atomic=true; bulk=true; lock=true; tags=false",
		"redis-capabilities":    "store=redis; touch=true; atomic=true; bulk=true; lock=true; tags=true",
		"failover-capabilities": "store=failover; touch=true; atomic=true; bulk=true; lock=true; tags=true",
		"laravel-compatibility": "Cache::store/get/put -> prismgo",
	}
	for _, name := range []string{
		"architecture", "config", "driver-prerequisites", "top-level-config",
		"memory-config", "redis-config", "file-config", "failover-config", "lock-config", "flexible-config",
		"facade", "named-store", "repository", "missing-store", "get", "fallbacks",
		"typed-retrieval", "existence", "put", "forever",
		"add", "put-many", "remember", "remember-forever", "flexible", "touch", "many", "pull",
		"forget", "forget-many", "flush", "counters", "lock", "lock-callback", "lock-block", "lock-restore",
		"lock-flush", "funnel", "without-overlapping", "tags-memory",
		"tags-redis", "tags-unsupported", "memo", "failover", "custom-driver", "resource-lifecycle",
		"events", "event-contract", "deferred", "key-prefixes", "encoding", "errors",
		"memory-capabilities", "file-capabilities", "redis-capabilities", "failover-capabilities", "laravel-compatibility",
	} {
		t.Run(name, func(t *testing.T) {
			result, err := cachedemo.Run(context.Background(), name)
			if err != nil {
				t.Fatalf("run %q against real Redis error = %v, want nil", name, err)
			}
			if result.Case != name || result.Value == "" {
				t.Fatalf("real Redis result = %#v, want case %q and nonempty value", result, name)
			}
			if want := wants[name]; want != "" && !strings.Contains(result.Value, want) {
				t.Fatalf("real Redis scenario %q value = %q, want to contain %q", name, result.Value, want)
			}
		})
	}
	if value, err := client.Get(context.Background(), outsideKey).Result(); err != nil || value != "preserve" {
		t.Fatalf("outside-prefix Redis key after cache flush = %q, error = %v, want preserve", value, err)
	}
}

func TestCacheDemoRedisTags(t *testing.T) {
	assertRedisScenario(t, "tags-redis", "before=tagged; after_flush=false")
}

func TestCacheDemoRedisCapabilities(t *testing.T) {
	assertRedisScenario(t, "redis-capabilities", "store=redis; touch=true; atomic=true; bulk=true; lock=true; tags=true")
}

func assertRedisScenario(t *testing.T, name, want string) {
	t.Helper()
	redisURL := os.Getenv("PRISMGO_REDIS_TEST_URL")
	if redisURL == "" {
		t.Skip("PRISMGO_REDIS_TEST_URL is required for real Redis integration")
	}
	parsed, err := url.Parse(redisURL)
	if err != nil || parsed.Hostname() == "" || parsed.Port() == "" {
		t.Fatalf("parse PRISMGO_REDIS_TEST_URL = %q: parsed=%v, error=%v, want host and port", redisURL, parsed, err)
	}
	t.Setenv("REDIS_CACHE_URL", redisURL)
	t.Setenv("REDIS_HOST", parsed.Hostname())
	t.Setenv("REDIS_PORT", parsed.Port())
	t.Setenv("CACHE_STORE", "redis")
	t.Setenv("CACHE_PREFIX", fmt.Sprintf("prismgo_demo_cache_%d", time.Now().UnixNano()))
	t.Setenv("APP_ENV", "testing")
	t.Setenv("APP_KEY", "base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=")
	t.Setenv("DB_CONNECTION", "sqlite")
	t.Setenv("QUEUE_CONNECTION", "sync")
	t.Setenv("SESSION_DRIVER", "file")
	t.Setenv("CACHE_FILE_PATH", t.TempDir())
	t.Setenv("CACHE_FILE_LOCK_PATH", t.TempDir())
	app := bootstrap.NewApplication(t.TempDir())
	if err := app.Boot(); err != nil {
		t.Fatalf("boot real Redis cache demo application error = %v, want nil", err)
	}
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Errorf("close real Redis cache demo application error = %v, want nil", err)
		}
	})
	t.Cleanup(func() {
		if err := cache.Flush(context.Background()); err != nil {
			t.Errorf("flush isolated real Redis cache prefix error = %v, want nil", err)
		}
	})
	result, err := cachedemo.Run(context.Background(), name)
	if err != nil || !strings.Contains(result.Value, want) {
		t.Fatalf("run real Redis scenario %q result = %#v, error = %v, want value containing %q", name, result, err, want)
	}
}
