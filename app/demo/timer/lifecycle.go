package timerdemo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/prismgo/framework/cmd"
	"github.com/prismgo/framework/console"
	"github.com/prismgo/framework/timer"
)

// stubKernel drives cmd.CronCommand without starting the scheduler goroutines.
type stubKernel struct {
	schedule *timer.Schedule
	started  bool
	stopped  bool
}

// Schedule returns the scheduler the cron command registers into.
func (k *stubKernel) Schedule() *timer.Schedule { return k.schedule }

// Start records the cron command's start call.
func (k *stubKernel) Start(context.Context) { k.started = true }

// Stop records the cron command's stop call.
func (k *stubKernel) Stop() { k.stopped = true }

// cronCommandScenario runs the cron command through its full register/start/stop lifecycle.
func cronCommandScenario() (string, error) {
	kernel := &stubKernel{schedule: timer.NewSchedule()}
	registered := false
	command := cmd.NewCronCommand(kernel, func(schedule *timer.Schedule) {
		registered = schedule != nil
		schedule.Call(func(context.Context) error { return nil }).Name("demo:heartbeat").EveryMinute()
	})

	parent, cancel := context.WithCancel(context.Background())
	cancel()

	var output bytes.Buffer
	commandCtx := console.NewCommandContext(
		parent, command, *command.Definition(), nil,
		console.NewIO(strings.NewReader(""), &output, io.Discard), nil, nil,
	)
	if err := command.Handle(commandCtx); err != nil {
		return "", fmt.Errorf("handle cron command: %w", err)
	}
	text := output.String()
	return fmt.Sprintf("registered=%t started=%t stopped=%t output=%t",
		registered, kernel.started, kernel.stopped,
		strings.Contains(text, "cron scheduler started") && strings.Contains(text, "cron scheduler stopped")), nil
}

// signalShutdownScenario proves a running task observes cancellation and Stop waits for it.
func signalShutdownScenario() (string, error) {
	schedule := timer.NewSchedule()
	started := make(chan struct{})
	canceled := make(chan struct{})
	schedule.Call(func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		close(canceled)
		return nil
	}).Every(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	schedule.Start(ctx)
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		cancel()
		schedule.Stop()
		return "", errors.New("scheduled task did not start")
	}
	cancel()
	schedule.Stop()
	select {
	case <-canceled:
		return "canceled=true stopped=true", nil
	default:
		return "", errors.New("scheduled task did not observe cancellation")
	}
}
