package demo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	cookiedemo "prismgo-demo/app/demo/cookie"
)

// CookieCommand runs the documented Cookie examples.
type CookieCommand struct {
	run func(string) (cookiedemo.Result, error)
}

// NewCookieCommand creates the Cookie demo command.
func NewCookieCommand() *CookieCommand {
	return &CookieCommand{run: cookiedemo.Run}
}

// Definition describes the demo:cookie command.
func (c *CookieCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:cookie {case? : Scenario name or list} {--json : Output JSON}",
		"Run Cookie examples mapped to the Cookie documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:cookie list",
		"go run ./demo demo:cookie middleware",
		"go run ./demo demo:cookie queue-make --json",
	}
	return definition
}

// Handle lists implemented Cookie scenarios or executes one by name.
func (c *CookieCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("cookie", "", catalog.StatusImplemented)
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("cookie", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented cookie scenario %q", name))
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
