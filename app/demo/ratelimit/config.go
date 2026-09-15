package ratelimitdemo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	configpkg "github.com/prismgo/framework/config"

	// Register the demo application's config namespaces so cache.limiter.* can be resolved.
	_ "prismgo-demo/config"
)

// configScenario resolves the documented limiter cache configuration items from an env file.
func configScenario() (string, error) {
	return withTempConfig(map[string]string{
		"CACHE_STORE":          "file",
		"CACHE_LIMITER_DRIVER": "redis",
	}, func(config *configpkg.Config) (string, error) {
		store := strings.TrimSpace(config.GetString("cache.default", ""))
		driver := strings.TrimSpace(config.GetString("cache.limiter.driver", ""))
		return fmt.Sprintf("store=%s driver=%s", store, driver), nil
	})
}

// configRegistrationScenario proves the cache namespace registers cache.limiter.driver and stores.
func configRegistrationScenario() (string, error) {
	return withTempConfig(map[string]string{
		"CACHE_LIMITER_DRIVER": "demo-plugin",
	}, func(config *configpkg.Config) (string, error) {
		driver := strings.TrimSpace(config.GetString("cache.limiter.driver", ""))
		stores := len(config.GetStringMap("cache.stores"))
		return fmt.Sprintf("namespace=cache driver=%s stores=%d", driver, stores), nil
	})
}

// withTempConfig loads a temporary .env file and runs inspect with the resulting config.
func withTempConfig(values map[string]string, inspect func(*configpkg.Config) (string, error)) (string, error) {
	root, err := os.MkdirTemp("", "prismgo-ratelimit-config-")
	if err != nil {
		return "", fmt.Errorf("create config temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(root) }()

	lines := make([]string, 0, len(values))
	for key, value := range values {
		lines = append(lines, key+"="+value)
	}
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		return "", fmt.Errorf("write config env file: %w", err)
	}
	config, err := configpkg.NewFromFile(path)
	if err != nil {
		return "", fmt.Errorf("load config env file: %w", err)
	}
	return inspect(config)
}
