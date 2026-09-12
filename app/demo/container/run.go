// Package containerdemo contains runnable examples of container binding and resolution.
package containerdemo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/prismgo/framework/container"
	containercontract "github.com/prismgo/framework/contracts/container"
	"github.com/prismgo/framework/foundation"
)

// Result records the observable behavior of a container example.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

type service struct{ name string }

// Run executes one documentation-backed container example.
func Run(name string) (Result, error) {
	value, err := run(name)
	if err != nil {
		return Result{}, fmt.Errorf("container demo %s: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

func run(name string) (string, error) {
	switch name {
	case "architecture":
		return architecture()
	case "dependency-resolution":
		return dependencyResolution()
	case "use-cases":
		return useCases()
	case "bind":
		return bindThroughApplication()
	case "transient-lifecycle":
		return transientLifecycle()
	case "singleton":
		return singleton()
	case "singleton-retry":
		return singletonRetry()
	case "instance":
		return instance()
	case "nil-instance":
		return nilInstance()
	case "alias":
		return alias()
	case "alias-close-order":
		return aliasCloseOrder()
	case "with-closer":
		return withCloser()
	case "with-context-closer":
		return withContextCloser()
	case "with-close-group":
		return withCloseGroup()
	case "make":
		return makeService()
	case "make-order":
		return makeOrder()
	case "factory":
		return factory()
	case "typed-make":
		return typedMake()
	case "typed-mismatch":
		return typedMismatch()
	case "value":
		return value()
	case "value-zero":
		return valueZero()
	case "call":
		return call()
	case "call-positional":
		return callPositional()
	case "call-results":
		return callResults()
	case "call-limits":
		return callLimits()
	case "has":
		return has()
	case "bound":
		return bound()
	case "resolved":
		return resolved()
	case "list":
		return list()
	case "forget":
		return forget()
	case "forget-no-close":
		return forgetNoClose()
	case "close-groups":
		return closeGroups()
	case "close-group":
		return closeGroup()
	case "close-order":
		return closeOrder()
	case "closer-ownership":
		return closerOwnership()
	case "close-pre-cancel":
		return closePreCancel()
	case "close-success":
		return closeSuccess()
	case "close-retry":
		return closeRetry()
	case "close-mid-cancel":
		return closeMidCancel()
	case "missing-loader":
		return missingLoader()
	case "missing-loader-error":
		return missingLoaderError()
	case "facade":
		return facade()
	case "provider-lifecycle":
		return providerLifecycle()
	case "error-not-registered":
		return errorNotRegistered()
	case "error-nil-result":
		return errorNilResult()
	case "error-no-current":
		return errorNoCurrent()
	case "laravel-mapping":
		return laravelMapping()
	default:
		return "", fmt.Errorf("unknown scenario %q", name)
	}
}

func architecture() (string, error) {
	c := container.NewContainer()
	if err := c.Instance("podcast.client", &service{name: "client"}); err != nil {
		return "", err
	}
	if err := c.Singleton("podcast.service", func(r containercontract.Resolver) (any, error) {
		client, err := r.Make("podcast.client")
		if err != nil {
			return nil, fmt.Errorf("resolve podcast client: %w", err)
		}
		return &service{name: client.(*service).name + "+podcast"}, nil
	}); err != nil {
		return "", err
	}
	raw, err := c.Make("podcast.service")
	if err != nil {
		return "", err
	}
	return raw.(*service).name, nil
}

func dependencyResolution() (string, error) {
	c := container.NewContainer()
	if err := c.Instance("parser.source", &service{name: "feed"}); err != nil {
		return "", err
	}
	if err := c.Bind("parser", func(r containercontract.Resolver) (any, error) {
		source, err := r.Make("parser.source")
		if err != nil {
			return nil, fmt.Errorf("resolve parser source: %w", err)
		}
		return &service{name: "parsed:" + source.(*service).name}, nil
	}); err != nil {
		return "", err
	}
	raw, err := c.Make("parser")
	if err != nil {
		return "", err
	}
	return raw.(*service).name, nil
}

func useCases() (string, error) {
	c := container.NewContainer()
	closed := false
	if err := c.Instance("shared.client", &service{name: "shared"}, container.WithCloser(func(*service) error {
		closed = true
		return nil
	})); err != nil {
		return "", err
	}
	if err := c.Close(context.Background()); err != nil {
		return "", err
	}
	return fmt.Sprintf("managed client closed=%t", closed), nil
}

func bindThroughApplication() (value string, err error) {
	app := foundation.NewApplication()
	defer func() {
		if closeErr := app.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close application: %w", closeErr))
		}
	}()
	created := 0
	if err := app.Bind("demo.generator", func(containercontract.Resolver) (any, error) {
		created++
		return &service{name: fmt.Sprintf("generator-%d", created)}, nil
	}); err != nil {
		return "", err
	}
	first, err := app.Make("demo.generator")
	if err != nil {
		return "", err
	}
	second, err := app.Make("demo.generator")
	if err != nil {
		return "", err
	}
	return first.(*service).name + "," + second.(*service).name, nil
}

func transientLifecycle() (string, error) {
	c := container.NewContainer()
	created, closed := 0, 0
	if err := c.Bind("transient", func(containercontract.Resolver) (any, error) {
		created++
		return &service{name: fmt.Sprintf("%d", created)}, nil
	}, container.WithCloser(func(*service) error {
		closed++
		return nil
	})); err != nil {
		return "", err
	}
	for range 2 {
		if _, err := c.Make("transient"); err != nil {
			return "", err
		}
	}
	if err := c.Close(context.Background()); err != nil {
		return "", err
	}
	return fmt.Sprintf("created=%d closed=%d", created, closed), nil
}

func singleton() (string, error) {
	c := container.NewContainer()
	created := 0
	if err := c.Singleton("shared", func(containercontract.Resolver) (any, error) {
		created++
		return &service{name: "shared"}, nil
	}); err != nil {
		return "", err
	}
	first, err := c.Make("shared")
	if err != nil {
		return "", err
	}
	second, err := c.Make("shared")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("same=%t created=%d", first == second, created), nil
}

func singletonRetry() (string, error) {
	c := container.NewContainer()
	attempts := 0
	wantFailure := errors.New("dependency unavailable")
	if err := c.Singleton("retry", func(containercontract.Resolver) (any, error) {
		attempts++
		if attempts == 1 {
			return nil, wantFailure
		}
		return &service{name: "ready"}, nil
	}); err != nil {
		return "", err
	}
	_, firstErr := c.Make("retry")
	if !errors.Is(firstErr, wantFailure) {
		return "", fmt.Errorf("first resolution error = %v, want %v", firstErr, wantFailure)
	}
	raw, err := c.Make("retry")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s after %d attempts", raw.(*service).name, attempts), nil
}

func instance() (string, error) {
	c := container.NewContainer()
	original := &service{name: "configured"}
	if err := c.Instance("settings", original); err != nil {
		return "", err
	}
	raw, err := c.Make("settings")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("same=%t resolved=%t", raw == original, c.Resolved("settings")), nil
}

func nilInstance() (string, error) {
	c := container.NewContainer()
	if err := c.Instance("optional", nil); err != nil {
		return "", err
	}
	before := c.Bound("optional")
	if err := c.Instance("optional", &service{name: "replaced"}); err != nil {
		return "", err
	}
	return fmt.Sprintf("bound before=%t after=%t", before, c.Bound("optional")), nil
}

func alias() (string, error) {
	c := container.NewContainer()
	if err := c.Instance("cache.manager", &service{name: "cache"}); err != nil {
		return "", err
	}
	if err := c.Alias("cache.manager", "cache"); err != nil {
		return "", err
	}
	canonical, err := c.Make("cache.manager")
	if err != nil {
		return "", err
	}
	aliased, err := c.Make("cache")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("same=%t bound=%t resolved=%t", canonical == aliased, c.Bound("cache"), c.Resolved("cache")), nil
}

func aliasCloseOrder() (string, error) {
	c := container.NewContainer()
	var order []string
	for _, key := range []string{"first", "second"} {
		if err := c.Instance(key, &service{name: key}, container.WithCloser(func(s *service) error {
			order = append(order, s.name)
			return nil
		})); err != nil {
			return "", err
		}
	}
	if err := c.Alias("first", "alias-first"); err != nil {
		return "", err
	}
	if _, err := c.Make("alias-first"); err != nil {
		return "", err
	}
	if err := c.Close(context.Background()); err != nil {
		return "", err
	}
	return strings.Join(order, ","), nil
}

func withCloser() (string, error) {
	c := container.NewContainer()
	closed := ""
	if err := c.Singleton("resource", func(containercontract.Resolver) (any, error) {
		return &service{name: "resource"}, nil
	}, container.WithCloser(func(s *service) error {
		closed = s.name
		return nil
	})); err != nil {
		return "", err
	}
	if _, err := c.Make("resource"); err != nil {
		return "", err
	}
	if err := c.Close(context.Background()); err != nil {
		return "", err
	}
	return "closed=" + closed, nil
}

func withContextCloser() (string, error) {
	c := container.NewContainer()
	key := struct{}{}
	ctx := context.WithValue(context.Background(), key, "shutdown")
	received := ""
	if err := c.Instance("resource", &service{name: "resource"}, container.WithContextCloser(func(closeCtx context.Context, _ *service) error {
		received, _ = closeCtx.Value(key).(string)
		return nil
	})); err != nil {
		return "", err
	}
	if err := c.Close(ctx); err != nil {
		return "", err
	}
	return "context=" + received, nil
}

func withCloseGroup() (string, error) {
	c := container.NewContainer()
	closed := false
	if err := c.Instance("reporter", &service{name: "reporter"},
		container.WithCloser(func(*service) error { closed = true; return nil }),
		container.WithCloseGroup(container.CloseGroupReporting)); err != nil {
		return "", err
	}
	if err := c.CloseGroup(context.Background(), container.CloseGroupNormal); err != nil {
		return "", err
	}
	before := closed
	if err := c.CloseGroup(context.Background(), container.CloseGroupReporting); err != nil {
		return "", err
	}
	return fmt.Sprintf("normal=%t reporting=%t", before, closed), nil
}

func makeService() (string, error) {
	c := container.NewContainer()
	if err := c.Instance("service", &service{name: "made"}); err != nil {
		return "", err
	}
	raw, err := c.Make("service")
	if err != nil {
		return "", err
	}
	return raw.(*service).name, nil
}

func makeOrder() (string, error) {
	c := container.NewContainer()
	if err := c.Instance("ready", &service{name: "ready"}); err != nil {
		return "", err
	}
	first, err := c.Make("ready")
	if err != nil {
		return "", err
	}
	second, err := c.Make("ready")
	if err != nil {
		return "", err
	}
	loaded, called := 0, 0
	c.SetMissingFactoryLoader(func(key string) error {
		loaded++
		if key != "deferred" {
			return nil
		}
		return c.Singleton(key, func(containercontract.Resolver) (any, error) {
			called++
			return &service{name: "deferred"}, nil
		})
	})
	for range 2 {
		if _, err := c.Make("deferred"); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("instance=%t loaded=%d factory=%d", first == second, loaded, called), nil
}

func factory() (string, error) {
	c := container.NewContainer()
	created := 0
	if err := c.Bind("fresh", func(containercontract.Resolver) (any, error) {
		created++
		return &service{name: fmt.Sprintf("fresh-%d", created)}, nil
	}); err != nil {
		return "", err
	}
	create, err := c.Factory("fresh")
	if err != nil {
		return "", err
	}
	before := created
	first, err := create()
	if err != nil {
		return "", err
	}
	second, err := create()
	if err != nil {
		return "", err
	}
	if err := c.Singleton("shared", func(containercontract.Resolver) (any, error) {
		return &service{name: "shared"}, nil
	}); err != nil {
		return "", err
	}
	resolve, err := c.Factory("shared")
	if err != nil {
		return "", err
	}
	sharedFirst, err := resolve()
	if err != nil {
		return "", err
	}
	sharedSecond, err := resolve()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("before=%d values=%s,%s shared=%t", before, first.(*service).name, second.(*service).name, sharedFirst == sharedSecond), nil
}
