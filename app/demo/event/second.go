package eventdemo

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	frameworkcmd "github.com/prismgo/framework/cmd"
	"github.com/prismgo/framework/console"
	"github.com/prismgo/framework/container"
	containercontract "github.com/prismgo/framework/contracts/container"
	eventcontract "github.com/prismgo/framework/contracts/event"
	providercontract "github.com/prismgo/framework/contracts/provider"
	"github.com/prismgo/framework/event"
	"github.com/prismgo/framework/exception"
	"github.com/prismgo/framework/foundation"
	frameworkhttp "github.com/prismgo/framework/http"
	"github.com/prismgo/framework/http/middleware"
	"github.com/prismgo/framework/kernel"
	"github.com/prismgo/framework/provider/publish"
	"github.com/prismgo/framework/queue"
)

func runSecondBatch(ctx context.Context, name, connection string) (string, error) {
	switch name {
	case "async-durability", "consistency-boundary", "laravel-compatibility":
		return runBoundary(ctx, name), nil
	case "app-lifecycle", "provider-lifecycle":
		return runApplicationLifecycle(name)
	case "server-lifecycle":
		return runServerLifecycle(ctx)
	case "request-lifecycle", "request-finished-ordering":
		return runRequestLifecycle(name)
	case "console-lifecycle":
		return runConsoleLifecycle(ctx)
	case "vendor-publish-event":
		return runVendorPublish(ctx)
	case "listener-error-isolation", "listener-panic-isolation", "async-failure-isolation", "queued-failure-handling", "isolated-testing", "queued-sync-testing", "queued-worker-testing":
		return runFailureAndTesting(ctx, name, connection)
	default:
		return runContract(ctx, name)
	}
}

func runBoundary(ctx context.Context, name string) string {
	switch name {
	case "async-durability":
		_, async := event.Async(func(context.Context, event.Event) error { return nil }).(event.AsyncListener)
		_, queued := event.Queued(event.ListenerFunc(func(context.Context, event.Event) error { return nil })).(event.ShouldQueue)
		return fmt.Sprintf("async-goroutine=%t; queued-worker-path=%t; durability-depends-on-driver", async, queued)
	case "consistency-boundary":
		bus := event.New()
		bus.ListenFunc("demo.consistency", func(context.Context, event.Event) error { return errors.New("listener failed") })
		bus.Dispatch(ctx, message{EventName: "demo.consistency"})
		return "dispatch-return=void; use direct service call for strong consistency"
	default:
		return "PrismGo: named Event, Listener, Subscriber, sync/async/queued; no propagation stop or automatic listener discovery"
	}
}

type demoProvider struct{ calls *[]string }

func (demoProvider) Name() string { return "demo.event.provider" }
func (p demoProvider) Register(providercontract.Application) error {
	*p.calls = append(*p.calls, "register")
	return nil
}
func (p demoProvider) Boot(providercontract.Application) error {
	*p.calls = append(*p.calls, "boot")
	return nil
}

func runApplicationLifecycle(name string) (string, error) {
	previous := foundation.App
	base, err := os.MkdirTemp("", "prismgo-event-lifecycle-")
	if err != nil {
		return "", fmt.Errorf("create lifecycle directory: %w", err)
	}
	defer os.RemoveAll(base)
	app := foundation.NewApplication(base)
	defer func() {
		foundation.App = previous
		if previous != nil {
			registry := previous.Container().(*container.Container)
			container.SetProvider(func() *container.Container { return registry })
		} else {
			container.SetProvider(nil)
		}
	}()
	bus := event.New()
	if err := app.Container().Instance("event.dispatcher", bus); err != nil {
		return "", fmt.Errorf("bind lifecycle dispatcher: %w", err)
	}
	var names []string
	bus.Listen("app.*", event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if name == "app-lifecycle" && (ev.Name() == event.EventAppBooting || ev.Name() == event.EventAppBooted || ev.Name() == event.EventAppTerminating || ev.Name() == event.EventAppTerminated) {
			names = append(names, ev.Name())
		}
		if name == "provider-lifecycle" && strings.HasPrefix(ev.Name(), "app.provider.") {
			if providerName := providerEventName(ev); providerName == "demo.event.provider" {
				names = append(names, ev.Name())
			}
		}
		return nil
	}))
	var calls []string
	if name == "provider-lifecycle" {
		if err := app.RegisterProvider(demoProvider{calls: &calls}); err != nil {
			return "", fmt.Errorf("register demo provider: %w", err)
		}
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot lifecycle application: %w", err)
	}
	if err := app.Close(); err != nil {
		return "", fmt.Errorf("close lifecycle application: %w", err)
	}
	if name == "provider-lifecycle" && !reflect.DeepEqual(calls, []string{"register", "boot"}) {
		return "", fmt.Errorf("provider calls = %v, want [register boot]", calls)
	}
	return strings.Join(names, ","), nil
}

