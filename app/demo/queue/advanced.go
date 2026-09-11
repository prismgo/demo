package queuedemo

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	queuecontract "github.com/prismgo/framework/contracts/queue"
	encodingpkg "github.com/prismgo/framework/encoding"
	"github.com/prismgo/framework/encryption"
	"github.com/prismgo/framework/queue"
	"github.com/prismgo/framework/queue/payload"
	"github.com/prismgo/rabbitmq"

	jobs "prismgo-demo/app/jobs/queuedemo"
)

const demoEncryptionSecret = "queue-demo-sensitive-value"

func runEncryption(ctx context.Context, connection string) (result Result, err error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo encryption uses an inspecting in-memory transport, got %s", connection)
	}
	encrypter, err := encryption.New(encryption.Config{
		Key: "base64:" + base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901")),
	})
	if err != nil {
		return Result{}, fmt.Errorf("queue demo encryption create encrypter: %w", err)
	}
	connector := &demoMemoryConnector{}
	driver := "demo-encryption"
	manager, err := queue.NewManager(queue.Config{
		Default: "encrypted",
		Connections: map[string]queue.ConnectionConfig{
			"encrypted": {Driver: driver, Queue: "demo-encryption"},
		},
		PayloadEncrypter: encrypter,
	}, queue.NewRegistry())
	if err != nil {
		return Result{}, fmt.Errorf("queue demo encryption manager: %w", err)
	}
	manager.Extend(driver, func() (queuecontract.Connector, error) { return connector, nil })
	defer func() {
		if closeErr := manager.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("queue demo encryption close manager: %w", closeErr)
		}
	}()

	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("encryption-%d", runID)
	queueName := fmt.Sprintf("demo-encryption-%d", runID)
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)
	jobID, err := manager.Dispatch(ctx, &jobs.EncryptionJob{
		TraceID: traceID, Label: "provider", Secret: demoEncryptionSecret, Encrypt: true,
	}, queue.OnQueue(queueName))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo encryption provider dispatch: %w", err)
	}
	if _, err := manager.Dispatch(ctx, &jobs.EncryptionJob{
		TraceID: traceID, Label: "option", Secret: demoEncryptionSecret,
	}, queue.OnQueue(queueName), queue.Encrypt()); err != nil {
		return Result{}, fmt.Errorf("queue demo encryption option dispatch: %w", err)
	}

	bodies := connector.queue.recordedBodies()
	if len(bodies) != 2 {
		return Result{}, fmt.Errorf("queue demo encryption recorded %d payloads, want 2", len(bodies))
	}
	for index, body := range bodies {
		var envelope payload.Envelope
		if err := payload.QueueCodec(encodingpkg.Msgpack()).Unmarshal(body, &envelope); err != nil {
			return Result{}, fmt.Errorf("queue demo encryption inspect payload %d: %w", index, err)
		}
		if !envelope.Encrypted || bytes.Contains(body, []byte(demoEncryptionSecret)) {
			return Result{}, fmt.Errorf("queue demo encryption payload %d was stored in plaintext", index)
		}
	}
	steps := jobs.TakeTrace(traceID)
	for _, expected := range []string{"provider:handled", "option:handled"} {
		if !advancedContainsStep(steps, expected) {
			return Result{}, fmt.Errorf("queue demo encryption missing %q: %v", expected, steps)
		}
	}
	steps = append([]string{"provider:encrypted", "option:encrypted"}, steps...)
	return Result{Case: "encryption", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func runCustomDriver(ctx context.Context, connection string) (result Result, err error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo custom-driver is hermetic and selected with sync, got %s", connection)
	}
	connector := &demoMemoryConnector{}
	driver := "demo-memory"
	manager, err := queue.NewManager(queue.Config{
		Default: "custom",
		Connections: map[string]queue.ConnectionConfig{
			"custom": {Driver: driver, Queue: "demo-custom", Options: map[string]any{"label": "custom-driver"}},
		},
	}, queue.NewRegistry())
	if err != nil {
		return Result{}, fmt.Errorf("queue demo custom-driver manager: %w", err)
	}
	manager.Extend(driver, func() (queuecontract.Connector, error) { return connector, nil })
	defer func() {
		if closeErr := manager.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("queue demo custom-driver close manager: %w", closeErr)
		}
	}()

	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("custom-%d", runID)
	queueName := fmt.Sprintf("demo-custom-%d", runID)
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)
	jobID, err := manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: traceID, Label: "custom:handled"}, queue.OnQueue(queueName))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo custom-driver dispatch: %w", err)
	}
	steps := jobs.TakeTrace(traceID)
	if connector.name != "custom" || connector.label != "custom-driver" || !advancedContainsStep(steps, "custom:handled") {
		return Result{}, fmt.Errorf("queue demo custom-driver connector=%q label=%q steps=%v", connector.name, connector.label, steps)
	}
	steps = append([]string{"connector:resolved", "options:received"}, steps...)
	return Result{Case: "custom-driver", Connection: "custom", Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func runErrors(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo errors uses sync for deterministic error boundaries, got %s", connection)
	}
	queueConnection, err := manager.Queue(connection)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo errors resolve connection: %w", err)
	}
	if _, err := queueConnection.Pop(ctx, []string{"demo-errors"}, queuecontract.PopNoWait); !errors.Is(err, queue.ErrEmpty) {
		return Result{}, fmt.Errorf("queue demo errors empty: got %v, want %w", err, queue.ErrEmpty)
	}
	if _, err := queue.NewRegistry().Unmarshal("demo.MissingJob", nil); !errors.Is(err, queue.ErrJobNotRegistered) {
		return Result{}, fmt.Errorf("queue demo errors registry: got %v, want %w", err, queue.ErrJobNotRegistered)
	}
	if err := (&demoMemoryQueue{SyncConnection: queue.NewSyncConnection()}).Later(ctx, "demo-errors", nil, time.Second); !errors.Is(err, queue.ErrUnsupportedOperation) {
		return Result{}, fmt.Errorf("queue demo errors unsupported: got %v, want %w", err, queue.ErrUnsupportedOperation)
	}
	closedManager, err := queue.NewManager(queue.Config{
		Default: "sync",
		Connections: map[string]queue.ConnectionConfig{
			"sync": {Driver: "sync", Queue: "demo-errors"},
		},
	}, queue.NewRegistry())
	if err != nil {
		return Result{}, fmt.Errorf("queue demo errors create close boundary: %w", err)
	}
	if err := closedManager.Close(); err != nil {
		return Result{}, fmt.Errorf("queue demo errors close manager: %w", err)
	}
	if _, err := closedManager.Queue(connection); !errors.Is(err, queue.ErrManagerClosed) {
		return Result{}, fmt.Errorf("queue demo errors closed manager: got %v, want %w", err, queue.ErrManagerClosed)
	}
	steps := []string{"empty:matched", "job-not-registered:matched", "unsupported-operation:matched", "manager-closed:matched"}
	return Result{Case: "errors", Connection: connection, Queue: "demo-errors", Processed: true, Steps: steps}, nil
}

