// Package commands runs isolated, documentation-backed application command examples.
package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Result records the command invocation and the observed behavior.
type Result struct {
	Case    string `json:"case"`
	Command string `json:"command"`
	Output  string `json:"output"`
}

// Run executes a commands example with the supplied Demo binary in an isolated directory.
func Run(ctx context.Context, executable, name string) (Result, error) {
	if executable == "" {
		return Result{}, fmt.Errorf("commands demo: executable is required")
	}
	root, err := os.MkdirTemp("", "prismgo-commands-*")
	if err != nil {
		return Result{}, fmt.Errorf("create commands demo directory: %w", err)
	}
	defer os.RemoveAll(root)
	if err := os.Mkdir(filepath.Join(root, "app"), 0o755); err != nil {
		return Result{}, fmt.Errorf("prepare commands demo directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("APP_ENV=testing\nAPP_KEY=base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=\n"), 0o600); err != nil {
		return Result{}, fmt.Errorf("prepare commands demo environment: %w", err)
	}

	switch name {
	case "application-entry", "binary-invocation", "list-namespace", "list-json", "list-markdown", "list-raw", "help":
		return runListing(ctx, executable, root, name)
	case "serve", "serve-port", "serve-reload", "serve-restart", "serve-stop", "serve-kill":
		return runServer(ctx, executable, root, name)
	case "generator-name", "generator-force", "generator-fullpath", "make-command", "make-controller", "make-event", "make-job":
		return runGenerator(ctx, executable, root, name)
	default:
		return Result{}, fmt.Errorf("commands demo: unknown case %q", name)
	}
}

func runListing(ctx context.Context, executable, root, name string) (Result, error) {
	args := map[string][]string{
		"application-entry": {"demo:list", "commands", "--json"},
		"binary-invocation": {"list", "--raw"},
		"list-namespace":    {"list", "make", "--raw"},
		"list-json":         {"list", "--format=json"},
		"list-markdown":     {"list", "--format=md"},
		"list-raw":          {"list", "--raw"},
		"help":              {"help", "migrate"},
	}[name]
	output, err := invoke(ctx, executable, root, nil, args...)
	if err != nil {
		return Result{}, err
	}
	switch name {
	case "application-entry":
		var details struct {
			Entries []map[string]any `json:"entries"`
		}
		if err := json.Unmarshal([]byte(output), &details); err != nil || len(details.Entries) != 101 {
			return Result{}, fmt.Errorf("application entry output has %d entries, want 101: %v", len(details.Entries), err)
		}
		output = fmt.Sprintf("app.HandleCommand routed os.Args to demo:list: %d catalog entries", len(details.Entries))
	case "list-json":
		if !json.Valid([]byte(output)) || !strings.Contains(output, `"name"`) {
			return Result{}, fmt.Errorf("list JSON output = %q, want command descriptors", output)
		}
	case "list-namespace":
		if !strings.Contains(output, "make:command") || strings.Contains(output, "migrate:install") {
			return Result{}, fmt.Errorf("namespace output = %q, want only make commands", output)
		}
	case "list-markdown":
		if !strings.Contains(output, "# PrismGo") || !strings.Contains(output, "- `make:command` -") {
			return Result{}, fmt.Errorf("Markdown output = %q, want command list", output)
		}
	case "list-raw", "binary-invocation":
		if !strings.Contains(output, "make:command Create a new console command") || strings.Contains(output, "Available Commands:") {
			return Result{}, fmt.Errorf("raw output = %q, want ungrouped name and description rows", output)
		}
	case "help":
		if !strings.Contains(output, "migrate") || !strings.Contains(output, "Usage:") {
			return Result{}, fmt.Errorf("help output = %q, want migrate usage", output)
		}
	}
	return Result{Case: name, Command: strings.Join(args, " "), Output: strings.TrimSpace(output)}, nil
}

func runGenerator(ctx context.Context, executable, root, name string) (Result, error) {
	command, artifact, expected := "", "", ""
	switch name {
	case "generator-name":
		command, artifact, expected = "make:command", "Admin/DailyReport", "app/cmd/admin/daily_report.go"
	case "generator-force":
		command, artifact, expected = "make:event", "DemoEvent", "app/events/demo_event.go"
	case "generator-fullpath":
		command, artifact, expected = "make:job", "DemoJob", "app/jobs/demo_job.go"
	case "make-command":
		command, artifact, expected = "make:command", "DemoCommand", "app/cmd/demo_command.go"
	case "make-controller":
		command, artifact, expected = "make:controller", "DemoController", "app/http/controllers/demo_controller.go"
	case "make-event":
		command, artifact, expected = "make:event", "DemoEvent", "app/events/demo_event.go"
	case "make-job":
		command, artifact, expected = "make:job", "DemoJob", "app/jobs/demo_job.go"
	}
	args := []string{command, artifact}
	if name == "generator-fullpath" {
		args = append(args, "--fullpath")
	}
	output, err := invoke(ctx, executable, root, nil, args...)
	if err != nil {
		return Result{}, err
	}
	path := filepath.Join(root, filepath.FromSlash(expected))
	content, err := os.ReadFile(path)
	if err != nil {
		return Result{}, fmt.Errorf("read generated artifact %s: %w", path, err)
	}
	if len(content) == 0 {
		return Result{}, fmt.Errorf("generated artifact %s is empty", path)
	}
	if name == "generator-fullpath" && !strings.Contains(output, path) {
		return Result{}, fmt.Errorf("fullpath output = %q, want %q", output, path)
	}
	if name == "generator-force" {
		if err := os.WriteFile(path, []byte("old content"), 0o644); err != nil {
			return Result{}, fmt.Errorf("prepare overwrite: %w", err)
		}
		args = append(args, "--force")
		output, err = invoke(ctx, executable, root, nil, args...)
		if err != nil {
			return Result{}, err
		}
		content, err = os.ReadFile(path)
		if err != nil {
			return Result{}, fmt.Errorf("read forced artifact %s: %w", path, err)
		}
		if string(content) == "old content" {
			return Result{}, fmt.Errorf("forced artifact content = %q, want regenerated Go file", content)
		}
	}
	return Result{Case: name, Command: strings.Join(args, " "), Output: strings.TrimSpace(output) + "\nverified " + filepath.ToSlash(expected)}, nil
}

func invoke(ctx context.Context, executable, root string, extraEnv []string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = root
	cmd.Env = commandEnvironment(extraEnv)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("run %q: %w: %s", strings.Join(args, " "), err, bytes.TrimSpace(output))
	}
	return string(output), nil
}

