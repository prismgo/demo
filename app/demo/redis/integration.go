package redisdemo

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	rediscontract "github.com/prismgo/framework/contracts/redis"
	"github.com/prismgo/framework/event"
	"github.com/prismgo/framework/foundation"
	"github.com/prismgo/framework/redis"
	goredis "github.com/redis/go-redis/v9"
)

// RunIntegration executes a Redis catalog scenario that needs a real Redis service.
func RunIntegration(name string) (Result, error) {
	value, err := runIntegration(name)
	if err != nil {
		return Result{}, fmt.Errorf("redis demo %s: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

func runIntegration(name string) (string, error) {
	switch name {
	case "client":
		return defaultClientScenario()
	case "named-client":
		return namedClientScenario()
	case "connection":
		return connectionScenario()
	case "strings":
		return withLiveClient(stringCommands)
	case "hashes":
		return withLiveClient(hashCommands)
	case "lists":
		return withLiveClient(listCommands)
	case "sets":
		return withLiveClient(setCommands)
	case "sorted-sets":
		return withLiveClient(sortedSetCommands)
	case "counters":
		return withLiveClient(counterCommands)
	case "keys":
		return withLiveClient(keyCommands)
	case "transaction":
		return withLiveClient(transactionScenario)
	case "lua":
		return withLiveClient(luaScenario)
	case "pipeline":
		return withLiveClient(pipelineScenario)
	case "publish", "subscribe", "psubscribe":
		return pubSubScenario(name)
	case "default-connection":
		return defaultConnectionScenario()
	case "named-connection":
		return namedConnectionScenario()
	case "default-connection-method":
		return defaultConnectionMethodScenario()
	case "multiple-connections":
		return multipleConnectionsScenario()
	case "purge-rebuild":
		return purgeRebuildScenario()
	case "command-executed-event":
		return commandExecutedEventScenario()
	case "command-failed-event":
		return commandFailedEventScenario()
	case "batch-executed-event":
		return batchExecutedEventScenario()
	case "batch-failed-event":
		return batchFailedEventScenario()
	case "event-payloads":
		return eventPayloadsScenario()
	case "global-command-listener":
		return globalCommandListenerScenario()
	case "global-failure-listener":
		return globalFailureListenerScenario()
	case "global-batch-listener":
		return globalBatchListenerScenario()
	case "connection-listener":
		return connectionListenerScenario()
	case "connection-failure-listener":
		return connectionFailureListenerScenario()
	case "listener-panic":
		return listenerPanicScenario()
	case "disable-events":
		return disableEventsScenario()
	case "enable-events":
		return enableEventsScenario()
	case "cache-driver":
		return cacheDriverScenario()
	case "cache-basic":
		return cacheBasicScenario()
	case "cache-ttl":
		return cacheTTLScenario()
	case "cache-atomic":
		return cacheAtomicScenario()
	case "cache-bulk":
		return cacheBulkScenario()
	case "cache-tags":
		return cacheTagsScenario()
	case "cache-flush":
		return cacheFlushScenario()
	case "queue-driver":
		return queueDriverScenario()
	case "queue-ready":
		return queueReadyScenario()
	case "queue-delayed":
		return queueDelayedScenario()
	case "queue-blocking-pop":
		return queueBlockingPopScenario()
	case "queue-failed":
		return queueFailedScenario()
	case "horizon-processes":
		return horizonProcessesScenario()
	case "horizon-control":
		return horizonControlScenario()
	case "horizon-metrics":
		return horizonMetricsScenario()
	case "horizon-queue-lengths":
		return horizonQueueLengthsScenario()
	case "horizon-summaries":
		return horizonSummariesScenario()
	case "horizon-job-diagnostics":
		return horizonJobDiagnosticsScenario()
	case "horizon-observability":
		return horizonObservabilityScenario()
	case "horizon-orphans":
		return horizonOrphansScenario()
	default:
		return "", fmt.Errorf("unknown redis scenario %q", name)
	}
}

// defaultClientScenario resolves the Facade default client and round-trips a value.
func defaultClientScenario() (string, error) {
	client, cleanup, err := liveClient()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := redis.Connection()
	if err != nil {
		return "", fmt.Errorf("resolve default connection: %w", err)
	}
	key := fmt.Sprintf("prismgo_demo_redis_%d:client", time.Now().UnixNano())
	if err := client.Set(ctx, key, "hello", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set default client key: %w", err)
	}
	defer func() { _ = client.Del(ctx, key).Err() }()
	value, err := client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("get default client key: %w", err)
	}
	return fmt.Sprintf("name=%s get=%s", conn.Name(), value), nil
}

// namedClientScenario resolves the documented cache connection by name.
func namedClientScenario() (string, error) {
	client, cleanup, err := liveClient("cache")
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := redis.Connection("cache")
	if err != nil {
		return "", fmt.Errorf("resolve cache connection: %w", err)
	}
	key := fmt.Sprintf("prismgo_demo_redis_%d:cache", time.Now().UnixNano())
	if err := client.Set(ctx, key, "hello", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set named client key: %w", err)
	}
	defer func() { _ = client.Del(ctx, key).Err() }()
	value, err := client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("get named client key: %w", err)
	}
	return fmt.Sprintf("name=%s get=%s", conn.Name(), value), nil
}

// connectionScenario exposes the Connection metadata and listener registration.
func connectionScenario() (string, error) {
	_, cleanup, err := openLive()
	if err != nil {
		return "", err
	}
	defer cleanup()
	conn, err := redis.Connection()
	if err != nil {
		return "", fmt.Errorf("resolve default connection: %w", err)
	}
	conn.Listen(func(context.Context, redis.CommandExecuted) {})
	conn.ListenForFailures(func(context.Context, redis.CommandFailed) {})
	return fmt.Sprintf("name=%s client=%t listeners=true", conn.Name(), conn.Client() != nil), nil
}

// stringCommands exercises the documented string command surface.
func stringCommands(ctx context.Context, client goredis.UniversalClient, prefix string) (string, error) {
	key := prefix + "string"
	if err := client.Set(ctx, key, "hello", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set string: %w", err)
	}
	appended, err := client.Append(ctx, key, "world").Result()
	if err != nil {
		return "", fmt.Errorf("append string: %w", err)
	}
	length, err := client.StrLen(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("strlen: %w", err)
	}
	slice, err := client.GetRange(ctx, key, 0, 4).Result()
	if err != nil {
		return "", fmt.Errorf("getrange: %w", err)
	}
	first, second := prefix+"a", prefix+"b"
	if err := client.MSet(ctx, first, "1", second, "2").Err(); err != nil {
		return "", fmt.Errorf("mset: %w", err)
	}
	values, err := client.MGet(ctx, first, second).Result()
	if err != nil {
		return "", fmt.Errorf("mget: %w", err)
	}
	return fmt.Sprintf("get=hello append=%d strlen=%d range=%s mget=%v", appended, length, slice, values), nil
}

// hashCommands exercises the documented hash command surface.
func hashCommands(ctx context.Context, client goredis.UniversalClient, prefix string) (string, error) {
	key := prefix + "hash"
	fields, err := client.HSet(ctx, key, "name", "alice", "age", "28").Result()
	if err != nil {
		return "", fmt.Errorf("hset: %w", err)
	}
	name, err := client.HGet(ctx, key, "name").Result()
	if err != nil {
		return "", fmt.Errorf("hget: %w", err)
	}
	all, err := client.HGetAll(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("hgetall: %w", err)
	}
	names := make([]string, 0, len(all))
	for field := range all {
		names = append(names, field+"="+all[field])
	}
	sort.Strings(names)
	removed, err := client.HDel(ctx, key, "name").Result()
	if err != nil {
		return "", fmt.Errorf("hdel: %w", err)
	}
	return fmt.Sprintf("hset=%d hget=%s fields=%s hdel=%d", fields, name, strings.Join(names, ","), removed), nil
}

// listCommands exercises the documented list command surface.
func listCommands(ctx context.Context, client goredis.UniversalClient, prefix string) (string, error) {
	key := prefix + "list"
	pushed, err := client.RPush(ctx, key, "a", "b", "c").Result()
	if err != nil {
		return "", fmt.Errorf("rpush: %w", err)
	}
	items, err := client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return "", fmt.Errorf("lrange: %w", err)
	}
	popped, err := client.LPop(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("lpop: %w", err)
	}
	length, err := client.LLen(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("llen: %w", err)
	}
	return fmt.Sprintf("rpush=%d range=%v lpop=%s llen=%d", pushed, items, popped, length), nil
}

