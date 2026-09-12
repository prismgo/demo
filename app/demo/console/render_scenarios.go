package consoledemo

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"
	"github.com/spf13/cobra"
)

func isRenderScenario(name string) bool {
	switch name {
	case "ansi-detection", "quiet-silent", "list-formats", "list-namespace", "list-raw-short", "argument-list", "normalize-definition", "clone-definition", "definition-usage", "bind-flags", "laravel-compatibility":
		return true
	}
	return false
}

func renderDefinitions() []console.Definition {
	return []console.Definition{
		{Name: "mail:send", Description: "Send mail", Aliases: []string{"send-mail"}},
		{Name: "mail:queue", Description: "Queue mail"},
		{Name: "cache:clear", Description: "Clear cache"},
		{Name: "hidden:task", Hidden: true},
	}
}

func renderList(definitions []console.Definition, opts console.CommandListOptions) (string, error) {
	var out bytes.Buffer
	if err := console.RenderCommandList(&out, definitions, opts); err != nil {
		return "", err
	}
	return out.String(), nil
}

func runRenderScenario(name string) (Result, error) {
	result := Result{Case: name}
	switch name {
	case "ansi-detection":
		return runKernelScenario(name)
	case "quiet-silent":
		for _, opts := range []console.CommandListOptions{{Output: console.OutputOptions{Quiet: true}}, {Output: console.OutputOptions{Silent: true}}} {
			out, err := renderList(renderDefinitions(), opts)
			if err != nil {
				return Result{}, fmt.Errorf("console demo %s: %w", name, err)
			}
			result.Values = append(result.Values, fmt.Sprintf("empty=%t", out == ""))
		}
	case "list-formats":
		for _, format := range []string{"txt", "json", "md"} {
			out, err := renderList(renderDefinitions(), console.CommandListOptions{AppName: "Demo", Description: "Demo commands", Format: format})
			if err != nil {
				return Result{}, fmt.Errorf("console demo list format %s: %w", format, err)
			}
			result.Values = append(result.Values, format+":"+out)
		}
	case "list-namespace":
		out, err := renderList(renderDefinitions(), console.CommandListOptions{Namespace: "mail", Format: "json"})
		if err != nil {
			return Result{}, fmt.Errorf("console demo list namespace: %w", err)
		}
		result.Output = out
	case "list-raw-short":
		for _, opts := range []console.CommandListOptions{{Raw: true}, {Short: true}} {
			out, err := renderList(renderDefinitions(), opts)
			if err != nil {
				return Result{}, fmt.Errorf("console demo list raw-short: %w", err)
			}
			result.Values = append(result.Values, out)
		}
	case "argument-list":
		args := []console.Argument{{Name: "user", Description: "User ID", Required: true}, {Name: "tags", Description: "Tags", IsArray: true}}
		for _, item := range console.ArgumentDescriptors(args) {
			result.Values = append(result.Values, item.Synopsis+":"+item.Description)
		}
		var out bytes.Buffer
		console.RenderArgumentList(&out, args, console.OutputOptions{})
		result.Output = out.String()
	case "normalize-definition":
		definition := console.Definition{Name: "  mail:send  ", Description: "  Send mail  ", Aliases: []string{" send ", "send", ""}, Examples: []string{" demo ", "demo"}}
		normalized, err := console.NormalizeDefinition(definition)
		if err != nil {
			return Result{}, fmt.Errorf("console demo normalize-definition: %w", err)
		}
		result.Definition = normalized
		_, invalid := console.NormalizeDefinition(console.Definition{})
		if invalid == nil {
			return Result{}, fmt.Errorf("console demo normalize-definition: empty name succeeded")
		}
		result.Values = []string{invalid.Error()}
	case "clone-definition":
		original := console.Definition{Name: "mail:send", Arguments: []console.Argument{{Name: "user", Suggestions: []string{"Ada"}}}, Options: []console.Option{{Name: "queue", Suggestions: []string{"sync"}}}, Aliases: []string{"send"}, Examples: []string{"mail:send Ada"}}
		cloned := console.CloneDefinition(original)
		cloned.Arguments[0].Suggestions[0] = "changed"
		cloned.Options[0].Suggestions[0] = "changed"
		cloned.Aliases[0] = "changed"
		cloned.Examples[0] = "changed"
		result.Definition = original
		result.Values = []string{original.Arguments[0].Suggestions[0], original.Options[0].Suggestions[0], original.Aliases[0], original.Examples[0]}
	case "definition-usage":
		definition := *console.MustDefinition("mail:send {user} {tags?*}", "Send mail")
		result.Values = []string{console.DefinitionUsage(definition)}
		definition.UsageText = "custom usage"
		result.Values = append(result.Values, console.DefinitionUsage(definition))
	case "bind-flags":
		definition := *console.MustDefinition("mail:send {--f|force} {--Q|queue=default} {--id=*}", "Send mail")
		cmd := &cobra.Command{Use: definition.Name}
		if err := console.BindDefinitionFlags(cmd, definition); err != nil {
			return Result{}, fmt.Errorf("console demo bind-flags: %w", err)
		}
		if err := cmd.ParseFlags([]string{"-f", "-Q", "--id=alpha", "--id=beta"}); err != nil {
			return Result{}, fmt.Errorf("console demo bind-flags parse: %w", err)
		}
		input := console.NewInput(definition, cmd, nil)
		result.Definition = definition
		result.Values = []string{fmt.Sprintf("force=%t", input.OptionBool("force")), fmt.Sprintf("queue=%q", input.Option("queue")), strings.Join(input.OptionStrings("id"), ","), cmd.Flags().Lookup("id").Value.Type()}
	case "laravel-compatibility":
		definition := *console.MustDefinition("mail:send {user} {--queue=default}", "Send mail")
		result.Definition = definition
		result.Values = []string{"Definition/Handle", "Argument/Option", "IO/Call", "queued commands use queue integration"}
	}
	return result, nil
}
