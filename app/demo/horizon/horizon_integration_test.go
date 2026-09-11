package horizondemo_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	qcontract "github.com/prismgo/framework/contracts/queue"
	"github.com/prismgo/framework/event"
	"github.com/prismgo/framework/foundation"
	"github.com/prismgo/framework/kernel"
	"github.com/prismgo/framework/queue"
	prismredis "github.com/prismgo/framework/redis"
	"github.com/prismgo/horizon"
	"github.com/prismgo/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	goredis "github.com/redis/go-redis/v9"

	demotest "prismgo-demo/app/demo/testing"
)

var redisJobHits atomic.Int64
var rabbitMQJobHits atomic.Int64

type liveQueueJob struct {
	Connection string
}

func (j *liveQueueJob) Handle(context.Context) error {
	switch j.Connection {
	case "redis":
		redisJobHits.Add(1)
	case "rabbitmq":
		rabbitMQJobHits.Add(1)
	default:
		return fmt.Errorf("unknown queue connection %q", j.Connection)
	}
	return nil
}

type liveStoreFactory struct {
	store horizon.Store
}

func (f liveStoreFactory) ResolveStore(context.Context, horizon.Config) (horizon.Store, error) {
	return f.store, nil
}

func TestHorizonWithRealQueue(t *testing.T) {
	redisURL := demotest.RequireIntegration(t, demotest.ServiceRedisURL)
	rabbitMQURL := demotest.RequireIntegration(t, demotest.ServiceRabbitMQ)
	redisOptions, err := goredis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("parse Redis integration URL: %v", err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	redisJobHits.Store(0)
	rabbitMQJobHits.Store(0)

	namespace := fmt.Sprintf("prismgo.demo.horizon.%d", time.Now().UnixNano())
	resources := liveResources{
		horizonPrefix:    namespace + ".store",
		redisQueuePrefix: namespace + ".queue",
		redisQueue:       namespace + ".redis",
		rabbitMQQueue:    namespace + ".rabbitmq",
		rabbitMQExchange: namespace + ".exchange",
		restartQueue:     namespace + ".restart",
	}

	cleanupClient := goredis.NewClient(redisOptions)
	if err := cleanupClient.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping real Redis at %s: %v", redisOptions.Addr, err)
	}
	t.Cleanup(func() {
		cleanupLiveRedis(t, cleanupClient, resources.horizonPrefix+":*", resources.redisQueuePrefix+":*")
		if err := cleanupClient.Close(); err != nil {
			t.Errorf("close Redis cleanup client: %v", err)
		}
		cleanupLiveRabbitMQ(t, rabbitMQURL, resources)
	})

	app := foundation.NewApplication()
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Errorf("close Horizon demo application: %v", err)
		}
	})
	redisManager, err := prismredis.NewManager(prismredis.Config{
		DefaultName: "default",
		Connections: map[string]prismredis.ConnectionConfig{
			"default": {
				Name:     "default",
				Addr:     redisOptions.Addr,
				Username: redisOptions.Username,
				Password: redisOptions.Password,
				DB:       redisOptions.DB,
			},
		},
	})
	if err != nil {
		t.Fatalf("create real Redis manager: %v", err)
	}
	if err := app.Instance("redis", redisManager); err != nil {
		t.Fatalf("bind real Redis manager: %v", err)
	}
	t.Cleanup(func() {
		if err := redisManager.Close(context.Background()); err != nil {
			t.Errorf("close real Redis manager: %v", err)
		}
	})

	queueManager := newLiveQueueManager(t, rabbitMQURL, resources)
	t.Cleanup(func() {
		if err := queueManager.Close(); err != nil {
			t.Errorf("close live queue manager: %v", err)
		}
	})
	bus := event.New()
	if err := app.Instance("queue.manager", queueManager); err != nil {
		t.Fatalf("bind live queue manager: %v", err)
	}
	if err := app.Instance("event.dispatcher", bus); err != nil {
		t.Fatalf("bind Horizon event dispatcher: %v", err)
	}
	if err := (queue.ServiceProvider{}).Boot(app); err != nil {
		t.Fatalf("boot queue event bridge: %v", err)
	}
	if err := (rabbitmq.ServiceProvider{}).Boot(app); err != nil {
		t.Fatalf("boot RabbitMQ extension: %v", err)
	}
	t.Cleanup(func() { queue.UseEventSink(nil) })

	store, err := horizon.NewRedisStore(
		horizon.RedisOptions{Connection: "default"},
		horizon.StoreOptions{Prefix: resources.horizonPrefix, HeartbeatTTL: time.Minute},
	)
	if err != nil {
		t.Fatalf("create real Horizon Redis store: %v", err)
	}
	manager, err := horizon.NewManager(liveHorizonConfig(resources),
		horizon.WithStoreFactory(liveStoreFactory{store: store}),
		horizon.WithQueueManager(horizon.NewQueueAdapter(queueManager)),
		horizon.WithWorkerRunner(horizon.NewQueueWorkerAdapter(queueManager)),
		horizon.WithEventDispatcher(bus),
	)
	if err != nil {
		t.Fatalf("create Horizon manager: %v", err)
	}
	if err := manager.RegisterMonitor(ctx); err != nil {
		t.Fatalf("register Horizon monitor: %v", err)
	}

	dispatcher := queue.NewDispatcher(queueManager)
	dispatchLiveJob(t, ctx, dispatcher, "redis", resources.redisQueue)
	dispatchLiveJob(t, ctx, dispatcher, "rabbitmq", resources.rabbitMQQueue)
	assertQueueSize(t, ctx, queueManager, "redis", resources.redisQueue, 1)
	assertQueueSize(t, ctx, queueManager, "rabbitmq", resources.rabbitMQQueue, 1)

	workLiveQueueOnce(t, ctx, manager, "redis", resources.redisQueue)
	workLiveQueueOnce(t, ctx, manager, "rabbitmq", resources.rabbitMQQueue)
	if got := redisJobHits.Load(); got != 1 {
		t.Fatalf("real Redis job handler hits = %d, want 1", got)
	}
	if got := rabbitMQJobHits.Load(); got != 1 {
		t.Fatalf("real RabbitMQ job handler hits = %d, want 1", got)
	}
	assertQueueSize(t, ctx, queueManager, "redis", resources.redisQueue, 0)
	assertQueueSize(t, ctx, queueManager, "rabbitmq", resources.rabbitMQQueue, 0)

	if err := app.Instance("horizon.manager", manager); err != nil {
		t.Fatalf("bind Horizon manager: %v", err)
	}
	k := kernel.New("demo-horizon-integration")
	for _, factory := range horizon.CommandFactories() {
		k.Register(factory())
	}
	if err := k.CallSilently(ctx, "horizon:snapshot"); err != nil {
		t.Fatalf("snapshot real Horizon state: %v", err)
	}
	processed := horizonProcessedJobs(t, ctx, store)
	if processed != 2 {
		t.Fatalf("real Horizon processed metric = %d, want 2", processed)
	}
	lengths, err := store.QueueLengthSnapshot(ctx)
	if err != nil {
		t.Fatalf("read real Horizon queue lengths: %v", err)
	}
	if len(lengths.Queues) != 2 {
		t.Fatalf("real Horizon queue length entries = %d, want 2", len(lengths.Queues))
	}
	for _, length := range lengths.Queues {
		if length.Size != 0 {
			t.Fatalf("real Horizon queue %q size = %d, want 0 after consumption", length.Queue, length.Size)
		}
	}
	t.Logf(
		"real Horizon flow passed: Redis hits=%d RabbitMQ hits=%d processed=%d drained queues=%d",
		redisJobHits.Load(), rabbitMQJobHits.Load(), processed, len(lengths.Queues),
	)
}

