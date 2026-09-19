package facadedemo_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prismgo/framework/foundation"
	"github.com/prismgo/sqlite"

	// Load the demo's configuration defaults before booting the test application.
	_ "prismgo-demo/config"

	facadedemo "prismgo-demo/app/demo/facade"
)

func TestFacadeDemoIntroduction(t *testing.T) {
	newTestApp(t)
	assertScenario(t, "introduction", "config=Prismgo cache=memory logger=stack filesystem=local")
}

func TestFacadeDemoHowFacadesWork(t *testing.T) {
	newTestApp(t)
	assertScenario(t, "how-facades-work", "same-instance=true container-make=true")
}

func TestFacadeDemoCoreFacadeResolve(t *testing.T) {
	newTestApp(t)
	assertScenario(t, "core-facade-resolve", "*config.Config name=Prismgo debug=false")
}

func TestFacadeDemoAvailableFacades(t *testing.T) {
	newTestApp(t)
	assertScenario(t, "list", "total=9 resolved=")
}

func TestFacadeDemoCacheFacade(t *testing.T) {
	newTestApp(t)
	assertScenario(t, "cache-facade", "default=memory value=facade-cache-value has=true")
}

func TestFacadeDemoConfigFacade(t *testing.T) {
	newTestApp(t)
	assertScenario(t, "config-facade", "name=Prismgo debug=false port=8080")
}

func TestFacadeDemoRouteFacade(t *testing.T) {
	newTestApp(t)
	assertScenario(t, "route-facade", "registered=1")
}

func TestFacadeDemoLoggerFacade(t *testing.T) {
	newTestApp(t)
	assertScenario(t, "logger-facade", "has-default=true")
}

func TestFacadeDemoDatabaseFacade(t *testing.T) {
	newTestApp(t)
	assertScenario(t, "database-facade", "*gorm.DB name=")
}

func TestFacadeDemoFilesystemFacade(t *testing.T) {
	newTestApp(t)
	assertScenario(t, "filesystem-facade", "default=local exists=true content=facade-fs-value")
}

func TestFacadeDemoClassReference(t *testing.T) {
	newTestApp(t)
	assertScenario(t, "class-reference", "config=*config.Config cache=*cache.Manager logger=*logger.Manager")
}

func TestFacadeDemoLaravelMapping(t *testing.T) {
	newTestApp(t)
	assertScenario(t, "laravel-mapping", "count=10 mappings=Cache=cache")
}

func TestFacadeDemoUnknownScenario(t *testing.T) {
	_, err := facadedemo.Run("not-configured")
	if err == nil || !strings.Contains(err.Error(), `unknown scenario "not-configured"`) {
		t.Fatalf("unknown scenario error = %v, want descriptive error", err)
	}
}

func assertScenario(t *testing.T, name, want string) {
	t.Helper()
	result, err := facadedemo.Run(name)
	if err != nil {
		t.Fatalf("facade scenario %q error = %v, want nil", name, err)
	}
	if result.Case != name || !strings.Contains(result.Value, want) {
		t.Fatalf("facade scenario %q result = %#v, want case %q and value containing %q", name, result, name, want)
	}
}

func newTestApp(t *testing.T) {
	t.Helper()
	basePath := t.TempDir()
	for _, dir := range []string{
		filepath.Join(basePath, "storage", "database"),
		filepath.Join(basePath, "storage", "framework", "cache", "data"),
		filepath.Join(basePath, "storage", "framework", "cache", "locks"),
		filepath.Join(basePath, "storage", "app", "private"),
		filepath.Join(basePath, "storage", "app", "public"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create test path %s: %v", dir, err)
		}
	}
	t.Setenv("APP_ENV", "testing")
	t.Setenv("APP_DEBUG", "false")
	t.Setenv("APP_KEY", "base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=")
	t.Setenv("DB_CONNECTION", "sqlite")
	t.Setenv("DB_DATABASE", filepath.Join(basePath, "storage", "database", "test.sqlite"))
	t.Setenv("CACHE_STORE", "memory")
	t.Setenv("FILESYSTEM_DISK", "local")
	t.Setenv("FILESYSTEM_LOCAL_ROOT", filepath.Join(basePath, "storage", "app", "private"))
	t.Setenv("FILESYSTEM_PUBLIC_ROOT", filepath.Join(basePath, "storage", "app", "public"))
	app := foundation.Configure(basePath).WithExtensionProviders(sqlite.ServiceProvider{}).Create()
	if err := app.Boot(); err != nil {
		t.Fatalf("boot test application: %v", err)
	}
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Errorf("close test application: %v", err)
		}
	})
}