func providerEventName(ev event.Event) string {
	switch value := ev.(type) {
	case event.ProviderRegistering:
		return value.Provider
	case event.ProviderRegistered:
		return value.Provider
	case event.ProviderBooting:
		return value.Provider
	case event.ProviderBooted:
		return value.Provider
	default:
		return ""
	}
}

func runServerLifecycle(ctx context.Context) (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("listen for lifecycle demo: %w", err)
	}
	defer listener.Close()
	serverCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	bus := event.New()
	var names []string
	bus.Listen("server.*", event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		names = append(names, ev.Name())
		if ev.Name() == event.EventServerStarted {
			cancel()
		}
		return nil
	}))
	server := frameworkhttp.NewServer(listener.Addr().String(), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }), time.Second)
	err = frameworkhttp.ListenAndServeGracefulContext(serverCtx, server, time.Second, frameworkhttp.WithListener(listener), frameworkhttp.WithDispatcher(bus))
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return "", fmt.Errorf("serve lifecycle demo: %w", err)
	}
	return strings.Join(names, ","), nil
}

func runRequestLifecycle(name string) (string, error) {
	bus := event.New()
	var names []string
	var finished []event.RequestFinished
	bus.Listen("request.*", event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		names = append(names, ev.Name())
		if value, ok := ev.(event.RequestFinished); ok {
			finished = append(finished, value)
		}
		return nil
	}))
	engine := gin.New()
	engine.Use(middleware.Event(bus))
	engine.GET("/demo", func(c *gin.Context) {
		if c.Query("fail") == "1" {
			c.String(http.StatusInternalServerError, "failure")
			return
		}
		c.String(http.StatusOK, "ok")
	})
	for _, path := range []string{"/demo", "/demo?fail=1"} {
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	}
	if len(finished) != 2 || finished[0].Status != 200 || finished[1].Status != 500 {
		return "", fmt.Errorf("finished payloads = %+v, want statuses 200 and 500", finished)
	}
	if name == "request-lifecycle" {
		return fmt.Sprintf("%s; status=%d,%d", strings.Join(names, ","), finished[0].Status, finished[1].Status), nil
	}
	return strings.Join(names, ","), nil
}

func runConsoleLifecycle(ctx context.Context) (string, error) {
	bus := event.Resolve()
	if bus == nil {
		return "", errors.New("application event dispatcher is unavailable")
	}
	var names []string
	listener := event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		names = append(names, ev.Name())
		return nil
	})
	for _, eventName := range []string{event.EventConsoleApplicationStarting, event.EventCommandStarting, event.EventCommandFinished} {
		bus.Listen(eventName, listener)
		defer bus.Forget(eventName)
	}
	commandKernel := kernel.New("demo-event")
	commandKernel.RegisterClosure(console.Definition{Name: "event:probe", Description: "Event lifecycle probe"}, func(console.CommandContext) error { return nil })
	if err := commandKernel.Call(ctx, "event:probe"); err != nil {
		return "", fmt.Errorf("run lifecycle command: %w", err)
	}
	return strings.Join(names, ","), nil
}

