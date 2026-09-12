package eventdemo_test

import (
	"context"
	"strings"
	"testing"

	"prismgo-demo/app/demo/catalog"
	eventdemo "prismgo-demo/app/demo/event"
	demotest "prismgo-demo/app/demo/testing"
)

func TestEventDemoFirstBatch(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	wantValues := map[string]string{
		"architecture":              "implements event.Dispatcher",
		"manual-registration":       "handled=1",
		"struct-registration":       "handled=1",
		"closure-listener":          "handled=1",
		"wildcard-prefix":           "handled=1; exact=false",
		"wildcard-all":              "handled=1; exact=false",
		"wildcard-exact-operations": "handled=1; exact=false",
		"event-definition":          "demo.event.created",
		"event-naming":              "demo.event.created,demo.event.updated",
		"payload-boundaries":        "serializable",
		"struct-listener":           "handled=1",
		"listener-func":             "handled=1",
		"propagation":               "handled=2",
		"queue-prerequisites":       "ShouldQueue=true",
		"should-queue":              "handled=1",
		"queued-wrapper":            "handled=1",
		"queued-sync":               "handled=1",
		"queue-routing-options":     "connection=sync; queue=demo-event; delay=0s; handled=1",
		"queue-retry-options":       "tries=3; backoff=[1s 2s]; timeout=5s; handled=1",
		"event-factory":             "registered:demo.event.registered",
		"event-factory-validation":  "rejected=2",
		"raw-queued-event":          "handled=1; event=demo.event.raw",
		"dispatch":                  "handled=2",
		"dispatch-isolation":        "handled=2",
		"nil-dispatch":              "nil-safe; handled=1",
		"subscriber":                "handled=2",
		"subscriber-registration":   "handled=2",
		"facade":                    "handled=1; before=true; after=false",
		"async":                     "handled=1; mode=async",
	}
	entries := catalog.Filter("event", "", catalog.StatusImplemented)
	if len(entries) != 57 {
		t.Fatalf("implemented event entries = %d, want 57", len(entries))
	}
	for _, entry := range entries {
		if entry.Case == "queued-redis" {
			continue
		}
		if _, ok := wantValues[entry.Case]; !ok {
			continue
		}
		t.Run(entry.Case, func(t *testing.T) {
			want, ok := wantValues[entry.Case]
			if !ok {
				t.Fatalf("case %q missing expected value", entry.Case)
			}
			result, err := eventdemo.Run(context.Background(), entry.Case, "sync")
			if err != nil {
				t.Fatalf("Run(%q) error = %v, want nil", entry.Case, err)
			}
			if result.Case != entry.Case || !strings.Contains(result.Value, want) {
				t.Fatalf("Run(%q) = %#v, want value containing %q", entry.Case, result, want)
			}
		})
	}
}

func TestEventDemoRejectsInvalidScenarioInputs(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	for _, testCase := range []struct{ name, connection, want string }{
		{name: "missing-case", connection: "sync", want: "unknown scenario"},
		{name: "dispatch", connection: "unknown", want: "unsupported connection"},
		{name: "queued-redis", connection: "sync", want: "requires --connection=redis"},
	} {
		t.Run(testCase.name+"/"+testCase.connection, func(t *testing.T) {
			_, err := eventdemo.Run(context.Background(), testCase.name, testCase.connection)
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("Run(%q,%q) error = %v, want %q", testCase.name, testCase.connection, err, testCase.want)
			}
		})
	}
}
