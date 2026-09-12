package configdemo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/prismgo/framework/config"
	"github.com/prismgo/framework/container"
)

func runInstanceCase(name string) (string, error) {
	switch name {
	case "new-from-file":
		return fileValue("APP_NAME=standalone\n", "app.name")
	case "standalone-read":
		cfg, err := repositoryFromContent("APP_DEBUG=true\nSERVER_PORT=9090\n")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("debug=%t; port=%d", cfg.GetBool("app.debug"), cfg.GetInt("app.server.port")), nil
	case "clone-isolation":
		return cloneIsolation()
	case "resolve":
		return config.Resolve().GetString("app.name"), nil
	case "facade-clone":
		original := config.Resolve()
		clone := config.Clone()
		return fmt.Sprintf("isolated=%t; same-name=%t", clone != original, clone.GetString("app.name") == original.GetString("app.name")), nil
	case "reload":
		original := config.Resolve()
		if err := config.Reload(); err != nil {
			return "", fmt.Errorf("reload current configuration: %w", err)
		}
		return fmt.Sprintf("same-instance=%t; empty=%t", config.Resolve() == original, config.Empty()), nil
	case "empty":
		return fmt.Sprintf("empty=%t", config.Empty()), nil
	case "test-binding":
		return testBinding()
	case "conventions":
		return "app.url <- APP_URL", nil
	case "mutable-settings":
		cfg, err := repositoryFromContent("APP_NAME=static\n")
		if err != nil {
			return "", err
		}
		settings := map[string]string{"label": "first"}
		settings["label"] = "second"
		return fmt.Sprintf("runtime=%s; mutable=%s", cfg.GetString("app.name"), settings["label"]), nil
	default:
		return "", fmt.Errorf("unknown scenario %q", name)
	}
}

func cloneIsolation() (string, error) {
	directory, err := os.MkdirTemp("", "prismgo-config-clone-*")
	if err != nil {
		return "", fmt.Errorf("create clone directory: %w", err)
	}
	defer os.RemoveAll(directory)
	first := filepath.Join(directory, "first.env")
	second := filepath.Join(directory, "second.env")
	if err := os.WriteFile(first, []byte("APP_NAME=before\n"), 0600); err != nil {
		return "", fmt.Errorf("write first env file: %w", err)
	}
	if err := os.WriteFile(second, []byte("APP_NAME=after\n"), 0600); err != nil {
		return "", fmt.Errorf("write second env file: %w", err)
	}
	cfg, err := config.NewFromFile(first)
	if err != nil {
		return "", fmt.Errorf("load first env file: %w", err)
	}
	clone := cfg.Clone()
	if err := cfg.ReloadFromFile(second); err != nil {
		return "", fmt.Errorf("reload second env file: %w", err)
	}
	return clone.GetString("app.name") + " -> " + cfg.GetString("app.name"), nil
}

func testBinding() (string, error) {
	cfg, err := repositoryFromContent("APP_NAME=bound-test\n")
	if err != nil {
		return "", err
	}
	registry := container.NewContainer()
	if err := registry.Instance("config.default", cfg); err != nil {
		return "", fmt.Errorf("bind test configuration: %w", err)
	}
	if err := (config.ServiceProvider{}).Register(providerApp{registry: registry}); err != nil {
		return "", fmt.Errorf("register config provider: %w", err)
	}
	value, err := registry.Make("config.default")
	if err != nil {
		return "", fmt.Errorf("resolve bound test configuration: %w", err)
	}
	bound, ok := value.(*config.Config)
	if !ok {
		return "", fmt.Errorf("bound configuration has type %T", value)
	}
	return fmt.Sprintf("same-instance=%t; name=%s", bound == cfg, bound.GetString("app.name")), nil
}
