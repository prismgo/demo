package queuedemo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	queuecontract "github.com/prismgo/framework/contracts/queue"
	"github.com/prismgo/framework/queue"

	jobs "prismgo-demo/app/jobs/queuedemo"
)

func runTransportDelay(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo transport-delay requires redis or rabbitmq, got %s", connection)
	}
	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("transport-delay-%d", runID)
	queueName := fmt.Sprintf("demo-transport-delay-%d", runID)
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)

	jobID, err := manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: traceID, Label: "after-due:handled"},
		queue.OnConnection(connection), queue.OnQueue(queueName), queue.Delay(time.Second))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo transport-delay dispatch on %s: %w", connection, err)
	}
	queueConnection, err := manager.Queue(connection)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo transport-delay resolve %s: %w", connection, err)
	}
	defer queueConnection.Clear(ctx, queueName) //nolint:errcheck // best-effort cleanup of the temporary queue
	if _, err := queueConnection.Pop(ctx, []string{queueName}, queuecontract.PopNoWait); !errors.Is(err, queue.ErrEmpty) {
		return Result{}, fmt.Errorf("queue demo transport-delay before due on %s: got %v, want %w", connection, err, queue.ErrEmpty)
	}
	// ttl_dlx rounds the requested delay up to the nearest configured bucket. The
	// documented default starts at five seconds, so wait through that boundary.
	timer := time.NewTimer(5500 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case <-timer.C:
	}
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
		Connection: connection, Queues: []string{queueName}, Once: true,
	}); err != nil {
		return Result{}, fmt.Errorf("queue demo transport-delay worker on %s: %w", connection, err)
	}
	steps := jobs.TakeTrace(traceID)
	if len(steps) != 1 || steps[0] != "after-due:handled" {
		return Result{}, fmt.Errorf("queue demo transport-delay trace on %s = %v, want [after-due:handled]", connection, steps)
	}
	return Result{
		Case: "transport-delay", Connection: connection, Queue: queueName, JobID: jobID,
		Processed: true, Steps: []string{"before-due:empty", steps[0]},
	}, nil
}

func runBlockingPop(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo blocking-pop requires redis or rabbitmq, got %s", connection)
	}
	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("blocking-pop-%d", runID)
	queueName := fmt.Sprintf("demo-blocking-pop-%d", runID)
	wakeQueue := queueName + "-wake"
	highQueue := queueName + "-high"
	lowQueue := queueName + "-low"
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)

	queueConnection, err := manager.Queue(connection)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo blocking-pop resolve %s: %w", connection, err)
	}
	defer func() {
		_ = queueConnection.Clear(ctx, wakeQueue)
		_ = queueConnection.Clear(ctx, highQueue)
		_ = queueConnection.Clear(ctx, lowQueue)
	}()

	reservedResult := make(chan queuecontract.ReservedJob, 1)
	popError := make(chan error, 1)
	popCtx, cancelPop := context.WithCancel(ctx)
	popDone := make(chan struct{})
	go func() {
		defer close(popDone)
		reserved, popErr := queueConnection.Pop(popCtx, []string{wakeQueue}, queuecontract.PopWaitAvailable)
		if popErr != nil {
			popError <- popErr
			return
		}
		reservedResult <- reserved
	}()
	defer func() {
		cancelPop()
		select {
		case <-popDone:
		case <-time.After(3 * time.Second):
		}
	}()
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case <-timer.C:
	}
	jobID, err := manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: traceID, Label: "blocking:reserved"},
		queue.OnConnection(connection), queue.OnQueue(wakeQueue))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo blocking-pop wake dispatch on %s: %w", connection, err)
	}
	select {
	case reserved := <-reservedResult:
		if reserved.ID() != jobID {
			return Result{}, fmt.Errorf("queue demo blocking-pop reserved ID on %s = %q, want %q", connection, reserved.ID(), jobID)
		}
		if err := reserved.Delete(ctx); err != nil {
			return Result{}, fmt.Errorf("queue demo blocking-pop delete on %s: %w", connection, err)
		}
	case err := <-popError:
		return Result{}, fmt.Errorf("queue demo blocking-pop wait on %s: %w", connection, err)
	case <-time.After(3 * time.Second):
		return Result{}, fmt.Errorf("queue demo blocking-pop wait on %s timed out", connection)
	}

	for _, item := range []struct {
		queue string
		label string
	}{{queue: lowQueue, label: "priority:low"}, {queue: highQueue, label: "priority:high"}} {
		if _, err := manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: traceID, Label: item.label},
			queue.OnConnection(connection), queue.OnQueue(item.queue)); err != nil {
			return Result{}, fmt.Errorf("queue demo blocking-pop dispatch %s on %s: %w", item.label, connection, err)
		}
	}
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
		Connection: connection, Queues: []string{highQueue, lowQueue}, MaxJobs: 2, StopWhenEmpty: true,
	}); err != nil {
		return Result{}, fmt.Errorf("queue demo blocking-pop priority worker on %s: %w", connection, err)
	}
	steps := jobs.TakeTrace(traceID)
	want := []string{"priority:high", "priority:low"}
	if len(steps) != len(want) || steps[0] != want[0] || steps[1] != want[1] {
		return Result{}, fmt.Errorf("queue demo blocking-pop priority on %s = %v, want %v", connection, steps, want)
	}
	return Result{
		Case: "blocking-pop", Connection: connection, Queue: queueName, JobID: jobID,
		Processed: true, Steps: append([]string{"blocking:woke"}, steps...),
	}, nil
}

