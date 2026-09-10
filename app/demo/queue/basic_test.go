package queuedemo

import (
	"context"
	"strings"
	"testing"

	"github.com/prismgo/framework/cache"
	"github.com/prismgo/framework/config"
	"github.com/prismgo/framework/container"
	"github.com/prismgo/framework/queue"
)

func TestBasicScenarioDispatchesAndProcessesSyncJob(t *testing.T) {
	manager, err := queue.NewManager(queue.Config{
		Default: "sync",
		Connections: map[string]queue.ConnectionConfig{
			"sync": {Driver: "sync", Queue: "demo-basic"},
		},
	}, queue.NewRegistry())
	if err != nil {
		t.Fatalf("create sync queue manager: %v", err)
	}
	t.Cleanup(func() { _ = manager.Close() })

	result, err := Run(context.Background(), manager, "basic", "sync")
	if err != nil {
		t.Fatalf("run basic queue scenario: %v", err)
	}
	if result.JobID == "" || !result.Processed || result.Connection != "sync" || !strings.HasPrefix(result.Queue, "demo-basic-") {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestQueueDemoStrategies(t *testing.T) {
	manager, err := queue.NewManager(queue.Config{
		Default: "sync",
		Connections: map[string]queue.ConnectionConfig{
			"sync": {Driver: "sync", Queue: "default"},
		},
	}, queue.NewRegistry())
	if err != nil {
		t.Fatalf("create strategy queue manager: %v", err)
	}
	t.Cleanup(func() { _ = manager.Close() })

	result, err := Run(context.Background(), manager, "strategies", "sync")
	if err != nil {
		t.Fatalf("run strategy queue scenario: %v", err)
	}
	want := []string{"global:before", "job:before", "handle", "job:after", "global:after"}
	if strings.Join(result.Steps, ",") != strings.Join(want, ",") {
		t.Fatalf("strategy steps = %v, want %v", result.Steps, want)
	}
	if !result.Processed || result.JobID == "" || !strings.HasPrefix(result.Queue, "demo-strategies-") {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestQueueDemoUnique(t *testing.T) {
	installMemoryCache(t)
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "unique", "sync")
	if err != nil {
		t.Fatalf("run unique queue scenario: %v", err)
	}
	if !containsStep(result.Steps, "duplicate:rejected") || !containsStep(result.Steps, "after-completion:accepted") {
		t.Fatalf("unique steps = %v", result.Steps)
	}
	if !result.Processed || result.JobID == "" || !strings.HasPrefix(result.Queue, "demo-unique-") {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestQueueDemoMiddleware(t *testing.T) {
	installMemoryCache(t)
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "middleware", "sync")
	if err != nil {
		t.Fatalf("run middleware queue scenario: %v", err)
	}
	wantOrder := "global:before,job:before,order:handled,job:after,global:after"
	if got := strings.Join(result.Steps[:5], ","); got != wantOrder {
		t.Fatalf("middleware order = %q, want %q; all steps=%v", got, wantOrder, result.Steps)
	}
	for _, absent := range []string{"skip:handled", "custom-skip:handled", "rate:second:handled", "overlap:second:handled", "throttle:second:handled"} {
		if containsStep(result.Steps, absent) {
			t.Fatalf("middleware should prevent %q; steps=%v", absent, result.Steps)
		}
	}
	for _, present := range []string{"rate:first:handled", "overlap:first:handled", "throttle:first:handled"} {
		if !containsStep(result.Steps, present) {
			t.Fatalf("middleware should execute %q; steps=%v", present, result.Steps)
		}
	}
	if !result.Processed || result.JobID == "" || !strings.HasPrefix(result.Queue, "demo-middleware-") {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestQueueDemoDispatch(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "dispatch", "sync")
	if err != nil {
		t.Fatalf("run dispatch queue scenario: %v", err)
	}
	want := []string{
		"basic:handled", "delay:handled", "later:handled", "delay-seconds:handled",
		"timeout:no-deadline", "sync-error:returned",
	}
	if strings.Join(result.Steps, ",") != strings.Join(want, ",") {
		t.Fatalf("dispatch steps = %v, want %v", result.Steps, want)
	}
	if !result.Processed || result.JobID == "" || !strings.HasPrefix(result.Queue, "demo-dispatch-") {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestQueueDemoChain(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "chain", "sync")
	if err != nil {
		t.Fatalf("run chain queue scenario: %v", err)
	}
	want := []string{"success:extract", "success:convert", "success:notify", "failure:first", "failure:stopped"}
	if strings.Join(result.Steps, ",") != strings.Join(want, ",") {
		t.Fatalf("chain steps = %v, want %v", result.Steps, want)
	}
	if !result.Processed || result.JobID == "" || !strings.HasPrefix(result.Queue, "demo-chain-") {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestQueueDemoBatch(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "batch", "sync")
	if err != nil {
		t.Fatalf("run batch queue scenario: %v", err)
	}
	want := []string{"success:one", "success:two", "progress:2/2 failed=0", "cancel:handled-before-mark", "cancel:marked"}
	if strings.Join(result.Steps, ",") != strings.Join(want, ",") {
		t.Fatalf("batch steps = %v, want %v", result.Steps, want)
	}
	if !result.Processed || result.JobID == "" || !strings.HasPrefix(result.Queue, "demo-batch-") {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestQueueDemoFailedCommands(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "failed-commands", "sync")
	if err != nil {
		t.Fatalf("run failed commands queue scenario: %v", err)
	}
	want := []string{"queue:failed:list", "queue:failed:find", "queue:forget", "queue:flush"}
	if strings.Join(result.Steps, ",") != strings.Join(want, ",") {
		t.Fatalf("failed command steps = %v, want %v", result.Steps, want)
	}
	if !result.Processed || result.JobID != "job-one" || result.Queue != "demo-failed" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestQueueDemoEncryption(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "encryption", "sync")
	if err != nil {
		t.Fatalf("run encrypted payload queue scenario: %v", err)
	}
	want := []string{"provider:encrypted", "option:encrypted", "provider:handled", "option:handled"}
	if strings.Join(result.Steps, ",") != strings.Join(want, ",") {
		t.Fatalf("encryption steps = %v, want %v", result.Steps, want)
	}
	if !result.Processed || result.JobID == "" || !strings.HasPrefix(result.Queue, "demo-encryption-") {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestQueueDemoCustomDriver(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "custom-driver", "sync")
	if err != nil {
		t.Fatalf("run custom queue driver scenario: %v", err)
	}
	want := []string{"connector:resolved", "options:received", "custom:handled"}
	if strings.Join(result.Steps, ",") != strings.Join(want, ",") {
		t.Fatalf("custom driver steps = %v, want %v", result.Steps, want)
	}
	if !result.Processed || result.Connection != "custom" || result.JobID == "" || !strings.HasPrefix(result.Queue, "demo-custom-") {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestQueueDemoErrors(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "errors", "sync")
	if err != nil {
		t.Fatalf("run queue error constants scenario: %v", err)
	}
	want := []string{"empty:matched", "job-not-registered:matched", "unsupported-operation:matched", "manager-closed:matched"}
	if strings.Join(result.Steps, ",") != strings.Join(want, ",") {
		t.Fatalf("error steps = %v, want %v", result.Steps, want)
	}
	if !result.Processed || result.JobID != "" || result.Queue != "demo-errors" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func newSyncQueueManager(t *testing.T) *queue.Manager {
	t.Helper()
	manager, err := queue.NewManager(queue.Config{
		Default: "sync",
		Connections: map[string]queue.ConnectionConfig{
			"sync": {Driver: "sync", Queue: "default"},
		},
	}, queue.NewRegistry())
	if err != nil {
		t.Fatalf("create sync queue manager: %v", err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	return manager
}

func installMemoryCache(t *testing.T) {
	t.Helper()
	cacheManager, err := cache.NewManager(cache.Config{
		Default: "memory",
		Prefix:  "queue-demo-test",
		Stores: map[string]cache.StoreConfig{
			"memory": {Driver: "memory"},
		},
	})
	if err != nil {
		t.Fatalf("create memory cache manager: %v", err)
	}
	registry := container.NewContainer()
	if err := registry.Instance("config.default", config.New()); err != nil {
		t.Fatalf("register test config: %v", err)
	}
	if err := registry.Instance("cache.manager", cacheManager); err != nil {
		t.Fatalf("register cache manager: %v", err)
	}
	container.SetProvider(func() *container.Container { return registry })
	t.Cleanup(func() {
		container.SetProvider(nil)
		_ = cacheManager.Close()
	})
}

func containsStep(steps []string, expected string) bool {
	for _, step := range steps {
		if step == expected {
			return true
		}
	}
	return false
}
