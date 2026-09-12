// Package loggerdemo contains runnable logger documentation examples.
package loggerdemo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/prismgo/framework/config"
	"github.com/prismgo/framework/foundation"
	"github.com/prismgo/framework/logger"
	"github.com/sirupsen/logrus"
)

// Result records an observable logger scenario outcome.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

// driverSequence gives each demo registration a unique process-wide name.
var driverSequence atomic.Uint64

// bufferDriver exposes writes and closure for isolated driver scenarios.
type bufferDriver struct {
	bytes.Buffer
	closed bool
}

func (d *bufferDriver) Close() error { d.closed = true; return nil }

func bufferedManager(opts logger.ChannelOptions) (*logger.Manager, *bufferDriver, error) {
	d := &bufferDriver{}
	name := fmt.Sprintf("demo-logger-%d", driverSequence.Add(1))
	logger.Extend(name, func(logger.ChannelOptions) (logger.Driver, error) { return d, nil })
	opts.Driver = name
	m, err := logger.NewManager(logger.Config{Default: "demo", Channels: map[string]logger.ChannelOptions{"demo": opts}})
	return m, d, err
}
func manager(opts logger.Config) (*logger.Manager, error) { return logger.NewManager(opts) }
func file(path string) (string, error)                    { data, err := os.ReadFile(path); return string(data), err }

