package demo

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/prismgo/framework/queue"

	qdemo "prismgo-demo/app/demo/queue"
)

func TestQueueDemoListShowsDocumentedScenariosAndConnections(t *testing.T) {
	command := NewQueueCommand()
	var output bytes.Buffer
	input := demoInput{arguments: map[string]string{"case": "list"}}

	if err := command.Handle(commandContext(command, input, &output)); err != nil {
		t.Fatalf("handle demo:queue list: %v", err)
	}

	for _, expected := range []string{
		"basic",
		"Creating Jobs",
		"implemented",
		"sync,redis,rabbitmq",
		"custom-driver",
		"strategy-envelope",
		"unique-options",
		"overlap-release",
		"poison-event",
		"infrastructure-events",
		"rabbitmq",
		"RabbitMQ Configuration",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("output does not contain %q:\n%s", expected, output.String())
		}
	}
}

func TestQueueDemoRunsSelectedScenario(t *testing.T) {
	command := newQueueCommand(func(_ context.Context, caseName, connection string) (qdemo.Result, error) {
		return qdemo.Result{
			Case: caseName, Connection: connection, Queue: "demo-basic", JobID: "job-123", Processed: true,
		}, nil
	})
	var output bytes.Buffer
	input := demoInput{
		arguments: map[string]string{"case": "basic"},
		options:   map[string]string{"connection": "redis"},
	}

	if err := command.Handle(commandContext(command, input, &output)); err != nil {
		t.Fatalf("handle demo:queue basic: %v", err)
	}
	for _, expected := range []string{"basic", "redis", "demo-basic", "job-123", "processed"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("output does not contain %q:\n%s", expected, output.String())
		}
	}
}

func TestQueueDemoRejectsUnsupportedScenarioConnection(t *testing.T) {
	called := false
	command := newQueueCommand(func(context.Context, string, string) (qdemo.Result, error) {
		called = true
		return qdemo.Result{}, nil
	})
	input := demoInput{
		arguments: map[string]string{"case": "middleware"},
		options:   map[string]string{"connection": "redis"},
	}

	err := command.Handle(commandContext(command, input, &bytes.Buffer{}))
	if err == nil || !strings.Contains(err.Error(), "supports sync") {
		t.Fatalf("unsupported connection error = %v", err)
	}
	if called {
		t.Fatal("scenario runner should not be called for an unsupported connection")
	}
}

func TestQueueDemoStateStoreConfigurationErrors(t *testing.T) {
	tests := []struct {
		name     string
		caseName string
		config   queue.Config
		want     string
	}{
		{name: "failed store", caseName: "failed-store", config: queue.Config{}, want: "QUEUE_FAILED_DRIVER=redis"},
		{name: "batch store", caseName: "batch-store", config: queue.Config{}, want: "QUEUE_BATCHING_DRIVER=redis"},
		{name: "restart store", caseName: "restart-store", config: queue.Config{}, want: "QUEUE_RESTART_CACHE"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := queueDemoConfigError(test.caseName, test.config)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("queueDemoConfigError(%q) = %v, want error containing %q", test.caseName, err, test.want)
			}
		})
	}
}

func TestQueueDemoAcceptsConfiguredStateStores(t *testing.T) {
	cfg := queue.Config{
		Failed:   queue.StateStoreConfig{Driver: " Redis "},
		Batching: queue.StateStoreConfig{Driver: "redis"},
		Restart:  queue.RestartConfig{Cache: "redis"},
	}
	for _, caseName := range []string{"failed-store", "batch-store", "restart-store"} {
		if err := queueDemoConfigError(caseName, cfg); err != nil {
			t.Fatalf("queueDemoConfigError(%q) = %v, want nil for configured store", caseName, err)
		}
	}
}
