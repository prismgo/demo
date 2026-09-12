package demo

import (
	"bytes"
	"strings"
	"testing"
)

func TestExceptionDemoCommandListsAndRejectsScenarios(t *testing.T) {
	command := NewExceptionCommand()
	var output bytes.Buffer
	if err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "list"}}, &output)); err != nil {
		t.Fatalf("handle demo:exception list error = %v, want nil", err)
	}
	for _, want := range []string{"default-report", "public-detail", "safe-response", "handler-report", "handler-render", "horizon-report", "Documentation"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("exception list output = %q, want %q", output.String(), want)
		}
	}
	for _, name := range []string{"missing"} {
		err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": name}}, &bytes.Buffer{}))
		if err == nil || !strings.Contains(err.Error(), name) {
			t.Fatalf("exception case %q error = %v, want descriptive rejection", name, err)
		}
	}
}
