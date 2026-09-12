package demo

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfigDemoCommandListsImplementedScenarios(t *testing.T) {
	command := NewConfigCommand()
	var output bytes.Buffer
	if err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "list"}}, &output)); err != nil {
		t.Fatalf("handle demo:config list error = %v, want nil", err)
	}
	for _, want := range []string{"quick-start", "get-default", "app-env", "Documentation"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("config list output = %q, want %q", output.String(), want)
		}
	}
}

func TestConfigDemoCommandRunsScenario(t *testing.T) {
	command := NewConfigCommand()
	var output bytes.Buffer
	if err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "get-default"}}, &output)); err != nil {
		t.Fatalf("handle demo:config get-default error = %v, want nil", err)
	}
	if !strings.Contains(output.String(), "fallback/7/true") {
		t.Fatalf("config command output = %q, want fallback/7/true", output.String())
	}
}

func TestConfigDemoCommandRejectsUnknownScenario(t *testing.T) {
	command := NewConfigCommand()
	err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "not-a-config-case"}}, &bytes.Buffer{}))
	if err == nil || !strings.Contains(err.Error(), `unknown or unimplemented config scenario "not-a-config-case"`) {
		t.Fatalf("unimplemented config scenario error = %v, want descriptive error", err)
	}
}
