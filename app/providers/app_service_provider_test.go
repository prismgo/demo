package providers

import (
	"testing"

	"github.com/prismgo/framework/foundation"
)

func TestAppServiceProviderLifecycle(t *testing.T) {
	app := foundation.NewApplication(t.TempDir())
	provider := AppServiceProvider{}

	if err := provider.Register(app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if err := provider.Boot(app); err != nil {
		t.Fatalf("Boot() error = %v", err)
	}
}
