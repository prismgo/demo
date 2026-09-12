package demo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	loggerdemo "prismgo-demo/app/demo/logger"
)

// LoggerCommand runs documentation-backed logger examples.
type LoggerCommand struct {
	run func(string) (loggerdemo.Result, error)
}

// NewLoggerCommand creates the logger demo command.
func NewLoggerCommand() *LoggerCommand {
	return &LoggerCommand{run: loggerdemo.Run}
}

// Definition describes the demo:logger command.
func (c *LoggerCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:logger {case? : Scenario name or list} {--json : Output JSON}",
		"Run logger examples mapped to the Logger documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:logger list",
		"go run ./demo demo:logger single-driver --json",
		"go run ./demo demo:logger daily-driver",
	}
	return definition
}

// Handle lists or executes implemented logger examples.
func (c *LoggerCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("logger", "", catalog.StatusImplemented)
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("logger", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented logger scenario %q", name))
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
