package demo

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"prismgo-demo/app/demo/catalog"
	ratelimitdemo "prismgo-demo/app/demo/ratelimit"
)

func TestRateLimitDemoCommand(t *testing.T) {
	command := NewRateLimitCommand()
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "list", want: "architecture"},
		{name: "per-minute", want: "max=3 decay=1m0s"},
		{name: "result-contract", want: "attempts=2"},
		{name: "named-lookup", want: "found=true missing=true blank=true"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var output bytes.Buffer
			err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": testCase.name}}, &output))
			if err != nil {
				t.Fatalf("demo:ratelimit %q error = %v, want nil", testCase.name, err)
			}
			if !strings.Contains(output.String(), testCase.want) {
				t.Fatalf("demo:ratelimit %q output = %q, want substring %q", testCase.name, output.String(), testCase.want)
			}
		})
	}
}

func TestRateLimitDemoCommandJSONAndStoreValidation(t *testing.T) {
	command := NewRateLimitCommand()
	var output bytes.Buffer
	err := command.Handle(commandContext(command, demoInput{
		arguments: map[string]string{"case": "per-minute"},
		bools:     map[string]bool{"json": true},
	}, &output))
	if err != nil {
		t.Fatalf("demo:ratelimit per-minute --json error = %v, want nil", err)
	}
	var result ratelimitdemo.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode demo:ratelimit JSON = %v, output = %q", err, output.String())
	}
	if result.Case != "per-minute" || result.Value != "max=3 decay=1m0s" {
		t.Fatalf("demo:ratelimit per-minute --json = %#v, want per-minute result", result)
	}

	if item, ok := catalog.Find("ratelimit", "unknown"); ok {
		t.Fatalf("unknown catalog entry = %#v, found=%t; want absent", item, ok)
	}
	output.Reset()
	err = command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "unknown"}}, &output))
	if err == nil || !strings.Contains(err.Error(), `unknown or unimplemented ratelimit scenario "unknown"`) {
		t.Fatalf("demo:ratelimit unknown error = %v, want unknown scenario error", err)
	}

	err = command.Handle(commandContext(command, demoInput{
		arguments: map[string]string{"case": "redis-store"},
		options:   map[string]string{"store": "memory"},
	}, &output))
	if err == nil || !strings.Contains(err.Error(), "requires --store=redis") {
		t.Fatalf("demo:ratelimit redis-store without redis store error = %v, want store requirement", err)
	}

	err = command.Handle(commandContext(command, demoInput{
		arguments: map[string]string{"case": "per-minute"},
		options:   map[string]string{"store": "redis"},
	}, &output))
	if err == nil || !strings.Contains(err.Error(), "does not support --store=redis") {
		t.Fatalf("demo:ratelimit per-minute with redis store error = %v, want unsupported store error", err)
	}
}
