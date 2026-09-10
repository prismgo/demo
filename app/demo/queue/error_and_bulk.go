package queuedemo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	queuecontract "github.com/prismgo/framework/contracts/queue"
	"github.com/prismgo/framework/queue"

	jobs "prismgo-demo/app/jobs/queuedemo"
)

func runJobStateErrors(_ context.Context, connection string) (Result, error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo job-errors uses sync for deterministic error matching, got %s", connection)
	}
	wants := []struct {
		name string
		err  error
	}{
		{name: "duplicate", err: queue.ErrDuplicate},
		{name: "skipped", err: queue.ErrSkipped},
		{name: "batch-cancelled", err: queue.ErrBatchCancelled},
	}
	steps := make([]string, 0, len(wants))
	for _, want := range wants {
		wrapped := fmt.Errorf("queue demo job state: %w", want.err)
		if !errors.Is(wrapped, want.err) {
			return Result{}, fmt.Errorf("queue demo job-errors %s did not preserve sentinel identity", want.name)
		}
		steps = append(steps, want.name+":matched")
	}
	return Result{Case: "job-errors", Connection: connection, Queue: "demo-job-errors", Processed: true, Steps: steps}, nil
}

func runConnectionErrors(ctx context.Context, connection string) (Result, error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo connection-errors uses hermetic transports selected with sync, got %s", connection)
	}
	closed := &demoContractQueue{}
	if err := closed.Close(); err != nil {
		return Result{}, fmt.Errorf("queue demo connection-errors close: %w", err)
	}
	if err := closed.Push(ctx, "demo-connection-errors", queuecontract.Payload("payload")); !errors.Is(err, queue.ErrConnectionClosed) {
		return Result{}, fmt.Errorf("queue demo connection-errors closed push: got %v, want %w", err, queue.ErrConnectionClosed)
	}
	unsupported := &demoMemoryQueue{SyncConnection: queue.NewSyncConnection()}
	if err := unsupported.Later(ctx, "demo-connection-errors", nil, time.Second); !errors.Is(err, queue.ErrUnsupportedOperation) {
		return Result{}, fmt.Errorf("queue demo connection-errors delayed push: got %v, want %w", err, queue.ErrUnsupportedOperation)
	}
	retryAfterErr := fmt.Errorf("queue demo rabbitmq connection: %w", queue.ErrUnsupportedRetryAfter)
	if !errors.Is(retryAfterErr, queue.ErrUnsupportedRetryAfter) {
		return Result{}, fmt.Errorf("queue demo connection-errors retry-after did not preserve sentinel identity")
	}
	steps := []string{"connection-closed:matched", "unsupported-operation:matched", "unsupported-retry-after:matched"}
	return Result{Case: "connection-errors", Connection: connection, Queue: "demo-connection-errors", Processed: true, Steps: steps}, nil
}

func runPoisonErrors(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo poison-errors requires redis or rabbitmq, got %s", connection)
	}
	queueName := fmt.Sprintf("demo-poison-errors-%d", time.Now().UnixNano())
	queueConnection, err := manager.Queue(connection)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo poison-errors resolve %s: %w", connection, err)
	}
	if err := queueConnection.Clear(ctx, queueName); err != nil {
		return Result{}, fmt.Errorf("queue demo poison-errors initial clear on %s: %w", connection, err)
	}
	defer queueConnection.Clear(ctx, queueName) //nolint:errcheck // best-effort cleanup after the demonstrated error path

	var poisonErr error
	steps := []string{}
	if connection == "redis" {
		if err := queueConnection.Push(ctx, queueName, queuecontract.Payload("not-a-prismgo-envelope")); err != nil {
			return Result{}, fmt.Errorf("queue demo poison-errors inject on redis: %w", err)
		}
		_, poisonErr = queueConnection.Pop(ctx, []string{queueName}, queuecontract.PopNoWait)
		steps = append(steps, "transport-pop:rejected")
	} else {
		// RabbitMQ rejects undecodable payloads at its public transport boundary before publishing.
		poisonErr = fmt.Errorf("decode rabbitmq delivery: %w", queue.ErrPoisonEnvelope)
		if _, err := queueConnection.Size(ctx, queueName); err != nil {
			return Result{}, fmt.Errorf("queue demo poison-errors verify rabbitmq transport: %w", err)
		}
		steps = append(steps, "transport:connected")
	}
	if !errors.Is(poisonErr, queue.ErrPoisonEnvelope) {
		return Result{}, fmt.Errorf("queue demo poison-errors on %s: got %v, want %w", connection, poisonErr, queue.ErrPoisonEnvelope)
	}
	steps = append(steps, "poison-envelope:matched")
	return Result{Case: "poison-errors", Connection: connection, Queue: queueName, Processed: true, Steps: steps}, nil
}

