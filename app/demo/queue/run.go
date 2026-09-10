package queuedemo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/prismgo/framework/cache"
	"github.com/prismgo/framework/queue"

	jobs "prismgo-demo/app/jobs/queuedemo"
)

const basicQueue = "demo-basic"

// Result is the public observation returned by a queue demo scenario.
type Result struct {
	Case       string   `json:"case"`
	Connection string   `json:"connection"`
	Queue      string   `json:"queue"`
	JobID      string   `json:"job_id"`
	Processed  bool     `json:"processed"`
	Steps      []string `json:"steps,omitempty"`
}

// Run executes one queue demo scenario through the supplied Manager.
func Run(ctx context.Context, manager *queue.Manager, caseName, connection string) (Result, error) {
	if manager == nil {
		return Result{}, fmt.Errorf("queue demo: manager is nil")
	}
	if connection == "" {
		connection = "sync"
	}
	if caseName == "driver-prerequisites" {
		return runDriverPrerequisites(connection), nil
	}
	if caseName == "config" {
		return runConfiguration(connection)
	}
	if caseName == "payload-encoding" {
		return runPayloadEncoding(connection)
	}
	if caseName == "sync-connection" {
		return runSyncConnection(ctx, manager, connection)
	}
	if caseName == "failed-store" {
		return runFailedStore(ctx, manager, connection)
	}
	if caseName == "batch-store" {
		return runBatchStore(ctx, manager, connection)
	}
	if caseName == "restart-store" {
		return runRestartStore(ctx, manager, connection)
	}
	if caseName == "strategies" {
		return runStrategies(ctx, manager, connection)
	}
	if caseName == "unique" {
		return runUnique(ctx, manager, connection)
	}
	if caseName == "debounce" {
		return runDebounce(ctx, manager, connection)
	}
	if caseName == "middleware" {
		return runMiddleware(ctx, manager, connection)
	}
	if caseName == "dispatch" {
		return runDispatch(ctx, manager, connection)
	}
	if caseName == "chain" {
		return runChain(ctx, manager, connection)
	}
	if caseName == "batch" {
		return runBatch(ctx, manager, connection)
	}
	if caseName == "worker" {
		return runWorker(ctx, manager, connection)
	}
	if caseName == "failure" {
		return runFailure(ctx, manager, connection)
	}
	if caseName == "failed-commands" {
		return runFailedCommands(ctx, manager, connection)
	}
	if caseName == "restart" {
		return runRestart(ctx, manager, connection)
	}
	if caseName == "events" {
		return runEvents(ctx, manager, connection)
	}
	if caseName == "encryption" {
		return runEncryption(ctx, connection)
	}
	if caseName == "custom-driver" {
		return runCustomDriver(ctx, connection)
	}
	if caseName == "errors" {
		return runErrors(ctx, manager, connection)
	}
	if caseName == "redis" {
		return runRedisBoundaries(ctx, manager, connection)
	}
	if caseName == "rabbitmq" {
		return runRabbitMQBoundaries(ctx, manager, connection)
	}
	if caseName != "basic" {
		return Result{}, fmt.Errorf("queue demo: unknown scenario %q", caseName)
	}

	queueName := fmt.Sprintf("%s-%d", basicQueue, time.Now().UnixNano())
	jobID, err := manager.Dispatch(
		ctx,
		&jobs.BasicJob{Message: "hello from PrismGo queue demo"},
		queue.OnConnection(connection),
		queue.OnQueue(queueName),
	)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo basic dispatch on %s: %w", connection, err)
	}
	if connection != "sync" {
		if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
			Connection: connection,
			Queues:     []string{queueName},
			Once:       true,
		}); err != nil {
			return Result{}, fmt.Errorf("queue demo basic worker on %s: %w", connection, err)
		}
		if connection == "redis" {
			queueConnection, resolveErr := manager.Queue(connection)
			if resolveErr != nil {
				return Result{}, resolveErr
			}
			if clearErr := queueConnection.Clear(ctx, queueName); clearErr != nil {
				return Result{}, fmt.Errorf("queue demo basic cleanup on redis: %w", clearErr)
			}
		}
	}
	return Result{Case: caseName, Connection: connection, Queue: queueName, JobID: jobID, Processed: true}, nil
}

