package containerdemo_test

import (
	"strings"
	"testing"

	containerdemo "prismgo-demo/app/demo/container"
)

func TestContainerDemoScenarios(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "architecture", want: "client+podcast"},
		{name: "dependency-resolution", want: "parsed:feed"},
		{name: "use-cases", want: "managed client closed=true"},
		{name: "bind", want: "generator-1,generator-2"},
		{name: "transient-lifecycle", want: "created=2 closed=0"},
		{name: "singleton", want: "same=true created=1"},
		{name: "singleton-retry", want: "ready after 2 attempts"},
		{name: "instance", want: "same=true resolved=true"},
		{name: "nil-instance", want: "bound before=false after=true"},
		{name: "alias", want: "same=true bound=true resolved=true"},
		{name: "alias-close-order", want: "second,first"},
		{name: "with-closer", want: "closed=resource"},
		{name: "with-context-closer", want: "context=shutdown"},
		{name: "with-close-group", want: "normal=false reporting=true"},
		{name: "make", want: "made"},
		{name: "make-order", want: "instance=true loaded=1 factory=1"},
		{name: "factory", want: "before=0 values=fresh-1,fresh-2 shared=true"},
		{name: "typed-make", want: "typed"},
		{name: "typed-mismatch", want: "zero=true mismatch=true"},
		{name: "value", want: "before=true after=ready created=1"},
		{name: "value-zero", want: "missing=true mismatch=true"},
		{name: "call", want: "injected"},
		{name: "call-positional", want: "explicit-7"},
		{name: "call-results", want: "value=result error=true"},
		{name: "call-limits", want: "missing-typed-binding=true"},
		{name: "has", want: "before=false has=true loads=1"},
		{name: "bound", want: "missing=false ready=true loads=0"},
		{name: "resolved", want: "before=false after=true"},
		{name: "list", want: "keys=first,second registered=true,false closable=true created=0"},
		{name: "forget", want: "between=false value=new"},
		{name: "forget-no-close", want: "closed=false bound=false"},
		{name: "close-groups", want: "closed=normal reporting-bound=true"},
		{name: "close-group", want: "closed=reporting,normal"},
		{name: "close-order", want: "third,second,first"},
		{name: "closer-ownership", want: "closed=2"},
		{name: "close-pre-cancel", want: "cancelled=true retained=true"},
		{name: "close-success", want: "bound=false resolved=false"},
		{name: "close-retry", want: "error=true retained=true cleared=true attempts=2"},
		{name: "close-mid-cancel", want: "cancelled=true remaining=true"},
		{name: "missing-loader", want: "make=made has=true factory=factory loads=4"},
		{name: "missing-loader-error", want: "make=true factory=true has=false"},
		{name: "facade", want: "make=facade same=true listed=1 cleared=true"},
		{name: "provider-lifecycle", want: "before=active cleared=true"},
		{name: "error-not-registered", want: "not-registered=true"},
		{name: "error-nil-result", want: "nil-result=true"},
		{name: "error-no-current", want: "no-current=true"},
		{name: "laravel-mapping", want: "bind=true singleton=true instance=true alias=true contextual-binding=unsupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := containerdemo.Run(tt.name)
			if err != nil {
				t.Fatalf("container scenario %q error = %v, want nil", tt.name, err)
			}
			if got.Case != tt.name || got.Value != tt.want {
				t.Fatalf("container scenario %q result = %#v, want case %q and value %q", tt.name, got, tt.name, tt.want)
			}
		})
	}
}

func TestContainerDemoUnknownScenario(t *testing.T) {
	_, err := containerdemo.Run("unknown")
	if err == nil || !strings.Contains(err.Error(), `unknown scenario "unknown"`) {
		t.Fatalf("unknown container scenario error = %v, want descriptive unknown scenario error", err)
	}
}
