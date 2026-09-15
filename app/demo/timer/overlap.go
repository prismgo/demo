package timerdemo

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prismgo/framework/cache"
	"github.com/prismgo/framework/timer"
)

// withoutOverlappingScenario proves a held lock skips execution and a released
// lock lets the same task run again.
func withoutOverlappingScenario() (string, error) {
	lock := cache.Lock("schedule:guarded-job", time.Minute)
	acquired, err := lock.Get(context.Background())
	if err != nil {
		return "", fmt.Errorf("acquire overlap lock: %w", err)
	}
	if !acquired {
		return "", errors.New("overlap lock was not acquired")
	}

	var blockedRuns atomic.Int64
	blockedSchedule := timer.NewSchedule()
	blockedSchedule.Call(func(context.Context) error {
		blockedRuns.Add(1)
		return nil
	}).Every(10 * time.Millisecond).WithoutOverlapping().Name("guarded-job")
	runSchedule(blockedSchedule, 60*time.Millisecond)
	blocked := blockedRuns.Load() == 0

	if err := lock.Release(context.Background()); err != nil {
		return "", fmt.Errorf("release overlap lock: %w", err)
	}

	var releasedRuns atomic.Int64
	releasedSchedule := timer.NewSchedule()
	releasedSchedule.Call(func(context.Context) error {
		releasedRuns.Add(1)
		return nil
	}).Every(10 * time.Millisecond).WithoutOverlapping().Name("guarded-job")
	runSchedule(releasedSchedule, 60*time.Millisecond)
	released := releasedRuns.Load() > 0

	invalid := expectPanic(
		func() { newTask().WithoutOverlapping(0) },
		func() { newTask().WithoutOverlapping(1, 2) },
	)
	return fmt.Sprintf("blocked=%t released=%t invalid=%s", blocked, released, invalid), nil
}

// crossProcessOverlapScenario simulates two processes sharing one overlap lock:
// the contender is skipped while the holder runs and resumes after it releases.
func crossProcessOverlapScenario() (string, error) {
	holder := timer.NewSchedule()
	contender := timer.NewSchedule()
	started := make(chan struct{})
	release := make(chan struct{})
	var startedOnce, releaseOnce sync.Once
	markStarted := func() { startedOnce.Do(func() { close(started) }) }
	open := func() { releaseOnce.Do(func() { close(release) }) }
	defer open()

	holder.Call(func(ctx context.Context) error {
		markStarted()
		select {
		case <-release:
		case <-ctx.Done():
		}
		return nil
	}).Every(5 * time.Millisecond).WithoutOverlapping().Name("cross-node-job")

	var contenderRuns atomic.Int64
	contender.Call(func(context.Context) error {
		contenderRuns.Add(1)
		return nil
	}).Every(5 * time.Millisecond).WithoutOverlapping().Name("cross-node-job")

	holderCtx, cancelHolder := context.WithCancel(context.Background())
	holder.Start(holderCtx)
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		cancelHolder()
		holder.Stop()
		return "", errors.New("overlap holder did not start")
	}

	contenderCtx, cancelContender := context.WithCancel(context.Background())
	contender.Start(contenderCtx)
	time.Sleep(60 * time.Millisecond)
	skipped := contenderRuns.Load() == 0

	open()
	cancelHolder()
	holder.Stop()
	time.Sleep(60 * time.Millisecond)
	ranAfterRelease := contenderRuns.Load() > 0

	cancelContender()
	contender.Stop()
	return fmt.Sprintf("skipped=%t ran-after-release=%t", skipped, ranAfterRelease), nil
}
