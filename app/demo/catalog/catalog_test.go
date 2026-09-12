package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogEntries(t *testing.T) {
	entries := All()
	if err := Validate(entries); err != nil {
		t.Fatalf("validate catalog entries: %v", err)
	}
	if len(entries) != 476 {
		t.Fatalf("catalog has %d entries, want 476", len(entries))
	}
	if item, ok := Find("commands", "list"); !ok || item.Status != StatusImplemented {
		t.Fatalf("implemented demo:list entry = %#v, %v", item, ok)
	}
	if item, ok := Find("queue", "basic-sync"); !ok || item.Since != SinceInitial {
		t.Fatalf("implemented queue version = %#v, %v; want %s", item, ok, SinceInitial)
	}
	if item, ok := Find("queue", "batch-events"); !ok || item.Since != SinceInitial {
		t.Fatalf("planned demo framework version = %#v, %v; want %s", item, ok, SinceInitial)
	}
	if item, ok := Find("horizon", "list"); !ok || item.Status != StatusImplemented || item.Example != "./demo/dev test-horizon" {
		t.Fatalf("Horizon live queue entry = %#v, %v; want implemented ./demo/dev test-horizon", item, ok)
	}
	for _, slug := range []string{
		"driver-prerequisites", "config", "payload-encoding", "sync-connection",
		"failed-store", "batch-store", "restart-store",
		"debounce-options", "job-control", "worker-command", "expiration",
		"failed-command-paths", "failed-retry", "failed-event", "batch-events", "poison-event", "infrastructure-events",
		"encryption-missing-key", "custom-queue-contract", "custom-reserved-job", "custom-pop-session", "custom-consumer-intent",
		"job-errors", "connection-errors", "poison-errors", "rabbitmq-errors", "bulk",
		"transport-delay", "blocking-pop", "redis-retry-after", "rabbitmq-retry-after", "rabbitmq-confirm",
		"rabbitmq-reconnect", "rabbitmq-topology", "rabbitmq-delay-modes", "poison-rejection",
	} {
		if item, ok := Find("queue", slug); !ok || item.Status != StatusImplemented {
			t.Fatalf("queue configuration entry %q = %#v, %v; want implemented", slug, item, ok)
		}
	}
	for _, slug := range []string{
		"architecture", "config", "provider", "short-keys", "nested-keys", "json-keys", "key-conflicts", "namespaces",
		"translator", "facade", "missing-default", "locale-argument", "has", "has-for-locale", "get-map",
		"replacements", "replacement-case", "stringable", "plural-pipe", "plural-intervals",
		"plural-replacements", "plural-count",
		"locale", "locale-validation", "fallback", "fallback-validation", "locale-resolver", "namespace-overrides",
		"add-lines", "add-lines-precedence",
		"missing-handler", "group-paths", "json-paths", "custom-loader",
		"translator-contract", "loader-contract", "selector-contract", "reset", "isolated",
	} {
		if item, ok := Find("translation", slug); !ok || item.Status != StatusImplemented {
			t.Fatalf("translation catalog entry %q = %#v, %v; want implemented", slug, item, ok)
		}
	}
	for _, slug := range []string{
		"architecture", "config", "driver-prerequisites", "top-level-config",
		"memory-config", "redis-config", "file-config", "failover-config", "lock-config", "flexible-config",
		"facade", "named-store", "repository", "missing-store",
		"get", "fallbacks", "typed-retrieval", "existence",
		"put", "forever", "add", "put-many", "remember", "remember-forever", "flexible", "touch", "many", "pull",
		"forget", "forget-many", "flush", "counters",
		"lock", "lock-callback", "lock-block", "lock-restore", "lock-flush", "funnel", "without-overlapping",
		"tags-memory", "tags-redis", "tags-unsupported", "memo", "failover", "custom-driver", "resource-lifecycle",
		"events", "event-contract", "deferred", "key-prefixes", "encoding", "errors",
		"memory-capabilities", "file-capabilities", "redis-capabilities", "failover-capabilities", "laravel-compatibility",
	} {
		if item, ok := Find("cache", slug); !ok || item.Status != StatusPlanned {
			t.Fatalf("cache catalog entry %q = %#v, %v; want planned", slug, item, ok)
		}
	}
	for _, slug := range []string{
		"architecture", "manual-registration", "struct-registration", "closure-listener",
		"wildcard-prefix", "wildcard-all", "wildcard-exact-operations",
		"event-definition", "event-naming", "payload-boundaries",
		"struct-listener", "listener-func", "propagation",
		"queue-prerequisites", "should-queue", "queued-wrapper", "queued-sync", "queued-redis",
		"queue-routing-options", "queue-retry-options", "event-factory", "event-factory-validation", "raw-queued-event",
		"dispatch", "dispatch-isolation", "nil-dispatch",
		"subscriber", "subscriber-registration", "facade", "async", "async-durability",
		"app-lifecycle", "provider-lifecycle", "server-lifecycle", "request-lifecycle", "request-finished-ordering",
		"console-lifecycle", "vendor-publish-event",
		"listener-error-isolation", "listener-panic-isolation", "async-failure-isolation", "queued-failure-handling", "consistency-boundary",
		"isolated-testing", "queued-sync-testing", "queued-worker-testing",
		"event-interface", "listener-interface", "listener-func-interface", "dispatcher-interface",
		"subscriber-interface", "should-queue-interface", "async-listener-interface", "queue-options-interface",
		"provider-register", "provider-boot", "laravel-compatibility",
	} {
		if item, ok := Find("event", slug); !ok || item.Status != StatusPlanned {
			t.Fatalf("event catalog entry %q = %#v, %v; want planned", slug, item, ok)
		}
	}
	for _, slug := range []string{
		"architecture", "config", "store-resolution", "memory-store", "redis-store", "config-registration",
		"auto-initialization", "quick-start", "explicit-limiter",
		"named-registration", "named-lookup", "unregistered-pass-through", "multi-rule-counters", "multi-rule-block",
		"limit-contract", "every", "per-second", "per-minute", "per-minutes", "per-hour", "per-day", "none",
		"by", "fallback-key", "after-count", "after-skip", "after-headers", "custom-response", "result-contract",
		"throttle", "throttle-for", "disabled-rule", "default-response", "success-headers", "over-limit-headers", "tightest-headers", "route-group",
		"route-compatibility", "facade", "instance",
		"hit", "hit-default-decay", "increment", "increment-default", "decrement", "decrement-default",
		"attempts", "attempts-missing", "too-many-attempts", "non-positive-limit", "expired-window",
		"remaining", "remaining-floor", "available-in", "available-in-elapsed", "reset-attempts", "clear",
		"attempt-success", "attempt-blocked", "attempt-error", "attempt-nil", "clean-key",
		"hashed-key", "unhashed-key", "manual-key", "key-design",
		"cache-layout", "fixed-window", "middleware-key",
		"cache-errors", "redis-errors", "counter-type-error", "error-policy", "laravel-compatibility",
	} {
		if item, ok := Find("ratelimit", slug); !ok || item.Status != StatusPlanned {
			t.Fatalf("ratelimit catalog entry %q = %#v, %v; want planned", slug, item, ok)
		}
	}
	for _, slug := range []string{
		"architecture", "config", "single-driver", "daily-driver", "stderr-driver", "null-driver",
		"top-level-config", "channel-config", "deployment-config", "stack", "stack-channel-isolation", "levels",
		"facade", "formatted-facade", "fatal", "with-field", "with-fields", "with-error", "error-stacktrace",
		"context-extractor", "context-noop", "named-channel", "missing-channel", "channel-context",
		"resolve-manager", "default-name", "driver-contract", "custom-driver", "custom-driver-replacement",
		"custom-formatter", "formatter-params", "line-formatter", "text-formatter", "json-formatter",
		"manual-manager", "provider-lifecycle", "provider-cleanup", "resource-cleanup", "closed-writes", "global-logrus",
	} {
		if item, ok := Find("logger", slug); !ok || item.Status != StatusPlanned {
			t.Fatalf("logger catalog entry %q = %#v, %v; want planned", slug, item, ok)
		}
	}
	for _, slug := range []string{
		"architecture", "requirements", "cron-command", "signal-shutdown", "registration", "deployment",
		"timezone", "debug-logging", "overlap-cache", "exception-config",
		"command", "command-validation", "call", "every", "second-frequencies", "minute-frequencies",
		"hourly", "hourly-at", "hour-steps", "daily", "daily-at", "twice-daily", "twice-daily-at", "at",
		"weekly", "weekly-on", "weekday-groups", "named-weekdays", "days",
		"monthly", "monthly-on", "twice-monthly", "last-day-of-month", "days-of-month",
		"quarterly", "quarterly-on", "yearly", "yearly-on",
		"without-overlapping", "cross-process-overlap", "start", "stop", "summary",
		"task-error", "task-panic", "task-success", "exception-reporter",
		"fixed-interval-model", "calendar-model", "immediate-run", "execution-concurrency",
		"time-parsing", "offset-parsing", "weekday-parsing",
		"timer-type", "resolved-command", "command-resolver", "new-schedule", "name", "description", "defaults", "standalone",
	} {
		if item, ok := Find("timer", slug); !ok || item.Status != StatusPlanned {
			t.Fatalf("timer catalog entry %q = %#v, %v; want planned", slug, item, ok)
		}
	}
	for _, slug := range []string{
		"architecture", "config", "local-prerequisites", "oss-prerequisites",
		"top-level-config", "local-config", "public-config", "oss-config", "links-config", "disk-config",
		"facade", "named-disk", "cloud", "interface-selection",
		"get", "json", "open-stream", "read-stream", "download", "file-existence", "directory-existence",
		"put", "put-reader", "prepend-append", "put-options", "put-file", "put-file-as", "upload-fields",
		"copy-move", "cross-disk-guard", "delete", "size", "last-modified", "file-info", "mime-type", "checksum", "path",
		"make-directory", "files", "all-files", "directories", "all-directories", "delete-directory",
		"public-url", "storage-link", "storage-link-options", "storage-unlink",
		"temporary-url", "temporary-url-capability", "temporary-upload-url", "temporary-upload-capability", "verify-temporary-url",
		"local-visibility", "oss-visibility",
		"custom-driver", "custom-driver-lifecycle", "driver-contract", "optional-driver-capabilities", "driver-factory-context",
		"manual-manager", "manager-from-config", "errors", "local-capabilities", "oss-capabilities",
		"laravel-compatibility", "best-practices",
	} {
		if item, ok := Find("filesystem", slug); !ok || item.Status != StatusPlanned {
			t.Fatalf("filesystem catalog entry %q = %#v, %v; want planned", slug, item, ok)
		}
	}
	if got := Filter("redis", LevelIntegration, StatusPlanned); len(got) != 1 {
		t.Fatalf("redis integration filter returned %d entries, want 1", len(got))
	}
	if got := Filter("queue", "", StatusImplemented); len(got) != 59 {
		t.Fatalf("implemented queue entries = %d, want 59", len(got))
	}
	if got := Filter("queue", "", StatusPlanned); len(got) != 0 {
		t.Fatalf("planned queue entries = %d, want 0", len(got))
	}
}

