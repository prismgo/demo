package demo

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"prismgo-demo/app/demo/catalog"
)

func TestCommandsDemoCommandListsOneHundredScenarios(t *testing.T) {
	command := NewCommandsCommand()
	var output bytes.Buffer
	input := demoInput{arguments: map[string]string{"case": "list"}, bools: map[string]bool{"json": true}}
	if err := command.Handle(commandContext(command, input, &output)); err != nil {
		t.Fatalf("demo:commands list --json error = %v, want nil", err)
	}
	var entries []catalog.Entry
	if err := json.Unmarshal(output.Bytes(), &entries); err != nil {
		t.Fatalf("decode commands list error = %v, want nil; output = %q", err, output.String())
	}
	if len(entries) != 100 {
		t.Fatalf("commands list entries = %d, want 100", len(entries))
	}
	for _, entry := range entries {
		if entry.Case == "list" || entry.Status != catalog.StatusImplemented {
			t.Fatalf("commands list entry = %#v, want implemented runnable scenario", entry)
		}
	}
}

func TestCommandsDemoCommandRejectsUnknownScenario(t *testing.T) {
	command := NewCommandsCommand()
	err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "missing"}}, &bytes.Buffer{}))
	if err == nil || !strings.Contains(err.Error(), `unknown or unimplemented commands scenario "missing"`) {
		t.Fatalf("demo:commands missing error = %v, want unknown scenario error", err)
	}
}
