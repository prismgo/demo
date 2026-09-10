package demo

import (
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
)

// ShowCommand renders detailed catalog entries for one or all features.
type ShowCommand struct{}

// NewShowCommand creates the detailed catalog command.
func NewShowCommand() *ShowCommand { return &ShowCommand{} }

// Definition describes the demo:show command.
func (c *ShowCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:show {feature? : Feature name; omit to show every entry} {--level= : compile, hermetic, scenario, integration} {--status= : implemented, planned, manual} {--json : Output JSON}",
		"Show framework demo implementation details",
	)
	definition.Examples = []string{
		"go run ./demo demo:show queue",
		"go run ./demo demo:show queue --status=planned",
		"go run ./demo demo:show queue --level=integration",
		"go run ./demo demo:show --status=planned --json",
	}
	return definition
}

// Handle renders detailed catalog entries matching the selected filters.
func (c *ShowCommand) Handle(ctx console.CommandContext) error {
	return renderCatalogDetails(ctx, strings.TrimSpace(ctx.Input().Argument("feature")))
}

func renderCatalogDetails(ctx console.CommandContext, feature string) error {
	if feature != "" {
		if _, ok := catalog.LookupFeature(feature); !ok {
			return ctx.Fail(unknownFeatureMessage(feature))
		}
	}
	level := catalog.Level(strings.TrimSpace(ctx.Input().Option("level")))
	if level != "" && !contains([]catalog.Level{catalog.LevelCompile, catalog.LevelHermetic, catalog.LevelScenario, catalog.LevelIntegration}, level) {
		return ctx.Fail("unknown coverage level: ", level)
	}
	status := catalog.Status(strings.TrimSpace(ctx.Input().Option("status")))
	if status != "" && !contains([]catalog.Status{catalog.StatusImplemented, catalog.StatusPlanned, catalog.StatusManual}, status) {
		return ctx.Fail("unknown coverage status: ", status)
	}

	entries := catalog.Filter(feature, level, status)
	output := detailOutput{Framework: frameworkVersion(), Feature: feature, Entries: entries}
	if feature != "" {
		summary, _ := catalog.SummaryFor(feature)
		output.Summary = &summary
	}
	if ctx.Input().OptionBool("json") {
		return encodeJSON(ctx, output)
	}
	if len(entries) == 0 {
		ctx.IO().Warn("No documentation demos matched the selected filters.")
		return nil
	}

	ctx.IO().Line("PrismGo Framework: " + frameworkVersion())
	if output.Summary != nil {
		ctx.IO().Line(fmt.Sprintf(
			"Module: %s — %d/%d implemented — %s",
			styledFeatureValue(ctx, output.Summary.Feature, output.Summary.Status),
			output.Summary.Implemented,
			output.Summary.Total,
			styledFeatureStatus(ctx, output.Summary.Status),
		))
	}
	ctx.IO().NewLine()
	rows := make([][]string, 0, len(entries))
	counts := map[catalog.Status]int{}
	for _, item := range entries {
		requirements := "-"
		if len(item.Requirements) > 0 {
			requirements = strings.Join(item.Requirements, ",")
		}
		row := make([]string, 0, 7)
		if feature == "" {
			row = append(row, styledEntryValue(ctx, item.Feature, item.Status))
		}
		row = append(row,
			styledEntryValue(ctx, item.Case, item.Status),
			item.Section,
			item.Since,
			item.Example,
			string(item.Level),
			requirements,
			styledEntryStatus(ctx, item.Status),
		)
		rows = append(rows, row)
		counts[item.Status]++
	}
	headers := []string{styledEntryValue(ctx, "Case", catalog.StatusPlanned), "Section", "Since", "Example", "Level", "Requires", "Status"}
	if feature == "" {
		headers = append([]string{styledEntryValue(ctx, "Feature", catalog.StatusPlanned)}, headers...)
	}
	if err := ctx.IO().Table(headers, rows); err != nil {
		return err
	}
	ctx.IO().NewLine()
	ctx.IO().Line(fmt.Sprintf(
		"Entries: %d | Implemented: %d | Planned: %d | Manual: %d",
		len(entries), counts[catalog.StatusImplemented], counts[catalog.StatusPlanned], counts[catalog.StatusManual],
	))
	return nil
}

type detailOutput struct {
	Framework string           `json:"framework"`
	Feature   string           `json:"feature,omitempty"`
	Summary   *catalog.Summary `json:"summary,omitempty"`
	Entries   []catalog.Entry  `json:"entries"`
}

func unknownFeatureMessage(name string) string {
	for _, feature := range catalog.Features() {
		if strings.HasPrefix(feature.Name, name) || strings.HasPrefix(name, feature.Name) {
			return fmt.Sprintf("unknown feature %q; did you mean %q?", name, feature.Name)
		}
	}
	return fmt.Sprintf("unknown feature %q; run demo:list to see available features", name)
}
