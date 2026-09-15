package lifecycledemo

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	containercontract "github.com/prismgo/framework/contracts/container"
	providercontract "github.com/prismgo/framework/contracts/provider"
	"github.com/prismgo/framework/event"
	"github.com/prismgo/framework/foundation"
)

// applicationEntryScenario asserts the documented construction contracts at compile time.
func applicationEntryScenario() (string, error) {
	var (
		_ func(...string) *foundation.Builder                                                  = foundation.Configure
		_ func(...string) *foundation.Application                                              = foundation.NewApplication
		_ func(*foundation.Application, context.Context, []string) error                       = (*foundation.Application).HandleCommand
		_ func(*foundation.Application, func(context.Context) error, ...context.Context) error = (*foundation.Application).RunContext
		_ func(*foundation.Application, context.Context, string) (*http.Server, error)         = (*foundation.Application).NewHTTPServer
	)
	return "entry=NewApplication builder=Configure command=HandleCommand runner=RunContext server=NewHTTPServer", nil
}

// runContextScenario verifies Boot, runner execution and Close orchestration.
func runContextScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	ran := false
	if err := app.RunContext(func(context.Context) error {
		ran = true
		return nil
	}); err != nil {
		return "", fmt.Errorf("run context lifecycle: %w", err)
	}
	return fmt.Sprintf("runner-ran=%t context-canceled=%t", ran, app.Context().Err() != nil), nil
}

// baseProvidersScenario verifies event/config/logger/translation register before Boot.
func baseProvidersScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	c := app.Container()
	return fmt.Sprintf("event=%t config=%t logger=%t translation=%t",
		c.Bound(foundation.ContainerKeyEventDispatcher),
		c.Bound("config.default"),
		c.Bound("logger.manager"),
		c.Bound("translator"),
	), nil
}

// defaultProvidersScenario verifies default provider order and deferred registration.
func defaultProvidersScenario(base string) (string, error) {
	app := buildApplication(base, nil, nil)
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	log := &eventLog{}
	recordProviderPhase(bus, log, "registering")

	boundBefore := app.Container().Bound("cache.manager")
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot default providers: %w", err)
	}
	boundAfter := app.Container().Bound("cache.manager")

	order := strings.Join(log.values(), ",")
	if order != defaultProviderOrder {
		return "", fmt.Errorf("default provider register order = %q, want %q", order, defaultProviderOrder)
	}
	return fmt.Sprintf("order=%s bound-before=%t bound-after=%t", order, boundBefore, boundAfter), nil
}

// extensionProvidersScenario verifies extensions run after defaults and bind lazily.
func extensionProvidersScenario(base string) (string, error) {
	var lazyCalls int32
	extension := markerProvider{
		id: "demo.extension",
		register: func(app providercontract.Application) error {
			return app.Container().Singleton("demo.extension.manager", func(containercontract.Resolver) (any, error) {
				atomic.AddInt32(&lazyCalls, 1)
				return "extension", nil
			})
		},
	}
	app := buildApplication(base, []providercontract.ServiceProvider{extension}, nil)
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	log := &eventLog{}
	recordProviderPhase(bus, log, "registering")
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot extension application: %w", err)
	}

	order := log.values()
	position := "after-defaults"
	if len(order) == 0 || order[len(order)-1] != "demo.extension" {
		position = "unexpected"
	}
	lazyAfterBoot := atomic.LoadInt32(&lazyCalls) == 0
	if _, err := app.Container().Make("demo.extension.manager"); err != nil {
		return "", fmt.Errorf("resolve extension service: %w", err)
	}
	lazyResolved := atomic.LoadInt32(&lazyCalls) == 1
	return fmt.Sprintf("position=%s lazy-after-boot=%t lazy-resolved=%t", position, lazyAfterBoot, lazyResolved), nil
}

// applicationProvidersScenario verifies application providers run last and can override.
func applicationProvidersScenario(base string) (string, error) {
	extension := markerProvider{
		id: "demo.extension",
		register: func(app providercontract.Application) error {
			return app.Container().Singleton("demo.layer", func(containercontract.Resolver) (any, error) {
				return "extension", nil
			})
		},
	}
	application := markerProvider{
		id: "demo.application",
		register: func(app providercontract.Application) error {
			return app.Container().Singleton("demo.layer", func(containercontract.Resolver) (any, error) {
				return "application", nil
			})
		},
	}
	app := buildApplication(base, []providercontract.ServiceProvider{extension}, []providercontract.ServiceProvider{application})
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	log := &eventLog{}
	recordProviderPhase(bus, log, "registering")
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot application providers: %w", err)
	}
	order := log.values()
	if len(order) < 2 || order[len(order)-2] != "demo.extension" || order[len(order)-1] != "demo.application" {
		return "", fmt.Errorf("application provider order = %v, want extension then application", order)
	}
	raw, err := app.Container().Make("demo.layer")
	if err != nil {
		return "", fmt.Errorf("resolve overridden binding: %w", err)
	}
	return fmt.Sprintf("order=demo.extension,demo.application override=%v", raw), nil
}

// exceptionHandlerScenario verifies the default exception handler construction and binding.
func exceptionHandlerScenario(base string) (string, error) {
	app := buildApplication(base, nil, nil)
	defer app.Close()

	raw, err := app.Container().Make(foundation.ContainerKeyExceptionHandler)
	if err != nil {
		return "", fmt.Errorf("resolve exception handler: %w", err)
	}
	return fmt.Sprintf("bound=%t handler=%T", app.Container().Bound(foundation.ContainerKeyExceptionHandler), raw), nil
}

