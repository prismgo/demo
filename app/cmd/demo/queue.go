package demo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/prismgo/framework/console"
	"github.com/prismgo/framework/queue"

	qdemo "prismgo-demo/app/demo/queue"
)

// QueueCommand runs the documentation-backed queue examples.
type QueueCommand struct {
	run func(context.Context, string, string) (qdemo.Result, error)
}

// NewQueueCommand creates the queue demo command.
func NewQueueCommand() *QueueCommand {
	return newQueueCommand(func(ctx context.Context, caseName, connection string) (qdemo.Result, error) {
		cfg := queue.BuildConfig()
		cfg.Default = connection
		if err := queueDemoConfigError(caseName, cfg); err != nil {
			return qdemo.Result{}, err
		}
		manager, err := queue.NewManager(cfg, queue.DefaultRegistry())
		if err != nil {
			return qdemo.Result{}, err
		}
		defer manager.Close()
		return qdemo.Run(ctx, manager, caseName, connection)
	})
}

func queueDemoConfigError(caseName string, cfg queue.Config) error {
	switch caseName {
	case "failed-store":
		if !strings.EqualFold(strings.TrimSpace(cfg.Failed.Driver), "redis") {
			return errors.New("queue demo failed-store requires QUEUE_FAILED_DRIVER=redis")
		}
	case "batch-store":
		if !strings.EqualFold(strings.TrimSpace(cfg.Batching.Driver), "redis") {
			return errors.New("queue demo batch-store requires QUEUE_BATCHING_DRIVER=redis")
		}
	case "restart-store":
		if strings.TrimSpace(cfg.Restart.Cache) == "" {
			return errors.New("queue demo restart-store requires QUEUE_RESTART_CACHE to name a Redis-backed cache store")
		}
	}
	return nil
}

func newQueueCommand(run func(context.Context, string, string) (qdemo.Result, error)) *QueueCommand {
	return &QueueCommand{run: run}
}

// Definition describes the demo:queue command.
func (c *QueueCommand) Definition() *console.Definition {
	definition := console.MustDefinition(
		"demo:queue {case? : Scenario name or list} {--connection=sync : sync, redis, or rabbitmq} {--json : Output JSON}",
		"Run queue examples mapped to the Queue documentation",
	)
	definition.Examples = []string{
		"go run ./demo demo:queue list",
		"go run ./demo demo:queue config",
		"go run ./demo demo:queue failed-store --connection=redis",
		"go run ./demo demo:queue basic --connection=sync",
		"go run ./demo demo:queue basic --connection=redis",
		"go run ./demo demo:queue basic --connection=rabbitmq",
		"go run ./demo demo:queue strategy-envelope",
		"go run ./demo demo:queue unique-options",
		"go run ./demo demo:queue overlap-release",
		"go run ./demo demo:queue failed-commands",
		"go run ./demo demo:queue failure --connection=redis",
		"go run ./demo demo:queue restart --connection=rabbitmq",
		"go run ./demo demo:queue events --connection=redis",
		"go run ./demo demo:queue encryption",
		"go run ./demo demo:queue custom-driver",
		"go run ./demo demo:queue errors",
		"go run ./demo demo:queue redis --connection=redis",
		"go run ./demo demo:queue rabbitmq --connection=rabbitmq",
	}
	return definition
}

// Handle lists queue scenarios or runs the selected scenario.
func (c *QueueCommand) Handle(ctx console.CommandContext) error {
	caseName := strings.TrimSpace(ctx.Input().Argument("case"))
	if caseName == "" || caseName == "list" {
		return renderQueueScenarios(ctx, qdemo.Scenarios())
	}
	connection := strings.TrimSpace(ctx.Input().Option("connection"))
	if connection == "" {
		connection = "sync"
	}
	if !contains([]string{"sync", "redis", "rabbitmq"}, connection) {
		return ctx.Fail(fmt.Sprintf("unknown queue connection %q", connection))
	}
	scenario, ok := queueScenario(caseName)
	if !ok {
		return ctx.Fail(fmt.Sprintf("unknown queue scenario %q", caseName))
	}
	if !contains(scenario.Connections, connection) {
		return ctx.Fail(fmt.Sprintf("queue scenario %q supports %s, not %s", caseName, strings.Join(scenario.Connections, ","), connection))
	}
	result, err := c.run(ctx.Context(), caseName, connection)
	if err != nil {
		return err
	}
	if ctx.Input().OptionBool("json") {
		encoder := json.NewEncoder(console.OutputWriter(ctx.IO()))
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}
	return ctx.IO().Table(
		[]string{"Case", "Connection", "Queue", "Job ID", "Status", "Steps"},
		[][]string{{result.Case, result.Connection, result.Queue, result.JobID, queueDemoStatus(result.Processed), queueDemoSteps(result.Steps)}},
	)
}

func queueScenario(name string) (qdemo.Scenario, bool) {
	for _, scenario := range qdemo.Scenarios() {
		if scenario.Name == name {
			return scenario, true
		}
	}
	return qdemo.Scenario{}, false
}

func queueDemoSteps(steps []string) string {
	if len(steps) == 0 {
		return "-"
	}
	return strings.Join(steps, " -> ")
}

func queueDemoStatus(processed bool) string {
	if processed {
		return "processed"
	}
	return "pending"
}

func renderQueueScenarios(ctx console.CommandContext, scenarios []qdemo.Scenario) error {
	if ctx.Input().OptionBool("json") {
		encoder := json.NewEncoder(console.OutputWriter(ctx.IO()))
		encoder.SetIndent("", "  ")
		return encoder.Encode(scenarios)
	}
	rows := make([][]string, 0, len(scenarios))
	for _, scenario := range scenarios {
		requires := "-"
		if len(scenario.Requires) > 0 {
			requires = strings.Join(scenario.Requires, ",")
		}
		rows = append(rows, []string{
			scenario.Name,
			scenario.Section,
			scenario.Status,
			strings.Join(scenario.Connections, ","),
			requires,
		})
	}
	return ctx.IO().Table([]string{"Case", "Documentation", "Status", "Connections", "Requires"}, rows)
}
