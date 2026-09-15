package providerdemo

import (
	"context"
	"fmt"
	"sync"

	"github.com/prismgo/framework/console"
	containercontract "github.com/prismgo/framework/contracts/container"
	eventcontract "github.com/prismgo/framework/contracts/event"
	providercontract "github.com/prismgo/framework/contracts/provider"
	"github.com/prismgo/framework/event"
	"github.com/prismgo/framework/foundation"
	providerpkg "github.com/prismgo/framework/provider"
)

// containerApplication adapts a bare container into the minimal Application view
// providers receive, so binding scenarios stay independent from the framework app.
type containerApplication struct {
	c containercontract.Container
}

// Container returns the wrapped container.
func (a containerApplication) Container() containercontract.Container { return a.c }

// stubProvider is a controllable provider used to observe binding, ordering and identity.
//
// An empty id makes Name() return "", which exercises the implicit Go type identity path.
type stubProvider struct {
	id       string
	register func(providercontract.Application) error
	boot     func(providercontract.Application) error
}

// Name returns the stable provider identity, or an empty string for implicit identity.
func (p stubProvider) Name() string { return p.id }

// Register runs the optional register hook.
func (p stubProvider) Register(app providercontract.Application) error {
	if p.register == nil {
		return nil
	}
	return p.register(app)
}

// Boot runs the optional boot hook.
func (p stubProvider) Boot(app providercontract.Application) error {
	if p.boot == nil {
		return nil
	}
	return p.boot(app)
}

// deferredProvider binds its services only when one of its declared keys is first resolved.
type deferredProvider struct {
	id       string
	keys     []string
	register func(providercontract.Application) error
	boot     func(providercontract.Application) error
}

// Name returns the stable provider identity.
func (p deferredProvider) Name() string { return p.id }

// Provides declares the deferred service keys handled by this provider.
func (p deferredProvider) Provides() []string { return p.keys }

// Register runs the optional register hook when the provider is first loaded.
func (p deferredProvider) Register(app providercontract.Application) error {
	if p.register == nil {
		return nil
	}
	return p.register(app)
}

// Boot runs the optional boot hook when the provider is first loaded.
func (p deferredProvider) Boot(app providercontract.Application) error {
	if p.boot == nil {
		return nil
	}
	return p.boot(app)
}

// deferredTerminableProvider combines deferred loading with a shutdown hook.
type deferredTerminableProvider struct {
	deferredProvider
	terminate func(context.Context) error
}

// Terminate runs the optional shutdown hook.
func (p deferredTerminableProvider) Terminate(ctx context.Context) error {
	if p.terminate == nil {
		return nil
	}
	return p.terminate(ctx)
}

// terminableProvider records provider shutdown behavior and optional register/boot errors.
type terminableProvider struct {
	id        string
	register  func(providercontract.Application) error
	boot      func(providercontract.Application) error
	terminate func(context.Context) error
}

// Name returns the stable provider identity.
func (p terminableProvider) Name() string { return p.id }

// Register runs the optional register hook.
func (p terminableProvider) Register(app providercontract.Application) error {
	if p.register == nil {
		return nil
	}
	return p.register(app)
}

// Boot runs the optional boot hook.
func (p terminableProvider) Boot(app providercontract.Application) error {
	if p.boot == nil {
		return nil
	}
	return p.boot(app)
}

// Terminate runs the optional shutdown hook.
func (p terminableProvider) Terminate(ctx context.Context) error {
	if p.terminate == nil {
		return nil
	}
	return p.terminate(ctx)
}

// mountCommandProvider declares a console command from provider Boot.
type mountCommandProvider struct {
	factory console.CommandFactory
}

// Name returns the stable provider identity.
func (mountCommandProvider) Name() string { return "demo.command.mount" }

// Register performs no binding.
func (mountCommandProvider) Register(providercontract.Application) error { return nil }

// Boot declares the console command through the deferred starting registrar.
func (p mountCommandProvider) Boot(providercontract.Application) error {
	return providerpkg.Commands(p.factory)
}

// probeCommand is an inline console command used by command declaration scenarios.
type probeCommand struct {
	name   string
	handle func(console.CommandContext) error
}

// Definition returns the command definition.
func (c probeCommand) Definition() *console.Definition {
	return &console.Definition{Name: c.name, Description: "Provider demo probe"}
}

// Handle delegates to the optional handler.
func (c probeCommand) Handle(ctx console.CommandContext) error {
	if c.handle == nil {
		return nil
	}
	return c.handle(ctx)
}

// workerPool is a minimal long-running resource started from Boot and drained on shutdown.
type workerPool struct {
	running bool
	drained bool
}

// Start marks the pool as running.
func (p *workerPool) Start() { p.running = true }

// Drain stops the pool, reporting the shutdown context state.
func (p *workerPool) Drain(ctx context.Context) error {
	p.drained = ctx.Err() == nil
	p.running = false
	return nil
}

// closerProbe is an empty container resource whose closer records shutdown order.
type closerProbe struct{}

// contextKey is a private key type for close-context propagation scenarios.
type contextKey struct{}

// providerPhaseRecorder records provider lifecycle events in dispatch order.
type providerPhaseRecorder struct {
	mu      sync.Mutex
	entries []string
}

// add appends one identity to the recorder.
func (r *providerPhaseRecorder) add(entry string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entry)
}

// values returns a copy of the recorded identities.
func (r *providerPhaseRecorder) values() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.entries...)
}

// recordProviderPhase subscribes a recorder for one provider lifecycle phase.
func recordProviderPhase(bus eventcontract.Dispatcher, phase string, recorder *providerPhaseRecorder) {
	bus.Listen("app.provider.*", event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		gotPhase, provider := providerEventInfo(ev)
		if gotPhase == phase && provider != "" {
			recorder.add(provider)
		}
		return nil
	}))
}

// providerEventInfo unpacks a provider lifecycle event into its phase and identity.
func providerEventInfo(ev event.Event) (phase, provider string) {
	switch value := ev.(type) {
	case event.ProviderRegistering:
		return "registering", value.Provider
	case event.ProviderRegistered:
		return "registered", value.Provider
	case event.ProviderBooting:
		return "booting", value.Provider
	case event.ProviderBooted:
		return "booted", value.Provider
	default:
		return "", ""
	}
}

// eventDispatcher resolves the event bus owned by the application container.
func eventDispatcher(app *foundation.Application) (eventcontract.Dispatcher, error) {
	raw, err := app.Container().Make(foundation.ContainerKeyEventDispatcher)
	if err != nil {
		return nil, err
	}
	bus, ok := raw.(eventcontract.Dispatcher)
	if !ok {
		return nil, fmt.Errorf("provider demo: %s has type %T, want event dispatcher", foundation.ContainerKeyEventDispatcher, raw)
	}
	return bus, nil
}

// contains reports whether values contains wanted.
func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

// bindDemoValue registers a lazy singleton through the provider Application container.
//
// Deferred providers must register bindings instead of constructed instances so the
// container can resolve the service immediately after the provider loads.
func bindDemoValue(app providercontract.Application, key, value string) error {
	return app.Container().Singleton(key, func(containercontract.Resolver) (any, error) {
		return value, nil
	})
}
