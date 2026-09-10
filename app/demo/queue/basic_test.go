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

func TestQueueDemoStrategyEnvelope(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "strategy-envelope", "sync")
	if err != nil {
		t.Fatalf("run strategy envelope queue scenario: %v", err)
	}
	want := []string{
		"connection:option", "queue:option", "delay:2s", "limits:option",
		"retry:option", "tags:option", "fail-on-timeout:provider", "silenced:provider",
	}
	if got := strings.Join(result.Steps, ","); got != strings.Join(want, ",") {
		t.Fatalf("strategy envelope steps = %q, want %q", got, strings.Join(want, ","))
	}
	if !result.Processed || result.JobID == "" || !strings.HasPrefix(result.Queue, "demo-strategy-envelope-") {
		t.Fatalf("strategy envelope result = %#v, want processed result with job ID", result)
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

func TestQueueDemoUniqueDispatchOptions(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "unique-options", "sync")
	if err != nil {
		t.Fatalf("run unique dispatch options queue scenario: %v", err)
	}
	for _, want := range []string{"key:option", "ttl:45s", "via:memory", "until-processing:true", "duplicate:rejected"} {
		if !containsStep(result.Steps, want) {
			t.Fatalf("unique dispatch option steps = %v, want step %q", result.Steps, want)
		}
	}
	if !result.Processed || result.JobID == "" || !strings.HasPrefix(result.Queue, "demo-unique-options-") {
		t.Fatalf("unique dispatch options result = %#v, want processed result with job ID", result)
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

func TestQueueDemoOverlapReleasePolicy(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "overlap-release", "sync")
	if err != nil {
		t.Fatalf("run overlap release queue scenario: %v", err)
	}
	want := []string{"first:locked", "second:released-after-25s", "first:unlocked"}
	if got := strings.Join(result.Steps, ","); got != strings.Join(want, ",") {
		t.Fatalf("overlap release steps = %q, want %q", got, strings.Join(want, ","))
	}
	if !result.Processed || result.JobID != "" || result.Queue != "demo-overlap-release" {
		t.Fatalf("overlap release result = %#v, want processed middleware result", result)
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

func TestQueueDemoFailedCleanupCommandPaths(t *testing.T) {
	manager := newSyncQueueManager(t)
	registry := container.NewContainer()
	if err := registry.Instance("queue.manager", manager); err != nil {
		t.Fatalf("register sync queue manager: %v", err)
	}
	container.SetProvider(func() *container.Container { return registry })
	t.Cleanup(func() { container.SetProvider(nil) })

	result, err := Run(context.Background(), manager, "failed-command-paths", "sync")
	if err != nil {
		t.Fatalf("run failed cleanup command paths: %v", err)
	}
	want := []string{"queue:failed:listed", "queue:forget:deleted", "queue:flush:emptied"}
	if got := strings.Join(result.Steps, ","); got != strings.Join(want, ",") {
		t.Fatalf("failed cleanup command steps = %q, want %q", got, strings.Join(want, ","))
	}
	if !result.Processed || result.JobID != "job-one" || result.Queue != "demo-failed-cleanup" {
		t.Fatalf("failed cleanup command result = %#v, want processed job-one result", result)
	}
}

func TestQueueDemoDriverPrerequisites(t *testing.T) {
	assertConfigurationScenario(t, "driver-prerequisites", []string{
		"sync:none", "redis:PRISMGO_REDIS_TEST_URL", "rabbitmq:PRISMGO_RABBITMQ_TEST_URL",
	})
}

func TestQueueDemoConfiguration(t *testing.T) {
	assertConfigurationScenario(t, "config", []string{
		"default:sync", "connections:sync,redis,rabbitmq", "state:failed,batching,restart",
	})
}

func TestQueueDemoPayloadEncoding(t *testing.T) {
	assertConfigurationScenario(t, "payload-encoding", []string{"explicit:json", "inherited:msgpack"})
}

func TestQueueDemoSyncConnectionConfiguration(t *testing.T) {
	assertConfigurationScenario(t, "sync-connection", []string{"queue:configured", "sync:handled"})
}

func assertConfigurationScenario(t *testing.T, name string, want []string) {
	t.Helper()
	manager := newSyncQueueManager(t)
	result, err := Run(context.Background(), manager, name, "sync")
	if err != nil {
		t.Fatalf("run %s queue scenario: %v", name, err)
	}
	if got, expected := strings.Join(result.Steps, ","), strings.Join(want, ","); got != expected {
		t.Fatalf("%s steps = %q, want %q", name, got, expected)
	}
	if !result.Processed || result.Case != name || result.Connection != "sync" {
		t.Fatalf("%s result = %#v, want processed sync result", name, result)
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

func TestQueueDemoEncryptionMissingKey(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "encryption-missing-key", "sync")
	if err != nil {
		t.Fatalf("run missing encryption key queue scenario: %v", err)
	}
	want := []string{"app-key:rejected", "dispatch:rejected", "transport:empty"}
	if got := strings.Join(result.Steps, ","); got != strings.Join(want, ",") {
		t.Fatalf("missing encryption key steps = %q, want %q", got, strings.Join(want, ","))
	}
	if !result.Processed || result.JobID != "" || result.Queue != "demo-encryption-missing-key" {
		t.Fatalf("missing encryption key result = %#v, want processed rejection without job ID", result)
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

func TestQueueDemoCustomQueueContract(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "custom-queue-contract", "sync")
	if err != nil {
		t.Fatalf("run custom queue contract scenario: %v", err)
	}
	want := "push:accepted,later:accepted,bulk:accepted=2,pop:priority,clear:size=0,close:released"
	if got := strings.Join(result.Steps, ","); got != want {
		t.Fatalf("custom queue contract steps = %q, want %q", got, want)
	}
	if !result.Processed || result.JobID == "" || result.Queue != "high" {
		t.Fatalf("custom queue contract result = %#v, want processed high-priority job", result)
	}
}

func TestQueueDemoCustomReservedJob(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "custom-reserved-job", "sync")
	if err != nil {
		t.Fatalf("run custom reserved job scenario: %v", err)
	}
	want := "metadata:read,payload:copied,attempts:1,release:2s,attempts:2,delete:acknowledged"
	if got := strings.Join(result.Steps, ","); got != want {
		t.Fatalf("custom reserved job steps = %q, want %q", got, want)
	}
	if !result.Processed || result.JobID == "" || result.Queue != "jobs" {
		t.Fatalf("custom reserved job result = %#v, want processed jobs result", result)
	}
}

func TestQueueDemoCustomPopSession(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "custom-pop-session", "sync")
	if err != nil {
		t.Fatalf("run custom pop session scenario: %v", err)
	}
	want := "session:created,session:pop,job:handled,session:closed"
	if got := strings.Join(result.Steps, ","); got != want {
		t.Fatalf("custom pop session steps = %q, want %q", got, want)
	}
	if !result.Processed || result.JobID == "" || !strings.HasPrefix(result.Queue, "demo-custom-pop-session-") {
		t.Fatalf("custom pop session result = %#v, want processed result with job ID", result)
	}
}

func TestQueueDemoCustomConsumerIntent(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "custom-consumer-intent", "sync")
	if err != nil {
		t.Fatalf("run custom consumer intent scenario: %v", err)
	}
	want := "intent:acquired,queues:received,job:handled,intent:released"
	if got := strings.Join(result.Steps, ","); got != want {
		t.Fatalf("custom consumer intent steps = %q, want %q", got, want)
	}
	if !result.Processed || result.JobID == "" || !strings.HasPrefix(result.Queue, "demo-custom-consumer-intent-") {
		t.Fatalf("custom consumer intent result = %#v, want processed result with job ID", result)
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

func TestQueueDemoJobStateErrors(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "job-errors", "sync")
	if err != nil {
		t.Fatalf("run queue job state errors scenario: %v", err)
	}
	want := "duplicate:matched,skipped:matched,batch-cancelled:matched"
	if got := strings.Join(result.Steps, ","); got != want {
		t.Fatalf("job error steps = %q, want %q", got, want)
	}
	if !result.Processed || result.JobID != "" || result.Queue != "demo-job-errors" {
		t.Fatalf("job error result = %#v, want processed result without job ID", result)
	}
}

func TestQueueDemoConnectionErrors(t *testing.T) {
	manager := newSyncQueueManager(t)

	result, err := Run(context.Background(), manager, "connection-errors", "sync")
	if err != nil {
		t.Fatalf("run queue connection errors scenario: %v", err)
	}
	want := "connection-closed:matched,unsupported-operation:matched,unsupported-retry-after:matched"
	if got := strings.Join(result.Steps, ","); got != want {
		t.Fatalf("connection error steps = %q, want %q", got, want)
	}
	if !result.Processed || result.JobID != "" || result.Queue != "demo-connection-errors" {
		t.Fatalf("connection error result = %#v, want processed result without job ID", result)
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
