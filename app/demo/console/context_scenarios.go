package consoledemo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/prismgo/framework/console"
	"github.com/spf13/cobra"
)

func isContextScenario(name string) bool {
	switch name {
	case "handle-context", "from-context", "from-command", "cobra-command", "context-methods", "trap", "trap-release", "fail", "manual-failure":
		return true
	}
	return false
}

func contextFixture() (console.CommandContext, *cobra.Command, *bytes.Buffer) {
	definition := *console.MustDefinition("demo:sample {user}", "Sample command")
	cmd := &cobra.Command{Use: definition.Name}
	input := console.NewInput(definition, cmd, []string{"Ada"})
	out := &bytes.Buffer{}
	ctx := console.NewCommandContext(context.Background(), &sampleCommand{}, definition, input, console.NewIO(bytes.NewReader(nil), out, &bytes.Buffer{}), nil, cmd)
	cmd.SetContext(ctx.Context())
	return ctx, cmd, out
}

func runContextScenario(name string) (Result, error) {
	result := Result{Case: name}
	ctx, cmd, out := contextFixture()
	result.Definition = *ctx.Definition()
	switch name {
	case "handle-context":
		command := &sampleCommand{}
		if err := command.Handle(ctx); err != nil {
			return Result{}, fmt.Errorf("console demo handle-context: %w", err)
		}
		result.Values = []string{ctx.CommandName(), command.handled[0]}
	case "from-context":
		found, ok := console.FromContext(ctx.Context())
		injected, injectedOK := console.FromContext(console.WithContext(context.Background(), ctx))
		_, missing := console.FromContext(context.Background())
		result.Values = []string{fmt.Sprintf("bound=%t", ok && found == ctx), fmt.Sprintf("injected=%t", injectedOK && injected == ctx), fmt.Sprintf("missing=%t", missing)}
	case "from-command":
		found, ok := console.FromCommand(cmd)
		required, err := console.MustFromCommand(cmd)
		if err != nil {
			return Result{}, fmt.Errorf("console demo from-command: %w", err)
		}
		_, missing := console.FromCommand(&cobra.Command{})
		_, missingErr := console.MustFromCommand(&cobra.Command{})
		if missingErr == nil {
			return Result{}, fmt.Errorf("console demo from-command: missing context was accepted")
		}
		result.Values = []string{fmt.Sprintf("found=%t", ok && found == ctx), fmt.Sprintf("required=%t", required == ctx), fmt.Sprintf("missing=%t", missing), missingErr.Error()}
	case "cobra-command":
		result.Values = []string{fmt.Sprintf("same=%t", console.CobraCommand(ctx) == cmd), fmt.Sprintf("nil=%t", console.CobraCommand(nil) == nil)}
	case "context-methods":
		definition := ctx.Definition()
		definition.Name = "changed"
		result.Values = []string{ctx.CommandName(), ctx.Definition().Name, ctx.Input().Argument("user"), fmt.Sprintf("io=%t", ctx.IO() != nil), fmt.Sprintf("context=%t", ctx.Context() != nil)}
	case "trap", "trap-release":
		return runTrapScenario(name, ctx)
	case "fail":
		cause := errors.New("disk unavailable")
		for _, failure := range []error{ctx.Fail("stopped"), ctx.Fail(cause), ctx.Fail("write failed:", cause)} {
			result.Values = append(result.Values, failure.Error())
		}
		result.Values = append(result.Values, fmt.Sprintf("wrapped=%t", errors.Is(ctx.Fail("write failed:", cause), cause)))
	case "manual-failure":
		manual, ok := console.IsManualFailure(fmt.Errorf("outer: %w", ctx.Fail("stopped")))
		if !ok {
			return Result{}, fmt.Errorf("console demo manual-failure: wrapped failure was not recognized")
		}
		_, ordinary := console.IsManualFailure(errors.New("ordinary"))
		result.Values = []string{fmt.Sprintf("recognized=%t", ok), manual.Message, fmt.Sprintf("ordinary=%t", ordinary)}
	}
	result.Output = out.String()
	return result, nil
}

func runTrapScenario(name string, ctx console.CommandContext) (Result, error) {
	result := Result{Case: name}
	if _, err := ctx.Trap(nil, func(os.Signal) {}); err == nil {
		return Result{}, fmt.Errorf("console demo %s: empty signal list accepted", name)
	}
	if _, err := ctx.Trap([]os.Signal{os.Interrupt}, nil); err == nil {
		return Result{}, fmt.Errorf("console demo %s: nil callback accepted", name)
	}
	received := make(chan os.Signal, 1)
	release, err := ctx.Trap([]os.Signal{os.Interrupt}, func(sig os.Signal) { received <- sig })
	if err != nil {
		return Result{}, fmt.Errorf("console demo %s register trap: %w", name, err)
	}
	defer console.ReleaseTraps(ctx)
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		return Result{}, fmt.Errorf("console demo %s find process: %w", name, err)
	}
	if err := process.Signal(os.Interrupt); err != nil {
		return Result{}, fmt.Errorf("console demo %s send signal: %w", name, err)
	}
	select {
	case sig := <-received:
		result.Values = []string{sig.String()}
	case <-time.After(time.Second):
		return Result{}, fmt.Errorf("console demo %s: signal callback timed out", name)
	}
	if name == "trap-release" {
		release()
		release()
		console.ReleaseTraps(ctx)
		result.Values = append(result.Values, "released")
	}
	return result, nil
}
