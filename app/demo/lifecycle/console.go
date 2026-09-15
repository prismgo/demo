package lifecycledemo

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/prismgo/framework/console"
	"github.com/prismgo/framework/container"
	providercontract "github.com/prismgo/framework/contracts/provider"
	"github.com/prismgo/framework/event"
	"github.com/prismgo/framework/foundation"
	providerpkg "github.com/prismgo/framework/provider"
)

// handleCommandScenario verifies Application.HandleCommand boots, runs and closes.
func handleCommandScenario(base string) (string, error) {
	ran := false
	factory := func() console.Command {
		return demoCommand{
			name: "lifecycle:probe",
			handle: func(console.CommandContext) error {
				ran = true
				return nil
			},
		}
	}
	app := foundation.Configure(base).
		WithRouting(func(r *foundation.Routing) { r.Commands(factory) }).
		Create()
	defer app.Close()

	if err := app.HandleCommand(context.Background(), []string{"demo", "lifecycle:probe"}); err != nil {
		return "", fmt.Errorf("handle lifecycle command: %w", err)
	}
	return fmt.Sprintf("ran=%t closed=%t", ran, app.Context().Err() != nil), nil
}

// mountCommandProvider mounts a console command from provider Boot via provider.Commands.
type mountCommandProvider struct {
	factory console.CommandFactory
}

// Name returns the stable provider identity.
func (mountCommandProvider) Name() string { return "demo.command.mount" }

// Register performs no binding.
func (mountCommandProvider) Register(providercontract.Application) error { return nil }

// Boot mounts the command declaration through the application starting registrar.
func (p mountCommandProvider) Boot(providercontract.Application) error {
	return providerpkg.Commands(p.factory)
}

// consoleStartingScenario verifies starting event and provider command mounting.
func consoleStartingScenario(base string) (string, error) {
	factory := func() console.Command { return demoCommand{name: "lifecycle:starting"} }
	app := foundation.Configure(base).
		WithProviders(mountCommandProvider{factory: factory}).
		Create()
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	log := &eventLog{}
	bus.Listen(event.EventConsoleApplicationStarting, event.ListenerFunc(func(context.Context, event.Event) error {
		log.add("starting")
		return nil
	}))

	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot console-starting application: %w", err)
	}
	definition, err := app.NewConsoleKernel().All()
	if err != nil {
		return "", fmt.Errorf("load console definitions: %w", err)
	}
	mounted := false
	for _, item := range definition {
		if item.Name == "lifecycle:starting" {
			mounted = true
		}
	}
	return fmt.Sprintf("mounted=%t starting-event=%t", mounted, len(log.values()) == 1), nil
}

// commandResolutionScenario verifies signature parsing into a normalized definition.
func commandResolutionScenario() (string, error) {
	definition := console.MustDefinition("demo:lifecycle {name} {--force}", "Lifecycle resolution probe")
	return fmt.Sprintf("name=%s arguments=%d options=%d",
		definition.Name, len(definition.Arguments), len(definition.Options)), nil
}

// commandExecutionScenario verifies command starting, handle and finished ordering.
func commandExecutionScenario(base string) (string, error) {
	log := &eventLog{}
	factory := func() console.Command {
		return demoCommand{
			name: "lifecycle:execution",
			handle: func(console.CommandContext) error {
				log.add("handle")
				return nil
			},
		}
	}
	app := foundation.Configure(base).
		WithRouting(func(r *foundation.Routing) { r.Commands(factory) }).
		Create()
	defer app.Close()
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot command-execution application: %w", err)
	}

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	bus.Listen(event.EventCommandStarting, event.ListenerFunc(func(context.Context, event.Event) error {
		log.add("starting")
		return nil
	}))
	bus.Listen(event.EventCommandFinished, event.ListenerFunc(func(context.Context, event.Event) error {
		log.add("finished")
		return nil
	}))

	runErr := app.NewConsoleKernel().Call(context.Background(), "lifecycle:execution")
	order := strings.Join(log.values(), ",")
	if order != "starting,handle,finished" {
		return "", fmt.Errorf("command execution order = %q, want starting,handle,finished", order)
	}
	return fmt.Sprintf("order=%s succeeded=%t", order, runErr == nil), nil
}

