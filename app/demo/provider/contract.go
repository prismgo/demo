package providerdemo

import (
	"context"
	"fmt"
	"strings"

	"github.com/prismgo/framework/container"
	containercontract "github.com/prismgo/framework/contracts/container"
	providercontract "github.com/prismgo/framework/contracts/provider"
	"github.com/prismgo/framework/foundation"
	providerpkg "github.com/prismgo/framework/provider"
)

// contractProvider is the minimal documented provider shape used for contract assertions.
type contractProvider struct{}

// Register performs no binding.
func (contractProvider) Register(providercontract.Application) error { return nil }

// Boot performs no work.
func (contractProvider) Boot(providercontract.Application) error { return nil }

// architectureScenario keeps the ServiceProvider and Application contracts referenced at compile time.
func architectureScenario() (string, error) {
	var provider providercontract.ServiceProvider = contractProvider{}
	var _ func(providercontract.Application) error = provider.Register
	var _ func(providercontract.Application) error = provider.Boot
	var _ providercontract.Application
	return "provider=ServiceProvider application=Application phases=Register,Boot", nil
}

// contractScenario asserts the optional provider interfaces at compile time.
func contractScenario() (string, error) {
	var (
		_ providercontract.ServiceProvider    = contractProvider{}
		_ providercontract.NamedProvider      = stubProvider{id: "demo.named"}
		_ providercontract.DeferrableProvider = deferredProvider{keys: []string{"demo.key"}}
		_ providercontract.TerminableProvider = terminableProvider{id: "demo.terminable"}
	)
	return "interfaces=ServiceProvider,NamedProvider,DeferrableProvider,TerminableProvider", nil
}

// registerContractScenario asserts the Register/Boot method signatures.
func registerContractScenario() (string, error) {
	provider := contractProvider{}
	var (
		_ func(providercontract.Application) error = provider.Register
		_ func(providercontract.Application) error = provider.Boot
	)
	return "register=func(Application) error boot=func(Application) error", nil
}

// applicationRegistrationScenario asserts the builder registration surface at compile time.
func applicationRegistrationScenario() (string, error) {
	var (
		_ func(*foundation.Builder, ...providercontract.ServiceProvider) *foundation.Builder = (*foundation.Builder).WithProviders
		_ func(*foundation.Builder, ...providercontract.ServiceProvider) *foundation.Builder = (*foundation.Builder).WithExtensionProviders
		_ func() []providerpkg.ServiceProvider                                               = providerpkg.DefaultProviders
	)
	return "builder=WithProviders extension=WithExtensionProviders defaults=DefaultProviders", nil
}

// containerAccessScenario proves a provider can bind services through the Application container.
func containerAccessScenario() (string, error) {
	c := container.NewContainer()
	provider := stubProvider{
		id: "demo.container",
		register: func(app providercontract.Application) error {
			return app.Container().Singleton("demo.container.service", func(containercontract.Resolver) (any, error) {
				return "container-access", nil
			})
		},
	}
	if err := provider.Register(containerApplication{c}); err != nil {
		return "", fmt.Errorf("register container access provider: %w", err)
	}
	raw, err := c.Make("demo.container.service")
	if err != nil {
		return "", fmt.Errorf("resolve container access service: %w", err)
	}
	return fmt.Sprintf("bound=%t value=%v", c.Bound("demo.container.service"), raw), nil
}

// preserveBindingScenario proves a provider keeps an existing container binding.
func preserveBindingScenario() (string, error) {
	c := container.NewContainer()
	if err := c.Instance("demo.preserved", "injected"); err != nil {
		return "", fmt.Errorf("seed preserved binding: %w", err)
	}
	provider := stubProvider{
		id: "demo.preserve",
		register: func(app providercontract.Application) error {
			if app.Container().Bound("demo.preserved") {
				return nil
			}
			return app.Container().Instance("demo.preserved", "overwritten")
		},
	}
	if err := provider.Register(containerApplication{c}); err != nil {
		return "", fmt.Errorf("register preserve provider: %w", err)
	}
	raw, err := c.Make("demo.preserved")
	if err != nil {
		return "", fmt.Errorf("resolve preserved service: %w", err)
	}
	return fmt.Sprintf("preserved=%t value=%v", c.Bound("demo.preserved"), raw), nil
}

// bindScenario proves a provider Bind registration is transient.
func bindScenario() (string, error) {
	c := container.NewContainer()
	created := 0
	provider := stubProvider{
		id: "demo.bind",
		register: func(app providercontract.Application) error {
			return app.Container().Bind("demo.transient", func(containercontract.Resolver) (any, error) {
				created++
				return fmt.Sprintf("transient-%d", created), nil
			})
		},
	}
	if err := provider.Register(containerApplication{c}); err != nil {
		return "", fmt.Errorf("register bind provider: %w", err)
	}
	first, err := c.Make("demo.transient")
	if err != nil {
		return "", fmt.Errorf("resolve first transient: %w", err)
	}
	second, err := c.Make("demo.transient")
	if err != nil {
		return "", fmt.Errorf("resolve second transient: %w", err)
	}
	return fmt.Sprintf("transient=%t created=%d", first != second, created), nil
}

