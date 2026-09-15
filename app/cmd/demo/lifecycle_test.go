package demo

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	lifecycledemo "prismgo-demo/app/demo/lifecycle"
)

func TestLifecycleDemoCommand(t *testing.T) {
	command := NewLifecycleCommand()
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "list", want: "application-entry"},
		{name: "run-context", want: "runner-ran=true context-canceled=true"},
		{name: "terminable-providers", want: "order=demo.p3,demo.p2,demo.p1"},
		{name: "http-pipeline", want: "order=received,middleware,handler,handled status=200"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var output bytes.Buffer
			err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": testCase.name}}, &output))
			if err != nil {
				t.Fatalf("demo:lifecycle %q error = %v, want nil", testCase.name, err)
			}
			if !strings.Contains(output.String(), testCase.want) {
				t.Fatalf("demo:lifecycle %q output = %q, want substring %q", testCase.name, output.String(), testCase.want)
			}
		})
	}
}

func TestLifecycleDemoCommandJSONAndUnknownScenario(t *testing.T) {
	command := NewLifecycleCommand()
	var output bytes.Buffer
	err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "command-resolution"}, bools: map[string]bool{"json": true}}, &output))
	if err != nil {
		t.Fatalf("demo:lifecycle command-resolution --json error = %v, want nil", err)
	}
	var result lifecycledemo.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode demo:lifecycle JSON = %v, output = %q", err, output.String())
	}
	if result.Case != "command-resolution" || result.Value != "name=demo:lifecycle arguments=1 options=1" {
		t.Fatalf("demo:lifecycle command-resolution --json = %#v, want command resolution result", result)
	}
	output.Reset()
	err = command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "unknown"}}, &output))
	if err == nil || !strings.Contains(err.Error(), `unknown or unimplemented lifecycle scenario "unknown"`) {
		t.Fatalf("demo:lifecycle unknown error = %v, want unknown scenario error", err)
	}
}
