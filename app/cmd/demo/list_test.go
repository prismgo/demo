package demo

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
)

func TestDemoListCommandShowsFeatureOverview(t *testing.T) {
	command := NewListCommand()
	var output bytes.Buffer
	if err := command.Handle(commandContext(command, demoInput{}, &output)); err != nil {
		t.Fatalf("handle demo:list: %v", err)
	}
	for _, expected := range []string{
		"PrismGo Framework: v0.2.2 (local workspace)",
		"Feature", "Description", "Since", "Progress", "Remaining", "Status",
		"Queues, jobs, and workers", "38/59", "in progress",
		"Modules: 29 | Implemented: 39/87 | Planned: 45 | Manual: 3",
		"go run ./demo demo:show <feature>",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("output does not contain %q:\n%s", expected, output.String())
		}
	}
	if strings.Contains(output.String(), "demo:queue basic") {
		t.Fatalf("overview unexpectedly contains entry details:\n%s", output.String())
	}
}

func TestDemoListCommandJSONAndFeatureStatusFilter(t *testing.T) {
	command := NewListCommand()
	var output bytes.Buffer
	input := demoInput{
		options: map[string]string{"status": "in-progress"},
		bools:   map[string]bool{"json": true},
	}
	if err := command.Handle(commandContext(command, input, &output)); err != nil {
		t.Fatalf("handle filtered demo:list: %v", err)
	}
	var result overviewOutput
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode JSON: %v\n%s", err, output.String())
	}
	if result.Framework != "v0.2.2" {
		t.Fatalf("framework = %q, want v0.2.2", result.Framework)
	}
	if len(result.Features) != 1 || result.Features[0].Feature != "queue" || result.Features[0].Status != catalog.FeatureStatusInProgress {
		t.Fatalf("features = %#v, want queue in progress", result.Features)
	}
	if strings.Contains(output.String(), "\x1b[") {
		t.Fatalf("JSON contains ANSI decoration: %q", output.String())
	}
}

func TestDemoListCommandColorsFeatureStatuses(t *testing.T) {
	command := NewListCommand()
	var output bytes.Buffer
	ioo := console.NewIOWithOutputOptions(
		strings.NewReader(""), &output, &output, console.OutputOptions{ANSI: true},
	)
	ctx := commandContextWithIO(command, demoInput{}, ioo)
	if err := command.Handle(ctx); err != nil {
		t.Fatalf("handle ANSI demo:list: %v", err)
	}
	for _, expected := range []string{
		"\x1b[32mimplemented\x1b[0m",
		"\x1b[32mcommands\x1b[0m",
		"\x1b[33min progress\x1b[0m",
		"\x1b[33mqueue\x1b[0m",
		"\x1b[39mcache\x1b[0m",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("ANSI output does not contain %q:\n%q", expected, output.String())
		}
	}
}

func TestDemoListCommandPreservesFeatureDetailShortcut(t *testing.T) {
	command := NewListCommand()
	var output bytes.Buffer
	input := demoInput{arguments: map[string]string{"feature": "cache"}}
	if err := command.Handle(commandContext(command, input, &output)); err != nil {
		t.Fatalf("handle demo:list cache: %v", err)
	}
	for _, expected := range []string{"Module: cache", "demo:cache list", "Entries: 1"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("detail shortcut output does not contain %q:\n%s", expected, output.String())
		}
	}
}

func TestDemoShowCommandJSONAndFilters(t *testing.T) {
	command := NewShowCommand()
	var output bytes.Buffer
	input := demoInput{
		options: map[string]string{"level": "integration", "status": "planned"},
		bools:   map[string]bool{"json": true},
	}
	if err := command.Handle(commandContext(command, input, &output)); err != nil {
		t.Fatalf("handle filtered demo:show: %v", err)
	}
	var result detailOutput
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode JSON: %v\n%s", err, output.String())
	}
	if result.Framework != "v0.2.2" || len(result.Entries) != 16 {
		t.Fatalf("detail output framework/entries = %q/%d, want v0.2.2/16", result.Framework, len(result.Entries))
	}
	for _, item := range result.Entries {
		if item.Level != catalog.LevelIntegration || item.Status != catalog.StatusPlanned || item.Since != catalog.SinceInitial {
			t.Fatalf("unexpected filtered item: %#v", item)
		}
	}
}

func TestDemoShowCommandShowsFeatureDetails(t *testing.T) {
	command := NewShowCommand()
	var output bytes.Buffer
	input := demoInput{
		arguments: map[string]string{"feature": "queue"},
		options:   map[string]string{"status": "implemented"},
	}
	if err := command.Handle(commandContext(command, input, &output)); err != nil {
		t.Fatalf("handle demo:show queue: %v", err)
	}
	for _, expected := range []string{
		"Module: queue — 38/59 implemented — in progress",
		"Case", "Section", "Since", "Example", "Level", "Requires", "Status",
		"basic-sync", "v0.1.0", "Entries: 38 | Implemented: 38 | Planned: 0 | Manual: 0",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("detail output does not contain %q:\n%s", expected, output.String())
		}
	}
}

