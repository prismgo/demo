package demo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	exceptiondemo "prismgo-demo/app/demo/exception"
)

// ExceptionCommand runs documentation-backed exception examples.
type ExceptionCommand struct {
	run func(string) (exceptiondemo.Result, error)
}

// NewExceptionCommand creates the exception demo command.
func NewExceptionCommand() *ExceptionCommand {
	return &ExceptionCommand{run: exceptiondemo.Run}
}

// Definition describes the demo:exception command.
func (c *ExceptionCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:exception {case? : Scenario name or list} {--json : Output JSON}",
		"Run exception examples mapped to the Exception documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:exception list",
		"go run ./demo demo:exception default-render --json",
		"go run ./demo demo:exception panic-recovery",
	}
	return definition
}

// Handle lists or executes implemented exception examples.
func (c *ExceptionCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("exception", "", catalog.StatusImplemented)
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("exception", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented exception scenario %q", name))
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
