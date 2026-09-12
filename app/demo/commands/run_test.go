package commands

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCommandsDemoIntegration(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "prismgo-demo")
	build := exec.Command("go", "build", "-o", executable, "../../..")
	build.Env = append(os.Environ(), "GOCACHE="+filepath.Join(os.TempDir(), "prismgo-demo-go-cache"))
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build Demo executable error = %v, want nil; output = %s", err, output)
	}
	for _, name := range []string{
		"application-entry", "binary-invocation", "list-namespace", "list-json", "list-markdown", "list-raw", "help",
		"serve", "serve-port", "serve-reload", "serve-restart", "serve-stop", "serve-kill",
		"generator-name", "generator-force", "generator-fullpath",
		"make-command", "make-controller", "make-event", "make-job",
		"make-listener", "make-middleware", "make-migration", "make-model", "make-provider", "make-resource", "make-seeder",
		"model-migration", "model-controller", "model-resource", "model-seeder", "model-api", "model-table", "model-composition",
		"model-unsupported-options", "controller-model", "controller-api", "controller-resource", "command-signature", "listener-queued",
		"listener-async", "listener-event", "listener-mode-conflict", "migration-create", "migration-table", "migration-path",
		"migration-realpath", "migration-inference", "stub-publish", "stub-override", "stub-force", "migration-registration",
		"production-guard", "sqlite-extension", "sqlite-fresh",
		"cron", "cron-shutdown", "key-generate", "key-show", "key-force", "key-format",
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()
			result, err := Run(ctx, executable, name)
			if err != nil {
				t.Fatalf("Run(%q) error = %v, want nil", name, err)
			}
			if result.Case != name || result.Command == "" || result.Output == "" {
				t.Fatalf("Run(%q) result = %#v, want case, command, and observed output", name, result)
			}
		})
	}
	t.Run("demo-command-entry", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, "app"), 0o755); err != nil {
			t.Fatalf("prepare application directory error = %v, want nil", err)
		}
		if err := os.WriteFile(filepath.Join(root, ".env"), []byte("APP_ENV=testing\nAPP_KEY=base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=\n"), 0o600); err != nil {
			t.Fatalf("prepare application environment error = %v, want nil", err)
		}
		cmd := exec.Command(executable, "demo:commands", "make-job", "--json")
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "DB_CONNECTION=sqlite", "QUEUE_CONNECTION=sync", "CACHE_STORE=memory", "SESSION_DRIVER=file")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("demo:commands make-job error = %v, want nil; output = %s", err, output)
		}
		var result Result
		if err := json.Unmarshal(output, &result); err != nil || result.Case != "make-job" {
			t.Fatalf("demo:commands result = %#v, JSON error = %v; want make-job", result, err)
		}
	})
}

func TestCommandsDemoMySQLIntegration(t *testing.T) {
	if os.Getenv("PRISMGO_COMMANDS_MYSQL_TEST_DSN") == "" {
		t.Skip("PRISMGO_COMMANDS_MYSQL_TEST_DSN is required for dedicated MySQL integration")
	}
	executable := filepath.Join(t.TempDir(), "prismgo-demo")
	build := exec.Command("go", "build", "-o", executable, "../../..")
	build.Env = append(os.Environ(), "GOCACHE="+filepath.Join(os.TempDir(), "prismgo-demo-go-cache"))
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build Demo executable error = %v, want nil; output = %s", err, output)
	}
	for _, name := range []string{
		"migrate", "migrate-install", "migrate-status", "migrate-rollback", "migrate-reset", "migrate-refresh", "migrate-fresh", "db-seed",
		"database-option", "migration-force", "migration-path-option", "migration-realpath-option", "migration-pretend",
		"migration-seed", "migration-seeder", "migrate-step", "rollback-step", "rollback-batch", "fresh-drop-views", "seed-class",
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			result, err := Run(ctx, executable, name)
			if err != nil {
				t.Fatalf("Run(%q) error = %v, want nil", name, err)
			}
			if result.Case != name || result.Command == "" || result.Output == "" {
				t.Fatalf("Run(%q) result = %#v, want case, command, and observed output", name, result)
			}
		})
	}
}

func TestCommandsDemoRedisIntegration(t *testing.T) {
	if os.Getenv("PRISMGO_REDIS_TEST_URL") == "" {
		t.Skip("PRISMGO_REDIS_TEST_URL is required for Redis integration")
	}
	executable := filepath.Join(t.TempDir(), "prismgo-demo")
	build := exec.Command("go", "build", "-o", executable, "../../..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build Demo executable error = %v, want nil; output = %s", err, output)
	}
	for _, name := range []string{
		"queue", "queue-work", "queue-failed", "queue-retry", "queue-forget", "queue-flush", "queue-restart",
		"worker-connection", "worker-queue", "worker-once", "worker-stop-when-empty", "worker-sleep", "worker-timeout",
		"worker-tries", "worker-backoff", "worker-max-jobs", "worker-max-time", "worker-retry-after",
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			result, err := Run(ctx, executable, name)
			if err != nil {
				t.Fatalf("Run(%q) error = %v, want nil", name, err)
			}
			if result.Case != name || result.Command == "" || result.Output == "" {
				t.Fatalf("Run(%q) result = %#v, want case, command, and observed output", name, result)
			}
		})
	}
}

func TestCommandsDemoPostgresIntegration(t *testing.T) {
	if os.Getenv("PRISMGO_POSTGRES_TEST_DSN") == "" {
		t.Skip("PRISMGO_POSTGRES_TEST_DSN is required for PostgreSQL integration")
	}
	executable := filepath.Join(t.TempDir(), "prismgo-demo")
	build := exec.Command("go", "build", "-o", executable, "../../..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build Demo executable error = %v, want nil; output = %s", err, output)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := Run(ctx, executable, "fresh-drop-types")
	if err != nil {
		t.Fatalf("Run(fresh-drop-types) error = %v, want nil", err)
	}
	if result.Case != "fresh-drop-types" || result.Command == "" || !strings.Contains(result.Output, "verified PostgreSQL enum and table removed") {
		t.Fatalf("Run(fresh-drop-types) result = %#v, want verified enum and table removal", result)
	}
}
