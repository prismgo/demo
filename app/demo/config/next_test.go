package configdemo_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/prismgo/framework/config"
	"github.com/prismgo/framework/container"
)

func TestConfigDemoApplicationDefaults(t *testing.T) {
	tests := []struct {
		name string
		env  string
		path string
		kind string
		want string
	}{
		{name: "key", env: "APP_KEY", path: "app.key", want: ""},
		{name: "debug", env: "APP_DEBUG", path: "app.debug", kind: "bool", want: "false"},
		{name: "url", env: "APP_URL", path: "app.url", want: "http://localhost:8080"},
		{name: "timezone", env: "APP_TIMEZONE", path: "app.timezone", want: "UTC"},
		{name: "locale", env: "APP_LOCALE", path: "app.locale", want: "en"},
		{name: "fallback locale", env: "APP_FALLBACK_LOCALE", path: "app.fallback_locale", want: "en"},
		{name: "cipher", env: "APP_CIPHER", path: "app.cipher", want: "AES-256-GCM"},
		{name: "previous keys", env: "APP_PREVIOUS_KEYS", path: "app.previous_keys", want: ""},
		{name: "host", env: "SERVER_HOST", path: "app.server.host", want: ""},
		{name: "port", env: "SERVER_PORT", path: "app.server.port", kind: "int", want: "8080"},
		{name: "timeout", env: "SERVER_TIMEOUT", path: "app.server.timeout", kind: "int", want: "15"},
		{name: "read timeout", env: "SERVER_READ_TIMEOUT", path: "app.server.read_timeout", want: "15s"},
		{name: "read header timeout", env: "SERVER_READ_HEADER_TIMEOUT", path: "app.server.read_header_timeout", want: "5s"},
		{name: "write timeout", env: "SERVER_WRITE_TIMEOUT", path: "app.server.write_timeout", want: "30s"},
		{name: "idle timeout", env: "SERVER_IDLE_TIMEOUT", path: "app.server.idle_timeout", want: "60s"},
		{name: "shutdown timeout", env: "SERVER_SHUTDOWN_TIMEOUT", path: "app.server.shutdown_timeout", want: "15s"},
		{name: "max header bytes", env: "SERVER_MAX_HEADER_BYTES", path: "app.server.max_header_bytes", kind: "int", want: "1048576"},
		{name: "max multipart memory", env: "SERVER_MAX_MULTIPART_MEMORY", path: "app.server.max_multipart_memory", kind: "int", want: "33554432"},
		{name: "trusted proxies", env: "SERVER_TRUSTED_PROXIES", path: "app.server.trusted_proxies", want: ""},
		{name: "client IP headers", env: "SERVER_CLIENT_IP_HEADERS", path: "app.server.client_ip_headers", want: "X-Forwarded-For,X-Real-IP"},
		{name: "access log", env: "SERVER_ACCESS_LOG", path: "app.server.access_log", kind: "bool", want: "true"},
		{name: "exception handler", env: "SERVER_EXCEPTION_HANDLER", path: "app.server.exception_handler", kind: "bool", want: "true"},
	}
	for _, test := range tests {
		t.Setenv(test.env, "")
	}
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatalf("write empty config file %s error = %v, want nil", path, err)
	}
	cfg, err := config.NewFromFile(path)
	if err != nil {
		t.Fatalf("load empty config file %s error = %v, want nil", path, err)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got string
			switch test.kind {
			case "bool":
				got = fmt.Sprint(cfg.GetBool(test.path))
			case "int":
				got = fmt.Sprint(cfg.GetInt(test.path))
			default:
				got = cfg.GetString(test.path)
			}
			if got != test.want {
				t.Fatalf("default %s (%s) = %q, want %q", test.name, test.path, got, test.want)
			}
		})
	}
}

func TestConfigDemoApplicationProcessOverrides(t *testing.T) {
	for _, test := range []struct {
		name  string
		key   string
		value string
		want  string
	}{
		{name: "server-port", key: "SERVER_PORT", value: "7070", want: "7070"},
		{name: "server-access-log", key: "SERVER_ACCESS_LOG", value: "true", want: "true"},
		{name: "app-locale", key: "APP_LOCALE", value: "de", want: "de"},
		{name: "app-previous-keys", key: "APP_PREVIOUS_KEYS", value: "one", want: "keys=1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(test.key, test.value)
			assertConfigCase(t, test.name, test.want)
		})
	}
}

func assertNewSetting(t *testing.T, name, key, want string) {
	t.Helper()
	t.Setenv(key, "")
	assertConfigCase(t, name, want)
}

func bindFacadeConfig(t *testing.T, cfg *config.Config) {
	t.Helper()
	registry := container.NewContainer()
	if err := registry.Instance("config.default", cfg); err != nil {
		t.Fatalf("bind config.default error = %v, want nil", err)
	}
	container.SetProvider(func() *container.Container { return registry })
	t.Cleanup(func() { container.SetProvider(nil) })
}

func facadeRepository(t *testing.T) *config.Config {
	t.Helper()
	t.Setenv("APP_NAME", "")
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("APP_NAME=facade-demo\n"), 0600); err != nil {
		t.Fatalf("write facade env file %s error = %v, want nil", path, err)
	}
	cfg, err := config.NewFromFile(path)
	if err != nil {
		t.Fatalf("load facade env file %s error = %v, want nil", path, err)
	}
	return cfg
}

