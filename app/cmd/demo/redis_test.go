package demo

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	redisdemo "prismgo-demo/app/demo/redis"
)

func TestRedisDemoCommand(t *testing.T) {
	command := NewRedisCommand()
	for _, testCase := range []struct {
		name string
		want string
	}{
		{name: "list", want: "named-config"},
		{name: "config", want: "client=go default=default"},
		{name: "named-config", want: "cache-db=7 cache-name=orders-client"},
		{name: "timeouts", want: "duration=3s/500ms/2s"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var output bytes.Buffer
			err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": testCase.name}}, &output))
			if err != nil {
				t.Fatalf("demo:redis %q error = %v, want nil", testCase.name, err)
			}
			if !strings.Contains(output.String(), testCase.want) {
				t.Fatalf("demo:redis %q output = %q, want substring %q", testCase.name, output.String(), testCase.want)
			}
		})
	}
}

func TestRedisDemoCommandJSONAndUnknownScenario(t *testing.T) {
	command := NewRedisCommand()
	var output bytes.Buffer
	err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "tls"}, bools: map[string]bool{"json": true}}, &output))
	if err != nil {
		t.Fatalf("demo:redis tls --json error = %v, want nil", err)
	}
	var result redisdemo.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode demo:redis JSON = %v, output = %q", err, output.String())
	}
	if result.Case != "tls" || result.Value != "tls=true plain=false" {
		t.Fatalf("demo:redis tls --json = %#v, want tls scenario result", result)
	}
	output.Reset()
	err = command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "unknown"}}, &output))
	if err == nil || !strings.Contains(err.Error(), `unknown or unimplemented redis scenario "unknown"`) {
		t.Fatalf("demo:redis unknown error = %v, want unknown scenario error", err)
	}
}

// TestRedisDemoCommandIntegrationRequiresURL verifies integration scenarios
// surface the missing Redis URL instead of silently succeeding.
func TestRedisDemoCommandIntegrationRequiresURL(t *testing.T) {
	if value := strings.TrimSpace(os.Getenv("PRISMGO_REDIS_TEST_URL")); value != "" {
		t.Skipf("PRISMGO_REDIS_TEST_URL is configured, cannot verify missing-URL error here")
	}
	command := NewRedisCommand()
	err := command.Handle(commandContext(command, demoInput{arguments: map[string]string{"case": "keys"}}, io.Discard))
	if err == nil || !strings.Contains(err.Error(), "PRISMGO_REDIS_TEST_URL") {
		t.Fatalf("demo:redis keys without URL error = %v, want PRISMGO_REDIS_TEST_URL requirement", err)
	}
}