func runRedisRetryAfter(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" {
		return Result{}, fmt.Errorf("queue demo redis-retry-after requires redis, got %s", connection)
	}
	runID := time.Now().UnixNano()
	queueName := fmt.Sprintf("demo-redis-retry-after-%d", runID)
	queueConnection, err := manager.Queue(connection)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo redis-retry-after resolve Redis: %w", err)
	}
	defer queueConnection.Clear(ctx, queueName) //nolint:errcheck // best-effort cleanup of the temporary queue
	jobID, err := manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: fmt.Sprintf("redis-retry-after-%d", runID), Label: "unused"},
		queue.OnConnection(connection), queue.OnQueue(queueName))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo redis-retry-after dispatch: %w", err)
	}
	first, err := queueConnection.Pop(ctx, []string{queueName}, queuecontract.PopNoWait)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo redis-retry-after first reservation: %w", err)
	}
	if first.ID() != jobID || first.Attempts() != 1 {
		return Result{}, fmt.Errorf("queue demo redis-retry-after first reservation = id %q attempts %d, want id %q attempts 1", first.ID(), first.Attempts(), jobID)
	}
	timer := time.NewTimer(350 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case <-timer.C:
	}
	second, err := queueConnection.Pop(ctx, []string{queueName}, queuecontract.PopNoWait)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo redis-retry-after second reservation: %w", err)
	}
	if second.ID() != jobID || second.Attempts() != 2 {
		return Result{}, fmt.Errorf("queue demo redis-retry-after second reservation = id %q attempts %d, want id %q attempts 2", second.ID(), second.Attempts(), jobID)
	}
	if err := second.Delete(ctx); err != nil {
		return Result{}, fmt.Errorf("queue demo redis-retry-after delete: %w", err)
	}
	return Result{
		Case: "redis-retry-after", Connection: connection, Queue: queueName, JobID: jobID,
		Processed: true, Steps: []string{"attempts:1", "retry-after:visible", "attempts:2"},
	}, nil
}