// singletonScenario proves a provider Singleton registration is shared.
func singletonScenario() (string, error) {
	c := container.NewContainer()
	created := 0
	provider := stubProvider{
		id: "demo.singleton",
		register: func(app providercontract.Application) error {
			return app.Container().Singleton("demo.shared", func(containercontract.Resolver) (any, error) {
				created++
				return "shared", nil
			})
		},
	}
	if err := provider.Register(containerApplication{c}); err != nil {
		return "", fmt.Errorf("register singleton provider: %w", err)
	}
	first, err := c.Make("demo.shared")
	if err != nil {
		return "", fmt.Errorf("resolve first shared: %w", err)
	}
	second, err := c.Make("demo.shared")
	if err != nil {
		return "", fmt.Errorf("resolve second shared: %w", err)
	}
	return fmt.Sprintf("singleton=%t created=%d", first == second, created), nil
}

// instanceScenario proves a provider Instance registration returns the constructed value.
func instanceScenario() (string, error) {
	c := container.NewContainer()
	provider := stubProvider{
		id: "demo.instance",
		register: func(app providercontract.Application) error {
			return app.Container().Instance("demo.configured", "configured")
		},
	}
	if err := provider.Register(containerApplication{c}); err != nil {
		return "", fmt.Errorf("register instance provider: %w", err)
	}
	raw, err := c.Make("demo.configured")
	if err != nil {
		return "", fmt.Errorf("resolve configured service: %w", err)
	}
	return fmt.Sprintf("same=%t resolved=%t", raw == "configured", c.Resolved("demo.configured")), nil
}

// aliasScenario proves a provider can alias an existing binding.
func aliasScenario() (string, error) {
	c := container.NewContainer()
	provider := stubProvider{
		id: "demo.alias",
		register: func(app providercontract.Application) error {
			if err := app.Container().Instance("demo.canonical", "canonical"); err != nil {
				return err
			}
			return app.Container().Alias("demo.canonical", "demo.alias")
		},
	}
	if err := provider.Register(containerApplication{c}); err != nil {
		return "", fmt.Errorf("register alias provider: %w", err)
	}
	canonical, err := c.Make("demo.canonical")
	if err != nil {
		return "", fmt.Errorf("resolve canonical service: %w", err)
	}
	aliased, err := c.Make("demo.alias")
	if err != nil {
		return "", fmt.Errorf("resolve aliased service: %w", err)
	}
	return fmt.Sprintf("same=%t bound=%t", canonical == aliased, c.Bound("demo.alias")), nil
}

// withCloserScenario proves WithCloser runs when the container closes the shared service.
func withCloserScenario() (string, error) {
	c := container.NewContainer()
	closed := false
	if err := c.Singleton("demo.resource", func(containercontract.Resolver) (any, error) {
		return "resource", nil
	}, container.WithCloser(func(string) error {
		closed = true
		return nil
	})); err != nil {
		return "", fmt.Errorf("bind closer resource: %w", err)
	}
	if _, err := c.Make("demo.resource"); err != nil {
		return "", fmt.Errorf("resolve closer resource: %w", err)
	}
	if err := c.Close(context.Background()); err != nil {
		return "", fmt.Errorf("close closer container: %w", err)
	}
	return fmt.Sprintf("closed=%t", closed), nil
}

// closeGroupScenario proves WithCloseGroup assigns the reporting shutdown phase.
func closeGroupScenario() (string, error) {
	c := container.NewContainer()
	closed := false
	if err := c.Instance("demo.reporter", "reporter",
		container.WithCloser(func(string) error { closed = true; return nil }),
		container.WithCloseGroup(container.CloseGroupReporting)); err != nil {
		return "", fmt.Errorf("bind reporting resource: %w", err)
	}
	if err := c.CloseGroup(context.Background(), container.CloseGroupNormal); err != nil {
		return "", fmt.Errorf("close normal group: %w", err)
	}
	normalClosed := closed
	if err := c.CloseGroup(context.Background(), container.CloseGroupReporting); err != nil {
		return "", fmt.Errorf("close reporting group: %w", err)
	}
	return fmt.Sprintf("normal=%t reporting=%t", normalClosed, closed), nil
}

// deferrableContractScenario asserts the DeferrableProvider contract.
func deferrableContractScenario() (string, error) {
	var deferred providercontract.DeferrableProvider = deferredProvider{keys: []string{"demo.key"}}
	keys := deferred.Provides()
	return fmt.Sprintf("interface=DeferrableProvider method=Provides keys=%s", strings.Join(keys, ",")), nil
}

// providesScenario reports the declared deferred service keys.
func providesScenario() (string, error) {
	provider := deferredProvider{id: "demo.provides", keys: []string{"demo.deferred.first", "demo.deferred.second"}}
	return "provides=" + strings.Join(provider.Provides(), ","), nil
}

// terminableContractScenario asserts the TerminableProvider contract.
func terminableContractScenario() (string, error) {
	var terminable providercontract.TerminableProvider = terminableProvider{id: "demo.terminable"}
	var _ func(context.Context) error = terminable.Terminate
	return "interface=TerminableProvider method=Terminate signature=func(context.Context) error", nil
}
