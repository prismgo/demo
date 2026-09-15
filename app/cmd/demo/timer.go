package demo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	timerdemo "prismgo-demo/app/demo/timer"
)

// TimerCommand runs the documented scheduler examples.
type TimerCommand struct {
	run func(string) (timerdemo.Result, error)
}

// NewTimerCommand creates the Timer demo command.
func NewTimerCommand() *TimerCommand {
	return &TimerCommand{run: timerdemo.Run}
}

// Definition describes the demo:timer command.
func (c *TimerCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:timer {case? : Scenario name or list} {--json : Output JSON}",
		"Run scheduled task examples mapped to the Task Scheduling documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:timer list",
		"go run ./demo demo:timer architecture",
		"go run ./demo demo:timer every --json",
		"go run ./demo demo:timer weekly-on",
	}
	return definition
}

// Handle lists implemented scheduler scenarios or executes one by name.
func (c *TimerCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("timer", "", catalog.StatusImplemented)
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("timer", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented timer scenario %q", name))
	}
	result, err := c.run(name)
	if err != nil {
		return err
	}
	if ctx.Input().OptionBool("json") {
		return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(result)
	}
	return ctx.IO().Table([]string{"Case", "Value"}, [][]string{{result.Case, result.Value}})
}