func runRabbitMQRetryAfter(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo rabbitmq-retry-after requires rabbitmq, got %s", connection)
	}
	queueName := fmt.Sprintf("demo-rabbitmq-retry-after-%d", time.Now().UnixNano())
	queueConnection, err := manager.Queue(connection)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-retry-after resolve transport: %w", err)
	}
	if _, err := queueConnection.Size(ctx, queueName); err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-retry-after verify transport: %w", err)
	}
	defer queueConnection.Clear(ctx, queueName) //nolint:errcheck // best-effort cleanup of the declared queue

	rejected, err := queue.NewManager(queue.Config{
		Default: "rabbitmq",
		Connections: map[string]queue.ConnectionConfig{
			"rabbitmq": {Driver: "rabbitmq", Queue: queueName, RetryAfter: time.Second},
		},
	}, queue.NewRegistry())
	if err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-retry-after manager: %w", err)
	}
	defer rejected.Close() //nolint:errcheck // best-effort cleanup of the rejected manager
	_, err = rejected.Queue("rabbitmq")
	if !errors.Is(err, queue.ErrUnsupportedRetryAfter) {
		return Result{}, fmt.Errorf("queue demo rabbitmq-retry-after queue error = %v, want %w", err, queue.ErrUnsupportedRetryAfter)
	}
	return Result{
		Case: "rabbitmq-retry-after", Connection: connection, Queue: queueName,
		Processed: true, Steps: []string{"transport:connected", "retry-after:rejected"},
	}, nil
}

func runRabbitMQPublisherConfirm(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo rabbitmq-confirm requires rabbitmq, got %s", connection)
	}
	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("rabbitmq-confirm-%d", runID)
	queueName := fmt.Sprintf("demo-rabbitmq-confirm-%d", runID)
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)

	var mu sync.Mutex
	queued := false
	previous := queue.CurrentEventSink()
	queue.UseEventSink(func(eventCtx context.Context, event queue.Event) {
		if observed, ok := event.(queue.JobQueued); ok && observed.Queue == queueName {
			mu.Lock()
			queued = observed.Connection == connection && observed.JobID != ""
			mu.Unlock()
		}
		if previous != nil {
			previous(eventCtx, event)
		}
	})
	jobID, err := manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: traceID, Label: "worker:handled"},
		queue.OnConnection(connection), queue.OnQueue(queueName))
	queue.UseEventSink(previous)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-confirm dispatch: %w", err)
	}
	mu.Lock()
	wasQueued := queued
	mu.Unlock()
	if !wasQueued {
		return Result{}, fmt.Errorf("queue demo rabbitmq-confirm did not observe %s after accepted publish", queue.EventJobQueued)
	}
	queueConnection, err := manager.Queue(connection)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-confirm resolve transport: %w", err)
	}
	defer queueConnection.Clear(ctx, queueName) //nolint:errcheck // best-effort cleanup of the temporary queue
	if size, err := queueConnection.Size(ctx, queueName); err != nil || size != 1 {
		return Result{}, fmt.Errorf("queue demo rabbitmq-confirm size = %d, want 1: %v", size, err)
	}
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
		Connection: connection, Queues: []string{queueName}, Once: true,
	}); err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-confirm worker: %w", err)
	}
	steps := jobs.TakeTrace(traceID)
	if len(steps) != 1 || steps[0] != "worker:handled" {
		return Result{}, fmt.Errorf("queue demo rabbitmq-confirm trace = %v, want [worker:handled]", steps)
	}
	return Result{
		Case: "rabbitmq-confirm", Connection: connection, Queue: queueName, JobID: jobID,
		Processed: true, Steps: []string{"publisher-confirm:accepted", queue.EventJobQueued, steps[0]},
	}, nil
}

