package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/prismgo/framework/cache"
	"github.com/prismgo/framework/console"
	"github.com/prismgo/framework/container"
	"github.com/prismgo/framework/event"
	"github.com/prismgo/framework/kernel"
)

type command struct {
	definition *console.Definition
	handle     func(console.CommandContext) error
}

func (c *command) Definition() *console.Definition         { return c.definition }
func (c *command) Handle(ctx console.CommandContext) error { return c.handle(ctx) }

type promptCommand struct {
	command
	custom bool
	values *[]string
}

func (c *promptCommand) PromptForMissingArgumentsUsing() map[string]console.MissingArgumentPrompt {
	if !c.custom {
		return nil
	}
	return map[string]console.MissingArgumentPrompt{
		"user": {Question: "Custom user", Ask: func(ctx console.CommandContext, _ console.Argument) (string, error) {
			return ctx.IO().Ask("Custom user")
		}},
		"team": {Question: "Custom team", Default: "dev"},
	}
}

func (c *promptCommand) AfterPromptingForMissingArguments(ctx console.CommandContext) error {
	*c.values = append(*c.values, "hook="+ctx.Argument("user")+":"+ctx.Argument("team"))
	return nil
}

type isolatedCommand struct {
	command
	started chan struct{}
	release chan struct{}
	count   atomic.Int32
}

func (c *isolatedCommand) IsolationKey(console.CommandContext) string {
	return "console-demo-isolation"
}
func (c *isolatedCommand) Handle(ctx console.CommandContext) error {
	c.count.Add(1)
	close(c.started)
	select {
	case <-c.release:
		return nil
	case <-ctx.Context().Done():
		return ctx.Context().Err()
	}
}

func setup() error {
	registry := container.NewContainer()
	manager, err := cache.NewManager(cache.Config{})
	if err != nil {
		return err
	}
	if err := registry.Instance("cache.manager", manager); err != nil {
		return err
	}
	if err := registry.Instance("event.dispatcher", event.New()); err != nil {
		return err
	}
	container.SetProvider(func() *container.Container { return registry })
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "helper mode is required")
		os.Exit(2)
	}
	mode := os.Args[1]
	if strings.HasPrefix(mode, "package-") {
		runPackage(mode)
		return
	}
	if err := setup(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer container.SetProvider(nil)
	values, err := runKernel(mode)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	data, err := json.Marshal(values)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Printf("SCENARIO_RESULT=%s\n", data)
}

func runPackage(mode string) {
	switch mode {
	case "package-output":
		console.Line("plain")
		console.Info("info")
		console.Comment("comment")
		console.Question("question")
		console.Success("success")
		console.Warn("warning")
		console.Error("error")
		console.Alert("alert")
	case "package-exit":
		console.Exit("exit failure")
	case "package-exit-if":
		console.ExitIf(fmt.Errorf("conditional failure"))
	case "package-exit-if-nil":
		console.ExitIf(nil)
		fmt.Println("continued")
	}
}

func runKernel(mode string) ([]string, error) {
	k := kernel.New("demo")
	var values []string
	switch mode {
	case "ansi-detection":
		plain := &strings.Builder{}
		if err := os.Setenv("NO_COLOR", ""); err != nil {
			return nil, fmt.Errorf("clear NO_COLOR: %w", err)
		}
		if err := os.Setenv("FORCE_COLOR", ""); err != nil {
			return nil, fmt.Errorf("clear FORCE_COLOR: %w", err)
		}
		values = append(values, fmt.Sprintf("plain=%t", console.ResolveANSI(plain, false, false)))
		if err := os.Setenv("FORCE_COLOR", "1"); err != nil {
			return nil, fmt.Errorf("enable FORCE_COLOR: %w", err)
		}
		values = append(values, fmt.Sprintf("force-color=%t", console.SupportsANSI(plain)))
		if err := os.Setenv("NO_COLOR", "1"); err != nil {
			return nil, fmt.Errorf("enable NO_COLOR: %w", err)
		}
		values = append(values, fmt.Sprintf("no-color=%t", console.SupportsANSI(plain)))
		values = append(values, fmt.Sprintf("forced=%t", console.ResolveANSI(plain, true, false)))
		values = append(values, fmt.Sprintf("disabled=%t", console.ResolveANSI(plain, true, true)))
		values = append(values, fmt.Sprintf("options=%+v", console.ResolveOutputOptions(plain, true, false, true, false)))
		return values, nil
	case "call", "call-silently", "call-input":
		target := &command{definition: console.MustDefinition("demo:target {user} {--queue=default}", "Target"), handle: func(ctx console.CommandContext) error {
			values = append(values, ctx.Argument("user"), ctx.Option("queue"), fmt.Sprintf("discard=%t", console.OutputWriter(ctx.IO()) == io.Discard))
			ctx.IO().Info("target output")
			return nil
		}}
		caller := &command{definition: console.MustDefinition("demo:caller", "Caller"), handle: func(ctx console.CommandContext) error {
			switch mode {
			case "call":
				return ctx.Call("demo:target Ada --queue=priority")
			case "call-silently":
				return ctx.CallSilently("demo:target Ada --queue=priority")
			default:
				return ctx.Call("demo:target", console.CallInput{Arguments: map[string]any{"user": "Ada"}, Options: map[string]any{"queue": "priority"}})
			}
		}}
		k.Register(target, caller)
		err := k.Call(context.Background(), "demo:caller")
		return values, err
	case "missing-input", "missing-input-custom":
		custom := mode == "missing-input-custom"
		signature := "demo:prompt {user : User name}"
		if custom {
			signature = "demo:prompt {user} {team}"
		}
		cmd := &promptCommand{custom: custom, values: &values}
		cmd.definition = console.MustDefinition(signature, "Prompt")
		cmd.handle = func(ctx console.CommandContext) error {
			values = append(values, ctx.Argument("user"))
			if custom {
				values = append(values, ctx.Argument("team"))
			}
			return nil
		}
		if custom {
			k.Register(cmd)
		} else {
			k.Register(&cmd.command)
		}
		err := k.RunContextArgv(context.Background(), []string{"demo", "demo:prompt"})
		return values, err
	case "isolatable":
		cmd := &isolatedCommand{started: make(chan struct{}), release: make(chan struct{})}
		cmd.definition = console.MustDefinition("demo:isolated", "Isolated")
		k.Register(cmd)
		first := make(chan error, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		go func() { first <- k.Call(ctx, "demo:isolated --isolated") }()
		select {
		case <-cmd.started:
		case err := <-first:
			if err != nil {
				return nil, fmt.Errorf("first isolated call ended before Handle: %w", err)
			}
			return nil, fmt.Errorf("first isolated call ended before Handle without error")
		case <-ctx.Done():
			close(cmd.release)
			<-first
			return nil, fmt.Errorf("first isolated call did not start: %w", ctx.Err())
		}
		second := k.Call(ctx, "demo:isolated --isolated")
		close(cmd.release)
		firstErr := <-first
		return []string{fmt.Sprintf("first=%v", firstErr), fmt.Sprintf("second=%v", second), fmt.Sprintf("count=%d", cmd.count.Load())}, nil
	default:
		return nil, fmt.Errorf("unknown helper mode %q", mode)
	}
}
