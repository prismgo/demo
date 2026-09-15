package demo

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"prismgo-demo/app/demo/catalog"
	providerdemo "prismgo-demo/app/demo/provider"
)

func TestProviderDemoCommand(t *testing.T) {
	command := NewProviderCommand()
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "list", want: "architecture"},
		{name: "singleton", want: "singleton=true created=1"},
		{name: "commands", want: "mounted=true command=provider:demo"},
		{name: "default-order", want: "order=redis,cache,queue,cookie,session,filesystem,database,database.schema,ratelimit,route"},
		{name: "deferred-resolution", want: "bound-before=false register=1 boot=1 value=deferred"},
		{name: "terminate-order", want: "order=p3,p2,p1"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var output bytes.Buffer
			err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": testCase.name}}, &output))
			if err != nil {
				t.Fatalf("demo:provider %q error = %v, want nil", testCase.name, err)
			}
			if !strings.Contains(output.String(), testCase.want) {
				t.Fatalf("demo:provider %q output = %q, want substring %q", testCase.name, output.String(), testCase.want)
			}
		})
	}
}

func TestProviderDemoCommandJSONAndUnknownScenario(t *testing.T) {
	command := NewProviderCommand()
	var output bytes.Buffer
	err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "singleton"}, bools: map[string]bool{"json": true}}, &output))
	if err != nil {
		t.Fatalf("demo:provider singleton --json error = %v, want nil", err)
	}
	var result providerdemo.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode demo:provider JSON = %v, output = %q", err, output.String())
	}
	if result.Case != "singleton" || result.Value != "singleton=true created=1" {
		t.Fatalf("demo:provider singleton --json = %#v, want singleton scenario result", result)
	}
	if item, ok := catalog.Find("service-provider", "unknown"); ok {
		t.Fatalf("unknown catalog entry = %#v, found=%t; want absent", item, ok)
	}
	output.Reset()
	err = command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "unknown"}}, &output))
	if err == nil || !strings.Contains(err.Error(), `unknown or unimplemented service provider scenario "unknown"`) {
		t.Fatalf("demo:provider unknown error = %v, want unknown scenario error", err)
	}
}
