package demo

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	schemademo "prismgo-demo/app/demo/schema"
)

func TestSchemaDemoCommand(t *testing.T) {
	command := NewSchemaCommand()
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "list", want: "architecture"},
		{name: "architecture", want: "builder=Builder blueprint=Blueprint column=ColumnDefinition index=IndexDefinition foreign=ForeignKeyDefinition err=ErrUnsupportedFeature"},
		{name: "create-dialect-options", want: "mysql=ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 sqlite="},
		{name: "drop-all-tables", want: "remaining=0"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var output bytes.Buffer
			err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": testCase.name}}, &output))
			if err != nil {
				t.Fatalf("demo:schema %q error = %v, want nil", testCase.name, err)
			}
			if !strings.Contains(output.String(), testCase.want) {
				t.Fatalf("demo:schema %q output = %q, want substring %q", testCase.name, output.String(), testCase.want)
			}
		})
	}
}

func TestSchemaDemoCommandJSONAndUnknownScenario(t *testing.T) {
	command := NewSchemaCommand()
	var output bytes.Buffer
	err := command.Handle(commandContext(command, demoInput{
		arguments: map[string]string{"case": "architecture"},
		bools:     map[string]bool{"json": true},
	}, &output))
	if err != nil {
		t.Fatalf("demo:schema architecture --json error = %v, want nil", err)
	}
	var result schemademo.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode demo:schema JSON = %v, output = %q", err, output.String())
	}
	if result.Case != "architecture" {
		t.Fatalf("demo:schema architecture --json = %#v, want architecture result", result)
	}

	output.Reset()
	err = command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "not-a-case"}}, &output))
	if err == nil || !strings.Contains(err.Error(), `unknown or unimplemented schema scenario "not-a-case"`) {
		t.Fatalf("demo:schema unknown error = %v, want unknown scenario error", err)
	}
}
