package demo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	lifecycledemo "prismgo-demo/app/demo/lifecycle"
)

// LifecycleCommand runs the documented application lifecycle examples.
type LifecycleCommand struct {
	run func(string) (lifecycledemo.Result, error)
}

// NewLifecycleCommand creates the lifecycle demo command.
func NewLifecycleCommand() *LifecycleCommand {
	return &LifecycleCommand{run: lifecycledemo.Run}
}

// Definition describes the demo:lifecycle command.
func (c *LifecycleCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:lifecycle {case? : Scenario name or list} {--json : Output JSON}",
		"Run application lifecycle examples mapped to the Lifecycle documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:lifecycle list",
		"go run ./demo demo:lifecycle run-context",
		"go run ./demo demo:lifecycle http-pipeline --json",
		"go run ./demo demo:lifecycle terminable-providers",
	}
	return definition
}

// Handle lists implemented lifecycle scenarios or executes one by name.
func (c *LifecycleCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("lifecycle", "", catalog.StatusImplemented)
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("lifecycle", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented lifecycle scenario %q", name))
	}
	result, err := c.run(entry.Case)
	if err != nil {
		return err
	}
	if ctx.Input().OptionBool("json") {
		return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(result)
	}
	return ctx.IO().Table([]string{"Case", "Value"}, [][]string{{result.Case, result.Value}})
}
