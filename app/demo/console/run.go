// Package consoledemo provides runnable examples of console definitions and input.
package consoledemo

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/prismgo/framework/console"
	"github.com/prismgo/framework/kernel"
	"github.com/spf13/cobra"
)

// Result records the actual definition or input observed by a console scenario.
type Result struct {
	Case        string             `json:"case"`
	Signature   string             `json:"signature,omitempty"`
	Definition  console.Definition `json:"definition"`
	Values      []string           `json:"values,omitempty"`
	Output      string             `json:"output,omitempty"`
	ErrorOutput string             `json:"error_output,omitempty"`
}

var signatures = map[string]string{
	"signature":               "mail:send {user : The user ID} {--Q|queue=default : Queue connection} {--force}",
	"required-argument":       "demo:sample {user}",
	"optional-argument":       "demo:sample {user?}",
	"default-argument":        "demo:sample {user=guest}",
	"required-array-argument": "demo:sample {ids*}",
	"optional-array-argument": "demo:sample {tags?*}",
	"argument-description":    "demo:sample {user : The user ID}",
	"boolean-option":          "demo:sample {--force}",
	"value-option":            "demo:sample {--queue=}",
	"default-option":          "demo:sample {--queue=default}",
	"short-option":            "demo:sample {--Q|queue=default}",
	"array-option":            "demo:sample {--id=*}",
	"option-description":      "demo:sample {--force : Force the operation}",
}

// Scenarios returns the supported cases in documentation order.
func Scenarios() []string {
	return []string{
		"architecture", "command-structure", "signature", "required-argument", "optional-argument",
		"default-argument", "required-array-argument", "optional-array-argument", "argument-description",
		"argument-validation", "boolean-option", "value-option", "default-option", "short-option",
		"array-option", "option-description", "option-validation", "must-definition", "parse-signature", "argument",
		"arguments", "missing-argument", "option", "option-strings", "option-bool", "option-int",
		"has-option", "optional-option-value", "line", "new-line", "info-comment-question-success",
		"warn-error", "alert", "table", "progress", "ask", "secret", "confirm", "choice", "choice-multiple",
		"choice-defaults-attempts", "anticipate", "handle-context", "from-context", "from-command", "cobra-command",
		"context-methods", "trap", "trap-release", "fail", "manual-failure", "call", "call-silently", "call-input",
		"isolatable", "missing-input", "missing-input-custom", "ansi-detection", "quiet-silent", "list-formats",
		"list-namespace", "list-raw-short", "argument-list", "normalize-definition", "clone-definition",
		"definition-usage", "bind-flags", "package-output", "package-exit", "laravel-compatibility",
	}
}

// Run executes a console example against the local framework checkout.
func Run(caseName string) (Result, error) {
	if isInputScenario(caseName) {
		return runInputScenario(caseName)
	}
	if isIOScenario(caseName) {
		return runIOScenario(caseName)
	}
	if isContextScenario(caseName) {
		return runContextScenario(caseName)
	}
	if isKernelScenario(caseName) {
		return runKernelScenario(caseName)
	}
	if isRenderScenario(caseName) {
		return runRenderScenario(caseName)
	}
	if caseName == "package-output" || caseName == "package-exit" {
		return runPackageScenario(caseName)
	}
	result := Result{Case: caseName}
	if signature, ok := signatures[caseName]; ok {
		definition, err := console.ParseSignature(signature)
		if err != nil {
			return Result{}, fmt.Errorf("console demo %s: %w", caseName, err)
		}
		result.Signature = signature
		result.Definition = definition
		return result, nil
	}

	switch caseName {
	case "architecture":
		result.Values = []string{"Command: Definition and Handle", "Definition: name, arguments, options", "Input: parsed values", "IO: terminal interaction", "Kernel: registration and dispatch"}
	case "command-structure":
		command := &sampleCommand{}
		k := kernel.New("demo")
		k.Register(command)
		definition := *command.Definition()
		input := console.NewInput(definition, &cobra.Command{}, []string{"Ada"})
		ctx := console.NewCommandContext(context.Background(), command, definition, input, console.NewIO(strings.NewReader(""), io.Discard, io.Discard), nil, nil)
		if err := command.Handle(ctx); err != nil {
			return Result{}, fmt.Errorf("console demo command-structure: %w", err)
		}
		result.Definition = *command.Definition()
		result.Values = append(result.Values, command.handled...)
	case "argument-validation":
		return validationResult(caseName, []string{
			"demo:sample {optional?} {required}",
			"demo:sample {ids*} {other?}",
			"demo:sample {user} {user?}",
		})
	case "option-validation":
		return validationResult(caseName, []string{
			"demo:sample {--force} {--force}",
			"demo:sample {--Q|queue=} {--Q|connection=}",
			"demo:sample {--LONG|queue=}",
			"demo:sample {--flag*}",
		})
	case "must-definition":
		result.Definition = *console.MustDefinition("demo:sample {user}", "  Send a message  ")
		result.Values = []string{mustDefinitionFailure("demo:sample {user")}
	case "parse-signature":
		definition, err := console.ParseSignature("demo:sample {user}")
		if err != nil {
			return Result{}, fmt.Errorf("console demo parse-signature: %w", err)
		}
		definition.Description = "Send a message"
		result.Definition = definition
		_, err = console.ParseSignature("demo:sample {user")
		if err == nil {
			return Result{}, fmt.Errorf("console demo parse-signature: invalid signature succeeded")
		}
		result.Values = []string{err.Error()}
	case "argument":
		definition, err := console.ParseSignature("demo:sample {user} {tags?*}")
		if err != nil {
			return Result{}, fmt.Errorf("console demo argument: %w", err)
		}
		input := console.NewInput(definition, &cobra.Command{}, []string{"Ada", "red", "blue"})
		result.Definition = definition
		result.Values = []string{input.Argument("user"), input.Argument("tags")}
	default:
		return Result{}, fmt.Errorf("console demo: unknown scenario %q", caseName)
	}
	return result, nil
}

func validationResult(caseName string, signatures []string) (Result, error) {
	result := Result{Case: caseName}
	for _, signature := range signatures {
		_, err := console.ParseSignature(signature)
		if err == nil {
			return Result{}, fmt.Errorf("console demo %s: invalid signature %q succeeded", caseName, signature)
		}
		result.Values = append(result.Values, err.Error())
	}
	return result, nil
}

func mustDefinitionFailure(signature string) (failure string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			failure = fmt.Sprint(recovered)
		}
	}()
	console.MustDefinition(signature, "invalid")
	return ""
}

type sampleCommand struct {
	handled []string
}

func (c *sampleCommand) Definition() *console.Definition {
	return console.MustDefinition("demo:sample {user}", "Handle a sample user")
}

func (c *sampleCommand) Handle(ctx console.CommandContext) error {
	c.handled = append(c.handled, strings.TrimSpace(ctx.Argument("user")))
	return nil
}
