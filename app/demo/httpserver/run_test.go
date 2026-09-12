package httpserver_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"prismgo-demo/app/demo/catalog"
	"prismgo-demo/app/demo/httpserver"
	demotest "prismgo-demo/app/demo/testing"
)

func TestHTTPServerDemoInProcess(t *testing.T) {
	for key, value := range map[string]string{
		"SERVER_HOST": "127.0.0.1", "SERVER_PORT": "9123", "SERVER_READ_TIMEOUT": "11s",
		"SERVER_TRUSTED_PROXIES": "127.0.0.1", "SERVER_CLIENT_IP_HEADERS": "X-Real-IP",
		"SERVER_MAX_MULTIPART_MEMORY": "4096", "SERVER_ACCESS_LOG": "false",
	} {
		t.Setenv(key, value)
	}
	demotest.NewApplication(t, demotest.Options{})
	for _, test := range []struct{ name, want string }{
		{name: "bootstrap", want: "*http.Server"},
		{name: "host", want: "127.0.0.1:9123"},
		{name: "read-timeout", want: "11s"},
		{name: "max-multipart-memory", want: "file-on-disk=true"},
		{name: "client-ip-headers", want: "body=198.51.100.9"},
		{name: "untrusted-client-ip", want: "body=192.0.2.7"},
		{name: "access-log", want: "enabled=false logged=false"},
		{name: "exception-handler", want: "status=500"},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := httpserver.Run(context.Background(), "", test.name)
			if err != nil || result.Case != test.name || !strings.Contains(result.Value, test.want) {
				t.Fatalf("Run(%q) = %#v, error = %v; want case %q and value containing %q", test.name, result, err, test.name, test.want)
			}
		})
	}
	if _, err := httpserver.Run(context.Background(), "", "not-a-case"); err == nil || !strings.Contains(err.Error(), `unknown scenario "not-a-case"`) {
		t.Fatalf("Run(\"not-a-case\") error = %v, want descriptive unknown scenario error", err)
	}
}

