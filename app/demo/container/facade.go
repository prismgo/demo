package containerdemo

import (
	"context"
	"errors"
	"fmt"

	"github.com/prismgo/framework/container"
	containercontract "github.com/prismgo/framework/contracts/container"
	"github.com/prismgo/framework/foundation"
)

func missingLoader() (string, error) {
	c := container.NewContainer()
	loads := 0
	c.SetMissingFactoryLoader(func(key string) error {
		loads++
		return c.Singleton(key, func(containercontract.Resolver) (any, error) {
			return &service{name: key}, nil
		})
	})
	made, err := c.Make("made")
	if err != nil {
		return "", err
	}
	has := c.Has("checked")
	create, err := c.Factory("factory")
	if err != nil {
		return "", err
	}
	fromFactory, err := create()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("make=%s has=%t factory=%s loads=%d", made.(*service).name, has, fromFactory.(*service).name, loads), nil
}

func missingLoaderError() (string, error) {
	c := container.NewContainer()
	wantErr := errors.New("loader failed")
	c.SetMissingFactoryLoader(func(string) error { return wantErr })
	_, makeErr := c.Make("missing")
	_, factoryErr := c.Factory("missing")
	return fmt.Sprintf("make=%t factory=%t has=%t", errors.Is(makeErr, wantErr), errors.Is(factoryErr, wantErr), c.Has("missing")), nil
}

func facade() (string, error) {
	c := container.NewContainer()
	container.SetProvider(func() *container.Container { return c })
	defer container.SetProvider(nil)
	if err := c.Instance("resource", &service{name: "facade"}); err != nil {
		return "", err
	}
	made, err := container.Make[*service]("resource")
	if err != nil {
		return "", err
	}
	read := container.Value[*service]("resource")
	listed := len(container.List())
	if err := container.Close(context.Background()); err != nil {
		return "", err
	}
	return fmt.Sprintf("make=%s same=%t listed=%d cleared=%t", made.name, made == read, listed, !c.Bound("resource")), nil
}

func providerLifecycle() (value string, err error) {
	app := foundation.NewApplication()
	defer func() {
		if closeErr := app.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close application: %w", closeErr))
		}
	}()
	if bindErr := app.Instance("lifecycle.demo", &service{name: "active"}); bindErr != nil {
		return "", bindErr
	}
	active, makeErr := container.Make[*service]("lifecycle.demo")
	if makeErr != nil {
		return "", makeErr
	}
	if closeErr := app.Close(); closeErr != nil {
		return "", closeErr
	}
	_, afterErr := container.Make[*service]("lifecycle.demo")
	return fmt.Sprintf("before=%s cleared=%t", active.name, errors.Is(afterErr, container.ErrNoCurrentContainer)), nil
}

func errorNotRegistered() (string, error) {
	_, err := container.NewContainer().Make("missing")
	return fmt.Sprintf("not-registered=%t", errors.Is(err, container.ErrFactoryNotRegistered)), nil
}

func errorNilResult() (string, error) {
	c := container.NewContainer()
	if err := c.Singleton("empty", func(containercontract.Resolver) (any, error) { return nil, nil }); err != nil {
		return "", err
	}
	_, err := c.Make("empty")
	return fmt.Sprintf("nil-result=%t", errors.Is(err, container.ErrFactoryReturnedNil)), nil
}

func errorNoCurrent() (string, error) {
	container.SetProvider(nil)
	_, err := container.Make[*service]("missing")
	return fmt.Sprintf("no-current=%t", errors.Is(err, container.ErrNoCurrentContainer)), nil
}

func laravelMapping() (string, error) {
	c := container.NewContainer()
	if err := c.Bind("transient", func(containercontract.Resolver) (any, error) { return &service{name: "bind"}, nil }); err != nil {
		return "", err
	}
	if err := c.Singleton("shared", func(containercontract.Resolver) (any, error) { return &service{name: "singleton"}, nil }); err != nil {
		return "", err
	}
	if err := c.Instance("ready", &service{name: "instance"}); err != nil {
		return "", err
	}
	if err := c.Alias("ready", "alias"); err != nil {
		return "", err
	}
	return fmt.Sprintf("bind=%t singleton=%t instance=%t alias=%t contextual-binding=unsupported", c.Bound("transient"), c.Bound("shared"), c.Bound("ready"), c.Bound("alias")), nil
}
