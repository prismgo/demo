// Package demo implements documentation-backed demo commands.
package demo

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/prismgo/framework/console"
	"github.com/prismgo/framework/version"

	"prismgo-demo/app/demo/catalog"
)

// ListCommand renders the feature-level documentation demo coverage catalog.
type ListCommand struct{}

// NewListCommand creates the demo discovery command.
func NewListCommand() *ListCommand { return &ListCommand{} }

// Definition describes the demo:list command.
func (c *ListCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:list {feature? : Show details for one feature} {--level= : Detail filter: compile, hermetic, scenario, integration} {--status= : Feature or detail status filter} {--json : Output JSON}",
		"List framework demo coverage by feature",
	)
	definition.Examples = []string{
		"go run ./demo demo:list",
		"go run ./demo demo:list --status=in-progress",
		"go run ./demo demo:list --json",
		"go run ./demo demo:list queue",
	}
	return definition
}

// Handle lists feature summaries or preserves the feature detail shortcut.
func (c *ListCommand) Handle(ctx console.CommandContext) error {
	feature := strings.TrimSpace(ctx.Input().Argument("feature"))
	if feature != "" {
		return renderCatalogDetails(ctx, feature)
	}
	if level := strings.TrimSpace(ctx.Input().Option("level")); level != "" {
		return ctx.Fail("the --level filter applies to details; use demo:show --level=", level)
	}

	status, err := featureStatusFilter(ctx.Input().Option("status"))
	if err != nil {
		return ctx.Fail(err)
	}
	summaries := catalog.Summaries()
	if status != "" {
		filtered := make([]catalog.Summary, 0, len(summaries))
		for _, summary := range summaries {
			if summary.Status == status {
				filtered = append(filtered, summary)
			}
		}
		summaries = filtered
	}
	if ctx.Input().OptionBool("json") {
		return encodeJSON(ctx, overviewOutput{Framework: frameworkVersion(), Features: summaries})
	}
	if len(summaries) == 0 {
		ctx.IO().Warn("No framework features matched the selected status.")
		return nil
	}

	ctx.IO().Line("PrismGo Framework: " + frameworkVersion() + " (local workspace)")
	ctx.IO().NewLine()
	rows := make([][]string, 0, len(summaries))
	implemented := 0
	planned := 0
	manual := 0
	total := 0
	for _, summary := range summaries {
		remaining := strconv.Itoa(summary.Remaining)
		if summary.Status == catalog.FeatureStatusManual {
			remaining = "-"
		}
		rows = append(rows, []string{
			styledFeatureValue(ctx, summary.Feature, summary.Status),
			summary.Description,
			summary.Since,
			fmt.Sprintf("%d/%d", summary.Implemented, summary.Total),
			remaining,
			styledFeatureStatus(ctx, summary.Status),
		})
		implemented += summary.Implemented
		planned += summary.Planned
		manual += summary.Manual
		total += summary.Total
	}
	headers := []string{styledFeatureValue(ctx, "Feature", catalog.FeatureStatusPlanned), "Description", "Since", "Progress", "Remaining", "Status"}
	if err := ctx.IO().Table(headers, rows); err != nil {
		return err
	}
	ctx.IO().NewLine()
	ctx.IO().Line(fmt.Sprintf(
		"Modules: %d | Implemented: %d/%d | Planned: %d | Manual: %d",
		len(summaries), implemented, total, planned, manual,
	))
	ctx.IO().NewLine()
	ctx.IO().Line("View details: go run ./demo demo:show <feature>")
	ctx.IO().Line("Example:      go run ./demo demo:show queue --status=planned")
	return nil
}

type overviewOutput struct {
	Framework string            `json:"framework"`
	Features  []catalog.Summary `json:"features"`
}

func featureStatusFilter(value string) (catalog.FeatureStatus, error) {
	normalized := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(value)), "-", "_")
	switch catalog.FeatureStatus(normalized) {
	case "":
		return "", nil
	case catalog.FeatureStatusImplemented, catalog.FeatureStatusInProgress, catalog.FeatureStatusPlanned, catalog.FeatureStatusManual:
		return catalog.FeatureStatus(normalized), nil
	default:
		return "", fmt.Errorf("unknown feature status %q: expected implemented, in-progress, planned, or manual", value)
	}
}

func styledFeatureStatus(ctx console.CommandContext, status catalog.FeatureStatus) string {
	label := strings.ReplaceAll(string(status), "_", " ")
	return styledFeatureValue(ctx, label, status)
}

func styledFeatureValue(ctx console.CommandContext, value string, status catalog.FeatureStatus) string {
	options := console.OutputOptionsForIO(ctx.IO())
	switch status {
	case catalog.FeatureStatusImplemented:
		return console.Styled(value, console.StyleInfo, options)
	case catalog.FeatureStatusInProgress:
		return console.Styled(value, console.StyleYellow, options)
	default:
		if !options.ANSI || value == "" {
			return value
		}
		// Keep every Feature cell's ANSI overhead equal so tabwriter measures
		// colored and default-foreground rows consistently.
		return "\x1b[39m" + value + "\x1b[0m"
	}
}

func styledEntryStatus(ctx console.CommandContext, status catalog.Status) string {
	return styledEntryValue(ctx, string(status), status)
}

func styledEntryValue(ctx console.CommandContext, value string, status catalog.Status) string {
	if status == catalog.StatusImplemented {
		return console.Styled(value, console.StyleInfo, console.OutputOptionsForIO(ctx.IO()))
	}
	options := console.OutputOptionsForIO(ctx.IO())
	if options.ANSI && value != "" {
		return "\x1b[39m" + value + "\x1b[0m"
	}
	return value
}

func frameworkVersion() string {
	return "v" + version.Framework
}

func encodeJSON(ctx console.CommandContext, value any) error {
	encoder := json.NewEncoder(console.OutputWriter(ctx.IO()))
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func contains[T comparable](values []T, wanted T) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
