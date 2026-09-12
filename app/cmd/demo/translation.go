package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"
	"github.com/prismgo/framework/container"

	translationdemo "prismgo-demo/app/demo/translation"
)

// TranslationCommand runs the documentation-backed translation examples.
type TranslationCommand struct {
	run func(context.Context, string) (translationdemo.Result, error)
}

// NewTranslationCommand creates the translation demo command.
func NewTranslationCommand() *TranslationCommand {
	return newTranslationCommand(func(_ context.Context, caseName string) (translationdemo.Result, error) {
		basePath, err := container.Make[string]("path.base")
		if err != nil {
			return translationdemo.Result{}, fmt.Errorf("translation demo resolve base path: %w", err)
		}
		return translationdemo.Run(caseName, basePath)
	})
}

func newTranslationCommand(run func(context.Context, string) (translationdemo.Result, error)) *TranslationCommand {
	return &TranslationCommand{run: run}
}

// Definition describes the demo:translation command.
func (c *TranslationCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:translation {case? : Scenario name or list} {--json : Output JSON}",
		"Run translation examples mapped to the Translation documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:translation list",
		"go run ./demo demo:translation short-keys",
		"go run ./demo demo:translation namespaces --json",
	}
	return definition
}

// Handle lists translation scenarios or runs the selected scenario.
func (c *TranslationCommand) Handle(ctx console.CommandContext) error {
	caseName := strings.TrimSpace(ctx.Input().Argument("case"))
	if caseName == "" || caseName == "list" {
		return renderTranslationScenarios(ctx, translationdemo.Scenarios())
	}
	if !hasTranslationScenario(caseName) {
		return ctx.Fail(fmt.Sprintf("unknown translation scenario %q", caseName))
	}

	result, err := c.run(ctx.Context(), caseName)
	if err != nil {
		return err
	}
	if ctx.Input().OptionBool("json") {
		encoder := json.NewEncoder(console.OutputWriter(ctx.IO()))
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}
	details := "-"
	if len(result.Details) > 0 {
		details = strings.Join(result.Details, " -> ")
	}
	return ctx.IO().Table(
		[]string{"Case", "Key", "Value", "Locale", "Details"},
		[][]string{{result.Case, result.Key, result.Value, result.Locale, details}},
	)
}

func hasTranslationScenario(name string) bool {
	for _, scenario := range translationdemo.Scenarios() {
		if scenario.Name == name {
			return true
		}
	}
	return false
}

func renderTranslationScenarios(ctx console.CommandContext, scenarios []translationdemo.Scenario) error {
	if ctx.Input().OptionBool("json") {
		encoder := json.NewEncoder(console.OutputWriter(ctx.IO()))
		encoder.SetIndent("", "  ")
		return encoder.Encode(scenarios)
	}
	rows := make([][]string, 0, len(scenarios))
	for _, scenario := range scenarios {
		rows = append(rows, []string{scenario.Name, scenario.Section, scenario.Status})
	}
	return ctx.IO().Table([]string{"Case", "Documentation", "Status"}, rows)
}
