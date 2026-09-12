package demo

import (
	"bytes"
	"context"
	"strings"
	"testing"

	eventdemo "prismgo-demo/app/demo/event"
)

func TestEventDemoCommandListsAllCases(t *testing.T) {
	command := NewEventCommand()
	var output bytes.Buffer
	if err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "list"}}, &output)); err != nil {
		t.Fatalf("list event cases: %v", err)
	}
	for _, want := range []string{"architecture", "queued-redis", "async", "async-durability", "queued-worker-testing", "Documentation"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("event list output = %q, want %q", output.String(), want)
		}
	}
}
func TestEventDemoCommandRunsCase(t *testing.T) {
	command := &EventCommand{run: func(_ context.Context, name, connection string) (eventdemo.Result, error) {
		return eventdemo.Result{Case: name, Value: connection + ":handled"}, nil
	}}
	var output bytes.Buffer
	input := demoInput{arguments: map[string]string{"case": "dispatch"}}
	if err := command.Handle(commandContext(command, input, &output)); err != nil {
		t.Fatalf("run event case: %v", err)
	}
	if !strings.Contains(output.String(), "sync:handled") {
		t.Fatalf("event output = %q, want sync:handled", output.String())
	}
}