// providerLayeringScenario verifies framework, extension and application layering.
func providerLayeringScenario(base string) (string, error) {
	extension := markerProvider{id: "demo.extension"}
	application := markerProvider{id: "demo.application"}
	app := buildApplication(base, []providercontract.ServiceProvider{extension}, []providercontract.ServiceProvider{application})
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	log := &eventLog{}
	recordProviderPhase(bus, log, "registering")
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot layered application: %w", err)
	}
	order := log.values()
	if len(order) != 12 || order[10] != "demo.extension" || order[11] != "demo.application" {
		return "", fmt.Errorf("layered provider order = %v, want 10 defaults then extension then application", order)
	}
	return "framework=10 extension=1 application=1 layered=true", nil
}

// registerPhaseScenario verifies every Register runs before any Boot and only binds.
func registerPhaseScenario(base string) (string, error) {
	var lazyCalls int32
	application := markerProvider{
		id: "demo.application",
		register: func(app providercontract.Application) error {
			return app.Container().Singleton("demo.lazy", func(containercontract.Resolver) (any, error) {
				atomic.AddInt32(&lazyCalls, 1)
				return "lazy", nil
			})
		},
	}
	app := buildApplication(base, nil, []providercontract.ServiceProvider{application})
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	log := &eventLog{}
	bus.Listen("app.provider.*", event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		phase, provider := providerEventInfo(ev)
		if phase != "" {
			log.add(phase + ":" + provider)
		}
		return nil
	}))
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot register-phase application: %w", err)
	}

	lastRegister, firstBoot := -1, -1
	for index, entry := range log.values() {
		if strings.HasPrefix(entry, "registering:") {
			lastRegister = index
		}
		if strings.HasPrefix(entry, "booting:") && firstBoot == -1 {
			firstBoot = index
		}
	}
	registersFirst := lastRegister >= 0 && firstBoot > lastRegister
	return fmt.Sprintf("registers-before-boots=%t binding-only=%t", registersFirst, atomic.LoadInt32(&lazyCalls) == 0), nil
}

// bootPhaseScenario verifies the default providers boot in repository order.
func bootPhaseScenario(base string) (string, error) {
	app := buildApplication(base, nil, nil)
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	log := &eventLog{}
	bus.Listen("app.provider.*", event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		phase, provider := providerEventInfo(ev)
		if phase == "booting" && strings.Contains(","+defaultProviderOrder+",", ","+provider+",") {
			log.add(provider)
		}
		return nil
	}))
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot default-provider application: %w", err)
	}
	order := strings.Join(log.values(), ",")
	if order != defaultProviderOrder {
		return "", fmt.Errorf("default provider boot order = %q, want %q", order, defaultProviderOrder)
	}
	return fmt.Sprintf("order=%s", order), nil
}

// bootedRunnerOrderScenario verifies app.booted dispatches before the RunContext runner.
func bootedRunnerOrderScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	log := &eventLog{}
	bus.Listen(event.EventAppBooted, event.ListenerFunc(func(context.Context, event.Event) error {
		log.add("booted")
		return nil
	}))
	if err := app.RunContext(func(context.Context) error {
		log.add("runner")
		return nil
	}); err != nil {
		return "", fmt.Errorf("run booted-runner lifecycle: %w", err)
	}
	return "order=" + strings.Join(log.values(), ","), nil
}

// deferredProviderMapScenario exercises the DeferrableProvider contract at compile time.
func deferredProviderMapScenario() (string, error) {
	var contract providercontract.DeferrableProvider = deferredProvider{key: "demo.deferred"}
	keys := contract.Provides()
	return fmt.Sprintf("interface=DeferrableProvider keys=%s", strings.Join(keys, ",")), nil
}

// deferredProviderResolutionScenario verifies a deferred provider loads on first resolution.
func deferredProviderResolutionScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	var registers, boots int32
	provider := deferredProvider{key: "demo.deferred", registers: &registers, boots: &boots}
	if err := app.RegisterProvider(provider); err != nil {
		return "", fmt.Errorf("register de deferred provider: %w", err)
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot deferred application: %w", err)
	}
	boundBefore := app.Container().Bound("demo.deferred")
	resolved, err := app.Container().Make("demo.deferred")
	if err != nil {
		return "", fmt.Errorf("resolve deferred service: %w", err)
	}
	return fmt.Sprintf("bound-before=%t register=%d boot=%d value=%v",
		boundBefore, atomic.LoadInt32(&registers), atomic.LoadInt32(&boots), resolved), nil
}

// dynamicProviderScenario verifies RegisterProvider runs Register and Boot after boot.
func dynamicProviderScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot dynamic application: %w", err)
	}

	var registered, booted bool
	provider := markerProvider{
		id:       "demo.dynamic",
		register: func(providercontract.Application) error { registered = true; return nil },
		boot:     func(providercontract.Application) error { booted = true; return nil },
	}
	if err := app.RegisterProvider(provider); err != nil {
		return "", fmt.Errorf("register dynamic provider: %w", err)
	}
	return fmt.Sprintf("registered=%t booted=%t", registered, booted), nil
}
