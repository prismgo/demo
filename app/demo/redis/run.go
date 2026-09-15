// Package redisdemo contains executable examples of Redis configuration,
// connections, native commands, and administration flows.
package redisdemo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	configpkg "github.com/prismgo/framework/config"
	rediscontract "github.com/prismgo/framework/contracts/redis"
	"github.com/prismgo/framework/foundation"
	"github.com/prismgo/framework/redis"
	goredis "github.com/redis/go-redis/v9"

	// Load the demo's configuration defaults before reading database.redis.
	_ "prismgo-demo/config"
)

// Result records one observable Redis scenario.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

// Run executes a Redis catalog scenario that needs no external service.
func Run(name string) (Result, error) {
	value, err := run(name)
	if err != nil {
		return Result{}, fmt.Errorf("redis demo %s: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

func run(name string) (string, error) {
	switch name {
	case "architecture":
		return architectureScenario()
	case "config":
		return configScenario()
	case "client-identifiers":
		return clientIdentifiersScenario()
	case "named-config":
		return namedConfigScenario()
	case "address-precedence",
		"database-precedence",
		"authentication",
		"url",
		"tls",
		"client-name",
		"timeouts",
		"max-retries":
		return optionScenario(name)
	case "environment":
		return environmentScenario()
	case "native-client":
		return nativeClientScenario()
	case "manager":
		return managerConstructionScenario()
	case "manager-repository":
		return managerFromRepositoryScenario()
	case "manager-application":
		return managerFromApplicationScenario()
	case "lazy-connection":
		return lazyConnectionScenario()
	case "connection-reuse":
		return connectionReuseScenario()
	case "connections-snapshot":
		return connectionsSnapshotScenario()
	case "snapshot-lazy":
		return connectionSnapshotLazyScenario()
	case "purge":
		return purgeScenario()
	case "close":
		return closeScenario()
	case "close-errors":
		return closeErrorsScenario()
	case "close-cancellation":
		return closeCancellationScenario()
	case "event-sensitive-parameters":
		return eventSensitiveParametersScenario()
	default:
		return "", fmt.Errorf("unknown scenario %q", name)
	}
}

// architectureScenario exercises every compile-facing contract named in the guide.
func architectureScenario() (string, error) {
	var (
		_ rediscontract.Factory    = (*redis.Manager)(nil)
		_ rediscontract.Connection = (*redis.NamedConnection)(nil)
	)
	events := []string{
		redis.EventCommandExecuted,
		redis.EventCommandFailed,
		redis.EventCommandBatchExecuted,
		redis.EventCommandBatchFailed,
	}
	for _, name := range events {
		if strings.TrimSpace(name) == "" {
			return "", fmt.Errorf("redis event name is empty")
		}
	}
	_ = redis.Resolve
	_ = redis.ManagerInstance
	_ = redis.Connection
	_ = redis.Client
	_ = redis.ManagerCloseOption
	return "factory=Manager connection=NamedConnection events=4 facade=5", nil
}

// nativeClientScenario keeps the go-redis command, pipeline, and Pub/Sub
// surface referenced at compile time.
func nativeClientScenario() (string, error) {
	var (
		_ goredis.UniversalClient = (*goredis.Client)(nil)
		_ goredis.UniversalClient = (*goredis.ClusterClient)(nil)
		_                         = goredis.UniversalClient.Get
		_                         = goredis.UniversalClient.HSet
		_                         = goredis.UniversalClient.LPush
		_                         = goredis.UniversalClient.SAdd
		_                         = goredis.UniversalClient.ZAdd
		_                         = goredis.UniversalClient.Incr
		_                         = goredis.UniversalClient.Exists
		_                         = goredis.UniversalClient.TxPipeline
		_                         = goredis.UniversalClient.Pipeline
		_                         = goredis.UniversalClient.Subscribe
		_                         = goredis.UniversalClient.PSubscribe
		_                         = goredis.UniversalClient.Publish
	)
	return "client=UniversalClient commands=7 pipelines=2 pubsub=3", nil
}

// configScenario reports the default database.redis configuration.
func configScenario() (string, error) {
	cfg := redis.ConfigFromRepository(nil)
	defaultConn, ok := cfg.Connections[redis.DefaultConnectionName]
	if !ok {
		return "", fmt.Errorf("default connection %q is not configured", redis.DefaultConnectionName)
	}
	cacheConn, ok := cfg.Connections["cache"]
	if !ok {
		return "", fmt.Errorf("cache connection is not configured")
	}
	names := make([]string, 0, len(cfg.Connections))
	for name := range cfg.Connections {
		names = append(names, name)
	}
	sort.Strings(names)
	return fmt.Sprintf("client=%s default=%s connections=%s default-addr=%s cache-db=%d",
		cfg.Client, cfg.DefaultName, strings.Join(names, ","),
		defaultConn.Host+":"+defaultConn.Port, cacheConn.DB), nil
}

// clientIdentifiersScenario verifies the accepted Redis client identifiers.
func clientIdentifiersScenario() (string, error) {
	for _, client := range []string{"go", "go-redis", "redis", "GO"} {
		if _, err := redis.NewManager(redis.Config{Client: client}); err != nil {
			return "", fmt.Errorf("client identifier %q: %w", client, err)
		}
	}
	_, err := redis.NewManager(redis.Config{Client: "memcached"})
	return fmt.Sprintf("accepted=4 rejected=%t", err != nil), nil
}

// namedConfigScenario verifies a named connection reads its own environment keys.
func namedConfigScenario() (string, error) {
	envPath, cleanup, err := writeEnvFile([]string{
		"REDIS_CACHE_DB=7",
		"REDIS_CACHE_NAME=orders-client",
	})
	if err != nil {
		return "", err
	}
	defer cleanup()
	repo, err := configpkg.NewFromFile(envPath)
	if err != nil {
		return "", err
	}
	cfg := redis.ConfigFromRepository(repo)
	cacheConn, ok := cfg.Connections["cache"]
	if !ok {
		return "", fmt.Errorf("cache connection is not configured")
	}
	options, err := resolveOptions(cfg, "cache")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("cache-db=%d cache-name=%s", cacheConn.DB, options.ClientName), nil
}

// optionScenario verifies the documented connection parameter precedence and mapping.
func optionScenario(name string) (string, error) {
	switch name {
	case "address-precedence":
		addrWins, err := resolveConn(redis.ConnectionConfig{
			Addr: "10.0.0.9:7000", Host: "127.0.0.1", Port: "6379",
			Options: map[string]any{"url": "redis://url-host:6380/0"},
		})
		if err != nil {
			return "", err
		}
		hostWins, err := resolveConn(redis.ConnectionConfig{
			Host: "192.168.1.5", Port: "7001",
			Options: map[string]any{"url": "redis://url-host:6380/0"},
		})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("addr=%s host-override=%s", addrWins.Addr, hostWins.Addr), nil
	case "database-precedence":
		dbWins, err := resolveConn(redis.ConnectionConfig{
			Host: "127.0.0.1", Port: "6379", DB: 5,
			Options: map[string]any{"database": 2, "db": 5},
		})
		if err != nil {
			return "", err
		}
		urlKept, err := resolveConn(redis.ConnectionConfig{
			Options: map[string]any{"url": "redis://127.0.0.1:6379/1"},
		})
		if err != nil {
			return "", err
		}
		urlOverride, err := resolveConn(redis.ConnectionConfig{
			DB: 3, Options: map[string]any{"url": "redis://127.0.0.1:6379/1"},
		})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("db=%d url-kept=%d url-override=%d", dbWins.DB, urlKept.DB, urlOverride.DB), nil
	case "authentication":
		override, err := resolveConn(redis.ConnectionConfig{
			Username: "acl-user", Password: "acl-pass",
			Options: map[string]any{"url": "redis://url-user:url-pass@127.0.0.1:6379/0"},
		})
		if err != nil {
			return "", err
		}
		urlOnly, err := resolveConn(redis.ConnectionConfig{
			Options: map[string]any{"url": "redis://url-user:url-pass@127.0.0.1:6379/0"},
		})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("override=%s/%s url=%s/%s",
			override.Username, override.Password, urlOnly.Username, urlOnly.Password), nil
	case "url":
		options, err := resolveConn(redis.ConnectionConfig{
			Options: map[string]any{"url": "redis://user:secret@127.0.0.1:6390/4?dial_timeout=3s&read_timeout=2s&write_timeout=5s&max_retries=3"},
		})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("addr=%s db=%d user=%s dial=%s read=%s write=%s retries=%d",
			options.Addr, options.DB, options.Username, options.DialTimeout, options.ReadTimeout, options.WriteTimeout, options.MaxRetries), nil
	case "tls":
		secured, err := resolveConn(redis.ConnectionConfig{Options: map[string]any{"scheme": "tls"}})
		if err != nil {
			return "", err
		}
		plain, err := resolveConn(redis.ConnectionConfig{})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("tls=%t plain=%t", secured.TLSConfig != nil, plain.TLSConfig != nil), nil
	case "client-name":
		options, err := resolveConn(redis.ConnectionConfig{Options: map[string]any{"name": "orders-client"}})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("name=%s", options.ClientName), nil
	case "timeouts":
		durations, err := resolveConn(redis.ConnectionConfig{
			Options: map[string]any{"timeout": "3s", "read_timeout": "500ms", "write_timeout": "2s"},
		})
		if err != nil {
			return "", err
		}
		seconds, err := resolveConn(redis.ConnectionConfig{
			Options: map[string]any{"timeout": "4", "read_timeout": "1", "write_timeout": "2"},
		})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("duration=%s/%s/%s seconds=%s/%s/%s",
			durations.DialTimeout, durations.ReadTimeout, durations.WriteTimeout,
			seconds.DialTimeout, seconds.ReadTimeout, seconds.WriteTimeout), nil
	case "max-retries":
		number, err := resolveConn(redis.ConnectionConfig{Options: map[string]any{"max_retries": 7}})
		if err != nil {
			return "", err
		}
		text, err := resolveConn(redis.ConnectionConfig{Options: map[string]any{"max_retries": "9"}})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("int=%d string=%d", number.MaxRetries, text.MaxRetries), nil
	}
	return "", fmt.Errorf("unknown option scenario %q", name)
}

// environmentScenario verifies the documented REDIS_* environment overrides.
func environmentScenario() (string, error) {
	restore := setScenarioEnv(map[string]string{
		"REDIS_URL": "", "REDIS_HOST": "10.2.0.1", "REDIS_PORT": "6389", "REDIS_MAIN_DB": "6",
		"REDIS_NAME": "env-client", "REDIS_TIMEOUT": "4s", "REDIS_READ_TIMEOUT": "1s",
		"REDIS_WRITE_TIMEOUT": "2s", "REDIS_MAX_RETRIES": "8", "REDIS_CLIENT": "go-redis",
	})
	defer restore()
	envPath, cleanup, err := writeEnvFile(nil)
	if err != nil {
		return "", err
	}
	defer cleanup()
	repo, err := configpkg.NewFromFile(envPath)
	if err != nil {
		return "", err
	}
	cfg := redis.ConfigFromRepository(repo)
	options, err := resolveOptions(cfg, redis.DefaultConnectionName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("client=%s addr=%s db=%d name=%s dial=%s read=%s write=%s retries=%d",
		cfg.Client, options.Addr, options.DB, options.ClientName,
		options.DialTimeout, options.ReadTimeout, options.WriteTimeout, options.MaxRetries), nil
}

// singleConnConfig wraps one connection in a manager-ready configuration.
func singleConnConfig(conn redis.ConnectionConfig) redis.Config {
	conn.Name = redis.DefaultConnectionName
	return redis.Config{
		DefaultName: redis.DefaultConnectionName,
		Connections: map[string]redis.ConnectionConfig{redis.DefaultConnectionName: conn},
	}
}

// resolveConn resolves one literal connection and returns its go-redis options.
func resolveConn(conn redis.ConnectionConfig) (*goredis.Options, error) {
	return resolveOptions(singleConnConfig(conn), redis.DefaultConnectionName)
}

// resolveOptions creates a manager, resolves one connection lazily, and exposes
// the go-redis options the framework derived from configuration.
func resolveOptions(cfg redis.Config, name string) (*goredis.Options, error) {
	manager, err := redis.NewManager(cfg)
	if err != nil {
		return nil, err
	}
	defer func() { _ = manager.Close(context.Background()) }()
	conn, err := manager.Connection(name)
	if err != nil {
		return nil, err
	}
	return clientOptions(conn)
}

// clientOptions exposes the go-redis options a managed connection was built from.
func clientOptions(conn rediscontract.Connection) (*goredis.Options, error) {
	client, ok := conn.Client().(*goredis.Client)
	if !ok {
		return nil, fmt.Errorf("client type = %T, want *goredis.Client", conn.Client())
	}
	return client.Options(), nil
}

// writeEnvFile materializes a temporary .env file for config loading.
func writeEnvFile(lines []string) (string, func(), error) {
	dir, err := os.MkdirTemp("", "prismgo-redis-env-")
	if err != nil {
		return "", nil, err
	}
	path := filepath.Join(dir, ".env")
	content := ""
	if len(lines) > 0 {
		content = strings.Join(lines, "\n") + "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		os.RemoveAll(dir)
		return "", nil, err
	}
	return path, func() { os.RemoveAll(dir) }, nil
}

// setScenarioEnv applies environment variables for one scenario and returns a restore closure.
func setScenarioEnv(values map[string]string) func() {
	previous := make(map[string]*string, len(values))
	for key, value := range values {
		if old, ok := os.LookupEnv(key); ok {
			copied := old
			previous[key] = &copied
		} else {
			previous[key] = nil
		}
		_ = os.Setenv(key, value)
	}
	return func() {
		for key, old := range previous {
			if old == nil {
				_ = os.Unsetenv(key)
				continue
			}
			_ = os.Setenv(key, *old)
		}
	}
}

// newHermeticManager builds a two-connection Manager that is never dialed.
//
// 设计思路：NewManager 只校验并保存配置，Connection 首次解析时才构造 go-redis
// client；因此默认地址无需真实 Redis 即可覆盖懒加载、缓存、快照、重建和关闭语义。
func newHermeticManager() (*redis.Manager, error) {
	return redis.NewManager(redis.Config{
		DefaultName: redis.DefaultConnectionName,
		Connections: map[string]redis.ConnectionConfig{
			redis.DefaultConnectionName: {Name: redis.DefaultConnectionName, Host: "127.0.0.1", Port: "6379", DB: 0},
			"cache":                     {Name: "cache", Host: "127.0.0.1", Port: "6379", DB: 1},
		},
	})
}

// closeManager releases a hermetic manager whose clients were never connected.
func closeManager(manager *redis.Manager) {
	if manager == nil {
		return
	}
	_ = manager.Close(context.Background())
}

// managerConstructionScenario constructs a Manager from explicit configuration.
func managerConstructionScenario() (string, error) {
	if _, err := redis.NewManager(redis.Config{Client: "memcached"}); err == nil {
		return "", fmt.Errorf("unsupported client identifier should be rejected")
	}
	manager, err := newHermeticManager()
	if err != nil {
		return "", err
	}
	defer closeManager(manager)
	if _, err := manager.Connection(); err != nil {
		return "", err
	}
	cache, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("default=%s cache=%s resolved=%d invalid-client=rejected",
		redis.DefaultConnectionName, cache.Name(), len(manager.Connections())), nil
}

// managerFromRepositoryScenario builds a Manager from a config repository.
func managerFromRepositoryScenario() (string, error) {
	envPath, cleanup, err := writeEnvFile([]string{"REDIS_HOST=10.9.0.1", "REDIS_PORT=6390", "REDIS_MAIN_DB=3"})
	if err != nil {
		return "", err
	}
	defer cleanup()
	repo, err := configpkg.NewFromFile(envPath)
	if err != nil {
		return "", err
	}
	cfg := redis.ConfigFromRepository(repo)
	manager, err := redis.NewManager(cfg)
	if err != nil {
		return "", err
	}
	defer closeManager(manager)
	conn, err := manager.DefaultConnection()
	if err != nil {
		return "", err
	}
	options, err := clientOptions(conn)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("source=repository connection=%s addr=%s db=%d", conn.Name(), options.Addr, options.DB), nil
}

// managerFromApplicationScenario resolves the Manager through the container Facade.
func managerFromApplicationScenario() (string, error) {
	_, cleanup, err := openHermeticApplication()
	if err != nil {
		return "", err
	}
	defer cleanup()
	factory := redis.Resolve()
	manager := redis.ManagerInstance()
	if factory == nil || manager == nil {
		return "", fmt.Errorf("redis facade resolved factory=%T manager=%T, want *redis.Manager", factory, manager)
	}
	conn, err := manager.Connection()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("factory=%T connection=%s", factory, conn.Name()), nil
}

// lazyConnectionScenario verifies clients are created only when resolved.
func lazyConnectionScenario() (string, error) {
	manager, err := newHermeticManager()
	if err != nil {
		return "", err
	}
	defer closeManager(manager)
	before := len(manager.Connections())
	conn, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	after := len(manager.Connections())
	if _, err := manager.Connection("missing"); err == nil {
		return "", fmt.Errorf("missing connection should be rejected")
	}
	return fmt.Sprintf("before=%d after=%d missing=error client=%t", before, after, conn.Client() != nil), nil
}

// connectionReuseScenario verifies same-name connections are cached and reused.
func connectionReuseScenario() (string, error) {
	manager, err := newHermeticManager()
	if err != nil {
		return "", err
	}
	defer closeManager(manager)
	first, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	second, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	other, err := manager.Connection()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("same=%t distinct=%t", first == second, first != other), nil
}

// connectionsSnapshotScenario verifies the snapshot is a detached copy.
func connectionsSnapshotScenario() (string, error) {
	manager, err := newHermeticManager()
	if err != nil {
		return "", err
	}
	defer closeManager(manager)
	if _, err := manager.Connection(); err != nil {
		return "", err
	}
	if _, err := manager.Connection("cache"); err != nil {
		return "", err
	}
	snapshot := manager.Connections()
	copied := manager.Connections()
	delete(snapshot, "cache")
	isolated := len(manager.Connections()) == 2 && len(copied) == 2 && len(snapshot) == 1
	return fmt.Sprintf("count=%d isolated=%t", len(copied), isolated), nil
}

// connectionSnapshotLazyScenario verifies snapshots never resolve new clients.
func connectionSnapshotLazyScenario() (string, error) {
	manager, err := newHermeticManager()
	if err != nil {
		return "", err
	}
	defer closeManager(manager)
	initial := len(manager.Connections())
	_ = manager.Connections()
	still := len(manager.Connections())
	if _, err := manager.DefaultConnection(); err != nil {
		return "", err
	}
	after := len(manager.Connections())
	return fmt.Sprintf("initial=%d snapshot=%d after-resolve=%d", initial, still, after), nil
}

// purgeScenario verifies Purge closes and allows rebuilding a connection.
func purgeScenario() (string, error) {
	manager, err := newHermeticManager()
	if err != nil {
		return "", err
	}
	defer closeManager(manager)
	conn, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	client := conn.Client()
	if err := manager.Purge("cache"); err != nil {
		return "", err
	}
	_, present := manager.Connections()["cache"]
	closed := errors.Is(client.Ping(context.Background()).Err(), goredis.ErrClosed)
	rebuilt, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("purged=%t closed=%t rebuilt=%t", !present, closed, rebuilt != conn), nil
}

// closeScenario verifies Close releases resolved connections only.
func closeScenario() (string, error) {
	manager, err := newHermeticManager()
	if err != nil {
		return "", err
	}
	if _, err := manager.Connection(); err != nil {
		return "", err
	}
	if _, err := manager.Connection("cache"); err != nil {
		return "", err
	}
	if err := manager.Close(context.Background()); err != nil {
		return "", err
	}
	empty, err := newHermeticManager()
	if err != nil {
		return "", err
	}
	emptyErr := empty.Close(context.Background())
	return fmt.Sprintf("resolved=2 remaining=%d empty-close=%t", len(manager.Connections()), emptyErr == nil && len(empty.Connections()) == 0), nil
}

// closeErrorsScenario verifies Close continues past a failing connection.
func closeErrorsScenario() (string, error) {
	manager, err := newHermeticManager()
	if err != nil {
		return "", err
	}
	cache, err := manager.Connection("cache")
	if err != nil {
		return "", err
	}
	if _, err := manager.Connection(); err != nil {
		return "", err
	}
	// 预先关闭 cache 的底层连接池，使 Manager.Close 再次关闭时返回首个错误。
	_ = cache.Client().Close()
	err = manager.Close(context.Background())
	remaining := len(manager.Connections())
	return fmt.Sprintf("first-error=%t remaining=%d", errors.Is(err, goredis.ErrClosed), remaining), nil
}

// closeCancellationScenario verifies a cancelled Close leaves state for retry.
func closeCancellationScenario() (string, error) {
	manager, err := newHermeticManager()
	if err != nil {
		return "", err
	}
	if _, err := manager.Connection("cache"); err != nil {
		return "", err
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	err = manager.Close(cancelled)
	retained := len(manager.Connections())
	if retryErr := manager.Close(context.Background()); retryErr != nil {
		return "", retryErr
	}
	return fmt.Sprintf("cancelled=%t retained=%d retried=%d",
		errors.Is(err, context.Canceled), retained, len(manager.Connections())), nil
}

// eventSensitiveParametersScenario shows events keep command parameters verbatim.
func eventSensitiveParametersScenario() (string, error) {
	manager, err := redis.NewManager(redis.Config{
		DefaultName: redis.DefaultConnectionName,
		Connections: map[string]redis.ConnectionConfig{
			redis.DefaultConnectionName: {
				Name:    redis.DefaultConnectionName,
				Addr:    "127.0.0.1:1",
				Options: map[string]any{"timeout": "100ms", "max_retries": -1},
			},
		},
	})
	if err != nil {
		return "", err
	}
	defer closeManager(manager)
	conn, err := manager.DefaultConnection()
	if err != nil {
		return "", err
	}
	var observed redis.CommandFailed
	seen := false
	conn.ListenForFailures(func(_ context.Context, ev redis.CommandFailed) {
		observed = ev
		seen = true
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 连接不可达，但 hook 仍会在失败事件中保留原始参数，验证事件默认不脱敏。
	_ = conn.Client().Set(ctx, "demo:token:key", "demo-token-value", 0).Err()
	if !seen {
		return "", fmt.Errorf("expected a failed command event for the unreachable connection")
	}
	if len(observed.Parameters) != 2 || observed.Parameters[1] != "demo-token-value" {
		return "", fmt.Errorf("event parameters = %#v, want raw values without masking", observed.Parameters)
	}
	return fmt.Sprintf("command=%s parameters=%d sensitive-verbatim=true", observed.Command, len(observed.Parameters)), nil
}

// openHermeticApplication boots an isolated Application without external services.
func openHermeticApplication() (*foundation.Application, func(), error) {
	basePath, err := os.MkdirTemp("", "prismgo-redis-hermetic-")
	if err != nil {
		return nil, nil, err
	}
	if err := os.WriteFile(filepath.Join(basePath, ".env"), nil, 0o600); err != nil {
		os.RemoveAll(basePath)
		return nil, nil, err
	}
	restore := setScenarioEnv(map[string]string{
		"APP_ENV":          "testing",
		"APP_KEY":          "base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=",
		"DB_CONNECTION":    "sqlite",
		"DB_DATABASE":      filepath.Join(basePath, "database.sqlite"),
		"CACHE_STORE":      "memory",
		"QUEUE_CONNECTION": "sync",
		"SESSION_DRIVER":   "file",
		"REDIS_URL":        "",
	})
	app := foundation.Configure(basePath).Create()
	if bootErr := app.Boot(); bootErr != nil {
		restore()
		os.RemoveAll(basePath)
		return nil, nil, fmt.Errorf("boot hermetic application: %w", bootErr)
	}
	cleanup := func() {
		_ = app.CloseContext(context.Background())
		restore()
		os.RemoveAll(basePath)
	}
	return app, cleanup, nil
}
