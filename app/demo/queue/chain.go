package queuedemo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/prismgo/framework/queue"

	jobs "prismgo-demo/app/jobs/queuedemo"
)

func runChain(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	traceID := fmt.Sprintf("chain-%d", time.Now().UnixNano())
	queueName := fmt.Sprintf("demo-chain-%d", time.Now().UnixNano())
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)
	successOptions := []queue.DispatchOption{
		queue.OnConnection(connection), queue.OnQueue(queueName), queue.Tries(1),
	}
	jobID, err := manager.Chain(
		&jobs.ChainJob{TraceID: traceID, Label: "success:extract"},
		&jobs.ChainJob{TraceID: traceID, Label: "success:convert"},
		&jobs.ChainJob{TraceID: traceID, Label: "success:notify"},
	).Options(successOptions...).Dispatch(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo successful chain on %s: %w", connection, err)
	}
	failureQueue := queueName + "-failure"
	_, failureErr := manager.Chain(
		&jobs.ChainJob{TraceID: traceID, Label: "failure:first", Fail: true},
		&jobs.ChainJob{TraceID: traceID, Label: "failure:must-not-run"},
	).Options(queue.OnConnection(connection), queue.OnQueue(failureQueue), queue.Tries(1)).Dispatch(ctx)
	if connection == "sync" {
		if failureErr == nil {
			return Result{}, fmt.Errorf("queue demo failed chain on sync did not return its error")
		}
	}
	if connection != "sync" && failureErr != nil {
		return Result{}, fmt.Errorf("queue demo failed chain dispatch on %s: %w", connection, failureErr)
	}
	if connection != "sync" {
		if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
			Connection: connection, Queues: []string{queueName, failureQueue}, MaxJobs: 4, StopWhenEmpty: true, Tries: 1,
		}); err != nil {
			return Result{}, fmt.Errorf("queue demo chain worker on %s: %w", connection, err)
		}
	}
	steps := jobs.TakeTrace(traceID)
	for _, step := range steps {
		if strings.Contains(step, "must-not-run") {
			return Result{}, fmt.Errorf("queue demo failed chain dispatched a later job: %v", steps)
		}
	}
	jobs.RecordTrace(traceID, "failure:stopped")
	steps = append(steps, "failure:stopped")
	return Result{Case: "chain", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}