type liveResources struct {
	horizonPrefix    string
	redisQueuePrefix string
	redisQueue       string
	rabbitMQQueue    string
	rabbitMQExchange string
	restartQueue     string
}

func newLiveQueueManager(t *testing.T, rabbitMQURL string, resources liveResources) *queue.Manager {
	t.Helper()
	registry := queue.NewRegistry()
	queue.RegisterTypeTo[*liveQueueJob](registry)
	manager, err := queue.NewManager(queue.Config{
		Default: "redis",
		Connections: map[string]queue.ConnectionConfig{
			"redis": {
				Driver:     "redis",
				Queue:      resources.redisQueue,
				Prefix:     resources.redisQueuePrefix,
				RetryAfter: time.Second,
				BlockFor:   2 * time.Second,
				Options: map[string]any{
					"connection": "default",
					"prefix":     resources.redisQueuePrefix,
				},
			},
			"rabbitmq": {
				Driver:   "rabbitmq",
				Queue:    resources.rabbitMQQueue,
				BlockFor: 2 * time.Second,
				Options: map[string]any{
					"url":                rabbitMQURL,
					"exchange":           resources.rabbitMQExchange,
					"exchange_type":      "direct",
					"declare":            true,
					"exchange_durable":   false,
					"queue_durable":      false,
					"message_persistent": false,
					"confirm":            true,
					"prefetch":           1,
					"publish_timeout":    2 * time.Second,
					"restart_queue":      resources.restartQueue,
					"restart_enabled":    true,
				},
			},
		},
	}, registry)
	if err != nil {
		t.Fatalf("create Redis and RabbitMQ queue manager: %v", err)
	}
	manager.Extend("rabbitmq", func() (qcontract.Connector, error) {
		return rabbitmq.Connector{}, nil
	})
	return manager
}

