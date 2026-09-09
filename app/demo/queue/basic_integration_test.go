package queuedemo_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prismgo/framework/cache"
	"github.com/prismgo/framework/config"
	"github.com/prismgo/framework/container"
	"github.com/prismgo/framework/queue"
	"github.com/prismgo/framework/redis"
	goredis "github.com/redis/go-redis/v9"

	qdemo "prismgo-demo/app/demo/queue"
	"prismgo-demo/app/demo/testing"
)

func TestQueueDemoBasicWithRealRedis(t *testing.T) {
	url := demotest.RequireIntegration(t, demotest.ServiceRedisURL)
	options, err := goredis.ParseURL(url)
	if err != nil {
		t.Fatalf("parse Redis test URL: %v", err)
	}
	registry := container.NewContainer()
	container.SetProvider(func() *container.Container { return registry })
	t.Cleanup(func() { container.SetProvider(nil) })
	redisManager, err := redis.NewManager(redis.Config{
		DefaultName: "default",
		Connections: map[string]redis.ConnectionConfig{
			"default": {Name: "default", Addr: options.Addr, Username: options.Username, Password: options.Password, DB: options.DB},
		},
	})
	if err != nil {
		t.Fatalf("create Redis manager: %v", err)
	}
	if err := registry.Instance("redis", redisManager); err != nil {
		t.Fatalf("register Redis manager: %v", err)
	}
	installIntegrationCache(t, registry, cache.StoreConfig{Driver: "redis", Redis: cache.RedisConfig{Connection: "default"}})
	t.Cleanup(func() { _ = redisManager.Close(context.Background()) })

	manager, err := queue.NewManager(queue.Config{
		Default: "redis",
		Connections: map[string]queue.ConnectionConfig{
			"redis": {
				Driver: "redis", Queue: "demo-basic",
				Prefix:  fmt.Sprintf("prismgo_demo_%d", time.Now().UnixNano()),
				Options: map[string]any{"connection": "default"},
			},
		},
	}, queue.NewRegistry())
	if err != nil {
		t.Fatalf("create Redis queue manager: %v", err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	assertBasicIntegration(t, manager, "redis")
}

func TestQueueDemoBasicWithRealRabbitMQ(t *testing.T) {
	url := demotest.RequireIntegration(t, demotest.ServiceRabbitMQ)
	registry := container.NewContainer()
	container.SetProvider(func() *container.Container { return registry })
	t.Cleanup(func() { container.SetProvider(nil) })
	installIntegrationCache(t, registry, cache.StoreConfig{Driver: "memory"})
	slug := fmt.Sprintf("prismgo.demo.%d", time.Now().UnixNano())
	manager, err := queue.NewManager(queue.Config{
		Default: "rabbitmq",
		Connections: map[string]queue.ConnectionConfig{
			"rabbitmq": {
				Driver: "rabbitmq", Queue: slug + ".queue", BlockFor: 2 * time.Second,
				Options: map[string]any{
					"url": url, "exchange": slug + ".exchange", "declare": true,
					"exchange_durable": false, "queue_durable": false,
					"message_persistent": false, "auto_delete": true,
					"confirm": true, "delay_mode": "ttl_dlx",
					"restart_queue": slug + ".restart", "restart_enabled": false,
				},
			},
		},
	}, queue.NewRegistry())
	if err != nil {
		t.Fatalf("create RabbitMQ queue manager: %v", err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	assertBasicIntegration(t, manager, "rabbitmq")
}

func TestQueueDemoDebounce(t *testing.T) {
	t.Run("redis", func(t *testing.T) {
		manager := newRealRedisQueueManager(t)
		assertDebounceIntegration(t, manager, "redis")
	})
	t.Run("rabbitmq", func(t *testing.T) {
		manager := newRealRabbitMQQueueManager(t)
		assertDebounceIntegration(t, manager, "rabbitmq")
	})
}

func TestQueueDemoUniqueUntilProcessingWithRealRedis(t *testing.T) {
	manager := newRealRedisQueueManager(t)
	result, err := qdemo.Run(context.Background(), manager, "unique", "redis")
	if err != nil {
		t.Fatalf("run unique Redis scenario: %v", err)
	}
	for _, expected := range []string{"duplicate:rejected", "after-completion:accepted", "until-processing:accepted"} {
		if !hasIntegrationStep(result.Steps, expected) {
			t.Fatalf("missing %q in unique steps %v", expected, result.Steps)
		}
	}
}

func TestQueueDemoDispatch(t *testing.T) {
	for _, connection := range []string{"redis", "rabbitmq"} {
		t.Run(connection, func(t *testing.T) {
			manager := newRealQueueManager(t, connection)
			result, err := qdemo.Run(context.Background(), manager, "dispatch", connection)
			if err != nil {
				t.Fatalf("run dispatch scenario on %s: %v", connection, err)
			}
			if len(result.Steps) != 5 || result.Steps[0] != "basic:handled" {
				t.Fatalf("dispatch steps = %v, want five steps starting with basic", result.Steps)
			}
			for _, expected := range []string{"delay:handled", "later:handled", "delay-seconds:handled", "timeout:deadline"} {
				if !hasIntegrationStep(result.Steps, expected) {
					t.Fatalf("dispatch steps missing %q: %v", expected, result.Steps)
				}
			}
		})
	}
}

func TestQueueDemoChain(t *testing.T) {
	for _, connection := range []string{"redis", "rabbitmq"} {
		t.Run(connection, func(t *testing.T) {
			manager := newRealQueueManager(t, connection)
			result, err := qdemo.Run(context.Background(), manager, "chain", connection)
			if err != nil {
				t.Fatalf("run chain scenario on %s: %v", connection, err)
			}
			extract := integrationStepIndex(result.Steps, "success:extract")
			convert := integrationStepIndex(result.Steps, "success:convert")
			notify := integrationStepIndex(result.Steps, "success:notify")
			if extract < 0 || convert <= extract || notify <= convert {
				t.Fatalf("successful chain is out of order: %v", result.Steps)
			}
			if !hasIntegrationStep(result.Steps, "failure:first") || !hasIntegrationStep(result.Steps, "failure:stopped") || hasIntegrationStep(result.Steps, "failure:must-not-run") {
				t.Fatalf("failed chain did not stop at its first job: %v", result.Steps)
			}
		})
	}
}

func TestQueueDemoBatch(t *testing.T) {
	for _, connection := range []string{"redis", "rabbitmq"} {
		t.Run(connection, func(t *testing.T) {
			manager := newRealQueueManager(t, connection)
			result, err := qdemo.Run(context.Background(), manager, "batch", connection)
			if err != nil {
				t.Fatalf("run batch scenario on %s: %v", connection, err)
			}
			want := "success:one,failure:two,progress:2/2 failed=1,cancel:marked"
			if got := strings.Join(result.Steps, ","); got != want {
				t.Fatalf("batch steps = %q, want %q", got, want)
			}
		})
	}
}

func TestQueueDemoWorker(t *testing.T) {
	for _, connection := range []string{"redis", "rabbitmq"} {
		t.Run(connection, func(t *testing.T) {
			manager := newRealQueueManager(t, connection)
			result, err := qdemo.Run(context.Background(), manager, "worker", connection)
			if err != nil {
				t.Fatalf("run worker scenario on %s: %v", connection, err)
			}
			for _, expected := range []string{
				"once:first", "once:stopped", "empty:one", "empty:two", "empty:stopped",
				"max-jobs:one", "max-jobs:two", "max-jobs:stopped", "max-time:stopped",
				"priority:high", "priority:low", "timeout:cancelled",
			} {
				if !hasIntegrationStep(result.Steps, expected) {
					t.Fatalf("worker steps missing %q: %v", expected, result.Steps)
				}
			}
			for _, absent := range []string{"once:must-remain", "max-jobs:must-remain"} {
				if hasIntegrationStep(result.Steps, absent) {
					t.Fatalf("worker processed %q past its stop boundary: %v", absent, result.Steps)
				}
			}
			if integrationStepIndex(result.Steps, "priority:high") >= integrationStepIndex(result.Steps, "priority:low") {
				t.Fatalf("worker priority order is invalid: %v", result.Steps)
			}
		})
	}
}

func newRealQueueManager(t *testing.T, connection string) *queue.Manager {
	t.Helper()
	if connection == "redis" {
		return newRealRedisQueueManager(t)
	}
	return newRealRabbitMQQueueManager(t)
}

func assertDebounceIntegration(t *testing.T, manager *queue.Manager, connection string) {
	t.Helper()
	result, err := qdemo.Run(context.Background(), manager, "debounce", connection)
	if err != nil {
		t.Fatalf("run debounce scenario on %s: %v", connection, err)
	}
	if got := strings.Join(result.Steps, ","); got != "new:handled" {
		t.Fatalf("debounce steps = %q, want new:handled", got)
	}
	if !result.Processed || result.JobID == "" || !strings.HasPrefix(result.Queue, "demo-debounce-") {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func hasIntegrationStep(steps []string, expected string) bool {
	for _, step := range steps {
		if step == expected {
			return true
		}
	}
	return false
}

func integrationStepIndex(steps []string, expected string) int {
	for index, step := range steps {
		if step == expected {
			return index
		}
	}
	return -1
}

func installIntegrationCache(t *testing.T, registry *container.Container, store cache.StoreConfig) {
	t.Helper()
	manager, err := cache.NewManager(cache.Config{
		Default: "queue-demo",
		Prefix:  fmt.Sprintf("prismgo_demo_cache_%d", time.Now().UnixNano()),
		Stores:  map[string]cache.StoreConfig{"queue-demo": store},
	})
	if err != nil {
		t.Fatalf("create integration cache manager: %v", err)
	}
	if err := registry.Instance("config.default", config.New()); err != nil {
		t.Fatalf("register integration config: %v", err)
	}
	if err := registry.Instance("cache.manager", manager); err != nil {
		t.Fatalf("register integration cache manager: %v", err)
	}
	t.Cleanup(func() { _ = manager.Close() })
}

func newRealRedisQueueManager(t *testing.T) *queue.Manager {
	t.Helper()
	url := demotest.RequireIntegration(t, demotest.ServiceRedisURL)
	options, err := goredis.ParseURL(url)
	if err != nil {
		t.Fatalf("parse Redis test URL: %v", err)
	}
	registry := container.NewContainer()
	container.SetProvider(func() *container.Container { return registry })
	t.Cleanup(func() { container.SetProvider(nil) })
	redisManager, err := redis.NewManager(redis.Config{
		DefaultName: "default",
		Connections: map[string]redis.ConnectionConfig{
			"default": {Name: "default", Addr: options.Addr, Username: options.Username, Password: options.Password, DB: options.DB},
		},
	})
	if err != nil {
		t.Fatalf("create Redis manager: %v", err)
	}
	if err := registry.Instance("redis", redisManager); err != nil {
		t.Fatalf("register Redis manager: %v", err)
	}
	installIntegrationCache(t, registry, cache.StoreConfig{Driver: "redis", Redis: cache.RedisConfig{Connection: "default"}})
	t.Cleanup(func() { _ = redisManager.Close(context.Background()) })
	manager, err := queue.NewManager(queue.Config{
		Default: "redis",
		Connections: map[string]queue.ConnectionConfig{
			"redis": {Driver: "redis", Queue: "queue-demo", Prefix: fmt.Sprintf("prismgo_demo_%d", time.Now().UnixNano()), Options: map[string]any{"connection": "default"}},
		},
	}, queue.NewRegistry())
	if err != nil {
		t.Fatalf("create Redis queue manager: %v", err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	return manager
}

func newRealRabbitMQQueueManager(t *testing.T) *queue.Manager {
	t.Helper()
	url := demotest.RequireIntegration(t, demotest.ServiceRabbitMQ)
	registry := container.NewContainer()
	container.SetProvider(func() *container.Container { return registry })
	t.Cleanup(func() { container.SetProvider(nil) })
	installIntegrationCache(t, registry, cache.StoreConfig{Driver: "memory"})
	slug := fmt.Sprintf("prismgo.demo.%d", time.Now().UnixNano())
	manager, err := queue.NewManager(queue.Config{
		Default: "rabbitmq",
		Connections: map[string]queue.ConnectionConfig{
			"rabbitmq": {Driver: "rabbitmq", Queue: slug + ".queue", BlockFor: 2 * time.Second, Options: map[string]any{
				"url": url, "exchange": slug + ".exchange", "declare": true,
				"exchange_durable": false, "queue_durable": false, "message_persistent": false,
				"auto_delete": true, "confirm": true, "delay_mode": "ttl_dlx",
				"restart_queue": slug + ".restart", "restart_enabled": false,
			}},
		},
	}, queue.NewRegistry())
	if err != nil {
		t.Fatalf("create RabbitMQ queue manager: %v", err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	return manager
}

func assertBasicIntegration(t *testing.T, manager *queue.Manager, connection string) {
	t.Helper()
	result, err := qdemo.Run(context.Background(), manager, "basic", connection)
	if err != nil {
		t.Fatalf("run basic scenario on %s: %v", connection, err)
	}
	if result.JobID == "" || !result.Processed || result.Connection != connection || !strings.HasPrefix(result.Queue, "demo-basic-") {
		t.Fatalf("unexpected result: %#v", result)
	}
}
