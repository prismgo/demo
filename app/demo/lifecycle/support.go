package lifecycledemo

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"

	"github.com/prismgo/framework/console"
	containercontract "github.com/prismgo/framework/contracts/container"
	eventcontract "github.com/prismgo/framework/contracts/event"
	providercontract "github.com/prismgo/framework/contracts/provider"
	"github.com/prismgo/framework/event"
	"github.com/prismgo/framework/foundation"
)

// defaultProviderOrder is the documented framework default provider registration order.
const defaultProviderOrder = "redis,cache,queue,cookie,session,filesystem,database,database.schema,ratelimit,route"

// eventLog records lifecycle observations in dispatch order.
type eventLog struct {
	mu      sync.Mutex
	entries []string
}

// add appends one observation.
func (l *eventLog) add(entry string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, entry)
}

// values returns a copy of the recorded observations.
func (l *eventLog) values() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.entries...)
}

// eventDispatcher resolves the event bus owned by the application container.
func eventDispatcher(app *foundation.Application) (eventcontract.Dispatcher, error) {
	raw, err := app.Container().Make(foundation.ContainerKeyEventDispatcher)
	if err != nil {
		return nil, err
	}
	bus, ok := raw.(eventcontract.Dispatcher)
	if !ok {
		return nil, fmt.Errorf("lifecycle demo: event.dispatcher has type %T, want event dispatcher", raw)
	}
	return bus, nil
}

// recordProviderPhase subscribes a recorder for one provider lifecycle phase.
func recordProviderPhase(bus eventcontract.Dispatcher, log *eventLog, phase string) {
	bus.Listen("app.provider.*", event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		gotPhase, provider := providerEventInfo(ev)
		if gotPhase == phase {
			log.add(provider)
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

// markerProvider is a controllable provider used to observe ordering and overrides.
type markerProvider struct {
	id       string
	register func(providercontract.Application) error
	boot     func(providercontract.Application) error
}

// Name returns the stable provider identity.
func (p markerProvider) Name() string { return p.id }

// Register runs the optional register hook.
func (p markerProvider) Register(app providercontract.Application) error {
	if p.register == nil {
		return nil
	}
	return p.register(app)
}

// Boot runs the optional boot hook.
func (p markerProvider) Boot(app providercontract.Application) error {
	if p.boot == nil {
		return nil
	}
	return p.boot(app)
}

// terminableMarker is a provider with a Terminate hook used to observe shutdown order.
type terminableMarker struct {
	id        string
	terminate func(context.Context) error
}

// Name returns the stable provider identity.
func (p terminableMarker) Name() string { return p.id }

// Register is a no-op binding phase.
func (p terminableMarker) Register(providercontract.Application) error { return nil }

// Boot is a no-op boot phase.
func (p terminableMarker) Boot(providercontract.Application) error { return nil }

// Terminate runs the optional termination hook.
func (p terminableMarker) Terminate(ctx context.Context) error {
	if p.terminate == nil {
		return nil
	}
	return p.terminate(ctx)
}

// deferredProvider registers a lazy binding only when its service key is first resolved.
type deferredProvider struct {
	key       string
	registers *int32
	boots     *int32
}

// Name returns the stable provider identity.
func (p deferredProvider) Name() string { return "demo.deferred.provider" }

// Provides declares the deferred service key.
func (p deferredProvider) Provides() []string { return []string{p.key} }

// Register binds the deferred service; called on first resolution.
func (p deferredProvider) Register(app providercontract.Application) error {
	if p.registers != nil {
		atomic.AddInt32(p.registers, 1)
	}
	return app.Container().Singleton(p.key, func(containercontract.Resolver) (any, error) {
		return p.key, nil
	})
}

// Boot marks the deferred provider as booted; called on first resolution after boot.
func (p deferredProvider) Boot(providercontract.Application) error {
	if p.boots != nil {
		atomic.AddInt32(p.boots, 1)
	}
	return nil
}

// closeProbe is an empty container resource whose closer records shutdown behavior.
type closeProbe struct{}

// listenerProvider registers lifecycle listeners from its Boot method, matching the
// documented "listen during provider boot" pattern.
type listenerProvider struct {
	booted     *int32
	terminated *int32
}

// Name returns the stable provider identity.
func (listenerProvider) Name() string { return "demo.lifecycle.listener" }

// Register performs no binding.
func (listenerProvider) Register(providercontract.Application) error { return nil }

// Boot subscribes lifecycle listeners through the global dispatcher facade.
func (p listenerProvider) Boot(providercontract.Application) error {
	event.ListenFunc(event.EventAppBooted, func(context.Context, event.Event) error {
		atomic.AddInt32(p.booted, 1)
		return nil
	})
	event.ListenFunc(event.EventAppTerminated, func(context.Context, event.Event) error {
		atomic.AddInt32(p.terminated, 1)
		return nil
	})
	return nil
}

// providerPhaseResults boots an application with two marker providers and returns the
// observed provider lifecycle sequence together with the first target phase payload.
//
// Only the demo marker providers are recorded so the sequence stays deterministic
// regardless of the framework default provider count.
func providerPhaseResults(base, target string) (phase, provider string, sequence []string, err error) {
	app := foundation.NewApplication(base)
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", "", nil, err
	}
	log := &eventLog{}
	bus.Listen("app.provider.*", event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		currentPhase, currentProvider := providerEventInfo(ev)
		if currentProvider != "demo.p1" && currentProvider != "demo.p2" {
			return nil
		}
		log.add(currentPhase + ":" + currentProvider)
		if currentPhase == target && provider == "" {
			phase, provider = currentPhase, currentProvider
		}
		return nil
	}))

	for _, id := range []string{"demo.p1", "demo.p2"} {
		if err := app.RegisterProvider(markerProvider{id: id}); err != nil {
			return "", "", nil, fmt.Errorf("register marker provider %s: %w", id, err)
		}
	}
	if err := app.Boot(); err != nil {
		return "", "", nil, fmt.Errorf("boot provider-phase application: %w", err)
	}
	return phase, provider, log.values(), nil
}

