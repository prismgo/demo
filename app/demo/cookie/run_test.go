package cookiedemo_test

import (
	"strings"
	"testing"

	"prismgo-demo/app/demo/catalog"
	cookiedemo "prismgo-demo/app/demo/cookie"
	demotest "prismgo-demo/app/demo/testing"
)

func TestCookieDemoFirstBatch(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	want := map[string]string{
		"value-object":        "name=theme value=dark minutes=10",
		"scope-deduplication": "headers=2 values=theme=last@/,theme=admin@/admin",
		"defaults":            "path=/ httpOnly=true secure=false sameSite=\"\" forever=2628000",
		"provider":            "lazy=true resolved=true queue=true",
		"middleware":          "status=201 cookies=theme=dark@/",
		"flush-failure":       "status=500 cookies=",
		"session-queue":       "",
		"make":                "equal=true name=theme value=dark",
		"forever":             "minutes=2628000 value=dark",
		"scope-options":       "path=/admin domain=example.test",
		"security-flags":      "httpOnly=false secure=true",
		"raw":                 "raw=true value=dark",
		"scope-option":        "names=one,two path=/admin domain=example.test",
		"minutes":             "minutes=5,0 maxAge=300,0 persistent=true,false",
		"expires-at":          "expires=2030-01-02T04:04:05Z maxAge=0",
		"max-age":             "expires=2030-01-02T03:04:05Z maxAge=42",
		"attach":              "first=one@/,second=two@/",
		"invalid-name":        "invalid=true headers=0",
		"attach-context":      "trace=value:request-42@/",
		"attach-now":          "expires=2030-01-02T03:09:05Z maxAge=300",
		"queue-make":          "status=201 cookies=theme=dark@/",
		"queue-forever":       "status=201 cookies=theme=dark@/ maxAge=157680000",
		"queue-cookie":        "status=201 cookies=theme=dark@/",
		"queued":              "found=true value=dark",
		"has-queued":          "found=true",
		"scoped-queued":       "found=true value=admin default=true:dark other-domain=false",
		"unqueue":             "remaining=false",
		"queue-from":          "status=201 cookies=theme=dark@/",
		"queue-not-found":     "missing=true",
		"process-queue":       "queued=true value=dark removed=true",
	}
	if len(want) != 30 {
		t.Fatalf("first cookie batch expectations = %d, want 30", len(want))
	}
	entries := catalog.Filter("cookie", "", catalog.StatusImplemented)
	if len(entries) != 50 {
		t.Fatalf("implemented cookie entries = %d, want 50", len(entries))
	}
	for _, entry := range entries {
		expected, ok := want[entry.Case]
		if !ok {
			continue
		}
		t.Run(entry.Case, func(t *testing.T) {
			result, err := cookiedemo.Run(entry.Case)
			if err != nil {
				t.Fatalf("Run(%q) error = %v, want nil", entry.Case, err)
			}
			if result.Case != entry.Case {
				t.Fatalf("Run(%q) case = %q, want %q", entry.Case, result.Case, entry.Case)
			}
			if entry.Case == "session-queue" {
				if !strings.HasPrefix(result.Value, "status=200 cookies=prismgo_session=") || !strings.HasSuffix(result.Value, "@/,theme=dark@/") {
					t.Fatalf("Run(%q) value = %q, want session ID cookie followed by queued theme cookie", entry.Case, result.Value)
				}
				return
			}
			if result.Value != expected {
				t.Fatalf("Run(%q) value = %q, want %q", entry.Case, result.Value, expected)
			}
		})
	}
}

func TestCookieDemoRejectsUnknownScenario(t *testing.T) {
	_, err := cookiedemo.Run("unknown")
	if err == nil || !strings.Contains(err.Error(), `unknown scenario "unknown"`) {
		t.Fatalf("Run(unknown) error = %v, want unknown scenario", err)
	}
}

func TestCookieDemoProviderPreservesCurrentApplication(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	if _, err := cookiedemo.Run("provider"); err != nil {
		t.Fatalf("Run(provider) error = %v, want nil", err)
	}
	result, err := cookiedemo.Run("process-queue")
	if err != nil {
		t.Fatalf("Run(process-queue) after provider error = %v, want nil", err)
	}
	if result.Value != "queued=true value=dark removed=true" {
		t.Fatalf("Run(process-queue) after provider = %q, want queued=true value=dark removed=true", result.Value)
	}
}
