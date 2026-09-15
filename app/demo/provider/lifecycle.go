package providerdemo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/prismgo/framework/container"
	containercontract "github.com/prismgo/framework/contracts/container"
	providercontract "github.com/prismgo/framework/contracts/provider"
	"github.com/prismgo/framework/foundation"
)

// deferredResolutionScenario verifies Register and Boot run on the first deferred resolution.
func deferredResolutionScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	registers, boots := 0, 0
	provider := deferredProvider{
		id:   "demo.deferred",
		keys: []string{"demo.deferred.service"},
		register: func(app providercontract.Application) error {
			registers++
			return app.Container().Singleton("demo.deferred.service", func(containercontract.Resolver) (any, error) {
				return "deferred", nil
			})
		},
		boot: func(providercontract.Application) error { boots++; return nil },
	}
	if err := app.RegisterProvider(provider); err != nil {
		return "", fmt.Errorf("register deferred provider: %w", err)
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot deferred-resolution application: %w", err)
	}
	boundBefore := app.Container().Bound("demo.deferred.service")
	raw, err := app.Container().Make("demo.deferred.service")
	if err != nil {
		return "", fmt.Errorf("resolve deferred service: %w", err)
	}
	return fmt.Sprintf("bound-before=%t register=%d boot=%d value=%v", boundBefore, registers, boots, raw), nil
}

// deferredMapCleanupScenario verifies the deferred service map is dropped after loading.
func deferredMapCleanupScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	first := deferredProvider{
		id:   "demo.deferred.first",
		keys: []string{"demo.deferred.shared"},
		register: func(app providercontract.Application) error {
			return bindDemoValue(app, "demo.deferred.shared", "first")
		},
	}
	if err := app.RegisterProvider(first); err != nil {
		return "", fmt.Errorf("register first deferred provider: %w", err)
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot deferred-map-cleanup application: %w", err)
	}
	if _, err := app.Container().Make("demo.deferred.shared"); err != nil {
		return "", fmt.Errorf("resolve loaded deferred service: %w", err)
	}
	second := deferredProvider{
		id:   "demo.deferred.second",
		keys: []string{"demo.deferred.shared"},
		register: func(app providercontract.Application) error {
			return bindDemoValue(app, "demo.deferred.shared", "second")
		},
	}
	remapped := app.RegisterProvider(second) == nil
	return fmt.Sprintf("loaded=true remapped=%t", remapped), nil
}

// deferredLateBootScenario verifies a provider resolved after startup boots immediately.
func deferredLateBootScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	boots := 0
	provider := deferredProvider{
		id:   "demo.deferred.late",
		keys: []string{"demo.deferred.late"},
		register: func(app providercontract.Application) error {
			return bindDemoValue(app, "demo.deferred.late", "late")
		},
		boot: func(providercontract.Application) error { boots++; return nil },
	}
	if err := app.RegisterProvider(provider); err != nil {
		return "", fmt.Errorf("register late deferred provider: %w", err)
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot late deferred application: %w", err)
	}
	bootedBefore := boots > 0
	if _, err := app.Container().Make("demo.deferred.late"); err != nil {
		return "", fmt.Errorf("resolve late deferred service: %w", err)
	}
	return fmt.Sprintf("booted-before=%t booted-after=%t", bootedBefore, boots > 0), nil
}

// deferredEmptyScenario verifies an empty Provides() declaration is rejected.
func deferredEmptyScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	err := app.RegisterProvider(deferredProvider{id: "demo.deferred.empty"})
	return fmt.Sprintf("empty-error=%t", err != nil), nil
}

// deferredConflictScenario verifies two deferred providers cannot share a service key.
func deferredConflictScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	firstErr := app.RegisterProvider(deferredProvider{id: "demo.deferred.a", keys: []string{"demo.deferred.conflict"}})
	secondErr := app.RegisterProvider(deferredProvider{id: "demo.deferred.b", keys: []string{"demo.deferred.conflict"}})
	return fmt.Sprintf("first-error=%t second-error=%t", firstErr != nil, secondErr != nil), nil
}

