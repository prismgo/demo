package sessiondemo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/prismgo/framework/session"
	"github.com/redis/go-redis/v9"
)

// RunRedis executes a session catalog scenario that needs a real Redis service.
func RunRedis(name string) (Result, error) {
	value, err := runRedis(name)
	if err != nil {
		return Result{}, fmt.Errorf("session demo %s: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

func runRedis(name string) (string, error) {
	switch name {
	case "redis-driver":
		return redisDriverScenario()
	case "redis-lock":
		return redisLockScenario()
	default:
		return "", fmt.Errorf("unknown redis scenario %q", name)
	}
}

// redisDriverScenario verifies shared reads, prefix isolation, TTL, GC, and recovery
// against the Redis instance named by PRISMGO_REDIS_TEST_URL.
func redisDriverScenario() (string, error) {
	rawURL := strings.TrimSpace(os.Getenv("PRISMGO_REDIS_TEST_URL"))
	if rawURL == "" {
		return "", errors.New("PRISMGO_REDIS_TEST_URL is required for the redis-driver scenario")
	}
	opts, err := redis.ParseURL(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse PRISMGO_REDIS_TEST_URL: %w", err)
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return "", fmt.Errorf("ping redis: %w", err)
	}
	prefix := fmt.Sprintf("prismgo_demo_session_%d", time.Now().UnixNano())
	cfg := session.DefaultConfig()
	cfg.Redis.Prefix = prefix
	cfg.Lifetime = 5 * time.Minute
	instances := make([]*session.RedisDriver, 0, 2)
	managers := make([]*session.Manager, 0, 2)
	for range 2 {
		driver, err := session.NewRedisDriverFromClient(client, cfg)
		if err != nil {
			cleanupRedis(ctx, client, prefix, instances)
			return "", err
		}
		manager, err := session.NewManager(cfg, driver)
		if err != nil {
			cleanupRedis(ctx, client, prefix, instances)
			return "", err
		}
		instances = append(instances, driver)
		managers = append(managers, manager)
	}
	defer cleanupRedis(ctx, client, prefix, instances)

	// 实例 A 写入并保存，实例 B 通过同一 session ID 恢复，验证跨实例共享。
	storeA := freshStore(managers[0])
	storeA.Put("user_id", int64(1001))
	if err := storeA.Save(ctx); err != nil {
		return "", fmt.Errorf("save via instance A: %w", err)
	}
	restoredB, err := restoreSession(managers[1], storeA.ID())
	if err != nil {
		return "", fmt.Errorf("restore via instance B: %w", err)
	}
	shared := fmt.Sprint(restoredB.Get("user_id")) == "1001"

	// TTL 由 expiresAt 控制，key 应带剩余寿命。
	ttl, err := client.TTL(ctx, redisSessionKey(prefix, storeA.ID())).Result()
	if err != nil {
		return "", fmt.Errorf("ttl session key: %w", err)
	}

	// GC 对 Redis 是 no-op：过期数据依赖 TTL，key 在 GC 后仍可读取。
	if err := instances[0].GC(ctx, time.Now().Add(time.Hour)); err != nil {
		return "", fmt.Errorf("redis gc: %w", err)
	}
	gcKept := client.Exists(ctx, redisSessionKey(prefix, storeA.ID())).Val() == 1

	// 前缀隔离：同 ID 换前缀的 driver 读不到对方数据。
	otherCfg := cfg
	otherCfg.Redis.Prefix = prefix + "_other"
	otherDriver, err := session.NewRedisDriverFromClient(client, otherCfg)
	if err != nil {
		return "", fmt.Errorf("build isolated driver: %w", err)
	}
	otherManager, err := session.NewManager(otherCfg, otherDriver)
	if err != nil {
		return "", fmt.Errorf("build isolated manager: %w", err)
	}
	otherStore, err := restoreSession(otherManager, storeA.ID())
	if err != nil {
		return "", fmt.Errorf("restore with isolated manager: %w", err)
	}
	isolated := otherStore.ID() != storeA.ID()

	// 损坏 payload 走可恢复错误：降级为新 session。
	if err := client.Set(ctx, redisSessionKey(prefix, storeA.ID()), []byte("{{{not-a-session-payload"), time.Hour).Err(); err != nil {
		return "", fmt.Errorf("corrupt session key: %w", err)
	}
	corruptedStore, err := restoreSession(managers[0], storeA.ID())
	if err != nil {
		return "", fmt.Errorf("restore corrupted session: %w", err)
	}
	corruptFresh := corruptedStore.ID() != storeA.ID()

	return fmt.Sprintf("shared=%t prefix=%s ttl=%ds gc-keeps=%t isolated=%t corrupt-fresh=%t",
		shared, prefix, int(ttl.Seconds()), gcKept, isolated, corruptFresh), nil
}

// redisSessionKey renders the payload key the redis driver stores for one session ID.
func redisSessionKey(prefix, id string) string {
	return prefix + ":sessions:" + id
}

// redisLockScenario verifies Redis lock ownership, contention, release, and
// expiration takeover against the instance named by PRISMGO_REDIS_TEST_URL.
func redisLockScenario() (string, error) {
	rawURL := strings.TrimSpace(os.Getenv("PRISMGO_REDIS_TEST_URL"))
	if rawURL == "" {
		return "", errors.New("PRISMGO_REDIS_TEST_URL is required for the redis-lock scenario")
	}
	opts, err := redis.ParseURL(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse PRISMGO_REDIS_TEST_URL: %w", err)
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return "", fmt.Errorf("ping redis: %w", err)
	}
	prefix := fmt.Sprintf("prismgo_demo_session_lock_%d", time.Now().UnixNano())
	cfg := session.DefaultConfig()
	cfg.Redis.Prefix = prefix
	driver, err := session.NewRedisDriverFromClient(client, cfg)
	if err != nil {
		cleanupRedis(ctx, client, prefix, nil)
		return "", err
	}
	defer cleanupRedis(ctx, client, prefix, []*session.RedisDriver{driver})

	id := newDemoSessionID()
	lock, err := driver.Lock(ctx, id, 10*time.Second, 200*time.Millisecond)
	if err != nil {
		return "", err
	}
	_, contendErr := driver.Lock(ctx, id, 10*time.Second, 100*time.Millisecond)
	contention := errors.Is(contendErr, session.ErrLockTimeout)
	if err := lock.Release(ctx); err != nil {
		return "", err
	}
	doubleRelease := errors.Is(lock.Release(ctx), session.ErrLockNotHeld)

	stale, err := driver.Lock(ctx, id, 100*time.Millisecond, 200*time.Millisecond)
	if err != nil {
		return "", err
	}
	time.Sleep(300 * time.Millisecond)
	fresh, err := driver.Lock(ctx, id, 10*time.Second, 200*time.Millisecond)
	if err != nil {
		return "", err
	}
	staleRelease := errors.Is(stale.Release(ctx), session.ErrLockNotHeld)
	if err := fresh.Release(ctx); err != nil {
		return "", err
	}
	return fmt.Sprintf("contention=%t released=true re-release-not-held=%t expired-takeover=true stale-release-not-held=%t",
		contention, doubleRelease, staleRelease), nil
}

// cleanupRedis removes every key created under the demo prefix and closes the client.
func cleanupRedis(ctx context.Context, client *redis.Client, prefix string, drivers []*session.RedisDriver) {
	for _, driver := range drivers {
		_ = driver.Close()
	}
	keys, err := client.Keys(ctx, prefix+"*").Result()
	if err == nil && len(keys) > 0 {
		_ = client.Del(ctx, keys...).Err()
	}
	_ = client.Close()
}