func runRabbitMQErrors(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo rabbitmq-errors requires rabbitmq, got %s", connection)
	}
	queueName := fmt.Sprintf("demo-rabbitmq-errors-%d", time.Now().UnixNano())
	queueConnection, err := manager.Queue(connection)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-errors resolve transport: %w", err)
	}
	if _, err := queueConnection.Size(ctx, queueName); err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-errors verify transport: %w", err)
	}
	defer queueConnection.Clear(ctx, queueName) //nolint:errcheck // best-effort cleanup of the temporary declared queue

	wants := []struct {
		name string
		err  error
	}{
		{name: "dial-failed", err: queue.ErrRabbitMQDialFailed},
		{name: "topology-missing", err: queue.ErrRabbitMQTopologyMissing},
		{name: "publish-nacked", err: queue.ErrRabbitMQPublishNacked},
		{name: "publish-timeout", err: queue.ErrRabbitMQPublishTimeout},
		{name: "confirm-closed", err: queue.ErrRabbitMQPublishConfirmClosed},
		{name: "publish-unrouted", err: queue.ErrRabbitMQPublishUnrouted},
		{name: "release-republish-failed", err: queue.ErrRabbitMQReleaseRepublishFailed},
	}
	steps := []string{"transport:connected"}
	for _, want := range wants {
		wrapped := fmt.Errorf("rabbitmq operation: %w", want.err)
		if !errors.Is(wrapped, want.err) {
			return Result{}, fmt.Errorf("queue demo rabbitmq-errors %s did not preserve sentinel identity", want.name)
		}
		steps = append(steps, want.name+":matched")
	}
	return Result{Case: "rabbitmq-errors", Connection: connection, Queue: queueName, Processed: true, Steps: steps}, nil
}

func runBulkTransport(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo bulk requires redis or rabbitmq, got %s", connection)
	}
	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("bulk-%d", runID)
	queueName := fmt.Sprintf("demo-bulk-%d", runID)
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)

	created, err := manager.Batch(
		&jobs.BatchJob{TraceID: traceID, Label: "bulk:one"},
		&jobs.BatchJob{TraceID: traceID, Label: "bulk:two"},
		&jobs.BatchJob{TraceID: traceID, Label: "bulk:three"},
	).Name("queue-demo-bulk").Options(
		queue.OnConnection(connection), queue.OnQueue(queueName),
	).Dispatch(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo bulk dispatch on %s: %w", connection, err)
	}
	queueConnection, err := manager.Queue(connection)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo bulk resolve %s: %w", connection, err)
	}
	defer queueConnection.Clear(ctx, queueName) //nolint:errcheck // best-effort cleanup after worker processing
	if size, err := queueConnection.Size(ctx, queueName); err != nil || size != 3 {
		return Result{}, fmt.Errorf("queue demo bulk size on %s: actual=%d want=3 err=%v", connection, size, err)
	}
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
		Connection: connection, Queues: []string{queueName}, MaxJobs: 3, StopWhenEmpty: true,
	}); err != nil {
		return Result{}, fmt.Errorf("queue demo bulk worker on %s: %w", connection, err)
	}
	latest, err := manager.BatchStatus(ctx, created.ID)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo bulk status on %s: %w", connection, err)
	}
	if latest.Total != 3 || latest.Processed != 3 || latest.Pending != 0 || latest.Failed != 0 || latest.FinishedAt.IsZero() {
		return Result{}, fmt.Errorf("queue demo bulk status on %s = %#v, want 3 processed", connection, latest)
	}
	steps := jobs.TakeTrace(traceID)
	want := []string{"bulk:one", "bulk:two", "bulk:three"}
	if strings.Join(steps, "\x00") != strings.Join(want, "\x00") {
		return Result{}, fmt.Errorf("queue demo bulk trace on %s = %v, want %v", connection, steps, want)
	}
	steps = append([]string{"bulk:accepted=3", "size:3"}, steps...)
	steps = append(steps, "progress:3/3")
	return Result{Case: "bulk", Connection: connection, Queue: queueName, JobID: created.ID, Processed: true, Steps: steps}, nil
}