// deferredTerminationScenario verifies unloaded deferred providers skip termination.
func deferredTerminationScenario(base string) (string, error) {
	unloaded := 0
	unloadedApp := foundation.NewApplication(base)
	unloadedProvider := deferredTerminableProvider{
		deferredProvider: deferredProvider{id: "demo.deferred.unloaded", keys: []string{"demo.deferred.unloaded"}},
		terminate:        func(context.Context) error { unloaded++; return nil },
	}
	if err := unloadedApp.RegisterProvider(unloadedProvider); err != nil {
		return "", fmt.Errorf("register unloaded deferred provider: %w", err)
	}
	if err := unloadedApp.Boot(); err != nil {
		return "", fmt.Errorf("boot unloaded deferred application: %w", err)
	}
	if err := unloadedApp.Close(); err != nil {
		return "", fmt.Errorf("close unloaded deferred application: %w", err)
	}

	loaded := 0
	loadedApp := foundation.NewApplication(base)
	loadedProvider := deferredTerminableProvider{
		deferredProvider: deferredProvider{
			id:   "demo.deferred.loaded",
			keys: []string{"demo.deferred.loaded"},
			register: func(app providercontract.Application) error {
				return bindDemoValue(app, "demo.deferred.loaded", "loaded")
			},
		},
		terminate: func(context.Context) error { loaded++; return nil },
	}
	if err := loadedApp.RegisterProvider(loadedProvider); err != nil {
		return "", fmt.Errorf("register loaded deferred provider: %w", err)
	}
	if err := loadedApp.Boot(); err != nil {
		return "", fmt.Errorf("boot loaded deferred application: %w", err)
	}
	if _, err := loadedApp.Container().Make("demo.deferred.loaded"); err != nil {
		return "", fmt.Errorf("resolve loaded deferred service: %w", err)
	}
	if err := loadedApp.Close(); err != nil {
		return "", fmt.Errorf("close loaded deferred application: %w", err)
	}
	return fmt.Sprintf("unloaded=%d loaded=%d", unloaded, loaded), nil
}

// workerLifecycleScenario verifies a terminable provider starts and drains a worker.
func workerLifecycleScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	pool := &workerPool{}
	started := false
	provider := terminableProvider{
		id: "demo.worker",
		boot: func(app providercontract.Application) error {
			pool.Start()
			started = pool.running
			return app.Container().Instance("demo.worker.pool", pool)
		},
		terminate: func(ctx context.Context) error {
			raw, err := app.Container().Make("demo.worker.pool")
			if err != nil {
				return err
			}
			resolved, ok := raw.(*workerPool)
			if !ok {
				return fmt.Errorf("worker pool has type %T", raw)
			}
			return resolved.Drain(ctx)
		},
	}
	if err := app.RegisterProvider(provider); err != nil {
		return "", fmt.Errorf("register worker provider: %w", err)
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot worker application: %w", err)
	}
	if err := app.Close(); err != nil {
		return "", fmt.Errorf("close worker application: %w", err)
	}
	return fmt.Sprintf("started=%t drained=%t", started, pool.drained), nil
}

// terminateOrderScenario verifies providers terminate in reverse registration order.
func terminateOrderScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	order := make([]string, 0, 3)
	for _, id := range []string{"p1", "p2", "p3"} {
		name := id
		provider := terminableProvider{
			id:        name,
			terminate: func(context.Context) error { order = append(order, name); return nil },
		}
		if err := app.RegisterProvider(provider); err != nil {
			return "", fmt.Errorf("register terminate provider %s: %w", name, err)
		}
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot terminate-order application: %w", err)
	}
	if err := app.Close(); err != nil {
		return "", fmt.Errorf("close terminate-order application: %w", err)
	}
	return "order=" + strings.Join(order, ","), nil
}

