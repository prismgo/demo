// Package demotest provides isolated application and integration-test helpers.
package demotest

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"prismgo-demo/bootstrap"
	_ "prismgo-demo/config"

	"github.com/prismgo/framework/foundation"
)

// Options selects hermetic framework drivers. Empty fields use local defaults.
type Options struct {
	Database string
	Cache    string
	Queue    string
	Session  string
}

// NewApplication boots an application rooted in t.TempDir and registers a
// cleanup that closes it. It never reads the development checkout's .env.
func NewApplication(t testing.TB, options Options) *foundation.Application {
	t.Helper()
	options = withDefaults(options)
	if err := validateHermeticOptions(options); err != nil {
		t.Fatal(err)
	}

	basePath := t.TempDir()
	paths := map[string]string{
		"DB_DATABASE":            filepath.Join(basePath, "storage", "database.sqlite"),
		"CACHE_FILE_PATH":        filepath.Join(basePath, "storage", "framework", "cache", "data"),
		"CACHE_FILE_LOCK_PATH":   filepath.Join(basePath, "storage", "framework", "cache", "locks"),
		"SESSION_FILES":          filepath.Join(basePath, "storage", "framework", "sessions"),
		"FILESYSTEM_LOCAL_ROOT":  filepath.Join(basePath, "storage", "app", "private"),
		"FILESYSTEM_PUBLIC_ROOT": filepath.Join(basePath, "storage", "app", "public"),
	}
	for _, path := range paths {
		if err := ensureParent(path); err != nil {
			t.Fatalf("prepare hermetic application path: %v", err)
		}
	}

	environment := map[string]string{
		"APP_ENV": "testing", "APP_DEBUG": "false", "APP_KEY": "base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=",
		"DB_CONNECTION": options.Database, "CACHE_STORE": options.Cache,
		"QUEUE_CONNECTION": options.Queue, "SESSION_DRIVER": options.Session,
		"REDIS_URL": "", "REDIS_CACHE_URL": "", "RABBITMQ_URL": "",
		"FILESYSTEM_DISK": "local",
	}
	for key, value := range paths {
		environment[key] = value
	}
	for key, value := range environment {
		t.Setenv(key, value)
	}

	app := bootstrap.NewApplication(basePath)
	if err := app.Boot(); err != nil {
		t.Fatalf("boot hermetic application: %v", err)
	}
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Errorf("close hermetic application: %v", err)
		}
	})
	return app
}

func withDefaults(options Options) Options {
	if options.Database == "" {
		options.Database = "sqlite"
	}
	if options.Cache == "" {
		options.Cache = "memory"
	}
	if options.Queue == "" {
		options.Queue = "sync"
	}
	if options.Session == "" {
		options.Session = "file"
	}
	return options
}

func validateHermeticOptions(options Options) error {
	allowed := map[string]map[string]bool{
		"database": {"sqlite": true},
		"cache":    {"memory": true, "file": true},
		"queue":    {"sync": true},
		"session":  {"file": true},
	}
	values := map[string]string{
		"database": options.Database,
		"cache":    options.Cache,
		"queue":    options.Queue,
		"session":  options.Session,
	}
	for component, value := range values {
		if !allowed[component][value] {
			return fmt.Errorf("demotest: %s driver %q is not hermetic", component, value)
		}
	}
	return nil
}

func ensureParent(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0o755)
}
