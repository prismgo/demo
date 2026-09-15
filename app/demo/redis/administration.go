package redisdemo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prismgo/framework/container"
	containercontract "github.com/prismgo/framework/contracts/container"
	eventcontract "github.com/prismgo/framework/contracts/event"
	rediscontract "github.com/prismgo/framework/contracts/redis"
	"github.com/prismgo/framework/event"
	"github.com/prismgo/framework/redis"
	"github.com/prismgo/horizon"
	goredis "github.com/redis/go-redis/v9"
)

// providerApp is the minimal providercontract.Application view needed by provider scenarios.
type providerApp struct {
	registry containercontract.Container
}

// Container returns the container the provider should bind into.
func (a providerApp) Container() containercontract.Container { return a.registry }

// providerRegistrationScenario verifies Redis ServiceProvider container bindings.
func providerRegistrationScenario() (string, error) {
	registry := container.NewContainer()
	if err := (redis.ServiceProvider{}).Register(providerApp{registry: registry}); err != nil {
		return "", fmt.Errorf("register redis provider: %w", err)
	}
	if !registry.Bound("redis") || !registry.Bound("redis.connection") {
		return "", fmt.Errorf("redis bindings = %t/%t, want both", registry.Bound("redis"), registry.Bound("redis.connection"))
	}
	if registry.Resolved("redis") {
		return "", fmt.Errorf("redis manager resolved during Register, want lazy binding")
	}
	return fmt.Sprintf("name=%s bound=true lazy=true", (redis.ServiceProvider{}).Name()), nil
}

// providerEventBridgeScenario verifies Boot bridges Redis events into the event dispatcher.
func providerEventBridgeScenario() (string, error) {
	registry := container.NewContainer()
	container.SetProvider(func() *container.Container { return registry })
	defer container.SetProvider(nil)

	bus := event.New()
	if err := registry.Instance("event.dispatcher", eventcontract.Dispatcher(bus)); err != nil {
		return "", fmt.Errorf("bind event dispatcher: %w", err)
	}
	if err := (redis.ServiceProvider{}).Boot(providerApp{registry: registry}); err != nil {
		return "", fmt.Errorf("boot redis provider: %w", err)
	}
	defer redis.UseEventSink(nil)

	var captured redis.CommandFailed
	received := 0
	bus.Listen(redis.EventCommandFailed, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if failed, ok := ev.(redis.CommandFailedEvent); ok {
			captured = failed.CommandFailed
			received++
		}
		return nil
	}))
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 连接不可达，但 Boot 安装的 sink 仍会把失败事件桥接到当前 event.dispatcher。
	_ = conn.Client().Set(ctx, "prismgo_demo_redis_bridge", "value", 0).Err()
	if received == 0 {
		return "", fmt.Errorf("bridged failure events = 0, want at least 1")
	}
	return fmt.Sprintf("bridged=true event=%s command=%s", redis.EventCommandFailed, captured.Command), nil
}

// containerFactoryScenario resolves the shared Redis Manager through the container.
func containerFactoryScenario() (string, error) {
	_, cleanup, err := openHermeticApplication()
	if err != nil {
		return "", err
	}
	defer cleanup()
	factory, err := container.Make[rediscontract.Factory]("redis")
	if err != nil {
		return "", fmt.Errorf("resolve redis factory: %w", err)
	}
	manager, ok := factory.(*redis.Manager)
	if !ok {
		return "", fmt.Errorf("container redis = %T, want *redis.Manager", factory)
	}
	conn, err := manager.Connection()
	if err != nil {
		return "", fmt.Errorf("resolve default connection: %w", err)
	}
	if conn.Name() != redis.DefaultConnectionName {
		return "", fmt.Errorf("manager default connection = %q, want %q", conn.Name(), redis.DefaultConnectionName)
	}
	return "key=redis type=*redis.Manager default=default", nil
}

// containerConnectionScenario resolves the default Connection through the container.
func containerConnectionScenario() (string, error) {
	_, cleanup, err := openHermeticApplication()
	if err != nil {
		return "", err
	}
	defer cleanup()
	conn, err := container.Make[rediscontract.Connection]("redis.connection")
	if err != nil {
		return "", fmt.Errorf("resolve redis connection: %w", err)
	}
	if conn.Name() != redis.DefaultConnectionName {
		return "", fmt.Errorf("container connection name = %q, want %q", conn.Name(), redis.DefaultConnectionName)
	}
	return fmt.Sprintf("key=redis.connection name=%s client=%t", conn.Name(), conn.Client() != nil), nil
}