func TestCatalogFeatureSummaries(t *testing.T) {
	summaries := Summaries()
	if len(summaries) != 29 {
		t.Fatalf("Summaries() returned %d features, want 29", len(summaries))
	}

	queue, ok := SummaryFor("queue")
	if !ok {
		t.Fatal("SummaryFor(queue) found = false")
	}
	if queue.Implemented != 59 || queue.Planned != 0 || queue.Manual != 0 || queue.Total != 59 || queue.Remaining != 0 {
		t.Fatalf("queue summary = %#v, want implemented=59 planned=0 manual=0 total=59 remaining=0", queue)
	}
	if queue.Status != FeatureStatusImplemented || queue.Since != SinceInitial {
		t.Fatalf("queue status/since = %q/%q, want %q/%q", queue.Status, queue.Since, FeatureStatusImplemented, SinceInitial)
	}

	cache, ok := SummaryFor("cache")
	if !ok {
		t.Fatal("SummaryFor(cache) found = false")
	}
	if cache.Implemented != 0 || cache.Planned != 57 || cache.Manual != 0 || cache.Total != 57 || cache.Remaining != 57 {
		t.Fatalf("cache summary = %#v, want implemented=0 planned=57 manual=0 total=57 remaining=57", cache)
	}
	if cache.Status != FeatureStatusPlanned || cache.Since != SinceInitial {
		t.Fatalf("cache status/since = %q/%q, want %q/%q", cache.Status, cache.Since, FeatureStatusPlanned, SinceInitial)
	}

	filesystem, ok := SummaryFor("filesystem")
	if !ok {
		t.Fatal("SummaryFor(filesystem) found = false")
	}
	if filesystem.Implemented != 1 || filesystem.Planned != 66 || filesystem.Manual != 0 || filesystem.Total != 67 || filesystem.Remaining != 66 {
		t.Fatalf("filesystem summary = %#v, want implemented=1 planned=66 manual=0 total=67 remaining=66", filesystem)
	}
	if filesystem.Status != FeatureStatusInProgress || filesystem.Since != SinceInitial {
		t.Fatalf("filesystem status/since = %q/%q, want %q/%q", filesystem.Status, filesystem.Since, FeatureStatusInProgress, SinceInitial)
	}

	horizonSummary, ok := SummaryFor("horizon")
	if !ok {
		t.Fatal("SummaryFor(horizon) found = false")
	}
	if horizonSummary.Implemented != 1 || horizonSummary.Planned != 0 || horizonSummary.Total != 1 || horizonSummary.Remaining != 0 {
		t.Fatalf("horizon summary = %#v, want implemented=1 planned=0 total=1 remaining=0", horizonSummary)
	}
	if horizonSummary.Status != FeatureStatusImplemented {
		t.Fatalf("horizon status = %q, want %q", horizonSummary.Status, FeatureStatusImplemented)
	}

	translation, ok := SummaryFor("translation")
	if !ok {
		t.Fatal("SummaryFor(translation) found = false")
	}
	if translation.Implemented != 39 || translation.Planned != 0 || translation.Manual != 0 || translation.Total != 39 || translation.Remaining != 0 {
		t.Fatalf("translation summary = %#v, want implemented=39 planned=0 manual=0 total=39 remaining=0", translation)
	}
	if translation.Status != FeatureStatusImplemented || translation.Since != SinceInitial {
		t.Fatalf("translation status/since = %q/%q, want %q/%q", translation.Status, translation.Since, FeatureStatusImplemented, SinceInitial)
	}

	event, ok := SummaryFor("event")
	if !ok {
		t.Fatal("SummaryFor(event) found = false")
	}
	if event.Implemented != 0 || event.Planned != 57 || event.Manual != 0 || event.Total != 57 || event.Remaining != 57 {
		t.Fatalf("event summary = %#v, want implemented=0 planned=57 manual=0 total=57 remaining=57", event)
	}
	if event.Status != FeatureStatusPlanned || event.Since != SinceInitial {
		t.Fatalf("event status/since = %q/%q, want %q/%q", event.Status, event.Since, FeatureStatusPlanned, SinceInitial)
	}

	logger, ok := SummaryFor("logger")
	if !ok {
		t.Fatal("SummaryFor(logger) found = false")
	}
	if logger.Implemented != 0 || logger.Planned != 40 || logger.Manual != 0 || logger.Total != 40 || logger.Remaining != 40 {
		t.Fatalf("logger summary = %#v, want implemented=0 planned=40 manual=0 total=40 remaining=40", logger)
	}
	if logger.Status != FeatureStatusPlanned || logger.Since != SinceInitial {
		t.Fatalf("logger status/since = %q/%q, want %q/%q", logger.Status, logger.Since, FeatureStatusPlanned, SinceInitial)
	}

	timer, ok := SummaryFor("timer")
	if !ok {
		t.Fatal("SummaryFor(timer) found = false")
	}
	if timer.Implemented != 0 || timer.Planned != 62 || timer.Manual != 0 || timer.Total != 62 || timer.Remaining != 62 {
		t.Fatalf("timer summary = %#v, want implemented=0 planned=62 manual=0 total=62 remaining=62", timer)
	}
	if timer.Status != FeatureStatusPlanned || timer.Since != SinceInitial {
		t.Fatalf("timer status/since = %q/%q, want %q/%q", timer.Status, timer.Since, FeatureStatusPlanned, SinceInitial)
	}

	ratelimit, ok := SummaryFor("ratelimit")
	if !ok {
		t.Fatal("SummaryFor(ratelimit) found = false")
	}
	if ratelimit.Implemented != 0 || ratelimit.Planned != 74 || ratelimit.Manual != 0 || ratelimit.Total != 74 || ratelimit.Remaining != 74 {
		t.Fatalf("ratelimit summary = %#v, want implemented=0 planned=74 manual=0 total=74 remaining=74", ratelimit)
	}
	if ratelimit.Status != FeatureStatusPlanned || ratelimit.Since != SinceInitial {
		t.Fatalf("ratelimit status/since = %q/%q, want %q/%q", ratelimit.Status, ratelimit.Since, FeatureStatusPlanned, SinceInitial)
	}

	commands, ok := SummaryFor("commands")
	if !ok || commands.Status != FeatureStatusImplemented {
		t.Fatalf("commands summary = %#v, %v; want implemented", commands, ok)
	}
	installation, ok := SummaryFor("installation")
	if !ok || installation.Status != FeatureStatusManual || installation.Remaining != 0 {
		t.Fatalf("installation summary = %#v, %v; want manual with no remaining planned entries", installation, ok)
	}
}

