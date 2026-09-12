package demo

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"prismgo-demo/app/demo/catalog"
	filesystemdemo "prismgo-demo/app/demo/filesystem"
)

func TestFilesystemDemoCommandListsRunnableScenarios(t *testing.T) {
	command := NewFilesystemCommand()
	var output bytes.Buffer
	if err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "list"}}, &output)); err != nil {
		t.Fatalf("handle demo:filesystem list error = %v, want nil", err)
	}
	for _, want := range []string{"architecture", "named-disk", "size"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("filesystem list output = %q, want %q", output.String(), want)
		}
	}
	if strings.Contains(output.String(), "oss-driver") {
		t.Fatalf("filesystem list output = %q, want test-only oss-driver omitted", output.String())
	}
}

func TestFilesystemDemoCommandListsSixtySixRunnableScenariosAsJSON(t *testing.T) {
	command := NewFilesystemCommand()
	var output bytes.Buffer
	input := demoInput{arguments: map[string]string{"case": "list"}, bools: map[string]bool{"json": true}}
	if err := command.Handle(commandContext(command, input, &output)); err != nil {
		t.Fatalf("handle demo:filesystem list --json error = %v, want nil", err)
	}
	var entries []catalog.Entry
	if err := json.Unmarshal(output.Bytes(), &entries); err != nil {
		t.Fatalf("decode filesystem list JSON error = %v, want nil; output = %q", err, output.String())
	}
	if len(entries) != 66 {
		t.Fatalf("runnable filesystem entries = %d, want 66", len(entries))
	}
}

func TestFilesystemDemoCommandRunsScenario(t *testing.T) {
	command := &FilesystemCommand{run: func(_ context.Context, name string) (filesystemdemo.Result, error) {
		return filesystemdemo.Result{Case: name, Value: "hello filesystem"}, nil
	}}
	var output bytes.Buffer
	if err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "get"}}, &output)); err != nil {
		t.Fatalf("handle demo:filesystem get error = %v, want nil", err)
	}
	if !strings.Contains(output.String(), "hello filesystem") {
		t.Fatalf("filesystem result = %q, want hello filesystem", output.String())
	}
}

func TestFilesystemDemoCommandRejectsUnrunnableScenario(t *testing.T) {
	command := NewFilesystemCommand()
	for _, name := range []string{"oss-driver", "unknown"} {
		err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": name}}, &bytes.Buffer{}))
		if err == nil || !strings.Contains(err.Error(), "unknown or unimplemented filesystem scenario") {
			t.Fatalf("filesystem scenario %q error = %v, want rejection", name, err)
		}
	}
}
