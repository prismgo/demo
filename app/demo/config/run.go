// Package configdemo contains runnable examples of configuration registration and reads.
package configdemo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/prismgo/framework/config"
	"github.com/prismgo/framework/container"
	containercontract "github.com/prismgo/framework/contracts/container"

	// Register the demo application's mail namespace before constructing a repository.
	_ "prismgo-demo/config"
)

// Result records the observable value returned by a configuration example.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

func init() {
	config.Add("config_demo", func() map[string]any {
		return map[string]any{
			"priority": config.Env("CONFIG_DEMO_PRIORITY", "default"),
			"values": map[string]any{
				"int64":       int64(9000000001),
				"uint":        uint(17),
				"float64":     1.25,
				"enabled":     true,
				"labels":      map[string]string{"region": "local", "tier": "demo"},
				"env_bool":    config.Env("CONFIG_DEMO_BOOL", false),
				"env_int":     config.Env("CONFIG_DEMO_INT", 0),
				"env_int64":   config.Env("CONFIG_DEMO_INT64", int64(0)),
				"env_uint":    config.Env("CONFIG_DEMO_UINT", uint(0)),
				"env_float64": config.Env("CONFIG_DEMO_FLOAT64", float64(0)),
				"env_string":  config.Env("CONFIG_DEMO_STRING", "fallback"),
			},
		}
	})
}

