package providerdemo

import (
	"context"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"
	containercontract "github.com/prismgo/framework/contracts/container"
	eventcontract "github.com/prismgo/framework/contracts/event"
	providercontract "github.com/prismgo/framework/contracts/provider"
	"github.com/prismgo/framework/event"
	"github.com/prismgo/framework/foundation"
	providerpkg "github.com/prismgo/framework/provider"
)

// defaultProviderIDs is the documented default provider registration order.
var defaultProviderIDs = []string{
	"redis", "cache", "queue", "cookie", "session", "filesystem",
	"database", "database.schema", "ratelimit", "route",
}

// baseProviderIDs is the documented immediate base provider registration order.
var baseProviderIDs = []string{"event", "config", "logger", "translation"}

// bootOrderScenario verifies every provider registers before any provider boots.
func bootOrderScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	trace := make([]string, 0, 4)
	provider := func(id string) stubProvider {
		return stubProvider{
			id:       id,
			register: func(providercontract.Application) error { trace = append(trace, "register:"+id); return nil },
			boot:     func(providercontract.Application) error { trace = append(trace, "boot:"+id); return nil },
		}
	}
	for _, id := range []string{"demo.p1", "demo.p2"} {
		if err := app.RegisterProvider(provider(id)); err != nil {
			return "", fmt.Errorf("register boot-order provider %s: %w", id, err)
		}
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot boot-order application: %w", err)
	}
	return "order=" + strings.Join(trace, ","), nil
}

// bootListenersScenario verifies Boot can resolve services and register listeners.
func bootListenersScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	booted := false
	provider := stubProvider{
		id: "demo.listener",
		boot: func(app providercontract.Application) error {
			return listenAppBooted(app, func() { booted = true })
		},
	}
	if err := app.RegisterProvider(provider); err != nil {
		return "", fmt.Errorf("register listener provider: %w", err)
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot listener application: %w", err)
	}
	return fmt.Sprintf("listen=app.booted booted=%t", booted), nil
}

// commandsScenario verifies a provider can declare console commands from Boot.
func commandsScenario(base string) (string, error) {
	factory := func() console.Command { return probeCommand{name: "provider:demo"} }
	app := foundation.Configure(base).WithProviders(mountCommandProvider{factory: factory}).Create()
	defer app.Close()
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot command provider application: %w", err)
	}
	definitions, err := app.NewConsoleKernel().All()
	if err != nil {
		return "", fmt.Errorf("load console definitions: %w", err)
	}
	return fmt.Sprintf("mounted=%t command=provider:demo", hasCommandDefinition(definitions, "provider:demo")), nil
}

// commandInputsScenario verifies provider.Commands accepts commands and factories and rejects bad input.
func commandInputsScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	instanceErr := providerpkg.Commands(probeCommand{name: "provider:demo"})
	factory := console.CommandFactory(func() console.Command { return probeCommand{name: "provider:demo.factory"} })
	factoryErr := providerpkg.Commands(factory)
	nilErr := providerpkg.Commands(nil)
	unsupportedErr := providerpkg.Commands(42)
	return fmt.Sprintf("instance=%t factory=%t nil-error=%t unsupported-error=%t",
		instanceErr == nil, factoryErr == nil, nilErr != nil, unsupportedErr != nil), nil
}

// commandDeferredScenario verifies command mounting is deferred until the console kernel starts.
func commandDeferredScenario(base string) (string, error) {
	factory := func() console.Command { return probeCommand{name: "provider:demo"} }
	app := foundation.Configure(base).WithProviders(mountCommandProvider{factory: factory}).Create()
	defer app.Close()
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot deferred-command application: %w", err)
	}
	kernel := app.NewConsoleKernel()
	before := hasCommandDefinition(kernel.Commands(), "provider:demo")
	if _, err := kernel.All(); err != nil {
		return "", fmt.Errorf("start console kernel: %w", err)
	}
	after := hasCommandDefinition(kernel.Commands(), "provider:demo")
	return fmt.Sprintf("mounted-before=%t mounted-after=%t", before, after), nil
}

// baseOrderScenario verifies base providers bind immediately and boot in documented order.
func baseOrderScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	bound := app.Container().Bound("event.dispatcher") &&
		app.Container().Bound("config.default") &&
		app.Container().Bound("logger.manager") &&
		app.Container().Bound("translator")
	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	recorder := &providerPhaseRecorder{}
	recordProviderPhase(bus, "booting", recorder)
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot base-order application: %w", err)
	}
	order := filterIdentities(recorder.values(), baseProviderIDs)
	return fmt.Sprintf("bound=%t order=%s", bound, strings.Join(order, ",")), nil
}

// defaultOrderScenario verifies automatic default provider registration order.
func defaultOrderScenario(base string) (string, error) {
	app := foundation.Configure(base).Create()
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	recorder := &providerPhaseRecorder{}
	recordProviderPhase(bus, "registering", recorder)
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot default-order application: %w", err)
	}
	order := filterIdentities(recorder.values(), defaultProviderIDs)
	return "order=" + strings.Join(order, ","), nil
}

