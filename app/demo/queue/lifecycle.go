package queuedemo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	queuecommand "github.com/prismgo/framework/cmd/queue"
	"github.com/prismgo/framework/console"
	"github.com/prismgo/framework/queue"
	"github.com/prismgo/framework/queue/payload"
	"github.com/prismgo/framework/queue/state"

	jobs "prismgo-demo/app/jobs/queuedemo"
)

func runFailedCleanupCommandPaths(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo failed-command-paths uses sync with an in-memory failed store, got %s", connection)
	}
	now := time.Now()
	items := []payload.FailedJob{
		{ID: "demo-cleanup-one", JobID: "job-one", Connection: connection, Queue: "demo-failed-cleanup", JobName: "DemoJob", Error: "first failure", FailedAt: now},
		{ID: "demo-cleanup-two", JobID: "job-two", Connection: connection, Queue: "demo-failed-cleanup", JobName: "DemoJob", Error: "second failure", FailedAt: now.Add(time.Second)},
	}
	for _, item := range items {
		if err := manager.Failed().Record(ctx, item); err != nil {
			return Result{}, fmt.Errorf("queue demo failed-command-paths record %s: %w", item.ID, err)
		}
	}

	listOutput, err := runQueueManagementCommand(ctx, queuecommand.NewFailedCommand(), queueWorkerInput{
		options: map[string]string{"page": "1", "page-size": "10000"},
	})
	if err != nil {
		return Result{}, fmt.Errorf("queue demo failed-command-paths list: %w", err)
	}
	if !strings.Contains(listOutput, items[0].ID) || !strings.Contains(listOutput, items[1].ID) {
		return Result{}, fmt.Errorf("queue demo failed-command-paths list output does not contain both demo records: %q", listOutput)
	}
	forgetOutput, err := runQueueManagementCommand(ctx, queuecommand.NewForgetCommand(), queueWorkerInput{
		arguments: map[string]string{"id": items[0].ID},
	})
	if err != nil {
		return Result{}, fmt.Errorf("queue demo failed-command-paths forget: %w", err)
	}
	if !strings.Contains(forgetOutput, "forgot failed job: "+items[0].ID) {
		return Result{}, fmt.Errorf("queue demo failed-command-paths forget output = %q", forgetOutput)
	}
	if _, err := manager.Failed().Find(ctx, items[0].ID); !errors.Is(err, queue.ErrEmpty) {
		return Result{}, fmt.Errorf("queue demo failed-command-paths forgotten record: got %v, want %w", err, queue.ErrEmpty)
	}
	flushOutput, err := runQueueManagementCommand(ctx, queuecommand.NewFlushCommand(), queueWorkerInput{})
	if err != nil {
		return Result{}, fmt.Errorf("queue demo failed-command-paths flush: %w", err)
	}
	if !strings.Contains(flushOutput, "flushed failed jobs") {
		return Result{}, fmt.Errorf("queue demo failed-command-paths flush output = %q", flushOutput)
	}
	page, err := manager.Failed().Page(ctx, state.PageRequest{Page: 1, PageSize: 10})
	if err != nil || page.Total != 0 {
		return Result{}, fmt.Errorf("queue demo failed-command-paths flushed page: page=%#v err=%v", page, err)
	}
	return Result{
		Case: "failed-command-paths", Connection: connection, Queue: items[0].Queue, JobID: items[0].JobID,
		Processed: true, Steps: []string{"queue:failed:listed", "queue:forget:deleted", "queue:flush:emptied"},
	}, nil
}

