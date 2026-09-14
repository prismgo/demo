package demo

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	sessiondemo "prismgo-demo/app/demo/session"
)

func TestSessionDemoCommand(t *testing.T) {
	command := NewSessionCommand()
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "list", want: "custom-manager"},
		{name: "flash", want: "flash r1:status=ok r2:status=ok r3:status=-"},
		{name: "get", want: "value=alice default=guest absent=<nil>"},
		{name: "middleware", want: "cookie=prismgo_session"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var output bytes.Buffer
			err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": testCase.name}}, &output))
			if err != nil {
				t.Fatalf("demo:session %q error = %v, want nil", testCase.name, err)
			}
			if !strings.Contains(output.String(), testCase.want) {
				t.Fatalf("demo:session %q output = %q, want substring %q", testCase.name, output.String(), testCase.want)
			}
		})
	}
}

func TestSessionDemoCommandJSONAndUnknownScenario(t *testing.T) {
	command := NewSessionCommand()
	var output bytes.Buffer
	err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "counters"}, bools: map[string]bool{"json": true}}, &output))
	if err != nil {
		t.Fatalf("demo:session counters --json error = %v, want nil", err)
	}
	var result sessiondemo.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode demo:session JSON = %v, output = %q", err, output.String())
	}
	if result.Case != "counters" || result.Value != "first=1 second=3 remaining=-2 invalid=true" {
		t.Fatalf("demo:session counters --json = %#v, want counters scenario result", result)
	}
	output.Reset()
	err = command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "unknown"}}, &output))
	if err == nil || !strings.Contains(err.Error(), `unknown or unimplemented session scenario "unknown"`) {
		t.Fatalf("demo:session unknown error = %v, want unknown scenario error", err)
	}
}

func TestSessionDemoCommandConnectionRouting(t *testing.T) {
	command := NewSessionCommand()
	var output bytes.Buffer
	err := command.Handle(commandContext(command, demoInput{
		arguments: map[string]string{"case": "redis-driver"},
		options:   map[string]string{"connection": "file"},
	}, &output))
	if err == nil || !strings.Contains(err.Error(), `requires --connection=redis`) {
		t.Fatalf("demo:session redis-driver with file connection error = %v, want redis connection requirement", err)
	}
	output.Reset()
	err = command.Handle(commandContext(command, demoInput{
		arguments: map[string]string{"case": "get"},
		options:   map[string]string{"connection": "redis"},
	}, &output))
	if err == nil || !strings.Contains(err.Error(), `does not support --connection=redis`) {
		t.Fatalf("demo:session get with redis connection error = %v, want unsupported connection", err)
	}
}
