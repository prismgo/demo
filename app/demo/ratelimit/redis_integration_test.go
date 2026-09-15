package ratelimitdemo_test

import (
	"net"
	"strconv"
	"strings"
	"testing"

	goredis "github.com/redis/go-redis/v9"

	ratelimitdemo "prismgo-demo/app/demo/ratelimit"
	demotest "prismgo-demo/app/demo/testing"
)

// TestRateLimitDemoRedisStore executes the redis-store scenario against a real
// Redis service; it skips only when PRISMGO_REDIS_TEST_URL is absent.
func TestRateLimitDemoRedisStore(t *testing.T) {
	startRedisApplication(t)
	expectValue(t, "redis-store", "store=redis exists=true attempts=1")
}

// TestRateLimitDemoRedisErrors executes the redis-errors scenario against a real
// Redis service; it skips only when PRISMGO_REDIS_TEST_URL is absent.
func TestRateLimitDemoRedisErrors(t *testing.T) {
	startRedisApplication(t)
	expectValue(t, "redis-errors", "error=true closed=true")
}

// TestRateLimitDemoRedisStoreRequiresURL verifies the scenario reports the missing
// Redis configuration instead of silently using the memory store.
func TestRateLimitDemoRedisStoreRequiresURL(t *testing.T) {
	if _, variable, ok := demotest.LookupIntegration(demotest.ServiceRedisURL); ok {
		t.Skipf("%s is configured, cannot verify missing-URL error here", variable)
	}
	t.Setenv("CACHE_LIMITER_DRIVER", "redis")
	demotest.NewApplication(t, demotest.Options{})
	if _, err := ratelimitdemo.Run("redis-store"); err == nil || !strings.Contains(err.Error(), "PRISMGO_REDIS_TEST_URL") {
		t.Fatalf("ratelimit redis-store without URL error = %v, want PRISMGO_REDIS_TEST_URL requirement", err)
	}
}

// startRedisApplication boots an isolated application whose limiter uses the real
// Redis connection described by PRISMGO_REDIS_TEST_URL. It skips when absent.
func startRedisApplication(t *testing.T) {
	t.Helper()
	rawURL := demotest.RequireIntegration(t, demotest.ServiceRedisURL)
	options, err := goredis.ParseURL(rawURL)
	if err != nil {
		t.Fatalf("parse PRISMGO_REDIS_TEST_URL %q: %v", rawURL, err)
	}
	host, port, err := net.SplitHostPort(options.Addr)
	if err != nil {
		t.Fatalf("split redis address %q: %v", options.Addr, err)
	}
	t.Setenv("CACHE_LIMITER_DRIVER", "redis")
	t.Setenv("REDIS_HOST", host)
	t.Setenv("REDIS_PORT", port)
	t.Setenv("REDIS_MAIN_DB", strconv.Itoa(options.DB))
	t.Setenv("REDIS_CACHE_DB", strconv.Itoa(options.DB))
	t.Setenv("REDIS_USERNAME", options.Username)
	t.Setenv("REDIS_PASSWORD", options.Password)
	demotest.NewApplication(t, demotest.Options{})
}