func runRabbitMQTopology(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo rabbitmq-topology requires rabbitmq, got %s", connection)
	}
	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("rabbitmq-topology-%d", runID)
	queueName := fmt.Sprintf("demo-rabbitmq-topology-%d", runID)
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)

	var mu sync.Mutex
	var topology queue.InfrastructureEvent
	previous := queue.CurrentEventSink()
	queue.UseEventSink(func(eventCtx context.Context, event queue.Event) {
		if observed, ok := event.(queue.InfrastructureEvent); ok && observed.Name() == queue.EventTopologyDeclared && observed.Queue == queueName {
			mu.Lock()
			topology = observed
			mu.Unlock()
		}
		if previous != nil {
			previous(eventCtx, event)
		}
	})
	jobID, err := manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: traceID, Label: "queue:routed"},
		queue.OnConnection(connection), queue.OnQueue(queueName))
	queue.UseEventSink(previous)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-topology dispatch: %w", err)
	}
	mu.Lock()
	observed := topology
	mu.Unlock()
	if observed.Connection != connection || observed.Driver != "rabbitmq" || observed.Queue != queueName || observed.Exchange == "" || observed.Error != "" || observed.Timestamp.IsZero() {
		return Result{}, fmt.Errorf("queue demo rabbitmq-topology event = %#v", observed)
	}
	queueConnection, err := manager.Queue(connection)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-topology resolve transport: %w", err)
	}
	defer queueConnection.Clear(ctx, queueName) //nolint:errcheck // best-effort cleanup of the temporary queue
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
		Connection: connection, Queues: []string{queueName}, Once: true,
	}); err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-topology worker: %w", err)
	}
	steps := jobs.TakeTrace(traceID)
	if len(steps) != 1 || steps[0] != "queue:routed" {
		return Result{}, fmt.Errorf("queue demo rabbitmq-topology trace = %v, want [queue:routed]", steps)
	}
	return Result{
		Case: "rabbitmq-topology", Connection: connection, Queue: queueName, JobID: jobID,
		Processed: true, Steps: []string{observed.Name(), "exchange:named", steps[0]},
	}, nil
}

func runRabbitMQDelayModes(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo rabbitmq-delay-modes requires rabbitmq, got %s", connection)
	}
	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("rabbitmq-delay-modes-%d", runID)
	queueName := fmt.Sprintf("demo-rabbitmq-delay-modes-%d", runID)
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)

	ttlConnection := "rabbitmq-ttl_dlx"
	jobID, err := manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: traceID, Label: "ttl_dlx:handled"},
		queue.OnConnection(ttlConnection), queue.OnQueue(queueName), queue.Delay(100*time.Millisecond))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-delay-modes ttl_dlx dispatch: %w", err)
	}
	timer := time.NewTimer(350 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case <-timer.C:
	}
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
		Connection: ttlConnection, Queues: []string{queueName}, Once: true,
	}); err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-delay-modes ttl_dlx worker: %w", err)
	}
	ttlQueue, err := manager.Queue(ttlConnection)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-delay-modes resolve ttl_dlx: %w", err)
	}
	defer ttlQueue.Clear(ctx, queueName) //nolint:errcheck // best-effort cleanup of the temporary queue

	_, err = manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: traceID, Label: "none:must-not-run"},
		queue.OnConnection("rabbitmq-none"), queue.OnQueue(queueName+"-none"), queue.Delay(time.Second))
	if !errors.Is(err, queue.ErrUnsupportedOperation) {
		return Result{}, fmt.Errorf("queue demo rabbitmq-delay-modes none error = %v, want %w", err, queue.ErrUnsupportedOperation)
	}

	pluginStep := "plugin:accepted"
	pluginQueueName := queueName + "-plugin"
	pluginJobID, pluginErr := manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: traceID, Label: "plugin:handled"},
		queue.OnConnection("rabbitmq-plugin"), queue.OnQueue(pluginQueueName), queue.Delay(100*time.Millisecond))
	if pluginErr != nil {
		message := strings.ToLower(pluginErr.Error())
		if !strings.Contains(message, "x-delayed-message") && !strings.Contains(message, "unknown exchange type") {
			return Result{}, fmt.Errorf("queue demo rabbitmq-delay-modes plugin dispatch: %w", pluginErr)
		}
		pluginStep = "plugin:requires-extension"
	} else {
		pluginTimer := time.NewTimer(200 * time.Millisecond)
		defer pluginTimer.Stop()
		select {
		case <-ctx.Done():
			return Result{}, ctx.Err()
		case <-pluginTimer.C:
		}
		if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
			Connection: "rabbitmq-plugin", Queues: []string{pluginQueueName}, Once: true,
		}); err != nil {
			return Result{}, fmt.Errorf("queue demo rabbitmq-delay-modes plugin worker for %s: %w", pluginJobID, err)
		}
	}
	steps := jobs.TakeTrace(traceID)
	if !resultHasStep(steps, "ttl_dlx:handled") {
		return Result{}, fmt.Errorf("queue demo rabbitmq-delay-modes trace = %v, want ttl_dlx:handled", steps)
	}
	return Result{
		Case: "rabbitmq-delay-modes", Connection: connection, Queue: queueName, JobID: jobID,
		Processed: true, Steps: []string{"ttl_dlx:handled", "none:rejected", pluginStep},
	}, nil
}

