package lifecycledemo_test

import (
	"strings"
	"testing"

	lifecycledemo "prismgo-demo/app/demo/lifecycle"
)

// runCase executes one scenario and fails with locating context when it errors.
func runCase(t *testing.T, name string) string {
	t.Helper()
	result, err := lifecycledemo.Run(name)
	if err != nil {
		t.Fatalf("lifecycle demo %q error = %v, want nil", name, err)
	}
	if result.Case != name {
		t.Fatalf("lifecycle demo %q result case = %q, want %q", name, result.Case, name)
	}
	return result.Value
}

// expectValue asserts the full observable value for a deterministic scenario.
func expectValue(t *testing.T, name, want string) {
	t.Helper()
	if got := runCase(t, name); got != want {
		t.Fatalf("lifecycle demo %q value = %q, want %q", name, got, want)
	}
}

func TestLifecycleDemoApplicationEntry(t *testing.T) {
	expectValue(t, "application-entry", "entry=NewApplication builder=Configure command=HandleCommand runner=RunContext server=NewHTTPServer")
}

func TestLifecycleDemoRunContext(t *testing.T) {
	expectValue(t, "run-context", "runner-ran=true context-canceled=true")
}

func TestLifecycleDemoBaseProviders(t *testing.T) {
	expectValue(t, "base-providers", "event=true config=true logger=true translation=true")
}

func TestLifecycleDemoDefaultProviders(t *testing.T) {
	const want = "order=redis,cache,queue,cookie,session,filesystem,database,database.schema,ratelimit,route" +
		" bound-before=false bound-after=true"
	expectValue(t, "default-providers", want)
}

func TestLifecycleDemoExtensionProviders(t *testing.T) {
	expectValue(t, "extension-providers", "position=after-defaults lazy-after-boot=true lazy-resolved=true")
}

func TestLifecycleDemoApplicationProviders(t *testing.T) {
	expectValue(t, "application-providers", "order=demo.extension,demo.application override=application")
}

func TestLifecycleDemoExceptionHandler(t *testing.T) {
	expectValue(t, "exception-handler", "bound=true handler=*exception.Handler")
}

func TestLifecycleDemoProviderLayering(t *testing.T) {
	expectValue(t, "provider-layering", "framework=10 extension=1 application=1 layered=true")
}

func TestLifecycleDemoRegisterPhase(t *testing.T) {
	expectValue(t, "register-phase", "registers-before-boots=true binding-only=true")
}

func TestLifecycleDemoBootPhase(t *testing.T) {
	expectValue(t, "boot-phase", "order=redis,cache,queue,cookie,session,filesystem,database,database.schema,ratelimit,route")
}

func TestLifecycleDemoBootedRunnerOrder(t *testing.T) {
	expectValue(t, "booted-runner-order", "order=booted,runner")
}

func TestLifecycleDemoDeferredProviderMap(t *testing.T) {
	expectValue(t, "deferred-provider-map", "interface=DeferrableProvider keys=demo.deferred")
}

func TestLifecycleDemoDeferredProviderResolution(t *testing.T) {
	expectValue(t, "deferred-provider-resolution", "bound-before=false register=1 boot=1 value=demo.deferred")
}

func TestLifecycleDemoDynamicProvider(t *testing.T) {
	expectValue(t, "dynamic-provider", "registered=true booted=true")
}

func TestLifecycleDemoHTTPPipeline(t *testing.T) {
	expectValue(t, "http-pipeline", "order=received,middleware,handler,handled status=200")
}

func TestLifecycleDemoRequestID(t *testing.T) {
	expectValue(t, "request-id", "header-present=true body-echoes=true")
}

func TestLifecycleDemoAccessLog(t *testing.T) {
	expectValue(t, "access-log", "logged=true")
}

func TestLifecycleDemoExceptionMiddleware(t *testing.T) {
	expectValue(t, "exception-middleware", "status=500 rendered=true")
}

func TestLifecycleDemoBusinessMiddleware(t *testing.T) {
	expectValue(t, "business-middleware", "business=ran")
}

func TestLifecycleDemoRequestErrorLog(t *testing.T) {
	expectValue(t, "request-error-log", "failed-event=true status=500")
}

func TestLifecycleDemoHandleCommand(t *testing.T) {
	expectValue(t, "handle-command", "ran=true closed=true")
}

func TestLifecycleDemoConsoleStarting(t *testing.T) {
	expectValue(t, "console-starting", "mounted=true starting-event=true")
}

func TestLifecycleDemoCommandResolution(t *testing.T) {
	expectValue(t, "command-resolution", "name=demo:lifecycle arguments=1 options=1")
}

func TestLifecycleDemoCommandExecution(t *testing.T) {
	expectValue(t, "command-execution", "order=starting,handle,finished succeeded=true")
}

func TestLifecycleDemoSharedApplication(t *testing.T) {
	expectValue(t, "shared-application", "http-wrote=true console-read=from-http")
}

func TestLifecycleDemoRunnerShutdown(t *testing.T) {
	expectValue(t, "runner-shutdown", "runner-returned=true closed=true")
}

func TestLifecycleDemoRootContextCancel(t *testing.T) {
	expectValue(t, "root-context-cancel", "canceled-before-terminating=true")
}

func TestLifecycleDemoTerminatingEventOrder(t *testing.T) {
	expectValue(t, "terminating-event-order", "order=context-canceled,terminating,terminate")
}

