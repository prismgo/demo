package ratelimitdemo_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"prismgo-demo/app/demo/catalog"
	ratelimitdemo "prismgo-demo/app/demo/ratelimit"
	demotest "prismgo-demo/app/demo/testing"
)

// expectValue executes one scenario and asserts its full observable value.
func expectValue(t *testing.T, name string, want string) {
	t.Helper()
	result, err := ratelimitdemo.Run(name)
	if err != nil {
		t.Fatalf("ratelimit demo %q error = %v, want nil", name, err)
	}
	if result.Case != name {
		t.Fatalf("ratelimit demo %q case = %q, want %q", name, result.Case, name)
	}
	if result.Value != want {
		t.Fatalf("ratelimit demo %q value = %q, want %q", name, result.Value, want)
	}
}

// newApplication boots an isolated application whose limiter uses the memory store.
func newApplication(t *testing.T) {
	t.Helper()
	t.Setenv("CACHE_LIMITER_DRIVER", "memory")
	demotest.NewApplication(t, demotest.Options{})
}

func TestRateLimitDemoArchitecture(t *testing.T) {
	expectValue(t, "architecture", "limiter=RateLimiter limit=Limit result=Result store=cache.Repository")
}

func TestRateLimitDemoConfiguration(t *testing.T) {
	expectValue(t, "config", "store=file driver=redis")
}

func TestRateLimitDemoStoreResolution(t *testing.T) {
	expectValue(t, "store-resolution", "dedicated-only=true fallback-default=true")
}

func TestRateLimitDemoMemoryStore(t *testing.T) {
	expectValue(t, "memory-store", "hits=1,2 reread=2 isolated=0")
}

func TestRateLimitDemoConfigurationRegistration(t *testing.T) {
	expectValue(t, "config-registration", "namespace=cache driver=demo-plugin stores=4")
}

func TestRateLimitDemoAutomaticInitialization(t *testing.T) {
	newApplication(t)
	expectValue(t, "auto-initialization", "resolved=true store=memory attempts=1")
}

func TestRateLimitDemoQuickStart(t *testing.T) {
	newApplication(t)
	expectValue(t, "quick-start", "allowed=5 blocked=1")
}

func TestRateLimitDemoExplicitLimiter(t *testing.T) {
	expectValue(t, "explicit-limiter", "manager=memory registered=true status=200")
}

func TestRateLimitDemoNamedRegistration(t *testing.T) {
	expectValue(t, "named-registration", "registered=true max=5 decay=1m0s key=ip:127.0.0.1")
}

func TestRateLimitDemoNamedLookup(t *testing.T) {
	expectValue(t, "named-lookup", "found=true missing=true blank=true")
}

func TestRateLimitDemoUnregisteredPassThrough(t *testing.T) {
	expectValue(t, "unregistered-pass-through", "resolver-nil=true status=200 body=open")
}

func TestRateLimitDemoMultipleRuleCounters(t *testing.T) {
	expectValue(t, "multi-rule-counters", "user=2 ip=2")
}

func TestRateLimitDemoMultipleRuleBlock(t *testing.T) {
	expectValue(t, "multi-rule-block", "first=200 second=429")
}

func TestRateLimitDemoLimitContract(t *testing.T) {
	expectValue(t, "limit-contract", "max=5 decay=1m0s key=user:1 fallback=user:fallback after=true response=true")
}

func TestRateLimitDemoEvery(t *testing.T) {
	expectValue(t, "every", "max=10 decay=30s")
}

func TestRateLimitDemoPerSecond(t *testing.T) {
	expectValue(t, "per-second", "max=2 decay=1s")
}

func TestRateLimitDemoPerMinute(t *testing.T) {
	expectValue(t, "per-minute", "max=3 decay=1m0s")
}

func TestRateLimitDemoPerMinutes(t *testing.T) {
	expectValue(t, "per-minutes", "max=20 decay=10m0s")
}

func TestRateLimitDemoPerHour(t *testing.T) {
	expectValue(t, "per-hour", "max=5 decay=1h0m0s")
}

func TestRateLimitDemoPerDay(t *testing.T) {
	expectValue(t, "per-day", "max=7 decay=24h0m0s")
}

func TestRateLimitDemoNone(t *testing.T) {
	expectValue(t, "none", "max=0 decay=0s")
}

func TestRateLimitDemoBy(t *testing.T) {
	expectValue(t, "by", "a=200,429 b=200")
}

func TestRateLimitDemoFallbackKey(t *testing.T) {
	expectValue(t, "fallback-key", "primary=1 fallback=1 first=200 second=429")
}

