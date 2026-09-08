package demotest

import (
	"os"
	"strings"
	"testing"
)

// Service identifies an optional real integration dependency.
type Service string

const (
	ServiceMySQL     Service = "mysql"
	ServiceSQLite    Service = "sqlite"
	ServiceRedisAddr Service = "redis-addr"
	ServiceRedisURL  Service = "redis-url"
	ServiceRabbitMQ  Service = "rabbitmq"
)

var serviceEnvironment = map[Service]string{
	ServiceMySQL:     "PRISMGO_MYSQL_TEST_DSN",
	ServiceSQLite:    "PRISMGO_SQLITE_TEST_DSN",
	ServiceRedisAddr: "PRISMGO_REDIS_TEST_ADDR",
	ServiceRedisURL:  "PRISMGO_REDIS_TEST_URL",
	ServiceRabbitMQ:  "PRISMGO_RABBITMQ_TEST_URL",
}

// LookupIntegration returns the configured value and its canonical variable.
func LookupIntegration(service Service) (value, variable string, ok bool) {
	variable, ok = serviceEnvironment[service]
	if !ok {
		return "", "", false
	}
	value = strings.TrimSpace(os.Getenv(variable))
	return value, variable, value != ""
}

// RequireIntegration skips the current test when its real service is absent.
func RequireIntegration(t testing.TB, service Service) string {
	t.Helper()
	value, variable, ok := LookupIntegration(service)
	if variable == "" {
		t.Fatalf("unknown integration service %q", service)
	}
	if !ok {
		t.Skipf("%s is not configured", variable)
	}
	return value
}
