package demo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	redisdemo "prismgo-demo/app/demo/redis"
)

// RedisCommand runs the documented Redis examples.
type RedisCommand struct {
	run            func(string) (redisdemo.Result, error)
	runIntegration func(string) (redisdemo.Result, error)
}

// NewRedisCommand creates the Redis demo command.
func NewRedisCommand() *RedisCommand {
	return &RedisCommand{run: redisdemo.Run, runIntegration: redisdemo.RunIntegration}
}

// Definition describes the demo:redis command.
func (c *RedisCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:redis {case? : Scenario name or list} {--json : Output JSON}",
		"Run Redis examples mapped to the Redis documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:redis list",
		"go run ./demo demo:redis config",
		"go run ./demo demo:redis strings --json",
		"go run ./demo demo:redis publish",
	}
	return definition
}

// Handle lists implemented Redis scenarios or executes one by name.
func (c *RedisCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("redis", "", catalog.StatusImplemented)
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("redis", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented redis scenario %q", name))
	}
	result, err := c.execute(entry)
	if err != nil {
		return err
	}
	if ctx.Input().OptionBool("json") {
		return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(result)
	}
	return ctx.IO().Table([]string{"Case", "Value"}, [][]string{{result.Case, result.Value}})
}

// execute routes one scenario to the hermetic or integration runner by level.
func (c *RedisCommand) execute(entry catalog.Entry) (redisdemo.Result, error) {
	if entry.Level == catalog.LevelIntegration {
		return c.runIntegration(entry.Case)
	}
	return c.run(entry.Case)
}