func runFailedRetryCommand(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo failed-retry requires redis or rabbitmq, got %s", connection)
	}
	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("failed-retry-%d", runID)
	queueName := fmt.Sprintf("demo-failed-retry-%d", runID)
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)
	jobID, err := manager.Dispatch(ctx, &jobs.OperationsJob{TraceID: traceID, Label: "retry", FailFirst: true},
		queue.OnConnection(connection), queue.OnQueue(queueName), queue.Tries(1))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo failed-retry dispatch on %s: %w", connection, err)
	}
	workerCtx, cancelWorker := context.WithCancel(ctx)
	workerDone := make(chan error, 1)
	go func() {
		workerDone <- queue.NewWorker(manager).Work(workerCtx, queue.WorkerOptions{
			Connection: connection, Queues: []string{queueName}, MaxJobs: 2, Tries: 1,
		})
	}()
	workerStopped := false
	defer func() {
		cancelWorker()
		if !workerStopped {
			<-workerDone
		}
	}()
	page, err := waitForFailedJob(ctx, manager.Failed(), jobID)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo failed-retry archive on %s: %w", connection, err)
	}
	failedID := page.Items[0].ID
	output, err := runQueueManagementCommand(ctx, queuecommand.NewRetryCommand(), queueWorkerInput{
		arguments: map[string]string{"ids": failedID},
	})
	if err != nil {
		return Result{}, fmt.Errorf("queue demo failed-retry command on %s: %w", connection, err)
	}
	if !strings.Contains(output, "retried failed job: "+failedID) {
		return Result{}, fmt.Errorf("queue demo failed-retry command output on %s = %q", connection, output)
	}
	select {
	case err := <-workerDone:
		workerStopped = true
		if err != nil {
			return Result{}, fmt.Errorf("queue demo failed-retry worker on %s: %w", connection, err)
		}
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case <-time.After(5 * time.Second):
		return Result{}, fmt.Errorf("queue demo failed-retry worker on %s did not stop", connection)
	}
	if _, err := manager.Failed().Find(ctx, failedID); !errors.Is(err, queue.ErrEmpty) {
		return Result{}, fmt.Errorf("queue demo failed-retry record on %s: got %v, want %w", connection, err, queue.ErrEmpty)
	}
	steps := jobs.TakeTrace(traceID)
	if countResultStep(steps, "retry:attempt") != 2 || !resultHasStep(steps, "retry:handled") {
		return Result{}, fmt.Errorf("queue demo failed-retry attempts on %s = %v, want two attempts and a handled job", connection, steps)
	}
	steps = append([]string{"queue:retry:requeued"}, steps...)
	return Result{Case: "failed-retry", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func runFailedEvent(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo failed-event requires redis or rabbitmq, got %s", connection)
	}
	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("failed-event-%d", runID)
	queueName := fmt.Sprintf("demo-failed-event-%d", runID)
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)
	jobID, err := manager.Dispatch(ctx, &jobs.OperationsJob{TraceID: traceID, Label: "failed-event", FailFirst: true},
		queue.OnConnection(connection), queue.OnQueue(queueName), queue.Tries(1))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo failed-event dispatch on %s: %w", connection, err)
	}
	var observed queue.JobFailed
	observer := func(eventCtx context.Context, event queue.Event) context.Context {
		failed, ok := event.(queue.JobFailed)
		if ok && failed.Queue == queueName {
			observed = failed
		}
		return eventCtx
	}
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
		Connection: connection, Queues: []string{queueName}, Once: true, Tries: 1, EventObserver: observer,
	}); err != nil {
		return Result{}, fmt.Errorf("queue demo failed-event worker on %s: %w", connection, err)
	}
	if observed.Name() != queue.EventJobFailed || observed.JobID != jobID || observed.Connection != connection || observed.Error == "" {
		return Result{}, fmt.Errorf("queue demo failed-event payload on %s = %#v", connection, observed)
	}
	if err := manager.Failed().Forget(ctx, observed.ID); err != nil {
		return Result{}, fmt.Errorf("queue demo failed-event cleanup on %s: %w", connection, err)
	}
	steps := []string{observed.Name(), "payload:job-id", "payload:error", "callback:invoked"}
	return Result{Case: "failed-event", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func runBatchEvents(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo batch-events requires redis or rabbitmq, got %s", connection)
	}
	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("batch-events-%d", runID)
	queueName := fmt.Sprintf("demo-batch-events-%d", runID)
	completedName := fmt.Sprintf("queue-demo-completed-%d", runID)
	cancelledName := fmt.Sprintf("queue-demo-cancelled-%d", runID)
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)

	var mu sync.Mutex
	eventNames := make([]string, 0, 4)
	observe := func(eventCtx context.Context, event queue.Event) context.Context {
		batchEvent, ok := event.(queue.BatchEvent)
		if !ok {
			return eventCtx
		}
		include := batchEvent.Batch.Name == completedName && batchEvent.Name() != queue.EventBatchCancelled
		include = include || batchEvent.Batch.Name == cancelledName && batchEvent.Name() == queue.EventBatchCancelled
		if include {
			mu.Lock()
			eventNames = append(eventNames, batchEvent.Name())
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
	created, err := manager.Batch(
		&jobs.BatchJob{TraceID: traceID, Label: "events:one"},
		&jobs.BatchJob{TraceID: traceID, Label: "events:two"},
	).Name(completedName).Options(queue.OnConnection(connection), queue.OnQueue(queueName)).Dispatch(ctx)
	queue.UseEventSink(previous)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo batch-events dispatch on %s: %w", connection, err)
	}
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
		Connection: connection, Queues: []string{queueName}, MaxJobs: 2, StopWhenEmpty: true, EventObserver: observe,
	}); err != nil {
		return Result{}, fmt.Errorf("queue demo batch-events worker on %s: %w", connection, err)
	}
	cancelled, err := manager.Batch(&jobs.BatchJob{TraceID: traceID, Label: "cancel:must-not-run"}).
		Name(cancelledName).Options(queue.OnConnection(connection), queue.OnQueue(queueName+"-cancel")).Dispatch(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo batch-events cancellable dispatch on %s: %w", connection, err)
	}
	previous = queue.CurrentEventSink()
	queue.UseEventSink(func(eventCtx context.Context, event queue.Event) {
		observe(eventCtx, event)
		if previous != nil {
			previous(eventCtx, event)
		}
	})
	if err := manager.CancelBatch(ctx, cancelled.ID); err != nil {
		queue.UseEventSink(previous)
		return Result{}, fmt.Errorf("queue demo batch-events cancel on %s: %w", connection, err)
	}
	queue.UseEventSink(previous)
	if err := clearRedisQueue(ctx, manager, connection, queueName+"-cancel"); err != nil {
		return Result{}, err
	}
	mu.Lock()
	steps := append([]string(nil), eventNames...)
	mu.Unlock()
	want := []string{queue.EventBatchCreated, queue.EventBatchUpdated, queue.EventBatchFinished, queue.EventBatchCancelled}
	if strings.Join(steps, "\x00") != strings.Join(want, "\x00") {
		return Result{}, fmt.Errorf("queue demo batch-events on %s: got %v, want %v", connection, steps, want)
	}
	return Result{Case: "batch-events", Connection: connection, Queue: queueName, JobID: created.ID, Processed: true, Steps: steps}, nil
}

func runQueueManagementCommand(ctx context.Context, command console.Command, input queueWorkerInput) (string, error) {
	definition := command.Definition()
	var output bytes.Buffer
	commandCtx := console.NewCommandContext(
		ctx,
		command,
		*definition,
		input,
		console.NewIO(strings.NewReader(""), &output, io.Discard),
		nil,
		nil,
	)
	if err := command.Handle(commandCtx); err != nil {
		return "", err
	}
	return output.String(), nil
}
