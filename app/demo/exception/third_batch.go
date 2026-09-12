package exceptiondemo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/console"
	"github.com/prismgo/framework/container"
	queuecontract "github.com/prismgo/framework/contracts/queue"
	"github.com/prismgo/framework/event"
	"github.com/prismgo/framework/exception"
	"github.com/prismgo/framework/foundation"
	"github.com/prismgo/framework/kernel"
	"github.com/prismgo/framework/queue"
	"github.com/prismgo/framework/routine"
	"github.com/prismgo/horizon"
)

func thirdBatch(name string) (string, error) {
	switch name {
	case "handler-render":
		h := exception.New(exception.WithLogging(false))
		returned := 0
		response, err := serve(h, requestOptions{register: func(engine *gin.Engine) {
			engine.GET("/failure", func(c *gin.Context) { returned = h.Render(c, errDemo) })
		}})
		if err != nil {
			return "", err
		}
		p, err := problem(response)
		return fmt.Sprintf("returned=%d; response=%d; type=%s", returned, response.status, p.Type), err
	case "should-report":
		h := exception.New(exception.WithDontReport(func(err error) bool { return errors.Is(err, errDemo) }))
		ignored := h.ShouldReport(errDemo, 500)
		client := h.ShouldReport(errors.New("client"), 422)
		h.ApplyOptions(exception.WithClientErrorLogging(false))
		return fmt.Sprintf("ignored=%t; client=%t; disabled_client=%t", ignored, client, h.ShouldReport(errors.New("client"), 422)), nil
	case "handler-level":
		h := exception.New(exception.WithLevel(func(_ error, status int) exception.Level {
			if status == http.StatusTooManyRequests {
				return exception.LevelInfo
			}
			return ""
		}))
		return fmt.Sprintf("429=%s; 500=%s", h.Level(errDemo, 429), h.Level(errDemo, 500)), nil
	case "handler-debug":
		enabled := false
		h := exception.New(exception.WithDebugResolver(func() bool { return enabled }))
		before := h.Debug()
		enabled = true
		return fmt.Sprintf("before=%t; after=%t", before, h.Debug()), nil
	case "apply-options":
		h := exception.New(exception.WithLogging(false))
		h.ApplyOptions(exception.WithLogging(true), exception.WithDebug(true))
		return fmt.Sprintf("logging=%t; debug=%t", h.LogErrors, h.Debug()), nil
	case "cli-report":
		return cliReport()
	case "routine-report":
		return routineReport()
	case "queue-report":
		return queueReport()
	case "horizon-report":
		return horizonReport()
	case "event-report":
		return eventReport()
	case "non-http-fields":
		return nonHTTPFields()
	case "scrub-keys", "scrub-nested", "scrub-service-key":
		return scrubCase(name)
	case "wrap-handler":
		return wrappedHandler()
	case "replace-handler":
		return replacedHandler()
	case "laravel-mapping":
		var predicate exception.Predicate = func(error) bool { return false }
		var reporter exception.Reporter = func(any, error, map[string]any) {}
		var renderer exception.Renderer = func(*gin.Context, error) (exception.Problem, bool) {
			return exception.Problem{Type: "demo", Title: "Teapot", Status: http.StatusTeapot}, true
		}
		h := exception.New(exception.WithLogging(false), exception.WithDontReport(predicate), exception.WithReporter(reporter), exception.WithRenderer(renderer))
		response, err := render(h, errDemo, "")
		return fmt.Sprintf("dont_report=%d; reporters=%d; status=%d; level=%s", len(h.DontReport), len(h.Reporters), response.status, h.Level(errDemo, 500)), err
	default:
		return "", fmt.Errorf("unknown third-batch exception scenario %q", name)
	}
}

// observeHandler temporarily installs a reporter in the current application.
func observeHandler(run func(context.Context) error) (reported []observedReport, err error) {
	app := foundation.App
	if app == nil {
		return nil, errors.New("exception demo application is unavailable")
	}
	previous, err := app.Container().Make(foundation.ContainerKeyExceptionHandler)
	if err != nil {
		return nil, fmt.Errorf("resolve exception handler: %w", err)
	}
	var mu sync.Mutex
	h := exception.New(exception.WithPanicStack(false), exception.WithReporter(func(_ any, _ error, fields map[string]any) {
		mu.Lock()
		reported = append(reported, observedReport{fields: fields})
		mu.Unlock()
	}))
	if err := app.Container().Instance(foundation.ContainerKeyExceptionHandler, h); err != nil {
		return nil, fmt.Errorf("install exception reporter: %w", err)
	}
	defer func() {
		if restoreErr := app.Container().Instance(foundation.ContainerKeyExceptionHandler, previous); restoreErr != nil {
			err = errors.Join(err, fmt.Errorf("restore exception handler: %w", restoreErr))
		}
	}()
	err = run(context.Background())
	return reported, err
}

