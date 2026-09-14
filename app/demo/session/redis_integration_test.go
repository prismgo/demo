package sessiondemo_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	sessiondemo "prismgo-demo/app/demo/session"
	demotest "prismgo-demo/app/demo/testing"
)

// TestSessionDemoRedisDriver executes the redis-driver scenario against a real
// Redis service; it skips only when PRISMGO_REDIS_TEST_URL is absent.
func TestSessionDemoRedisDriver(t *testing.T) {
	demotest.RequireIntegration(t, demotest.ServiceRedisURL)
	result, err := sessiondemo.RunRedis("redis-driver")
	if err != nil {
		t.Fatalf("session demo redis-driver error = %v, want nil", err)
	}
	want := regexp.MustCompile(`^shared=true prefix=prismgo_demo_session_\d{15,25} ttl=300s gc-keeps=true isolated=true corrupt-fresh=true$`)
	if !want.MatchString(result.Value) {
		t.Fatalf("session demo redis-driver value = %q, want shared/prefix/ttl/gc/isolated/corrupt observations", result.Value)
	}
	if result.Case != "redis-driver" {
		t.Fatalf("session demo redis-driver case = %q, want %q", result.Case, "redis-driver")
	}
}

// TestSessionDemoRedisLock executes the redis-lock scenario against a real Redis
// service; it skips only when PRISMGO_REDIS_TEST_URL is absent.
func TestSessionDemoRedisLock(t *testing.T) {
	demotest.RequireIntegration(t, demotest.ServiceRedisURL)
	result, err := sessiondemo.RunRedis("redis-lock")
	if err != nil {
		t.Fatalf("session demo redis-lock error = %v, want nil", err)
	}
	want := "contention=true released=true re-release-not-held=true expired-takeover=true stale-release-not-held=true"
	if result.Value != want {
		t.Fatalf("session demo redis-lock value = %q, want %q", result.Value, want)
	}
	if result.Case != "redis-lock" {
		t.Fatalf("session demo redis-lock case = %q, want %q", result.Case, "redis-lock")
	}
}

// TestSessionDemoRedisLockRequiresURL verifies the runner errors clearly without config.
func TestSessionDemoRedisLockRequiresURL(t *testing.T) {
	_, variable, ok := demotest.LookupIntegration(demotest.ServiceRedisURL)
	if ok {
		t.Skipf("%s is configured, cannot verify missing-URL error here", variable)
	}
	if _, err := sessiondemo.RunRedis("redis-lock"); err == nil || !strings.Contains(err.Error(), variable) {
		t.Fatalf("session demo redis-lock without URL error = %v, want %s requirement error", err, variable)
	}
}

// TestSessionDemoRedisDriverRequiresURL verifies the runner errors clearly without config.
func TestSessionDemoRedisDriverRequiresURL(t *testing.T) {
	_, variable, ok := demotest.LookupIntegration(demotest.ServiceRedisURL)
	if ok {
		t.Skipf("%s is configured, cannot verify missing-URL error here", variable)
	}
	if _, err := sessiondemo.RunRedis("redis-driver"); err == nil || !strings.Contains(err.Error(), variable) {
		t.Fatalf("session demo redis-driver without URL error = %v, want %s requirement error", err, variable)
	}
}

// TestSessionDemoRedisRejectsUnknownScenario keeps the redis dispatcher strict.
func TestSessionDemoRedisRejectsUnknownScenario(t *testing.T) {
	if _, err := sessiondemo.RunRedis("no-such-redis-case"); err == nil || !strings.Contains(err.Error(), fmt.Sprintf("unknown redis scenario %q", "no-such-redis-case")) {
		t.Fatalf("session demo unknown redis scenario error = %v, want unknown redis scenario error", err)
	}
}
