package timerdemo_test

import (
	"testing"

	"prismgo-demo/app/demo/catalog"
	demotest "prismgo-demo/app/demo/testing"
	timerdemo "prismgo-demo/app/demo/timer"
)

// expectValue executes one scenario and asserts its full observable value.
func expectValue(t *testing.T, name string, want string) {
	t.Helper()
	result, err := timerdemo.Run(name)
	if err != nil {
		t.Fatalf("timer demo %q error = %v, want nil", name, err)
	}
	if result.Case != name {
		t.Fatalf("timer demo %q case = %q, want %q", name, result.Case, name)
	}
	if result.Value != want {
		t.Fatalf("timer demo %q value = %q, want %q", name, result.Value, want)
	}
}

// newApplication boots an isolated application whose container binds the config,
// logger, and exception handlers the scheduler reads while running tasks.
func newApplication(t *testing.T) {
	t.Helper()
	demotest.NewApplication(t, demotest.Options{})
}

// expectRunning boots an application, then runs a scenario that starts the scheduler.
func expectRunning(t *testing.T, name string, want string) {
	t.Helper()
	newApplication(t)
	expectValue(t, name, want)
}

func TestTimerDemoArchitecture(t *testing.T) {
	expectValue(t, "architecture", "schedule=Schedule task=ScheduledTask resolver=CommandResolver timer=Timer")
}

func TestTimerDemoRequirements(t *testing.T) {
	expectValue(t, "requirements", "process=long-running dependencies=time.Local,cache signals=SIGINT,SIGTERM")
}

func TestTimerDemoCronCommand(t *testing.T) {
	expectValue(t, "cron-command", "registered=true started=true stopped=true output=true")
}

func TestTimerDemoSignalShutdown(t *testing.T) {
	expectRunning(t, "signal-shutdown", "canceled=true stopped=true")
}

func TestTimerDemoRegistration(t *testing.T) {
	expectValue(t, "registration", "tasks=3 names=overtime:detect,followup:generate,area:sync")
}

func TestTimerDemoDeployment(t *testing.T) {
	expectValue(t, "deployment", "managers=systemd,supervisor restart=always daemonize=false")
}

func TestTimerDemoTimezone(t *testing.T) {
	expectValue(t, "timezone", "location=demo-zone offset=+08:00")
}

func TestTimerDemoDebugLogging(t *testing.T) {
	newApplication(t)
	expectValue(t, "debug-logging", "app.debug=false")
}

func TestTimerDemoOverlapCacheConfiguration(t *testing.T) {
	newApplication(t)
	expectValue(t, "overlap-cache", "cache.default=memory")
}

func TestTimerDemoExceptionConfiguration(t *testing.T) {
	newApplication(t)
	expectValue(t, "exception-config", "reporter=bound")
}

func TestTimerDemoCommand(t *testing.T) {
	expectValue(t, "command", "name=overtime:detect args=--take=100 description=超时工单检测")
}

func TestTimerDemoCommandValidation(t *testing.T) {
	expectValue(t, "command-validation", "empty=panic nil-resolver=panic unknown=panic")
}

func TestTimerDemoCall(t *testing.T) {
	expectRunning(t, "call", "name=tenant_sync description=同步租户缓存 immediate=true")
}

func TestTimerDemoEvery(t *testing.T) {
	expectRunning(t, "every", "immediate=true repeated=true")
}

func TestTimerDemoSecondFrequencies(t *testing.T) {
	expectRunning(t, "second-frequencies", "methods=7 every-second-fired=true")
}

func TestTimerDemoMinuteFrequencies(t *testing.T) {
	expectRunning(t, "minute-frequencies", "immediate=6 deferred=2")
}

func TestTimerDemoHourly(t *testing.T) {
	expectRunning(t, "hourly", "immediate=true")
}

func TestTimerDemoHourlyAt(t *testing.T) {
	expectRunning(t, "hourly-at", "offsets=3 immediate=false")
}

func TestTimerDemoHourSteps(t *testing.T) {
	expectRunning(t, "hour-steps", "methods=5 immediate=false")
}

func TestTimerDemoDaily(t *testing.T) {
	expectRunning(t, "daily", "immediate=true")
}

func TestTimerDemoDailyAt(t *testing.T) {
	expectRunning(t, "daily-at", "at=18:30 immediate=false")
}

func TestTimerDemoTwiceDaily(t *testing.T) {
	expectRunning(t, "twice-daily", "hours=2 immediate=false")
}

func TestTimerDemoTwiceDailyAt(t *testing.T) {
	expectRunning(t, "twice-daily-at", "hours=2 offset=15 immediate=false")
}

func TestTimerDemoAt(t *testing.T) {
	expectRunning(t, "at", "at=18:30 weekdays=5 immediate=false")
}

func TestTimerDemoWeekly(t *testing.T) {
	expectRunning(t, "weekly", "weekday=sunday immediate=false")
}

func TestTimerDemoWeeklyOn(t *testing.T) {
	expectRunning(t, "weekly-on", "weekday=monday at=09:30 immediate=false")
}

func TestTimerDemoWeekdayGroups(t *testing.T) {
	expectRunning(t, "weekday-groups", "weekdays=5 weekends=2 immediate=false")
}

func TestTimerDemoNamedWeekdays(t *testing.T) {
	expectRunning(t, "named-weekdays", "named=7 immediate=false")
}

func TestTimerDemoDays(t *testing.T) {
	expectRunning(t, "days", "ints=3 strings=3 immediate=false")
}

func TestTimerDemoMonthly(t *testing.T) {
	expectRunning(t, "monthly", "day=1 immediate=false")
}

func TestTimerDemoMonthlyOn(t *testing.T) {
	expectRunning(t, "monthly-on", "day=15 at=08:30 immediate=false invalid=panic")
}