// containerNamedConnectionScenario resolves a named Connection through the container factory.
func containerNamedConnectionScenario() (string, error) {
	_, cleanup, err := openHermeticApplication()
	if err != nil {
		return "", err
	}
	defer cleanup()
	factory, err := container.Make[rediscontract.Factory]("redis")
	if err != nil {
		return "", fmt.Errorf("resolve redis factory: %w", err)
	}
	def, err := factory.DefaultConnection()
	if err != nil {
		return "", err
	}
	cache, err := factory.Connection("cache")
	if err != nil {
		return "", fmt.Errorf("resolve cache connection: %w", err)
	}
	if cache.Name() != "cache" || cache == def {
		return "", fmt.Errorf("cache connection = %#v, want distinct cache connection", cache.Name())
	}
	return "factory=redis connection=cache distinct=true", nil
}

// lifecycleCloseScenario verifies the container closes resolved Redis connections.
func lifecycleCloseScenario() (string, error) {
	app, cleanup, err := openHermeticApplication()
	if err != nil {
		return "", err
	}
	defer cleanup()
	manager := redis.ManagerInstance()
	if manager == nil {
		return "", fmt.Errorf("redis manager was not resolved")
	}
	def, err := manager.Connection()
	if err != nil {
		return "", err
	}
	if _, err := manager.Connection("cache"); err != nil {
		return "", err
	}
	if err := app.CloseContext(context.Background()); err != nil {
		return "", fmt.Errorf("close application: %w", err)
	}
	closed := errors.Is(def.Client().Ping(context.Background()).Err(), goredis.ErrClosed)
	return fmt.Sprintf("resolved=2 remaining=%d client-closed=%t", len(manager.Connections()), closed), nil
}

// horizonConfigReader is a minimal horizon.ConfigReader backed by explicit maps.
type horizonConfigReader struct {
	values map[string]string
	maps   map[string]map[string]any
}

// GetString returns one scalar configuration value.
func (r horizonConfigReader) GetString(path string, defaultValue ...any) string {
	if value, ok := r.values[path]; ok {
		return value
	}
	if len(defaultValue) > 0 {
		return fmt.Sprint(defaultValue[0])
	}
	return ""
}

// GetStringMap returns one nested configuration map.
func (r horizonConfigReader) GetStringMap(path string) map[string]any {
	return r.maps[path]
}

// horizonConfigScenario parses an explicit Horizon Redis store configuration.
func horizonConfigScenario() (string, error) {
	reader := horizonConfigReader{
		values: map[string]string{"app.env": "local"},
		maps: map[string]map[string]any{
			"horizon": {
				"use":                   "redis",
				"connection":            "cache",
				"prefix":                "demo_horizon",
				"path":                  "horizon",
				"encoding":              "json",
				"heartbeat_ttl_seconds": 30,
			},
		},
	}
	cfg, err := horizon.LoadConfigFrom(reader)
	if err != nil {
		return "", fmt.Errorf("load horizon config: %w", err)
	}
	if cfg.Store != "redis" || cfg.Connection != "cache" || cfg.Prefix != "demo_horizon" {
		return "", fmt.Errorf("horizon store/connection/prefix = %s/%s/%s, want redis/cache/demo_horizon", cfg.Store, cfg.Connection, cfg.Prefix)
	}
	if cfg.Encoding != "json" || cfg.HeartbeatTTL != 30*time.Second || cfg.Environment != "local" {
		return "", fmt.Errorf("horizon encoding/ttl/environment = %s/%s/%s, want json/30s/local", cfg.Encoding, cfg.HeartbeatTTL, cfg.Environment)
	}
	return fmt.Sprintf("store=%s connection=%s prefix=%s encoding=%s ttl=%s env=%s",
		cfg.Store, cfg.Connection, cfg.Prefix, cfg.Encoding, cfg.HeartbeatTTL, cfg.Environment), nil
}
