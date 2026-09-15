package demo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	routedemo "prismgo-demo/app/demo/route"
)

// RouteCommand runs the documented HTTP router examples.
type RouteCommand struct {
	run func(string) (routedemo.Result, error)
}

// NewRouteCommand creates the Route demo command.
func NewRouteCommand() *RouteCommand {
	return &RouteCommand{run: routedemo.Run}
}

// Definition describes the demo:route command.
func (c *RouteCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:route {case? : Scenario name or list} {--json : Output JSON}",
		"Run HTTP router examples mapped to the Route documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:route list",
		"go run ./demo demo:route facade --json",
		"go run ./demo demo:route url-escaping",
		"go run ./demo demo:route list-entries --json",
	}
	return definition
}

// Handle lists implemented Route scenarios or executes one by name.
func (c *RouteCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("route", "", catalog.StatusImplemented)
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	if name == "list-entries" {
		name = "list"
	}
	entry, ok := catalog.Find("route", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented route scenario %q", name))
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