func TestRateLimitDemoAfterCount(t *testing.T) {
	newApplication(t)
	expectValue(t, "after-count", "first=500 second=500 third=429 attempts=2")
}

func TestRateLimitDemoAfterSkip(t *testing.T) {
	newApplication(t)
	expectValue(t, "after-skip", "ok=200 ok-attempts=0 server-error=500 error-attempts=1")
}

func TestRateLimitDemoAfterHeaders(t *testing.T) {
	newApplication(t)
	expectValue(t, "after-headers", "status=200 limit=3 remaining=3 attempts=1")
}

func TestRateLimitDemoCustomResponse(t *testing.T) {
	newApplication(t)
	expectValue(t, "custom-response", `first=200 second=429 body={"custom":true,"max":1}`)
}

func TestRateLimitDemoResultContract(t *testing.T) {
	expectValue(t, "result-contract", "limit=5 key=user:1 max=5 attempts=2 remaining=3 retry=42 reset=1700000000")
}

func TestRateLimitDemoThrottle(t *testing.T) {
	newApplication(t)
	expectValue(t, "throttle", "first=200 remaining=0 retry-after=true second=429")
}

func TestRateLimitDemoThrottleFor(t *testing.T) {
	expectValue(t, "throttle-for", "limiter-a=200,429 limiter-b=200,200")
}

func TestRateLimitDemoDisabledRule(t *testing.T) {
	expectValue(t, "disabled-rule", "first=200 second=200 has-limit-header=false")
}

func TestRateLimitDemoDefaultResponse(t *testing.T) {
	newApplication(t)
	expectValue(t, "default-response", `first=200 second=429 body={"type":"too_many_requests","title":"Too Many Requests","status":429,"detail":"too many requests","message":"too many requests"}`)
}

func TestRateLimitDemoSuccessHeaders(t *testing.T) {
	newApplication(t)
	expectValue(t, "success-headers", "status=200 limit=3 remaining=2")
}

func TestRateLimitDemoOverLimitHeaders(t *testing.T) {
	newApplication(t)
	expectValue(t, "over-limit-headers", "status=429 limit=1 remaining=0 retry-positive=true reset-positive=true")
}

func TestRateLimitDemoTightestHeaders(t *testing.T) {
	newApplication(t)
	expectValue(t, "tightest-headers", "limit=2 remaining=1")
}

func TestRateLimitDemoRouteGroup(t *testing.T) {
	newApplication(t)
	expectValue(t, "route-group", "a1=200 a2=200 b1=429")
}

func TestRateLimitDemoRouteCompatibility(t *testing.T) {
	newApplication(t)
	expectValue(t, "route-compatibility", "max=2 statuses=200,200,429")
}

func TestRateLimitDemoFacade(t *testing.T) {
	newApplication(t)
	expectValue(t, "facade", "hit=1 increment=3 attempts=3 remaining=2 retries=2 decrement=2 blocked=true")
}

func TestRateLimitDemoInstance(t *testing.T) {
	expectValue(t, "instance", "registered=true hit=1 increment=3 decrement=2 remaining=3 reset=0")
}

func TestRateLimitDemoHit(t *testing.T) {
	expectValue(t, "hit", "first=1 second=2 attempts=2")
}

func TestRateLimitDemoHitDefaultDecay(t *testing.T) {
	expectValue(t, "hit-default-decay", "zero=1m0s negative=1m0s")
}

func TestRateLimitDemoIncrement(t *testing.T) {
	expectValue(t, "increment", "value=5")
}

func TestRateLimitDemoIncrementDefault(t *testing.T) {
	expectValue(t, "increment-default", "first=1 second=2")
}

func TestRateLimitDemoDecrement(t *testing.T) {
	expectValue(t, "decrement", "before=5 after=3")
}

func TestRateLimitDemoDecrementDefault(t *testing.T) {
	expectValue(t, "decrement-default", "before=3 after=2")
}

func TestRateLimitDemoAttempts(t *testing.T) {
	expectValue(t, "attempts", "attempts=2")
}

func TestRateLimitDemoMissingAttempts(t *testing.T) {
	expectValue(t, "attempts-missing", "attempts=0")
}

func TestRateLimitDemoTooManyAttempts(t *testing.T) {
	expectValue(t, "too-many-attempts", "before=false after-one=false after-two=true")
}

func TestRateLimitDemoNonPositiveLimit(t *testing.T) {
	expectValue(t, "non-positive-limit", "zero-limit=false negative-limit=false")
}

func TestRateLimitDemoExpiredWindow(t *testing.T) {
	expectValue(t, "expired-window", "allowed=true attempts=0")
}

