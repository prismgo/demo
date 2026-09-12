package cachedemo

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/prismgo/framework/cache"
)

func advancedOperationResult(ctx context.Context, name, key string) (string, []string, error) {
	switch name {
	case "add":
		if err := cache.Forget(ctx, key); err != nil {
			return "", nil, err
		}
		first, err := cache.Add(ctx, key, "first", time.Minute)
		if err != nil {
			return "", nil, err
		}
		second, err := cache.Add(ctx, key, "second", time.Minute)
		if err != nil {
			return "", nil, err
		}
		value, err := cache.Get[string](ctx, key)
		return fmt.Sprintf("first=%t; second=%t; value=%s", first, second, value), nil, err
	case "put-many":
		if err := cache.PutMany(ctx, map[string]any{key + ":a": "alpha"}, time.Minute); err != nil {
			return "", nil, err
		}
		if err := cache.SetMultiple(ctx, map[string]any{key + ":b": "beta"}, time.Minute); err != nil {
			return "", nil, err
		}
		values, err := cache.Many[string](ctx, []string{key + ":a", key + ":b"})
		return fmt.Sprintf("%s/%s", values[key+":a"], values[key+":b"]), []string{"aliases:PutMany,SetMultiple"}, err
	case "remember":
		if err := cache.Forget(ctx, key); err != nil {
			return "", nil, err
		}
		calls := 0
		loader := func(context.Context) (string, error) { calls++; return "loaded", nil }
		first, err := cache.Remember(ctx, key, time.Minute, loader)
		if err != nil {
			return "", nil, err
		}
		second, err := cache.Remember(ctx, key, time.Minute, loader)
		return fmt.Sprintf("%s/%s; loader_calls=%d", first, second, calls), nil, err
	case "remember-forever":
		if err := cache.Forget(ctx, key); err != nil {
			return "", nil, err
		}
		calls := 0
		loader := func(context.Context) (string, error) { calls++; return "permanent", nil }
		first, err := cache.RememberForever(ctx, key, loader)
		if err != nil {
			return "", nil, err
		}
		second, err := cache.Sear(ctx, key, loader)
		return fmt.Sprintf("%s/%s; loader_calls=%d", first, second, calls), []string{"aliases:RememberForever,Sear"}, err
	case "flexible":
		return flexibleExample(ctx, key)
	case "touch":
		if err := cache.Put(ctx, key, "retained", time.Second); err != nil {
			return "", nil, err
		}
		updated, err := cache.Touch(ctx, key, time.Minute)
		if err != nil {
			return "", nil, err
		}
		select {
		case <-ctx.Done():
			return "", nil, ctx.Err()
		case <-time.After(1100 * time.Millisecond):
		}
		value, err := cache.Get[string](ctx, key)
		if err != nil {
			return "", nil, err
		}
		missing, err := cache.Touch(ctx, key+":missing", time.Minute)
		return fmt.Sprintf("updated=%t; value=%s; missing=%t", updated, value, missing), nil, err
	case "many":
		if err := cache.Put(ctx, key+":present", "one", time.Minute); err != nil {
			return "", nil, err
		}
		if err := cache.Forget(ctx, key+":missing"); err != nil {
			return "", nil, err
		}
		keys := []string{key + ":present", key + ":missing"}
		values, err := cache.GetMultiple[string](ctx, keys, cache.Value("fallback"))
		return fmt.Sprintf("present=%s; missing=%s", values[key+":present"], values[key+":missing"]), []string{"alias:GetMultiple"}, err
	case "pull":
		if err := cache.Put(ctx, key, "taken", time.Minute); err != nil {
			return "", nil, err
		}
		value, err := cache.Pull[string](ctx, key)
		if err != nil {
			return "", nil, err
		}
		after, err := cache.Has(ctx, key)
		return fmt.Sprintf("value=%s; remains=%t", value, after), nil, err
	case "forget":
		if err := cache.Put(ctx, key, "remove", time.Minute); err != nil {
			return "", nil, err
		}
		if err := cache.Forget(ctx, key); err != nil {
			return "", nil, err
		}
		first, err := cache.Has(ctx, key)
		if err != nil {
			return "", nil, err
		}
		if err := cache.Put(ctx, key, "remove-again", time.Minute); err != nil {
			return "", nil, err
		}
		if err := cache.Delete(ctx, key); err != nil {
			return "", nil, err
		}
		second, err := cache.Has(ctx, key)
		return fmt.Sprintf("after_forget=%t; after_delete=%t", first, second), nil, err
	case "forget-many":
		keys := []string{key + ":a", key + ":b"}
		if err := cache.PutMany(ctx, map[string]any{keys[0]: "a", keys[1]: "b"}, time.Minute); err != nil {
			return "", nil, err
		}
		if err := cache.ForgetMany(ctx, keys[:1]); err != nil {
			return "", nil, err
		}
		if err := cache.DeleteMultiple(ctx, keys[1:]); err != nil {
			return "", nil, err
		}
		values, err := cache.Many[string](ctx, keys)
		return fmt.Sprintf("a=%q; b=%q", values[keys[0]], values[keys[1]]), []string{"aliases:ForgetMany,DeleteMultiple"}, err
	case "flush":
		return flushExample(ctx, key)
	case "counters":
		if err := cache.Forget(ctx, key); err != nil {
			return "", nil, err
		}
		first, err := cache.Increment(ctx, key)
		if err != nil {
			return "", nil, err
		}
		second, err := cache.Increment(ctx, key, 4)
		if err != nil {
			return "", nil, err
		}
		third, err := cache.Decrement(ctx, key, 2)
		return fmt.Sprintf("%d -> %d -> %d", first, second, third), nil, err
	case "lock":
		return lockExample(ctx, key)
	case "lock-callback":
		return lockCallbackExample(ctx, key)
	case "lock-block":
		return lockBlockExample(ctx, key)
	case "lock-restore":
		return lockRestoreExample(ctx, key)
	case "lock-flush":
		return lockFlushExample(ctx, key)
	case "funnel":
		return funnelExample(ctx, key)
	case "without-overlapping":
		return overlapExample(ctx, key)
	case "tags-memory":
		return memoryTagsExample(ctx, key)
	default:
		return finalOperationResult(ctx, name, key)
	}
}

