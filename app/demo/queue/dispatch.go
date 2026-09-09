package queuedemo

import (
	"context"
	"fmt"
	"time"

	"github.com/prismgo/framework/queue"

	jobs "prismgo-demo/app/jobs/queuedemo"
)

func runDispatch(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	traceID := fmt.Sprintf("dispatch-%d", time.Now().UnixNano())
	queueName := fmt.Sprintf("demo-dispatch-%d", time.Now().UnixNano())
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)

	dispatch := func(job *jobs.DispatchJob, options ...queue.DispatchOption) (string, error) {
		base := queue.OnConnection(connection).OnQueue(queueName)
		return manager.Dispatch(ctx, job, append([]queue.DispatchOption{base}, options...)...)
	}
	jobID, err := dispatch(&jobs.DispatchJob{TraceID: traceID, Label: "basic"})
	if err != nil {
		return Result{}, fmt.Errorf("queue demo dispatch basic on %s: %w", connection, err)
	}

	delay := time.Duration(0)
	delaySeconds := 0
	if connection == "redis" {
		delay = 100 * time.Millisecond
		delaySeconds = 1
	}
	if connection == "rabbitmq" {
		delay = 5 * time.Second
		delaySeconds = 5
	}
	for _, item := range []struct {
		label  string
		option queue.DispatchOption
	}{
		{label: "delay", option: queue.Delay(delay)},
		// The Later facade delegates to Dispatch with DelaySeconds; use the
		// supplied manager here so the demo remains isolated and reusable.
		{label: "later", option: queue.DelaySeconds(delaySeconds)},
		{label: "delay-seconds", option: queue.OnQueue(queueName).DelaySeconds(delaySeconds)},
	} {
		if _, err := dispatch(&jobs.DispatchJob{TraceID: traceID, Label: item.label}, item.option); err != nil {
			return Result{}, fmt.Errorf("queue demo dispatch %s on %s: %w", item.label, connection, err)
		}
	}
	if _, err := dispatch(
		&jobs.DispatchJob{TraceID: traceID, Label: "timeout", CheckDeadline: true},
		queue.Tries(3).Timeout(2*time.Second).Backoff(time.Second, 2*time.Second).RetryUntil(time.Now().Add(time.Minute)),
	); err != nil {
		return Result{}, fmt.Errorf("queue demo dispatch options on %s: %w", connection, err)
	}

	if connection != "sync" {
		wait := delay
		if secondsDelay := time.Duration(delaySeconds) * time.Second; secondsDelay > wait {
			wait = secondsDelay
		}
		if wait > 0 {
			timer := time.NewTimer(wait + 300*time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return Result{}, ctx.Err()
			case <-timer.C:
			}
		}
		if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
			Connection: connection, Queues: []string{queueName}, MaxJobs: 5, StopWhenEmpty: true,
		}); err != nil {
			return Result{}, fmt.Errorf("queue demo dispatch worker on %s: %w", connection, err)
		}
	}
	if connection == "sync" {
		if _, err := dispatch(&jobs.DispatchJob{TraceID: traceID, Label: "sync-error", ReturnError: true}); err == nil {
			return Result{}, fmt.Errorf("queue demo dispatch sync error was not returned")
		}
		jobs.RecordTrace(traceID, "sync-error:returned")
	}
	steps := jobs.TakeTrace(traceID)
	return Result{Case: "dispatch", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}
