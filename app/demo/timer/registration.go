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

// commandScenario registers a task through a command signature and resolver.
func commandScenario() (string, error) {
	var args string
	schedule := timer.NewSchedule()
	schedule.SetResolver(func(_ string, values []string) (timer.ResolvedCommand, error) {
		args = strings.Join(values, " ")
		return timer.ResolvedCommand{
			Description: "超时工单检测",
			Fn:          func(context.Context) error { return nil },
		}, nil
	})
	schedule.Command("overtime:detect --take=100")

	name, description := summaryOf(schedule)
	return fmt.Sprintf("name=%s args=%s description=%s", name, args, description), nil
}

// commandValidationScenario proves invalid command registration fails during registration.
func commandValidationScenario() (string, error) {
	empty := panics(func() {
		schedule := timer.NewSchedule()
		schedule.SetResolver(func(string, []string) (timer.ResolvedCommand, error) {
			return timer.ResolvedCommand{}, nil
		})
		schedule.Command("")
	})
	nilResolver := panics(func() {
		timer.NewSchedule().Command("overtime:detect")
	})
	unknown := panics(func() {
		schedule := timer.NewSchedule()
		schedule.SetResolver(func(string, []string) (timer.ResolvedCommand, error) {
			return timer.ResolvedCommand{}, errors.New("command not registered")
		})
		schedule.Command("missing:command")
	})
	return fmt.Sprintf("empty=%s nil-resolver=%s unknown=%s", panicState(empty), panicState(nilResolver), panicState(unknown)), nil
}

// callScenario registers a closure task and proves it runs immediately.
func callScenario() (string, error) {
	schedule := timer.NewSchedule()
	var runs atomic.Int64
	schedule.Call(func(context.Context) error {
		runs.Add(1)
		return nil
	}).EveryMinute().Name("tenant_sync").Description("同步租户缓存")

	name, description := summaryOf(schedule)
	ctx, cancel := context.WithCancel(context.Background())
	schedule.Start(ctx)
	time.Sleep(60 * time.Millisecond)
	cancel()
	schedule.Stop()
	return fmt.Sprintf("name=%s description=%s immediate=%t", name, description, runs.Load() > 0), nil
}

// panics reports whether fn panics.
func panics(fn func()) bool {
	panicked := false
	func() {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		fn()
	}()
	return panicked
}

// panicState renders a recovered flag as the observable registration outcome.
func panicState(panicked bool) string {
	if panicked {
		return "panic"
	}
	return "ok"
}
