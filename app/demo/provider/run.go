// Package providerdemo contains runnable examples of service provider registration and lifecycle.
package providerdemo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/prismgo/framework/container"
	"github.com/prismgo/framework/foundation"
)

// Result records one observable service provider scenario.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

// Run executes a service provider catalog scenario.
//
// Each run prepares a throwaway base path, a hermetic environment and isolated
// foundation/container globals so scenarios can build and close their own Application
// without disturbing the running demo process.
func Run(name string) (Result, error) {
	base, err := os.MkdirTemp("", "prismgo-provider-")
	if err != nil {
		return Result{}, fmt.Errorf("provider demo %s: create base directory: %w", name, err)
	}
	defer os.RemoveAll(base)

	if err := prepareBasePaths(base); err != nil {
		return Result{}, fmt.Errorf("provider demo %s: %w", name, err)
	}
	restoreEnv := setHermeticEnv(base)
	defer restoreEnv()

	restoreGlobals := isolateGlobals()
	defer restoreGlobals()

	value, err := run(base, name)
	if err != nil {
		return Result{}, fmt.Errorf("provider demo %s: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

// run dispatches one named scenario.
func run(base, name string) (string, error) {
	switch name {
	case "architecture":
		return architectureScenario()
	case "contract":
		return contractScenario()
	case "register-contract":
		return registerContractScenario()
	case "container-access":
		return containerAccessScenario()
	case "preserve-binding":
		return preserveBindingScenario()
	case "bind":
		return bindScenario()
	case "singleton":
		return singletonScenario()
	case "instance":
		return instanceScenario()
	case "alias":
		return aliasScenario()
	case "with-closer":
		return withCloserScenario()
	case "close-group":
		return closeGroupScenario()
	case "boot-order":
		return bootOrderScenario(base)
	case "boot-listeners":
		return bootListenersScenario(base)
	case "commands":
		return commandsScenario(base)
	case "command-inputs":
		return commandInputsScenario(base)
	case "command-deferred":
		return commandDeferredScenario(base)
	case "publishes":
		return publishesScenario(base)
	case "publish-tags":
		return publishTagsScenario(base)
	case "publish-environment":
		return publishEnvironmentScenario(base)
	case "application-registration":
		return applicationRegistrationScenario()
	case "base-order":
		return baseOrderScenario(base)
	case "default-order":
		return defaultOrderScenario(base)
	case "extension-order":
		return extensionOrderScenario(base)
	case "application-order":
		return applicationOrderScenario(base)
	case "named-identity":
		return namedIdentityScenario(base)
	case "implicit-identity":
		return implicitIdentityScenario(base)
	case "duplicate-identity":
		return duplicateIdentityScenario(base)
	case "deferrable-contract":
		return deferrableContractScenario()
	case "provides":
		return providesScenario()
	case "deferred-resolution":
		return deferredResolutionScenario(base)
	case "deferred-map-cleanup":
		return deferredMapCleanupScenario(base)
	case "deferred-late-boot":
		return deferredLateBootScenario(base)
	case "deferred-empty":
		return deferredEmptyScenario(base)
	case "deferred-conflict":
		return deferredConflictScenario(base)
	case "deferred-termination":
		return deferredTerminationScenario(base)
	case "terminable-contract":
		return terminableContractScenario()
	case "worker-lifecycle":
		return workerLifecycleScenario(base)
	case "terminate-order":
		return terminateOrderScenario(base)
	case "terminate-eligibility":
		return terminateEligibilityScenario(base)
	case "terminate-context":
		return terminateContextScenario(base)
	case "closer-order":
		return closerOrderScenario(base)
	case "full-lifecycle":
		return fullLifecycleScenario(base)
	default:
		return "", fmt.Errorf("unknown scenario %q", name)
	}
}

// prepareBasePaths creates the storage directories a hermetic application expects.
func prepareBasePaths(base string) error {
	paths := []string{
		filepath.Join(base, "storage", "database.sqlite"),
		filepath.Join(base, "storage", "framework", "cache", "data"),
		filepath.Join(base, "storage", "framework", "sessions"),
		filepath.Join(base, "storage", "app", "private"),
		filepath.Join(base, "storage", "app", "public"),
	}
	for _, path := range paths {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("prepare base path %s: %w", path, err)
		}
	}
	// An empty .env keeps config loading on the process environment without the
	// "file not found" warning a bare temporary base path would produce.
	if err := os.WriteFile(filepath.Join(base, ".env"), nil, 0o600); err != nil {
		return fmt.Errorf("prepare base environment: %w", err)
	}
	return nil
}

// isolateGlobals restores the process-wide Application and container provider after
// a scenario builds and closes its own throwaway Application.
func isolateGlobals() func() {
	previous := foundation.App
	return func() {
		foundation.App = previous
		if previous != nil {
			if registry, ok := previous.Container().(*container.Container); ok {
				container.SetProvider(func() *container.Container { return registry })
				return
			}
		}
		container.SetProvider(nil)
	}
}

// setHermeticEnv applies the isolated environment used by every scenario and returns
// a restore closure. It never reads the development checkout .env.
func setHermeticEnv(base string) func() {
	values := map[string]string{
		"APP_ENV":                "testing",
		"APP_DEBUG":              "false",
		"APP_KEY":                "base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=",
		"DB_CONNECTION":          "sqlite",
		"DB_DATABASE":            filepath.Join(base, "storage", "database.sqlite"),
		"CACHE_STORE":            "memory",
		"QUEUE_CONNECTION":       "sync",
		"SESSION_DRIVER":         "file",
		"SESSION_FILES":          filepath.Join(base, "storage", "framework", "sessions"),
		"FILESYSTEM_DISK":        "local",
		"FILESYSTEM_LOCAL_ROOT":  filepath.Join(base, "storage", "app", "private"),
		"FILESYSTEM_PUBLIC_ROOT": filepath.Join(base, "storage", "app", "public"),
		"SERVER_ACCESS_LOG":      "false",
		"REDIS_URL":              "",
		"REDIS_CACHE_URL":        "",
		"RABBITMQ_URL":           "",
	}
	return setEnv(values)
}

// setEnv snapshots the listed variables, applies values and returns a restore closure.
func setEnv(values map[string]string) func() {
	previous := make(map[string]*string, len(values))
	for key, value := range values {
		if old, ok := os.LookupEnv(key); ok {
			copied := old
			previous[key] = &copied
		} else {
			previous[key] = nil
		}
		_ = os.Setenv(key, value)
	}
	return func() {
		for key, old := range previous {
			if old == nil {
				_ = os.Unsetenv(key)
				continue
			}
			_ = os.Setenv(key, *old)
		}
	}
}
