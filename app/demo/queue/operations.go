package queuedemo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prismgo/framework/queue"
	"github.com/prismgo/framework/queue/payload"
	"github.com/prismgo/framework/queue/state"

	jobs "prismgo-demo/app/jobs/queuedemo"
)

func runFailure(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo failure requires redis or rabbitmq, got %s", connection)
	}
	traceID := fmt.Sprintf("failure-%d", time.Now().UnixNano())
	queueName := fmt.Sprintf("demo-failure-%d", time.Now().UnixNano())
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)

	jobID, err := manager.Dispatch(ctx, &jobs.OperationsJob{
		TraceID: traceID, Label: "failure", FailFirst: true,
	}, queue.OnConnection(connection), queue.OnQueue(queueName), queue.Tries(1))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo failure dispatch on %s: %w", connection, err)
	}
	worker := queue.NewWorker(manager)
	workerCtx, cancelWorker := context.WithCancel(ctx)
	defer cancelWorker()
	workerDone := make(chan error, 1)
	go func() {
		workerDone <- worker.Work(workerCtx, queue.WorkerOptions{
			Connection: connection, Queues: []string{queueName}, MaxJobs: 2, Tries: 1,
		})
	}()
	page, err := waitForFailedJob(ctx, manager.Failed(), jobID)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo failure list on %s: %w", connection, err)
	}
	if len(page.Items) != 1 || page.Items[0].JobID != jobID || page.Items[0].Error == "" {
		return Result{}, fmt.Errorf("queue demo failure archive on %s: %#v", connection, page.Items)
	}
	failedID := page.Items[0].ID
	jobs.RecordTrace(traceID, "failure:archived")
	if err := queue.NewDispatcher(manager).RetryFailed(ctx, failedID); err != nil {
		return Result{}, fmt.Errorf("queue demo failure retry on %s: %w", connection, err)
	}
	select {
	case err := <-workerDone:
		if err != nil {
			return Result{}, fmt.Errorf("queue demo failure retry worker on %s: %w", connection, err)
		}
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case <-time.After(5 * time.Second):
		return Result{}, fmt.Errorf("queue demo failure retry worker on %s did not stop", connection)
	}
	if _, err := manager.Failed().Find(ctx, failedID); !errors.Is(err, queue.ErrEmpty) {
		return Result{}, fmt.Errorf("queue demo failure retry record on %s: got %v, want %w", connection, err, queue.ErrEmpty)
	}
	steps := jobs.TakeTrace(traceID)
	for _, expected := range []string{"failure:failed-callback", "failure:archived", "failure:handled"} {
		if !resultHasStep(steps, expected) {
			return Result{}, fmt.Errorf("queue demo failure missing %q: %v", expected, steps)
		}
	}
	return Result{Case: "failure", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func waitForFailedJob(ctx context.Context, store queue.FailedStore, jobID string) (state.PageEnvelope[payload.FailedJob], error) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	for {
		page, err := store.Page(ctx, state.PageRequest{Page: 1, PageSize: 10})
		if err != nil {
			return state.PageEnvelope[payload.FailedJob]{}, err
		}
		for _, item := range page.Items {
			if item.JobID == jobID {
				return state.PageEnvelope[payload.FailedJob]{Items: []payload.FailedJob{item}, Total: 1, Page: 1, PageSize: 10}, nil
			}
		}
		select {
		case <-ctx.Done():
			return state.PageEnvelope[payload.FailedJob]{}, ctx.Err()
		case <-timer.C:
			return state.PageEnvelope[payload.FailedJob]{}, fmt.Errorf("timed out waiting for failed job %s", jobID)
		case <-ticker.C:
		}
	}
}