// indexOf returns the position of entry in entries, or -1 when absent.
func indexOf(entries []string, entry string) int {
	for index, current := range entries {
		if current == entry {
			return index
		}
	}
	return -1
}

// demoCommand is an inline console command used by console lifecycle scenarios.
type demoCommand struct {
	name   string
	handle func(console.CommandContext) error
}

// Definition returns the command definition.
func (c demoCommand) Definition() *console.Definition {
	return &console.Definition{Name: c.name, Description: "Lifecycle demo probe"}
}

// Handle delegates to the optional handler.
func (c demoCommand) Handle(ctx console.CommandContext) error {
	if c.handle == nil {
		return nil
	}
	return c.handle(ctx)
}

// buildApplication assembles an Application without booting it, so callers can attach
// lifecycle listeners before the provider lifecycle starts.
func buildApplication(base string, extension []providercontract.ServiceProvider, application []providercontract.ServiceProvider) *foundation.Application {
	builder := foundation.Configure(base)
	if len(extension) > 0 {
		builder = builder.WithExtensionProviders(extension...)
	}
	if len(application) > 0 {
		builder = builder.WithProviders(application...)
	}
	return builder.Create()
}

// buildHTTPApplication assembles and boots an Application with routes, middleware and commands.
func buildHTTPApplication(
	base string,
	routes func(*foundation.Application, *gin.Engine) error,
	middleware func(*foundation.Middleware),
	commands ...console.CommandFactory,
) (*foundation.Application, error) {
	builder := foundation.Configure(base)
	if middleware != nil {
		builder = builder.WithMiddleware(middleware)
	}
	builder = builder.WithRouting(func(r *foundation.Routing) {
		if routes != nil {
			r.Routes(routes)
		}
		if len(commands) > 0 {
			r.Commands(commands...)
		}
	})
	app := builder.Create()
	if err := app.Boot(); err != nil {
		return nil, err
	}
	return app, nil
}
