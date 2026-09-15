package demo

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	timerdemo "prismgo-demo/app/demo/timer"
)

func TestTimerDemoCommand(t *testing.T) {
	command := NewTimerCommand()
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "list", want: "architecture"},
		{name: "architecture", want: "schedule=Schedule task=ScheduledTask resolver=CommandResolver timer=Timer"},
		{name: "command", want: "args=--take=100"},
		{name: "command-validation", want: "empty=panic nil-resolver=panic unknown=panic"},
		{name: "timezone", want: "location=demo-zone offset=+08:00"},
		{name: "quarterly-on", want: "day=45 at=10:00"},
		{name: "summary", want: "names=overtime:detect,tenant_sync"},
		{name: "weekday-parsing", want: "aliases=ok ranges=ok slices=ok invalid=panic"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var output bytes.Buffer
			err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": testCase.name}}, &output))
			if err != nil {
				t.Fatalf("demo:timer %q error = %v, want nil", testCase.name, err)
			}
			if !strings.Contains(output.String(), testCase.want) {
				t.Fatalf("demo:timer %q output = %q, want substring %q", testCase.name, output.String(), testCase.want)
			}
		})
	}
}

func TestTimerDemoCommandJSONAndUnknownScenario(t *testing.T) {
	command := NewTimerCommand()
	var output bytes.Buffer
	err := command.Handle(commandContext(command, demoInput{
		arguments: map[string]string{"case": "architecture"},
		bools:     map[string]bool{"json": true},
	}, &output))
	if err != nil {
		t.Fatalf("demo:timer architecture --json error = %v, want nil", err)
	}
	var result timerdemo.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode demo:timer JSON = %v, output = %q", err, output.String())
	}
	if result.Case != "architecture" {
		t.Fatalf("demo:timer architecture --json = %#v, want architecture result", result)
	}

	output.Reset()
	err = command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "not-a-case"}}, &output))
	if err == nil || !strings.Contains(err.Error(), `unknown or unimplemented timer scenario "not-a-case"`) {
		t.Fatalf("demo:timer unknown error = %v, want unknown scenario error", err)
	}
}