func runRabbitMQReconnect(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo rabbitmq-reconnect requires rabbitmq, got %s", connection)
	}
	management, err := newRabbitMQManagementClient()
	if err != nil {
		return Result{}, err
	}
	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("rabbitmq-reconnect-%d", runID)
	queueName := fmt.Sprintf("demo-rabbitmq-reconnect-%d", runID)
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)
	queueConnection, err := manager.Queue(connection)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-reconnect resolve transport: %w", err)
	}
	defer queueConnection.Clear(ctx, queueName) //nolint:errcheck // best-effort cleanup of the temporary queue
	wantEvents := []string{
		queue.EventConnectionDisconnected,
		queue.EventConnectionReconnecting,
		queue.EventConnectionReconnected,
	}
	events := make(chan string, len(wantEvents)+2)
	consumerStarted := make(chan struct{}, 1)
	previous := queue.CurrentEventSink()
	queue.UseEventSink(func(eventCtx context.Context, event queue.Event) {
		if observed, ok := event.(queue.InfrastructureEvent); ok && observed.Connection == connection {
			if observed.Name() == queue.EventConsumerStarted && observed.Queue == queueName {
				select {
				case consumerStarted <- struct{}{}:
				default:
				}
			}
			for _, eventName := range wantEvents {
				if observed.Name() == eventName {
					events <- observed.Name()
					break
				}
			}
		}
		if previous != nil {
			previous(eventCtx, event)
		}
	})
	defer queue.UseEventSink(previous)
	workerCtx, cancelWorker := context.WithCancel(ctx)
	workerDone := make(chan error, 1)
	go func() {
		workerDone <- queue.NewWorker(manager).Work(workerCtx, queue.WorkerOptions{
			Connection: connection, Queues: []string{queueName}, MaxJobs: 1, Sleep: 25 * time.Millisecond,
		})
	}()
	workerStopped := false
	defer func() {
		cancelWorker()
		if !workerStopped {
			select {
			case <-workerDone:
			case <-time.After(3 * time.Second):
			}
		}
	}()
	select {
	case <-consumerStarted:
	case err := <-workerDone:
		workerStopped = true
		return Result{}, fmt.Errorf("queue demo rabbitmq-reconnect worker stopped before fault injection: %v", err)
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case <-time.After(3 * time.Second):
		return Result{}, fmt.Errorf("queue demo rabbitmq-reconnect worker did not start consumer")
	}
	managedConnection, err := management.connectionForQueue(ctx, queueName)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-reconnect select connection: %w", err)
	}
	if err := management.closeConnection(ctx, managedConnection); err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-reconnect inject disconnect: %w", err)
	}

	observedEvents := make([]string, 0, len(wantEvents))
	timeout := time.NewTimer(8 * time.Second)
	defer timeout.Stop()
	for len(observedEvents) < len(wantEvents) {
		select {
		case eventName := <-events:
			if eventName == wantEvents[len(observedEvents)] {
				observedEvents = append(observedEvents, eventName)
			}
		case <-ctx.Done():
			return Result{}, ctx.Err()
		case <-timeout.C:
			return Result{}, fmt.Errorf("queue demo rabbitmq-reconnect events = %v, want %v", observedEvents, wantEvents)
		}
	}

	jobID, err := manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: traceID, Label: "worker:handled"},
		queue.OnConnection(connection), queue.OnQueue(queueName))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo rabbitmq-reconnect post-reconnect dispatch: %w", err)
	}
	select {
	case err := <-workerDone:
		workerStopped = true
		if err != nil {
			return Result{}, fmt.Errorf("queue demo rabbitmq-reconnect worker: %w", err)
		}
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case <-time.After(5 * time.Second):
		return Result{}, fmt.Errorf("queue demo rabbitmq-reconnect worker did not process post-reconnect job")
	}
	steps := jobs.TakeTrace(traceID)
	if len(steps) != 1 || steps[0] != "worker:handled" {
		return Result{}, fmt.Errorf("queue demo rabbitmq-reconnect trace = %v, want [worker:handled]", steps)
	}
	return Result{
		Case: "rabbitmq-reconnect", Connection: connection, Queue: queueName, JobID: jobID,
		Processed: true, Steps: append(observedEvents, steps[0]),
	}, nil
}

