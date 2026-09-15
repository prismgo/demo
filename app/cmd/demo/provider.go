package demo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	providerdemo "prismgo-demo/app/demo/provider"
)

// ProviderCommand runs the documented service provider examples.
type ProviderCommand struct {
	run func(string) (providerdemo.Result, error)
}

// NewProviderCommand creates the Provider demo command.
func NewProviderCommand() *ProviderCommand {
	return &ProviderCommand{run: providerdemo.Run}
}

// Definition describes the demo:provider command.
func (c *ProviderCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:provider {case? : Scenario name or list} {--json : Output JSON}",
		"Run service provider examples mapped to the Service Provider documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:provider list",
		"go run ./demo demo:provider singleton --json",
		"go run ./demo demo:provider deferred-resolution",
		"go run ./demo demo:provider terminate-order --json",
	}
	return definition
}

// Handle lists implemented Provider scenarios or executes one by name.
func (c *ProviderCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("service-provider", "", catalog.StatusImplemented)
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("service-provider", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented service provider scenario %q", name))
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
