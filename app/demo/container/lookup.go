package containerdemo

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/prismgo/framework/container"
	containercontract "github.com/prismgo/framework/contracts/container"
)

func typedMake() (string, error) {
	c := container.NewContainer()
	container.SetProvider(func() *container.Container { return c })
	defer container.SetProvider(nil)
	if err := c.Instance("typed", &service{name: "typed"}); err != nil {
		return "", err
	}
	got, err := container.Make[*service]("typed")
	if err != nil {
		return "", err
	}
	return got.name, nil
}

func typedMismatch() (string, error) {
	c := container.NewContainer()
	container.SetProvider(func() *container.Container { return c })
	defer container.SetProvider(nil)
	if err := c.Instance("typed", &service{name: "typed"}); err != nil {
		return "", err
	}
	got, err := container.Make[string]("typed")
	return fmt.Sprintf("zero=%t mismatch=%t", got == "", err != nil), nil
}

func value() (string, error) {
	c := container.NewContainer()
	container.SetProvider(func() *container.Container { return c })
	defer container.SetProvider(nil)
	created := 0
	if err := c.Singleton("lazy", func(containercontract.Resolver) (any, error) {
		created++
		return &service{name: "ready"}, nil
	}); err != nil {
		return "", err
	}
	before := container.Value[*service]("lazy")
	if _, err := c.Make("lazy"); err != nil {
		return "", err
	}
	after := container.Value[*service]("lazy")
	return fmt.Sprintf("before=%t after=%s created=%d", before == nil, after.name, created), nil
}

func valueZero() (string, error) {
	c := container.NewContainer()
	container.SetProvider(func() *container.Container { return c })
	defer container.SetProvider(nil)
	if err := c.Instance("present", &service{name: "present"}); err != nil {
		return "", err
	}
	return fmt.Sprintf("missing=%t mismatch=%t", container.Value[*service]("missing") == nil, container.Value[string]("present") == ""), nil
}

func call() (string, error) {
	c := container.NewContainer()
	// Call uses the fully qualified parameter type as its lookup key.
	key := "*" + reflect.TypeOf(service{}).PkgPath() + ".service"
	if err := c.Instance(key, &service{name: "injected"}); err != nil {
		return "", err
	}
	out, err := c.Call(func(s *service) string { return s.name })
	if err != nil {
		return "", err
	}
	return out[0].(string), nil
}

func callPositional() (string, error) {
	c := container.NewContainer()
	out, err := c.Call(func(prefix string, count int) string { return fmt.Sprintf("%s-%d", prefix, count) }, "explicit", 7)
	if err != nil {
		return "", err
	}
	return out[0].(string), nil
}

func callResults() (string, error) {
	c := container.NewContainer()
	wantErr := errors.New("callback failed")
	out, err := c.Call(func() (string, error) { return "result", wantErr })
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("value=%s error=%t", out[0], errors.Is(out[1].(error), wantErr)), nil
}

func callLimits() (string, error) {
	c := container.NewContainer()
	_, err := c.Call(func(_ *service) {})
	return fmt.Sprintf("missing-typed-binding=%t", errors.Is(err, container.ErrFactoryNotRegistered)), nil
}

func has() (string, error) {
	c := container.NewContainer()
	loads := 0
	c.SetMissingFactoryLoader(func(key string) error {
		loads++
		return c.Instance(key, &service{name: "loaded"})
	})
	before := c.Bound("deferred")
	found := c.Has("deferred")
	return fmt.Sprintf("before=%t has=%t loads=%d", before, found, loads), nil
}

func bound() (string, error) {
	c := container.NewContainer()
	loads := 0
	c.SetMissingFactoryLoader(func(string) error { loads++; return nil })
	missing := c.Bound("deferred")
	if err := c.Instance("ready", &service{name: "ready"}); err != nil {
		return "", err
	}
	return fmt.Sprintf("missing=%t ready=%t loads=%d", missing, c.Bound("ready"), loads), nil
}

func resolved() (string, error) {
	c := container.NewContainer()
	if err := c.Singleton("lazy", func(containercontract.Resolver) (any, error) { return &service{name: "ready"}, nil }); err != nil {
		return "", err
	}
	before := c.Resolved("lazy")
	if _, err := c.Make("lazy"); err != nil {
		return "", err
	}
	return fmt.Sprintf("before=%t after=%t", before, c.Resolved("lazy")), nil
}

func list() (string, error) {
	c := container.NewContainer()
	if err := c.Instance("first", &service{name: "first"}, container.WithCloser(func(*service) error { return nil })); err != nil {
		return "", err
	}
	created := 0
	if err := c.Singleton("second", func(containercontract.Resolver) (any, error) { created++; return &service{name: "second"}, nil }); err != nil {
		return "", err
	}
	entries := c.List()
	return fmt.Sprintf("keys=%s,%s registered=%t,%t closable=%t created=%d", entries[0].Key, entries[1].Key, entries[0].Registered, entries[1].Registered, entries[0].Closable, created), nil
}

func forget() (string, error) {
	c := container.NewContainer()
	if err := c.Instance("replace", &service{name: "old"}); err != nil {
		return "", err
	}
	if err := c.Forget("replace"); err != nil {
		return "", err
	}
	between := c.Bound("replace")
	if err := c.Instance("replace", &service{name: "new"}); err != nil {
		return "", err
	}
	got, err := c.Make("replace")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("between=%t value=%s", between, got.(*service).name), nil
}

func forgetNoClose() (string, error) {
	c := container.NewContainer()
	closed := false
	if err := c.Instance("external", &service{name: "external"}, container.WithCloser(func(*service) error { closed = true; return nil })); err != nil {
		return "", err
	}
	if err := c.Forget("external"); err != nil {
		return "", err
	}
	if err := c.Close(context.Background()); err != nil {
		return "", err
	}
	return fmt.Sprintf("closed=%t bound=%t", closed, c.Bound("external")), nil
}
