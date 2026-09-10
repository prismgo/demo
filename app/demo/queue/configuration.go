package queuedemo

import (
	"context"
	"fmt"
	"time"

	"github.com/prismgo/framework/queue"

	jobs "prismgo-demo/app/jobs/queuedemo"
)

func runDriverPrerequisites(connection string) Result {
	return Result{
		Case:       "driver-prerequisites",
		Connection: connection,
		Processed:  true,
		Steps: []string{
			"sync:none",
			"redis:PRISMGO_REDIS_TEST_URL",
			"rabbitmq:PRISMGO_RABBITMQ_TEST_URL",
		},
	}
}

func runConfiguration(connection string) (Result, error) {
	manager, err := queue.NewManager(queue.Config{
		Default:  "sync",
		Encoding: "msgpack",
		Connections: map[string]queue.ConnectionConfig{
			"sync":     {Driver: "sync", Queue: "default"},
			"redis":    {Driver: "redis", Queue: "default"},
			"rabbitmq": {Driver: "rabbitmq", Queue: "default"},
		},
		Failed:   queue.StateStoreConfig{Driver: "memory"},
		Batching: queue.StateStoreConfig{Driver: "memory"},
		Restart:  queue.RestartConfig{Key: "prismgo:queue:restart"},
	}, queue.NewRegistry())
	if err != nil {
		return Result{}, fmt.Errorf("queue demo configuration: %w", err)
	}
	if err := manager.Close(); err != nil {
		return Result{}, fmt.Errorf("queue demo configuration close: %w", err)
	}
	return Result{
		Case:       "config",
		Connection: connection,
		Processed:  true,
		Steps: []string{
			"default:sync",
			"connections:sync,redis,rabbitmq",
			"state:failed,batching,restart",
		},
	}, nil
}

func runPayloadEncoding(connection string) (Result, error) {
	for _, encoding := range []string{"json", ""} {
		manager, err := queue.NewManager(queue.Config{
			Default:  "sync",
			Encoding: encoding,
			Connections: map[string]queue.ConnectionConfig{
				"sync": {Driver: "sync", Queue: "default"},
			},
		}, queue.NewRegistry())
		if err != nil {
			return Result{}, fmt.Errorf("queue demo payload encoding %q: %w", encoding, err)
		}
		if err := manager.Close(); err != nil {
			return Result{}, fmt.Errorf("queue demo payload encoding %q close: %w", encoding, err)
		}
	}
	return Result{
		Case:       "payload-encoding",
		Connection: connection,
		Processed:  true,
		Steps:      []string{"explicit:json", "inherited:msgpack"},
	}, nil
}

func runSyncConnection(ctx context.Context, manager *queue.Manager, connection string) (Result, error) {
	if connection != "sync" {
		return Result{}, fmt.Errorf("queue demo sync-connection requires sync, got %s", connection)
	}
	traceID := fmt.Sprintf("sync-connection-%d", time.Now().UnixNano())
	queueName := fmt.Sprintf("demo-sync-%d", time.Now().UnixNano())
	jobs.ResetTrace(traceID)
	defer jobs.TakeTrace(traceID)
	jobID, err := manager.Dispatch(ctx, &jobs.OperationsJob{TraceID: traceID, Label: "sync"}, queue.OnConnection(connection), queue.OnQueue(queueName))
	if err != nil {
		return Result{}, fmt.Errorf("queue demo sync connection dispatch: %w", err)
	}
	steps := jobs.TakeTrace(traceID)
	if !resultHasStep(steps, "sync:handled") {
		return Result{}, fmt.Errorf("queue demo sync connection steps = %v, want sync:handled", steps)
	}
	return Result{
		Case:       "sync-connection",
		Connection: connection,
		Queue:      queueName,
		JobID:      jobID,
		Processed:  true,
		Steps:      []string{"queue:configured", "sync:handled"},
	}, nil
}