// extensionOrderScenario verifies extension providers register after framework defaults.
func extensionOrderScenario(base string) (string, error) {
	alpha := stubProvider{id: "demo.ext.alpha"}
	beta := stubProvider{id: "demo.ext.beta"}
	app := foundation.Configure(base).WithExtensionProviders(alpha, beta).Create()
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	recorder := &providerPhaseRecorder{}
	recordProviderPhase(bus, "registering", recorder)
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot extension-order application: %w", err)
	}
	order := recorder.values()
	extensions := filterIdentities(order, []string{"demo.ext.alpha", "demo.ext.beta"})
	afterDefaults := len(order) >= 2 && order[len(order)-2] == "demo.ext.alpha" && order[len(order)-1] == "demo.ext.beta"
	return fmt.Sprintf("order=%s after-defaults=%t", strings.Join(extensions, ","), afterDefaults), nil
}

// applicationOrderScenario verifies application providers register last and can override.
func applicationOrderScenario(base string) (string, error) {
	extension := stubProvider{
		id: "demo.ext",
		register: func(app providercontract.Application) error {
			return app.Container().Singleton("demo.layer", func(containercontract.Resolver) (any, error) {
				return "extension", nil
			})
		},
	}
	application := stubProvider{
		id: "demo.app",
		register: func(app providercontract.Application) error {
			return app.Container().Instance("demo.layer", "application")
		},
	}
	app := foundation.Configure(base).WithExtensionProviders(extension).WithProviders(application).Create()
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	recorder := &providerPhaseRecorder{}
	recordProviderPhase(bus, "registering", recorder)
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot application-order application: %w", err)
	}
	order := recorder.values()
	positioned := len(order) >= 2 && order[len(order)-2] == "demo.ext" && order[len(order)-1] == "demo.app"
	if !positioned {
		return "", fmt.Errorf("application provider order = %v, want extension then application", order)
	}
	raw, err := app.Container().Make("demo.layer")
	if err != nil {
		return "", fmt.Errorf("resolve overridden layer: %w", err)
	}
	return fmt.Sprintf("order=demo.ext,demo.app override=%v", raw), nil
}

// namedIdentityScenario verifies an explicit Name() becomes the provider identity.
func namedIdentityScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	recorder := &providerPhaseRecorder{}
	recordProviderPhase(bus, "registering", recorder)
	if err := app.RegisterProvider(stubProvider{id: "demo.named"}); err != nil {
		return "", fmt.Errorf("register named provider: %w", err)
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot named-identity application: %w", err)
	}
	identities := recorder.values()
	if len(identities) != 1 {
		return "", fmt.Errorf("named identity events = %v, want exactly one", identities)
	}
	return "identity=" + identities[0], nil
}

// implicitIdentityScenario verifies a provider without Name() uses its Go type path.
func implicitIdentityScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	recorder := &providerPhaseRecorder{}
	recordProviderPhase(bus, "registering", recorder)
	if err := app.RegisterProvider(stubProvider{}); err != nil {
		return "", fmt.Errorf("register anonymous provider: %w", err)
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot implicit-identity application: %w", err)
	}
	identities := recorder.values()
	if len(identities) != 1 {
		return "", fmt.Errorf("implicit identity events = %v, want exactly one", identities)
	}
	return "identity=" + identities[0], nil
}

// duplicateIdentityScenario verifies a repeated identity reuses the first provider instance.
func duplicateIdentityScenario(base string) (string, error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	registers := 0
	first := stubProvider{
		id: "demo.dup",
		register: func(app providercontract.Application) error {
			registers++
			return app.Container().Instance("demo.dup.value", "first")
		},
	}
	second := stubProvider{
		id: "demo.dup",
		register: func(app providercontract.Application) error {
			registers++
			return app.Container().Instance("demo.dup.value", "second")
		},
	}
	if err := app.RegisterProvider(first); err != nil {
		return "", fmt.Errorf("register first duplicate provider: %w", err)
	}
	if err := app.RegisterProvider(second); err != nil {
		return "", fmt.Errorf("register second duplicate provider: %w", err)
	}
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot duplicate-identity application: %w", err)
	}
	raw, err := app.Container().Make("demo.dup.value")
	if err != nil {
		return "", fmt.Errorf("resolve duplicate value: %w", err)
	}
	return fmt.Sprintf("registers=%d value=%v", registers, raw), nil
}

// listenAppBooted resolves the application event dispatcher and subscribes to app.booted.
func listenAppBooted(app providercontract.Application, onBooted func()) error {
	raw, err := app.Container().Make(foundation.ContainerKeyEventDispatcher)
	if err != nil {
		return fmt.Errorf("resolve event dispatcher: %w", err)
	}
	bus, ok := raw.(eventcontract.Dispatcher)
	if !ok {
		return fmt.Errorf("event dispatcher has type %T", raw)
	}
	bus.Listen(event.EventAppBooted, event.ListenerFunc(func(context.Context, event.Event) error {
		onBooted()
		return nil
	}))
	return nil
}

// hasCommandDefinition reports whether a definition snapshot contains the command name.
func hasCommandDefinition(definitions []console.Definition, name string) bool {
	for _, definition := range definitions {
		if definition.Name == name {
			return true
		}
	}
	return false
}

// filterIdentities returns recorded identities that appear in allowed, in recorded order.
func filterIdentities(recorded, allowed []string) []string {
	filtered := make([]string, 0, len(recorded))
	for _, identity := range recorded {
		if contains(allowed, identity) {
			filtered = append(filtered, identity)
		}
	}
	return filtered
}