func runPoisonEnvelopeRejection(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo poison-rejection requires rabbitmq, got %s", connection)
	}
	management, err := newRabbitMQManagementClient()
	if err != nil {
		return Result{}, err
	}
	queueName := fmt.Sprintf("demo-poison-rejection-%d", time.Now().UnixNano())
	queueConnection, err := manager.Queue(connection)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo poison-rejection resolve transport: %w", err)
	}
	defer queueConnection.Clear(ctx, queueName) //nolint:errcheck // best-effort cleanup of the temporary queue

	var mu sync.Mutex
	var topology queue.InfrastructureEvent
	var poison queue.PoisonEnvelope
	previous := queue.CurrentEventSink()
	queue.UseEventSink(func(eventCtx context.Context, event queue.Event) {
		mu.Lock()
		switch observed := event.(type) {
		case queue.InfrastructureEvent:
			if observed.Name() == queue.EventTopologyDeclared && observed.Queue == queueName {
				topology = observed
			}
		case queue.PoisonEnvelope:
			if observed.Queue == queueName {
				poison = observed
			}
		}
		mu.Unlock()
		if previous != nil {
			previous(eventCtx, event)
		}
	})
	defer queue.UseEventSink(previous)
	if _, err := queueConnection.Size(ctx, queueName); err != nil {
		return Result{}, fmt.Errorf("queue demo poison-rejection declare queue: %w", err)
	}
	mu.Lock()
	exchange := topology.Exchange
	mu.Unlock()
	if exchange == "" {
		return Result{}, fmt.Errorf("queue demo poison-rejection did not observe declared exchange")
	}
	body := []byte("not-a-prismgo-envelope")
	if err := management.publish(ctx, exchange, queueName, body); err != nil {
		return Result{}, fmt.Errorf("queue demo poison-rejection inject message: %w", err)
	}
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
		Connection: connection, Queues: []string{queueName}, Once: true,
	}); err != nil {
		return Result{}, fmt.Errorf("queue demo poison-rejection worker: %w", err)
	}
	mu.Lock()
	observed := poison
	mu.Unlock()
	if observed.Name() != queue.EventPoisonEnvelope || observed.Connection != connection || observed.Driver != "rabbitmq" ||
		observed.Queue != queueName || observed.Action != queue.PoisonEnvelopeActionReject || observed.BodySize != len(body) ||
		observed.BodyEncoding != "base64" || observed.Error == "" || observed.Timestamp.IsZero() {
		return Result{}, fmt.Errorf("queue demo poison-rejection event = %#v", observed)
	}
	return Result{
		Case: "poison-rejection", Connection: connection, Queue: queueName,
		Processed: true, Steps: []string{"message:injected", "action:reject", "queue:unblocked"},
	}, nil
}
