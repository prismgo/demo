package demo

import (
	"bytes"
	"context"
	"strings"
	"testing"

	translationdemo "prismgo-demo/app/demo/translation"
)

func TestTranslationDemoCommandListsImplementedScenarios(t *testing.T) {
	command := NewTranslationCommand()
	var output bytes.Buffer
	input := demoInput{arguments: map[string]string{"case": "list"}}

	if err := command.Handle(commandContext(command, input, &output)); err != nil {
		t.Fatalf("handle demo:translation list: %v", err)
	}
	for _, want := range []string{"architecture", "short-keys", "namespaces", "Using Facade Helpers", "implemented"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("translation list output does not contain %q:\n%s", want, output.String())
		}
	}
}

func TestTranslationDemoCommandRunsSelectedScenario(t *testing.T) {
	command := newTranslationCommand(func(_ context.Context, caseName string) (translationdemo.Result, error) {
		return translationdemo.Result{Case: caseName, Key: "messages.welcome", Value: "Welcome", Locale: "en"}, nil
	})
	var output bytes.Buffer
	input := demoInput{arguments: map[string]string{"case": "short-keys"}}

	if err := command.Handle(commandContext(command, input, &output)); err != nil {
		t.Fatalf("handle demo:translation short-keys: %v", err)
	}
	for _, want := range []string{"short-keys", "messages.welcome", "Welcome", "en"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("translation result output does not contain %q:\n%s", want, output.String())
		}
	}
}

func TestTranslationDemoCommandRejectsUnknownScenario(t *testing.T) {
	called := false
	command := newTranslationCommand(func(context.Context, string) (translationdemo.Result, error) {
		called = true
		return translationdemo.Result{}, nil
	})
	err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "missing"}}, &bytes.Buffer{}))
	if err == nil || !strings.Contains(err.Error(), `unknown translation scenario "missing"`) {
		t.Fatalf("unknown scenario error = %v, want descriptive error", err)
	}
	if called {
		t.Fatal("translation runner called for unknown scenario")
	}
}
