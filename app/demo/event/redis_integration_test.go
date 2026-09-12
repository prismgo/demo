package eventdemo_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prismgo/framework/queue"
	"github.com/prismgo/framework/redis"
	goredis "github.com/redis/go-redis/v9"

	eventdemo "prismgo-demo/app/demo/event"
	demotest "prismgo-demo/app/demo/testing"
)

func TestEventDemoQueuedRedis(t *testing.T) {
	testEventDemoRedisWorker(t, "queued-redis")
}

func TestEventDemoQueuedWorkerTesting(t *testing.T) {
	testEventDemoRedisWorker(t, "queued-worker-testing")
}

func testEventDemoRedisWorker(t *testing.T, scenario string) {
	t.Helper()
	url := demotest.RequireIntegration(t, demotest.ServiceRedisURL)
	options, err := goredis.ParseURL(url)
	if err != nil {
		t.Fatalf("parse Redis test URL: %v", err)
	}
	app := demotest.NewApplication(t, demotest.Options{})
	redisManager, err := redis.NewManager(redis.Config{DefaultName: "default", Connections: map[string]redis.ConnectionConfig{"default": {Name: "default", Addr: options.Addr, Username: options.Username, Password: options.Password, DB: options.DB}}})
	if err != nil {
		t.Fatalf("new Redis manager: %v", err)
	}
	t.Cleanup(func() {
		if err := redisManager.Close(context.Background()); err != nil {
			t.Errorf("close Redis manager: %v", err)
		}
	})
	if err := app.Container().Instance("redis", redisManager); err != nil {
		t.Fatalf("bind Redis manager: %v", err)
	}
	manager, err := queue.NewManager(queue.Config{Default: "redis", Connections: map[string]queue.ConnectionConfig{"redis": {Driver: "redis", Queue: "demo-event", Prefix: fmt.Sprintf("demo_event_%d", time.Now().UnixNano()), Options: map[string]any{"connection": "default"}}}}, queue.DefaultRegistry())
	if err != nil {
		t.Fatalf("new Redis queue manager: %v", err)
	}
	t.Cleanup(func() {
		if err := manager.Close(); err != nil {
			t.Errorf("close queue manager: %v", err)
		}
	})
	if err := app.Container().Instance("queue.manager", manager); err != nil {
		t.Fatalf("bind queue manager: %v", err)
	}
	if err := app.Container().Instance("queue.dispatcher", queue.NewDispatcher(manager)); err != nil {
		t.Fatalf("bind queue dispatcher: %v", err)
	}
	result, err := eventdemo.Run(context.Background(), scenario, "redis")
	if err != nil {
		t.Fatalf("run %s: %v", scenario, err)
	}
	if !strings.Contains(result.Value, "handled=1; event=demo.event.registered") {
		t.Fatalf("%s result = %#v, want worker handled registered event", scenario, result)
	}
}