// terminateEligibilityScenario verifies providers with failed Register skip termination.
func terminateEligibilityScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	goodTerminated := false
	good := terminableProvider{
		id:        "demo.good",
		terminate: func(context.Context) error { goodTerminated = true; return nil },
	}
	failedTerminated := false
	failing := terminableProvider{
		id:        "demo.failing",
		register:  func(providercontract.Application) error { return errors.New("register failed") },
		terminate: func(context.Context) error { failedTerminated = true; return nil },
	}
	if err := app.RegisterProvider(good); err != nil {
		return "", fmt.Errorf("register good provider: %w", err)
	}
	if err := app.RegisterProvider(failing); err != nil {
		return "", fmt.Errorf("register failing provider: %w", err)
	}
	bootErr := app.Boot()
	if err := app.Close(); err != nil {
		return "", fmt.Errorf("close terminate-eligibility application: %w", err)
	}
	return fmt.Sprintf("boot-error=%t failing=%t good=%t", bootErr != nil, failedTerminated, goodTerminated), nil
}

// terminateContextScenario verifies the close context reaches provider Terminate.
func terminateContextScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	received := ""
	canceled := true
	provider := terminableProvider{
		id: "demo.terminate.context",
		terminate: func(ctx context.Context) error {
			received, _ = ctx.Value(contextKey{}).(string)
			canceled = ctx.Err() != nil
			return nil
		},
	}
	if err := app.RegisterProvider(provider); err != nil {
		return "", fmt.Errorf("register terminate-context provider: %w", err)
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot terminate-context application: %w", err)
	}
	ctx := context.WithValue(context.Background(), contextKey{}, "shutdown")
	if err := app.CloseContext(ctx); err != nil {
		return "", fmt.Errorf("close terminate-context application: %w", err)
	}
	return fmt.Sprintf("value=%s canceled=%t", received, canceled), nil
}

// closerOrderScenario verifies container closers run after provider termination.
func closerOrderScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	order := make([]string, 0, 2)
	provider := terminableProvider{
		id:        "demo.closer.order",
		terminate: func(context.Context) error { order = append(order, "terminate"); return nil },
	}
	if err := app.RegisterProvider(provider); err != nil {
		return "", fmt.Errorf("register closer-order provider: %w", err)
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot closer-order application: %w", err)
	}
	if err := app.Container().Instance("demo.closer.probe", &closerProbe{}, container.WithCloser(func(*closerProbe) error {
		order = append(order, "closer")
		return nil
	})); err != nil {
		return "", fmt.Errorf("bind closer probe: %w", err)
	}
	if err := app.Close(); err != nil {
		return "", fmt.Errorf("close closer-order application: %w", err)
	}
	return "order=" + strings.Join(order, ","), nil
}

// fullLifecycleScenario verifies a deferred, terminable provider runs the full lifecycle.
func fullLifecycleScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	registers, boots, terminates := 0, 0, 0
	provider := deferredTerminableProvider{
		deferredProvider: deferredProvider{
			id:   "demo.full",
			keys: []string{"demo.full.service"},
			register: func(app providercontract.Application) error {
				registers++
				return bindDemoValue(app, "demo.full.service", "full")
			},
			boot: func(providercontract.Application) error { boots++; return nil },
		},
		terminate: func(context.Context) error { terminates++; return nil },
	}
	if err := app.RegisterProvider(provider); err != nil {
		return "", fmt.Errorf("register full-lifecycle provider: %w", err)
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot full-lifecycle application: %w", err)
	}
	if _, err := app.Container().Make("demo.full.service"); err != nil {
		return "", fmt.Errorf("resolve full-lifecycle service: %w", err)
	}
	if err := app.Close(); err != nil {
		return "", fmt.Errorf("close full-lifecycle application: %w", err)
	}
	return fmt.Sprintf("register=%d boot=%d terminate=%d", registers, boots, terminates), nil
}
