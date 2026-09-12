package consoledemo

import (
	"reflect"
	"strings"
	"testing"

	"github.com/prismgo/framework/console"

	"prismgo-demo/app/demo/catalog"
)

func runCase(t *testing.T, caseName string) Result {
	t.Helper()
	result, err := Run(caseName)
	if err != nil {
		t.Fatalf("Run(%q) error = %v, want nil", caseName, err)
	}
	if result.Case != caseName {
		t.Fatalf("Run(%q) case = %q, want %q", caseName, result.Case, caseName)
	}
	return result
}

func checkArgument(t *testing.T, caseName, name string, required, array bool, defaultValue, description string) {
	t.Helper()
	result := runCase(t, caseName)
	if len(result.Definition.Arguments) != 1 {
		t.Fatalf("%s arguments = %#v, want one", caseName, result.Definition.Arguments)
	}
	arg := result.Definition.Arguments[0]
	if arg.Name != name || arg.Required != required || arg.IsArray != array || arg.Description != description {
		t.Errorf("%s argument = %#v, want name=%q required=%t array=%t description=%q", caseName, arg, name, required, array, description)
	}
	if defaultValue == "" && arg.DefaultValue != nil {
		t.Errorf("%s default = %q, want nil", caseName, *arg.DefaultValue)
	}
	if defaultValue != "" && (arg.DefaultValue == nil || *arg.DefaultValue != defaultValue) {
		t.Errorf("%s default = %v, want %q", caseName, arg.DefaultValue, defaultValue)
	}
}

func checkOption(t *testing.T, caseName, name, shortcut string, mode console.OptionValueMode, array bool, defaultValue, description string) {
	t.Helper()
	result := runCase(t, caseName)
	if len(result.Definition.Options) != 1 {
		t.Fatalf("%s options = %#v, want one", caseName, result.Definition.Options)
	}
	opt := result.Definition.Options[0]
	if opt.Name != name || opt.Shortcut != shortcut || opt.ValueMode != mode || opt.IsArray != array || opt.Description != description {
		t.Errorf("%s option = %#v, want name=%q shortcut=%q mode=%d array=%t description=%q", caseName, opt, name, shortcut, mode, array, description)
	}
	if defaultValue == "" && opt.DefaultValue != nil {
		t.Errorf("%s default = %q, want nil", caseName, *opt.DefaultValue)
	}
	if defaultValue != "" && (opt.DefaultValue == nil || *opt.DefaultValue != defaultValue) {
		t.Errorf("%s default = %v, want %q", caseName, opt.DefaultValue, defaultValue)
	}
}

func checkErrors(t *testing.T, caseName string, fragments []string) {
	t.Helper()
	result := runCase(t, caseName)
	if len(result.Values) != len(fragments) {
		t.Fatalf("%s errors = %#v, want %d errors", caseName, result.Values, len(fragments))
	}
	for index, fragment := range fragments {
		if !strings.Contains(result.Values[index], fragment) {
			t.Errorf("%s error[%d] = %q, want substring %q", caseName, index, result.Values[index], fragment)
		}
	}
}

func TestConsoleDemoArchitecture(t *testing.T) {
	result := runCase(t, "architecture")
	if len(result.Values) != 5 {
		t.Errorf("architecture concepts = %#v, want five responsibilities", result.Values)
	}
}

func TestConsoleDemoCommandStructure(t *testing.T) {
	result := runCase(t, "command-structure")
	if result.Definition.Name != "demo:sample" || !reflect.DeepEqual(result.Values, []string{"Ada"}) {
		t.Errorf("command-structure = %#v, want demo:sample Handle invoked with Ada", result)
	}
}

func TestConsoleDemoSignature(t *testing.T) {
	result := runCase(t, "signature")
	if result.Definition.Name != "mail:send" || len(result.Definition.Arguments) != 1 || len(result.Definition.Options) != 2 {
		t.Errorf("signature definition = %#v, want mail:send with one argument and two options", result.Definition)
	}
}

func TestConsoleDemoRequiredArgument(t *testing.T) {
	checkArgument(t, "required-argument", "user", true, false, "", "")
}
func TestConsoleDemoOptionalArgument(t *testing.T) {
	checkArgument(t, "optional-argument", "user", false, false, "", "")
}
func TestConsoleDemoDefaultArgument(t *testing.T) {
	checkArgument(t, "default-argument", "user", false, false, "guest", "")
}
func TestConsoleDemoRequiredArrayArgument(t *testing.T) {
	checkArgument(t, "required-array-argument", "ids", true, true, "", "")
}
func TestConsoleDemoOptionalArrayArgument(t *testing.T) {
	checkArgument(t, "optional-array-argument", "tags", false, true, "", "")
}
func TestConsoleDemoArgumentDescription(t *testing.T) {
	checkArgument(t, "argument-description", "user", true, false, "", "The user ID")
}