func TestFeatureStatus(t *testing.T) {
	tests := []struct {
		name    string
		summary Summary
		want    FeatureStatus
	}{
		{name: "implemented", summary: Summary{Implemented: 2, Total: 2}, want: FeatureStatusImplemented},
		{name: "in progress", summary: Summary{Implemented: 1, Planned: 1, Total: 2}, want: FeatureStatusInProgress},
		{name: "planned", summary: Summary{Planned: 2, Total: 2}, want: FeatureStatusPlanned},
		{name: "manual", summary: Summary{Manual: 2, Total: 2}, want: FeatureStatusManual},
		{name: "empty", summary: Summary{}, want: FeatureStatusPlanned},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := featureStatus(test.summary); got != test.want {
				t.Fatalf("featureStatus(%#v) = %q, want %q", test.summary, got, test.want)
			}
		})
	}
}

func TestCatalogValidationRejectsInvalidEntries(t *testing.T) {
	valid := All()[0]
	tests := []struct {
		name  string
		items []Entry
	}{
		{name: "empty", items: []Entry{{}}},
		{name: "since", items: []Entry{func() Entry { item := valid; item.Since = ""; return item }()}},
		{name: "feature metadata", items: []Entry{func() Entry { item := valid; item.Feature = "missing"; return item }()}},
		{name: "level", items: []Entry{func() Entry { item := valid; item.Level = "unknown"; return item }()}},
		{name: "status", items: []Entry{func() Entry { item := valid; item.Status = "unknown"; return item }()}},
		{name: "duplicate", items: []Entry{valid, valid}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := Validate(test.items); err == nil {
				t.Fatal("Validate() error = nil")
			}
		})
	}
}

