package timerdemo

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prismgo/framework/timer"
)

// fixedIntervalModelScenario proves Every runs immediately and repeats.
func fixedIntervalModelScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("health-check").Every(15 * time.Millisecond)
	counts := recorder.run(80 * time.Millisecond)
	runs := counts["health-check"]
	return fmt.Sprintf("immediate=%t repeated=%t", runs >= 1, runs >= 2), nil
}

// calendarModelScenario contrasts the interval mode with a deferred calendar hit.
func calendarModelScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("interval-model").Every(15 * time.Millisecond)
	recorder.task("calendar-model").DailyAt("23:59")
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("interval=%t calendar=%t", counts["interval-model"] > 0, counts["calendar-model"] > 0), nil
}

// immediateRunScenario counts the documented helpers that run on start.
func immediateRunScenario() (string, error) {
	recorder := newRecorder()
	immediateNames := []string{
		"every-minute", "every-two-minutes", "every-five-minutes", "every-ten-minutes",
		"every-fifteen-minutes", "every-thirty-minutes", "hourly", "daily",
	}
	recorder.task("every-minute").EveryMinute()
	recorder.task("every-two-minutes").EveryTwoMinutes()
	recorder.task("every-five-minutes").EveryFiveMinutes()
	recorder.task("every-ten-minutes").EveryTenMinutes()
	recorder.task("every-fifteen-minutes").EveryFifteenMinutes()
	recorder.task("every-thirty-minutes").EveryThirtyMinutes()
	recorder.task("hourly").Hourly()
	recorder.task("daily").Daily()

	deferredNames := []string{"every-three-minutes", "every-four-minutes"}
	recorder.task("every-three-minutes").EveryThreeMinutes()
	recorder.task("every-four-minutes").EveryFourMinutes()
	counts := recorder.run(80 * time.Millisecond)

	immediate := 0
	for _, name := range immediateNames {
		if counts[name] > 0 {
			immediate++
		}
	}
	deferred := 0
	for _, name := range deferredNames {
		if counts[name] == 0 {
			deferred++
		}
	}
	return fmt.Sprintf("immediate=%d deferred=%d", immediate, deferred), nil
}

// executionConcurrencyScenario proves distinct tasks run in parallel and a single
// task never overlaps itself.
func executionConcurrencyScenario() (string, error) {
	release := make(chan struct{})
	var releaseOnce sync.Once
	open := func() { releaseOnce.Do(func() { close(release) }) }
	defer open()

	schedule := timer.NewSchedule()
	var started atomic.Int64
	for _, name := range []string{"worker-a", "worker-b"} {
		schedule.Call(func(ctx context.Context) error {
			started.Add(1)
			select {
			case <-release:
			case <-ctx.Done():
			}
			return nil
		}).Every(5 * time.Millisecond).Name(name)
	}
	ctx, cancel := context.WithCancel(context.Background())
	schedule.Start(ctx)
	parallel := waitUntil(func() bool { return started.Load() >= 2 }, time.Second)
	open()
	cancel()
	schedule.Stop()

	serialSchedule := timer.NewSchedule()
	var active, maxActive atomic.Int64
	serialSchedule.Call(func(ctx context.Context) error {
		current := active.Add(1)
		for {
			observed := maxActive.Load()
			if current <= observed || maxActive.CompareAndSwap(observed, current) {
				break
			}
		}
		select {
		case <-time.After(15 * time.Millisecond):
		case <-ctx.Done():
		}
		active.Add(-1)
		return nil
	}).Every(2 * time.Millisecond).Name("serial-job")
	serialCtx, serialCancel := context.WithCancel(context.Background())
	serialSchedule.Start(serialCtx)
	time.Sleep(80 * time.Millisecond)
	serialCancel()
	serialSchedule.Stop()
	serial := maxActive.Load() == 1

	return fmt.Sprintf("parallel=%t serial=%t", parallel, serial), nil
}
