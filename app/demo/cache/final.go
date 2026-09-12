package cachedemo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/prismgo/framework/cache"
	"github.com/prismgo/framework/event"
)

func finalOperationResult(ctx context.Context, name, key string) (string, []string, error) {
	switch name {
	case "tags-redis":
		return taggedExample(ctx, "redis", key)
	case "tags-unsupported":
		err := cache.Store("file").Tags("demo").Put(ctx, key, "value", time.Minute)
		if !errors.Is(err, cache.ErrTagsUnsupported) {
			return "", nil, fmt.Errorf("file tagged put error = %v, want ErrTagsUnsupported", err)
		}
		return "ErrTagsUnsupported", nil, nil
	case "memo":
		return memoExample(ctx, key)
	case "failover":
		return failoverExample(ctx, key)
	case "custom-driver":
		return customDriverExample(ctx, key)
	case "resource-lifecycle":
		return lifecycleExample(ctx, key)
	case "events":
		return eventsExample(ctx, key)
	case "event-contract":
		return eventContractExample(ctx, key)
	case "deferred":
		return deferredExample(ctx, key)
	case "key-prefixes":
		return keyPrefixesExample(ctx, key)
	case "encoding":
		return encodingExample(ctx, key)
	case "errors":
		return errorsExample(ctx, key)
	case "memory-capabilities", "file-capabilities", "redis-capabilities", "failover-capabilities":
		return capabilitiesExample(ctx, strings.TrimSuffix(name, "-capabilities"), key)
	case "laravel-compatibility":
		return laravelExample(ctx, key)
	default:
		return "", nil, fmt.Errorf("unknown scenario %q", name)
	}
}

func taggedExample(ctx context.Context, store, key string) (string, []string, error) {
	tagged := cache.TagsFrom(store, "demo", "users")
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
	return fmt.Sprintf("before=%v; after_flush=%t", before, after), []string{"store:" + store}, err
}

func memoExample(ctx context.Context, key string) (string, []string, error) {
	repo := cache.Store("memory")
	if err := repo.Put(ctx, key, "first", time.Minute); err != nil {
		return "", nil, err
	}
	memo := repo.Memo()
	first, err := memo.Get(ctx, key)
	if err != nil {
		return "", nil, err
	}
	if err := repo.Put(ctx, key, "second", time.Minute); err != nil {
		return "", nil, err
	}
	cached, err := memo.Get(ctx, key)
	if err != nil {
		return "", nil, err
	}
	if err := memo.Put(ctx, key, "third", time.Minute); err != nil {
		return "", nil, err
	}
	updated, err := memo.Get(ctx, key)
	return fmt.Sprintf("first=%v; memo=%v; after_write=%v", first, cached, updated), nil, err
}

func failoverExample(ctx context.Context, key string) (result string, details []string, resultErr error) {
	var failedOver bool
	event.ListenFunc(cache.EventCacheFailedOver, func(_ context.Context, ev event.Event) error {
		payload, ok := ev.(cache.CacheEvent)
		if ok && payload.Store == "fallback" && payload.From == "absent" && errors.Is(payload.Error, cache.ErrStoreNotFound) {
			failedOver = true
		}
		return nil
	})
	defer event.Forget(cache.EventCacheFailedOver)
	m, err := cache.NewManager(cache.Config{Default: "fallback", Stores: map[string]cache.StoreConfig{
		"fallback": {Driver: "failover", Stores: []string{"absent", "memory"}},
		"memory":   {Driver: "memory"},
	}})
	if err != nil {
		return "", nil, err
	}
	defer func() { resultErr = errors.Join(resultErr, m.Close()) }()
	repo := m.Default()
	if err := repo.Put(ctx, key, "survived", time.Minute); err != nil {
		return "", nil, err
	}
	value, err := repo.Get(ctx, key)
	if err != nil {
		return "", nil, err
	}
	child, err := m.Store("memory").Get(ctx, key)
	if err != nil {
		return "", nil, err
	}
	if !failedOver {
		return "", nil, fmt.Errorf("failover event observed = false, want true")
	}
	return fmt.Sprintf("value=%v; fallback=%v; event=true", value, child), nil, nil
}

func eventsExample(ctx context.Context, key string) (string, []string, error) {
	var names []string
	for _, name := range []string{cache.EventCacheWriting, cache.EventCacheWritten, cache.EventCacheRetrieving, cache.EventCacheHit} {
		event.ListenFunc(name, func(_ context.Context, ev event.Event) error {
			payload, ok := ev.(cache.CacheEvent)
			if ok && payload.Key == key {
				names = append(names, payload.Name())
			}
			return nil
		})
		defer event.Forget(name)
	}
	if err := cache.PutFrom(ctx, "memory", key, "observed", time.Minute); err != nil {
		return "", nil, err
	}
	if _, err := cache.GetFrom[string](ctx, "memory", key); err != nil {
		return "", nil, err
	}
	return strings.Join(names, ","), nil, nil
}

