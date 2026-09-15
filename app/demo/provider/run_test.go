package providerdemo_test

import (
	"strings"
	"testing"

	providerdemo "prismgo-demo/app/demo/provider"

	"prismgo-demo/app/demo/catalog"
)

// expectValue executes one scenario and asserts its full observable value.
func expectValue(t *testing.T, name string, want string) {
	t.Helper()
	result, err := providerdemo.Run(name)
	if err != nil {
		t.Fatalf("provider demo %q error = %v, want nil", name, err)
	}
	if result.Case != name {
		t.Fatalf("provider demo %q case = %q, want %q", name, result.Case, name)
	}
	if result.Value != want {
		t.Fatalf("provider demo %q value = %q, want %q", name, result.Value, want)
	}
}

func TestProviderDemoArchitecture(t *testing.T) {
	expectValue(t, "architecture", "provider=ServiceProvider application=Application phases=Register,Boot")
}

func TestProviderDemoContract(t *testing.T) {
	expectValue(t, "contract", "interfaces=ServiceProvider,NamedProvider,DeferrableProvider,TerminableProvider")
}

func TestProviderDemoRegisterContract(t *testing.T) {
	expectValue(t, "register-contract", "register=func(Application) error boot=func(Application) error")
}

func TestProviderDemoContainerAccess(t *testing.T) {
	expectValue(t, "container-access", "bound=true value=container-access")
}

func TestProviderDemoPreserveBinding(t *testing.T) {
	expectValue(t, "preserve-binding", "preserved=true value=injected")
}

func TestProviderDemoBind(t *testing.T) {
	expectValue(t, "bind", "transient=true created=2")
}

func TestProviderDemoSingleton(t *testing.T) {
	expectValue(t, "singleton", "singleton=true created=1")
}

func TestProviderDemoInstance(t *testing.T) {
	expectValue(t, "instance", "same=true resolved=true")
}

func TestProviderDemoAlias(t *testing.T) {
	expectValue(t, "alias", "same=true bound=true")
}

func TestProviderDemoWithCloser(t *testing.T) {
	expectValue(t, "with-closer", "closed=true")
}

func TestProviderDemoCloseGroup(t *testing.T) {
	expectValue(t, "close-group", "normal=false reporting=true")
}

func TestProviderDemoBootOrder(t *testing.T) {
	expectValue(t, "boot-order", "order=register:demo.p1,register:demo.p2,boot:demo.p1,boot:demo.p2")
}

func TestProviderDemoBootListeners(t *testing.T) {
	expectValue(t, "boot-listeners", "listen=app.booted booted=true")
}

func TestProviderDemoCommands(t *testing.T) {
	expectValue(t, "commands", "mounted=true command=provider:demo")
}

func TestProviderDemoCommandInputs(t *testing.T) {
	expectValue(t, "command-inputs", "instance=true factory=true nil-error=true unsupported-error=true")
}

func TestProviderDemoCommandDeferred(t *testing.T) {
	expectValue(t, "command-deferred", "mounted-before=false mounted-after=true")
}

func TestProviderDemoPublishes(t *testing.T) {
	expectValue(t, "publishes", "entries=1 tags=config copied=true")
}

func TestProviderDemoPublishTags(t *testing.T) {
	expectValue(t, "publish-tags", "providers=demo.config,demo.lang tags=config,lang filtered=1")
}

func TestProviderDemoPublishEnvironment(t *testing.T) {
	expectValue(t, "publish-environment", "available=false entries=0 copy-error=true")
}

func TestProviderDemoApplicationRegistration(t *testing.T) {
	expectValue(t, "application-registration", "builder=WithProviders extension=WithExtensionProviders defaults=DefaultProviders")
}

func TestProviderDemoBaseOrder(t *testing.T) {
	expectValue(t, "base-order", "bound=true order=event,config,logger,translation")
}

func TestProviderDemoDefaultOrder(t *testing.T) {
	expectValue(t, "default-order", "order=redis,cache,queue,cookie,session,filesystem,database,database.schema,ratelimit,route")
}

