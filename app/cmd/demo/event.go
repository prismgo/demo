package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	eventdemo "prismgo-demo/app/demo/event"
)

// EventCommand runs documented event examples.
type EventCommand struct {
	run func(context.Context, string, string) (eventdemo.Result, error)
}

// NewEventCommand creates the event demo command.
func NewEventCommand() *EventCommand { return &EventCommand{run: eventdemo.Run} }

// Definition describes the demo:event command.
func (c *EventCommand) Definition() *console.Definition {
	definition := console.MustDefinition("demo:event {case? : Scenario name or list} {--connection=sync : sync or redis} {--json : Output JSON}", "Run event examples mapped to the Event documentation")
	definition.Examples = []string{"go run ./demo demo:event list", "go run ./demo demo:event dispatch --json", "go run ./demo demo:event queued-redis --connection=redis"}
	return definition
}

// Handle lists or executes implemented event examples.
func (c *EventCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("event", "", catalog.StatusImplemented)
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("event", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented event scenario %q", name))
	}
	connection := strings.TrimSpace(ctx.Input().Option("connection"))
	if connection == "" {
		connection = "sync"
	}
	result, err := c.run(ctx.Context(), name, connection)
	if err != nil {
		return err
	}
	if ctx.Input().OptionBool("json") {
		return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(result)
	}
	return ctx.IO().Table([]string{"Case", "Value"}, [][]string{{result.Case, result.Value}})
}
