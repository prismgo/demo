// Package facadedemo contains runnable examples of framework facade access.
package facadedemo

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prismgo/framework/cache"
	"github.com/prismgo/framework/config"
	"github.com/prismgo/framework/container"
	"github.com/prismgo/framework/database"
	"github.com/prismgo/framework/facade"
	"github.com/prismgo/framework/filesystem"
	"github.com/prismgo/framework/logger"
	"github.com/prismgo/framework/route"
)

// Result records an observable facade demo outcome.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

// Run executes one documented facade scenario in the current application.
func Run(name string) (Result, error) {
	value, err := run(name)
	if err != nil {
		return Result{}, fmt.Errorf("facade demo %s: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

func run(name string) (string, error) {
	switch name {
	case "introduction":
		return introduction()
	case "how-facades-work":
		return howFacadesWork()
	case "core-facade-resolve":
		return coreFacadeResolve()
	case "list":
		return list()
	case "cache-facade":
		return cacheFacade()
	case "config-facade":
		return configFacade()
	case "route-facade":
		return routeFacade()
	case "logger-facade":
		return loggerFacade()
	case "database-facade":
		return databaseFacade()
	case "filesystem-facade":
		return filesystemFacade()
	case "class-reference":
		return classReference()
	case "laravel-mapping":
		return laravelMapping()
	default:
		return "", fmt.Errorf("unknown scenario %q", name)
	}
}

func introduction() (string, error) {
	cfgName := config.GetString("app.name")
	cacheName := cache.DefaultName()
	loggerName := logger.DefaultName()
	fsName := filesystem.DefaultName()
	return fmt.Sprintf("config=%s cache=%s logger=%s filesystem=%s", cfgName, cacheName, loggerName, fsName), nil
}

func howFacadesWork() (string, error) {
	r1 := config.Resolve()
	r2 := facade.Resolve[*config.Config]("config.default")
	sameInstance := r1 == r2
	r3, err := container.Make[*config.Config]("config.default")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("same-instance=%t container-make=%t", sameInstance, r1 == r3), nil
}

func coreFacadeResolve() (string, error) {
	cfg := facade.Resolve[*config.Config]("config.default")
	appName := cfg.GetString("app.name")
	debug := cfg.GetBool("app.debug")
	return fmt.Sprintf("type=%T name=%s debug=%t", cfg, appName, debug), nil
}

func list() (string, error) {
	knownKeys := []string{
		"cache.manager", "config.default", "database.default",
		"event.dispatcher", "filesystem.manager", "logger.manager",
		"queue.manager", "route.router", "session.manager",
	}
	entries := container.List()
	registered := make(map[string]bool, len(entries))
	for _, e := range entries {
		registered[e.Key] = true
	}
	resolved := make([]string, 0, len(knownKeys))
	for _, key := range knownKeys {
		if registered[key] {
			resolved = append(resolved, key)
		}
	}
	sort.Strings(resolved)
	return fmt.Sprintf("total=%d resolved=%s", len(resolved), strings.Join(resolved, ",")), nil
}

func cacheFacade() (string, error) {
	ctx := context.Background()
	key := "demo:facade:cache"
	defaultName := cache.DefaultName()
	if err := cache.Put(ctx, key, "facade-cache-value", time.Minute); err != nil {
		return "", err
	}
	value, err := cache.Get[string](ctx, key)
	if err != nil {
		return "", err
	}
	has, err := cache.Has(ctx, key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("default=%s value=%s has=%t", defaultName, value, has), nil
}

func configFacade() (string, error) {
	appName := config.GetString("app.name")
	debug := config.GetBool("app.debug")
	port := config.GetInt("app.server.port")
	return fmt.Sprintf("name=%s debug=%t port=%d", appName, debug, port), nil
}

func routeFacade() (string, error) {
	route.Get("/facade-demo", func(*gin.Context) {})
	routes := route.List()
	count := len(routes)
	return fmt.Sprintf("registered=%d", count), nil
}

func loggerFacade() (string, error) {
	defaultName := logger.DefaultName()
	hasDefault := defaultName != ""
	return fmt.Sprintf("default=%s has-default=%t", defaultName, hasDefault), nil
}

func databaseFacade() (string, error) {
	db := database.Resolve()
	dbType := fmt.Sprintf("%T", db)
	name := db.Name()
	return fmt.Sprintf("type=%s name=%s", dbType, name), nil
}

func filesystemFacade() (string, error) {
	ctx := context.Background()
	defaultName := filesystem.DefaultName()
	key := "facade-demo.txt"
	if err := filesystem.Put(ctx, key, "facade-fs-value"); err != nil {
		return "", err
	}
	exists, err := filesystem.Exists(ctx, key)
	if err != nil {
		return "", err
	}
	content, err := filesystem.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("default=%s exists=%t content=%s", defaultName, exists, string(content)), nil
}

func classReference() (string, error) {
	var cfg any = config.Resolve()
	var mgr any = cache.Resolve()
	var log any = logger.Resolve()
	return fmt.Sprintf("config=%T cache=%T logger=%T", cfg, mgr, log), nil
}

func laravelMapping() (string, error) {
	pairs := []string{
		"Cache=cache", "Config=config", "DB=database",
		"Event=event", "File=filesystem", "Log=logger",
		"Queue=queue", "Route=route", "Session=session",
		"Translator=translation",
	}
	return fmt.Sprintf("count=%d mappings=%s", len(pairs), strings.Join(pairs, " ")), nil
}
