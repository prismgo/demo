package consoledemo

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"
)

func isIOScenario(caseName string) bool {
	switch caseName {
	case "line", "new-line", "info-comment-question-success", "warn-error", "alert", "table", "progress", "ask", "secret", "confirm", "choice", "choice-multiple", "choice-defaults-attempts", "anticipate":
		return true
	}
	return false
}

func runIOScenario(caseName string) (Result, error) {
	result := Result{Case: caseName}
	var input string
	switch caseName {
	case "ask":
		input = "Ada\n\n"
	case "secret":
		input = "secret-value\n"
	case "confirm":
		input = "yes\nno\n\n"
	case "choice":
		input = "2\nred\n\n"
	case "choice-multiple":
		input = "1, blue\n"
	case "choice-defaults-attempts":
		input = "\ninvalid\n2\ninvalid\ninvalid\n"
	case "anticipate":
		input = "Taylor\n\n"
	}
	var out, errOut bytes.Buffer
	styled := caseName == "line" || caseName == "info-comment-question-success" || caseName == "warn-error" || caseName == "alert"
	ioo := console.NewIOWithOutputOptions(strings.NewReader(input), &out, &errOut, console.OutputOptions{ANSI: styled})
	switch caseName {
	case "line":
		ioo.Line("plain")
		ioo.Line("colored", "info")
	case "new-line":
		ioo.Line("before")
		ioo.NewLine()
		ioo.NewLine(2)
		ioo.Line("after")
	case "info-comment-question-success":
		ioo.Info("info")
		ioo.Comment("comment")
		ioo.Question("question")
		ioo.Success("success")
	case "warn-error":
		ioo.Warn("warning")
		ioo.Error("failure")
	case "alert":
		ioo.Alert("Attention")
	case "table":
		if err := ioo.Table([]string{"Name", "Role"}, [][]string{{"Ada", "Admin"}, {"Bob", "User"}}); err != nil {
			return Result{}, fmt.Errorf("console demo table: %w", err)
		}
		failing := console.NewIO(strings.NewReader(""), &errorWriter{}, &errOut)
		if err := failing.Table([]string{"Name"}, [][]string{{"Ada"}}); err == nil {
			return Result{}, fmt.Errorf("console demo table: failing writer succeeded")
		} else {
			result.Values = []string{err.Error()}
		}
	case "progress":
		bar := ioo.Progress(3)
		bar.Advance(0)
		bar.Advance(2)
		bar.Finish()
	case "ask":
		name, err := ioo.Ask("Name")
		if err != nil {
			return Result{}, fmt.Errorf("console demo ask name: %w", err)
		}
		fallback, err := ioo.Ask("Name", "Guest")
		if err != nil {
			return Result{}, fmt.Errorf("console demo ask default: %w", err)
		}
		result.Values = []string{name, fallback}
	case "secret":
		secret, err := ioo.Secret("Password")
		if err != nil {
			return Result{}, fmt.Errorf("console demo secret: %w", err)
		}
		result.Values = []string{secret}
	case "confirm":
		for _, defaultYes := range []bool{false, false, true} {
			confirmed, err := ioo.Confirm("Continue", defaultYes)
			if err != nil {
				return Result{}, fmt.Errorf("console demo confirm: %w", err)
			}
			result.Values = append(result.Values, fmt.Sprintf("%t", confirmed))
		}
	case "choice":
		options := []string{"red", "blue"}
		for _, fallback := range []string{"", "", "red"} {
			choice, err := ioo.Choice("Color", options, fallback)
			if err != nil {
				return Result{}, fmt.Errorf("console demo choice: %w", err)
			}
			result.Values = append(result.Values, choice)
		}
	case "choice-multiple":
		choices, err := ioo.ChoiceWithOptions("Colors", []string{"red", "blue"}, console.ChoiceOptions{Multiple: true})
		if err != nil {
			return Result{}, fmt.Errorf("console demo choice-multiple: %w", err)
		}
		result.Values = choices
	case "choice-defaults-attempts":
		options := []string{"red", "blue"}
		defaults, err := ioo.ChoiceWithOptions("Colors", options, console.ChoiceOptions{Multiple: true, Defaults: []string{"red", "blue"}})
		if err != nil {
			return Result{}, fmt.Errorf("console demo choice defaults: %w", err)
		}
		retried, err := ioo.ChoiceWithOptions("Color", options, console.ChoiceOptions{Attempts: 2})
		if err != nil {
			return Result{}, fmt.Errorf("console demo choice retry: %w", err)
		}
		_, err = ioo.ChoiceWithOptions("Color", options, console.ChoiceOptions{Attempts: 2})
		if err == nil {
			return Result{}, fmt.Errorf("console demo choice retry limit: invalid choices succeeded")
		}
		result.Values = append(defaults, append(retried, err.Error())...)
	case "anticipate":
		name, err := ioo.Anticipate("Name", []string{"Taylor", "Dayle"}, "Dayle")
		if err != nil {
			return Result{}, fmt.Errorf("console demo anticipate candidate: %w", err)
		}
		fallback, err := ioo.Anticipate("Name", []string{"Taylor", "Dayle"}, "Dayle")
		if err != nil {
			return Result{}, fmt.Errorf("console demo anticipate fallback: %w", err)
		}
		result.Values = []string{name, fallback}
	}
	result.Output, result.ErrorOutput = out.String(), errOut.String()
	return result, nil
}

type errorWriter struct{}

func (w *errorWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("forced table write failure")
}
