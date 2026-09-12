package configdemo_test

import (
	configdemo "prismgo-demo/app/demo/config"
	"strings"
	"testing"
)

func assertConfigCase(t *testing.T, name, want string) {
	t.Helper()
	for key, value := range map[string]string{
		"MAIL_HOST": "", "MAIL_PORT": "", "MAIL_FROM_ADDRESS": "", "APP_NAME": "", "APP_ENV": "",
		"CONFIG_DEMO_PRIORITY": "", "CONFIG_DEMO_UNSET": "", "CONFIG_DEMO_BOOL": "", "CONFIG_DEMO_INT": "",
		"CONFIG_DEMO_INT64": "", "CONFIG_DEMO_UINT": "", "CONFIG_DEMO_FLOAT64": "", "CONFIG_DEMO_STRING": "",
	} {
		t.Setenv(key, value)
	}
	result, err := configdemo.Run(name)
	if err != nil {
		t.Fatalf("config case %q error = %v, want nil", name, err)
	}
	if result.Case != name || result.Value != want {
		t.Fatalf("config case %q result = %#v, want case %q and value %q", name, result, name, want)
	}
}

func TestConfigDemoQuickStart(t *testing.T) { assertConfigCase(t, "quick-start", "127.0.0.1:1025") }
func TestConfigDemoRuntimeScope(t *testing.T) {
	assertConfigCase(t, "runtime-scope", "first -> second")
}
func TestConfigDemoLazyLoading(t *testing.T) {
	assertConfigCase(t, "lazy-loading", "resolved before=false; after=true")
}
func TestConfigDemoRegistrationTiming(t *testing.T) {
	assertConfigCase(t, "registration-timing", "first-load -> second-load")
}
func TestConfigDemoMissingEnvFile(t *testing.T) { assertConfigCase(t, "missing-env-file", "127.0.0.1") }
func TestConfigDemoEnvPrecedence(t *testing.T)  { assertConfigCase(t, "env-precedence", "from-file") }
func TestConfigDemoEnv(t *testing.T)            { assertConfigCase(t, "env", "from-env") }
func TestConfigDemoEnvDefault(t *testing.T) {
	assertConfigCase(t, "env-default", "missing=fallback; blank=fallback")
}
func TestConfigDemoEnvBool(t *testing.T)    { assertConfigCase(t, "env-bool", "bool:true") }
func TestConfigDemoEnvInt(t *testing.T)     { assertConfigCase(t, "env-int", "int:42") }
func TestConfigDemoEnvInt64(t *testing.T)   { assertConfigCase(t, "env-int64", "int64:9000000001") }
func TestConfigDemoEnvUint(t *testing.T)    { assertConfigCase(t, "env-uint", "uint:17") }
func TestConfigDemoEnvFloat64(t *testing.T) { assertConfigCase(t, "env-float64", "float64:1.25") }
func TestConfigDemoEnvString(t *testing.T)  { assertConfigCase(t, "env-string", "string:hello") }
func TestConfigDemoAdd(t *testing.T)        { assertConfigCase(t, "add", "mail: 3 fields") }
func TestConfigDemoNestedPath(t *testing.T) {
	assertConfigCase(t, "nested-path", "noreply@example.com")
}
func TestConfigDemoMapNormalization(t *testing.T) {
	assertConfigCase(t, "map-normalization", "nested map[string]any=true")
}
func TestConfigDemoGetString(t *testing.T) {
	assertConfigCase(t, "get-string", "Get=127.0.0.1; GetString=127.0.0.1")
}
func TestConfigDemoGetInt(t *testing.T)       { assertConfigCase(t, "get-int", "1025") }
func TestConfigDemoGetInt64(t *testing.T)     { assertConfigCase(t, "get-int64", "9000000001") }
func TestConfigDemoGetUint(t *testing.T)      { assertConfigCase(t, "get-uint", "17") }
func TestConfigDemoGetFloat64(t *testing.T)   { assertConfigCase(t, "get-float64", "1.25") }
func TestConfigDemoGetBool(t *testing.T)      { assertConfigCase(t, "get-bool", "true") }
func TestConfigDemoGetStringMap(t *testing.T) { assertConfigCase(t, "get-string-map", "mail fields=3") }
func TestConfigDemoGetStringMapString(t *testing.T) {
	assertConfigCase(t, "get-string-map-string", "local")
}
func TestConfigDemoGetDefault(t *testing.T) { assertConfigCase(t, "get-default", "fallback/7/true") }
func TestConfigDemoMissingMap(t *testing.T) {
	assertConfigCase(t, "missing-map", "any=true:0; string=true:0")
}
func TestConfigDemoAppRegistration(t *testing.T) {
	assertConfigCase(t, "app-registration", "registered-app")
}
func TestConfigDemoAppName(t *testing.T) { assertConfigCase(t, "app-name", "Prismgo") }
func TestConfigDemoAppEnv(t *testing.T)  { assertConfigCase(t, "app-env", "production") }

func TestConfigDemoEnvironmentOverrides(t *testing.T) {
	for _, test := range []struct{ name, key, value, want string }{
		{name: "quick-start", key: "MAIL_HOST", value: "smtp.local", want: "smtp.local:1025"},
		{name: "app-name", key: "APP_NAME", value: "Example", want: "Example"},
		{name: "app-env", key: "APP_ENV", value: "testing", want: "testing"},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, key := range []string{"MAIL_HOST", "MAIL_PORT", "APP_NAME", "APP_ENV"} {
				t.Setenv(key, "")
			}
			t.Setenv(test.key, test.value)
			result, err := configdemo.Run(test.name)
			if err != nil || result.Value != test.want {
				t.Fatalf("config case %q with %s=%q result = %#v, error = %v, want %q", test.name, test.key, test.value, result, err, test.want)
			}
		})
	}
}

func TestConfigDemoUnknownScenario(t *testing.T) {
	_, err := configdemo.Run("unknown")
	if err == nil || !strings.Contains(err.Error(), `unknown scenario "unknown"`) {
		t.Fatalf("unknown config case error = %v, want descriptive unknown scenario error", err)
	}
}

func TestConfigDemoMissingEnvFileUsesProcessEnvironment(t *testing.T) {
	t.Setenv("MAIL_HOST", "process-host")
	result, err := configdemo.Run("missing-env-file")
	if err != nil || result.Value != "process-host" {
		t.Fatalf("missing .env result = %#v, error = %v, want process-host", result, err)
	}
}

func TestConfigDemoProcessEnvironmentOverridesFile(t *testing.T) {
	t.Setenv("CONFIG_DEMO_PRIORITY", "from-process")
	result, err := configdemo.Run("env-precedence")
	if err != nil || result.Value != "from-process" {
		t.Fatalf("environment precedence result = %#v, error = %v, want from-process", result, err)
	}
}
