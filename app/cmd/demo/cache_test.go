package demo

import (
	"bytes"
	"context"
	"strings"
	"testing"

	cachedemo "prismgo-demo/app/demo/cache"
)

func TestCacheDemoCommandListsAllScenarios(t *testing.T) {
	command := NewCacheCommand()
	var output bytes.Buffer
	if err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "list"}}, &output)); err != nil {
		t.Fatalf("handle demo:cache list error = %v, want nil", err)
	}
	for _, want := range []string{"architecture", "redis-config", "forever", "tags-memory", "laravel-compatibility", "Documentation"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("cache list output = %q, want %q", output.String(), want)
		}
	}
}

func TestCacheDemoCommandRunsScenario(t *testing.T) {
	command := &CacheCommand{run: func(_ context.Context, name string) (cachedemo.Result, error) {
		return cachedemo.Result{Case: name, Key: "example", Value: "stored"}, nil
	}}
	var output bytes.Buffer
	if err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "get"}}, &output)); err != nil {
		t.Fatalf("handle demo:cache get error = %v, want nil", err)
	}
	if !strings.Contains(output.String(), "stored") || !strings.Contains(output.String(), "example") {
		t.Fatalf("cache result output = %q, want stored value and example key", output.String())
	}
}

func TestCacheDemoCommandRejectsUnknownScenario(t *testing.T) {
	command := NewCacheCommand()
	err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "not-a-cache-case"}}, &bytes.Buffer{}))
	if err == nil || !strings.Contains(err.Error(), `unknown or unimplemented cache scenario "not-a-cache-case"`) {
		t.Fatalf("unknown scenario error = %v, want descriptive error", err)
	}
}
