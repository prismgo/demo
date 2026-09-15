// Package lifecycledemo contains runnable examples of the application lifecycle.
package lifecycledemo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/prismgo/framework/container"
	"github.com/prismgo/framework/foundation"
)

// Result records one observable application lifecycle scenario.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

// Run executes a lifecycle catalog scenario that needs no external services.
//
// Each run prepares a throwaway application base path, a hermetic environment and
// isolated foundation/container globals so scenarios can build and close their own
// Application without disturbing the running demo process.
func Run(name string) (Result, error) {
	base, err := os.MkdirTemp("", "prismgo-lifecycle-")
	if err != nil {
		return Result{}, fmt.Errorf("lifecycle demo %s: create base directory: %w", name, err)
	}
	defer os.RemoveAll(base)

	if err := prepareBasePaths(base); err != nil {
		return Result{}, fmt.Errorf("lifecycle demo %s: %w", name, err)
	}
	restoreEnv := setHermeticEnv(base)
	defer restoreEnv()

	restoreGlobals := isolateGlobals()
	defer restoreGlobals()

	restoreGin := quietGin()
	defer restoreGin()

	value, err := run(base, name)
	if err != nil {
		return Result{}, fmt.Errorf("lifecycle demo %s: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

func run(base, name string) (string, error) {
	switch name {
	case "application-entry":
		return applicationEntryScenario()
	case "run-context":
		return runContextScenario(base)
	case "base-providers":
		return baseProvidersScenario(base)
	case "default-providers":
		return defaultProvidersScenario(base)
	case "extension-providers":
		return extensionProvidersScenario(base)
	case "application-providers":
		return applicationProvidersScenario(base)
	case "exception-handler":
		return exceptionHandlerScenario(base)
	case "provider-layering":
		return providerLayeringScenario(base)
	case "register-phase":
		return registerPhaseScenario(base)
	case "boot-phase":
		return bootPhaseScenario(base)
	case "booted-runner-order":
		return bootedRunnerOrderScenario(base)
	case "deferred-provider-map":
		return deferredProviderMapScenario()
	case "deferred-provider-resolution":
		return deferredProviderResolutionScenario(base)
	case "dynamic-provider":
		return dynamicProviderScenario(base)
	case "http-pipeline":
		return httpPipelineScenario(base)
	case "request-id":
		return requestIDScenario(base)
	case "access-log":
		return accessLogScenario(base)
	case "exception-middleware":
		return exceptionMiddlewareScenario(base)
	case "business-middleware":
		return businessMiddlewareScenario(base)
	case "request-error-log":
		return requestErrorLogScenario(base)
	case "handle-command":
		return handleCommandScenario(base)
	case "console-starting":
		return consoleStartingScenario(base)
	case "command-resolution":
		return commandResolutionScenario()
	case "command-execution":
		return commandExecutionScenario(base)
	case "shared-application":
		return sharedApplicationScenario(base)
	case "runner-shutdown":
		return runnerShutdownScenario(base)
	case "signal-shutdown":
		return signalShutdownScenario(base)
	case "root-context-cancel":
		return rootContextCancelScenario(base)
	case "terminating-event-order":
		return terminatingEventOrderScenario(base)
	case "terminable-providers":
		return terminableProvidersScenario(base)
	case "cleanup-functions":
		return cleanupFunctionsScenario(base)
	case "resource-close-order":
		return resourceCloseOrderScenario(base)
	case "shutdown-error-reporting":
		return shutdownErrorReportingScenario(base)
	case "terminated-event-order":
		return terminatedEventOrderScenario(base)
	case "close-retry":
		return closeRetryScenario(base)
	case "best-effort-events":
		return bestEffortEventsScenario(base)
	case "app-booting-event":
		return appBootingEventScenario(base)
	case "app-booted-event":
		return appBootedEventScenario(base)
	case "app-terminating-event":
		return appTerminatingEventScenario(base)
	case "app-terminated-event":
		return appTerminatedEventScenario(base)
	case "provider-registering-event":
		return providerRegisteringEventScenario(base)
	case "provider-registered-event":
		return providerRegisteredEventScenario(base)
	case "provider-booting-event":
		return providerBootingEventScenario(base)
	case "provider-booted-event":
		return providerBootedEventScenario(base)
	case "server-starting-event":
		return serverStartingEventScenario(base)
	case "server-started-event":
		return serverStartedEventScenario(base)
	case "server-stopping-event":
		return serverStoppingEventScenario(base)
	case "server-stopped-event":
		return serverStoppedEventScenario(base)
	case "request-received-event":
		return requestReceivedEventScenario(base)
	case "request-handled-event":
		return requestHandledEventScenario(base)
	case "request-failed-event":
		return requestFailedEventScenario(base)
	case "request-finished-event":
		return requestFinishedEventScenario(base)
	case "console-application-starting-event":
		return consoleApplicationStartingEventScenario(base)
	case "console-command-starting-event":
		return consoleCommandStartingEventScenario(base)
	case "console-command-finished-event":
		return consoleCommandFinishedEventScenario(base)
	case "listeners":
		return listenersScenario(base)
	case "payload-boundaries":
		return payloadBoundariesScenario()
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
	// "file not found" warning that a bare temporary base path would produce.
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

// quietGin forces Gin test mode with discarded output for the scenario duration.
func quietGin() func() {
	mode := gin.Mode()
	writer := gin.DefaultWriter
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	return func() {
		gin.DefaultWriter = writer
		gin.SetMode(mode)
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