func eventContractExample(ctx context.Context, key string) (string, []string, error) {
	var got cache.CacheEvent
	event.ListenFunc(cache.EventCacheWritten, func(_ context.Context, ev event.Event) error {
		payload, ok := ev.(cache.CacheEvent)
		if ok && payload.Key == key {
			got = payload
		}
		return nil
	})
	defer event.Forget(cache.EventCacheWritten)
	if err := cache.PutFrom(ctx, "memory", key, "payload", time.Minute); err != nil {
		return "", nil, err
	}
	if got.Name() != cache.EventCacheWritten || got.Store != "memory" || got.Key != key {
		return "", nil, fmt.Errorf("cache event = %#v, want written/memory/%s", got, key)
	}
	return fmt.Sprintf("event=%s; store=%s; key=%s", got.Name(), got.Store, got.Key), nil, nil
}

func deferredExample(ctx context.Context, key string) (string, []string, error) {
	if err := cache.ForgetFrom(ctx, "memory", key); err != nil {
		return "", nil, err
	}
	deferredCtx, run := cache.WithDeferred(ctx)
	loader := func(context.Context) (string, error) { return "refreshed", nil }
	if _, err := cache.FlexibleFrom(deferredCtx, "memory", key, cache.FlexibleWindow{Fresh: time.Millisecond, Stale: time.Minute}, loader); err != nil {
		return "", nil, err
	}
	select {
	case <-ctx.Done():
		return "", nil, ctx.Err()
	case <-time.After(3 * time.Millisecond):
	}
	stale, err := cache.FlexibleFrom(deferredCtx, "memory", key, cache.FlexibleWindow{Fresh: time.Millisecond, Stale: time.Minute}, func(context.Context) (string, error) { return "updated", nil })
	if err != nil {
		return "", nil, err
	}
	run()
	deadline := time.Now().Add(time.Second)
	for {
		value, err := cache.FlexibleFrom(ctx, "memory", key, cache.FlexibleWindow{Fresh: time.Hour, Stale: 2 * time.Hour}, loader)
		if err != nil {
			return "", nil, err
		}
		if value == "updated" {
			return fmt.Sprintf("stale=%s; after_deferred=%s", stale, value), nil, nil
		}
		if time.Now().After(deadline) {
			return "", nil, fmt.Errorf("deferred cache value = %q, want updated", value)
		}
		select {
		case <-ctx.Done():
			return "", nil, ctx.Err()
		case <-time.After(time.Millisecond):
		}
	}
}

func keyPrefixesExample(ctx context.Context, key string) (string, []string, error) {
	repo := cache.Store("memory")
	prefix := repo.GetStore().Prefix()
	if err := repo.Put(ctx, key, "prefixed", time.Minute); err != nil {
		return "", nil, err
	}
	if !strings.HasSuffix(prefix, ":memory") {
		return "", nil, fmt.Errorf("cache prefix = %q, want suffix :memory", prefix)
	}
	return "cache=" + prefix + "; lock=" + prefix + ":locks", nil, nil
}

func encodingExample(ctx context.Context, key string) (result string, details []string, resultErr error) {
	m, store, err := customManager()
	if err != nil {
		return "", nil, err
	}
	defer func() { resultErr = errors.Join(resultErr, m.Close()) }()
	repo := m.Default()
	for _, value := range []any{int64(42), map[string]string{"name": "Ada"}} {
		if err := repo.Put(ctx, key+fmt.Sprintf(":%T", value), value, time.Minute); err != nil {
			return "", nil, err
		}
	}
	integerPayload, err := store.Get(ctx, "demo:tenant:"+key+":int64")
	if err != nil {
		return "", nil, err
	}
	objectPayload, err := store.Get(ctx, "demo:tenant:"+key+":map[string]string")
	if err != nil {
		return "", nil, err
	}
	integerBytes, integerOK := integerPayload.([]byte)
	objectBytes, objectOK := objectPayload.([]byte)
	if !integerOK || string(integerBytes) != "42" || !objectOK || len(objectBytes) == 0 || objectBytes[0] == '{' {
		return "", nil, fmt.Errorf("encoded payloads integer=%#v object=%#v, want decimal ASCII and structured codec bytes", integerPayload, objectPayload)
	}
	number, err := repo.Get(ctx, key+":int64")
	if err != nil {
		return "", nil, err
	}
	object, err := repo.Get(ctx, key+":map[string]string")
	if err != nil {
		return "", nil, err
	}
	fields, ok := object.(map[string]any)
	if !ok {
		return "", nil, fmt.Errorf("decoded object = %T, want map[string]any", object)
	}
	return fmt.Sprintf("number=%v; name=%v", number, fields["name"]), []string{"numeric:decimal", "structured:codec"}, nil
}

