package demo

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"prismgo-demo/app/demo/catalog"

	"github.com/prismgo/framework/console"
)

func TestDemoListCommand(t *testing.T) {
	command := NewListCommand()
	var output bytes.Buffer
	ctx := commandContext(command, demoInput{arguments: map[string]string{"feature": "cache"}}, &output)
	if err := command.Handle(ctx); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"cache", "demo:cache list", "planned", "Coverage entries: 1"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("output does not contain %q:\n%s", expected, output.String())
		}
	}
}

func TestDemoListCommandJSONAndFilters(t *testing.T) {
	command := NewListCommand()
	var output bytes.Buffer
	input := demoInput{
		options: map[string]string{"level": "integration", "status": "planned"},
		bools:   map[string]bool{"json": true},
	}
	if err := command.Handle(commandContext(command, input, &output)); err != nil {
		t.Fatal(err)
	}
	var entries []catalog.Entry
	if err := json.Unmarshal(output.Bytes(), &entries); err != nil {
		t.Fatalf("decode JSON: %v\n%s", err, output.String())
	}
	if len(entries) != 2 {
		t.Fatalf("integration planned entries = %d, want 2", len(entries))
	}
	for _, item := range entries {
		if item.Level != catalog.LevelIntegration || item.Status != catalog.StatusPlanned {
			t.Fatalf("unexpected filtered item: %#v", item)
		}
	}
}

func TestDemoListCommandRejectsUnknownFilter(t *testing.T) {
	command := NewListCommand()
	for _, input := range []demoInput{
		{options: map[string]string{"level": "remote"}},
		{options: map[string]string{"status": "done"}},
	} {
		if err := command.Handle(commandContext(command, input, io.Discard)); err == nil {
			t.Fatal("Handle() error = nil")
		}
	}
}

func commandContext(command console.Command, input console.Input, output io.Writer) console.CommandContext {
	return console.NewCommandContext(
		context.Background(), command, *command.Definition(), input,
		console.NewIO(strings.NewReader(""), output, output), nil, nil,
	)
}

type demoInput struct {
	arguments map[string]string
	options   map[string]string
	bools     map[string]bool
}

func (i demoInput) Argument(name string) string   { return i.arguments[name] }
func (i demoInput) Arguments(string) []string     { return nil }
func (i demoInput) Option(name string) string     { return i.options[name] }
func (i demoInput) OptionStrings(string) []string { return nil }
func (i demoInput) OptionBool(name string) bool   { return i.bools[name] }
func (i demoInput) OptionInt(string) (int, error) { return 0, nil }
func (i demoInput) HasOption(name string) bool    { _, ok := i.options[name]; return ok }
