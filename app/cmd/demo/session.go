package demo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	sessiondemo "prismgo-demo/app/demo/session"
)

// SessionCommand runs the documented Session examples.
type SessionCommand struct {
	run      func(string) (sessiondemo.Result, error)
	runRedis func(string) (sessiondemo.Result, error)
}

// NewSessionCommand creates the Session demo command.
func NewSessionCommand() *SessionCommand {
	return &SessionCommand{run: sessiondemo.Run, runRedis: sessiondemo.RunRedis}
}

// Definition describes the demo:session command.
func (c *SessionCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:session {case? : Scenario name or list} {--connection=file : file or redis} {--json : Output JSON}",
		"Run Session examples mapped to the Session documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:session list",
		"go run ./demo demo:session flash",
		"go run ./demo demo:session put --json",
		"go run ./demo demo:session redis-driver --connection=redis",
	}
	return definition
}

// Handle lists implemented Session scenarios or executes one by name.
func (c *SessionCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("session", "", catalog.StatusImplemented)
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("session", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented session scenario %q", name))
	}
	connection := strings.TrimSpace(ctx.Input().Option("connection"))
	if connection == "" {
		connection = "file"
	}
	result, err := c.execute(entry, connection)
	if err != nil {
		return err
	}
	if ctx.Input().OptionBool("json") {
		return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(result)
	}
	return ctx.IO().Table([]string{"Case", "Value"}, [][]string{{result.Case, result.Value}})
}

// execute routes one scenario to the local or Redis-backed runner by connection.
func (c *SessionCommand) execute(entry catalog.Entry, connection string) (sessiondemo.Result, error) {
	if entry.Level == catalog.LevelIntegration {
		if connection != "redis" {
			return sessiondemo.Result{}, fmt.Errorf("session scenario %q requires --connection=redis", entry.Case)
		}
		return c.runRedis(entry.Case)
	}
	if connection != "file" {
		return sessiondemo.Result{}, fmt.Errorf("session scenario %q does not support --connection=%s", entry.Case, connection)
	}
	return c.run(entry.Case)
}
