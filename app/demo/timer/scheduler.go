package timerdemo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/prismgo/framework/container"
	"github.com/prismgo/framework/exception"
	"github.com/prismgo/framework/timer"
)

// exceptionHandlerKey is the container binding the scheduler reports through.
const exceptionHandlerKey = "exception.handler"

// captureReports installs a temporary reporter on the application exception
// handler. It returns a counter read function and a function that restores the
// original reporter list.
func captureReports() (func() int, func()) {
	handler, err := container.Make[*exception.Handler](exceptionHandlerKey)
	if err != nil || handler == nil {
		return func() int { return 0 }, func() {}
	}
	var total atomic.Int64
	original := handler.Reporters
	reporters := make([]exception.Reporter, len(original), len(original)+1)
	copy(reporters, original)
	reporters = append(reporters, func(any, error, map[string]any) { total.Add(1) })
	handler.Reporters = reporters
	return func() int { return int(total.Load()) }, func() { handler.Reporters = original }
}

// startScenario proves Start returns while the registered task runs.
func startScenario() (string, error) {
	schedule := timer.NewSchedule()
	var runs atomic.Int64
	schedule.Call(func(context.Context) error {
		runs.Add(1)
		return nil
	}).Every(10 * time.Millisecond).Name("start-job")

	ctx, cancel := context.WithCancel(context.Background())
	schedule.Start(ctx)
	ran := waitUntil(func() bool { return runs.Load() > 0 }, time.Second)
	cancel()
	schedule.Stop()
	return fmt.Sprintf("ran=%t", ran), nil
}

// stopScenario proves Stop returns and no further executions happen.
func stopScenario() (string, error) {
	schedule := timer.NewSchedule()
	var runs atomic.Int64
	schedule.Call(func(context.Context) error {
		runs.Add(1)
		return nil
	}).Every(10 * time.Millisecond).Name("stop-job")

	ctx, cancel := context.WithCancel(context.Background())
	schedule.Start(ctx)
	ran := waitUntil(func() bool { return runs.Load() > 0 }, time.Second)
	cancel()
	schedule.Stop()
	afterStop := runs.Load()
	time.Sleep(60 * time.Millisecond)
	return fmt.Sprintf("ran=%t idle=%t", ran, runs.Load() == afterStop), nil
}

// summaryScenario renders the registered task list through Summary.
func summaryScenario() (string, error) {
	schedule := timer.NewSchedule()
	schedule.SetResolver(func(string, []string) (timer.ResolvedCommand, error) {
		return timer.ResolvedCommand{
			Description: "超时工单检测",
			Fn:          func(context.Context) error { return nil },
		}, nil
	})
	schedule.Command("overtime:detect")
	schedule.Call(func(context.Context) error { return nil }).Name("tenant_sync").Description("同步租户缓存")

	names := make([]string, 0, 2)
	for _, line := range strings.Split(strings.TrimSpace(schedule.Summary()), "\n") {
		if fields := strings.Fields(line); len(fields) > 0 {
			names = append(names, fields[0])
		}
	}
	return fmt.Sprintf("lines=%d names=%s", len(names), strings.Join(names, ",")), nil
}

// taskErrorScenario proves a returned error is reported and the loop continues.
func taskErrorScenario() (string, error) {
	reports, restore := captureReports()
	defer restore()

	schedule := timer.NewSchedule()
	var runs atomic.Int64
	schedule.Call(func(context.Context) error {
		runs.Add(1)
		return errors.New("scheduled task failed")
	}).Every(10 * time.Millisecond).Name("flaky-job")
	runSchedule(schedule, 80*time.Millisecond)
	return fmt.Sprintf("reported=%t continued=%t", reports() > 0, runs.Load() >= 2), nil
}

// taskPanicScenario proves a panic is reported and the loop continues.
func taskPanicScenario() (string, error) {
	reports, restore := captureReports()
	defer restore()

	schedule := timer.NewSchedule()
	var runs atomic.Int64
	schedule.Call(func(context.Context) error {
		runs.Add(1)
		panic("scheduled task exploded")
	}).Every(10 * time.Millisecond).Name("panic-job")
	runSchedule(schedule, 80*time.Millisecond)
	return fmt.Sprintf("reported=%t continued=%t", reports() > 0, runs.Load() >= 2), nil
}

// taskSuccessScenario proves a successful task runs without exception reports.
func taskSuccessScenario() (string, error) {
	reports, restore := captureReports()
	defer restore()

	schedule := timer.NewSchedule()
	var runs atomic.Int64
	schedule.Call(func(context.Context) error {
		runs.Add(1)
		return nil
	}).Every(10 * time.Millisecond).Name("stable-job")
	runSchedule(schedule, 60*time.Millisecond)
	return fmt.Sprintf("reported=%t ran=%t", reports() > 0, runs.Load() > 0), nil
}

// exceptionReporterScenario checks the reporter binding and its default filters.
func exceptionReporterScenario() (string, error) {
	handler, err := container.Make[*exception.Handler](exceptionHandlerKey)
	if err != nil || handler == nil {
		return "bound=false normal=false canceled=false", nil
	}
	normal := handler.ShouldReport(errors.New("boom"), 500)
	canceled := !handler.ShouldReport(context.Canceled, 500)
	return fmt.Sprintf("bound=true normal=%t canceled=%t", normal, canceled), nil
}
