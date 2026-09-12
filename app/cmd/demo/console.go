package demo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	consoledemo "prismgo-demo/app/demo/console"
)

// ConsoleCommand runs the documentation-backed console examples.
type ConsoleCommand struct{}

// NewConsoleCommand creates the console demo command.
func NewConsoleCommand() *ConsoleCommand {
	return &ConsoleCommand{}
}

// Definition describes the demo:console command.
func (c *ConsoleCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:console {case? : Scenario name or list} {--json : Output JSON}",
		"Run console examples mapped to the Console documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:console list",
		"go run ./demo demo:console signature",
		"go run ./demo demo:console command-structure --json",
	}
	return definition
}

// Handle lists console scenarios or runs the selected scenario.
func (c *ConsoleCommand) Handle(ctx console.CommandContext) error {
	caseName := strings.TrimSpace(ctx.Argument("case"))
	if caseName == "" || caseName == "list" {
		if ctx.OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(consoledemo.Scenarios())
		}
		rows := make([][]string, 0, len(consoledemo.Scenarios()))
		for _, name := range consoledemo.Scenarios() {
			rows = append(rows, []string{name, "implemented"})
		}
		return ctx.IO().Table([]string{"Case", "Status"}, rows)
	}
	result, err := consoledemo.Run(caseName)
	if err != nil {
		return ctx.Fail(fmt.Sprintf("demo:console %s: %v", caseName, err))
	}
	if ctx.OptionBool("json") {
		encoder := json.NewEncoder(console.OutputWriter(ctx.IO()))
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}
	rows := [][]string{{"case", result.Case}, {"signature", result.Signature}, {"name", result.Definition.Name}}
	for _, value := range result.Values {
		rows = append(rows, []string{"observed", value})
	}
	if result.Output != "" {
		rows = append(rows, []string{"stdout", fmt.Sprintf("%q", result.Output)})
	}
	if result.ErrorOutput != "" {
		rows = append(rows, []string{"stderr", fmt.Sprintf("%q", result.ErrorOutput)})
	}
	for _, argument := range result.Definition.Arguments {
		rows = append(rows, []string{"argument", fmt.Sprintf("%s required=%t array=%t", argument.Name, argument.Required, argument.IsArray)})
	}
	for _, option := range result.Definition.Options {
		rows = append(rows, []string{"option", fmt.Sprintf("%s shortcut=%s mode=%d array=%t", option.Name, option.Shortcut, option.ValueMode, option.IsArray)})
	}
	return ctx.IO().Table([]string{"Field", "Value"}, rows)
}