// sharedApplicationScenario verifies container state written by HTTP is visible to Console.
func sharedApplicationScenario(base string) (string, error) {
	sharedRead := ""
	factory := func() console.Command {
		return demoCommand{
			name: "lifecycle:shared",
			handle: func(console.CommandContext) error {
				value, err := container.Make[string]("demo.shared")
				if err != nil {
					return err
				}
				sharedRead = value
				return nil
			},
		}
	}
	app := foundation.Configure(base).
		WithRouting(func(r *foundation.Routing) {
			r.Commands(factory)
			r.Routes(func(app *foundation.Application, engine *gin.Engine) error {
				engine.GET("/shared", func(c *gin.Context) {
					_ = app.Container().Instance("demo.shared", "from-http")
					c.String(http.StatusOK, "written")
				})
				return nil
			})
		}).
		Create()
	defer app.Close()
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot shared-application application: %w", err)
	}

	server, err := app.NewHTTPServer(context.Background(), "0")
	if err != nil {
		return "", fmt.Errorf("construct shared HTTP server: %w", err)
	}
	response := serveRequest(server, http.MethodGet, "/shared", nil)

	if err := app.NewConsoleKernel().Call(context.Background(), "lifecycle:shared"); err != nil {
		return "", fmt.Errorf("run shared console command: %w", err)
	}
	return fmt.Sprintf("http-wrote=%t console-read=%s", response.Code == http.StatusOK, sharedRead), nil
}

// consoleApplicationStartingEventScenario verifies the console.application.starting payload and timing.
func consoleApplicationStartingEventScenario(base string) (string, error) {
	app := foundation.Configure(base).Create()
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	kernelName := ""
	bus.Listen(event.EventConsoleApplicationStarting, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if value, ok := ev.(event.ConsoleApplicationStarting); ok {
			kernelName = value.KernelName
		}
		return nil
	}))
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot console-application-starting application: %w", err)
	}
	if _, err := app.NewConsoleKernel().All(); err != nil {
		return "", fmt.Errorf("start console application: %w", err)
	}
	if kernelName == "" {
		return "", fmt.Errorf("console.application.starting dispatched without kernel name")
	}
	return "kernel=" + kernelName, nil
}

// consoleCommandStartingEventScenario verifies the console.command.starting payload.
func consoleCommandStartingEventScenario(base string) (string, error) {
	factory := func() console.Command { return demoCommand{name: "lifecycle:command-starting"} }
	app := foundation.Configure(base).
		WithRouting(func(r *foundation.Routing) { r.Commands(factory) }).
		Create()
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	var starting event.CommandStarting
	bus.Listen(event.EventCommandStarting, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if value, ok := ev.(event.CommandStarting); ok {
			starting = value
		}
		return nil
	}))
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot console-command-starting application: %w", err)
	}
	if err := app.NewConsoleKernel().Call(context.Background(), "lifecycle:command-starting"); err != nil {
		return "", fmt.Errorf("run command starting probe: %w", err)
	}
	return fmt.Sprintf("command=%s input-snapshot=%t", starting.Command, len(starting.Input) > 0), nil
}

// consoleCommandFinishedEventScenario verifies the console.command.finished payload.
func consoleCommandFinishedEventScenario(base string) (string, error) {
	factory := func() console.Command { return demoCommand{name: "lifecycle:command-finished"} }
	app := foundation.Configure(base).
		WithRouting(func(r *foundation.Routing) { r.Commands(factory) }).
		Create()
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	var finished event.CommandFinished
	bus.Listen(event.EventCommandFinished, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if value, ok := ev.(event.CommandFinished); ok {
			finished = value
		}
		return nil
	}))
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot console-command-finished application: %w", err)
	}
	if err := app.NewConsoleKernel().Call(context.Background(), "lifecycle:command-finished"); err != nil {
		return "", fmt.Errorf("run command finished probe: %w", err)
	}
	return fmt.Sprintf("command=%s succeeded=%t duration-nonnegative=%t",
		finished.Command, finished.Succeeded, finished.Duration >= 0), nil
}
