package demo

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestConsoleDemoCommandListsScenarios(t *testing.T) {
	command := NewConsoleCommand()
	var output bytes.Buffer
	if err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "list"}}, &output)); err != nil {
		t.Fatalf("handle demo:console list error = %v, want nil", err)
	}
	for _, want := range []string{"architecture", "command-structure", "argument", "implemented"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("console list output = %q, want substring %q", output.String(), want)
		}
	}
}

func TestConsoleDemoCommandRunsScenarioJSON(t *testing.T) {
	command := NewConsoleCommand()
	var output bytes.Buffer
	input := demoInput{arguments: map[string]string{"case": "required-argument"}, bools: map[string]bool{"json": true}}
	if err := command.Handle(commandContext(command, input, &output)); err != nil {
		t.Fatalf("handle demo:console required-argument error = %v, want nil", err)
	}
	var result struct {
		Case       string `json:"case"`
		Definition struct {
			Arguments []struct {
				Name     string `json:"Name"`
				Required bool   `json:"Required"`
			} `json:"Arguments"`
		} `json:"definition"`
	}
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode demo:console JSON error = %v, output = %q", err, output.String())
	}
	if result.Case != "required-argument" || len(result.Definition.Arguments) != 1 || result.Definition.Arguments[0].Name != "user" || !result.Definition.Arguments[0].Required {
		t.Errorf("demo:console JSON = %#v, want required user argument", result)
	}
}

func TestConsoleDemoCommandRejectsUnknownScenario(t *testing.T) {
	command := NewConsoleCommand()
	var output bytes.Buffer
	err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "missing"}}, &output))
	if err == nil || !strings.Contains(err.Error(), `unknown scenario "missing"`) {
		t.Errorf("handle demo:console missing error = %v, want unknown scenario", err)
	}
}
