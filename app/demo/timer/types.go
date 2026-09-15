package timerdemo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/prismgo/framework/timer"
)

// timerTypeScenario exercises the legacy Timer base type and its zero value.
func timerTypeScenario() (string, error) {
	timerType := timer.Timer{Interval: 90 * time.Second}
	return fmt.Sprintf("interval=%s zero=%t", timerType.Interval, timer.Timer{}.Interval == 0), nil
}

// resolvedCommandScenario runs the command/resolver bridge result directly.
func resolvedCommandScenario() (string, error) {
	var ran atomic.Bool
	resolved := timer.ResolvedCommand{
		Description: "同步租户缓存",
		Fn: func(context.Context) error {
			ran.Store(true)
			return nil
		},
	}
	if err := resolved.Fn(context.Background()); err != nil {
		return "", fmt.Errorf("run resolved command: %w", err)
	}
	return fmt.Sprintf("description=%s ran=%t", resolved.Description, ran.Load()), nil
}

// commandResolverScenario injects a resolver and checks signature parsing and errors.
func commandResolverScenario() (string, error) {
	var name string
	var args []string
	schedule := timer.NewSchedule()
	schedule.SetResolver(func(command string, values []string) (timer.ResolvedCommand, error) {
		name = command
		args = values
		return timer.ResolvedCommand{
			Description: "命令解析",
			Fn:          func(context.Context) error { return nil },
		}, nil
	})
	schedule.Command("overtime:detect --take=10 extra")

	failing := timer.NewSchedule()
	failing.SetResolver(func(string, []string) (timer.ResolvedCommand, error) {
		return timer.ResolvedCommand{}, errors.New("command not registered")
	})
	invalid := expectPanic(func() { failing.Command("missing:command") })
	return fmt.Sprintf("name=%s args=%s invalid=%s", name, strings.Join(args, " "), invalid), nil
}

// newScheduleScenario builds an empty scheduler and registers one command task.
func newScheduleScenario() (string, error) {
	schedule := timer.NewSchedule()
	empty := strings.TrimSpace(schedule.Summary()) == ""
	schedule.SetResolver(func(string, []string) (timer.ResolvedCommand, error) {
		return timer.ResolvedCommand{
			Description: "超时工单检测",
			Fn:          func(context.Context) error { return nil },
		}, nil
	})
	schedule.Command("overtime:detect")
	tasks := len(strings.Split(strings.TrimSpace(schedule.Summary()), "\n"))
	return fmt.Sprintf("empty=%t tasks=%d", empty, tasks), nil
}

// nameScenario proves Name feeds the Summary name column.
func nameScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("tenant_sync").EveryMinute()
	name, _ := summaryOf(recorder.schedule)
	return "name=" + name, nil
}

// descriptionScenario proves Description feeds the Summary description column.
func descriptionScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("tenant_sync").EveryMinute().Description("同步租户缓存")
	_, description := summaryOf(recorder.schedule)
	return "description=" + description, nil
}

// defaultsScenario proves an unconfigured task keeps the immediate minute default.
func defaultsScenario() (string, error) {
	schedule := timer.NewSchedule()
	var runs atomic.Int64
	schedule.Call(func(context.Context) error {
		runs.Add(1)
		return nil
	}).Name("default-job")
	runSchedule(schedule, 60*time.Millisecond)
	return fmt.Sprintf("immediate=%t", runs.Load() > 0), nil
}

// standaloneScenario drives a scheduler without the Console Kernel.
func standaloneScenario() (string, error) {
	schedule := timer.NewSchedule()
	var commandRuns, callRuns atomic.Int64
	schedule.SetResolver(func(string, []string) (timer.ResolvedCommand, error) {
		return timer.ResolvedCommand{
			Description: "同步客户缓存",
			Fn: func(context.Context) error {
				commandRuns.Add(1)
				return nil
			},
		}, nil
	})
	schedule.Command("customer:sync --take=500").EveryMinute()
	schedule.Call(func(context.Context) error {
		callRuns.Add(1)
		return nil
	}).Every(10 * time.Millisecond).Name("report_cache").Description("刷新报表缓存")
	runSchedule(schedule, 80*time.Millisecond)
	return fmt.Sprintf("tasks=2 command=%t call=%t", commandRuns.Load() > 0, callRuns.Load() > 0), nil
}