func TestRateLimitDemoRemaining(t *testing.T) {
	expectValue(t, "remaining", "remaining=3 retries=3")
}

func TestRateLimitDemoRemainingFloor(t *testing.T) {
	expectValue(t, "remaining-floor", "remaining=0")
}

func TestRateLimitDemoAvailableIn(t *testing.T) {
	expectValue(t, "available-in", "window=1m0s")
}

func TestRateLimitDemoAvailableInElapsed(t *testing.T) {
	expectValue(t, "available-in-elapsed", "elapsed=0")
}

func TestRateLimitDemoResetAttempts(t *testing.T) {
	expectValue(t, "reset-attempts", "before=3 after=0 timer=true")
}

func TestRateLimitDemoClear(t *testing.T) {
	expectValue(t, "clear", "before=3 after=0 timer=0")
}

func TestRateLimitDemoAttemptSuccess(t *testing.T) {
	expectValue(t, "attempt-success", "value=ok allowed=true attempts=1")
}

func TestRateLimitDemoAttemptBlocked(t *testing.T) {
	expectValue(t, "attempt-blocked", "value=<nil> allowed=false called=false attempts=1")
}

func TestRateLimitDemoAttemptError(t *testing.T) {
	expectValue(t, "attempt-error", "allowed=true error=true attempts=0")
}

func TestRateLimitDemoAttemptNil(t *testing.T) {
	expectValue(t, "attempt-nil", "allowed=false error=true")
}

func TestRateLimitDemoCleanKey(t *testing.T) {
	expectValue(t, "clean-key", "clean=user:42")
}

func TestRateLimitDemoHashedKey(t *testing.T) {
	sum := sha256.Sum256([]byte("user:1"))
	want := "hashed=ratelimit:api:" + hex.EncodeToString(sum[:])
	expectValue(t, "hashed-key", want)
}

func TestRateLimitDemoUnhashedKey(t *testing.T) {
	expectValue(t, "unhashed-key", "plain=ratelimit:api:user:1")
}

func TestRateLimitDemoManualKey(t *testing.T) {
	expectValue(t, "manual-key", "manual=1 scoped=0")
}

func TestRateLimitDemoKeyDesign(t *testing.T) {
	expectValue(t, "key-design", "business=orders:42 dimension=user:42 isolation=ratelimit:api:user:42")
}

func TestRateLimitDemoCacheLayout(t *testing.T) {
	expectValue(t, "cache-layout", "add=timer:orders increment=orders touch=orders ttl=2m0s")
}

func TestRateLimitDemoFixedWindow(t *testing.T) {
	expectValue(t, "fixed-window", "initial-blocked=false at-limit-blocked=true after-clear-blocked=false")
}

func TestRateLimitDemoMiddlewareKey(t *testing.T) {
	expectValue(t, "middleware-key", "plain=ratelimit:api:user:1 normalized=ratelimit:api:user:1")
}

func TestRateLimitDemoCacheErrors(t *testing.T) {
	expectValue(t, "cache-errors", "hit=true attempts=true too-many=true reset=true available=true")
}

func TestRateLimitDemoCounterTypeError(t *testing.T) {
	expectValue(t, "counter-type-error", "invalid=true")
}

func TestRateLimitDemoErrorPolicy(t *testing.T) {
	expectValue(t, "error-policy", "fail-open=true fail-closed=false")
}

func TestRateLimitDemoLaravelCompatibility(t *testing.T) {
	expectValue(t, "laravel-compatibility", "compatible=for,limit-builders,attempt,counters,headers,after,response boundary=sha256-hash,timer-prefix,no-artisan")
}

func TestRateLimitDemoRejectsUnknownScenario(t *testing.T) {
	_, err := ratelimitdemo.Run("unknown")
	if err == nil || err.Error() != `ratelimit demo unknown: unknown scenario "unknown"` {
		t.Fatalf("Run(unknown) error = %v, want unknown scenario error", err)
	}
}

func TestRateLimitDemoCatalogCoverage(t *testing.T) {
	entries := catalog.Filter("ratelimit", "", catalog.StatusImplemented)
	if len(entries) != 74 {
		t.Fatalf("implemented ratelimit entries = %d, want 74", len(entries))
	}
	for _, slug := range []string{"architecture", "store-resolution", "quick-start", "per-minute", "after-count", "throttle", "throttle-for", "hit", "cache-layout", "redis-errors", "laravel-compatibility"} {
		if item, ok := catalog.Find("ratelimit", slug); !ok || item.Status != catalog.StatusImplemented {
			t.Fatalf("ratelimit catalog entry %q = %#v, %v; want implemented", slug, item, ok)
		}
	}
}