func runFailedCommands(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo failed-commands uses sync with an in-memory failed store, got %s", connection)
	}
	store := manager.Failed()
	now := time.Now()
	items := []payload.FailedJob{
		{ID: "demo-failed-one", JobID: "job-one", Connection: connection, Queue: "demo-failed", JobName: "DemoJob", Error: "first failure", FailedAt: now},
		{ID: "demo-failed-two", JobID: "job-two", Connection: connection, Queue: "demo-failed", JobName: "DemoJob", Error: "second failure", FailedAt: now.Add(time.Second)},
	}
	for _, item := range items {
		if err := store.Record(ctx, item); err != nil {
			return Result{}, fmt.Errorf("queue demo failed-commands record %s: %w", item.ID, err)
		}
	}
	page, err := store.Page(ctx, state.PageRequest{Page: 1, PageSize: 1})
	if err != nil || page.Total != 2 || len(page.Items) != 1 || page.Items[0].ID != items[0].ID {
		return Result{}, fmt.Errorf("queue demo failed-commands page: page=%#v err=%v", page, err)
	}
	if _, err := store.Find(ctx, items[0].ID); err != nil {
		return Result{}, fmt.Errorf("queue demo failed-commands find: %w", err)
	}
	if err := store.Forget(ctx, items[0].ID); err != nil {
		return Result{}, fmt.Errorf("queue demo failed-commands forget: %w", err)
	}
	if _, err := store.Find(ctx, items[0].ID); !errors.Is(err, queue.ErrEmpty) {
		return Result{}, fmt.Errorf("queue demo failed-commands forgotten record: got %v, want %w", err, queue.ErrEmpty)
	}
	if err := store.Flush(ctx); err != nil {
		return Result{}, fmt.Errorf("queue demo failed-commands flush: %w", err)
	}
	page, err = store.Page(ctx, state.PageRequest{Page: 1, PageSize: 10})
	if err != nil || page.Total != 0 || len(page.Items) != 0 {
		return Result{}, fmt.Errorf("queue demo failed-commands flushed page: page=%#v err=%v", page, err)
	}
	steps := []string{"queue:failed:list", "queue:failed:find", "queue:forget", "queue:flush"}
	return Result{Case: "failed-commands", Connection: connection, Queue: "demo-failed", JobID: items[0].JobID, Processed: true, Steps: steps}, nil
}

func runFailedStore(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" {
		return Result{}, fmt.Errorf("queue demo failed-store requires redis, got %s", connection)
	}
	id := fmt.Sprintf("demo-failed-store-%d", time.Now().UnixNano())
	jobID := fmt.Sprintf("demo-job-%d", time.Now().UnixNano())
	store := manager.Failed()
	failed := payload.FailedJob{
		ID: id, JobID: jobID, Connection: connection, Queue: "demo-failed-store",
		JobName: "OperationsJob", Error: "demonstration failure", FailedAt: time.Now(),
	}
	if err := store.Record(ctx, failed); err != nil {
		return Result{}, fmt.Errorf("queue demo failed store record: %w", err)
	}
	found, err := store.Find(ctx, id)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo failed store find: %w", err)
	}
	if found.JobID != jobID {
		return Result{}, fmt.Errorf("queue demo failed store job ID = %q, want %q", found.JobID, jobID)
	}
	if err := store.Forget(ctx, id); err != nil {
		return Result{}, fmt.Errorf("queue demo failed store forget: %w", err)
	}
	if _, err := store.Find(ctx, id); !errors.Is(err, queue.ErrEmpty) {
		return Result{}, fmt.Errorf("queue demo failed store forgotten record: got %v, want %w", err, queue.ErrEmpty)
	}
	return Result{
		Case: "failed-store", Connection: connection, Queue: failed.Queue, JobID: jobID,
		Processed: true, Steps: []string{"recorded", "found", "forgotten"},
	}, nil
}

func runBatchStore(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" {
		return Result{}, fmt.Errorf("queue demo batch-store requires redis, got %s", connection)
	}
	traceID := fmt.Sprintf("batch-store-%d", time.Now().UnixNano())
	queueName := fmt.Sprintf("demo-batch-store-%d", time.Now().UnixNano())
	status, err := manager.Batch(
		&jobs.OperationsJob{TraceID: traceID, Label: "batch-store"},
	).Name("queue-demo-state-store").Options(
		queue.OnConnection(connection), queue.OnQueue(queueName),
	).Dispatch(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo batch store create: %w", err)
	}
	stored, err := manager.BatchStatus(ctx, status.ID)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo batch store read: %w", err)
	}
	if stored.ID != status.ID || stored.Name != "queue-demo-state-store" || stored.Total != 1 {
		return Result{}, fmt.Errorf("queue demo batch store status = %#v, want created batch %q", stored, status.ID)
	}
	if err := manager.CancelBatch(ctx, status.ID); err != nil {
		return Result{}, fmt.Errorf("queue demo batch store cancel: %w", err)
	}
	cancelled, err := manager.BatchStatus(ctx, status.ID)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo batch store read cancelled: %w", err)
	}
	if !cancelled.Cancelled {
		return Result{}, fmt.Errorf("queue demo batch store cancelled = false, want true")
	}
	if err := clearRedisQueue(ctx, manager, connection, queueName); err != nil {
		return Result{}, err
	}
	return Result{
		Case: "batch-store", Connection: connection, Queue: queueName, JobID: status.ID,
		Processed: true, Steps: []string{"created", "read", "cancelled"},
	}, nil
}

