package demo

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"prismgo-demo/app/demo/catalog"
	routedemo "prismgo-demo/app/demo/route"
)

func TestRouteDemoCommand(t *testing.T) {
	command := NewRouteCommand()
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "list", want: "isolated-router"},
		{name: "mount", want: "status=200 body=pong routes=1"},
		{name: "nested-groups", want: "url=/api/admin/users name=api.admin.users.index"},
		{name: "list-entries", want: "routes=2 GET /users=users.index;GET /users/{id}=users.show"},
		{name: "best-practices", want: "url=/api/v1/prismgos/7 routes=1 name=api.prismgos.show"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var output bytes.Buffer
			err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": testCase.name}}, &output))
			if err != nil {
				t.Fatalf("demo:route %q error = %v, want nil", testCase.name, err)
			}
			if !strings.Contains(output.String(), testCase.want) {
				t.Fatalf("demo:route %q output = %q, want substring %q", testCase.name, output.String(), testCase.want)
			}
		})
	}
}

func TestRouteDemoCommandJSONAndUnknownScenario(t *testing.T) {
	command := NewRouteCommand()
	var output bytes.Buffer
	err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "url"}, bools: map[string]bool{"json": true}}, &output))
	if err != nil {
		t.Fatalf("demo:route url --json error = %v, want nil", err)
	}
	var result routedemo.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode demo:route JSON = %v, output = %q", err, output.String())
	}
	if result.Case != "url" || result.Value != "url=/users/100" {
		t.Fatalf("demo:route url --json = %#v, want url scenario result", result)
	}
	if item, ok := catalog.Find("route", "unknown"); ok {
		t.Fatalf("unknown catalog entry = %#v, found=%t; want absent", item, ok)
	}
	output.Reset()
	err = command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "unknown"}}, &output))
	if err == nil || !strings.Contains(err.Error(), `unknown or unimplemented route scenario "unknown"`) {
		t.Fatalf("demo:route unknown error = %v, want unknown scenario error", err)
	}
}