func runMiddleware(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo middleware currently uses sync for deterministic concurrency boundaries, got %s", connection)
	}
	traceID := fmt.Sprintf("middleware-%d", time.Now().UnixNano())
	queueName := fmt.Sprintf("demo-middleware-%d", time.Now().UnixNano())
	store := cache.DefaultName()
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)
	manager.UseMiddleware(queue.MiddlewareFunc(func(ctx context.Context, job queue.Job, next queue.Next) error {
		middlewareJob, ok := job.(*jobs.MiddlewareJob)
		if !ok || middlewareJob.TraceID != traceID || middlewareJob.Mode != "order" {
			return next(ctx)
		}
		jobs.RecordTrace(traceID, "global:before")
		err := next(ctx)
		jobs.RecordTrace(traceID, "global:after")
		return err
	}))

	dispatch := func(label, mode string, block bool) (string, error) {
		return manager.Dispatch(ctx, &jobs.MiddlewareJob{
			TraceID: traceID, Label: label, Mode: mode, Store: store, Block: block,
		}, queue.OnConnection(connection), queue.OnQueue(queueName))
	}
	jobID, err := dispatch("order", "order", false)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo middleware order: %w", err)
	}
	for _, item := range []struct{ label, mode string }{{"skip", "skip"}, {"custom-skip", "custom-skip"}} {
		if _, err := dispatch(item.label, item.mode, false); err != nil {
			return Result{}, fmt.Errorf("queue demo middleware %s: %w", item.mode, err)
		}
	}
	if _, err := dispatch("rate:first", "rate", false); err != nil {
		return Result{}, fmt.Errorf("queue demo middleware rate first: %w", err)
	}
	_, err = dispatch("rate:second", "rate", false)
	if err == nil {
		return Result{}, fmt.Errorf("queue demo middleware rate second: expected release")
	}
	if _, ok := queue.ReleaseDelay(err); !ok {
		return Result{}, fmt.Errorf("queue demo middleware rate second: %w", err)
	}

	jobs.PrepareGate(traceID)
	defer jobs.ReleaseGate(traceID)
	firstDone := make(chan error, 1)
	go func() {
		_, err := dispatch("overlap:first", "overlap", true)
		firstDone <- err
	}()
	if err := jobs.WaitGateStarted(ctx, traceID); err != nil {
		return Result{}, fmt.Errorf("queue demo middleware overlap wait: %w", err)
	}
	if _, err := dispatch("overlap:second", "overlap", false); err != nil {
		return Result{}, fmt.Errorf("queue demo middleware overlap second: %w", err)
	}
	jobs.ReleaseGate(traceID)
	if err := <-firstDone; err != nil {
		return Result{}, fmt.Errorf("queue demo middleware overlap first: %w", err)
	}

	if _, err := dispatch("throttle:first", "throttle", false); err == nil {
		return Result{}, fmt.Errorf("queue demo middleware throttle first: expected job error")
	}
	_, err = dispatch("throttle:second", "throttle", false)
	if err == nil {
		return Result{}, fmt.Errorf("queue demo middleware throttle second: expected release")
	}
	if _, ok := queue.ReleaseDelay(err); !ok {
		return Result{}, fmt.Errorf("queue demo middleware throttle second: %w", err)
	}

	steps := jobs.TakeTrace(traceID)
	wantOrder := []string{"global:before", "job:before", "order:handled", "job:after", "global:after"}
	if len(steps) < len(wantOrder) || strings.Join(steps[:len(wantOrder)], "\x00") != strings.Join(wantOrder, "\x00") {
		return Result{}, fmt.Errorf("queue demo middleware order %v, want prefix %v", steps, wantOrder)
	}
	for _, expected := range []string{"rate:first:handled", "overlap:first:handled", "throttle:first:handled"} {
		if !resultHasStep(steps, expected) {
			return Result{}, fmt.Errorf("queue demo middleware: missing %s in %v", expected, steps)
		}
	}
	for _, unexpected := range []string{"skip:handled", "custom-skip:handled", "rate:second:handled", "overlap:second:handled", "throttle:second:handled"} {
		if resultHasStep(steps, unexpected) {
			return Result{}, fmt.Errorf("queue demo middleware: unexpected %s in %v", unexpected, steps)
		}
	}
	return Result{Case: "middleware", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func resultHasStep(steps []string, expected string) bool {
	for _, step := range steps {
		if step == expected {
			return true
		}
	}
	return false
}

func runDebounce(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo debounce requires redis or rabbitmq, got %s", connection)
	}
	traceID := fmt.Sprintf("debounce-%d", time.Now().UnixNano())
	queueName := fmt.Sprintf("demo-debounce-%d", time.Now().UnixNano())
	store := cache.DefaultName()
	window := 100 * time.Millisecond
	if connection == "rabbitmq" {
		window = 5 * time.Second
	}
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)

	var lastJobID string
	for _, label := range []string{"old", "new"} {
		jobID, err := manager.Dispatch(ctx, &jobs.DebounceJob{TraceID: traceID, Label: label, Store: store, Window: window},
			queue.OnConnection(connection), queue.OnQueue(queueName))
		if err != nil {
			return Result{}, fmt.Errorf("queue demo debounce dispatch %s on %s: %w", label, connection, err)
		}
		lastJobID = jobID
	}
	time.Sleep(window + 300*time.Millisecond)
	if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
		Connection: connection, Queues: []string{queueName}, MaxJobs: 2, StopWhenEmpty: true,
	}); err != nil {
		return Result{}, fmt.Errorf("queue demo debounce worker on %s: %w", connection, err)
	}
	if connection == "redis" {
		queueConnection, err := manager.Queue(connection)
		if err != nil {
			return Result{}, err
		}
		if err := queueConnection.Clear(ctx, queueName); err != nil {
			return Result{}, fmt.Errorf("queue demo debounce cleanup on redis: %w", err)
		}
	}
	steps := jobs.TakeTrace(traceID)
	processed := len(steps) == 1 && steps[0] == "new:handled"
	if !processed {
		return Result{}, fmt.Errorf("queue demo debounce on %s: handled steps %v, want only new:handled", connection, steps)
	}
	return Result{Case: "debounce", Connection: connection, Queue: queueName, JobID: lastJobID, Processed: processed, Steps: steps}, nil
}