func runRestartStore(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" {
		return Result{}, fmt.Errorf("queue demo restart-store requires redis, got %s", connection)
	}
	traceID := fmt.Sprintf("restart-store-%d", time.Now().UnixNano())
	queueName := fmt.Sprintf("demo-restart-store-%d", time.Now().UnixNano())
	jobs.ResetTrace(traceID)
	jobs.PrepareGate(traceID)
	defer jobs.ReleaseGate(traceID)
	defer jobs.TakeTrace(traceID)
	jobID, err := manager.Dispatch(ctx, &jobs.OperationsJob{TraceID: traceID, Label: "restart-store", Block: true},
		queue.OnConnection(connection), queue.OnQueue(queueName))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo restart store dispatch: %w", err)
	}
	workerCtx, cancelWorker := context.WithCancel(ctx)
	workerDone := make(chan error, 1)
	go func() {
		workerDone <- queue.NewWorker(manager).Work(workerCtx, queue.WorkerOptions{
			Connection: connection, Queues: []string{queueName}, StopWhenEmpty: true, Sleep: 5 * time.Millisecond,
		})
	}()
	workerStopped := false
	defer func() {
		cancelWorker()
		if !workerStopped {
			<-workerDone
		}
	}()
	if err := jobs.WaitGateStarted(ctx, traceID); err != nil {
		return Result{}, fmt.Errorf("queue demo restart store wait: %w", err)
	}
	if err := manager.RequestRestart(ctx); err != nil {
		return Result{}, fmt.Errorf("queue demo restart store request: %w", err)
	}
	jobs.ReleaseGate(traceID)
	select {
	case err := <-workerDone:
		workerStopped = true
		if err != nil {
			return Result{}, fmt.Errorf("queue demo restart store worker: %w", err)
		}
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case <-time.After(5 * time.Second):
		return Result{}, fmt.Errorf("queue demo restart store worker did not observe signal")
	}
	if err := clearRedisQueue(ctx, manager, connection, queueName); err != nil {
		return Result{}, err
	}
	return Result{
		Case: "restart-store", Connection: connection, Queue: queueName, JobID: jobID,
		Processed: true, Steps: []string{"requested", "worker-observed"},
	}, nil
}

