package demo

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"prismgo-demo/app/demo/catalog"
	cookiedemo "prismgo-demo/app/demo/cookie"
)

func TestCookieDemoCommand(t *testing.T) {
	command := NewCookieCommand()
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "list", want: "scope-deduplication"},
		{name: "middleware", want: "status=201 cookies=theme=dark@/"},
		{name: "queue-forget", want: "name=theme path=/admin domain=example.test maxAge=-1"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var output bytes.Buffer
			err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": testCase.name}}, &output))
			if err != nil {
				t.Fatalf("demo:cookie %q error = %v, want nil", testCase.name, err)
			}
			if !strings.Contains(output.String(), testCase.want) {
				t.Fatalf("demo:cookie %q output = %q, want substring %q", testCase.name, output.String(), testCase.want)
			}
		})
	}
}

func TestCookieDemoCommandJSONAndUnknownScenario(t *testing.T) {
	command := NewCookieCommand()
	var output bytes.Buffer
	err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "make"}, bools: map[string]bool{"json": true}}, &output))
	if err != nil {
		t.Fatalf("demo:cookie make --json error = %v, want nil", err)
	}
	var result cookiedemo.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode demo:cookie JSON = %v, output = %q", err, output.String())
	}
	if result.Case != "make" || result.Value != "equal=true name=theme value=dark" {
		t.Fatalf("demo:cookie make --json = %#v, want make scenario result", result)
	}
	if item, ok := catalog.Find("cookie", "unknown"); ok {
		t.Fatalf("unknown catalog entry = %#v, found=%t; want absent", item, ok)
	}
	output.Reset()
	err = command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "unknown"}}, &output))
	if err == nil || !strings.Contains(err.Error(), `unknown or unimplemented cookie scenario "unknown"`) {
		t.Fatalf("demo:cookie unknown error = %v, want unknown scenario error", err)
	}
}
