package demo

import (
	"bytes"
	"strings"
	"testing"
)

func TestContainerDemoCommand(t *testing.T) {
	command := NewContainerCommand()
	for _, tt := range []struct {
		name string
		want string
	}{
		{name: "list", want: "singleton-retry"},
		{name: "list-entries", want: "keys=first,second registered=true,false closable=true created=0"},
		{name: "factory", want: "before=0 values=fresh-1,fresh-2 shared=true"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": tt.name}}, &output))
			if err != nil {
				t.Fatalf("demo:container %q error = %v, want nil", tt.name, err)
			}
			if !strings.Contains(output.String(), tt.want) {
				t.Fatalf("demo:container %q output = %q, want substring %q", tt.name, output.String(), tt.want)
			}
		})
	}
}

func TestContainerDemoCommandRejectsUnknownScenario(t *testing.T) {
	command := NewContainerCommand()
	err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "unknown"}}, &bytes.Buffer{}))
	if err == nil || !strings.Contains(err.Error(), `unknown or unimplemented container scenario "unknown"`) {
		t.Fatalf("unknown container scenario error = %v, want descriptive error", err)
	}
}
