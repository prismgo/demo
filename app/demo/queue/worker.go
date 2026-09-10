package queuedemo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	queuecommand "github.com/prismgo/framework/cmd/queue"
	"github.com/prismgo/framework/console"
	"github.com/prismgo/framework/queue"

	jobs "prismgo-demo/app/jobs/queuedemo"
)

func runWorkerCommand(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo worker-command requires redis or rabbitmq, got %s", connection)
	}
	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("worker-command-%d", runID)
	queueName := fmt.Sprintf("demo-worker-command-%d", runID)
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)
	jobID, err := manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: traceID, Label: "command:handled"},
		queue.OnConnection(connection), queue.OnQueue(queueName))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo worker-command dispatch on %s: %w", connection, err)
	}

	command := queuecommand.NewWorkCommand()
	definition := command.Definition()
	if definition.Name != "queue" || len(definition.Aliases) != 1 || definition.Aliases[0] != "queue:work" {
		return Result{}, fmt.Errorf("queue demo worker-command definition = %#v, want queue with queue:work alias", definition)
	}
	var output bytes.Buffer
	commandCtx := console.NewCommandContext(
		ctx,
		command,
		*definition,
		queueWorkerInput{
			arguments: map[string]string{"connection": connection},
			options:   map[string]string{"queue": queueName},
			bools:     map[string]bool{"once": true},
		},
		console.NewIO(strings.NewReader(""), &output, io.Discard),
		nil,
		nil,
	)
	if err := command.Handle(commandCtx); err != nil {
		return Result{}, fmt.Errorf("queue demo worker-command handle on %s: %w", connection, err)
	}
	if err := clearRedisQueue(ctx, manager, connection, queueName); err != nil {
		return Result{}, err
	}
	steps := jobs.TakeTrace(traceID)
	if len(steps) != 1 || steps[0] != "command:handled" {
		return Result{}, fmt.Errorf("queue demo worker-command on %s steps = %v, want command:handled", connection, steps)
	}
	steps = append([]string{
		"command:" + definition.Name,
		"alias:" + definition.Aliases[0],
		"output:" + strings.TrimSpace(output.String()),
	}, steps...)
	return Result{Case: "worker-command", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

// queueWorkerInput supplies the documented queue CLI arguments through the public console input seam.
type queueWorkerInput struct {
	arguments map[string]string
	options   map[string]string
	bools     map[string]bool
}

func (i queueWorkerInput) Argument(name string) string { return i.arguments[name] }

func (i queueWorkerInput) Arguments(name string) []string {
	value := i.Argument(name)
	if value == "" {
		return nil
	}
	return []string{value}
}

func (i queueWorkerInput) Option(name string) string { return i.options[name] }

func (i queueWorkerInput) OptionStrings(name string) []string {
	value := i.Option(name)
	if value == "" {
		return nil
	}
	return []string{value}
}

func (i queueWorkerInput) OptionBool(name string) bool { return i.bools[name] }

func (i queueWorkerInput) OptionInt(string) (int, error) { return 0, nil }

func (i queueWorkerInput) HasOption(name string) bool {
	if _, ok := i.options[name]; ok {
		return true
	}
	_, ok := i.bools[name]
	return ok
}

func runExpirationPolicies(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo expiration requires redis or rabbitmq, got %s", connection)
	}
	runID := time.Now().UnixNano()
	traceID := fmt.Sprintf("expiration-%d", runID)
	queueName := fmt.Sprintf("demo-expiration-%d", runID)
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)

	jobID, err := manager.Dispatch(ctx, &jobs.ControlJob{TraceID: traceID, Label: "expired", Mode: "error"},
		queue.OnConnection(connection), queue.OnQueue(queueName), queue.Tries(3), queue.RetryUntil(time.Now().Add(-time.Second)))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo expiration retry-until dispatch on %s: %w", connection, err)
	}
	if _, err := manager.Dispatch(ctx, &jobs.ControlJob{TraceID: traceID, Label: "timeout", Mode: "timeout"},
		queue.OnConnection(connection), queue.OnQueue(queueName), queue.Tries(3), queue.Timeout(time.Second), queue.FailOnTimeout()); err != nil {
		return Result{}, fmt.Errorf("queue demo expiration timeout dispatch on %s: %w", connection, err)
	}
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
		Connection: connection, Queues: []string{queueName}, MaxJobs: 2, Tries: 3, TimeoutGrace: 100 * time.Millisecond,
	}); err != nil {
		return Result{}, fmt.Errorf("queue demo expiration worker on %s: %w", connection, err)
	}
	if err := clearRedisQueue(ctx, manager, connection, queueName); err != nil {
		return Result{}, err
	}
	steps := jobs.TakeTrace(traceID)
	for _, expected := range []string{"expired:failed-callback", "timeout:cancelled", "timeout:failed-callback"} {
		if !resultHasStep(steps, expected) {
			return Result{}, fmt.Errorf("queue demo expiration on %s missing %q: %v", connection, expected, steps)
		}
	}
	if countResultStep(steps, "expired:attempt") != 1 || countResultStep(steps, "timeout:attempt") != 1 {
		return Result{}, fmt.Errorf("queue demo expiration attempt counts on %s = %v, want one each", connection, steps)
	}
	return Result{Case: "expiration", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func runWorker(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo worker requires redis or rabbitmq, got %s", connection)
	}
	traceID := fmt.Sprintf("worker-%d", time.Now().UnixNano())
	queueBase := fmt.Sprintf("demo-worker-%d", time.Now().UnixNano())
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)
	worker := queue.NewWorker(manager)
	dispatch := func(queueName, label string, waitForCancel bool) (string, error) {
		return manager.Dispatch(ctx, &jobs.WorkerJob{TraceID: traceID, Label: label, WaitForCancel: waitForCancel},
			queue.OnConnection(connection), queue.OnQueue(queueName), queue.Tries(1))
	}
	work := func(label string, options queue.WorkerOptions) error {
		options.Connection = connection
		if err := worker.Work(ctx, options); err != nil {
			return fmt.Errorf("queue demo worker %s on %s: %w", label, connection, err)
		}
		return nil
	}
	clearRedis := func(queueNames ...string) error {
		if connection != "redis" {
			return nil
		}
		queueConn, err := manager.Queue(connection)
		if err != nil {
			return err
		}
		for _, queueName := range queueNames {
			if err := queueConn.Clear(ctx, queueName); err != nil {
				return err
			}
		}
		return nil
	}

	onceQueue := queueBase + "-once"
	jobID, err := dispatch(onceQueue, "once:first", false)
	if err != nil {
		return Result{}, err
	}
	if _, err := dispatch(onceQueue, "once:must-remain", false); err != nil {
		return Result{}, err
	}
	if err := work("once", queue.WorkerOptions{Queues: []string{onceQueue}, Once: true}); err != nil {
		return Result{}, err
	}
	jobs.RecordTrace(traceID, "once:stopped")
	if err := clearRedis(onceQueue); err != nil {
		return Result{}, err
	}

	emptyQueue := queueBase + "-empty"
	for _, label := range []string{"empty:one", "empty:two"} {
		if _, err := dispatch(emptyQueue, label, false); err != nil {
			return Result{}, err
		}
	}
	if err := work("stop-when-empty", queue.WorkerOptions{Queues: []string{emptyQueue}, StopWhenEmpty: true}); err != nil {
		return Result{}, err
	}
	jobs.RecordTrace(traceID, "empty:stopped")

	maxJobsQueue := queueBase + "-max-jobs"
	for _, label := range []string{"max-jobs:one", "max-jobs:two", "max-jobs:must-remain"} {
		if _, err := dispatch(maxJobsQueue, label, false); err != nil {
			return Result{}, err
		}
	}
	if err := work("max-jobs", queue.WorkerOptions{Queues: []string{maxJobsQueue}, MaxJobs: 2}); err != nil {
		return Result{}, err
	}
	jobs.RecordTrace(traceID, "max-jobs:stopped")
	if err := clearRedis(maxJobsQueue); err != nil {
		return Result{}, err
	}

	maxTimeQueue := queueBase + "-max-time"
	maxTimeStart := time.Now()
	if err := work("max-time", queue.WorkerOptions{Queues: []string{maxTimeQueue}, MaxTime: 20 * time.Millisecond, Sleep: 5 * time.Millisecond}); err != nil {
		return Result{}, err
	}
	if time.Since(maxTimeStart) < 20*time.Millisecond {
		return Result{}, fmt.Errorf("queue demo worker max-time stopped before its configured boundary")
	}
	jobs.RecordTrace(traceID, "max-time:stopped")

	highQueue, lowQueue := queueBase+"-high", queueBase+"-low"
	if _, err := dispatch(lowQueue, "priority:low", false); err != nil {
		return Result{}, err
	}
	if _, err := dispatch(highQueue, "priority:high", false); err != nil {
		return Result{}, err
	}
	if err := work("priority", queue.WorkerOptions{Queues: []string{highQueue, lowQueue}, MaxJobs: 2, StopWhenEmpty: true}); err != nil {
		return Result{}, err
	}

	timeoutQueue := queueBase + "-timeout"
	if _, err := dispatch(timeoutQueue, "timeout", true); err != nil {
		return Result{}, err
	}
	if err := work("timeout", queue.WorkerOptions{
		Queues: []string{timeoutQueue}, Once: true, Timeout: 50 * time.Millisecond, TimeoutGrace: 100 * time.Millisecond, Tries: 1,
	}); err != nil {
		return Result{}, err
	}
	steps := jobs.TakeTrace(traceID)
	for _, absent := range []string{"once:must-remain", "max-jobs:must-remain"} {
		if resultHasStep(steps, absent) {
			return Result{}, fmt.Errorf("queue demo worker processed %q past its stop boundary: %v", absent, steps)
		}
	}
	highIndex, lowIndex := queueStepIndex(steps, "priority:high"), queueStepIndex(steps, "priority:low")
	if highIndex < 0 || lowIndex <= highIndex {
		return Result{}, fmt.Errorf("queue demo worker priority order is invalid: %v", steps)
	}
	for _, required := range []string{"once:first", "empty:one", "empty:two", "timeout:cancelled"} {
		if !resultHasStep(steps, required) {
			return Result{}, fmt.Errorf("queue demo worker missing %q: %v", required, steps)
		}
	}
	return Result{Case: "worker", Connection: connection, Queue: queueBase, JobID: jobID, Processed: true, Steps: steps}, nil
}

func queueStepIndex(steps []string, expected string) int {
	for index, step := range steps {
		if step == expected {
			return index
		}
	}
	return -1
}