type observedReport struct {
	fields map[string]any
}

func cliReport() (string, error) {
	reports, err := observeHandler(func(ctx context.Context) error {
		k := kernel.New("exception-demo")
		k.RegisterClosure(console.Definition{Name: "exception:fail", Description: "Report a failed demo command"}, func(console.CommandContext) error { return errDemo })
		k.RegisterClosure(console.Definition{Name: "exception:panic", Description: "Report a panicking demo command"}, func(console.CommandContext) error { panic("demo CLI panic") })
		failure := k.CallSilently(ctx, "exception:fail")
		if !errors.Is(failure, errDemo) {
			return fmt.Errorf("CLI command error = %v, want %v", failure, errDemo)
		}
		panicErr := k.CallSilently(ctx, "exception:panic")
		if panicErr == nil || !strings.Contains(panicErr.Error(), "panic recovered") {
			return fmt.Errorf("CLI panic error = %v, want recovered panic", panicErr)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(reports) != 2 {
		return "", fmt.Errorf("CLI reports = %d, want 2", len(reports))
	}
	return fmt.Sprintf("reports=%d; component=%v; commands=%v,%v", len(reports), firstField(reports, "component"), firstField(reports, "command"), reports[len(reports)-1].fields["command"]), nil
}

func routineReport() (string, error) {
	reports, err := observeHandler(func(ctx context.Context) error {
		done := make(chan struct{}, 2)
		for _, task := range []func(context.Context) error{
			func(context.Context) error { return errDemo },
			func(context.Context) error { panic("routine demo panic") },
		} {
			routine.Task(ctx, task).Name("exception-demo").Component("routine").OnError(func(error) { done <- struct{}{} }).OnPanic(func(error) { done <- struct{}{} }).Go()
		}
		for range 2 {
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				return errors.New("wait for routine exception reports: timeout")
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("reports=%d; component=%v; routine=%v", len(reports), firstField(reports, "component"), firstField(reports, "routine")), nil
}

type demoQueueConnection struct{ queuecontract.Queue }

type demoQueueConnector struct{ connection queuecontract.Queue }

func (c demoQueueConnector) Connect(context.Context, string, queuecontract.ConnectorConfig) (queuecontract.Queue, error) {
	return demoQueueConnection{Queue: c.connection}, nil
}

type failingQueueJob struct{}

func (*failingQueueJob) Handle(context.Context) error { return errDemo }

func queueReport() (string, error) {
	reports, err := observeHandler(func(ctx context.Context) (runErr error) {
		manager, err := queue.NewManager(queue.Config{
			Default: "exception-demo",
			Connections: map[string]queue.ConnectionConfig{
				"exception-demo": {Driver: "exception-demo", Queue: "exception-demo"},
			},
		}, queue.NewRegistry())
		if err != nil {
			return fmt.Errorf("create demo queue manager: %w", err)
		}
		defer func() {
			if closeErr := manager.Close(); closeErr != nil {
				runErr = errors.Join(runErr, fmt.Errorf("close demo queue manager: %w", closeErr))
			}
		}()
		manager.Extend("exception-demo", func() (queuecontract.Connector, error) {
			return demoQueueConnector{connection: queue.NewSyncConnection()}, nil
		})
		if _, err := manager.Dispatch(ctx, &failingQueueJob{}, queue.OnConnection("exception-demo"), queue.OnQueue("exception-demo"), queue.Tries(1)); err != nil {
			return fmt.Errorf("dispatch failing queue job: %w", err)
		}
		workerCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if err := queue.NewWorker(manager).Work(workerCtx, queue.WorkerOptions{Connection: "exception-demo", Queues: []string{"exception-demo"}, Once: true, Tries: 1}); err != nil {
			return fmt.Errorf("run failing queue job: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("reports=%d; component=%v; subsystem=%v", len(reports), firstField(reports, "component"), firstField(reports, "subsystem")), nil
}

func firstField(reports []observedReport, key string) any {
	if len(reports) == 0 {
		return nil
	}
	return reports[0].fields[key]
}

type demoHorizonStore struct{ store horizon.Store }

func (s demoHorizonStore) ResolveStore(context.Context, horizon.Config) (horizon.Store, error) {
	return s.store, nil
}

type failingHorizonControlStore struct {
	horizon.Store
	calls int
}

func (s *failingHorizonControlStore) Control(ctx context.Context) (horizon.ControlState, error) {
	s.calls++
	if s.calls == 2 {
		return horizon.ControlState{}, errDemo
	}
	return s.Store.Control(ctx)
}

type failingWorkerRunner struct{}

func (failingWorkerRunner) Begin(context.Context, queue.WorkerOptions) (queuecontract.WorkerSession, error) {
	return failingWorkerSession{}, nil
}

type failingWorkerSession struct{}

func (failingWorkerSession) Activate(context.Context) error { return nil }
func (failingWorkerSession) Work(context.Context) error     { return errDemo }
func (failingWorkerSession) Close() error                   { return nil }

func horizonReport() (value string, err error) {
	app := foundation.App
	if app == nil {
		return "", errors.New("exception demo application is unavailable")
	}
	bound := app.Container().Bound("horizon.manager")
	var previous any
	if bound {
		previous, err = app.Container().Make("horizon.manager")
		if err != nil {
			return "", fmt.Errorf("resolve Horizon manager: %w", err)
		}
	}
	manager, err := horizon.NewManager(horizon.Config{Store: "memory"},
		horizon.WithStoreFactory(demoHorizonStore{store: horizon.NewMemoryStore(horizon.StoreOptions{})}),
		horizon.WithWorkerRunner(failingWorkerRunner{}),
	)
	if err != nil {
		return "", fmt.Errorf("create Horizon demo manager: %w", err)
	}
	if err := app.Container().Instance("horizon.manager", manager); err != nil {
		return "", fmt.Errorf("install Horizon demo manager: %w", err)
	}
	defer func() {
		var restoreErr error
		if bound {
			restoreErr = app.Container().Instance("horizon.manager", previous)
		} else {
			registry, ok := app.Container().(*container.Container)
			if !ok {
				restoreErr = fmt.Errorf("Horizon demo container = %T, want *container.Container", app.Container())
			} else {
				restoreErr = registry.Forget("horizon.manager")
			}
		}
		if restoreErr != nil {
			err = errors.Join(err, fmt.Errorf("restore Horizon manager: %w", restoreErr))
		}
	}()
	reports, err := observeHandler(func(ctx context.Context) error {
		k := kernel.New("exception-demo")
		foundWork := false
		foundSupervisor := false
		for _, factory := range horizon.CommandFactories() {
			command := factory()
			switch command.Definition().Name {
			case "horizon:work":
				foundWork = true
				k.Register(command)
			case "horizon:supervisor":
				foundSupervisor = true
				k.Register(command)
			}
		}
		if !foundWork || !foundSupervisor {
			return fmt.Errorf("Horizon demo commands found: work=%t supervisor=%t, want both", foundWork, foundSupervisor)
		}
		failure := k.CallSilently(ctx, "horizon:work", console.CallInput{Options: map[string]any{"name": "exception-demo-worker", "once": true}})
		if !errors.Is(failure, errDemo) {
			return fmt.Errorf("Horizon worker error = %v, want %v", failure, errDemo)
		}
		store := &failingHorizonControlStore{Store: horizon.NewMemoryStore(horizon.StoreOptions{})}
		supervisorManager, err := horizon.NewManager(horizon.Config{
			Store: "memory", LoopInterval: 10 * time.Millisecond,
			Supervisors: map[string]horizon.SupervisorConfig{
				"exception-demo-supervisor": {Name: "exception-demo-supervisor"},
			},
		}, horizon.WithStoreFactory(demoHorizonStore{store: store}))
		if err != nil {
			return fmt.Errorf("create Horizon supervisor demo manager: %w", err)
		}
		if err := app.Container().Instance("horizon.manager", supervisorManager); err != nil {
			return fmt.Errorf("install Horizon supervisor demo manager: %w", err)
		}
		supervisorCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		if err := k.CallSilently(supervisorCtx, "horizon:supervisor", console.CallInput{Arguments: map[string]any{"name": "exception-demo-supervisor", "connection": "sync"}}); err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("run Horizon supervisor demo: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	count := 0
	for _, report := range reports {
		if report.fields["component"] == "horizon" {
			count++
		}
	}
	return fmt.Sprintf("reports=%d; component=%v; subsystems=%v,%v", count, firstField(reports, "component"), firstField(reports, "subsystem"), lastHorizonSubsystem(reports)), nil
}

func lastHorizonSubsystem(reports []observedReport) any {
	for i := len(reports) - 1; i >= 0; i-- {
		if reports[i].fields["component"] == "horizon" {
			return reports[i].fields["subsystem"]
		}
	}
	return nil
}

type demoEvent struct{}

func (demoEvent) Name() string { return "exception.demo.failed" }

func eventReport() (string, error) {
	reports, err := observeHandler(func(ctx context.Context) error {
		bus := event.New()
		bus.ListenFunc("exception.demo.failed", func(context.Context, event.Event) error { return errDemo })
		bus.ListenFunc("exception.demo.failed", func(context.Context, event.Event) error { panic("event demo panic") })
		bus.Dispatch(ctx, demoEvent{})
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("reports=%d; component=%v; event=%v", len(reports), firstField(reports, "component"), firstField(reports, "event")), nil
}

func nonHTTPFields() (string, error) {
	reports, err := observeHandler(func(ctx context.Context) error {
		exception.Report(ctx, errDemo, map[string]any{"component": "demo", "caller": "non-http", "status": 500})
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("reports=%d; status=%v; caller=%v", len(reports), firstField(reports, "status"), firstField(reports, "caller")), nil
}

func scrubCase(name string) (string, error) {
	reports, err := observeHandler(func(ctx context.Context) error {
		exception.Report(ctx, errDemo, map[string]any{
			"password": "secret-value", "api_token": "token-value", "service_key": "exception.handler",
			"nested": map[string]any{"authorization": "bearer secret", "items": []any{map[string]string{"cookie": "session-value"}}},
		})
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(reports) != 1 {
		return "", fmt.Errorf("scrub reports = %d, want 1", len(reports))
	}
	fields := reports[0].fields
	switch name {
	case "scrub-keys":
		return fmt.Sprintf("password=%v; token=%v", fields["password"], fields["api_token"]), nil
	case "scrub-nested":
		nested, ok := fields["nested"].(map[string]any)
		if !ok {
			return "", fmt.Errorf("nested fields = %T, want map[string]any", fields["nested"])
		}
		items, ok := nested["items"].([]any)
		if !ok || len(items) != 1 {
			return "", fmt.Errorf("nested items = %#v, want one item", nested["items"])
		}
		item, ok := items[0].(map[string]string)
		if !ok {
			return "", fmt.Errorf("nested item = %T, want map[string]string", items[0])
		}
		return fmt.Sprintf("authorization=%v; cookie=%v", nested["authorization"], item["cookie"]), nil
	default:
		return fmt.Sprintf("service_key=%v", fields["service_key"]), nil
	}
}

func wrappedHandler() (string, error) {
	called := false
	h := exception.BuildAndRegister(nil, func(next *exception.Handler) *exception.Handler {
		next.ApplyOptions(exception.WithReporter(func(any, error, map[string]any) { called = true }))
		return next
	})
	h.Report(context.Background(), errDemo, nil)
	response, err := render(h, errDemo, "")
	return fmt.Sprintf("reported=%t; status=%d; safe=%t", called, response.status, !strings.Contains(response.body, errDemo.Error())), err
}

func replacedHandler() (string, error) {
	replacement := exception.New(exception.WithLogging(false), exception.WithRecovery(false), exception.WithPanicStack(false), exception.WithResponseRenderer(func(c *gin.Context, _ error) bool {
		c.Data(http.StatusServiceUnavailable, "text/plain", []byte("service unavailable"))
		return true
	}))
	h := exception.BuildAndRegister([]exception.Option{exception.WithDebug(true)}, func(*exception.Handler) *exception.Handler { return replacement })
	response, err := render(h, errDemo, "")
	return fmt.Sprintf("replaced=%t; status=%d; body=%s; logging=%t; recovery=%t; debug=%t", h == replacement, response.status, response.body, h.LogErrors, h.RecoverPanics, h.Debug()), err
}