func runRestart(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo restart requires redis or rabbitmq, got %s", connection)
	}
	traceID := fmt.Sprintf("restart-%d", time.Now().UnixNano())
	queueName := fmt.Sprintf("demo-restart-%d", time.Now().UnixNano())
	jobs.ResetTrace(traceID)
	jobs.PrepareGate(traceID)
	defer jobs.ReleaseGate(traceID)
	defer jobs.TakeTrace(traceID)

	dispatch := func(label string, block bool) (string, error) {
		return manager.Dispatch(ctx, &jobs.OperationsJob{TraceID: traceID, Label: label, Block: block},
			queue.OnConnection(connection), queue.OnQueue(queueName))
	}
	jobID, err := dispatch("restart:current", true)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo restart current dispatch on %s: %w", connection, err)
	}
	if _, err := dispatch("restart:pending", false); err != nil {
		return Result{}, fmt.Errorf("queue demo restart pending dispatch on %s: %w", connection, err)
	}
	workerCtx, cancelWorker := context.WithCancel(ctx)
	defer cancelWorker()
	workerDone := make(chan error, 1)
	go func() {
		workerDone <- queue.NewWorker(manager).Work(workerCtx, queue.WorkerOptions{
			Connection: connection, Queues: []string{queueName}, StopWhenEmpty: true, Sleep: 5 * time.Millisecond,
		})
	}()
	if err := jobs.WaitGateStarted(ctx, traceID); err != nil {
		return Result{}, fmt.Errorf("queue demo restart wait on %s: %w", connection, err)
	}
	if err := manager.RequestRestart(ctx); err != nil {
		return Result{}, fmt.Errorf("queue demo restart signal on %s: %w", connection, err)
	}
	jobs.RecordTrace(traceID, "restart:requested")
	jobs.ReleaseGate(traceID)
	select {
	case err := <-workerDone:
		if err != nil {
			return Result{}, fmt.Errorf("queue demo restart worker on %s: %w", connection, err)
		}
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case <-time.After(5 * time.Second):
		return Result{}, fmt.Errorf("queue demo restart worker on %s did not stop", connection)
	}
	jobs.RecordTrace(traceID, "restart:worker-stopped")
	steps := jobs.TakeTrace(traceID)
	for _, expected := range []string{"restart:current:handled", "restart:requested", "restart:worker-stopped"} {
		if !resultHasStep(steps, expected) {
			return Result{}, fmt.Errorf("queue demo restart missing %q: %v", expected, steps)
		}
	}
	if resultHasStep(steps, "restart:pending:handled") {
		return Result{}, fmt.Errorf("queue demo restart processed pending job after signal: %v", steps)
	}
	if err := clearRedisQueue(ctx, manager, connection, queueName); err != nil {
		return Result{}, err
	}
	return Result{Case: "restart", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func runEvents(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo events requires redis or rabbitmq, got %s", connection)
	}
	traceID := fmt.Sprintf("events-%d", time.Now().UnixNano())
	queueName := fmt.Sprintf("demo-events-%d", time.Now().UnixNano())
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)

	var mu sync.Mutex
	eventNames := make([]string, 0, 5)
	observe := func(eventCtx context.Context, event queue.Event) context.Context {
		if queueEventBelongsTo(event, queueName) {
			mu.Lock()
			eventNames = append(eventNames, event.Name())
			mu.Unlock()
		}
		return eventCtx
	}
	previous := queue.CurrentEventSink()
	queue.UseEventSink(func(eventCtx context.Context, event queue.Event) {
		observe(eventCtx, event)
		if previous != nil {
			previous(eventCtx, event)
		}
	})
	jobID, err := manager.Dispatch(ctx, &jobs.OperationsJob{
		TraceID: traceID, Label: "events", FailFirst: true,
	}, queue.OnConnection(connection), queue.OnQueue(queueName), queue.Tries(2))
	queue.UseEventSink(previous)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo events dispatch on %s: %w", connection, err)
	}
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
		Connection: connection, Queues: []string{queueName}, MaxJobs: 2, StopWhenEmpty: true, Tries: 2, EventObserver: observe,
	}); err != nil {
		return Result{}, fmt.Errorf("queue demo events worker on %s: %w", connection, err)
	}
	mu.Lock()
	steps := append([]string(nil), eventNames...)
	mu.Unlock()
	want := []string{queue.EventJobQueued, queue.EventJobProcessing, queue.EventJobReleased, queue.EventJobProcessing, queue.EventJobProcessed}
	if strings.Join(steps, "\x00") != strings.Join(want, "\x00") {
		return Result{}, fmt.Errorf("queue demo events on %s: got %v, want %v", connection, steps, want)
	}
	return Result{Case: "events", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func queueEventBelongsTo(event queue.Event, queueName string) bool {
	switch item := event.(type) {
	case queue.JobQueued:
		return item.Queue == queueName
	case queue.JobProcessing:
		return item.Queue == queueName
	case queue.JobProcessed:
		return item.Queue == queueName
	case queue.JobReleased:
		return item.Queue == queueName
	case queue.JobFailed:
		return item.Queue == queueName
	default:
		return false
	}
}

func clearRedisQueue(ctx context.Context, manager *queue.Manager, connection, queueName string) error {
	if connection != "redis" {
		return nil
	}
	queueConnection, err := manager.Queue(connection)
	if err != nil {
		return err
	}
	if err := queueConnection.Clear(ctx, queueName); err != nil {
		return fmt.Errorf("queue demo cleanup on redis: %w", err)
	}
	return nil
}
