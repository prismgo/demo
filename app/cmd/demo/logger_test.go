package demo

import (
	"bytes"
	"strings"
	"testing"

	loggerdemo "prismgo-demo/app/demo/logger"
)

func TestLoggerDemoCommandListsAndRunsCases(t *testing.T) {
	command := NewLoggerCommand()
	var output bytes.Buffer
	if err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "list"}}, &output)); err != nil {
		t.Fatalf("logger list error = %v, want nil", err)
	}
	for _, want := range []string{"single-driver", "global-logrus", "Documentation"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("logger list = %q, want %q", output.String(), want)
		}
	}
	command.run = func(name string) (loggerdemo.Result, error) {
		return loggerdemo.Result{Case: name, Value: "observed"}, nil
	}
	output.Reset()
	if err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "single-driver"}, bools: map[string]bool{"json": true}}, &output)); err != nil {
		t.Fatalf("logger case error = %v, want nil", err)
	}
	if !strings.Contains(output.String(), `"value":"observed"`) {
		t.Fatalf("logger JSON = %q, want observed result", output.String())
	}
	err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "missing"}}, &bytes.Buffer{}))
	if err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("missing logger case error = %v, want named rejection", err)
	}
}
