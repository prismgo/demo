package commands

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
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