func TestConfigDemoAppKey(t *testing.T)   { assertNewSetting(t, "app-key", "APP_KEY", "set=true") }
func TestConfigDemoAppDebug(t *testing.T) { assertNewSetting(t, "app-debug", "APP_DEBUG", "true") }
func TestConfigDemoAppURL(t *testing.T) {
	assertNewSetting(t, "app-url", "APP_URL", "https://example.test")
}
func TestConfigDemoAppTimezone(t *testing.T) {
	assertNewSetting(t, "app-timezone", "APP_TIMEZONE", "Asia/Shanghai")
}
func TestConfigDemoAppLocale(t *testing.T) { assertNewSetting(t, "app-locale", "APP_LOCALE", "zh_CN") }
func TestConfigDemoAppFallbackLocale(t *testing.T) {
	assertNewSetting(t, "app-fallback-locale", "APP_FALLBACK_LOCALE", "fr")
}
func TestConfigDemoAppCipher(t *testing.T) {
	assertNewSetting(t, "app-cipher", "APP_CIPHER", "AES-256-GCM")
}
func TestConfigDemoAppPreviousKeys(t *testing.T) {
	assertNewSetting(t, "app-previous-keys", "APP_PREVIOUS_KEYS", "keys=2")
}
func TestConfigDemoServerHost(t *testing.T) {
	assertNewSetting(t, "server-host", "SERVER_HOST", "127.0.0.1")
}
func TestConfigDemoServerPort(t *testing.T) {
	assertNewSetting(t, "server-port", "SERVER_PORT", "9090")
}
func TestConfigDemoServerTimeout(t *testing.T) {
	assertNewSetting(t, "server-timeout", "SERVER_TIMEOUT", "19")
}
func TestConfigDemoServerReadTimeout(t *testing.T) {
	assertNewSetting(t, "server-read-timeout", "SERVER_READ_TIMEOUT", "21s")
}
func TestConfigDemoServerReadHeaderTimeout(t *testing.T) {
	assertNewSetting(t, "server-read-header-timeout", "SERVER_READ_HEADER_TIMEOUT", "6s")
}
func TestConfigDemoServerWriteTimeout(t *testing.T) {
	assertNewSetting(t, "server-write-timeout", "SERVER_WRITE_TIMEOUT", "32s")
}
func TestConfigDemoServerIdleTimeout(t *testing.T) {
	assertNewSetting(t, "server-idle-timeout", "SERVER_IDLE_TIMEOUT", "62s")
}
func TestConfigDemoServerShutdownTimeout(t *testing.T) {
	assertNewSetting(t, "server-shutdown-timeout", "SERVER_SHUTDOWN_TIMEOUT", "17s")
}
func TestConfigDemoServerMaxHeaderBytes(t *testing.T) {
	assertNewSetting(t, "server-max-header-bytes", "SERVER_MAX_HEADER_BYTES", "2097152")
}
func TestConfigDemoServerMaxMultipartMemory(t *testing.T) {
	assertNewSetting(t, "server-max-multipart-memory", "SERVER_MAX_MULTIPART_MEMORY", "16777216")
}
func TestConfigDemoServerTrustedProxies(t *testing.T) {
	assertNewSetting(t, "server-trusted-proxies", "SERVER_TRUSTED_PROXIES", "10.0.0.0/8")
}
func TestConfigDemoServerClientIPHeaders(t *testing.T) {
	assertNewSetting(t, "server-client-ip-headers", "SERVER_CLIENT_IP_HEADERS", "X-Real-IP")
}
func TestConfigDemoServerAccessLog(t *testing.T) {
	assertNewSetting(t, "server-access-log", "SERVER_ACCESS_LOG", "false")
}
func TestConfigDemoServerExceptionHandler(t *testing.T) {
	assertNewSetting(t, "server-exception-handler", "SERVER_EXCEPTION_HANDLER", "false")
}
func TestConfigDemoNewFromFile(t *testing.T) {
	assertConfigCase(t, "new-from-file", "standalone")
}
func TestConfigDemoStandaloneRead(t *testing.T) {
	t.Setenv("APP_DEBUG", "")
	t.Setenv("SERVER_PORT", "")
	assertConfigCase(t, "standalone-read", "debug=true; port=9090")
}
func TestConfigDemoCloneIsolation(t *testing.T) {
	assertConfigCase(t, "clone-isolation", "before -> after")
}
func TestConfigDemoResolve(t *testing.T) {
	bindFacadeConfig(t, facadeRepository(t))
	assertConfigCase(t, "resolve", "facade-demo")
}
func TestConfigDemoFacadeClone(t *testing.T) {
	bindFacadeConfig(t, facadeRepository(t))
	assertConfigCase(t, "facade-clone", "isolated=true; same-name=true")
}
func TestConfigDemoReload(t *testing.T) {
	bindFacadeConfig(t, config.New())
	assertConfigCase(t, "reload", "same-instance=true; empty=false")
}
func TestConfigDemoEmpty(t *testing.T) {
	bindFacadeConfig(t, config.New())
	assertConfigCase(t, "empty", "empty=true")
}
func TestConfigDemoTestBinding(t *testing.T) {
	assertConfigCase(t, "test-binding", "same-instance=true; name=bound-test")
}
func TestConfigDemoConventions(t *testing.T) {
	assertConfigCase(t, "conventions", "app.url <- APP_URL")
}
func TestConfigDemoMutableSettings(t *testing.T) {
	assertConfigCase(t, "mutable-settings", "runtime=static; mutable=second")
}
