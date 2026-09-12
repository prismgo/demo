package consoledemo

import (
	"fmt"

	"github.com/prismgo/framework/console"
	"github.com/spf13/cobra"
)

func isInputScenario(caseName string) bool {
	switch caseName {
	case "arguments", "missing-argument", "option", "option-strings", "option-bool", "option-int", "has-option", "optional-option-value":
		return true
	}
	return false
}

func scenarioInput(signature string, args, flags []string) (console.Definition, console.Input, error) {
	definition, err := console.ParseSignature(signature)
	if err != nil {
		return console.Definition{}, nil, fmt.Errorf("parse console signature %q: %w", signature, err)
	}
	cmd := &cobra.Command{Use: definition.Name}
	if err := console.BindDefinitionFlags(cmd, definition); err != nil {
		return console.Definition{}, nil, fmt.Errorf("bind console flags for %q: %w", signature, err)
	}
	if err := cmd.ParseFlags(flags); err != nil {
		return console.Definition{}, nil, fmt.Errorf("parse console flags for %q: %w", signature, err)
	}
	return definition, console.NewInput(definition, cmd, args), nil
}

func runInputScenario(caseName string) (Result, error) {
	result := Result{Case: caseName}
	var signature string
	var args, flags []string
	switch caseName {
	case "arguments":
		signature, args = "demo:sample {user} {tags?*}", []string{"Ada", "red", "blue"}
	case "missing-argument":
		signature = "demo:sample {user?} {tags?*}"
	case "option":
		signature, flags = "demo:sample {--queue=}", []string{"--queue=priority"}
	case "option-strings":
		signature, flags = "demo:sample {--id=*}", []string{"--id=alpha", "--id=beta"}
	case "option-bool":
		signature, flags = "demo:sample {--force} {--enabled=}", []string{"--force", "--enabled=true"}
	case "option-int":
		signature, flags = "demo:sample {--retries=}", []string{"--retries=42"}
	case "has-option":
		signature = "demo:sample {--force}"
	case "optional-option-value":
		signature, flags = "demo:sample {--queue=default}", []string{"--queue"}
	}
	definition, input, err := scenarioInput(signature, args, flags)
	if err != nil {
		return Result{}, fmt.Errorf("console demo %s: %w", caseName, err)
	}
	result.Signature, result.Definition = signature, definition
	switch caseName {
	case "arguments":
		result.Values = input.Arguments("tags")
	case "missing-argument":
		result.Values = []string{input.Argument("user"), fmt.Sprintf("tags_nil=%t", input.Arguments("tags") == nil)}
	case "option":
		_, unsetInput, err := scenarioInput(signature, nil, nil)
		if err != nil {
			return Result{}, fmt.Errorf("console demo option unset fixture: %w", err)
		}
		result.Values = []string{input.Option("queue"), unsetInput.Option("queue")}
	case "option-strings":
		result.Values = input.OptionStrings("id")
	case "option-bool":
		result.Values = []string{fmt.Sprintf("force=%t", input.OptionBool("force")), fmt.Sprintf("enabled=%t", input.OptionBool("enabled")), fmt.Sprintf("absent=%t", input.OptionBool("absent"))}
	case "option-int":
		value, err := input.OptionInt("retries")
		if err != nil {
			return Result{}, fmt.Errorf("console demo option-int: %w", err)
		}
		result.Values = []string{fmt.Sprintf("retries=%d", value)}
		_, invalidInput, err := scenarioInput(signature, nil, []string{"--retries=invalid"})
		if err != nil {
			return Result{}, fmt.Errorf("console demo option-int invalid fixture: %w", err)
		}
		_, err = invalidInput.OptionInt("retries")
		if err == nil {
			return Result{}, fmt.Errorf("console demo option-int: invalid integer succeeded")
		}
		result.Values = append(result.Values, err.Error())
	case "has-option":
		result.Values = []string{fmt.Sprintf("force=%t", input.HasOption("force")), fmt.Sprintf("absent=%t", input.HasOption("absent"))}
		parent := &cobra.Command{Use: "root"}
		parent.PersistentFlags().String("global", "default", "global option")
		child := &cobra.Command{Use: "child"}
		parent.AddCommand(child)
		inherited := console.NewInput(definition, child, nil)
		result.Values = append(result.Values, fmt.Sprintf("global=%t", inherited.HasOption("global")))
	case "optional-option-value":
		result.Values = []string{input.Option("queue"), fmt.Sprintf("strings=%q", input.OptionStrings("queue")), fmt.Sprintf("present=%t", input.HasOption("queue"))}
	}
	return result, nil
}
