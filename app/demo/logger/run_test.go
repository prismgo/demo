package loggerdemo_test

import (
	"path/filepath"
	"strings"
	"testing"

	"prismgo-demo/app/demo/catalog"
	loggerdemo "prismgo-demo/app/demo/logger"
	demotest "prismgo-demo/app/demo/testing"
)

func TestLoggerDemoCases(t *testing.T) {
	expected := map[string]string{
		"architecture":              "manager=*logger.Manager; channel=*logger.channel",
		"config":                    "default=stack; single=single; path=",
		"single-driver":             "created=true",
		"daily-driver":              "rotated=true",
		"stderr-driver":             "stderr=true",
		"null-driver":               "suppressed=true",
		"top-level-config":          "default=sink",
		"channel-config":            "driver=single; formatter=json; level=warn; path=storage/logs/app.log; children=1",
		"deployment-config":         "single=debug; error=error; stderr=warn",
		"stack":                     "first=true; second=true",
		"stack-channel-isolation":   "first_info=true; second_info=false; second_json=true",
		"levels":                    "info=false; warn=true",
		"facade":                    "recorded=true",
		"formatted-facade":          "recorded=true",
		"fatal":                     "logged=true; exited=true",
		"with-field":                "field=true",
		"with-fields":               "order=true; tenant=true",
		"with-error":                "error=true",
		"error-stacktrace":          "error=true; stack=true",
		"context-extractor":         "manager=true; override=true",
		"context-noop":              "request_absent=true",
		"named-channel":             "first=true; second=false",
		"missing-channel":           "fallback=true",
		"channel-context":           "tenant=true; isolated=true",
		"resolve-manager":           "manager=*logger.Manager",
		"default-name":              "default=demo",
		"driver-contract":           "Driver=Write+Close",
		"custom-driver":             "written=true",
		"custom-driver-replacement": "old=false; replacement=true",
		"custom-formatter":          "output=true",
		"formatter-params":          "channel=demo; prefix=demo:",
		"line-formatter":            "line=true",
		"text-formatter":            "text=true",
		"json-formatter":            "level=info; tenant=blue",
		"manual-manager":            "default=demo; written=true",
		"provider-lifecycle":        "bound=true; lazy=true; singleton=true",
		"provider-cleanup":          "cleanup=true",
		"resource-cleanup":          "closed=true; written=true",
		"closed-writes":             "unchanged=true",
		"global-logrus":             "recorded=true",
	}
	if len(expected) != 40 {
		t.Fatalf("logger case count = %d, want 40", len(expected))
	}
	entries := catalog.Filter("logger", "", catalog.StatusImplemented)
	if len(entries) != len(expected) {
		t.Fatalf("implemented logger entries = %d, want %d", len(entries), len(expected))
	}
	for _, entry := range entries {
		t.Run(entry.Case, func(t *testing.T) {
			if entry.Test != "TestLoggerDemoCases/"+entry.Case {
				t.Fatalf("case %q test target = %q, want TestLoggerDemoCases/%s", entry.Case, entry.Test, entry.Case)
			}
			want, ok := expected[entry.Case]
			if !ok {
				t.Fatalf("case %q has no expectation; want all catalog cases covered", entry.Case)
			}
			path := filepath.Join(t.TempDir(), "app.log")
			t.Setenv("APP_LOGGER_FILE", path)
			t.Setenv("ERROR_LOGGER_DRIVER", "null")
			if entry.Case == "deployment-config" {
				t.Setenv("APP_LOGGER_LEVEL", "debug")
				t.Setenv("ERROR_LOGGER_LEVEL", "error")
				t.Setenv("STDERR_LOGGER_LEVEL", "warn")
			}
			_ = demotest.NewApplication(t, demotest.Options{})
			result, err := loggerdemo.Run(entry.Case)
			if err != nil {
				t.Fatalf("case %q error = %v, want nil", entry.Case, err)
			}
			if result.Case != entry.Case || !strings.Contains(result.Value, want) {
				t.Fatalf("case %q result = %#v, want value containing %q", entry.Case, result, want)
			}
			if entry.Case == "config" && !strings.Contains(result.Value, "path="+path) {
				t.Fatalf("config result = %#v, want environment path %q", result, path)
			}
		})
	}
}
func TestLoggerDemoUnknownCase(t *testing.T) {
	_, err := loggerdemo.Run("missing")
	if err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("unknown case error = %v, want named error", err)
	}
}
