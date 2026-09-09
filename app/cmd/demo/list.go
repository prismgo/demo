// Package demo implements documentation-backed demo commands.
package demo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
)

// ListCommand renders the documentation-to-demo coverage catalog.
type ListCommand struct{}

// NewListCommand creates the demo discovery command.
func NewListCommand() *ListCommand { return &ListCommand{} }

// Definition describes the demo:list command.
func (c *ListCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:list {feature? : Filter by feature} {--level= : compile, hermetic, scenario, integration} {--status= : implemented, planned, manual} {--json : Output JSON}",
		"List documentation demos, verification levels, dependencies, and remaining gaps",
	)
	definition.Examples = []string{
		"go run ./dev demo:list",
		"go run ./dev demo:list cache",
		"go run ./dev demo:list --status=planned",
		"go run ./dev demo:list --level=integration --json",
	}
	return definition
}

// Handle lists catalog entries matching the command filters.
func (c *ListCommand) Handle(ctx console.CommandContext) error {
	level := catalog.Level(strings.TrimSpace(ctx.Input().Option("level")))
	if level != "" && !contains([]catalog.Level{catalog.LevelCompile, catalog.LevelHermetic, catalog.LevelScenario, catalog.LevelIntegration}, level) {
		return ctx.Fail("unknown coverage level: ", level)
	}
	status := catalog.Status(strings.TrimSpace(ctx.Input().Option("status")))
	if status != "" && !contains([]catalog.Status{catalog.StatusImplemented, catalog.StatusPlanned, catalog.StatusManual}, status) {
		return ctx.Fail("unknown coverage status: ", status)
	}

	entries := catalog.Filter(strings.TrimSpace(ctx.Input().Argument("feature")), level, status)
	if ctx.Input().OptionBool("json") {
		encoder := json.NewEncoder(console.OutputWriter(ctx.IO()))
		encoder.SetIndent("", "  ")
		return encoder.Encode(entries)
	}

	rows := make([][]string, 0, len(entries))
	counts := map[catalog.Status]int{}
	for _, item := range entries {
		requirements := "-"
		if len(item.Requirements) > 0 {
			requirements = strings.Join(item.Requirements, ",")
		}
		rows = append(rows, []string{
			item.Feature,
			item.Section,
			item.Example,
			string(item.Level),
			string(item.Status),
			requirements,
		})
		counts[item.Status]++
	}
	if len(rows) == 0 {
		ctx.IO().Warn("No documentation demos matched the selected filters.")
		return nil
	}
	if err := ctx.IO().Table([]string{"Feature", "Section", "Example", "Level", "Status", "Requires"}, rows); err != nil {
		return err
	}
	ctx.IO().NewLine()
	ctx.IO().Info(fmt.Sprintf(
		"Coverage entries: %d (implemented: %d, planned: %d, manual: %d)",
		len(entries), counts[catalog.StatusImplemented], counts[catalog.StatusPlanned], counts[catalog.StatusManual],
	))
	return nil
}

func contains[T comparable](values []T, wanted T) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