func flexibleExample(ctx context.Context, key string) (string, []string, error) {
	if err := cache.Forget(ctx, key); err != nil {
		return "", nil, err
	}
	deferredCtx, runDeferred := cache.WithDeferred(ctx)
	var calls atomic.Int32
	loader := func(context.Context) (string, error) { return fmt.Sprintf("version-%d", calls.Add(1)), nil }
	window := cache.FlexibleWindow{Fresh: 10 * time.Millisecond, Stale: time.Minute}
	first, err := cache.Flexible(deferredCtx, key, window, loader)
	if err != nil {
		return "", nil, err
	}
	select {
	case <-ctx.Done():
		return "", nil, ctx.Err()
	case <-time.After(20 * time.Millisecond):
	}
	stale, err := cache.Flexible(deferredCtx, key, window, loader)
	if err != nil {
		return "", nil, err
	}
	before := calls.Load()
	runDeferred()
	deadline := time.Now().Add(time.Second)
	var updated string
	for {
		updated, err = cache.Flexible(ctx, key, cache.FlexibleWindow{Fresh: time.Hour, Stale: 2 * time.Hour}, loader)
		if err != nil {
			return "", nil, err
		}
		if updated == "version-2" {
			break
		}
		if time.Now().After(deadline) {
			return "", nil, fmt.Errorf("deferred refresh value = %q, want version-2", updated)
		}
		select {
		case <-ctx.Done():
			return "", nil, ctx.Err()
		case <-time.After(time.Millisecond):
		}
	}
	return fmt.Sprintf("first=%s; stale=%s; refreshed=%s", first, stale, updated), []string{fmt.Sprintf("calls_before_deferred:%d", before), fmt.Sprintf("calls_after_deferred:%d", calls.Load())}, nil
}

func flushExample(ctx context.Context, key string) (string, []string, error) {
	if err := cache.Put(ctx, key, "clear-me", time.Minute); err != nil {
		return "", nil, err
	}
	if err := cache.Flush(ctx); err != nil {
		return "", nil, err
	}
	afterFlush, err := cache.Has(ctx, key)
	if err != nil {
		return "", nil, err
	}
	if err := cache.Put(ctx, key, "clear-again", time.Minute); err != nil {
		return "", nil, err
	}
	if err := cache.Clear(ctx); err != nil {
		return "", nil, err
	}
	afterClear, err := cache.Has(ctx, key)
	return fmt.Sprintf("after_flush=%t; after_clear=%t", afterFlush, afterClear), []string{"scope:current-store-prefix"}, err
}

func lockExample(ctx context.Context, key string) (string, []string, error) {
	lock := cache.Lock(key, time.Minute)
	got, err := lock.Get(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("acquire lock: %w", err)
	}
	if !got {
		return "", nil, fmt.Errorf("acquire lock: got false, want true")
	}
	contender, err := cache.Lock(key, time.Minute).Get(ctx)
	if err != nil {
		return "", nil, err
	}
	if err := lock.Release(ctx); err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("acquired=%t; contender=%t; released=true", got, contender), nil, nil
}

func lockCallbackExample(ctx context.Context, key string) (string, []string, error) {
	called := false
	got, err := cache.Lock(key, time.Minute).Get(ctx, func(context.Context) error { called = true; return nil })
	if err != nil {
		return "", nil, err
	}
	again := cache.Lock(key, time.Minute)
	reacquired, err := again.Get(ctx)
	if err != nil {
		return "", nil, err
	}
	if reacquired {
		if err := again.Release(ctx); err != nil {
			return "", nil, err
		}
	}
	return fmt.Sprintf("acquired=%t; callback=%t; reacquired=%t", got, called, reacquired), nil, nil
}