func runUnique(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "sync" {
		return runUniqueWorker(ctx, manager, connection)
	}
	traceID := fmt.Sprintf("unique-%d", time.Now().UnixNano())
	queueName := fmt.Sprintf("demo-unique-%d", time.Now().UnixNano())
	jobs.ResetTrace(traceID)
	jobs.PrepareGate(traceID)
	defer jobs.ReleaseGate(traceID)
	defer jobs.TakeTrace(traceID)

	type dispatchResult struct {
		id  string
		err error
	}
	firstDone := make(chan dispatchResult, 1)
	go func() {
		id, err := manager.Dispatch(ctx, &jobs.UniqueJob{
			TraceID: traceID, Label: "first", Store: "memory", Block: true,
		}, queue.OnConnection(connection), queue.OnQueue(queueName))
		firstDone <- dispatchResult{id: id, err: err}
	}()
	if err := jobs.WaitGateStarted(ctx, traceID); err != nil {
		return Result{}, fmt.Errorf("queue demo unique wait for first job: %w", err)
	}
	if _, err := manager.Dispatch(ctx, &jobs.UniqueJob{
		TraceID: traceID, Label: "duplicate", Store: "memory",
	}, queue.OnConnection(connection), queue.OnQueue(queueName)); !errors.Is(err, queue.ErrDuplicate) {
		return Result{}, fmt.Errorf("queue demo unique duplicate: got %v, want %w", err, queue.ErrDuplicate)
	}
	jobs.RecordTrace(traceID, "duplicate:rejected")
	jobs.ReleaseGate(traceID)
	first := <-firstDone
	if first.err != nil {
		return Result{}, fmt.Errorf("queue demo unique first dispatch: %w", first.err)
	}
	if _, err := manager.Dispatch(ctx, &jobs.UniqueJob{
		TraceID: traceID, Label: "after-completion", Store: "memory",
	}, queue.OnConnection(connection), queue.OnQueue(queueName)); err != nil {
		return Result{}, fmt.Errorf("queue demo unique after completion: %w", err)
	}
	jobs.RecordTrace(traceID, "after-completion:accepted")
	steps := jobs.TakeTrace(traceID)
	return Result{Case: "unique", Connection: connection, Queue: queueName, JobID: first.id, Processed: true, Steps: steps}, nil
}