func TestProviderDemoExtensionOrder(t *testing.T) {
	expectValue(t, "extension-order", "order=demo.ext.alpha,demo.ext.beta after-defaults=true")
}

func TestProviderDemoApplicationOrder(t *testing.T) {
	expectValue(t, "application-order", "order=demo.ext,demo.app override=application")
}

func TestProviderDemoNamedIdentity(t *testing.T) {
	expectValue(t, "named-identity", "identity=demo.named")
}

func TestProviderDemoImplicitIdentity(t *testing.T) {
	expectValue(t, "implicit-identity", "identity=providerdemo.stubProvider")
}

func TestProviderDemoDuplicateIdentity(t *testing.T) {
	expectValue(t, "duplicate-identity", "registers=1 value=first")
}

func TestProviderDemoDeferrableContract(t *testing.T) {
	expectValue(t, "deferrable-contract", "interface=DeferrableProvider method=Provides keys=demo.key")
}

func TestProviderDemoProvides(t *testing.T) {
	expectValue(t, "provides", "provides=demo.deferred.first,demo.deferred.second")
}

func TestProviderDemoDeferredResolution(t *testing.T) {
	expectValue(t, "deferred-resolution", "bound-before=false register=1 boot=1 value=deferred")
}

func TestProviderDemoDeferredMapCleanup(t *testing.T) {
	expectValue(t, "deferred-map-cleanup", "loaded=true remapped=true")
}

func TestProviderDemoDeferredLateBoot(t *testing.T) {
	expectValue(t, "deferred-late-boot", "booted-before=false booted-after=true")
}

func TestProviderDemoDeferredEmpty(t *testing.T) {
	expectValue(t, "deferred-empty", "empty-error=true")
}

func TestProviderDemoDeferredConflict(t *testing.T) {
	expectValue(t, "deferred-conflict", "first-error=false second-error=true")
}

func TestProviderDemoDeferredTermination(t *testing.T) {
	expectValue(t, "deferred-termination", "unloaded=0 loaded=1")
}

func TestProviderDemoTerminableContract(t *testing.T) {
	expectValue(t, "terminable-contract", "interface=TerminableProvider method=Terminate signature=func(context.Context) error")
}

func TestProviderDemoWorkerLifecycle(t *testing.T) {
	expectValue(t, "worker-lifecycle", "started=true drained=true")
}

func TestProviderDemoTerminateOrder(t *testing.T) {
	expectValue(t, "terminate-order", "order=p3,p2,p1")
}

func TestProviderDemoTerminateEligibility(t *testing.T) {
	expectValue(t, "terminate-eligibility", "boot-error=true failing=false good=true")
}

func TestProviderDemoTerminateContext(t *testing.T) {
	expectValue(t, "terminate-context", "value=shutdown canceled=false")
}

func TestProviderDemoCloserOrder(t *testing.T) {
	expectValue(t, "closer-order", "order=terminate,closer")
}

func TestProviderDemoFullLifecycle(t *testing.T) {
	expectValue(t, "full-lifecycle", "register=1 boot=1 terminate=1")
}

func TestProviderDemoRejectsUnknownScenario(t *testing.T) {
	_, err := providerdemo.Run("unknown")
	if err == nil || !strings.Contains(err.Error(), `unknown scenario "unknown"`) {
		t.Fatalf("Run(unknown) error = %v, want unknown scenario error", err)
	}
}

func TestProviderDemoCatalogCoverage(t *testing.T) {
	entries := catalog.Filter("service-provider", "", catalog.StatusImplemented)
	if len(entries) != 42 {
		t.Fatalf("implemented service-provider entries = %d, want 42", len(entries))
	}
	for _, slug := range []string{"architecture", "singleton", "commands", "default-order", "deferred-resolution", "terminate-order", "full-lifecycle"} {
		if item, ok := catalog.Find("service-provider", slug); !ok || item.Status != catalog.StatusImplemented {
			t.Fatalf("service-provider %s entry = %#v, %v; want implemented", slug, item, ok)
		}
	}
}