func TestTimerDemoTwiceMonthly(t *testing.T) {
	expectRunning(t, "twice-monthly", "days=1,16 at=10:00 immediate=false invalid=panic")
}

func TestTimerDemoLastDayOfMonth(t *testing.T) {
	expectRunning(t, "last-day-of-month", "at=23:55 immediate=false invalid=panic")
}

func TestTimerDemoDaysOfMonth(t *testing.T) {
	expectRunning(t, "days-of-month", "days=3 at=09:00 immediate=false invalid=panic")
}

func TestTimerDemoQuarterly(t *testing.T) {
	expectRunning(t, "quarterly", "day=1 immediate=false")
}

func TestTimerDemoQuarterlyOn(t *testing.T) {
	expectRunning(t, "quarterly-on", "day=45 at=10:00 immediate=false invalid=panic")
}

func TestTimerDemoYearly(t *testing.T) {
	expectRunning(t, "yearly", "month=1 day=1 immediate=false")
}

func TestTimerDemoYearlyOn(t *testing.T) {
	expectRunning(t, "yearly-on", "month=12 day=31 at=23:59 last-token=ok immediate=false invalid=panic")
}

func TestTimerDemoWithoutOverlapping(t *testing.T) {
	expectRunning(t, "without-overlapping", "blocked=true released=true invalid=panic")
}

func TestTimerDemoStart(t *testing.T) {
	expectRunning(t, "start", "ran=true")
}

func TestTimerDemoStop(t *testing.T) {
	expectRunning(t, "stop", "ran=true idle=true")
}

func TestTimerDemoSummary(t *testing.T) {
	expectValue(t, "summary", "lines=2 names=overtime:detect,tenant_sync")
}

func TestTimerDemoTaskError(t *testing.T) {
	expectRunning(t, "task-error", "reported=true continued=true")
}

func TestTimerDemoTaskPanic(t *testing.T) {
	expectRunning(t, "task-panic", "reported=true continued=true")
}

func TestTimerDemoTaskSuccess(t *testing.T) {
	expectRunning(t, "task-success", "reported=false ran=true")
}

func TestTimerDemoExceptionReporter(t *testing.T) {
	newApplication(t)
	expectValue(t, "exception-reporter", "bound=true normal=true canceled=true")
}

func TestTimerDemoFixedIntervalModel(t *testing.T) {
	expectRunning(t, "fixed-interval-model", "immediate=true repeated=true")
}

func TestTimerDemoCalendarModel(t *testing.T) {
	expectRunning(t, "calendar-model", "interval=true calendar=false")
}

func TestTimerDemoImmediateRun(t *testing.T) {
	expectRunning(t, "immediate-run", "immediate=8 deferred=2")
}

func TestTimerDemoExecutionConcurrency(t *testing.T) {
	expectRunning(t, "execution-concurrency", "parallel=true serial=true")
}

func TestTimerDemoTimeParsing(t *testing.T) {
	expectValue(t, "time-parsing", "hhmm=ok hhmmss=ok invalid=panic")
}

func TestTimerDemoOffsetParsing(t *testing.T) {
	expectValue(t, "offset-parsing", "forms=ok invalid=panic")
}

func TestTimerDemoWeekdayParsing(t *testing.T) {
	expectValue(t, "weekday-parsing", "aliases=ok ranges=ok slices=ok invalid=panic")
}

func TestTimerDemoTimerType(t *testing.T) {
	expectValue(t, "timer-type", "interval=1m30s zero=true")
}

func TestTimerDemoResolvedCommand(t *testing.T) {
	expectValue(t, "resolved-command", "description=同步租户缓存 ran=true")
}

func TestTimerDemoCommandResolver(t *testing.T) {
	expectValue(t, "command-resolver", "name=overtime:detect args=--take=10 extra invalid=panic")
}

func TestTimerDemoNewSchedule(t *testing.T) {
	expectValue(t, "new-schedule", "empty=true tasks=1")
}

func TestTimerDemoName(t *testing.T) {
	expectValue(t, "name", "name=tenant_sync")
}

func TestTimerDemoDescription(t *testing.T) {
	expectValue(t, "description", "description=同步租户缓存")
}

func TestTimerDemoDefaults(t *testing.T) {
	expectRunning(t, "defaults", "immediate=true")
}

func TestTimerDemoStandalone(t *testing.T) {
	expectRunning(t, "standalone", "tasks=2 command=true call=true")
}

func TestTimerDemoRejectsUnknownScenario(t *testing.T) {
	_, err := timerdemo.Run("unknown")
	if err == nil || err.Error() != `timer demo unknown: unknown scenario "unknown"` {
		t.Fatalf("Run(unknown) error = %v, want unknown scenario error", err)
	}
}

func TestTimerDemoCatalogCoverage(t *testing.T) {
	implemented := catalog.Filter("timer", "", catalog.StatusImplemented)
	if len(implemented) != 62 {
		t.Fatalf("implemented timer entries = %d, want 62", len(implemented))
	}
	for _, slug := range []string{"architecture", "cron-command", "call", "every", "daily-at", "weekly-on", "days", "monthly", "quarterly-on", "yearly-on", "without-overlapping", "cross-process-overlap", "task-panic", "execution-concurrency", "weekday-parsing", "standalone"} {
		if item, ok := catalog.Find("timer", slug); !ok || item.Status != catalog.StatusImplemented {
			t.Fatalf("timer catalog entry %q = %#v, %v; want implemented", slug, item, ok)
		}
	}
	if planned := catalog.Filter("timer", "", catalog.StatusPlanned); len(planned) != 0 {
		t.Fatalf("planned timer entries = %d, want 0", len(planned))
	}
}
