package demo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	ratelimitdemo "prismgo-demo/app/demo/ratelimit"
)

// RateLimitCommand runs the documented rate limiter examples.
type RateLimitCommand struct {
	run func(string) (ratelimitdemo.Result, error)
}

// NewRateLimitCommand creates the RateLimit demo command.
func NewRateLimitCommand() *RateLimitCommand {
	return &RateLimitCommand{run: ratelimitdemo.Run}
}

// Definition describes the demo:ratelimit command.
func (c *RateLimitCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:ratelimit {case? : Scenario name or list} {--store=memory : Limiter cache store} {--json : Output JSON}",
		"Run rate limiter examples mapped to the Rate Limiting documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:ratelimit list",
		"go run ./demo demo:ratelimit architecture",
		"go run ./demo demo:ratelimit per-minute --json",
		"go run ./demo demo:ratelimit redis-store --store=redis",
	}
	return definition
}

// Handle lists implemented rate limiter scenarios or executes one by name.
func (c *RateLimitCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("ratelimit", "", catalog.StatusImplemented)
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("ratelimit", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented ratelimit scenario %q", name))
	}
	store := strings.TrimSpace(ctx.Input().Option("store"))
	if store == "" {
		store = "memory"
	}
	if entry.Level == catalog.LevelIntegration {
		if store != "redis" {
			return ctx.Fail(fmt.Sprintf("ratelimit scenario %q requires --store=redis", entry.Case))
		}
	} else if store != "memory" {
		return ctx.Fail(fmt.Sprintf("ratelimit scenario %q does not support --store=%s", entry.Case, store))
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
