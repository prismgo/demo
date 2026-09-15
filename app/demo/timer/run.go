// Package timerdemo contains runnable examples of the Laravel-style task scheduler.
package timerdemo

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prismgo/framework/timer"
)

// Result records one observable scheduler scenario.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

// Run executes a timer catalog scenario.
func Run(name string) (Result, error) {
	value, err := run(name)
	if err != nil {
		return Result{}, fmt.Errorf("timer demo %s: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

func run(name string) (string, error) {
	switch name {
	case "architecture":
		return architectureScenario()
	case "requirements":
		return requirementsScenario()
	case "cron-command":
		return cronCommandScenario()
	case "signal-shutdown":
		return signalShutdownScenario()
	case "registration":
		return registrationScenario()
	case "deployment":
		return deploymentScenario()
	case "timezone":
		return timezoneScenario()
	case "debug-logging":
		return debugLoggingScenario()
	case "overlap-cache":
		return overlapCacheScenario()
	case "exception-config":
		return exceptionConfigurationScenario()
	case "command":
		return commandScenario()
	case "command-validation":
		return commandValidationScenario()
	case "call":
		return callScenario()
	case "every":
		return everyScenario()
	case "second-frequencies":
		return secondFrequenciesScenario()
	case "minute-frequencies":
		return minuteFrequenciesScenario()
	case "hourly":
		return hourlyScenario()
	case "hourly-at":
		return hourlyAtScenario()
	case "hour-steps":
		return hourStepsScenario()
	case "daily":
		return dailyScenario()
	case "daily-at":
		return dailyAtScenario()
	case "twice-daily":
		return twiceDailyScenario()
	case "twice-daily-at":
		return twiceDailyAtScenario()
	case "at":
		return atScenario()
	case "weekly":
		return weeklyScenario()
	case "weekly-on":
		return weeklyOnScenario()
	case "weekday-groups":
		return weekdayGroupsScenario()
	case "named-weekdays":
		return namedWeekdaysScenario()
	case "days":
		return daysScenario()
	case "monthly":
		return monthlyScenario()
	case "monthly-on":
		return monthlyOnScenario()
	case "twice-monthly":
		return twiceMonthlyScenario()
	case "last-day-of-month":
		return lastDayOfMonthScenario()
	case "days-of-month":
		return daysOfMonthScenario()
	case "quarterly":
		return quarterlyScenario()
	case "quarterly-on":
		return quarterlyOnScenario()
	case "yearly":
		return yearlyScenario()
	case "yearly-on":
		return yearlyOnScenario()
	case "without-overlapping":
		return withoutOverlappingScenario()
	case "cross-process-overlap":
		return crossProcessOverlapScenario()
	case "start":
		return startScenario()
	case "stop":
		return stopScenario()
	case "summary":
		return summaryScenario()
	case "task-error":
		return taskErrorScenario()
	case "task-panic":
		return taskPanicScenario()
	case "task-success":
		return taskSuccessScenario()
	case "exception-reporter":
		return exceptionReporterScenario()
	case "fixed-interval-model":
		return fixedIntervalModelScenario()
	case "calendar-model":
		return calendarModelScenario()
	case "immediate-run":
		return immediateRunScenario()
	case "execution-concurrency":
		return executionConcurrencyScenario()
	case "time-parsing":
		return timeParsingScenario()
	case "offset-parsing":
		return offsetParsingScenario()
	case "weekday-parsing":
		return weekdayParsingScenario()
	case "timer-type":
		return timerTypeScenario()
	case "resolved-command":
		return resolvedCommandScenario()
	case "command-resolver":
		return commandResolverScenario()
	case "new-schedule":
		return newScheduleScenario()
	case "name":
		return nameScenario()
	case "description":
		return descriptionScenario()
	case "defaults":
		return defaultsScenario()
	case "standalone":
		return standaloneScenario()
	default:
		return "", fmt.Errorf("unknown scenario %q", name)
	}
}

// recorder runs one schedule and counts how many times each task executes.
// A recorder is single-use: Start spawns one goroutine per task, so run must
// only be called once per recorder.
type recorder struct {
	schedule *timer.Schedule
	mu       sync.Mutex
	counts   map[string]int
}

func newRecorder() *recorder {
	return &recorder{schedule: timer.NewSchedule(), counts: map[string]int{}}
}

// task registers a named closure task and returns it for frequency chaining.
func (r *recorder) task(name string) *timer.ScheduledTask {
	return r.schedule.Call(func(context.Context) error {
		r.mu.Lock()
		r.counts[name]++
		r.mu.Unlock()
		return nil
	}).Name(name)
}

// run starts the schedule, records executions for window, then stops it and
// returns an isolated copy of the per-task execution counts.
func (r *recorder) run(window time.Duration) map[string]int {
	runSchedule(r.schedule, window)

	r.mu.Lock()
	defer r.mu.Unlock()
	snapshot := make(map[string]int, len(r.counts))
	for name, count := range r.counts {
		snapshot[name] = count
	}
	return snapshot
}

// summaryOf splits one Summary line into its name and description columns.
func summaryOf(s *timer.Schedule) (string, string) {
	fields := strings.Fields(s.Summary())
	if len(fields) == 0 {
		return "", ""
	}
	return fields[0], strings.Join(fields[1:], " ")
}

// anyFired reports whether any recorded task executed during the run window.
func anyFired(counts map[string]int) bool {
	for _, count := range counts {
		if count > 0 {
			return true
		}
	}
	return false
}

// runSchedule starts the schedule, waits for window, then cancels and stops it.
func runSchedule(schedule *timer.Schedule, window time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	schedule.Start(ctx)
	time.Sleep(window)
	cancel()
	schedule.Stop()
}

// waitUntil polls cond until it reports true or timeout elapses.
func waitUntil(cond func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(2 * time.Millisecond)
	}
	return cond()
}

// newTask returns a standalone task used to probe registration-time validation.
func newTask() *timer.ScheduledTask {
	return timer.NewSchedule().Call(func(context.Context) error { return nil })
}

// expectOK reports "ok" when no probe panics and "panic" otherwise.
func expectOK(probes ...func()) string {
	for _, probe := range probes {
		if panics(probe) {
			return "panic"
		}
	}
	return "ok"
}

// expectPanic reports "panic" when every probe panics and "ok" otherwise.
func expectPanic(probes ...func()) string {
	for _, probe := range probes {
		if !panics(probe) {
			return "ok"
		}
	}
	return "panic"
}