// Run executes one documentation-backed configuration example.
func Run(name string) (Result, error) {
	value, err := run(name)
	if err != nil {
		return Result{}, fmt.Errorf("config demo %s: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

func run(name string) (string, error) {
	switch name {
	case "lazy-loading":
		return lazyLoading()
	case "registration-timing":
		first, err := fileValue("MAIL_HOST=first-load\n", "mail.host")
		if err != nil {
			return "", err
		}
		second, err := fileValue("MAIL_HOST=second-load\n", "mail.host")
		return first + " -> " + second, err
	case "missing-env-file":
		return missingEnvFile()
	case "env-precedence":
		return fileValue("CONFIG_DEMO_PRIORITY=from-file\n", "config_demo.priority")
	case "env":
		return fileValue("CONFIG_DEMO_STRING=from-env\n", "config_demo.values.env_string")
	case "env-default":
		return fmt.Sprintf("missing=%v; blank=%v", config.Env("CONFIG_DEMO_UNSET", "fallback"), config.Env("", "fallback")), nil
	case "env-bool":
		return typedEnvValue("CONFIG_DEMO_BOOL=true\n", "env_bool")
	case "env-int":
		return typedEnvValue("CONFIG_DEMO_INT=42\n", "env_int")
	case "env-int64":
		return typedEnvValue("CONFIG_DEMO_INT64=9000000001\n", "env_int64")
	case "env-uint":
		return typedEnvValue("CONFIG_DEMO_UINT=17\n", "env_uint")
	case "env-float64":
		return typedEnvValue("CONFIG_DEMO_FLOAT64=1.25\n", "env_float64")
	case "env-string":
		return typedEnvValue("CONFIG_DEMO_STRING=hello\n", "env_string")
	case "app-registration":
		return fileValue("APP_NAME=registered-app\n", "app.name")
	case "runtime-scope":
		first, err := repositoryFromContent("CONFIG_DEMO_PRIORITY=first\n")
		if err != nil {
			return "", err
		}
		second, err := repositoryFromContent("CONFIG_DEMO_PRIORITY=second\n")
		if err != nil {
			return "", err
		}
		return first.GetString("config_demo.priority") + " -> " + second.GetString("config_demo.priority"), nil
	case "new-from-file", "standalone-read", "clone-isolation", "resolve", "facade-clone", "reload", "empty", "test-binding", "conventions", "mutable-settings":
		return runInstanceCase(name)
	}
	if _, ok := settingFor(name); ok {
		return runApplicationSetting(name)
	}

	cfg, err := repositoryFromContent("")
	if err != nil {
		return "", fmt.Errorf("construct repository: %w", err)
	}
	switch name {
	case "quick-start":
		return fmt.Sprintf("%s:%d", cfg.GetString("mail.host"), cfg.GetInt("mail.port")), nil
	case "add":
		return fmt.Sprintf("mail: %d fields", len(cfg.GetStringMap("mail"))), nil
	case "nested-path":
		return cfg.GetString("mail.from.address"), nil
	case "map-normalization":
		_, ok := cfg.GetStringMap("config_demo.values")["labels"].(map[string]any)
		return fmt.Sprintf("nested map[string]any=%t", ok), nil
	case "get-string":
		return fmt.Sprintf("Get=%s; GetString=%s", cfg.Get("mail.host"), cfg.GetString("mail.host")), nil
	case "get-int":
		return fmt.Sprintf("%d", cfg.GetInt("mail.port")), nil
	case "get-int64":
		return fmt.Sprintf("%d", cfg.GetInt64("config_demo.values.int64")), nil
	case "get-uint":
		return fmt.Sprintf("%d", cfg.GetUint("config_demo.values.uint")), nil
	case "get-float64":
		return fmt.Sprintf("%.2f", cfg.GetFloat64("config_demo.values.float64")), nil
	case "get-bool":
		return fmt.Sprintf("%t", cfg.GetBool("config_demo.values.enabled")), nil
	case "get-string-map":
		return fmt.Sprintf("mail fields=%d", len(cfg.GetStringMap("mail"))), nil
	case "get-string-map-string":
		return cfg.GetStringMapString("config_demo.values.labels")["region"], nil
	case "get-default":
		return fmt.Sprintf("%s/%d/%t", cfg.GetString("missing.path", "fallback"), cfg.GetInt("missing.path", 7), cfg.GetBool("missing.path", true)), nil
	case "missing-map":
		anyMap := cfg.GetStringMap("missing.path")
		stringMap := cfg.GetStringMapString("missing.path")
		return fmt.Sprintf("any=%t:%d; string=%t:%d", anyMap != nil, len(anyMap), stringMap != nil, len(stringMap)), nil
	case "app-name":
		return cfg.GetString("app.name"), nil
	case "app-env":
		return cfg.GetString("app.env"), nil
	default:
		return "", fmt.Errorf("unknown scenario %q", name)
	}
}

func fileValue(content, path string) (string, error) {
	cfg, err := repositoryFromContent(content)
	if err != nil {
		return "", err
	}
	return cfg.GetString(path), nil
}

func typedEnvValue(content, key string) (string, error) {
	cfg, err := repositoryFromContent(content)
	if err != nil {
		return "", err
	}
	value := cfg.GetStringMap("config_demo.values")[key]
	return fmt.Sprintf("%T:%v", value, value), nil
}

func repositoryFromContent(content string) (*config.Config, error) {
	file, err := os.CreateTemp("", "prismgo-config-demo-*.env")
	if err != nil {
		return nil, fmt.Errorf("create env file: %w", err)
	}
	defer os.Remove(file.Name())
	if _, err := file.WriteString(content); err != nil {
		file.Close()
		return nil, fmt.Errorf("write env file: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("close env file: %w", err)
	}
	cfg, err := config.NewFromFile(file.Name())
	if err != nil {
		return nil, fmt.Errorf("load env file: %w", err)
	}
	return cfg, nil
}

func missingEnvFile() (string, error) {
	directory, err := os.MkdirTemp("", "prismgo-config-demo-missing-*")
	if err != nil {
		return "", fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(directory)
	path := filepath.Join(directory, ".env")
	cfg, err := config.NewFromFile(path)
	if err != nil {
		return "", fmt.Errorf("load missing env file: %w", err)
	}
	return cfg.GetString("mail.host"), nil
}

type providerApp struct{ registry containercontract.Container }

func (a providerApp) Container() containercontract.Container { return a.registry }

func lazyLoading() (string, error) {
	registry := container.NewContainer()
	if err := (config.ServiceProvider{}).Register(providerApp{registry: registry}); err != nil {
		return "", fmt.Errorf("register provider: %w", err)
	}
	before := registry.Resolved("config.default")
	_, err := registry.Make("config.default")
	if err != nil {
		return "", fmt.Errorf("resolve repository: %w", err)
	}
	return fmt.Sprintf("resolved before=%t; after=%t", before, registry.Resolved("config.default")), nil
}
