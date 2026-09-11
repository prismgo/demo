package bootstrap

import (
	"testing"

	pcontract "github.com/prismgo/framework/contracts/provider"
)

func TestExtensionsIncludeRabbitMQProvider(t *testing.T) {
	exts := Extensions()
	if got, want := len(exts), 2; got != want {
		t.Fatalf("len(Extensions()) = %d, want %d", got, want)
	}
	named, ok := exts[1].(pcontract.NamedProvider)
	if !ok {
		t.Fatalf("Extensions()[1] type = %T, want provider.NamedProvider", exts[1])
	}
	if got, want := named.Name(), "prismgo.extension.rabbitmq"; got != want {
		t.Fatalf("Extensions()[1].Name() = %q, want %q", got, want)
	}
}
