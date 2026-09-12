package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	cachedemo "prismgo-demo/app/demo/cache"
	"prismgo-demo/app/demo/catalog"
)

// CacheCommand runs documentation-backed cache scenarios.
type CacheCommand struct {
	run func(context.Context, string) (cachedemo.Result, error)
}

// NewCacheCommand creates the cache demo command.
func NewCacheCommand() *CacheCommand {
	return &CacheCommand{run: cachedemo.Run}
}

// Definition describes the demo:cache command.
func (c *CacheCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:cache {case? : Scenario name or list} {--json : Output JSON}",
		"Run cache examples mapped to the Cache documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:cache list",
		"go run ./demo demo:cache get",
		"go run ./demo demo:cache named-store --json",
	}
	return definition
}

// Handle lists implemented cache scenarios or runs one by name.
func (c *CacheCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("cache", "", catalog.StatusImplemented)
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("cache", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented cache scenario %q", name))
	}
	result, err := c.run(ctx.Context(), name)
	if err != nil {
		return err
	}
	if ctx.Input().OptionBool("json") {
		return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(result)
	}
	return ctx.IO().Table([]string{"Case", "Key", "Value", "Details"}, [][]string{{result.Case, result.Key, result.Value, strings.Join(result.Details, "; ")}})
}
