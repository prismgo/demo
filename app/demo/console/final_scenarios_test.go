package consoledemo

import (
	"reflect"
	"strings"
	"testing"
)

func checkFinalScenario(t *testing.T, name string) {
	t.Helper()
	tests := map[string]struct {
		values      []string
		output      []string
		errorOutput []string
	}{
		"choice-defaults-attempts": {values: []string{"red", "blue", "blue", "invalid choice"}, output: []string{"[default: red,blue]", "Color: Color:"}},
		"anticipate":               {values: []string{"Taylor", "Dayle"}, output: []string{"Name [default: Dayle]: "}},
		"handle-context":           {values: []string{"demo:sample", "Ada"}},
		"from-context":             {values: []string{"bound=true", "injected=true", "missing=false"}},
		"from-command":             {values: []string{"found=true", "required=true", "missing=false", "not available"}},
		"cobra-command":            {values: []string{"same=true", "nil=true"}},
		"context-methods":          {values: []string{"demo:sample", "demo:sample", "Ada", "io=true", "context=true"}},
		"trap":                     {values: []string{"interrupt"}},
		"trap-release":             {values: []string{"interrupt", "released"}},
		"fail":                     {values: []string{"stopped", "disk unavailable", "write failed: disk unavailable", "wrapped=true"}},
		"manual-failure":           {values: []string{"recognized=true", "stopped", "ordinary=false"}},
		"call":                     {values: []string{"Ada", "priority", "discard=false"}, output: []string{"target output"}},
		"call-silently":            {values: []string{"Ada", "priority", "discard=true"}},
		"call-input":               {values: []string{"Ada", "priority", "discard=false"}, output: []string{"target output"}},
		"isolatable":               {values: []string{"first=<nil>", "already running", "count=1"}},
		"missing-input":            {values: []string{"Ada"}, output: []string{"What is User name?"}},
		"missing-input-custom":     {values: []string{"hook=Ada:dev", "Ada", "dev"}, output: []string{"Custom user", "Custom team [default: dev]"}},
		"ansi-detection":           {values: []string{"plain=false", "force-color=true", "no-color=false", "forced=true", "disabled=false", "options={ANSI:true Quiet:true Silent:false}"}},
		"quiet-silent":             {values: []string{"empty=true", "empty=true"}},
		"list-formats":             {values: []string{"txt:Demo commands", "json:{", "md:# Demo"}},
		"list-namespace":           {output: []string{"mail:send", "mail:queue", `"namespace": "mail"`}},
		"list-raw-short":           {values: []string{"cache:clear Clear cache", "cache:clear\n"}},
		"argument-list":            {values: []string{"user:User ID", "tags...:Tags"}, output: []string{"user", "User ID", "tags...", "Tags"}},
		"normalize-definition":     {values: []string{"command name is required"}},
		"clone-definition":         {values: []string{"Ada", "sync", "send", "mail:send Ada"}},
		"definition-usage":         {values: []string{"mail:send <user> [tags...]", "custom usage"}},
		"bind-flags":               {values: []string{"force=true", "queue=\"\"", "alpha,beta", "stringArray"}},
		"package-output":           {output: []string{"plain", "info", "comment", "question", "success", "warning", "error", "alert"}},
		"package-exit":             {values: []string{"package-exit=1:exit failure", "package-exit-if=1:conditional failure", "nil=continued"}},
		"laravel-compatibility":    {values: []string{"Definition/Handle", "Argument/Option", "IO/Call", "queued commands use queue integration"}},
	}
	tt, ok := tests[name]
	if !ok {
		t.Fatalf("console case %q has no expectation, want one", name)
	}
	result := runCase(t, name)
	if len(result.Values) != len(tt.values) {
		t.Fatalf("console %s values = %#v, want %d values %#v", name, result.Values, len(tt.values), tt.values)
	}
	for i, want := range tt.values {
		if !strings.Contains(result.Values[i], want) {
			t.Errorf("console %s value[%d] = %q, want substring %q", name, i, result.Values[i], want)
		}
	}
	for _, want := range tt.output {
		if !strings.Contains(result.Output, want) {
			t.Errorf("console %s stdout = %q, want substring %q", name, result.Output, want)
		}
	}
	for _, want := range tt.errorOutput {
		if !strings.Contains(result.ErrorOutput, want) {
			t.Errorf("console %s stderr = %q, want substring %q", name, result.ErrorOutput, want)
		}
	}
	if name == "call-silently" && strings.Contains(result.Output, "target output") {
		t.Errorf("console %s stdout = %q, want target output suppressed", name, result.Output)
	}
	if name == "list-namespace" && strings.Contains(result.Output, "cache:clear") {
		t.Errorf("console %s stdout = %q, want cache namespace excluded", name, result.Output)
	}
	if name == "normalize-definition" && (result.Definition.Name != "mail:send" || !reflect.DeepEqual(result.Definition.Aliases, []string{"send"}) || !reflect.DeepEqual(result.Definition.Examples, []string{"demo"})) {
		t.Errorf("console %s definition = %#v, want trimmed name and deduplicated aliases/examples", name, result.Definition)
	}
}

