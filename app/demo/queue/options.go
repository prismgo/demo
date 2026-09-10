package queuedemo

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/prismgo/framework/cache"
	queuecontract "github.com/prismgo/framework/contracts/queue"
	encodingpkg "github.com/prismgo/framework/encoding"
	"github.com/prismgo/framework/queue"
	"github.com/prismgo/framework/queue/payload"

	jobs "prismgo-demo/app/jobs/queuedemo"
)

func runStrategyEnvelope(ctx context.Context, connection string) (Result, error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo strategy-envelope is hermetic and selected with sync, got %s", connection)
	}
	manager, transport, err := newHoldingQueueManager("strategy-envelope")
	if err != nil {
		return Result{}, err
	}
	defer manager.Close()

	runID := time.Now().UnixNano()
	queueName := fmt.Sprintf("demo-strategy-envelope-%d", runID)
	retryUntil := time.Now().Add(10 * time.Minute)
	jobID, err := manager.Dispatch(ctx, &jobs.StrategyJob{TraceID: fmt.Sprintf("strategy-envelope-%d", runID)},
		queue.OnConnection("inspect"),
		queue.OnQueue(queueName),
		queue.Delay(2*time.Second),
		queue.Tries(3),
		queue.MaxExceptions(2),
		queue.Timeout(5*time.Second),
		queue.Backoff(time.Second, 2*time.Second),
		queue.RetryUntil(retryUntil),
		queue.Tags("dispatch-override"),
	)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo strategy-envelope dispatch: %w", err)
	}
	record, err := transport.singleRecord()
	if err != nil {
		return Result{}, fmt.Errorf("queue demo strategy-envelope inspect transport: %w", err)
	}
	envelope, err := decodeDemoEnvelope(record.body)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo strategy-envelope decode: %w", err)
	}
	if record.queue != queueName || record.delay != 2*time.Second || envelope.Queue != queueName {
		return Result{}, fmt.Errorf("queue demo strategy-envelope route = queue:%q envelope:%q delay:%v, want %q and 2s", record.queue, envelope.Queue, record.delay, queueName)
	}
	if envelope.MaxTries != 3 || envelope.MaxExceptions != 2 || envelope.TimeoutSec != 5 {
		return Result{}, fmt.Errorf("queue demo strategy-envelope limits = tries:%d exceptions:%d timeout:%d", envelope.MaxTries, envelope.MaxExceptions, envelope.TimeoutSec)
	}
	if len(envelope.BackoffSec) != 2 || envelope.BackoffSec[0] != 1 || envelope.BackoffSec[1] != 2 || envelope.RetryUntil != retryUntil.Unix() {
		return Result{}, fmt.Errorf("queue demo strategy-envelope retry = backoff:%v until:%d, want [1 2] and %d", envelope.BackoffSec, envelope.RetryUntil, retryUntil.Unix())
	}
	if len(envelope.Tags) != 1 || envelope.Tags[0] != "dispatch-override" || !envelope.FailOnTimeout || !envelope.Silenced {
		return Result{}, fmt.Errorf("queue demo strategy-envelope metadata = tags:%v fail-on-timeout:%t silenced:%t", envelope.Tags, envelope.FailOnTimeout, envelope.Silenced)
	}
	steps := []string{
		"connection:option", "queue:option", "delay:2s", "limits:option",
		"retry:option", "tags:option", "fail-on-timeout:provider", "silenced:provider",
	}
	return Result{Case: "strategy-envelope", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func runUniqueDispatchOptions(ctx context.Context, connection string) (Result, error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo unique-options is hermetic and selected with sync, got %s", connection)
	}
	manager, transport, err := newHoldingQueueManager("unique-options")
	if err != nil {
		return Result{}, err
	}
	defer manager.Close()
	cacheManager, err := newDemoMemoryCache("unique-options")
	if err != nil {
		return Result{}, err
	}
	defer cacheManager.Close()

	runID := time.Now().UnixNano()
	queueName := fmt.Sprintf("demo-unique-options-%d", runID)
	uniqueKey := fmt.Sprintf("queue-demo:unique-option:%d", runID)
	store := cacheManager.Default()
	dispatch := func(providerKey string) (string, error) {
		return manager.Dispatch(ctx, &jobs.UniqueJob{UniqueKey: providerKey},
			queue.OnConnection("inspect"),
			queue.OnQueue(queueName),
			queue.Unique(uniqueKey, 45*time.Second),
			queue.UniqueVia(store),
			queue.UniqueUntilProcessing(),
		)
	}
	jobID, err := dispatch("provider-first")
	if err != nil {
		return Result{}, fmt.Errorf("queue demo unique-options first dispatch: %w", err)
	}
	if _, err := dispatch("provider-second"); !errors.Is(err, queue.ErrDuplicate) {
		return Result{}, fmt.Errorf("queue demo unique-options duplicate: got %v, want %w", err, queue.ErrDuplicate)
	}
	record, err := transport.singleRecord()
	if err != nil {
		return Result{}, fmt.Errorf("queue demo unique-options inspect transport: %w", err)
	}
	envelope, err := decodeDemoEnvelope(record.body)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo unique-options decode: %w", err)
	}
	if envelope.UniqueKey != uniqueKey || envelope.UniqueForSec != 45 || envelope.UniqueVia != store.Name() || !envelope.UniqueUntil {
		return Result{}, fmt.Errorf("queue demo unique-options envelope = key:%q ttl:%d via:%q until:%t", envelope.UniqueKey, envelope.UniqueForSec, envelope.UniqueVia, envelope.UniqueUntil)
	}
	steps := []string{"key:option", "ttl:45s", "via:" + store.Name(), "until-processing:true", "duplicate:rejected"}
	return Result{Case: "unique-options", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func runDebounceDispatchOptions(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo debounce-options requires redis or rabbitmq, got %s", connection)
	}
	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("debounce-options-%d", runID)
	queueName := fmt.Sprintf("demo-debounce-options-%d", runID)
	window := 100 * time.Millisecond
	if connection == "rabbitmq" {
		window = 5 * time.Second
	}
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)

	store := cache.Default()
	debounceKey := fmt.Sprintf("queue-demo:debounce-option:%d", runID)
	var jobID string
	for _, label := range []string{"option:old", "option:new"} {
		id, err := manager.Dispatch(ctx, &jobs.DebounceJob{
			TraceID:     traceID,
			Label:       label,
			Store:       store.Name(),
			Window:      time.Minute,
			ProviderKey: debounceKey + ":provider:" + label,
		},
			queue.OnConnection(connection),
			queue.OnQueue(queueName),
			queue.Debounce(debounceKey, window),
			queue.DebounceVia(store),
		)
		if err != nil {
			return Result{}, fmt.Errorf("queue demo debounce-options dispatch %s on %s: %w", label, connection, err)
		}
		jobID = id
	}
	timer := time.NewTimer(window + 300*time.Millisecond)
	select {
	case <-ctx.Done():
		timer.Stop()
		return Result{}, ctx.Err()
	case <-timer.C:
	}
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
		Connection: connection, Queues: []string{queueName}, MaxJobs: 2, StopWhenEmpty: true,
	}); err != nil {
		return Result{}, fmt.Errorf("queue demo debounce-options worker on %s: %w", connection, err)
	}
	if err := clearRedisQueue(ctx, manager, connection, queueName); err != nil {
		return Result{}, err
	}
	steps := jobs.TakeTrace(traceID)
	if len(steps) != 1 || steps[0] != "option:new:handled" {
		return Result{}, fmt.Errorf("queue demo debounce-options on %s steps = %v, want option:new:handled", connection, steps)
	}
	steps = append(steps, "key:option", "via:"+store.Name())
	return Result{Case: "debounce-options", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func runOverlapReleasePolicy(ctx context.Context, connection string) (Result, error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo overlap-release is hermetic and selected with sync, got %s", connection)
	}
	runID := time.Now().UnixNano()
	key := fmt.Sprintf("queue-demo:overlap-release:%d", runID)
	cacheManager, err := newDemoMemoryCache("overlap-release")
	if err != nil {
		return Result{}, err
	}
	defer cacheManager.Close()
	store := cacheManager.Default()
	started := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan error, 1)
	job := &jobs.MiddlewareJob{}
	go func() {
		firstDone <- queue.WithoutOverlapping(key).Via(store).Handle(ctx, job, func(ctx context.Context) error {
			close(started)
			select {
			case <-release:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()
	select {
	case <-started:
	case firstErr := <-firstDone:
		close(release)
		return Result{}, fmt.Errorf("queue demo overlap-release acquire lock: %w", firstErr)
	case <-ctx.Done():
		close(release)
		<-firstDone
		return Result{}, fmt.Errorf("queue demo overlap-release wait for lock holder: %w", ctx.Err())
	}

	wantDelay := 25 * time.Second
	secondErr := queue.WithoutOverlapping(key).ReleaseAfter(wantDelay).ExpireAfter(time.Minute).Via(store).Shared().
		Handle(ctx, job, func(context.Context) error {
			return errors.New("queue demo overlap-release unexpectedly acquired held lock")
		})
	delay, ok := queue.ReleaseDelay(secondErr)
	close(release)
	firstErr := <-firstDone
	if firstErr != nil {
		return Result{}, fmt.Errorf("queue demo overlap-release lock holder: %w", firstErr)
	}
	if !ok || delay != wantDelay {
		return Result{}, fmt.Errorf("queue demo overlap-release delay = %v, %t, want %v, true: %w", delay, ok, wantDelay, secondErr)
	}
	steps := []string{"first:locked", "second:released-after-25s", "first:unlocked"}
	return Result{Case: "overlap-release", Connection: connection, Queue: "demo-overlap-release", Processed: true, Steps: steps}, nil
}

func newHoldingQueueManager(label string) (*queue.Manager, *demoHoldingQueue, error) {
	transport := &demoHoldingQueue{queue: queue.NewSyncConnection()}
	driver := fmt.Sprintf("demo-holding-%s-%d", label, time.Now().UnixNano())
	queue.Extend(driver, demoHoldingConnector{queue: transport})
	manager, err := queue.NewManager(queue.Config{
		Default: "inspect",
		Connections: map[string]queue.ConnectionConfig{
			"inspect": {Driver: driver, Queue: "demo-holding"},
		},
	}, queue.NewRegistry())
	if err != nil {
		return nil, nil, fmt.Errorf("queue demo %s manager: %w", label, err)
	}
	return manager, transport, nil
}

func newDemoMemoryCache(label string) (*cache.Manager, error) {
	manager, err := cache.NewManager(cache.Config{
		Default: "memory",
		Prefix:  "queue-demo-" + label,
		Stores: map[string]cache.StoreConfig{
			"memory": {Driver: "memory"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("queue demo %s memory cache: %w", label, err)
	}
	return manager, nil
}

func decodeDemoEnvelope(body queuecontract.Payload) (*payload.Envelope, error) {
	var envelope payload.Envelope
	if err := payload.QueueCodec(encodingpkg.Msgpack()).Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	return &envelope, nil
}

type demoHoldingConnector struct {
	queue *demoHoldingQueue
}

func (c demoHoldingConnector) Connect(context.Context, string, map[string]any) (queuecontract.Queue, error) {
	return c.queue, nil
}

type demoQueueRecord struct {
	queue string
	body  queuecontract.Payload
	delay time.Duration
}

type demoHoldingQueue struct {
	queue *queue.SyncConnection
	mu    sync.Mutex
	items []demoQueueRecord
}

func (q *demoHoldingQueue) Push(ctx context.Context, queueName string, body queuecontract.Payload) error {
	q.record(queueName, body, 0)
	return q.queue.Push(ctx, queueName, body)
}

func (q *demoHoldingQueue) Later(ctx context.Context, queueName string, body queuecontract.Payload, delay time.Duration) error {
	q.record(queueName, body, delay)
	return q.queue.Later(ctx, queueName, body, delay)
}

func (q *demoHoldingQueue) Bulk(ctx context.Context, queueName string, bodies []queuecontract.Payload) (queuecontract.BulkResult, error) {
	for _, body := range bodies {
		q.record(queueName, body, 0)
	}
	return q.queue.Bulk(ctx, queueName, bodies)
}

func (q *demoHoldingQueue) Pop(ctx context.Context, queues []string, wait ...queuecontract.PopWaitMode) (queuecontract.ReservedJob, error) {
	return q.queue.Pop(ctx, queues, wait...)
}

func (q *demoHoldingQueue) Size(ctx context.Context, queueName string) (int64, error) {
	return q.queue.Size(ctx, queueName)
}

func (q *demoHoldingQueue) Clear(ctx context.Context, queueName string) error {
	return q.queue.Clear(ctx, queueName)
}

func (q *demoHoldingQueue) Close() error {
	return q.queue.Close()
}

func (q *demoHoldingQueue) record(queueName string, body queuecontract.Payload, delay time.Duration) {
	q.mu.Lock()
	q.items = append(q.items, demoQueueRecord{queue: queueName, body: append(queuecontract.Payload(nil), body...), delay: delay})
	q.mu.Unlock()
}

func (q *demoHoldingQueue) singleRecord() (demoQueueRecord, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) != 1 {
		return demoQueueRecord{}, fmt.Errorf("recorded %d payloads, want 1", len(q.items))
	}
	record := q.items[0]
	record.body = append(queuecontract.Payload(nil), record.body...)
	return record, nil
}

var (
	_ queuecontract.Connector = demoHoldingConnector{}
	_ queuecontract.Queue     = (*demoHoldingQueue)(nil)
)