func lockBlockExample(ctx context.Context, key string) (string, []string, error) {
	held := cache.Lock(key, time.Minute)
	got, err := held.Get(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("acquire blocking example lock: %w", err)
	}
	if !got {
		return "", nil, fmt.Errorf("acquire blocking example lock: got false, want true")
	}
	_, timeoutErr := cache.Lock(key, time.Minute).Block(ctx, 5*time.Millisecond, nil)
	if err := held.Release(ctx); err != nil {
		return "", nil, err
	}
	if !errors.Is(timeoutErr, cache.ErrLockTimeout) {
		return "", nil, fmt.Errorf("blocked lock error = %v, want ErrLockTimeout", timeoutErr)
	}
	called := false
	acquired, err := cache.Lock(key, time.Minute).Block(ctx, time.Second, func(context.Context) error { called = true; return nil })
	return fmt.Sprintf("timeout=true; acquired=%t; callback=%t", acquired, called), nil, err
}

func lockRestoreExample(ctx context.Context, key string) (string, []string, error) {
	lock := cache.Lock(key, time.Minute)
	got, err := lock.Get(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("acquire restorable lock: %w", err)
	}
	if !got {
		return "", nil, fmt.Errorf("acquire restorable lock: got false, want true")
	}
	owner := lock.Owner()
	if err := cache.RestoreLock(key, owner).Release(ctx); err != nil {
		return "", nil, err
	}
	other := cache.Lock(key, time.Minute)
	got, err = other.Get(ctx)
	if err != nil {
		return "", nil, err
	}
	if got {
		if err := other.ForceRelease(ctx); err != nil {
			return "", nil, err
		}
	}
	return fmt.Sprintf("restored=true; reacquired=%t; force_released=%t", got, got), []string{"owner:" + owner}, nil
}

func lockFlushExample(ctx context.Context, key string) (string, []string, error) {
	lock := cache.Lock(key, time.Minute)
	got, err := lock.Get(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("acquire flushable lock: %w", err)
	}
	if !got {
		return "", nil, fmt.Errorf("acquire flushable lock: got false, want true")
	}
	if err := cache.FlushLocks(ctx); err != nil {
		return "", nil, err
	}
	again := cache.Lock(key, time.Minute)
	reacquired, err := again.Get(ctx)
	if err != nil {
		return "", nil, err
	}
	if reacquired {
		if err := again.Release(ctx); err != nil {
			return "", nil, err
		}
	}
	return fmt.Sprintf("flushed=true; reacquired=%t", reacquired), nil, nil
}

func funnelExample(ctx context.Context, key string) (string, []string, error) {
	failed := false
	outer, err := cache.Funnel(key).Limit(1).Then(ctx, func(ctx context.Context) error {
		entered, err := cache.Funnel(key).Limit(1).Then(ctx, func(context.Context) error { return nil }, func(context.Context) error {
			failed = true
			return nil
		})
		if err != nil {
			return err
		}
		if entered {
			return fmt.Errorf("occupied funnel entered = true, want false")
		}
		return nil
	})
	if err != nil {
		return "", nil, err
	}
	if !outer {
		return "", nil, fmt.Errorf("outer funnel entered = false, want true")
	}
	entered, err := cache.Funnel(key).Limit(1).Then(ctx, func(context.Context) error { return nil })
	return fmt.Sprintf("failure_callback=%t; entered_after_release=%t", failed, entered), nil, err
}

func overlapExample(ctx context.Context, key string) (string, []string, error) {
	outer, err := cache.WithoutOverlapping(ctx, key, func(ctx context.Context) error {
		entered, timeoutErr := cache.WithoutOverlapping(ctx, key, func(context.Context) error { return nil }, cache.WithOverlapWait(5*time.Millisecond))
		if entered || !errors.Is(timeoutErr, cache.ErrLockTimeout) {
			return fmt.Errorf("nested overlap = %t, error = %v, want false and ErrLockTimeout", entered, timeoutErr)
		}
		return nil
	})
	if err != nil {
		return "", nil, err
	}
	if !outer {
		return "", nil, fmt.Errorf("outer overlap entered = false, want true")
	}
	called := false
	entered, err := cache.WithoutOverlapping(ctx, key, func(context.Context) error { called = true; return nil })
	return fmt.Sprintf("overlap_blocked=true; entered=%t; callback=%t", entered, called), nil, err
}

func memoryTagsExample(ctx context.Context, key string) (string, []string, error) {
	tagged := cache.Store("memory").Tags("demo", "users")
	if err := tagged.Put(ctx, key, "tagged", time.Minute); err != nil {
		return "", nil, err
	}
	before, err := tagged.Get(ctx, key)
	if err != nil {
		return "", nil, err
	}
	if err := tagged.Flush(ctx); err != nil {
		return "", nil, err
	}
	after, err := tagged.Has(ctx, key)
	return fmt.Sprintf("before=%v; after_flush=%t", before, after), []string{"store:memory", "tags:demo,users"}, err
}