// setCommands exercises the documented set command surface.
func setCommands(ctx context.Context, client goredis.UniversalClient, prefix string) (string, error) {
	key := prefix + "set"
	added, err := client.SAdd(ctx, key, "a", "b", "c").Result()
	if err != nil {
		return "", fmt.Errorf("sadd: %w", err)
	}
	cardinality, err := client.SCard(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("scard: %w", err)
	}
	member, err := client.SIsMember(ctx, key, "a").Result()
	if err != nil {
		return "", fmt.Errorf("sismember: %w", err)
	}
	members, err := client.SMembers(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("smembers: %w", err)
	}
	sort.Strings(members)
	removed, err := client.SRem(ctx, key, "a").Result()
	if err != nil {
		return "", fmt.Errorf("srem: %w", err)
	}
	return fmt.Sprintf("sadd=%d card=%d member=%t members=%v srem=%d", added, cardinality, member, members, removed), nil
}

// sortedSetCommands exercises the documented sorted set command surface.
func sortedSetCommands(ctx context.Context, client goredis.UniversalClient, prefix string) (string, error) {
	key := prefix + "ranking"
	added, err := client.ZAdd(ctx, key,
		goredis.Z{Score: 100, Member: "alice"},
		goredis.Z{Score: 90, Member: "bob"},
	).Result()
	if err != nil {
		return "", fmt.Errorf("zadd: %w", err)
	}
	ranking, err := client.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return "", fmt.Errorf("zrange: %w", err)
	}
	score, err := client.ZScore(ctx, key, "alice").Result()
	if err != nil {
		return "", fmt.Errorf("zscore: %w", err)
	}
	rank, err := client.ZRevRank(ctx, key, "alice").Result()
	if err != nil {
		return "", fmt.Errorf("zrevrank: %w", err)
	}
	incremented, err := client.ZIncrBy(ctx, key, 5, "alice").Result()
	if err != nil {
		return "", fmt.Errorf("zincrby: %w", err)
	}
	return fmt.Sprintf("zadd=%d range=%v score=%v rank=%d incr=%v", added, ranking, score, rank, incremented), nil
}

