package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
	filesystemdemo "prismgo-demo/app/demo/filesystem"
)

// FilesystemCommand runs documentation-backed filesystem scenarios.
type FilesystemCommand struct {
	run func(context.Context, string) (filesystemdemo.Result, error)
}

// NewFilesystemCommand creates the filesystem demo command.
func NewFilesystemCommand() *FilesystemCommand {
	return &FilesystemCommand{run: filesystemdemo.Run}
}

// Definition describes the demo:filesystem command.
func (c *FilesystemCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:filesystem {case? : Scenario name or list} {--disk= : Disk for OSS scenarios} {--json : Output JSON}",
		"Run filesystem examples mapped to the Filesystem documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:filesystem list",
		"go run ./demo demo:filesystem get",
		"go run ./demo demo:filesystem named-disk --json",
	}
	return definition
}

// Handle lists implemented filesystem scenarios or runs one by name.
func (c *FilesystemCommand) Handle(ctx console.CommandContext) error {
	name := strings.TrimSpace(ctx.Input().Argument("case"))
	if name == "" || name == "list" {
		entries := catalog.Filter("filesystem", "", catalog.StatusImplemented)
		runnable := make([]catalog.Entry, 0, len(entries))
		for _, entry := range entries {
			if strings.HasPrefix(entry.Example, "demo:filesystem ") {
				runnable = append(runnable, entry)
			}
		}
		if ctx.Input().OptionBool("json") {
			return json.NewEncoder(console.OutputWriter(ctx.IO())).Encode(runnable)
		}
		rows := make([][]string, 0, len(runnable))
		for _, entry := range runnable {
			rows = append(rows, []string{entry.Case, entry.Section, string(entry.Level)})
		}
		return ctx.IO().Table([]string{"Case", "Documentation", "Level"}, rows)
	}
	entry, ok := catalog.Find("filesystem", name)
	if !ok || entry.Status != catalog.StatusImplemented || !strings.HasPrefix(entry.Example, "demo:filesystem ") {
		return ctx.Fail(fmt.Sprintf("unknown or unimplemented filesystem scenario %q", name))
	}
	if selected := strings.TrimSpace(ctx.Input().Option("disk")); selected != "" {
		if !slices.Contains(entry.Requirements, "oss") || selected != "oss" {
			return ctx.Fail(fmt.Sprintf("filesystem scenario %q does not use disk %q", name, selected))
		}
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