func TestCatalogDocuments(t *testing.T) {
	docsRoot, ok := FindDocsRoot(".")
	if !ok {
		t.Skip("sibling docs checkout is not available")
	}
	if err := ValidateDocuments(docsRoot, All()); err != nil {
		t.Fatalf("validate catalog documents: %v", err)
	}
}

func TestFindDocsRootAndMissingHeading(t *testing.T) {
	root := t.TempDir()
	docsRoot := filepath.Join(root, "docs")
	for _, locale := range []string{"zh_CN", "en"} {
		if err := os.MkdirAll(filepath.Join(docsRoot, locale), 0o755); err != nil {
			t.Fatalf("create %s docs directory: %v", locale, err)
		}
		if err := os.WriteFile(filepath.Join(docsRoot, locale, "sample.md"), []byte("# Sample\n\n## Present\n"), 0o644); err != nil {
			t.Fatalf("write %s sample document: %v", locale, err)
		}
	}
	found, ok := FindDocsRoot(filepath.Join(root, "nested"))
	if !ok || found != docsRoot {
		t.Fatalf("FindDocsRoot() = %q, %v; want %q, true", found, ok, docsRoot)
	}
	item := Entry{DocumentZH: "zh_CN/sample.md", DocumentEN: "en/sample.md", HeadingZH: "Missing", HeadingEN: "Present"}
	if err := ValidateDocuments(docsRoot, []Entry{item}); err == nil || !strings.Contains(err.Error(), "Missing") {
		t.Fatalf("missing heading error = %v", err)
	}
}