// counterCommands exercises the documented counter command surface.
func counterCommands(ctx context.Context, client goredis.UniversalClient, prefix string) (string, error) {
	key := prefix + "counter"
	incr, err := client.Incr(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("incr: %w", err)
	}
	incrBy, err := client.IncrBy(ctx, key, 3).Result()
	if err != nil {
		return "", fmt.Errorf("incrby: %w", err)
	}
	decrBy, err := client.DecrBy(ctx, key, 1).Result()
	if err != nil {
		return "", fmt.Errorf("decrby: %w", err)
	}
	float, err := client.IncrByFloat(ctx, key, 0.5).Result()
	if err != nil {
		return "", fmt.Errorf("incrbyfloat: %w", err)
	}
	return fmt.Sprintf("incr=%d incrby=%d decrby=%d float=%v", incr, incrBy, decrBy, float), nil
}

// keyCommands exercises the documented existence, expiration, and deletion commands.
func keyCommands(ctx context.Context, client goredis.UniversalClient, prefix string) (string, error) {
	key := prefix + "key"
	if err := client.Set(ctx, key, "value", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set key: %w", err)
	}
	exists, err := client.Exists(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("exists: %w", err)
	}
	ttl, err := client.TTL(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("ttl: %w", err)
	}
	persisted, err := client.Persist(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("persist: %w", err)
	}
	kind, err := client.Type(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("type: %w", err)
	}
	removed, err := client.Del(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("del: %w", err)
	}
	gone, err := client.Exists(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("exists after delete: %w", err)
	}
	return fmt.Sprintf("exists=%t ttl-positive=%t persist=%t type=%s del=%d gone=%t",
		exists == 1, ttl > 0, persisted, kind, removed, gone == 0), nil
}

// transactionScenario shows MULTI/EXEC reads only resolve after Exec.
func transactionScenario(ctx context.Context, client goredis.UniversalClient, prefix string) (string, error) {
	key := prefix + "transaction"
	pipe := client.TxPipeline()
	incremented := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Minute)
	before := incremented.Val()
	if _, err := pipe.Exec(ctx); err != nil {
		return "", fmt.Errorf("exec transaction: %w", err)
	}
	ttl, err := client.TTL(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("transaction ttl: %w", err)
	}
	return fmt.Sprintf("before=%d after=%d ttl-positive=%t", before, incremented.Val(), ttl > 0), nil
}

// luaScenario runs an atomic conditional script through goredis.NewScript.
func luaScenario(ctx context.Context, client goredis.UniversalClient, prefix string) (string, error) {
	key := prefix + "lock"
	script := goredis.NewScript(`
		if redis.call("exists", KEYS[1]) == 0 then
			redis.call("set", KEYS[1], ARGV[1])
			return 1
		end
		return 0
	`)
	first, err := script.Run(ctx, client, []string{key}, "locked").Int()
	if err != nil {
		return "", fmt.Errorf("run script first: %w", err)
	}
	second, err := script.Run(ctx, client, []string{key}, "locked").Int()
	if err != nil {
		return "", fmt.Errorf("run script second: %w", err)
	}
	value, err := client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("read script value: %w", err)
	}
	return fmt.Sprintf("first=%d second=%d value=%s", first, second, value), nil
}

// pipelineScenario batches commands and reads every queued result.
func pipelineScenario(ctx context.Context, client goredis.UniversalClient, prefix string) (string, error) {
	first, second := prefix+"pipe-a", prefix+"pipe-b"
	pipe := client.Pipeline()
	pipe.Set(ctx, first, "1", 0)
	pipe.Set(ctx, second, "2", 0)
	get := pipe.Get(ctx, first)
	commands, err := pipe.Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("exec pipeline: %w", err)
	}
	values, err := client.MGet(ctx, first, second).Result()
	if err != nil {
		return "", fmt.Errorf("read pipeline values: %w", err)
	}
	return fmt.Sprintf("commands=%d get=%s mget=%v", len(commands), get.Val(), values), nil
}

