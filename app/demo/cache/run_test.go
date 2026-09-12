package cachedemo_test

import (
	"context"
	"strings"
	"testing"

	cachedemo "prismgo-demo/app/demo/cache"
	demotest "prismgo-demo/app/demo/testing"
)

func TestCacheDemoArchitecture(t *testing.T) {
	assertScenario(t, "architecture", "*cache.Manager -> memory")
}
func TestCacheDemoConfiguration(t *testing.T) {
	assertScenario(t, "config", "cache.default=memory; stores=4")
}
func TestCacheDemoDriverPrerequisites(t *testing.T) {
	assertScenario(t, "driver-prerequisites", "redis: named connection")
}
func TestCacheDemoTopLevelConfiguration(t *testing.T) {
	assertScenario(t, "top-level-config", "default=memory; encoding=inherited; prefix=prismgo_cache")
}
func TestCacheDemoMemoryConfiguration(t *testing.T) {
	assertScenario(t, "memory-config", "memory: driver=memory; prefix=memory")
}
func TestCacheDemoRedisConfiguration(t *testing.T) {
	assertScenario(t, "redis-config", "redis: driver=redis; prefix=redis; connection=cache")
}
func TestCacheDemoFileConfiguration(t *testing.T) {
	assertScenario(t, "file-config", "file: driver=file; prefix=file")
}
func TestCacheDemoFailoverConfiguration(t *testing.T) {
	assertScenario(t, "failover-config", "failover: driver=failover")
}
func TestCacheDemoLockConfiguration(t *testing.T) {
	assertScenario(t, "lock-config", "prefix=locks; retry_sleep_ms=50")
}
func TestCacheDemoFlexibleConfiguration(t *testing.T) {
	assertScenario(t, "flexible-config", "refresh_timeout=30 seconds")
}
func TestCacheDemoFacade(t *testing.T)       { assertScenario(t, "facade", "facade-value") }
func TestCacheDemoNamedStore(t *testing.T)   { assertScenario(t, "named-store", "file-value") }
func TestCacheDemoRepository(t *testing.T)   { assertScenario(t, "repository", "injected") }
func TestCacheDemoMissingStore(t *testing.T) { assertScenario(t, "missing-store", "ErrStoreNotFound") }
func TestCacheDemoGet(t *testing.T)          { assertScenario(t, "get", "stored") }
func TestCacheDemoFallbacks(t *testing.T) {
	assertScenario(t, "fallbacks", "default-value -> lazy-value -> stored")
}
func TestCacheDemoTypedRetrieval(t *testing.T) {
	assertScenario(t, "typed-retrieval", "hello/7/1.5/true")
}
func TestCacheDemoExistence(t *testing.T) {
	assertScenario(t, "existence", "missing_before=true; has_after=true")
}
func TestCacheDemoPutAndSet(t *testing.T) { assertScenario(t, "put", "second") }
func TestCacheDemoForever(t *testing.T)   { assertScenario(t, "forever", "permanent") }

func assertScenario(t *testing.T, name, want string) {
	t.Helper()
	setCacheDefaults(t)
	demotest.NewApplication(t, demotest.Options{})
	result, err := cachedemo.Run(context.Background(), name)
	if err != nil {
		t.Fatalf("run cache demo %q error = %v, want nil", name, err)
	}
	if result.Case != name || !strings.Contains(result.Value, want) {
		t.Fatalf("cache demo %q result = %#v, want case %q and value containing %q", name, result, name, want)
	}
}

func setCacheDefaults(t *testing.T) {
	t.Helper()
	for key, value := range map[string]string{
		"CACHE_PREFIX": "prismgo_cache", "CACHE_ENCODING": "",
		"CACHE_MEMORY_PREFIX": "memory", "CACHE_REDIS_PREFIX": "redis", "CACHE_REDIS_CONNECTION": "cache",
		"CACHE_FILE_PREFIX": "file", "CACHE_FAILOVER_STORES": "redis,memory",
		"CACHE_LOCK_PREFIX": "locks", "CACHE_LOCK_RETRY_SLEEP_MS": "50",
		"CACHE_FLEXIBLE_REFRESH_TIMEOUT": "30",
	} {
		t.Setenv(key, value)
	}
}

func TestCacheDemoConfigurationOverrides(t *testing.T) {
	setCacheDefaults(t)
	t.Setenv("CACHE_PREFIX", "tenant_cache")
	t.Setenv("CACHE_LOCK_RETRY_SLEEP_MS", "25")
	demotest.NewApplication(t, demotest.Options{})
	for _, test := range []struct {
		name string
		want string
	}{
		{name: "top-level-config", want: "prefix=tenant_cache"},
		{name: "lock-config", want: "retry_sleep_ms=25"},
	} {
		result, err := cachedemo.Run(context.Background(), test.name)
		if err != nil || !strings.Contains(result.Value, test.want) {
			t.Fatalf("override scenario %q result = %#v, error = %v, want %q", test.name, result, err, test.want)
		}
	}
}

func TestCacheDemoUnknownScenario(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	_, err := cachedemo.Run(context.Background(), "not-configured")
	if err == nil || !strings.Contains(err.Error(), `unknown scenario "not-configured"`) {
		t.Fatalf("unknown scenario error = %v, want descriptive error", err)
	}
}

func TestCacheDemoFallbackDoesNotWrite(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	result, err := cachedemo.Run(context.Background(), "fallbacks")
	if err != nil {
		t.Fatalf("run fallback demo error = %v, want nil", err)
	}
	if len(result.Details) != 2 || result.Details[0] != "lazy_calls:1" || result.Details[1] != "cached_after_fallback:false" {
		t.Fatalf("fallback details = %v, want lazy_calls:1 and cached_after_fallback:false", result.Details)
	}
}