func TestLifecycleDemoTerminableProviders(t *testing.T) {
	expectValue(t, "terminable-providers", "order=demo.p3,demo.p2,demo.p1")
}

func TestLifecycleDemoCleanupFunctions(t *testing.T) {
	expectValue(t, "cleanup-functions", "order=demo.c3,demo.c2,demo.c1")
}

func TestLifecycleDemoResourceCloseOrder(t *testing.T) {
	expectValue(t, "resource-close-order", "order=demo.normal,demo.reporting")
}

func TestLifecycleDemoShutdownErrorReporting(t *testing.T) {
	expectValue(t, "shutdown-error-reporting", "reported=true reporting-closed=true error=true")
}

func TestLifecycleDemoTerminatedEventOrder(t *testing.T) {
	expectValue(t, "terminated-event-order",
		"order=terminate,cleanup,terminated duration-positive=true close-duration-positive=true error=cleanup failed")
}

func TestLifecycleDemoCloseRetry(t *testing.T) {
	expectValue(t, "close-retry", "first-error=true retry-clean=true stable-once=true failing-twice=true")
}

func TestLifecycleDemoBestEffortEvents(t *testing.T) {
	expectValue(t, "best-effort-events", "boot-ok=true close-ok=true")
}

func TestLifecycleDemoAppBootingEvent(t *testing.T) {
	expectValue(t, "app-booting-event", "args=0 before-provider-boot=true")
}

func TestLifecycleDemoAppBootedEvent(t *testing.T) {
	expectValue(t, "app-booted-event", "duration-nonnegative=true after-provider-boot=true cache-bound=true")
}

func TestLifecycleDemoAppTerminatingEvent(t *testing.T) {
	expectValue(t, "app-terminating-event", "reason=application shutdown context-canceled=true")
}

func TestLifecycleDemoAppTerminatedEvent(t *testing.T) {
	expectValue(t, "app-terminated-event", "duration-positive=true close-duration-positive=true error-empty=true")
}

func TestLifecycleDemoProviderRegisteringEvent(t *testing.T) {
	expectValue(t, "provider-registering-event", "phase=registering provider=demo.p1 before-registered=true")
}

func TestLifecycleDemoProviderRegisteredEvent(t *testing.T) {
	expectValue(t, "provider-registered-event", "phase=registered provider=demo.p1 after-registering=true")
}

func TestLifecycleDemoProviderBootingEvent(t *testing.T) {
	expectValue(t, "provider-booting-event", "phase=booting provider=demo.p1 after-all-registers=true")
}

func TestLifecycleDemoProviderBootedEvent(t *testing.T) {
	expectValue(t, "provider-booted-event", "phase=booted provider=demo.p1 after-booting=true")
}

func TestLifecycleDemoServerStartingEvent(t *testing.T) {
	expectValue(t, "server-starting-event", "addr=:0 pid-positive=true before-started=true")
}

func TestLifecycleDemoServerStartedEvent(t *testing.T) {
	expectValue(t, "server-started-event", "addr=:0 pid-positive=true after-starting=true")
}

func TestLifecycleDemoServerStoppingEvent(t *testing.T) {
	expectValue(t, "server-stopping-event", "addr=:0 reason=context canceled after-started=true")
}

func TestLifecycleDemoServerStoppedEvent(t *testing.T) {
	expectValue(t, "server-stopped-event", "addr=:0 error-empty=true after-stopping=true")
}

func TestLifecycleDemoRequestReceivedEvent(t *testing.T) {
	expectValue(t, "request-received-event", "method=GET path=/received client-ip=true request-id=true received-at=true")
}

func TestLifecycleDemoRequestHandledEvent(t *testing.T) {
	expectValue(t, "request-handled-event", "status=200 duration-nonnegative=true handled=true")
}

func TestLifecycleDemoRequestFailedEvent(t *testing.T) {
	expectValue(t, "request-failed-event", "status=500 method=GET path=/failed error-recorded=true")
}

func TestLifecycleDemoRequestFinishedEvent(t *testing.T) {
	expectValue(t, "request-finished-event", "status=200 after-handled=true error-empty=true")
}

func TestLifecycleDemoConsoleApplicationStartingEvent(t *testing.T) {
	expectValue(t, "console-application-starting-event", "kernel=PrismGo")
}

func TestLifecycleDemoConsoleCommandStartingEvent(t *testing.T) {
	expectValue(t, "console-command-starting-event", "command=lifecycle:command-starting input-snapshot=true")
}

func TestLifecycleDemoConsoleCommandFinishedEvent(t *testing.T) {
	expectValue(t, "console-command-finished-event", "command=lifecycle:command-finished succeeded=true duration-nonnegative=true")
}

func TestLifecycleDemoListeners(t *testing.T) {
	expectValue(t, "listeners", "booted-listener=true terminated-listener=true")
}

func TestLifecycleDemoPayloadBoundaries(t *testing.T) {
	expectValue(t, "payload-boundaries", "payloads=19 serializable=true")
}

func TestLifecycleDemoRejectsUnknownScenario(t *testing.T) {
	_, err := lifecycledemo.Run("no-such-case")
	if err == nil || !strings.Contains(err.Error(), `unknown scenario "no-such-case"`) {
		t.Fatalf("lifecycle demo unknown scenario error = %v, want unknown scenario error", err)
	}
}
