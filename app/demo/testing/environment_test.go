package demotest

import (
	"path/filepath"
	"testing"

	"github.com/prismgo/framework/config"
)

func TestLookupIntegration(t *testing.T) {
	t.Setenv("PRISMGO_REDIS_TEST_ADDR", " 127.0.0.1:16379 ")
	value, variable, ok := LookupIntegration(ServiceRedisAddr)
	if !ok || value != "127.0.0.1:16379" || variable != "PRISMGO_REDIS_TEST_ADDR" {
		t.Fatalf("LookupIntegration() = %q, %q, %v", value, variable, ok)
	}
	if _, variable, ok := LookupIntegration(Service("unknown")); ok || variable != "" {
		t.Fatalf("unknown service = %q, %v", variable, ok)
	}
}

func TestNewApplicationUsesHermeticDefaults(t *testing.T) {
	app := NewApplication(t, Options{})
	if got := config.GetString("app.env"); got != "testing" {
		t.Fatalf("app.env = %q, want testing", got)
	}
	if got := config.GetString("database.default"); got != "sqlite" {
		t.Fatalf("database.default = %q, want sqlite", got)
	}
	databasePath := config.GetString("database.connections.sqlite.database")
	if filepath.Dir(databasePath) != app.StoragePath() {
		t.Fatalf("sqlite path = %q, want file under %q", databasePath, app.StoragePath())
	}
}

func TestValidateHermeticOptions(t *testing.T) {
	if err := validateHermeticOptions(withDefaults(Options{})); err != nil {
		t.Fatalf("validate default hermetic options: %v", err)
	}
	if err := validateHermeticOptions(Options{Database: "mysql", Cache: "memory", Queue: "sync", Session: "file"}); err == nil {
		t.Fatal("mysql unexpectedly accepted as hermetic")
	}
}
