package cookiedemo_test

import (
	"testing"

	"prismgo-demo/app/demo/catalog"
	cookiedemo "prismgo-demo/app/demo/cookie"
	demotest "prismgo-demo/app/demo/testing"
)

func TestCookieDemoSecondBatch(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	want := map[string]string{
		"request-cookie":        "value=dark",
		"request-not-found":     "missing=true",
		"request-context":       "value=value:request-42",
		"request-signer":        "value=dark",
		"request-encryptor":     "value=dark",
		"queue-expire":          "name=theme path=/admin domain=example.test maxAge=-1 value=\"\"",
		"queue-forget":          "name=theme path=/admin domain=example.test maxAge=-1 value=\"\"",
		"forget":                "name=theme path=/admin domain=example.test maxAge=-1 value=\"\"",
		"deletion-scope":        "sameScope=true deleted=true",
		"same-site-default":     "default=true disabled=true",
		"same-site-modes":       "modes=Lax,Strict,None",
		"security-contracts":    "signer=true encryptor=true",
		"outgoing-order":        "value=theme.enc-dark steps=encrypt,sign",
		"incoming-order":        "value=dark steps=verify,decrypt",
		"passthrough":           "value=dark",
		"attach-security":       "value=theme.enc-dark steps=encrypt,sign",
		"queue-security":        "value=theme.enc-dark steps=encrypt,sign cleared=true",
		"sensitive-error":       "redacted=true signature=true typed=true",
		"errors":                "invalid=true missing=true queue=true signature=true encryption=true decryption=true",
		"laravel-compatibility": "make=dark forever=2628000 read=dark forget=-1",
	}
	if len(want) != 20 {
		t.Fatalf("second cookie batch expectations = %d, want 20", len(want))
	}
	entries := catalog.Filter("cookie", "", catalog.StatusImplemented)
	if len(entries) != 50 {
		t.Fatalf("implemented cookie entries = %d, want 50", len(entries))
	}
	covered := 0
	for _, entry := range entries {
		expected, ok := want[entry.Case]
		if !ok {
			continue
		}
		covered++
		t.Run(entry.Case, func(t *testing.T) {
			result, err := cookiedemo.Run(entry.Case)
			if err != nil {
				t.Fatalf("Run(%q) error = %v, want nil", entry.Case, err)
			}
			if result.Case != entry.Case || result.Value != expected {
				t.Fatalf("Run(%q) = %#v, want case=%q value=%q", entry.Case, result, entry.Case, expected)
			}
		})
	}
	if covered != len(want) {
		t.Fatalf("second cookie batch covered = %d, want %d", covered, len(want))
	}
}