// pubSubScenario exercises publish, channel subscribe, and pattern subscribe.
func pubSubScenario(name string) (string, error) {
	client, cleanup, err := liveClient()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	prefix := fmt.Sprintf("prismgo_demo_redis_pubsub_%d:", time.Now().UnixNano())
	switch name {
	case "publish":
		channel := prefix + "notifications"
		sub := client.Subscribe(ctx, channel)
		defer func() { _ = sub.Close() }()
		if _, err := sub.Receive(ctx); err != nil {
			return "", fmt.Errorf("confirm subscription: %w", err)
		}
		subscribers, err := client.Publish(ctx, channel, "order-received").Result()
		if err != nil {
			return "", fmt.Errorf("publish message: %w", err)
		}
		message, err := sub.ReceiveMessage(ctx)
		if err != nil {
			return "", fmt.Errorf("receive message: %w", err)
		}
		return fmt.Sprintf("subscribers=%d channel=%s payload=%s", subscribers, message.Channel, message.Payload), nil
	case "subscribe":
		channel := prefix + "notifications"
		sub := client.Subscribe(ctx, channel)
		defer func() { _ = sub.Close() }()
		if _, err := sub.Receive(ctx); err != nil {
			return "", fmt.Errorf("confirm subscription: %w", err)
		}
		if err := client.Publish(ctx, channel, "hello").Err(); err != nil {
			return "", fmt.Errorf("publish message: %w", err)
		}
		message, err := sub.ReceiveMessage(ctx)
		if err != nil {
			return "", fmt.Errorf("receive message: %w", err)
		}
		return fmt.Sprintf("subscribed=true channel=%s payload=%s", message.Channel, message.Payload), nil
	case "psubscribe":
		pattern := prefix + "orders:*"
		sub := client.PSubscribe(ctx, pattern)
		defer func() { _ = sub.Close() }()
		if _, err := sub.Receive(ctx); err != nil {
			return "", fmt.Errorf("confirm pattern subscription: %w", err)
		}
		if err := client.Publish(ctx, prefix+"orders:1", "created").Err(); err != nil {
			return "", fmt.Errorf("publish pattern message: %w", err)
		}
		message, err := sub.ReceiveMessage(ctx)
		if err != nil {
			return "", fmt.Errorf("receive pattern message: %w", err)
		}
		return fmt.Sprintf("pattern=%s channel=%s payload=%s", pattern, message.Channel, message.Payload), nil
	}
	return "", fmt.Errorf("unknown pubsub scenario %q", name)
}

// commandFn is one Redis command scenario bound to a live client and key prefix.
type commandFn func(context.Context, goredis.UniversalClient, string) (string, error)

// withLiveClient runs one command scenario and removes every key it created.
func withLiveClient(fn commandFn) (string, error) {
	client, cleanup, err := liveClient()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	prefix := fmt.Sprintf("prismgo_demo_redis_%d:", time.Now().UnixNano())
	defer func() {
		keys, keyErr := client.Keys(ctx, prefix+"*").Result()
		if keyErr == nil && len(keys) > 0 {
			_ = client.Del(ctx, keys...).Err()
		}
	}()
	return fn(ctx, client, prefix)
}

// liveClient boots an isolated application and resolves one Facade client.
func liveClient(name ...string) (goredis.UniversalClient, func(), error) {
	_, cleanup, err := openLive()
	if err != nil {
		return nil, nil, err
	}
	client, err := redis.Client(name...)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	return client, cleanup, nil
}

// redisConnectionClient resolves one Redis connection client from the current application.
func redisConnectionClient(name string) (goredis.UniversalClient, error) {
	return redis.Client(name)
}

// liveOptions customizes the isolated Application used by integration scenarios.
type liveOptions struct {
	// cacheDatabase overrides the logical database of the cache connection.
	cacheDatabase *int
	// configure allows a scenario to adjust the Application builder before boot.
	configure func(*foundation.Builder)
	// env adds scenario-specific environment overrides after the Redis connection variables.
	env map[string]string
}

// openLive boots an application whose Redis connections point at PRISMGO_REDIS_TEST_URL.
func openLive() (app *foundation.Application, cleanup func(), err error) {
	return openLiveWith(liveOptions{})
}

// openLiveWith boots an isolated application with optional Redis connection overrides.
func openLiveWith(opts liveOptions) (app *foundation.Application, cleanup func(), err error) {
	rawURL := strings.TrimSpace(os.Getenv("PRISMGO_REDIS_TEST_URL"))
	if rawURL == "" {
		return nil, nil, errors.New("PRISMGO_REDIS_TEST_URL is required for Redis integration scenarios")
	}
	redisEnv, err := redisConnectionEnv(rawURL, opts.cacheDatabase)
	if err != nil {
		return nil, nil, err
	}
	basePath, err := os.MkdirTemp("", "prismgo-redis-live-")
	if err != nil {
		return nil, nil, err
	}
	// 放置空 .env，让配置加载只使用注入的环境变量而不打印缺失警告。
	if err := os.WriteFile(filepath.Join(basePath, ".env"), nil, 0o600); err != nil {
		os.RemoveAll(basePath)
		return nil, nil, err
	}
	env := map[string]string{
		"APP_ENV":          "testing",
		"APP_KEY":          "base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=",
		"DB_CONNECTION":    "sqlite",
		"DB_DATABASE":      filepath.Join(basePath, "database.sqlite"),
		"CACHE_STORE":      "memory",
		"QUEUE_CONNECTION": "sync",
		"SESSION_DRIVER":   "file",
	}
	for key, value := range redisEnv {
		env[key] = value
	}
	for key, value := range opts.env {
		env[key] = value
	}
	restore := setScenarioEnv(env)
	builder := foundation.Configure(basePath)
	if opts.configure != nil {
		opts.configure(builder)
	}
	app = builder.Create()
	if bootErr := app.Boot(); bootErr != nil {
		restore()
		os.RemoveAll(basePath)
		return nil, nil, fmt.Errorf("boot redis integration application: %w", bootErr)
	}
	cleanup = func() {
		_ = app.CloseContext(context.Background())
		restore()
		os.RemoveAll(basePath)
	}
	return app, cleanup, nil
}