func liveHorizonConfig(resources liveResources) horizon.Config {
	return horizon.Config{
		Store:        "redis",
		Environment:  "local",
		HeartbeatTTL: time.Minute,
		Supervisors: map[string]horizon.SupervisorConfig{
			"redis-live": {
				Name:       "redis-live",
				Connection: "redis",
				Queues:     []string{resources.redisQueue},
			},
			"rabbitmq-live": {
				Name:       "rabbitmq-live",
				Connection: "rabbitmq",
				Queues:     []string{resources.rabbitMQQueue},
			},
		},
	}
}

func dispatchLiveJob(t *testing.T, ctx context.Context, dispatcher *queue.Dispatcher, connection, queueName string) {
	t.Helper()
	if _, err := dispatcher.Dispatch(
		ctx,
		&liveQueueJob{Connection: connection},
		queue.OnConnection(connection),
		queue.OnQueue(queueName),
	); err != nil {
		t.Fatalf("dispatch real %s job to %q: %v", connection, queueName, err)
	}
}

func workLiveQueueOnce(t *testing.T, ctx context.Context, manager *horizon.Manager, connection, queueName string) {
	t.Helper()
	session, err := manager.WorkerRunner().Begin(ctx, queue.WorkerOptions{
		Connection:    connection,
		Queues:        []string{queueName},
		Once:          true,
		StopWhenEmpty: true,
		RetryAfter:    time.Second,
		Tries:         1,
	})
	if err != nil {
		t.Fatalf("begin real %s Horizon worker: %v", connection, err)
	}
	defer func() {
		if err := session.Close(); err != nil {
			t.Errorf("close real %s Horizon worker: %v", connection, err)
		}
	}()
	if err := session.Activate(ctx); err != nil {
		t.Fatalf("activate real %s Horizon worker: %v", connection, err)
	}
	if err := session.Work(ctx); err != nil {
		t.Fatalf("consume real %s Horizon job: %v", connection, err)
	}
}

func assertQueueSize(t *testing.T, ctx context.Context, manager *queue.Manager, connection, queueName string, want int64) {
	t.Helper()
	queueConnection, err := manager.Queue(connection)
	if err != nil {
		t.Fatalf("resolve %s queue connection: %v", connection, err)
	}
	got, err := queueConnection.Size(ctx, queueName)
	if err != nil {
		t.Fatalf("read real %s queue %q size: %v", connection, queueName, err)
	}
	if got != want {
		t.Fatalf("real %s queue %q size = %d, want %d", connection, queueName, got, want)
	}
}

func horizonProcessedJobs(t *testing.T, ctx context.Context, store horizon.Store) int64 {
	t.Helper()
	windows, err := store.EventMetricWindows(ctx, horizon.EventMetricWindowQuery{})
	if err != nil {
		t.Fatalf("read real Horizon event metrics: %v", err)
	}
	var processed int64
	for _, window := range windows.Items {
		processed += window.Processed
	}
	return processed
}

func cleanupLiveRedis(t *testing.T, client *goredis.Client, patterns ...string) {
	t.Helper()
	ctx := context.Background()
	for _, pattern := range patterns {
		var cursor uint64
		for {
			keys, next, err := client.Scan(ctx, cursor, pattern, 100).Result()
			if err != nil {
				t.Errorf("scan Redis cleanup pattern %q: %v", pattern, err)
				break
			}
			if len(keys) > 0 {
				if err := client.Del(ctx, keys...).Err(); err != nil {
					t.Errorf("delete Redis cleanup pattern %q: %v", pattern, err)
					break
				}
			}
			cursor = next
			if cursor == 0 {
				break
			}
		}
	}
}

func cleanupLiveRabbitMQ(t *testing.T, url string, resources liveResources) {
	t.Helper()
	conn, err := amqp.Dial(url)
	if err != nil {
		t.Errorf("dial RabbitMQ for Horizon cleanup: %v", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Errorf("close RabbitMQ cleanup connection: %v", err)
		}
	}()
	channel, err := conn.Channel()
	if err != nil {
		t.Errorf("open RabbitMQ cleanup channel: %v", err)
		return
	}
	defer func() {
		if err := channel.Close(); err != nil {
			t.Errorf("close RabbitMQ cleanup channel: %v", err)
		}
	}()
	_, _ = channel.QueueDelete(resources.rabbitMQQueue, false, false, false)
	_, _ = channel.QueueDelete(resources.restartQueue, false, false, false)
	_ = channel.ExchangeDelete(resources.rabbitMQExchange, false, false)
}
