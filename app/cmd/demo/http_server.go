package demo

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	"prismgo-demo/app/demo/httpserver"
)

// HTTPServerCommand runs documentation-backed HTTP server examples.
type HTTPServerCommand struct{}

// NewHTTPServerCommand creates the HTTP server demo command.
func NewHTTPServerCommand() *HTTPServerCommand { return &HTTPServerCommand{} }

// Definition describes the demo:http-server command.
func (c *HTTPServerCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:http-server {case? : Scenario name or list} {--json : Output JSON}",
		"Run HTTP server startup and lifecycle examples",
	)
	definition.Examples = []string{
		"go run ./demo demo:http-server list",
		"go run ./demo demo:http-server routes --json",
		"go run ./demo demo:http-server reload",
	}
	return definition
}

// Handle lists or executes an HTTP server example.
func (c *HTTPServerCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("http-server", "", catalog.StatusImplemented)
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(entries)
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("http-server", name)
	if !ok || entry.Status != catalog.StatusImplemented {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented http-server scenario %q", name))
	}
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate Demo executable: %w", err)
	}
	result, err := httpserver.Run(ctx.Context(), executable, name)
	if err != nil {
		return err
	}
	if ctx.Input().OptionBool("json") {
		return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(result)
	}
	return ctx.IO().Table([]string{"Case", "Observed"}, [][]string{{result.Case, result.Value}})
}
