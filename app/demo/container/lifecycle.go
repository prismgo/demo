package containerdemo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/prismgo/framework/container"
	containercontract "github.com/prismgo/framework/contracts/container"
)

func closeGroups() (string, error) {
	c := container.NewContainer()
	var closed []string
	for _, item := range []struct {
		key   string
		group container.CloseGroup
	}{
		{key: "normal", group: container.CloseGroupNormal},
		{key: "reporting", group: container.CloseGroupReporting},
	} {
		if err := c.Instance(item.key, &service{name: item.key}, container.WithCloseGroup(item.group), container.WithCloser(func(s *service) error {
			closed = append(closed, s.name)
			return nil
		})); err != nil {
			return "", err
		}
	}
	if err := c.CloseGroup(context.Background(), container.CloseGroupNormal); err != nil {
		return "", err
	}
	return fmt.Sprintf("closed=%s reporting-bound=%t", strings.Join(closed, ","), c.Bound("reporting")), nil
}

func closeGroup() (string, error) {
	c := container.NewContainer()
	var closed []string
	for _, item := range []struct {
		key   string
		group container.CloseGroup
	}{
		{key: "normal", group: container.CloseGroupNormal},
		{key: "reporting", group: container.CloseGroupReporting},
	} {
		if err := c.Instance(item.key, &service{name: item.key}, container.WithCloseGroup(item.group), container.WithCloser(func(s *service) error {
			closed = append(closed, s.name)
			return nil
		})); err != nil {
			return "", err
		}
	}
	if err := c.Close(context.Background()); err != nil {
		return "", err
	}
	return "closed=" + strings.Join(closed, ","), nil
}

func closeOrder() (string, error) {
	c := container.NewContainer()
	var closed []string
	for _, name := range []string{"first", "second", "third"} {
		if err := c.Instance(name, &service{name: name}, container.WithCloser(func(s *service) error {
			closed = append(closed, s.name)
			return nil
		})); err != nil {
			return "", err
		}
	}
	if err := c.Close(context.Background()); err != nil {
		return "", err
	}
	return strings.Join(closed, ","), nil
}

func closerOwnership() (string, error) {
	c := container.NewContainer()
	closed := 0
	closer := container.WithCloser(func(*service) error { closed++; return nil })
	factory := func(containercontract.Resolver) (any, error) { return &service{name: "resource"}, nil }
	if err := c.Bind("transient", factory, closer); err != nil {
		return "", err
	}
	if err := c.Singleton("shared", factory, closer); err != nil {
		return "", err
	}
	if err := c.Instance("ready", &service{name: "ready"}, closer); err != nil {
		return "", err
	}
	if _, err := c.Make("transient"); err != nil {
		return "", err
	}
	if _, err := c.Make("shared"); err != nil {
		return "", err
	}
	if err := c.Close(context.Background()); err != nil {
		return "", err
	}
	return fmt.Sprintf("closed=%d", closed), nil
}

func closePreCancel() (string, error) {
	c := container.NewContainer()
	if err := c.Instance("resource", &service{name: "resource"}); err != nil {
		return "", err
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := c.Close(ctx)
	return fmt.Sprintf("cancelled=%t retained=%t", errors.Is(err, context.Canceled), c.Bound("resource")), nil
}

func closeSuccess() (string, error) {
	c := container.NewContainer()
	if err := c.Instance("resource", &service{name: "resource"}); err != nil {
		return "", err
	}
	if err := c.Close(context.Background()); err != nil {
		return "", err
	}
	return fmt.Sprintf("bound=%t resolved=%t", c.Bound("resource"), c.Resolved("resource")), nil
}

func closeRetry() (string, error) {
	c := container.NewContainer()
	attempts := 0
	wantErr := errors.New("temporary close failure")
	if err := c.Instance("resource", &service{name: "resource"}, container.WithCloser(func(*service) error {
		attempts++
		if attempts == 1 {
			return wantErr
		}
		return nil
	})); err != nil {
		return "", err
	}
	firstErr := c.Close(context.Background())
	retained := c.Bound("resource")
	if err := c.Close(context.Background()); err != nil {
		return "", err
	}
	return fmt.Sprintf("error=%t retained=%t cleared=%t attempts=%d", errors.Is(firstErr, wantErr), retained, !c.Bound("resource"), attempts), nil
}

func closeMidCancel() (string, error) {
	c := container.NewContainer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := c.Instance("remaining", &service{name: "remaining"}, container.WithContextCloser(func(closeCtx context.Context, _ *service) error {
		return closeCtx.Err()
	})); err != nil {
		return "", err
	}
	if err := c.Instance("cancel", &service{name: "cancel"}, container.WithContextCloser(func(context.Context, *service) error {
		cancel()
		return context.Canceled
	})); err != nil {
		return "", err
	}
	err := c.Close(ctx)
	return fmt.Sprintf("cancelled=%t remaining=%t", errors.Is(err, context.Canceled), c.Bound("remaining")), nil
}