// redisConnectionEnv derives database.redis environment overrides from a Redis URL.
//
// 需求背景：demo 的 database.redis 为 host/port 提供默认值，而 framework 的地址优先级
// 是 addr > host/port > url；只设置 REDIS_URL 会被默认 host/port 覆盖。这里显式回填
// host/port/database，确保集成场景连接 PRISMGO_REDIS_TEST_URL 指向的真实服务。
func redisConnectionEnv(rawURL string, cacheDatabase *int) (map[string]string, error) {
	options, err := goredis.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url %q: %w", rawURL, err)
	}
	host, port, err := net.SplitHostPort(options.Addr)
	if err != nil {
		return nil, fmt.Errorf("split redis address %q: %w", options.Addr, err)
	}
	cacheDB := options.DB
	if cacheDatabase != nil {
		cacheDB = *cacheDatabase
	}
	return map[string]string{
		"REDIS_URL":       rawURL,
		"REDIS_HOST":      host,
		"REDIS_PORT":      port,
		"REDIS_MAIN_DB":   strconv.Itoa(options.DB),
		"REDIS_CACHE_URL": "",
		"REDIS_CACHE_DB":  strconv.Itoa(cacheDB),
		"REDIS_USERNAME":  options.Username,
		"REDIS_PASSWORD":  options.Password,
	}, nil
}

// liveManager boots an isolated application and resolves the Facade Manager.
func liveManager() (*redis.Manager, func(), error) {
	return liveManagerWith(liveOptions{})
}

// liveManagerWith boots an isolated application with options and resolves the Manager.
func liveManagerWith(opts liveOptions) (*redis.Manager, func(), error) {
	_, cleanup, err := openLiveWith(opts)
	if err != nil {
		return nil, nil, err
	}
	manager := redis.ManagerInstance()
	if manager == nil {
		cleanup()
		return nil, nil, errors.New("redis manager facade was not resolved")
	}
	return manager, cleanup, nil
}

// newRedisPrefix returns a unique key prefix for one integration scenario.
func newRedisPrefix() string {
	return fmt.Sprintf("prismgo_demo_redis_%d:", time.Now().UnixNano())
}

// deleteRedisKeys removes every key created under prefix.
func deleteRedisKeys(ctx context.Context, client goredis.UniversalClient, prefix string) {
	keys, err := client.Keys(ctx, prefix+"*").Result()
	if err != nil || len(keys) == 0 {
		return
	}
	_ = client.Del(ctx, keys...).Err()
}

// optionsDatabase reports the logical database of one managed connection.
func optionsDatabase(conn rediscontract.Connection) (int, error) {
	options, err := clientOptions(conn)
	if err != nil {
		return 0, err
	}
	return options.DB, nil
}

// defaultConnectionScenario resolves the default connection with and without a name.
func defaultConnectionScenario() (string, error) {
	manager, cleanup, err := liveManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := manager.Connection()
	if err != nil {
		return "", err
	}
	named, err := manager.Connection(redis.DefaultConnectionName)
	if err != nil {
		return "", err
	}
	if err := conn.Client().Ping(ctx).Err(); err != nil {
		return "", fmt.Errorf("ping default connection: %w", err)
	}
	return fmt.Sprintf("name=%s same-as-named=%t ping=pong", conn.Name(), conn == named), nil
}

// namedConnectionScenario resolves a non-default connection by name.
func namedConnectionScenario() (string, error) {
	manager, cleanup, err := liveManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	if err := conn.Client().Ping(ctx).Err(); err != nil {
		return "", fmt.Errorf("ping cache connection: %w", err)
	}
	return fmt.Sprintf("name=%s ping=pong", conn.Name()), nil
}

// defaultConnectionMethodScenario verifies DefaultConnection matches Connection().
func defaultConnectionMethodScenario() (string, error) {
	manager, cleanup, err := liveManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	explicit, err := manager.DefaultConnection()
	if err != nil {
		return "", err
	}
	implicit, err := manager.Connection()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("name=%s same=%t", explicit.Name(), explicit == implicit), nil
}

