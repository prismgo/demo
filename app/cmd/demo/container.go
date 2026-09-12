package demo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	containerdemo "prismgo-demo/app/demo/container"
)

// ContainerCommand runs documentation-backed container examples.
type ContainerCommand struct {
	run func(string) (containerdemo.Result, error)
}

// NewContainerCommand creates the container demo command.
func NewContainerCommand() *ContainerCommand {
	return &ContainerCommand{run: containerdemo.Run}
}

// Definition describes the demo:container command.
func (c *ContainerCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:container {case? : Scenario name or list} {--json : Output JSON}",
		"Run container examples mapped to the Container documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:container list",
		"go run ./demo demo:container singleton",
		"go run ./demo demo:container factory --json",
	}
	return definition
}

// Handle lists implemented container scenarios or executes one by name.
func (c *ContainerCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("container", "", catalog.StatusImplemented)
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
	entry, ok := catalog.Find("container", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented container scenario %q", name))
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
