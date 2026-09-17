package demo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	schemademo "prismgo-demo/app/demo/schema"
)

// SchemaCommand runs the documented schema builder examples.
type SchemaCommand struct {
	run func(string) (schemademo.Result, error)
}

// NewSchemaCommand creates the Schema demo command.
func NewSchemaCommand() *SchemaCommand {
	return &SchemaCommand{run: schemademo.Run}
}

// Definition describes the demo:schema command.
func (c *SchemaCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:schema {case? : Scenario name or list} {--json : Output JSON}",
		"Run schema builder examples mapped to the Database: Schema Builder documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:schema list",
		"go run ./demo demo:schema architecture",
		"go run ./demo demo:schema create --json",
		"go run ./demo demo:schema change-column",
	}
	return definition
}

// Handle lists implemented schema scenarios or executes one by name.
func (c *SchemaCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("schema", "", catalog.StatusImplemented)
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("schema", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented schema scenario %q", name))
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