// Run executes a logger example selected by its Catalog case name.
func Run(name string) (Result, error) {
	var value string
	var err error
	switch name {
	case "architecture", "config", "top-level-config", "channel-config", "deployment-config":
		value, err = configuration(name)
	case "single-driver", "daily-driver", "stderr-driver", "null-driver", "stack", "stack-channel-isolation", "levels", "facade", "formatted-facade", "fatal", "with-field", "with-fields", "with-error", "error-stacktrace", "context-extractor", "context-noop", "named-channel", "missing-channel", "channel-context", "resolve-manager", "default-name", "driver-contract", "custom-driver", "custom-driver-replacement", "custom-formatter", "formatter-params", "line-formatter", "text-formatter", "json-formatter", "manual-manager", "provider-lifecycle", "provider-cleanup", "resource-cleanup", "closed-writes", "global-logrus":
		value, err = scenario(name)
	default:
		return Result{}, fmt.Errorf("unknown logger scenario %q", name)
	}
	if err != nil {
		return Result{}, fmt.Errorf("logger scenario %q: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}
func configuration(name string) (string, error) {
	switch name {
	case "architecture", "top-level-config":
		m, err := manager(logger.Config{Default: "sink", Channels: map[string]logger.ChannelOptions{"sink": {Driver: "null", Level: "info"}}})
		if err != nil {
			return "", err
		}
		defer m.Close()
		if name == "architecture" {
			return fmt.Sprintf("manager=%T; channel=%T", m, m.Default()), nil
		}
		return "default=" + m.DefaultName(), nil
	case "config":
		return fmt.Sprintf("default=%s; single=%v; path=%v", config.GetString("logging.default"), config.Get("logging.channels.single.driver"), config.Get("logging.channels.single.path")), nil
	case "channel-config":
		opts := logger.ChannelOptions{Driver: "single", Formatter: "json", Level: "warn", Path: "storage/logs/app.log", Channels: []string{"sink"}}
		return fmt.Sprintf("driver=%s; formatter=%s; level=%s; path=%s; children=%d", opts.Driver, opts.Formatter, opts.Level, opts.Path, len(opts.Channels)), nil
	case "deployment-config":
		return fmt.Sprintf("single=%v; error=%v; stderr=%v", config.Get("logging.channels.single.level"), config.Get("logging.channels.error.level"), config.Get("logging.channels.stderr.level")), nil
	}
	return "", fmt.Errorf("unknown configuration case %q", name)
}
func scenario(name string) (string, error) {
	switch name {
	case "single-driver", "daily-driver":
		return fileScenario(name)
	case "stderr-driver":
		old := os.Stderr
		r, w, err := os.Pipe()
		if err != nil {
			return "", err
		}
		os.Stderr = w
		defer func() { os.Stderr = old; r.Close(); w.Close() }()
		m, err := manager(logger.Config{Default: "demo", Channels: map[string]logger.ChannelOptions{"demo": {Driver: "stderr", Level: "info"}}})
		if err != nil {
			return "", err
		}
		m.Default().Info("stderr marker")
		if err := m.Close(); err != nil {
			return "", err
		}
		if err := w.Close(); err != nil {
			return "", err
		}
		out, err := io.ReadAll(r)
		return fmt.Sprintf("stderr=%t", strings.Contains(string(out), "stderr marker")), err
	case "null-driver":
		old := os.Stderr
		r, w, err := os.Pipe()
		if err != nil {
			return "", err
		}
		os.Stderr = w
		defer func() { os.Stderr = old; r.Close(); w.Close() }()
		m, err := manager(logger.Config{Default: "demo", Channels: map[string]logger.ChannelOptions{"demo": {Driver: "null", Level: "info"}}})
		if err != nil {
			return "", err
		}
		m.Default().Info("discarded")
		if err := m.Close(); err != nil {
			return "", err
		}
		if err := w.Close(); err != nil {
			return "", err
		}
		out, err := io.ReadAll(r)
		return fmt.Sprintf("suppressed=%t", !strings.Contains(string(out), "discarded")), err
	case "stack", "stack-channel-isolation", "named-channel", "missing-channel", "channel-context":
		return multiChannel(name)
	case "fatal":
		return fatalProcess()
	case "facade", "formatted-facade", "resolve-manager", "provider-lifecycle", "provider-cleanup", "global-logrus":
		return application(name)
	case "driver-contract":
		var _ logger.Driver = (*bufferDriver)(nil)
		return "Driver=Write+Close", nil
	case "custom-driver-replacement":
		id := fmt.Sprintf("demo-replace-%d", driverSequence.Add(1))
		old := &bufferDriver{}
		newDriver := &bufferDriver{}
		logger.Extend(id, func(logger.ChannelOptions) (logger.Driver, error) { return old, nil })
		logger.Extend(id, func(logger.ChannelOptions) (logger.Driver, error) { return newDriver, nil })
		m, err := manager(logger.Config{Default: "demo", Channels: map[string]logger.ChannelOptions{"demo": {Driver: id, Level: "info"}}})
		if err != nil {
			return "", err
		}
		m.Default().Info("replacement marker")
		if err := m.Close(); err != nil {
			return "", err
		}
		return fmt.Sprintf("old=%t; replacement=%t", strings.Contains(old.String(), "replacement marker"), strings.Contains(newDriver.String(), "replacement marker")), nil
	case "custom-formatter", "formatter-params":
		return formatterScenario(name)
	case "resource-cleanup", "closed-writes":
		m, d, err := bufferedManager(logger.ChannelOptions{Level: "info"})
		if err != nil {
			return "", err
		}
		held := m.Default()
		held.Info("before close")
		if err := m.Close(); err != nil {
			return "", err
		}
		before := d.String()
		m.Default().Info("after close")
		if name == "resource-cleanup" {
			return fmt.Sprintf("closed=%t; written=%t", d.closed, strings.Contains(before, "before close")), nil
		}
		return fmt.Sprintf("unchanged=%t", d.String() == before), nil
	case "context-extractor":
		return contextExtraction()
	default:
		return basic(name)
	}
}
func fileScenario(name string) (string, error) {
	dir, err := os.MkdirTemp("", "prismgo-logger-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "nested", "app.log")
	opts := logger.ChannelOptions{Driver: "single", Level: "info", Path: path}
	if name == "daily-driver" {
		day := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
		opts.Driver = "daily"
		opts.Now = func() time.Time { return day }
		m, err := manager(logger.Config{Default: "demo", Channels: map[string]logger.ChannelOptions{"demo": opts}})
		if err != nil {
			return "", err
		}
		m.Default().Info("first day")
		day = day.Add(24 * time.Hour)
		m.Default().Info("second day")
		if err := m.Close(); err != nil {
			return "", err
		}
		first, e1 := file(filepath.Join(dir, "nested", "app-2026-01-02.log"))
		second, e2 := file(filepath.Join(dir, "nested", "app-2026-01-03.log"))
		return fmt.Sprintf("rotated=%t", strings.Contains(first, "first day") && !strings.Contains(first, "second day") && strings.Contains(second, "second day")), errors.Join(e1, e2)
	}
	m, err := manager(logger.Config{Default: "demo", Channels: map[string]logger.ChannelOptions{"demo": opts}})
	if err != nil {
		return "", err
	}
	m.Default().Info("single marker")
	if err := m.Close(); err != nil {
		return "", err
	}
	out, err := file(path)
	return fmt.Sprintf("created=%t", strings.Contains(out, "single marker")), err
}

type contextKey struct{}

func multiChannel(name string) (string, error) {
	first := &bufferDriver{}
	second := &bufferDriver{}
	firstName := fmt.Sprintf("demo-first-%d", driverSequence.Add(1))
	secondName := fmt.Sprintf("demo-second-%d", driverSequence.Add(1))
	logger.Extend(firstName, func(logger.ChannelOptions) (logger.Driver, error) { return first, nil })
	logger.Extend(secondName, func(logger.ChannelOptions) (logger.Driver, error) { return second, nil })
	m, err := manager(logger.Config{Default: "stack", Channels: map[string]logger.ChannelOptions{
		"stack":  {Driver: "stack", Channels: []string{"first", "second"}},
		"first":  {Driver: firstName, Formatter: "line", Level: "debug"},
		"second": {Driver: secondName, Formatter: "json", Level: "error"},
	}})
	if err != nil {
		return "", err
	}
	defer m.Close()
	switch name {
	case "stack":
		m.Default().Error("stack marker")
		return fmt.Sprintf("first=%t; second=%t", strings.Contains(first.String(), "stack marker"), strings.Contains(second.String(), "stack marker")), nil
	case "stack-channel-isolation":
		m.Default().Info("info marker")
		m.Default().Error("error marker")
		return fmt.Sprintf("first_info=%t; second_info=%t; second_json=%t", strings.Contains(first.String(), "info marker"), strings.Contains(second.String(), "info marker"), strings.Contains(second.String(), `"msg":"error marker"`)), nil
	case "named-channel":
		m.Channel("first").Info("first only")
		return fmt.Sprintf("first=%t; second=%t", strings.Contains(first.String(), "first only"), strings.Contains(second.String(), "first only")), nil
	case "missing-channel":
		m.Channel("missing").Error("fallback marker")
		return fmt.Sprintf("fallback=%t", strings.Contains(first.String(), "fallback marker") && strings.Contains(second.String(), "fallback marker")), nil
	case "channel-context":
		m.Channel("second").WithField("tenant", "blue").Error("context marker")
		return fmt.Sprintf("tenant=%t; isolated=%t", strings.Contains(second.String(), `"tenant":"blue"`), !strings.Contains(first.String(), "context marker")), nil
	}
	return "", fmt.Errorf("unknown multi-channel case %q", name)
}
func contextExtraction() (string, error) {
	first := &bufferDriver{}
	second := &bufferDriver{}
	firstName := fmt.Sprintf("demo-context-%d", driverSequence.Add(1))
	secondName := fmt.Sprintf("demo-context-%d", driverSequence.Add(1))
	logger.Extend(firstName, func(logger.ChannelOptions) (logger.Driver, error) { return first, nil })
	logger.Extend(secondName, func(logger.ChannelOptions) (logger.Driver, error) { return second, nil })
	m, err := manager(logger.Config{
		Default: "first",
		ContextExtractor: func(ctx context.Context) map[string]any {
			return map[string]any{"scope": "manager", "request_id": ctx.Value(contextKey{})}
		},
		Channels: map[string]logger.ChannelOptions{
			"first": {Driver: firstName, Level: "info"},
			"second": {Driver: secondName, Level: "info", ContextExtractor: func(context.Context) map[string]any {
				return map[string]any{"scope": "channel"}
			}},
		},
	})
	if err != nil {
		return "", err
	}
	defer m.Close()
	ctx := context.WithValue(context.Background(), contextKey{}, "req-42")
	m.Default().WithContext(ctx).Info("manager context")
	m.Channel("second").WithContext(ctx).Info("channel context")
	return fmt.Sprintf("manager=%t; override=%t", strings.Contains(first.String(), `"scope":"manager"`) && strings.Contains(first.String(), `"request_id":"req-42"`), strings.Contains(second.String(), `"scope":"channel"`) && !strings.Contains(second.String(), "request_id")), nil
}

func basic(name string) (string, error) {
	opts := logger.ChannelOptions{Level: "debug"}
	if name == "text-formatter" {
		opts.Formatter = "text"
	}
	if name == "json-formatter" {
		opts.Formatter = "json"
	}
	m, d, err := bufferedManager(opts)
	if err != nil {
		return "", err
	}
	defer m.Close()
	lg := m.Default()
	switch name {
	case "levels":
		m.Close()
		m, d, err = bufferedManager(logger.ChannelOptions{Level: "warn"})
		if err != nil {
			return "", err
		}
		defer m.Close()
		m.Default().Info("filtered marker")
		m.Default().Warn("warning marker")
		return fmt.Sprintf("info=%t; warn=%t", strings.Contains(d.String(), "filtered marker"), strings.Contains(d.String(), "warning marker")), nil
	case "with-field":
		lg.WithField("order_id", 42).Info("field marker")
		return fmt.Sprintf("field=%t", strings.Contains(d.String(), `"order_id":42`)), nil
	case "with-fields":
		lg.WithFields(map[string]any{"order_id": 42, "tenant": "blue"}).Info("fields marker")
		return fmt.Sprintf("order=%t; tenant=%t", strings.Contains(d.String(), `"order_id":42`), strings.Contains(d.String(), `"tenant":"blue"`)), nil
	case "with-error":
		lg.WithError(errors.New("disk unavailable")).Error("error marker")
		return fmt.Sprintf("error=%t", strings.Contains(d.String(), `"error":"disk unavailable"`)), nil
	case "error-stacktrace":
		lg.WithError(errors.New("stack failure")).Error("stack marker")
		return fmt.Sprintf("error=%t; stack=%t", strings.Contains(d.String(), "stack failure"), strings.Contains(d.String(), "[stacktrace]")), nil
	case "context-noop":
		lg.WithContext(context.WithValue(context.Background(), contextKey{}, "req-42")).Info("context marker")
		return fmt.Sprintf("request_absent=%t", !strings.Contains(d.String(), "req-42")), nil
	case "default-name":
		return "default=" + m.DefaultName(), nil
	case "custom-driver":
		lg.Info("custom marker")
		return fmt.Sprintf("written=%t", strings.Contains(d.String(), "custom marker")), nil
	case "line-formatter":
		lg.Info("line marker")
		return fmt.Sprintf("line=%t", strings.Contains(d.String(), "demo.INFO: line marker")), nil
	case "text-formatter":
		lg.WithField("tenant", "blue").Info("text marker")
		return fmt.Sprintf("text=%t", strings.Contains(d.String(), "level=info") && strings.Contains(d.String(), "tenant=blue")), nil
	case "json-formatter":
		lg.WithField("tenant", "blue").Info("json marker")
		var row map[string]any
		if err := json.Unmarshal(bytes.TrimSpace(d.Bytes()), &row); err != nil {
			return "", err
		}
		return fmt.Sprintf("level=%v; tenant=%v", row["level"], row["tenant"]), nil
	case "manual-manager":
		lg.Info("manual marker")
		return fmt.Sprintf("default=%s; written=%t", m.DefaultName(), strings.Contains(d.String(), "manual marker")), nil
	}
	return "", fmt.Errorf("unknown basic case %q", name)
}

type prefixFormatter struct{ prefix string }

func (f prefixFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(f.prefix + entry.Message + "\n"), nil
}
func formatterScenario(name string) (string, error) {
	id := fmt.Sprintf("demo-formatter-%d", driverSequence.Add(1))
	var received map[string]any
	logger.RegisterFormatter(id, func(params map[string]any) (logger.Formatter, error) {
		received = params
		return prefixFormatter{prefix: fmt.Sprint(params["prefix"])}, nil
	})
	m, d, err := bufferedManager(logger.ChannelOptions{Level: "info", Formatter: id, FormatterParams: map[string]any{"prefix": "demo:"}})
	if err != nil {
		return "", err
	}
	defer m.Close()
	m.Default().Info("formatted marker")
	if name == "custom-formatter" {
		return fmt.Sprintf("output=%t", d.String() == "demo:formatted marker\n"), nil
	}
	return fmt.Sprintf("channel=%v; prefix=%v", received["channel"], received["prefix"]), nil
}
func application(name string) (string, error) {
	if name == "provider-cleanup" {
		return providerCleanup()
	}
	if name == "provider-lifecycle" {
		app := foundation.App
		if app == nil {
			return "", fmt.Errorf("application is unavailable")
		}
		bound := app.Container().Bound("logger.manager")
		before := app.Container().Resolved("logger.manager")
		m := logger.Resolve()
		return fmt.Sprintf("bound=%t; lazy=%t; singleton=%t", bound, !before && app.Container().Resolved("logger.manager"), m == logger.Resolve()), nil
	}
	m := logger.Resolve()
	if m == nil {
		return "", fmt.Errorf("logger manager is unavailable")
	}
	switch name {
	case "resolve-manager":
		return fmt.Sprintf("manager=%T", m), nil
	case "facade", "formatted-facade", "global-logrus":
		// The current application owns the facade manager; use its configured single file.
		path := config.GetString("logging.channels.single.path")
		if path == "" {
			return "", fmt.Errorf("single channel path is empty")
		}
		marker := fmt.Sprintf("demo-logger-%d", driverSequence.Add(1))
		if name == "facade" {
			logger.Info(marker)
		} else if name == "formatted-facade" {
			logger.Infof("formatted %s", marker)
		} else {
			logrus.Info(marker)
		}
		out, readErr := file(path)
		if readErr != nil {
			return "", readErr
		}
		return fmt.Sprintf("recorded=%t", strings.Contains(out, marker)), nil
	}
	return "", fmt.Errorf("unknown application case %q", name)
}

// providerCleanup verifies application-owned driver cleanup in an isolated process.
func providerCleanup() (string, error) {
	dir, err := os.MkdirTemp("", "prismgo-logger-provider-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "provider.log")
	source := filepath.Join(dir, "main.go")
	code := `package main

import (
	"fmt"
	"os"

	"github.com/prismgo/framework/foundation"
	"github.com/prismgo/framework/logger"

	// Register the Demo logging configuration for this isolated application.
	_ "prismgo-demo/config"
)

func main() {
	app := foundation.Configure(os.Args[1]).Create()
	if err := app.Boot(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	m := logger.Resolve()
	m.Default().Info("provider marker")
	if err := app.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	before, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	m.Default().Info("late marker")
	after, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Printf("cleanup=%t", string(before) == string(after))
}`
	if err := os.WriteFile(source, []byte(code), 0o600); err != nil {
		return "", err
	}
	command := exec.Command("go", "run", source, dir, path)
	command.Env = append(os.Environ(), "APP_LOGGER_FILE="+path, "LOG_CHANNEL=single", "ERROR_LOGGER_DRIVER=null")
	out, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("provider cleanup child: %w: %s", err, out)
	}
	if !strings.Contains(string(out), "cleanup=true") {
		return "", fmt.Errorf("provider cleanup output = %q, want cleanup=true", out)
	}
	return "cleanup=true", nil
}
func fatalProcess() (string, error) {
	dir, err := os.MkdirTemp("", "prismgo-logger-fatal-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "fatal.log")
	source := filepath.Join(dir, "main.go")
	code := fmt.Sprintf(`package main

import (
	"fmt"
	"os"

	"github.com/prismgo/framework/logger"
)

func main() {
	m, err := logger.NewManager(logger.Config{
		Default: "demo",
		Channels: map[string]logger.ChannelOptions{
			"demo": {Driver: "single", Level: "info", Path: %q},
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	m.Default().Fatal("fatal marker")
}
`, path)
	if err := os.WriteFile(source, []byte(code), 0o600); err != nil {
		return "", err
	}
	command := exec.Command("go", "run", source)
	output, runErr := command.CombinedOutput()
	logged, readErr := file(path)
	if readErr != nil {
		return "", readErr
	}
	var exitErr *exec.ExitError
	if !errors.As(runErr, &exitErr) {
		return "", fmt.Errorf("fatal child error = %v, output = %q; want exit error", runErr, output)
	}
	return fmt.Sprintf("logged=%t; exited=%t", strings.Contains(logged, "fatal marker"), exitErr.ExitCode() != 0), nil
}
