package timerdemo

import (
	"context"
	"fmt"
	"strings"

	"github.com/prismgo/framework/timer"
)

// architectureScenario keeps every documented scheduler contract referenced at compile time.
func architectureScenario() (string, error) {
	var (
		_ func() *timer.Schedule                                                  = timer.NewSchedule
		_ func(*timer.Schedule, timer.CommandResolver)                            = (*timer.Schedule).SetResolver
		_ func(*timer.Schedule, string) *timer.ScheduledTask                      = (*timer.Schedule).Command
		_ func(*timer.Schedule, func(context.Context) error) *timer.ScheduledTask = (*timer.Schedule).Call
		_ func(*timer.Schedule, context.Context)                                  = (*timer.Schedule).Start
		_ func(*timer.Schedule)                                                   = (*timer.Schedule).Stop
		_ func(*timer.Schedule) string                                            = (*timer.Schedule).Summary
		_ timer.ResolvedCommand
		_ timer.CommandResolver
		_ timer.Timer
	)
	return "schedule=Schedule task=ScheduledTask resolver=CommandResolver timer=Timer", nil
}

// requirementsScenario reports the operating conditions of a long-running scheduler process.
func requirementsScenario() (string, error) {
	return "process=long-running dependencies=time.Local,cache signals=SIGINT,SIGTERM", nil
}

// deploymentScenario reports how the long-running scheduler process is supervised.
func deploymentScenario() (string, error) {
	return "managers=systemd,supervisor restart=always daemonize=false", nil
}

// registrationScenario mirrors app/schedule/register.go: one function adds every task.
func registrationScenario() (string, error) {
	schedule := timer.NewSchedule()
	schedule.SetResolver(func(name string, _ []string) (timer.ResolvedCommand, error) {
		return timer.ResolvedCommand{
			Description: "scheduled " + name,
			Fn:          func(context.Context) error { return nil },
		}, nil
	})
	schedule.Command("overtime:detect --take=100000").EveryFiveMinutes()
	schedule.Command("followup:generate").EveryTenMinutes()
	schedule.Command("area:sync").Sundays().At("03:00")

	names := make([]string, 0, 3)
	for _, line := range strings.Split(strings.TrimSpace(schedule.Summary()), "\n") {
		if fields := strings.Fields(line); len(fields) > 0 {
			names = append(names, fields[0])
		}
	}
	return fmt.Sprintf("tasks=%d names=%s", len(names), strings.Join(names, ",")), nil
}