func runRedisBoundaries(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" {
		return Result{}, fmt.Errorf("queue demo redis requires redis, got %s", connection)
	}
	return runTransportBoundaries(ctx, manager, connection, "demo-redis")
}

func runRabbitMQBoundaries(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo rabbitmq requires rabbitmq, got %s", connection)
	}
	if got := rabbitmq.NormalizeQueues(nil); len(got) != 1 || got[0] != "default" {
		return Result{}, fmt.Errorf("queue demo rabbitmq default queue: %v", got)
	}
	if got := rabbitmq.SanitizeDelayBuckets([]time.Duration{-time.Second, 2 * time.Second}); len(got) != 1 || got[0] != 2*time.Second {
		return Result{}, fmt.Errorf("queue demo rabbitmq delay buckets: %v", got)
	}
	result, err := runTransportBoundaries(ctx, manager, connection, "demo-rabbitmq")
	if err != nil {
		return Result{}, err
	}
	result.Steps = append([]string{"config:default-queue", "config:delay-buckets"}, result.Steps...)
	return result, nil
}

func extendRabbitMQ(m *queue.Manager) {
	m.Extend("rabbitmq", func() (queuecontract.Connector, error) {
		return rabbitmq.Connector{}, nil
	})
}

func runTransportBoundaries(ctx context.Context, manager *queue.Manager, connection, prefix string) (Result, error) {
	queueName := fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	queueConnection, err := manager.Queue(connection)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo %s resolve transport: %w", connection, err)
	}
	if err := queueConnection.Clear(ctx, queueName); err != nil {
		return Result{}, fmt.Errorf("queue demo %s initial clear: %w", connection, err)
	}
	traceID := fmt.Sprintf("%s-%d", connection, time.Now().UnixNano())
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)
	dispatch := func(label string) (string, error) {
		return manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: traceID, Label: label}, queue.OnConnection(connection), queue.OnQueue(queueName))
	}
	if _, err := dispatch("cleared:must-not-run"); err != nil {
		return Result{}, fmt.Errorf("queue demo %s first dispatch: %w", connection, err)
	}
	if size, err := queueConnection.Size(ctx, queueName); err != nil || size != 1 {
		return Result{}, fmt.Errorf("queue demo %s size after dispatch: size=%d err=%v", connection, size, err)
	}
	if err := queueConnection.Clear(ctx, queueName); err != nil {
		return Result{}, fmt.Errorf("queue demo %s clear: %w", connection, err)
	}
	if size, err := queueConnection.Size(ctx, queueName); err != nil || size != 0 {
		return Result{}, fmt.Errorf("queue demo %s size after clear: size=%d err=%v", connection, size, err)
	}
	if _, err := queueConnection.Pop(ctx, []string{queueName}, queuecontract.PopNoWait); !errors.Is(err, queue.ErrEmpty) {
		return Result{}, fmt.Errorf("queue demo %s empty pop: got %v, want %w", connection, err, queue.ErrEmpty)
	}
	jobID, err := dispatch("worker:handled")
	if err != nil {
		return Result{}, fmt.Errorf("queue demo %s second dispatch: %w", connection, err)
	}
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{Connection: connection, Queues: []string{queueName}, Once: true}); err != nil {
		return Result{}, fmt.Errorf("queue demo %s worker: %w", connection, err)
	}
	trace := jobs.TakeTrace(traceID)
	if advancedContainsStep(trace, "cleared:must-not-run") || !advancedContainsStep(trace, "worker:handled") {
		return Result{}, fmt.Errorf("queue demo %s transport trace: %v", connection, trace)
	}
	steps := []string{"size:1", "clear:size=0", "empty:matched", "worker:handled"}
	return Result{Case: connection, Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func advancedContainsStep(steps []string, expected string) bool {
	for _, step := range steps {
		if step == expected {
			return true
		}
	}
	return false
}

type demoMemoryConnector struct {
	name  string
	label string
	queue *demoMemoryQueue
}

func (c *demoMemoryConnector) Connect(_ context.Context, name string, config queuecontract.ConnectorConfig) (queuecontract.Queue, error) {
	c.name = name
	c.label, _ = config.Options["label"].(string)
	c.queue = &demoMemoryQueue{SyncConnection: queue.NewSyncConnection()}
	return c.queue, nil
}

type demoMemoryQueue struct {
	*queue.SyncConnection
	mu     sync.Mutex
	bodies []queuecontract.Payload
}

func (q *demoMemoryQueue) Push(ctx context.Context, queueName string, body queuecontract.Payload) error {
	q.mu.Lock()
	q.bodies = append(q.bodies, append(queuecontract.Payload(nil), body...))
	q.mu.Unlock()
	return q.SyncConnection.Push(ctx, queueName, body)
}

func (q *demoMemoryQueue) Later(context.Context, string, queuecontract.Payload, time.Duration) error {
	return fmt.Errorf("demo memory queue delay: %w", queue.ErrUnsupportedOperation)
}

func (q *demoMemoryQueue) recordedBodies() []queuecontract.Payload {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]queuecontract.Payload, len(q.bodies))
	for index, body := range q.bodies {
		out[index] = append(queuecontract.Payload(nil), body...)
	}
	return out
}

var (
	_ queuecontract.Connector = (*demoMemoryConnector)(nil)
	_ queuecontract.Queue     = (*demoMemoryQueue)(nil)
)