// multipleConnectionsScenario verifies named connections isolate their databases.
func multipleConnectionsScenario() (string, error) {
	cacheDB := 5
	manager, cleanup, err := liveManagerWith(liveOptions{cacheDatabase: &cacheDB})
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defaultConn, err := manager.Connection()
	if err != nil {
		return "", err
	}
	cacheConn, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	prefix := newRedisPrefix()
	defer deleteRedisKeys(ctx, defaultConn.Client(), prefix)
	defer deleteRedisKeys(ctx, cacheConn.Client(), prefix)
	defaultDB, err := optionsDatabase(defaultConn)
	if err != nil {
		return "", err
	}
	cacheDBValue, err := optionsDatabase(cacheConn)
	if err != nil {
		return "", err
	}
	key := prefix + "shared"
	if err := defaultConn.Client().Set(ctx, key, "default-value", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set default key: %w", err)
	}
	_, cacheErr := cacheConn.Client().Get(ctx, key).Result()
	return fmt.Sprintf("default-db=%d cache-db=%d isolated=%t", defaultDB, cacheDBValue, errors.Is(cacheErr, goredis.Nil)), nil
}

// purgeRebuildScenario verifies Purge closes and rebuilds the chosen connection.
func purgeRebuildScenario() (string, error) {
	manager, cleanup, err := liveManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	client := conn.Client()
	prefix := newRedisPrefix()
	key := prefix + "purge"
	if err := client.Set(ctx, key, "value", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set purge key: %w", err)
	}
	if err := manager.Purge("cache"); err != nil {
		return "", fmt.Errorf("purge cache connection: %w", err)
	}
	closed := errors.Is(client.Ping(ctx).Err(), goredis.ErrClosed)
	rebuilt, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	if rebuilt == conn {
		return "", fmt.Errorf("purged connection was reused, want a rebuilt connection")
	}
	defer deleteRedisKeys(ctx, rebuilt.Client(), prefix)
	if err := rebuilt.Client().Ping(ctx).Err(); err != nil {
		return "", fmt.Errorf("ping rebuilt cache connection: %w", err)
	}
	return fmt.Sprintf("closed=%t rebuilt=%t ping=pong", closed, rebuilt != conn), nil
}

// commandExecutedEventScenario observes the success event on the global bus.
func commandExecutedEventScenario() (string, error) {
	client, cleanup, err := liveClient()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var captured redis.CommandExecutedEvent
	received := 0
	event.Listen(redis.EventCommandExecuted, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if executed, ok := ev.(redis.CommandExecutedEvent); ok {
			captured = executed
			received++
		}
		return nil
	}))
	prefix := newRedisPrefix()
	defer deleteRedisKeys(ctx, client, prefix)
	if err := client.Set(ctx, prefix+"event", "value", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set event key: %w", err)
	}
	if received == 0 {
		return "", fmt.Errorf("success event received = 0, want at least 1")
	}
	return fmt.Sprintf("event=%s command=%s connection=%s", captured.Name(), captured.Command, captured.ConnectionName), nil
}

// commandFailedEventScenario observes the failure event on the global bus.
func commandFailedEventScenario() (string, error) {
	client, cleanup, err := liveClient()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var captured redis.CommandFailedEvent
	received := 0
	event.Listen(redis.EventCommandFailed, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if failed, ok := ev.(redis.CommandFailedEvent); ok {
			captured = failed
			received++
		}
		return nil
	}))
	prefix := newRedisPrefix()
	key := prefix + "failed"
	defer deleteRedisKeys(ctx, client, prefix)
	if err := client.Set(ctx, key, "value", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set failed key: %w", err)
	}
	if err := client.Incr(ctx, key).Err(); err == nil {
		return "", fmt.Errorf("incr on string key unexpectedly succeeded")
	}
	if received == 0 {
		return "", fmt.Errorf("failure event received = 0, want at least 1")
	}
	return fmt.Sprintf("event=%s command=%s error=true", captured.Name(), captured.Command), nil
}

// batchExecutedEventScenario observes the successful pipeline event.
func batchExecutedEventScenario() (string, error) {
	client, cleanup, err := liveClient()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var captured redis.CommandBatchExecutedEvent
	received := 0
	event.Listen(redis.EventCommandBatchExecuted, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if batch, ok := ev.(redis.CommandBatchExecutedEvent); ok {
			captured = batch
			received++
		}
		return nil
	}))
	prefix := newRedisPrefix()
	defer deleteRedisKeys(ctx, client, prefix)
	pipe := client.Pipeline()
	pipe.Set(ctx, prefix+"batch-a", "1", 0)
	pipe.Get(ctx, prefix+"batch-a")
	if _, err := pipe.Exec(ctx); err != nil {
		return "", fmt.Errorf("exec pipeline: %w", err)
	}
	if received == 0 {
		return "", fmt.Errorf("batch event received = 0, want at least 1")
	}
	return fmt.Sprintf("event=%s commands=%d", captured.Name(), len(captured.Commands)), nil
}

