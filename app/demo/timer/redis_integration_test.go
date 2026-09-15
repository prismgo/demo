package timerdemo_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/prismgo/framework/cache"

	demotest "prismgo-demo/app/demo/testing"
	"prismgo-demo/bootstrap"

	// Register the demo cache and Redis connection configuration for the application.
	_ "prismgo-demo/config"
)

// TestTimerDemoCrossProcessOverlap runs the overlap scenario against a real Redis
// cache so two schedulers share one distributed lock; it skips only when
// PRISMGO_REDIS_TEST_URL is absent.
func TestTimerDemoCrossProcessOverlap(t *testing.T) {
	redisURL := demotest.RequireIntegration(t, demotest.ServiceRedisURL)
	parsed, err := url.Parse(redisURL)
	if err != nil || parsed.Hostname() == "" || parsed.Port() == "" {
		t.Fatalf("parse PRISMGO_REDIS_TEST_URL = %q: parsed=%v, error=%v, want host and port", redisURL, parsed, err)
	}
	t.Setenv("REDIS_CACHE_URL", redisURL)
	t.Setenv("REDIS_HOST", parsed.Hostname())
	t.Setenv("REDIS_PORT", parsed.Port())
	t.Setenv("CACHE_STORE", "redis")
	t.Setenv("CACHE_PREFIX", fmt.Sprintf("prismgo_demo_timer_%d", time.Now().UnixNano()))
	t.Setenv("APP_ENV", "testing")
	t.Setenv("APP_KEY", "base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=")
	t.Setenv("DB_CONNECTION", "sqlite")
	t.Setenv("QUEUE_CONNECTION", "sync")
	t.Setenv("SESSION_DRIVER", "file")
	t.Setenv("CACHE_FILE_PATH", t.TempDir())
	t.Setenv("CACHE_FILE_LOCK_PATH", t.TempDir())

	app := bootstrap.NewApplication(t.TempDir())
	if err := app.Boot(); err != nil {
		t.Fatalf("boot real Redis timer demo application error = %v, want nil", err)
	}
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Errorf("close real Redis timer demo application error = %v, want nil", err)
		}
	})
	t.Cleanup(func() {
		if err := cache.Flush(context.Background()); err != nil {
			t.Errorf("flush isolated real Redis cache prefix error = %v, want nil", err)
		}
	})

	expectValue(t, "cross-process-overlap", "skipped=true ran-after-release=true")
}