func errorsExample(ctx context.Context, key string) (result string, details []string, resultErr error) {
	_, miss := cache.GetFrom[string](ctx, "memory", key+":missing")
	_, absent := cache.Store("absent").Get(ctx, key)
	unsupported := cache.Store("file").Tags("demo").Put(ctx, key, "value", time.Minute)
	lock := cache.LockFrom("memory", key+":held", time.Minute)
	got, err := lock.Get(ctx)
	if err != nil || !got {
		return "", nil, fmt.Errorf("acquire error demo lock got=%t error=%v, want true/nil", got, err)
	}
	defer func() { resultErr = errors.Join(resultErr, lock.Release(context.WithoutCancel(ctx))) }()
	_, timeout := cache.LockFrom("memory", key+":held", time.Minute).Block(ctx, time.Millisecond, nil)
	notHeld := cache.LockFrom("memory", key+":free", time.Minute).Release(ctx)
	if err := cache.PutFrom(ctx, "memory", key+":counter", "not-integer", time.Minute); err != nil {
		return "", nil, err
	}
	_, invalid := cache.IncrementFrom(ctx, "memory", key+":counter")
	if !errors.Is(miss, cache.ErrCacheMiss) || !errors.Is(absent, cache.ErrStoreNotFound) ||
		!errors.Is(unsupported, cache.ErrTagsUnsupported) || !errors.Is(timeout, cache.ErrLockTimeout) ||
		!errors.Is(notHeld, cache.ErrLockNotHeld) || !errors.Is(invalid, cache.ErrInvalidCounter) {
		return "", nil, fmt.Errorf("cache errors miss=%v absent=%v unsupported=%v timeout=%v not_held=%v invalid=%v, want respective sentinels", miss, absent, unsupported, timeout, notHeld, invalid)
	}
	return "ErrCacheMiss,ErrStoreNotFound,ErrTagsUnsupported,ErrLockTimeout,ErrLockNotHeld,ErrInvalidCounter", nil, nil
}

func capabilitiesExample(ctx context.Context, store, key string) (result string, details []string, resultErr error) {
	repo := cache.Store(store)
	if store == "failover" {
		// The configured failover starts with Redis, so use an isolated manager with local children.
		m, err := cache.NewManager(cache.Config{Default: "failover", Stores: map[string]cache.StoreConfig{
			"failover": {Driver: "failover", Stores: []string{"memory"}},
			"memory":   {Driver: "memory"},
		}})
		if err != nil {
			return "", nil, err
		}
		defer func() { resultErr = errors.Join(resultErr, m.Close()) }()
		repo = m.Default()
	}
	if err := repo.Put(ctx, key, "value", time.Minute); err != nil {
		return "", nil, err
	}
	touched, err := repo.Touch(ctx, key, time.Minute)
	if err != nil {
		return "", nil, err
	}
	added, err := repo.Add(ctx, key+":once", "once", time.Minute)
	if err != nil {
		return "", nil, err
	}
	if err := repo.PutMany(ctx, map[string]any{key + ":bulk": "bulk"}, time.Minute); err != nil {
		return "", nil, err
	}
	bulk, err := repo.Many(ctx, []string{key + ":bulk"})
	if err != nil {
		return "", nil, err
	}
	lock := repo.Lock(key+":lock", time.Minute)
	locked, err := lock.Get(ctx)
	if err != nil {
		return "", nil, err
	}
	if locked {
		defer func() { resultErr = errors.Join(resultErr, lock.Release(context.WithoutCancel(ctx))) }()
	}
	if !touched || !added || bulk[key+":bulk"] != "bulk" || !locked || !repo.SupportsFlushingLocks() {
		return "", nil, fmt.Errorf("%s capabilities touch=%t add=%t bulk=%v lock=%t lock_flush=%t, want all supported", store, touched, added, bulk, locked, repo.SupportsFlushingLocks())
	}
	if store == "file" {
		if repo.SupportsTags() {
			return "", nil, fmt.Errorf("file SupportsTags = true, want false")
		}
	} else if !repo.SupportsTags() {
		return "", nil, fmt.Errorf("%s SupportsTags = false, want true", store)
	}
	return fmt.Sprintf("store=%s; touch=true; atomic=true; bulk=true; lock=true; tags=%t", store, repo.SupportsTags()), nil, nil
}

func laravelExample(ctx context.Context, key string) (string, []string, error) {
	if err := cache.PutFrom(ctx, "memory", key, "prismgo", time.Minute); err != nil {
		return "", nil, err
	}
	value, err := cache.GetFrom[string](ctx, "memory", key)
	return fmt.Sprintf("Cache::store/get/put -> %s", value), []string{"cache.Store", "cache.GetFrom", "cache.PutFrom"}, err
}
