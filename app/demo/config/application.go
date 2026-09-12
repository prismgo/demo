package configdemo

import (
	"fmt"
	"strings"
)

type applicationSetting struct {
	env   string
	path  string
	value string
	kind  string
}

func settingFor(name string) (applicationSetting, bool) {
	switch name {
	case "app-key":
		return applicationSetting{env: "APP_KEY", path: "app.key", value: "demo-only-key", kind: "secret"}, true
	case "app-debug":
		return applicationSetting{env: "APP_DEBUG", path: "app.debug", value: "true", kind: "bool"}, true
	case "app-url":
		return applicationSetting{env: "APP_URL", path: "app.url", value: "https://example.test", kind: "string"}, true
	case "app-timezone":
		return applicationSetting{env: "APP_TIMEZONE", path: "app.timezone", value: "Asia/Shanghai", kind: "string"}, true
	case "app-locale":
		return applicationSetting{env: "APP_LOCALE", path: "app.locale", value: "zh_CN", kind: "string"}, true
	case "app-fallback-locale":
		return applicationSetting{env: "APP_FALLBACK_LOCALE", path: "app.fallback_locale", value: "fr", kind: "string"}, true
	case "app-cipher":
		return applicationSetting{env: "APP_CIPHER", path: "app.cipher", value: "AES-256-GCM", kind: "string"}, true
	case "app-previous-keys":
		return applicationSetting{env: "APP_PREVIOUS_KEYS", path: "app.previous_keys", value: "old-one,old-two", kind: "key-count"}, true
	case "server-host":
		return applicationSetting{env: "SERVER_HOST", path: "app.server.host", value: "127.0.0.1", kind: "string"}, true
	case "server-port":
		return applicationSetting{env: "SERVER_PORT", path: "app.server.port", value: "9090", kind: "int"}, true
	case "server-timeout":
		return applicationSetting{env: "SERVER_TIMEOUT", path: "app.server.timeout", value: "19", kind: "int"}, true
	case "server-read-timeout":
		return applicationSetting{env: "SERVER_READ_TIMEOUT", path: "app.server.read_timeout", value: "21s", kind: "string"}, true
	case "server-read-header-timeout":
		return applicationSetting{env: "SERVER_READ_HEADER_TIMEOUT", path: "app.server.read_header_timeout", value: "6s", kind: "string"}, true
	case "server-write-timeout":
		return applicationSetting{env: "SERVER_WRITE_TIMEOUT", path: "app.server.write_timeout", value: "32s", kind: "string"}, true
	case "server-idle-timeout":
		return applicationSetting{env: "SERVER_IDLE_TIMEOUT", path: "app.server.idle_timeout", value: "62s", kind: "string"}, true
	case "server-shutdown-timeout":
		return applicationSetting{env: "SERVER_SHUTDOWN_TIMEOUT", path: "app.server.shutdown_timeout", value: "17s", kind: "string"}, true
	case "server-max-header-bytes":
		return applicationSetting{env: "SERVER_MAX_HEADER_BYTES", path: "app.server.max_header_bytes", value: "2097152", kind: "int"}, true
	case "server-max-multipart-memory":
		return applicationSetting{env: "SERVER_MAX_MULTIPART_MEMORY", path: "app.server.max_multipart_memory", value: "16777216", kind: "int"}, true
	case "server-trusted-proxies":
		return applicationSetting{env: "SERVER_TRUSTED_PROXIES", path: "app.server.trusted_proxies", value: "10.0.0.0/8", kind: "string"}, true
	case "server-client-ip-headers":
		return applicationSetting{env: "SERVER_CLIENT_IP_HEADERS", path: "app.server.client_ip_headers", value: "X-Real-IP", kind: "string"}, true
	case "server-access-log":
		return applicationSetting{env: "SERVER_ACCESS_LOG", path: "app.server.access_log", value: "false", kind: "bool"}, true
	case "server-exception-handler":
		return applicationSetting{env: "SERVER_EXCEPTION_HANDLER", path: "app.server.exception_handler", value: "false", kind: "bool"}, true
	default:
		return applicationSetting{}, false
	}
}

func runApplicationSetting(name string) (string, error) {
	setting, ok := settingFor(name)
	if !ok {
		return "", fmt.Errorf("unknown application setting %q", name)
	}
	cfg, err := repositoryFromContent(setting.env + "=" + setting.value + "\n")
	if err != nil {
		return "", err
	}
	switch setting.kind {
	case "secret":
		return fmt.Sprintf("set=%t", cfg.GetString(setting.path) != ""), nil
	case "key-count":
		value := cfg.GetString(setting.path)
		if value == "" {
			return "keys=0", nil
		}
		return fmt.Sprintf("keys=%d", len(strings.Split(value, ","))), nil
	case "bool":
		return fmt.Sprintf("%t", cfg.GetBool(setting.path)), nil
	case "int":
		return fmt.Sprintf("%d", cfg.GetInt(setting.path)), nil
	default:
		return cfg.GetString(setting.path), nil
	}
}