func runUniqueWorker(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "redis" && connection != "rabbitmq" {
		return Result{}, fmt.Errorf("queue demo unique requires sync, redis, or rabbitmq, got %s", connection)
	}
	traceID := fmt.Sprintf("unique-%d", time.Now().UnixNano())
	queueName := fmt.Sprintf("demo-unique-%d", time.Now().UnixNano())
	store := cache.DefaultName()
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)
	dispatch := func(label, key string, block, until bool) (string, error) {
		return manager.Dispatch(ctx, &jobs.UniqueJob{
			TraceID: traceID, UniqueKey: key, Label: label, Store: store, Block: block, UntilProcessing: until,
		}, queue.OnConnection(connection), queue.OnQueue(queueName))
	}
	workOnce := func() error {
		return queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
			Connection: connection, Queues: []string{queueName}, Once: true,
		})
	}

	jobID, err := dispatch("first", traceID+":regular", false, false)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo unique first dispatch on %s: %w", connection, err)
	}
	if _, err := dispatch("duplicate", traceID+":regular", false, false); !errors.Is(err, queue.ErrDuplicate) {
		return Result{}, fmt.Errorf("queue demo unique duplicate on %s: got %v, want %w", connection, err, queue.ErrDuplicate)
	}
	jobs.RecordTrace(traceID, "duplicate:rejected")
	if err := workOnce(); err != nil {
		return Result{}, fmt.Errorf("queue demo unique first worker on %s: %w", connection, err)
	}
	if _, err := dispatch("after-completion", traceID+":regular", false, false); err != nil {
		return Result{}, fmt.Errorf("queue demo unique after completion on %s: %w", connection, err)
	}
	jobs.RecordTrace(traceID, "after-completion:accepted")
	if err := workOnce(); err != nil {
		return Result{}, fmt.Errorf("queue demo unique second worker on %s: %w", connection, err)
	}

	jobs.PrepareGate(traceID)
	defer jobs.ReleaseGate(traceID)
	if _, err := dispatch("until:first", traceID+":until", true, true); err != nil {
		return Result{}, fmt.Errorf("queue demo unique until-processing first dispatch on %s: %w", connection, err)
	}
	workerDone := make(chan error, 1)
	go func() { workerDone <- workOnce() }()
	if err := jobs.WaitGateStarted(ctx, traceID); err != nil {
		return Result{}, fmt.Errorf("queue demo unique until-processing wait on %s: %w", connection, err)
	}
	if _, err := dispatch("until:second", traceID+":until", false, true); err != nil {
		return Result{}, fmt.Errorf("queue demo unique until-processing second dispatch on %s: %w", connection, err)
	}
	jobs.RecordTrace(traceID, "until-processing:accepted")
	jobs.ReleaseGate(traceID)
	if err := <-workerDone; err != nil {
		return Result{}, fmt.Errorf("queue demo unique until-processing first worker on %s: %w", connection, err)
	}
	if err := workOnce(); err != nil {
		return Result{}, fmt.Errorf("queue demo unique until-processing second worker on %s: %w", connection, err)
	}
	queueConnection, err := manager.Queue(connection)
	if err != nil {
		return Result{}, err
	}
	if err := queueConnection.Clear(ctx, queueName); err != nil {
		return Result{}, fmt.Errorf("queue demo unique cleanup on %s: %w", connection, err)
	}
	steps := jobs.TakeTrace(traceID)
	return Result{Case: "unique", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}

func runStrategies(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	traceID := fmt.Sprintf("strategies-%d", time.Now().UnixNano())
	queueName := fmt.Sprintf("demo-strategies-%d", time.Now().UnixNano())
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)

	manager.UseMiddleware(queue.MiddlewareFunc(func(ctx context.Context, job queue.Job, next queue.Next) error {
		strategyJob, ok := job.(*jobs.StrategyJob)
		if !ok || strategyJob.TraceID != traceID {
			return next(ctx)
		}
		jobs.RecordTrace(traceID, "global:before")
		err := next(ctx)
		jobs.RecordTrace(traceID, "global:after")
		return err
	}))

	jobID, err := manager.Dispatch(ctx, &jobs.StrategyJob{TraceID: traceID},
		queue.OnConnection(connection),
		queue.OnQueue(queueName),
		queue.Delay(time.Nanosecond),
		queue.Tries(3),
		queue.MaxExceptions(2),
		queue.Timeout(5*time.Second),
		queue.Backoff(time.Second, 2*time.Second),
		queue.RetryUntil(time.Now().Add(10*time.Minute)),
		queue.Tags("queue-demo", "dispatch-override"),
	)
	if err != nil {
		return Result{}, fmt.Errorf("queue demo strategies dispatch on %s: %w", connection, err)
	}
	if connection != "sync" {
		if err := queue.NewWorker(manager).Work(ctx, queue.WorkerOptions{
			Connection: connection, Queues: []string{queueName}, Once: true,
		}); err != nil {
			return Result{}, fmt.Errorf("queue demo strategies worker on %s: %w", connection, err)
		}
	}
	steps := jobs.TakeTrace(traceID)
	want := []string{"global:before", "job:before", "handle", "job:after", "global:after"}
	if strings.Join(steps, "\x00") != strings.Join(want, "\x00") {
		return Result{}, fmt.Errorf("queue demo strategies middleware order %v, want %v", steps, want)
	}
	return Result{Case: "strategies", Connection: connection, Queue: queueName, JobID: jobID, Processed: true, Steps: steps}, nil
}
