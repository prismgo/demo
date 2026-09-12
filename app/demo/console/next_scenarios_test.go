package consoledemo

import (
	"reflect"
	"strings"
	"testing"
)

func checkNextScenario(t *testing.T, caseName string) {
	t.Helper()
	tests := []struct {
		name        string
		values      []string
		output      []string
		errorOutput []string
	}{
		{name: "arguments", values: []string{"red", "blue"}},
		{name: "missing-argument", values: []string{"", "tags_nil=true"}},
		{name: "option", values: []string{"priority", ""}},
		{name: "option-strings", values: []string{"alpha", "beta"}},
		{name: "option-bool", values: []string{"force=true", "enabled=true", "absent=false"}},
		{name: "option-int", values: []string{"retries=42", "invalid syntax"}},
		{name: "has-option", values: []string{"force=true", "absent=false", "global=true"}},
		{name: "optional-option-value", values: []string{"", "strings=[\"\"]", "present=true"}},
		{name: "line", output: []string{"plain\n", "\x1b[", "colored"}},
		{name: "new-line", output: []string{"before\n\n\n\nafter\n"}},
		{name: "info-comment-question-success", output: []string{"\x1b[", "info", "comment", "question", "success"}},
		{name: "warn-error", errorOutput: []string{"\x1b[", "warning", "failure"}},
		{name: "alert", output: []string{"\x1b[", "********", "*     Attention     *"}},
		{name: "table", values: []string{"forced table write failure"}, output: []string{"Name", "Role", "Ada", "Admin", "Bob", "User"}},
		{name: "progress", output: []string{"\r[1/3]\r[3/3]\n"}},
		{name: "ask", values: []string{"Ada", "Guest"}, output: []string{"Name: ", "[default: Guest]"}},
		{name: "secret", values: []string{"secret-value"}, output: []string{"Password: "}},
		{name: "confirm", values: []string{"true", "false", "true"}, output: []string{"[y/N]", "[Y/n]"}},
		{name: "choice", values: []string{"blue", "red", "red"}, output: []string{"1) red", "2) blue", "[default: red]"}},
		{name: "choice-multiple", values: []string{"red", "blue"}, output: []string{"1) red", "2) blue"}},
	}
	for _, tt := range tests {
		if tt.name == caseName {
			result := runCase(t, tt.name)
			if tt.values != nil {
				if len(result.Values) != len(tt.values) {
					t.Fatalf("%s values = %#v, want %#v", tt.name, result.Values, tt.values)
				}
				for i, want := range tt.values {
					if tt.name == "option-int" && i == 1 {
						if !strings.Contains(result.Values[i], want) {
							t.Errorf("%s values[%d] = %q, want substring %q", tt.name, i, result.Values[i], want)
						}
						continue
					}
					if result.Values[i] != want {
						t.Errorf("%s values[%d] = %q, want %q", tt.name, i, result.Values[i], want)
					}
				}
			} else if len(result.Values) != 0 {
				t.Errorf("%s values = %#v, want empty", tt.name, result.Values)
			}
			for _, want := range tt.output {
				if !strings.Contains(result.Output, want) {
					t.Errorf("%s output = %q, want substring %q", tt.name, result.Output, want)
				}
			}
			for _, want := range tt.errorOutput {
				if !strings.Contains(result.ErrorOutput, want) {
					t.Errorf("%s error output = %q, want substring %q", tt.name, result.ErrorOutput, want)
				}
			}
			if tt.name == "table" {
				const want = "Name  Role\nAda   Admin\nBob   User\n"
				if result.Output != want {
					t.Errorf("table output = %q, want %q", result.Output, want)
				}
			}
			if tt.name == "warn-error" && result.Output != "" {
				t.Errorf("warn-error stdout = %q, want empty", result.Output)
			}
			if tt.name == "secret" && strings.Contains(result.Output, result.Values[0]) {
				t.Errorf("secret prompt output = %q, want no echoed secret", result.Output)
			}
			return
		}
	}
	t.Fatalf("console test case %q has no fixture, want one", caseName)
}

func TestConsoleDemoNextScenarioOrder(t *testing.T) {
	if got, want := Scenarios()[20:40], []string{"arguments", "missing-argument", "option", "option-strings", "option-bool", "option-int", "has-option", "optional-option-value", "line", "new-line", "info-comment-question-success", "warn-error", "alert", "table", "progress", "ask", "secret", "confirm", "choice", "choice-multiple"}; !reflect.DeepEqual(got, want) {
		t.Errorf("next console scenarios = %#v, want %#v", got, want)
	}
}

func TestConsoleDemoArguments(t *testing.T)           { checkNextScenario(t, "arguments") }
func TestConsoleDemoMissingArgument(t *testing.T)     { checkNextScenario(t, "missing-argument") }
func TestConsoleDemoOption(t *testing.T)              { checkNextScenario(t, "option") }
func TestConsoleDemoOptionStrings(t *testing.T)       { checkNextScenario(t, "option-strings") }
func TestConsoleDemoOptionBool(t *testing.T)          { checkNextScenario(t, "option-bool") }
func TestConsoleDemoOptionInt(t *testing.T)           { checkNextScenario(t, "option-int") }
func TestConsoleDemoHasOption(t *testing.T)           { checkNextScenario(t, "has-option") }
func TestConsoleDemoOptionalOptionValue(t *testing.T) { checkNextScenario(t, "optional-option-value") }
func TestConsoleDemoLine(t *testing.T)                { checkNextScenario(t, "line") }
func TestConsoleDemoNewLine(t *testing.T)             { checkNextScenario(t, "new-line") }
func TestConsoleDemoInfoCommentQuestionSuccess(t *testing.T) {
	checkNextScenario(t, "info-comment-question-success")
}
func TestConsoleDemoWarnError(t *testing.T)      { checkNextScenario(t, "warn-error") }
func TestConsoleDemoAlert(t *testing.T)          { checkNextScenario(t, "alert") }
func TestConsoleDemoTable(t *testing.T)          { checkNextScenario(t, "table") }
func TestConsoleDemoProgress(t *testing.T)       { checkNextScenario(t, "progress") }
func TestConsoleDemoAsk(t *testing.T)            { checkNextScenario(t, "ask") }
func TestConsoleDemoSecret(t *testing.T)         { checkNextScenario(t, "secret") }
func TestConsoleDemoConfirm(t *testing.T)        { checkNextScenario(t, "confirm") }
func TestConsoleDemoChoice(t *testing.T)         { checkNextScenario(t, "choice") }
func TestConsoleDemoChoiceMultiple(t *testing.T) { checkNextScenario(t, "choice-multiple") }

func TestConsoleDemoInputFixtureErrors(t *testing.T) {
	for _, tt := range []struct {
		name      string
		signature string
		flags     []string
		want      string
	}{
		{name: "invalid signature", signature: "demo:sample {user", want: "parse console signature"},
		{name: "unknown flag", signature: "demo:sample {--force}", flags: []string{"--missing"}, want: "parse console flags"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := scenarioInput(tt.signature, nil, tt.flags)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("scenarioInput(%q, %q) error = %v, want substring %q", tt.signature, tt.flags, err, tt.want)
			}
		})
	}
}