func TestConsoleDemoArgumentValidation(t *testing.T) {
	checkErrors(t, "argument-validation", []string{"cannot follow optional", "cannot follow array", "duplicated argument"})
}

func TestConsoleDemoBooleanOption(t *testing.T) {
	checkOption(t, "boolean-option", "force", "", console.OptionValueNone, false, "", "")
}
func TestConsoleDemoValueOption(t *testing.T) {
	checkOption(t, "value-option", "queue", "", console.OptionValueOptional, false, "", "")
}
func TestConsoleDemoDefaultOption(t *testing.T) {
	checkOption(t, "default-option", "queue", "", console.OptionValueOptional, false, "default", "")
}
func TestConsoleDemoShortOption(t *testing.T) {
	checkOption(t, "short-option", "queue", "Q", console.OptionValueOptional, false, "default", "")
}
func TestConsoleDemoArrayOption(t *testing.T) {
	checkOption(t, "array-option", "id", "", console.OptionValueOptional, true, "", "")
}
func TestConsoleDemoOptionDescription(t *testing.T) {
	checkOption(t, "option-description", "force", "", console.OptionValueNone, false, "", "Force the operation")
}

func TestConsoleDemoOptionValidation(t *testing.T) {
	checkErrors(t, "option-validation", []string{"duplicated option", "duplicated option shortcut", "single character", "must accept values"})
}

func TestConsoleDemoMustDefinition(t *testing.T) {
	result := runCase(t, "must-definition")
	if result.Definition.Description != "Send a message" {
		t.Errorf("MustDefinition description = %q, want %q", result.Definition.Description, "Send a message")
	}
	if len(result.Values) != 1 || !strings.Contains(result.Values[0], "unclosed brace") {
		t.Errorf("MustDefinition failure = %#v, want unclosed brace panic", result.Values)
	}
}

func TestConsoleDemoParseSignature(t *testing.T) {
	result := runCase(t, "parse-signature")
	if result.Definition.Description != "Send a message" || len(result.Values) != 1 || !strings.Contains(result.Values[0], "unclosed brace") {
		t.Errorf("ParseSignature result = %#v, want description and parse error", result)
	}
}

func TestConsoleDemoArgument(t *testing.T) {
	result := runCase(t, "argument")
	if !reflect.DeepEqual(result.Values, []string{"Ada", "red"}) {
		t.Errorf("Argument values = %#v, want [Ada red]", result.Values)
	}
}

func TestConsoleDemoUnknownScenario(t *testing.T) {
	_, err := Run("missing")
	if err == nil || !strings.Contains(err.Error(), `unknown scenario "missing"`) {
		t.Errorf("Run(missing) error = %v, want unknown scenario", err)
	}
}

func TestConsoleDemoScenariosReturnsCopy(t *testing.T) {
	cases := Scenarios()
	if len(cases) != 20 {
		t.Fatalf("Scenarios length = %d, want 20", len(cases))
	}
	cases[0] = "changed"
	if Scenarios()[0] != "architecture" {
		t.Errorf("Scenarios()[0] = %q after caller mutation, want architecture", Scenarios()[0])
	}
}

func TestConsoleDemoValidationRejectsValidFixture(t *testing.T) {
	_, err := validationResult("argument-validation", []string{"demo:sample {user}"})
	if err == nil || !strings.Contains(err.Error(), `invalid signature "demo:sample {user}" succeeded`) {
		t.Errorf("validationResult(valid signature) error = %v, want fixture error", err)
	}
}

func TestConsoleDemoScenariosMatchCatalog(t *testing.T) {
	for _, caseName := range Scenarios() {
		entry, ok := catalog.Find("console", caseName)
		if !ok {
			t.Errorf("console scenario %q has no catalog entry; want an implemented entry", caseName)
			continue
		}
		if entry.Status != catalog.StatusImplemented || entry.Example != "demo:console "+caseName {
			t.Errorf("console catalog entry %q = %#v, want implemented demo:console example", caseName, entry)
		}
	}
}
