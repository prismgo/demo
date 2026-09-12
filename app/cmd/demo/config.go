package demo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	configdemo "prismgo-demo/app/demo/config"
)

// ConfigCommand runs documentation-backed configuration examples.
type ConfigCommand struct {
	run func(string) (configdemo.Result, error)
}

// NewConfigCommand creates the configuration demo command.
func NewConfigCommand() *ConfigCommand {
	return &ConfigCommand{run: configdemo.Run}
}

// Definition describes the demo:config command.
func (c *ConfigCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:config {case? : Scenario name or list} {--json : Output JSON}",
		"Run configuration examples mapped to the Config documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:config list",
		"go run ./demo demo:config quick-start",
		"go run ./demo demo:config get-default --json",
	}
	return definition
}

// Handle lists implemented configuration scenarios or executes one by name.
func (c *ConfigCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("config", "", catalog.StatusImplemented)
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("config", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented config scenario %q", name))
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
