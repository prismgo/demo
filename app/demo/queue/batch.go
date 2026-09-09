package queuedemo

import (
	"context"
	"fmt"
	"time"

	"github.com/prismgo/framework/queue"

	jobs "prismgo-demo/app/jobs/queuedemo"
)

func runBatch(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	traceID := fmt.Sprintf("batch-%d", time.Now().UnixNano())
	queueName := fmt.Sprintf("demo-batch-%d", time.Now().UnixNano())
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)

	second := &jobs.BatchJob{TraceID: traceID, Label: "success:two"}
	if connection != "sync" {
		second.Label = "failure:two"
		second.Fail = true
	}
	created, err := manager.Batch(
		&jobs.BatchJob{TraceID: traceID, Label: "success:one"}, second,
	).Name("queue-demo-export").Options(
		queue.OnConnection(connection), queue.OnQueue(queueName), queue.Tries(1),
	).Dispatch(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo batch dispatch on %s: %w", connection, err)
	}
	if connection != "sync" {
		if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
			Connection: connection, Queues: []string{queueName}, MaxJobs: 2, StopWhenEmpty: true, Tries: 1,
		}); err != nil {
			return Result{}, fmt.Errorf("queue demo batch worker on %s: %w", connection, err)
		}
	}
	latest, err := manager.BatchStatus(ctx, created.ID)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo inspect batch on %s: %w", connection, err)
	}
	wantFailed := 0
	if connection != "sync" {
		wantFailed = 1
	}
	if latest.Name != "queue-demo-export" || latest.Total != 2 || latest.Pending != 0 || latest.Processed != 2 || latest.Failed != wantFailed || latest.FinishedAt.IsZero() {
		return Result{}, fmt.Errorf("queue demo batch progress on %s: %#v", connection, latest)
	}
	jobs.RecordTrace(traceID, fmt.Sprintf("progress:%d/%d failed=%d", latest.Processed, latest.Total, latest.Failed))

	cancelQueue := queueName + "-cancel"
	cancelLabel := "cancel:must-not-run"
	if connection == "sync" {
		cancelLabel = "cancel:handled-before-mark"
	}
	cancelled, err := manager.Batch(&jobs.BatchJob{TraceID: traceID, Label: cancelLabel}).
		Name("queue-demo-cancel").Options(queue.OnConnection(connection), queue.OnQueue(cancelQueue)).Dispatch(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo cancellable batch dispatch on %s: %w", connection, err)
	}
	if err := manager.CancelBatch(ctx, cancelled.ID); err != nil {
		return Result{}, fmt.Errorf("queue demo cancel batch on %s: %w", connection, err)
	}
	if connection != "sync" {
		if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
			Connection: connection, Queues: []string{cancelQueue}, Once: true,
		}); err != nil {
			return Result{}, fmt.Errorf("queue demo cancelled batch worker on %s: %w", connection, err)
		}
	}
	cancelStatus, err := manager.BatchStatus(ctx, cancelled.ID)
	if err != nil || !cancelStatus.Cancelled || cancelStatus.CancelledAt.IsZero() {
		return Result{}, fmt.Errorf("queue demo cancelled batch status on %s: status=%#v err=%v", connection, cancelStatus, err)
	}
	steps := jobs.TakeTrace(traceID)
	if connection != "sync" {
		for _, step := range steps {
			if step == "cancel:must-not-run" {
				return Result{}, fmt.Errorf("queue demo cancelled batch executed pending job")
			}
		}
	}
	steps = append(steps, "cancel:marked")
	return Result{Case: "batch", Connection: connection, Queue: queueName, JobID: created.ID, Processed: true, Steps: steps}, nil
}
