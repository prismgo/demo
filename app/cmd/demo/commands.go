package demo

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	"prismgo-demo/app/demo/commands"
)

// CommandsCommand runs isolated examples of application commands.
type CommandsCommand struct{}

// NewCommandsCommand creates the commands demo command.
func NewCommandsCommand() *CommandsCommand { return &CommandsCommand{} }

// Definition describes the demo:commands command.
func (c *CommandsCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:commands {case? : Scenario name or list} {--json : Output JSON}",
		"Run application command examples in an isolated directory",
	)
	definition.Examples = []string{
		"go run ./demo demo:commands list",
		"go run ./demo demo:commands list-json",
		"go run ./demo demo:commands make-job --json",
	}
	return definition
}

// Handle lists or runs a commands scenario.
func (c *CommandsCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		catalogEntries := catalog.Filter("commands", "", catalog.StatusImplemented)
		entries := make([]catalog.Entry, 0, len(catalogEntries))
		for _, entry := range catalogEntries {
			if entry.Case != "list" {
				entries = append(entries, entry)
			}
		}
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("commands", name)
	if !ok || entry.Status != catalog.StatusImplemented || name == "list" {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented commands scenario %q", name))
	}
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate Demo executable: %w", err)
	}
	result, err := commands.Run(ctx.Context(), executable, name)
	if err != nil {
		return err
	}
	if ctx.Input().OptionBool("json") {
		return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(result)
	}
	return ctx.IO().Table([]string{"Case", "Command", "Observed"}, [][]string{{result.Case, result.Command, result.Output}})
}