func TestDemoShowCommandColorsModuleAndCases(t *testing.T) {
	command := NewShowCommand()
	var output bytes.Buffer
	input := demoInput{
		arguments: map[string]string{"feature": "queue"},
		options:   map[string]string{"status": "implemented"},
	}
	ioo := console.NewIOWithOutputOptions(
		strings.NewReader(""), &output, &output, console.OutputOptions{ANSI: true},
	)
	if err := command.Handle(commandContextWithIO(command, input, ioo)); err != nil {
		t.Fatalf("handle ANSI demo:show queue: %v", err)
	}
	for _, expected := range []string{
		"\x1b[33mqueue\x1b[0m",
		"\x1b[32mbasic-sync\x1b[0m",
		"\x1b[32mimplemented\x1b[0m",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("ANSI detail output does not contain %q:\n%q", expected, output.String())
		}
	}
}

func TestDemoShowCommandIncludesFeatureWhenShowingAllEntries(t *testing.T) {
	command := NewShowCommand()
	var output bytes.Buffer
	input := demoInput{options: map[string]string{"status": "manual"}}
	if err := command.Handle(commandContext(command, input, &output)); err != nil {
		t.Fatalf("handle demo:show manual entries: %v", err)
	}
	for _, expected := range []string{"Feature", "Case", "installation", "manual", "Entries: 3"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("all-detail output does not contain %q:\n%s", expected, output.String())
		}
	}
}

func TestDemoShowCommandReportsEmptyFilteredFeature(t *testing.T) {
	command := NewShowCommand()
	var output bytes.Buffer
	input := demoInput{
		arguments: map[string]string{"feature": "cache"},
		options:   map[string]string{"status": "implemented"},
	}
	if err := command.Handle(commandContext(command, input, &output)); err != nil {
		t.Fatalf("handle empty demo:show result: %v", err)
	}
	if !strings.Contains(output.String(), "No documentation demos matched the selected filters.") {
		t.Fatalf("empty result warning missing:\n%s", output.String())
	}
}

func TestUnknownFeatureMessage(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "queu", want: `did you mean "queue"`},
		{name: "missing", want: "run demo:list"},
	}
	for _, test := range tests {
		if got := unknownFeatureMessage(test.name); !strings.Contains(got, test.want) {
			t.Fatalf("unknownFeatureMessage(%q) = %q, want substring %q", test.name, got, test.want)
		}
	}
}

func TestDemoCatalogCommandsRejectUnknownFiltersAndFeatures(t *testing.T) {
	tests := []struct {
		name    string
		command console.Command
		input   demoInput
	}{
		{name: "list level", command: NewListCommand(), input: demoInput{options: map[string]string{"level": "integration"}}},
		{name: "list status", command: NewListCommand(), input: demoInput{options: map[string]string{"status": "done"}}},
		{name: "show level", command: NewShowCommand(), input: demoInput{options: map[string]string{"level": "remote"}}},
		{name: "show status", command: NewShowCommand(), input: demoInput{options: map[string]string{"status": "done"}}},
		{name: "show feature", command: NewShowCommand(), input: demoInput{arguments: map[string]string{"feature": "queu"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.command.Handle(commandContext(test.command, test.input, io.Discard)); err == nil {
				t.Fatal("Handle() error = nil")
			}
		})
	}
}

func commandContext(command console.Command, input console.Input, output io.Writer) console.CommandContext {
	return commandContextWithIO(command, input, console.NewIO(strings.NewReader(""), output, output))
}

func commandContextWithIO(command console.Command, input console.Input, ioo console.IO) console.CommandContext {
	return console.NewCommandContext(
		context.Background(), command, *command.Definition(), input, ioo, nil, nil,
	)
}

type demoInput struct {
	arguments map[string]string
	options   map[string]string
	bools     map[string]bool
}

// Argument returns a scalar test argument.
func (i demoInput) Argument(name string) string { return i.arguments[name] }

// Arguments returns no variadic arguments for this test input.
func (i demoInput) Arguments(string) []string { return nil }

// Option returns a scalar test option.
func (i demoInput) Option(name string) string { return i.options[name] }

// OptionStrings returns no repeated options for this test input.
func (i demoInput) OptionStrings(string) []string { return nil }

// OptionBool returns a boolean test option.
func (i demoInput) OptionBool(name string) bool { return i.bools[name] }

// OptionInt returns the zero value for unused integer options.
func (i demoInput) OptionInt(string) (int, error) { return 0, nil }

// HasOption reports whether a scalar test option exists.
func (i demoInput) HasOption(name string) bool { _, ok := i.options[name]; return ok }