// batchFailedEventScenario observes the failed pipeline event.
func batchFailedEventScenario() (string, error) {
	client, cleanup, err := liveClient()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var captured redis.CommandBatchFailedEvent
	received := 0
	event.Listen(redis.EventCommandBatchFailed, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if batch, ok := ev.(redis.CommandBatchFailedEvent); ok {
			captured = batch
			received++
		}
		return nil
	}))
	prefix := newRedisPrefix()
	key := prefix + "batch-failed"
	defer deleteRedisKeys(ctx, client, prefix)
	if err := client.Set(ctx, key, "value", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set batch key: %w", err)
	}
	pipe := client.Pipeline()
	pipe.Set(ctx, prefix+"batch-ok", "1", 0)
	pipe.Incr(ctx, key)
	if _, err := pipe.Exec(ctx); err == nil {
		return "", fmt.Errorf("pipeline with failing command unexpectedly succeeded")
	}
	if received == 0 {
		return "", fmt.Errorf("batch failure event received = 0, want at least 1")
	}
	return fmt.Sprintf("event=%s commands=%d error=true", captured.Name(), len(captured.Commands)), nil
}

// eventPayloadsScenario snapshots the success and failure event payload fields.
func eventPayloadsScenario() (string, error) {
	client, cleanup, err := liveClient()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var executed redis.CommandExecutedEvent
	var failed redis.CommandFailedEvent
	executedCount, failedCount := 0, 0
	event.Listen(redis.EventCommandExecuted, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if value, ok := ev.(redis.CommandExecutedEvent); ok {
			executed = value
			executedCount++
		}
		return nil
	}))
	event.Listen(redis.EventCommandFailed, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if value, ok := ev.(redis.CommandFailedEvent); ok {
			failed = value
			failedCount++
		}
		return nil
	}))
	prefix := newRedisPrefix()
	key := prefix + "payload"
	defer deleteRedisKeys(ctx, client, prefix)
	if err := client.Set(ctx, key, "value", 0).Err(); err != nil {
		return "", fmt.Errorf("set payload key: %w", err)
	}
	if err := client.Incr(ctx, key).Err(); err == nil {
		return "", fmt.Errorf("incr on string key unexpectedly succeeded")
	}
	if executedCount == 0 || failedCount == 0 {
		return "", fmt.Errorf("event payloads executed=%d failed=%d, want both", executedCount, failedCount)
	}
	return fmt.Sprintf("executed=%s/%s/params:%d failed=%s/%s/params:%d/error:%t",
		executed.Command, executed.ConnectionName, len(executed.Parameters),
		failed.Command, failed.ConnectionName, len(failed.Parameters), failed.Error != nil), nil
}

// globalCommandListenerScenario registers a global success listener.
func globalCommandListenerScenario() (string, error) {
	client, cleanup, err := liveClient()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	received := 0
	event.Listen(redis.EventCommandExecuted, event.ListenerFunc(func(_ context.Context, _ event.Event) error {
		received++
		return nil
	}))
	registered := event.Has(redis.EventCommandExecuted)
	prefix := newRedisPrefix()
	defer deleteRedisKeys(ctx, client, prefix)
	if err := client.Set(ctx, prefix+"global", "value", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set global key: %w", err)
	}
	return fmt.Sprintf("registered=%t received=%d command=set", registered, received), nil
}

// globalFailureListenerScenario registers a global failure listener.
func globalFailureListenerScenario() (string, error) {
	client, cleanup, err := liveClient()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	received := 0
	event.Listen(redis.EventCommandFailed, event.ListenerFunc(func(_ context.Context, _ event.Event) error {
		received++
		return nil
	}))
	registered := event.Has(redis.EventCommandFailed)
	prefix := newRedisPrefix()
	key := prefix + "global-failed"
	defer deleteRedisKeys(ctx, client, prefix)
	if err := client.Set(ctx, key, "value", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set global key: %w", err)
	}
	if err := client.Incr(ctx, key).Err(); err == nil {
		return "", fmt.Errorf("incr on string key unexpectedly succeeded")
	}
	return fmt.Sprintf("registered=%t received=%d command=incr", registered, received), nil
}

// globalBatchListenerScenario registers a global batch listener.
func globalBatchListenerScenario() (string, error) {
	client, cleanup, err := liveClient()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	batches, commands := 0, 0
	event.Listen(redis.EventCommandBatchExecuted, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if batch, ok := ev.(redis.CommandBatchExecutedEvent); ok {
			batches++
			commands = len(batch.Commands)
		}
		return nil
	}))
	registered := event.Has(redis.EventCommandBatchExecuted)
	prefix := newRedisPrefix()
	defer deleteRedisKeys(ctx, client, prefix)
	pipe := client.Pipeline()
	pipe.Set(ctx, prefix+"global-batch-a", "1", 0)
	pipe.Get(ctx, prefix+"global-batch-a")
	if _, err := pipe.Exec(ctx); err != nil {
		return "", fmt.Errorf("exec pipeline: %w", err)
	}
	return fmt.Sprintf("registered=%t batches=%d commands=%d", registered, batches, commands), nil
}

