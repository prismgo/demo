package routedemo

import (
	"fmt"

	"github.com/prismgo/framework/container"
	containercontract "github.com/prismgo/framework/contracts/container"
	"github.com/prismgo/framework/route"
)

// demoProviderApp adapts a bare container to the minimal provider.Application contract.
type demoProviderApp struct {
	registry containercontract.Container
}

// Container returns the container the provider registers into.
func (a demoProviderApp) Container() containercontract.Container { return a.registry }

// providerRegistrationScenario proves Register is lazy and only Make builds the Router.
func providerRegistrationScenario() (string, error) {
	registry := container.NewContainer()
	provider := route.ServiceProvider{}
	if err := provider.Register(demoProviderApp{registry: registry}); err != nil {
		return "", fmt.Errorf("register route provider: %w", err)
	}

	bound := registry.Bound("route.router")
	lazy := !registry.Resolved("route.router")
	raw, err := registry.Make("route.router")
	if err != nil {
		return "", fmt.Errorf("resolve route.router: %w", err)
	}
	_, ok := raw.(*route.Router)
	return fmt.Sprintf("name=%s bound=%t lazy=%t router=%t", provider.Name(), bound, lazy, ok), nil
}

// providerPreservesRouterScenario proves an explicitly injected Router is not overwritten.
func providerPreservesRouterScenario() (string, error) {
	registry := container.NewContainer()
	explicit := route.New()
	explicit.Get("/explicit", textHandler("explicit"))
	if err := registry.Instance("route.router", explicit); err != nil {
		return "", fmt.Errorf("seed explicit router: %w", err)
	}
	if err := (route.ServiceProvider{}).Register(demoProviderApp{registry: registry}); err != nil {
		return "", fmt.Errorf("register route provider: %w", err)
	}

	raw, err := registry.Make("route.router")
	if err != nil {
		return "", fmt.Errorf("resolve route.router: %w", err)
	}
	preserved, ok := raw.(*route.Router)
	if !ok {
		return "", fmt.Errorf("route.router resolved %T, want *route.Router", raw)
	}
	return fmt.Sprintf("preserved=%t routes=%d", preserved == explicit, len(preserved.List())), nil
}

// providerSingletonScenario proves repeated resolutions reuse one Router instance.
func providerSingletonScenario() (string, error) {
	registry := container.NewContainer()
	if err := (route.ServiceProvider{}).Register(demoProviderApp{registry: registry}); err != nil {
		return "", fmt.Errorf("register route provider: %w", err)
	}

	first, err := registry.Make("route.router")
	if err != nil {
		return "", fmt.Errorf("first resolve: %w", err)
	}
	second, err := registry.Make("route.router")
	if err != nil {
		return "", fmt.Errorf("second resolve: %w", err)
	}
	return fmt.Sprintf("singleton=%t resolved=%t", first == second, registry.Resolved("route.router")), nil
}