func runVendorPublish(ctx context.Context) (string, error) {
	base, err := os.MkdirTemp("", "prismgo-event-publish-")
	if err != nil {
		return "", fmt.Errorf("create publish directory: %w", err)
	}
	defer os.RemoveAll(base)
	source, target := filepath.Join(base, "source.txt"), filepath.Join(base, "target.txt")
	if err := os.WriteFile(source, []byte("event demo"), 0o600); err != nil {
		return "", fmt.Errorf("write publish source: %w", err)
	}
	providerName := fmt.Sprintf("event-demo-%d", time.Now().UnixNano())
	if err := publish.Register(providerName, map[string]string{source: target}, "event-demo"); err != nil {
		return "", fmt.Errorf("register publish source: %w", err)
	}
	bus := event.Resolve()
	if bus == nil {
		return "", errors.New("application event dispatcher is unavailable")
	}
	var captured event.VendorTagPublished
	bus.ListenFunc(event.EventVendorTagPublished, func(_ context.Context, ev event.Event) error {
		captured = ev.(event.VendorTagPublished)
		return nil
	})
	defer bus.Forget(event.EventVendorTagPublished)
	commandKernel := kernel.New("demo-event")
	commandKernel.Register(frameworkcmd.NewVendorPublishCommand())
	if err := commandKernel.Call(ctx, "vendor:publish --provider="+providerName+" --tag=event-demo"); err != nil {
		return "", fmt.Errorf("publish demo resource: %w", err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		return "", fmt.Errorf("read published resource: %w", err)
	}
	if string(data) != "event demo" || captured.Published != 1 {
		return "", fmt.Errorf("published data = %q, event = %+v, want event demo and 1 published", data, captured)
	}
	return fmt.Sprintf("tag=%s; published=%d; skipped=%d", strings.Join(captured.Tag, ","), captured.Published, captured.Skipped), nil
}

type providerView struct{ registry *container.Container }

func (p providerView) Container() containercontract.Container { return p.registry }

func runContract(ctx context.Context, name string) (string, error) {
	bus := event.New()
	var count atomic.Int32
	listener := counter{count: &count}
	ev := message{EventName: "demo.event.contract"}
	switch name {
	case "event-interface":
		var contract eventcontract.Event = ev
		return "name=" + contract.Name(), nil
	case "listener-interface":
		var contract eventcontract.Listener = listener
		if err := contract.Handle(ctx, ev); err != nil {
			return "", err
		}
	case "listener-func-interface":
		var contract eventcontract.Listener = event.ListenerFunc(func(context.Context, event.Event) error { count.Add(1); return nil })
		if err := contract.Handle(ctx, ev); err != nil {
			return "", err
		}
	case "dispatcher-interface":
		var contract eventcontract.Dispatcher = bus
		contract.Listen(ev.Name(), listener)
		contract.Dispatch(ctx, ev)
		return fmt.Sprintf("has=%t; handled=%d", contract.Has(ev.Name()), count.Load()), nil
	case "subscriber-interface":
		var contract eventcontract.Subscriber = subscriber{count: &count}
		bus.Subscribe(contract)
		bus.Dispatch(ctx, message{EventName: "demo.event.created"})
	case "should-queue-interface":
		var contract event.ShouldQueue = queuedCounter{counter: listener}
		return fmt.Sprintf("should-queue=%t", contract.ShouldQueue()), nil
	case "async-listener-interface":
		contract, ok := event.Async(func(context.Context, event.Event) error { return nil }).(event.AsyncListener)
		return fmt.Sprintf("async=%t; implemented=%t", contract.Async(), ok), nil
	case "queue-options-interface":
		var contract event.QueueOptionsProvider = routedCounter{connection: "sync", queueName: "event-demo", tries: 3}
		return fmt.Sprintf("connection=%s; queue=%s; tries=%d", contract.QueueConnection(), contract.QueueName(), contract.QueueTries()), nil
	case "provider-register":
		registry := container.NewContainer()
		if err := (event.ServiceProvider{}).Register(providerView{registry: registry}); err != nil {
			return "", fmt.Errorf("register event provider: %w", err)
		}
		resolved, err := registry.Make("event.dispatcher")
		if err != nil {
			return "", fmt.Errorf("resolve event dispatcher: %w", err)
		}
		return fmt.Sprintf("bound=%t; dispatcher=%T", registry.Bound("event.dispatcher"), resolved), nil
	case "provider-boot":
		if err := (event.ServiceProvider{}).Boot(providerView{registry: container.NewContainer()}); err != nil {
			return "", fmt.Errorf("boot event provider: %w", err)
		}
		const queuedJob = "github.com/prismgo/framework/event.queuedListenerJob"
		return fmt.Sprintf("queued-job-registered=%t", queue.DefaultRegistry().Has(queuedJob)), nil
	default:
		return "", fmt.Errorf("unknown second-batch contract %q", name)
	}
	return fmt.Sprintf("handled=%d", count.Load()), nil
}

func runFailureAndTesting(ctx context.Context, name, connection string) (value string, err error) {
	if name == "queued-worker-testing" {
		if connection != "redis" {
			return "", errors.New("queued-worker-testing requires --connection=redis")
		}
		return runQueued(ctx, "queued-redis", connection)
	}
	if name == "queued-sync-testing" {
		return runQueued(ctx, "queued-sync", "sync")
	}
	bus := event.New()
	var handled atomic.Int32
	var reported atomic.Int32
	reportedDone := make(chan struct{}, 1)
	if name != "isolated-testing" {
		handler := exception.New(exception.WithReporter(func(_ any, _ error, fields map[string]any) {
			if fields["component"] == "event" {
				reported.Add(1)
				select {
				case reportedDone <- struct{}{}:
				default:
				}
			}
		}))
		app := foundation.App
		if app == nil {
			return "", errors.New("application is unavailable for event exception reporting")
		}
		previous, resolveErr := app.Container().Make("exception.handler")
		if resolveErr != nil {
			return "", fmt.Errorf("resolve existing exception handler: %w", resolveErr)
		}
		if bindErr := app.Container().Instance("exception.handler", handler); bindErr != nil {
			return "", fmt.Errorf("bind event exception reporter: %w", bindErr)
		}
		defer func() {
			if restoreErr := app.Container().Instance("exception.handler", previous); restoreErr != nil && err == nil {
				err = fmt.Errorf("restore exception handler: %w", restoreErr)
			}
		}()
	}
	eventName := "demo.event.failure"
	failure := event.ListenerFunc(func(context.Context, event.Event) error { return errors.New("demo listener failure") })
	switch name {
	case "listener-error-isolation":
		bus.Listen(eventName, failure)
	case "listener-panic-isolation":
		bus.ListenFunc(eventName, func(context.Context, event.Event) error { panic("demo listener panic") })
	case "async-failure-isolation":
		bus.Listen(eventName, event.Async(func(context.Context, event.Event) error { return errors.New("demo async failure") }))
	case "queued-failure-handling":
		bus.Listen(eventName, event.Queued(failure))
	case "isolated-testing":
		bus.ListenFunc(eventName, func(context.Context, event.Event) error { handled.Add(1); return nil })
		bus.Dispatch(ctx, message{EventName: "demo.event.other"})
	default:
		return "", fmt.Errorf("unknown failure scenario %q", name)
	}
	if name != "isolated-testing" {
		bus.Listen(eventName, counter{count: &handled})
	}
	bus.Dispatch(ctx, message{EventName: eventName})
	if name == "async-failure-isolation" {
		select {
		case <-reportedDone:
		case <-time.After(3 * time.Second):
			return "", errors.New("async failure was not reported within 3s")
		}
	}
	if handled.Load() != 1 {
		return "", fmt.Errorf("healthy listener count = %d, want 1", handled.Load())
	}
	if name != "isolated-testing" && reported.Load() != 1 {
		return "", fmt.Errorf("reported listener failures = %d, want 1", reported.Load())
	}
	return fmt.Sprintf("handled=%d; reported=%d", handled.Load(), reported.Load()), nil
}