// connectionListenerScenario registers success listeners on one connection.
func connectionListenerScenario() (string, error) {
	manager, cleanup, err := liveManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	var commands []string
	conn.Listen(func(_ context.Context, ev redis.CommandExecuted) {
		commands = append(commands, ev.Command)
	})
	client := conn.Client()
	prefix := newRedisPrefix()
	defer deleteRedisKeys(ctx, client, prefix)
	key := prefix + "listener"
	if err := client.Set(ctx, key, "value", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set listener key: %w", err)
	}
	if err := client.Get(ctx, key).Err(); err != nil {
		return "", fmt.Errorf("get listener key: %w", err)
	}
	if len(commands) == 0 {
		return "", fmt.Errorf("connection success events = 0, want at least 1")
	}
	return fmt.Sprintf("successes=%d sequence=%s", len(commands), strings.Join(commands, ",")), nil
}

// connectionFailureListenerScenario registers failure listeners on one connection.
func connectionFailureListenerScenario() (string, error) {
	manager, cleanup, err := liveManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	var commands []string
	conn.ListenForFailures(func(_ context.Context, ev redis.CommandFailed) {
		commands = append(commands, ev.Command)
	})
	client := conn.Client()
	prefix := newRedisPrefix()
	key := prefix + "listener-failed"
	defer deleteRedisKeys(ctx, client, prefix)
	if err := client.Set(ctx, key, "value", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set listener key: %w", err)
	}
	if err := client.Incr(ctx, key).Err(); err == nil {
		return "", fmt.Errorf("incr on string key unexpectedly succeeded")
	}
	if len(commands) == 0 {
		return "", fmt.Errorf("connection failure events = 0, want at least 1")
	}
	return fmt.Sprintf("failures=%d command=%s error=true", len(commands), strings.Join(commands, ",")), nil
}

// listenerPanicScenario verifies a panicking listener is isolated and reported.
func listenerPanicScenario() (string, error) {
	reports := make(chan map[string]any, 1)
	_, cleanup, err := openLiveWith(liveOptions{configure: func(builder *foundation.Builder) {
		builder.WithExceptions(func(exceptions *foundation.Exceptions) {
			exceptions.Report(func(_ any, _ error, fields map[string]any) {
				select {
				case reports <- fields:
				default:
				}
			})
		})
	}})
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	manager := redis.ManagerInstance()
	conn, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	surviving := 0
	conn.Listen(func(context.Context, redis.CommandExecuted) {
		panic("demo redis listener panic")
	})
	conn.Listen(func(context.Context, redis.CommandExecuted) {
		surviving++
	})
	prefix := newRedisPrefix()
	defer deleteRedisKeys(ctx, conn.Client(), prefix)
	if err := conn.Client().Set(ctx, prefix+"panic", "value", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("command result changed by listener panic: %w", err)
	}
	reported := false
	select {
	case <-reports:
		reported = true
	default:
	}
	return fmt.Sprintf("command=ok surviving=%d reported=%t", surviving, reported), nil
}

// disableEventsScenario verifies DisableEvents stops dispatching events.
func disableEventsScenario() (string, error) {
	manager, cleanup, err := liveManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	events := 0
	conn.Listen(func(context.Context, redis.CommandExecuted) {
		events++
	})
	client := conn.Client()
	prefix := newRedisPrefix()
	defer deleteRedisKeys(ctx, client, prefix)
	if err := client.Set(ctx, prefix+"disable-a", "1", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set before disable: %w", err)
	}
	before := events
	manager.DisableEvents()
	if err := client.Set(ctx, prefix+"disable-b", "2", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set after disable: %w", err)
	}
	if events != before {
		return "", fmt.Errorf("events after disable = %d, want %d", events, before)
	}
	return fmt.Sprintf("events-before=%d events-after=%d", before, events), nil
}

// enableEventsScenario verifies the toggle covers existing and future connections.
func enableEventsScenario() (string, error) {
	manager, cleanup, err := liveManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cacheConn, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	cacheEvents := 0
	cacheConn.Listen(func(context.Context, redis.CommandExecuted) {
		cacheEvents++
	})
	prefix := newRedisPrefix()
	defer deleteRedisKeys(ctx, cacheConn.Client(), prefix)
	if err := cacheConn.Client().Set(ctx, prefix+"enable-existing", "1", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set before disable: %w", err)
	}
	manager.DisableEvents()
	futureConn, err := manager.Connection()
	if err != nil {
		return "", err
	}
	defer deleteRedisKeys(ctx, futureConn.Client(), prefix)
	futureEvents := 0
	futureConn.Listen(func(context.Context, redis.CommandExecuted) {
		futureEvents++
	})
	if err := cacheConn.Client().Set(ctx, prefix+"enable-disabled", "2", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set while disabled: %w", err)
	}
	if err := futureConn.Client().Set(ctx, prefix+"enable-future-disabled", "3", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("future connection set while disabled: %w", err)
	}
	manager.EnableEvents()
	if err := cacheConn.Client().Set(ctx, prefix+"enable-reenabled", "4", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("set after re-enable: %w", err)
	}
	if err := futureConn.Client().Set(ctx, prefix+"enable-future-reenabled", "5", time.Minute).Err(); err != nil {
		return "", fmt.Errorf("future connection set after re-enable: %w", err)
	}
	return fmt.Sprintf("existing=%d future=%d", cacheEvents, futureEvents), nil
}