func TestConsoleDemoFinalScenarioOrder(t *testing.T) {
	want := []string{"choice-defaults-attempts", "anticipate", "handle-context", "from-context", "from-command", "cobra-command", "context-methods", "trap", "trap-release", "fail", "manual-failure", "call", "call-silently", "call-input", "isolatable", "missing-input", "missing-input-custom", "ansi-detection", "quiet-silent", "list-formats", "list-namespace", "list-raw-short", "argument-list", "normalize-definition", "clone-definition", "definition-usage", "bind-flags", "package-output", "package-exit", "laravel-compatibility"}
	if got := Scenarios()[40:]; !reflect.DeepEqual(got, want) {
		t.Errorf("final console scenarios = %#v, want %#v", got, want)
	}
}

func TestConsoleDemoChoiceDefaultsAttempts(t *testing.T) {
	checkFinalScenario(t, "choice-defaults-attempts")
}
func TestConsoleDemoAnticipate(t *testing.T)          { checkFinalScenario(t, "anticipate") }
func TestConsoleDemoHandleContext(t *testing.T)       { checkFinalScenario(t, "handle-context") }
func TestConsoleDemoFromContext(t *testing.T)         { checkFinalScenario(t, "from-context") }
func TestConsoleDemoFromCommand(t *testing.T)         { checkFinalScenario(t, "from-command") }
func TestConsoleDemoCobraCommand(t *testing.T)        { checkFinalScenario(t, "cobra-command") }
func TestConsoleDemoContextMethods(t *testing.T)      { checkFinalScenario(t, "context-methods") }
func TestConsoleDemoTrap(t *testing.T)                { checkFinalScenario(t, "trap") }
func TestConsoleDemoTrapRelease(t *testing.T)         { checkFinalScenario(t, "trap-release") }
func TestConsoleDemoFail(t *testing.T)                { checkFinalScenario(t, "fail") }
func TestConsoleDemoManualFailure(t *testing.T)       { checkFinalScenario(t, "manual-failure") }
func TestConsoleDemoCall(t *testing.T)                { checkFinalScenario(t, "call") }
func TestConsoleDemoCallSilently(t *testing.T)        { checkFinalScenario(t, "call-silently") }
func TestConsoleDemoCallInput(t *testing.T)           { checkFinalScenario(t, "call-input") }
func TestConsoleDemoIsolatable(t *testing.T)          { checkFinalScenario(t, "isolatable") }
func TestConsoleDemoMissingInput(t *testing.T)        { checkFinalScenario(t, "missing-input") }
func TestConsoleDemoMissingInputCustom(t *testing.T)  { checkFinalScenario(t, "missing-input-custom") }
func TestConsoleDemoANSIDetection(t *testing.T)       { checkFinalScenario(t, "ansi-detection") }
func TestConsoleDemoQuietSilent(t *testing.T)         { checkFinalScenario(t, "quiet-silent") }
func TestConsoleDemoListFormats(t *testing.T)         { checkFinalScenario(t, "list-formats") }
func TestConsoleDemoListNamespace(t *testing.T)       { checkFinalScenario(t, "list-namespace") }
func TestConsoleDemoListRawShort(t *testing.T)        { checkFinalScenario(t, "list-raw-short") }
func TestConsoleDemoArgumentList(t *testing.T)        { checkFinalScenario(t, "argument-list") }
func TestConsoleDemoNormalizeDefinition(t *testing.T) { checkFinalScenario(t, "normalize-definition") }
func TestConsoleDemoCloneDefinition(t *testing.T)     { checkFinalScenario(t, "clone-definition") }
func TestConsoleDemoDefinitionUsage(t *testing.T)     { checkFinalScenario(t, "definition-usage") }
func TestConsoleDemoBindFlags(t *testing.T)           { checkFinalScenario(t, "bind-flags") }
func TestConsoleDemoPackageOutput(t *testing.T)       { checkFinalScenario(t, "package-output") }
func TestConsoleDemoPackageExit(t *testing.T)         { checkFinalScenario(t, "package-exit") }
func TestConsoleDemoLaravelCompatibility(t *testing.T) {
	checkFinalScenario(t, "laravel-compatibility")
}