func TestHTTPServerDemoLocalIntegration(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "prismgo-demo")
	build := exec.Command("go", "build", "-o", executable, "../../..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build Demo executable error = %v, want nil; output = %s", err, output)
	}
	type integrationCase struct {
		name    string
		want    string
		notWant string
		env     []string
	}
	cases := []integrationCase{
		{name: "bootstrap", want: "*http.Server"},
		{name: "routes", want: `status=200 body={"status":"ok"}`},
		{name: "serve", want: "HTTP request reached configured port"},
		{name: "port-flag", want: "HTTP request reached configured port"},
		{name: "port-flag-scope", want: ".env remains SERVER_PORT=8123"},
		{name: "host", want: "127.0.0.1:9123", env: []string{"SERVER_HOST=127.0.0.1", "SERVER_PORT=9123"}},
		{name: "port", want: "9123", env: []string{"SERVER_PORT=9123"}},
		{name: "timeout-fallback", want: "read=7s write=7s shutdown=7s"},
		{name: "read-timeout", want: "11s", env: []string{"SERVER_READ_TIMEOUT=11s"}},
		{name: "read-header-timeout", want: "4s", env: []string{"SERVER_READ_HEADER_TIMEOUT=4s"}},
		{name: "write-timeout", want: "12s", env: []string{"SERVER_WRITE_TIMEOUT=12s"}},
		{name: "idle-timeout", want: "13s", env: []string{"SERVER_IDLE_TIMEOUT=13s"}},
		{name: "shutdown-timeout", want: "graceful shutdown within 2s", env: []string{"SERVER_SHUTDOWN_TIMEOUT=2s"}},
		{name: "max-header-bytes", want: "max=2048 oversized-header-status=431", env: []string{"SERVER_MAX_HEADER_BYTES=2048"}},
		{name: "max-multipart-memory", want: "multipart memory=4096 file-on-disk=true", env: []string{"SERVER_MAX_MULTIPART_MEMORY=4096"}},
		{name: "duration-string", want: "read=1m30s header=4s write=12s idle=13s", env: []string{"SERVER_READ_TIMEOUT=1m30s", "SERVER_READ_HEADER_TIMEOUT=4s", "SERVER_WRITE_TIMEOUT=12s", "SERVER_IDLE_TIMEOUT=13s"}},
		{name: "duration-seconds", want: "9s", env: []string{"SERVER_READ_TIMEOUT=9"}},
		{name: "trusted-proxies", want: "body=203.0.113.10", env: []string{"SERVER_TRUSTED_PROXIES=127.0.0.1"}},
		{name: "client-ip-headers", want: "body=198.51.100.9", env: []string{"SERVER_TRUSTED_PROXIES=127.0.0.1", "SERVER_CLIENT_IP_HEADERS=X-Real-IP"}},
		{name: "untrusted-client-ip", want: "body=192.0.2.7", env: []string{"SERVER_TRUSTED_PROXIES=127.0.0.1"}},
		{name: "access-log", want: "enabled=true logged=true"},
		{name: "exception-handler", want: "status=500", notWant: "internal demo detail", env: []string{"SERVER_EXCEPTION_HANDLER=true"}},
		{name: "debug-exception", want: "internal demo detail", env: []string{"SERVER_EXCEPTION_HANDLER=true", "APP_DEBUG=true"}},
		{name: "pid-file", want: "pid="},
		{name: "stop", want: "server shutdown completed"},
		{name: "kill", want: "server killed"},
		{name: "reload", want: "new server started"},
		{name: "restart", want: "new server started"},
		{name: "missing-pid", want: "read pid file failed"},
	}
	if len(cases) != 29 {
		t.Fatalf("HTTP server integration cases = %d, want 29", len(cases))
	}
	covered := make(map[string]bool, len(cases))
	for _, test := range cases {
		covered[test.name] = true
	}
	for _, entry := range catalog.Filter("http-server", "", catalog.StatusImplemented) {
		if !covered[entry.Case] {
			t.Fatalf("HTTP server catalog case %q has no local integration scenario, want coverage", entry.Case)
		}
	}
	runCase := func(t *testing.T, test integrationCase) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, "app"), 0o755); err != nil {
			t.Fatalf("prepare app directory error = %v, want nil", err)
		}
		if err := os.WriteFile(filepath.Join(root, ".env"), []byte("APP_ENV=testing\nAPP_KEY=base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=\n"), 0o600); err != nil {
			t.Fatalf("prepare app environment error = %v, want nil", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, executable, "demo:http-server", test.name, "--json")
		command.Dir = root
		command.Env = append(isolatedHTTPEnvironment(), test.env...)
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("demo:http-server %s error = %v, want nil; output = %s", test.name, err, output)
		}
		var result httpserver.Result
		if err := json.Unmarshal(output, &result); err != nil {
			t.Fatalf("decode demo:http-server %s output error = %v, want JSON; output = %s", test.name, err, output)
		}
		if result.Case != test.name || !strings.Contains(result.Value, test.want) || (test.notWant != "" && strings.Contains(result.Value, test.notWant)) {
			t.Fatalf("demo:http-server %s result = %#v, want case %q, value containing %q, and not containing %q", test.name, result, test.name, test.want, test.notWant)
		}
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) { runCase(t, test) })
	}
	t.Run("access-log-disabled", func(t *testing.T) {
		runCase(t, integrationCase{name: "access-log", want: "enabled=false logged=false", env: []string{"SERVER_ACCESS_LOG=false"}})
	})
	t.Run("exception-handler-disabled", func(t *testing.T) {
		runCase(t, integrationCase{name: "exception-handler", want: "status=200 body=", notWant: "internal demo detail", env: []string{"SERVER_EXCEPTION_HANDLER=false"}})
	})
}

func isolatedHTTPEnvironment() []string {
	environment := make([]string, 0, len(os.Environ())+8)
	for _, value := range os.Environ() {
		key, _, _ := strings.Cut(value, "=")
		if strings.HasPrefix(key, "SERVER_") || key == "APP_DEBUG" || key == "GIN_MODE" {
			continue
		}
		environment = append(environment, value)
	}
	return append(environment,
		"APP_DEBUG=false", "GIN_MODE=release", "DB_CONNECTION=sqlite",
		"CACHE_STORE=memory", "QUEUE_CONNECTION=sync", "SESSION_DRIVER=file",
	)
}