func commandEnvironment(extraEnv []string) []string {
	environment := os.Environ()
	environment = append(environment,
		"APP_ENV=testing", "APP_KEY=base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=",
		"DB_CONNECTION=sqlite", "QUEUE_CONNECTION=sync", "CACHE_STORE=memory", "SESSION_DRIVER=file",
		"REDIS_URL=", "RABBITMQ_URL=", "GIN_MODE=release",
	)
	return append(environment, extraEnv...)
}

func unusedPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("allocate HTTP demo port: %w", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func waitHTTP(ctx context.Context, port int, wantUp bool) error {
	client := &http.Client{Timeout: 200 * time.Millisecond}
	url := "http://127.0.0.1:" + strconv.Itoa(port) + "/"
	deadline := time.NewTimer(8 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		response, err := client.Get(url)
		if err == nil {
			response.Body.Close()
		}
		if (err == nil) == wantUp {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return fmt.Errorf("HTTP port %d reachable = %t, want %t", port, err == nil, wantUp)
		case <-tick.C:
		}
	}
}

func runServer(ctx context.Context, executable, root, name string) (Result, error) {
	port, err := unusedPort()
	if err != nil {
		return Result{}, err
	}
	portText := strconv.Itoa(port)
	args := []string{"serve"}
	env := []string{"SERVER_HOST=127.0.0.1", "SERVER_PORT=" + portText, "SERVER_ACCESS_LOG=false"}
	if name == "serve-port" {
		args = append(args, "--port="+portText)
		env = append(env, "SERVER_PORT=1")
	}
	serverCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	cmd := exec.CommandContext(serverCtx, executable, args...)
	cmd.Dir = root
	cmd.Env = commandEnvironment(env)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return Result{}, fmt.Errorf("start serve demo: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = invoke(cleanupCtx, executable, root, env, "serve", "--port="+portText, "--kill")
		cancel()
		<-done
	}()
	if err := waitHTTP(ctx, port, true); err != nil {
		return Result{}, err
	}
	if name == "serve" || name == "serve-port" {
		return Result{Case: name, Command: strings.Join(args, " "), Output: "HTTP request reached configured port " + portText}, nil
	}
	flag := strings.TrimPrefix(name, "serve-")
	control := []string{"serve", "--port=" + portText, "--" + flag}
	output, err := invoke(ctx, executable, root, env, control...)
	if err != nil {
		return Result{}, err
	}
	wantUp := flag == "reload" || flag == "restart"
	if err := waitHTTP(ctx, port, wantUp); err != nil {
		return Result{}, fmt.Errorf("%w; control output = %q", err, output)
	}
	return Result{Case: name, Command: strings.Join(control, " "), Output: strings.TrimSpace(output)}, nil
}
