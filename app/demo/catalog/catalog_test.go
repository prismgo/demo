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
	if len(entries) != 1432 {
		t.Fatalf("catalog has %d entries, want 1432", len(entries))
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
		"architecture", "dependency-resolution", "use-cases", "bind", "transient-lifecycle", "singleton", "singleton-retry",
		"instance", "nil-instance", "alias", "alias-close-order", "with-closer", "with-context-closer", "with-close-group",
		"make", "make-order", "factory", "typed-make", "typed-mismatch", "value", "value-zero",
		"call", "call-positional", "call-results", "call-limits", "has", "bound", "resolved", "list", "forget", "forget-no-close",
		"close-groups", "close-group", "close-order", "closer-ownership", "close-pre-cancel", "close-success", "close-retry", "close-mid-cancel",
		"missing-loader", "missing-loader-error", "facade", "provider-lifecycle", "error-not-registered", "error-nil-result", "error-no-current", "laravel-mapping",
	} {
		if item, ok := Find("container", slug); !ok || item.Status != StatusPlanned {
			t.Fatalf("container catalog entry %q = %#v, found=%v; want planned", slug, item, ok)
		}
	}
	for _, slug := range []string{
		"architecture", "contract", "register-contract", "container-access", "preserve-binding",
		"bind", "singleton", "instance", "alias", "with-closer", "close-group",
		"boot-order", "boot-listeners", "commands", "command-inputs", "command-deferred",
		"publishes", "publish-tags", "publish-environment",
		"application-registration", "base-order", "default-order", "extension-order", "application-order",
		"named-identity", "implicit-identity", "duplicate-identity",
		"deferrable-contract", "provides", "deferred-resolution", "deferred-map-cleanup", "deferred-late-boot", "deferred-empty", "deferred-conflict", "deferred-termination",
		"terminable-contract", "worker-lifecycle", "terminate-order", "terminate-eligibility", "terminate-context", "closer-order", "full-lifecycle",
	} {
		if item, ok := Find("service-provider", slug); !ok || item.Status != StatusPlanned {
			t.Fatalf("service-provider catalog entry %q = %#v, %v; want planned", slug, item, ok)
		}
	}
	for _, slug := range []string{
		"architecture", "config", "driver-prerequisites", "top-level-config",
		"memory-config", "redis-config", "file-config", "failover-config", "lock-config", "flexible-config",
		"facade", "named-store", "repository", "missing-store",
		"get", "fallbacks", "typed-retrieval", "existence", "put", "forever",
		"add", "put-many", "remember", "remember-forever", "flexible", "touch", "many", "pull",
		"forget", "forget-many", "flush", "counters", "lock", "lock-callback", "lock-block", "lock-restore",
		"lock-flush", "funnel", "without-overlapping", "tags-memory",
		"tags-redis", "tags-unsupported", "memo", "failover", "custom-driver", "resource-lifecycle",
		"events", "event-contract", "deferred", "key-prefixes", "encoding", "errors",
		"memory-capabilities", "file-capabilities", "redis-capabilities", "failover-capabilities", "laravel-compatibility",
	} {
		if item, ok := Find("cache", slug); !ok || item.Status != StatusImplemented {
			t.Fatalf("cache catalog entry %q = %#v, %v; want implemented", slug, item, ok)
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
		"application-entry", "run-context",
		"base-providers", "default-providers", "extension-providers", "application-providers", "exception-handler",
		"provider-layering", "register-phase", "boot-phase", "booted-runner-order",
		"deferred-provider-map", "deferred-provider-resolution", "dynamic-provider",
		"http-pipeline", "request-id", "access-log", "exception-middleware", "business-middleware", "request-error-log",
		"handle-command", "console-starting", "command-resolution", "command-execution", "shared-application",
		"runner-shutdown", "signal-shutdown", "root-context-cancel", "terminating-event-order", "terminable-providers",
		"cleanup-functions", "resource-close-order", "shutdown-error-reporting", "terminated-event-order", "close-retry",
		"best-effort-events", "app-booting-event", "app-booted-event", "app-terminating-event", "app-terminated-event",
		"provider-registering-event", "provider-registered-event", "provider-booting-event", "provider-booted-event",
		"server-starting-event", "server-started-event", "server-stopping-event", "server-stopped-event",
		"request-received-event", "request-handled-event", "request-failed-event", "request-finished-event",
		"console-application-starting-event", "console-command-starting-event", "console-command-finished-event",
		"listeners", "payload-boundaries",
	} {
		if item, ok := Find("lifecycle", slug); !ok || item.Status != StatusPlanned {
			t.Fatalf("lifecycle catalog entry %q = %#v, %v; want planned", slug, item, ok)
		}
	}
	for _, slug := range []string{
		"architecture", "config", "file-driver", "redis-driver", "top-level-config", "cookie-config",
		"file-config", "redis-config", "lock-config", "manager-config",
		"middleware", "recovery", "response-buffering", "store-from", "custom-manager",
		"get", "all", "subsets", "has", "exists", "missing", "put", "counters",
		"flash", "now", "reflash", "keep", "forget", "flush", "pull", "regenerate", "invalidate",
		"blocking", "file-lock", "redis-lock", "id-cookie", "expire-on-close", "queued-cookies",
		"encryption", "encryptor", "custom-encryptor", "sensitive-error",
		"driver-contract", "locker-contract", "extend", "extend-validation", "unknown-driver",
		"errors", "recoverable-errors", "laravel-compatibility",
	} {
		if item, ok := Find("session", slug); !ok || item.Status != StatusPlanned {
			t.Fatalf("session catalog entry %q = %#v, %v; want planned", slug, item, ok)
		}
	}
	for _, slug := range []string{"redis-driver", "redis-lock"} {
		item, ok := Find("session", slug)
		if !ok || item.Level != LevelIntegration || len(item.Requirements) != 1 || item.Requirements[0] != "redis" {
			t.Fatalf("session Redis entry %q = %#v, %v; want integration requiring redis", slug, item, ok)
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
		"architecture", "config", "client-identifiers", "named-config", "address-precedence", "database-precedence", "authentication",
		"url", "tls", "client-name", "timeouts", "max-retries", "environment",
		"client", "named-client", "connection", "strings", "hashes", "lists", "sets", "sorted-sets", "counters", "keys", "native-client",
		"transaction", "lua", "pipeline", "publish", "subscribe", "psubscribe",
		"manager", "manager-repository", "manager-application", "default-connection", "named-connection", "default-connection-method", "lazy-connection",
		"multiple-connections", "connection-reuse", "connections-snapshot", "snapshot-lazy", "purge", "purge-rebuild",
		"close", "close-errors", "close-cancellation",
		"command-executed-event", "command-failed-event", "batch-executed-event", "batch-failed-event", "event-payloads", "event-sensitive-parameters",
		"global-command-listener", "global-failure-listener", "global-batch-listener",
		"connection-listener", "connection-failure-listener", "listener-panic", "disable-events", "enable-events",
		"provider-registration", "provider-event-bridge", "container-factory", "container-connection", "container-named-connection", "lifecycle-close",
		"cache-driver", "cache-basic", "cache-ttl", "cache-atomic", "cache-bulk", "cache-tags", "cache-flush",
		"queue-driver", "queue-ready", "queue-delayed", "queue-blocking-pop", "queue-failed",
		"horizon-config", "horizon-processes", "horizon-control", "horizon-metrics", "horizon-queue-lengths", "horizon-summaries",
		"horizon-job-diagnostics", "horizon-observability", "horizon-orphans",
		"facade-manager", "manager-close-option", "manager-contract", "connection-contract", "event-aliases", "event-wrappers",
	} {
		if item, ok := Find("redis", slug); !ok || item.Status != StatusPlanned {
			t.Fatalf("redis catalog entry %q = %#v, %v; want planned", slug, item, ok)
		}
	}
	for _, slug := range []string{
		"architecture", "facade", "mount", "isolated-router",
		"http-methods", "match", "any",
		"laravel-parameters", "gin-parameters", "parameter-read", "optional-parameters", "wildcard-parameters",
		"where", "where-number", "where-alpha", "where-alphanumeric", "where-uuid", "where-ulid", "where-in",
		"group-constraints", "global-pattern", "constraint-override",
		"route-name", "url", "url-escaping", "url-missing-parameter", "group-name-prefix", "duplicate-route-name",
		"group-prefix", "group-chain", "nested-groups",
		"group-middleware", "route-middleware", "named-middleware", "middleware-function-name", "route-without-middleware", "registrar-without-middleware",
		"bind", "model", "binding-context", "missing-handler", "default-missing",
		"controller-action", "controller-validation",
		"api-resource", "resource", "resource-create", "resource-edit", "resource-only", "resource-except", "resource-names", "resource-parameters",
		"api-resources", "nested-resource", "resource-controller-contract", "create-controller-contract", "edit-controller-contract",
		"redirect", "permanent-redirect", "static", "fallback", "domain", "domain-placeholder",
		"rate-limiter", "limit", "throttle-route", "throttle-group", "throttle-unknown", "throttle-over-limit",
		"current-route", "route-info", "list", "list-command", "handler-order",
		"resolve", "reset", "clone", "add", "router-group", "route-scope-bindings", "registrar-scope-bindings", "registrar-overrides", "facade-contract",
		"provider-registration", "provider-preserves-router", "provider-singleton", "best-practices",
	} {
		if item, ok := Find("route", slug); !ok || item.Status != StatusPlanned {
			t.Fatalf("route catalog entry %q = %#v, %v; want planned", slug, item, ok)
		}
	}
	for _, slug := range []string{
		"architecture", "sqlite-extension", "sqlite-connection-scope",
		"default-string-length", "default-time-precision", "default-morph-key-type", "morph-using-uuids", "morph-using-ulids", "explicit-tag-precedence",
		"bind", "new", "named-connection", "facade",
		"create", "create-validation", "create-dialect-options", "table", "add-column", "change-column", "rename-column", "drop-column", "drop-columns", "raw",
		"rename", "drop", "builder-drop-columns", "drop-all-tables", "drop-all-views", "drop-all-types",
		"id", "increments", "signed-integers", "unsigned-integers", "string-char", "text-types", "uuid-ulid", "network-addresses", "remember-token",
		"boolean", "floating-point", "decimal", "date", "datetime", "time", "timestamp", "year", "timestamps", "soft-deletes",
		"binary", "json", "enum-set", "spatial-types", "vector", "foreign-id", "foreign-id-for", "morphs", "nullable-morphs", "nullable-timestamps",
		"nullable", "not-null", "unsigned", "auto-increment", "primary-modifier", "unique-modifier", "index-modifier", "default", "comment", "first", "after",
		"charset", "collation", "use-current", "use-current-on-update", "invisible", "stored-as", "virtual-as", "from", "instant", "lock", "change-semantics",
		"drop-remember-token", "drop-timestamps", "drop-timestamps-tz", "drop-soft-deletes", "drop-soft-deletes-tz", "drop-morphs", "drop-constrained-foreign-id", "drop-foreign-id-for",
		"primary-index", "unique-index", "index", "fulltext-index", "spatial-index", "named-indexes", "index-naming", "rename-index",
		"drop-index", "drop-unique", "drop-primary", "drop-fulltext", "drop-spatial-index",
		"constrained", "constrained-explicit", "foreign", "foreign-actions", "cascade-actions", "restrict-actions", "null-actions", "no-action-actions", "foreign-name", "foreign-dialect", "drop-foreign",
		"table-view-existence", "tables", "table-listing", "views", "schemas", "types", "schema-filter", "has-columns", "columns", "column-type", "has-index", "indexes", "foreign-keys", "metadata-types",
		"when-has-column", "when-missing-column", "when-missing-index",
		"disable-foreign-keys", "enable-foreign-keys", "without-foreign-keys", "foreign-key-toggle-dialects",
		"create-database", "drop-database", "ensure-extension", "ensure-vector-extension",
		"sync-models", "sync-models-columns", "sync-models-defaults", "sync-models-boundaries", "dialect-compatibility", "laravel-compatibility",
	} {
		if item, ok := Find("schema", slug); !ok || item.Status != StatusPlanned {
			t.Fatalf("schema catalog entry %q = %#v, %v; want planned", slug, item, ok)
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
	if got := Filter("redis", LevelIntegration, StatusPlanned); len(got) != 54 {
		t.Fatalf("redis integration filter returned %d entries, want 54", len(got))
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
	if cache.Implemented != 57 || cache.Planned != 0 || cache.Manual != 0 || cache.Total != 57 || cache.Remaining != 0 {
		t.Fatalf("cache summary = %#v, want implemented=57 planned=0 manual=0 total=57 remaining=0", cache)
	}
	if cache.Status != FeatureStatusImplemented || cache.Since != SinceInitial {
		t.Fatalf("cache status/since = %q/%q, want %q/%q", cache.Status, cache.Since, FeatureStatusInProgress, SinceInitial)
	}

	containerSummary, ok := SummaryFor("container")
	if !ok || containerSummary.Implemented != 0 || containerSummary.Planned != 47 || containerSummary.Manual != 0 || containerSummary.Total != 47 || containerSummary.Remaining != 47 || containerSummary.Status != FeatureStatusPlanned {
		t.Fatalf("container summary = %#v, found=%v; want 47 planned entries and no implemented or manual entries", containerSummary, ok)
	}

	config, ok := SummaryFor("config")
	if !ok || config.Implemented != 0 || config.Planned != 62 || config.Manual != 0 || config.Total != 62 || config.Remaining != 62 || config.Status != FeatureStatusPlanned {
		t.Fatalf("config summary = %#v, found=%v; want 62 planned entries and no implemented or manual entries", config, ok)
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

	sessionSummary, ok := SummaryFor("session")
	if !ok {
		t.Fatal("SummaryFor(session) found = false, want true")
	}
	if sessionSummary.Implemented != 0 || sessionSummary.Planned != 50 || sessionSummary.Manual != 0 || sessionSummary.Total != 50 || sessionSummary.Remaining != 50 {
		t.Fatalf("session summary = %#v, want implemented=0 planned=50 manual=0 total=50 remaining=50", sessionSummary)
	}
	if sessionSummary.Status != FeatureStatusPlanned || sessionSummary.Since != SinceInitial {
		t.Fatalf("session status/since = %q/%q, want %q/%q", sessionSummary.Status, sessionSummary.Since, FeatureStatusPlanned, SinceInitial)
	}

	lifecycle, ok := SummaryFor("lifecycle")
	if !ok {
		t.Fatal("SummaryFor(lifecycle) found = false")
	}
	if lifecycle.Implemented != 0 || lifecycle.Planned != 57 || lifecycle.Manual != 0 || lifecycle.Total != 57 || lifecycle.Remaining != 57 {
		t.Fatalf("lifecycle summary = %#v, want implemented=0 planned=57 manual=0 total=57 remaining=57", lifecycle)
	}
	if lifecycle.Status != FeatureStatusPlanned || lifecycle.Since != SinceInitial {
		t.Fatalf("lifecycle status/since = %q/%q, want %q/%q", lifecycle.Status, lifecycle.Since, FeatureStatusPlanned, SinceInitial)
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

	redis, ok := SummaryFor("redis")
	if !ok {
		t.Fatal("SummaryFor(redis) found = false")
	}
	if redis.Implemented != 0 || redis.Planned != 93 || redis.Manual != 0 || redis.Total != 93 || redis.Remaining != 93 {
		t.Fatalf("redis summary = %#v, want implemented=0 planned=93 manual=0 total=93 remaining=93", redis)
	}
	if redis.Status != FeatureStatusPlanned || redis.Since != SinceInitial {
		t.Fatalf("redis status/since = %q/%q, want %q/%q", redis.Status, redis.Since, FeatureStatusPlanned, SinceInitial)
	}

	routeSummary, ok := SummaryFor("route")
	if !ok {
		t.Fatal("SummaryFor(route) found = false")
	}
	if routeSummary.Implemented != 0 || routeSummary.Planned != 87 || routeSummary.Manual != 0 || routeSummary.Total != 87 || routeSummary.Remaining != 87 {
		t.Fatalf("route summary = %#v, want implemented=0 planned=87 manual=0 total=87 remaining=87", routeSummary)
	}
	if routeSummary.Status != FeatureStatusPlanned || routeSummary.Since != SinceInitial {
		t.Fatalf("route status/since = %q/%q, want %q/%q", routeSummary.Status, routeSummary.Since, FeatureStatusPlanned, SinceInitial)
	}

	schemaSummary, ok := SummaryFor("schema")
	if !ok {
		t.Fatal("SummaryFor(schema) found = false")
	}
	if schemaSummary.Implemented != 0 || schemaSummary.Planned != 143 || schemaSummary.Manual != 0 || schemaSummary.Total != 143 || schemaSummary.Remaining != 143 {
		t.Fatalf("schema summary = %#v, want implemented=0 planned=143 manual=0 total=143 remaining=143", schemaSummary)
	}
	if schemaSummary.Status != FeatureStatusPlanned || schemaSummary.Since != SinceInitial {
		t.Fatalf("schema status/since = %q/%q, want %q/%q", schemaSummary.Status, schemaSummary.Since, FeatureStatusPlanned, SinceInitial)
	}

	serviceProvider, ok := SummaryFor("service-provider")
	if !ok {
		t.Fatal("SummaryFor(service-provider) found = false")
	}
	if serviceProvider.Implemented != 0 || serviceProvider.Planned != 42 || serviceProvider.Manual != 0 || serviceProvider.Total != 42 || serviceProvider.Remaining != 42 {
		t.Fatalf("service-provider summary = %#v, want implemented=0 planned=42 manual=0 total=42 remaining=42", serviceProvider)
	}
	if serviceProvider.Status != FeatureStatusPlanned || serviceProvider.Since != SinceInitial {
		t.Fatalf("service-provider status/since = %q/%q, want %q/%q", serviceProvider.Status, serviceProvider.Since, FeatureStatusPlanned, SinceInitial)
	}

	commands, ok := SummaryFor("commands")
	if !ok || commands.Status != FeatureStatusImplemented || commands.Implemented != 101 || commands.Planned != 0 || commands.Total != 101 {
		t.Fatalf("commands summary = %#v, %v; want all 101 commands entries implemented", commands, ok)
	}
	installation, ok := SummaryFor("installation")
	if !ok || installation.Status != FeatureStatusManual || installation.Remaining != 0 {
		t.Fatalf("installation summary = %#v, %v; want manual with no remaining planned entries", installation, ok)
	}
}

func TestConfigDocumentationCoverage(t *testing.T) {
	entries := Filter("config", "", "")
	docsRoot, ok := FindDocsRoot(".")
	if !ok {
		t.Skip("sibling docs checkout is not available")
	}

	for _, locale := range []struct {
		path    string
		heading func(Entry) string
	}{
		{path: filepath.Join(docsRoot, "zh_CN", "config.md"), heading: func(item Entry) string { return item.HeadingZH }},
		{path: filepath.Join(docsRoot, "en", "config.md"), heading: func(item Entry) string { return item.HeadingEN }},
	} {
		content, err := os.ReadFile(locale.path)
		if err != nil {
			t.Fatalf("read config documentation %s: %v", locale.path, err)
		}
		configRows := 0
		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "### ") {
				heading := strings.TrimSpace(strings.TrimLeft(line, "#"))
				found := false
				for _, item := range entries {
					if locale.heading(item) == heading {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("config document %s heading %q has no catalog entry; actual entries=%d, want heading mapped", locale.path, heading, len(entries))
				}
			}
			if !strings.HasPrefix(line, "| `app.") {
				continue
			}
			columns := strings.Split(line, "|")
			if len(columns) < 4 {
				t.Fatalf("config document %s row %q has %d columns, want at least 4", locale.path, line, len(columns))
			}
			path := strings.Trim(strings.TrimSpace(columns[1]), "`")
			env := strings.Trim(strings.TrimSpace(columns[2]), "`")
			section := path + " and " + env
			found := false
			for _, item := range entries {
				if item.Section == section {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("config document %s row %q has no catalog entry; want section %q", locale.path, line, section)
			}
			configRows++
		}
		if configRows != 24 {
			t.Errorf("config document %s has %d application configuration rows, want 24", locale.path, configRows)
		}
	}
}

func TestDatabaseDocumentationCoverage(t *testing.T) {
	entries := Filter("database", "", "")
	if len(entries) != 62 {
		t.Fatalf("database catalog entries = %d, want 62", len(entries))
	}
	docsRoot, ok := FindDocsRoot(".")
	if !ok {
		t.Skip("sibling docs checkout is not available")
	}

	for _, locale := range []struct {
		path    string
		heading func(Entry) string
	}{
		{path: filepath.Join(docsRoot, "zh_CN", "database.md"), heading: func(item Entry) string { return item.HeadingZH }},
		{path: filepath.Join(docsRoot, "en", "database.md"), heading: func(item Entry) string { return item.HeadingEN }},
	} {
		content, err := os.ReadFile(locale.path)
		if err != nil {
			t.Fatalf("read database documentation %s: %v", locale.path, err)
		}
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "### ") || strings.HasPrefix(line, "#### ") {
				level := len(line) - len(strings.TrimLeft(line, "#"))
				leaf := true
				for _, next := range lines[i+1:] {
					next = strings.TrimSpace(next)
					if !strings.HasPrefix(next, "#") {
						continue
					}
					nextLevel := len(next) - len(strings.TrimLeft(next, "#"))
					leaf = nextLevel <= level
					break
				}
				if leaf {
					heading := strings.TrimSpace(strings.TrimLeft(line, "#"))
					found := false
					for _, item := range entries {
						if locale.heading(item) == heading {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("database document %s leaf heading %q has no catalog entry; actual entries=%d, want heading mapped", locale.path, heading, len(entries))
					}
				}
			}
			if !strings.HasPrefix(line, "| `database.") {
				continue
			}
			columns := strings.Split(line, "|")
			if len(columns) < 4 {
				t.Fatalf("database document %s row %q has %d columns, want at least 4", locale.path, line, len(columns))
			}
			path := strings.Trim(strings.TrimSpace(columns[1]), "`")
			found := false
			for _, item := range entries {
				if strings.HasPrefix(item.Section, path) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("database document %s config row %q has no catalog entry; want section starting with %q", locale.path, line, path)
			}
		}
	}
}

func TestCommandsDocumentationCoverage(t *testing.T) {
	entries := Filter("commands", "", "")
	if len(entries) != 101 {
		t.Fatalf("commands catalog entries = %d, want 101", len(entries))
	}

	docsRoot, ok := FindDocsRoot(".")
	if !ok {
		t.Skip("sibling docs checkout is not available")
	}
	for _, locale := range []struct {
		path    string
		heading func(Entry) string
	}{
		{path: filepath.Join(docsRoot, "zh_CN", "commands.md"), heading: func(item Entry) string { return item.HeadingZH }},
		{path: filepath.Join(docsRoot, "en", "commands.md"), heading: func(item Entry) string { return item.HeadingEN }},
	} {
		content, err := os.ReadFile(locale.path)
		if err != nil {
			t.Fatalf("read commands documentation %s: %v", locale.path, err)
		}
		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
				continue
			}
			heading := strings.TrimSpace(strings.TrimLeft(line, "#"))
			found := false
			for _, item := range entries {
				if locale.heading(item) == heading {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("commands document %s heading %q has no catalog entry; want a mapped entry", locale.path, heading)
			}
		}
	}
}

func TestConsoleDocumentationCoverage(t *testing.T) {
	entries := Filter("console", "", "")
	if len(entries) != 70 {
		t.Fatalf("console catalog entries = %d, want 70", len(entries))
	}
	for _, item := range entries {
		if item.Status != StatusImplemented {
			t.Errorf("console entry %q status = %q, want %q", item.Case, item.Status, StatusImplemented)
		}
	}

	docsRoot, ok := FindDocsRoot(".")
	if !ok {
		t.Skip("sibling docs checkout is not available")
	}
	for _, locale := range []struct {
		path    string
		heading func(Entry) string
	}{
		{path: filepath.Join(docsRoot, "zh_CN", "console.md"), heading: func(item Entry) string { return item.HeadingZH }},
		{path: filepath.Join(docsRoot, "en", "console.md"), heading: func(item Entry) string { return item.HeadingEN }},
	} {
		content, err := os.ReadFile(locale.path)
		if err != nil {
			t.Fatalf("read console documentation %s: %v", locale.path, err)
		}
		lines := strings.Split(string(content), "\n")
		for index, line := range lines {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
				continue
			}
			if strings.HasPrefix(line, "## ") {
				hasChild := false
				for _, next := range lines[index+1:] {
					next = strings.TrimSpace(next)
					if strings.HasPrefix(next, "## ") {
						break
					}
					if strings.HasPrefix(next, "### ") {
						hasChild = true
						break
					}
				}
				if hasChild {
					continue
				}
			}
			heading := strings.TrimSpace(strings.TrimLeft(line, "#"))
			found := false
			for _, item := range entries {
				if locale.heading(item) == heading {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("console documentation %s heading %q has no catalog entry; actual entries=%d, want heading mapped", locale.path, heading, len(entries))
			}
		}
	}
}

func TestCookieDocumentationCoverage(t *testing.T) {
	sections := []struct {
		zh string
		en string
	}{
		{zh: "简介", en: "Introduction"},
		{zh: "默认值", en: "Default Values"},
		{zh: "服务提供者", en: "Service Provider"},
		{zh: "中间件", en: "Middleware"},
		{zh: "基础创建", en: "Basic Creation"},
		{zh: "长期 Cookie", en: "Forever Cookies"},
		{zh: "构造参数与选项", en: "Constructor Parameters and Options"},
		{zh: "显式过期时间与 Max-Age", en: "Explicit Expiry Time and Max-Age"},
		{zh: "直接附加到响应", en: "Direct Attachment"},
		{zh: "Attach 选项", en: "Attach Options"},
		{zh: "排队 Cookie", en: "Queuing Cookies"},
		{zh: "查询已排队 Cookie", en: "Querying Queued Cookies"},
		{zh: "移除排队项", en: "Removing Queued Items"},
		{zh: "请求级 vs 进程级 API", en: "Request-Level vs Process-Level API"},
		{zh: "基础读取", en: "Basic Retrieval"},
		{zh: "带安全扩展读取", en: "Retrieval with Security Extensions"},
		{zh: "删除 Cookie", en: "Deleting Cookies"},
		{zh: "SameSite 策略", en: "SameSite Policy"},
		{zh: "安全契约接口", en: "Security Contract Interfaces"},
		{zh: "写出与读取顺序", en: "Outgoing and Incoming Order"},
		{zh: "默认透传实现", en: "Default Passthrough Implementation"},
		{zh: "错误脱敏", en: "Error Sanitization"},
		{zh: "错误常量", en: "Error Constants"},
		{zh: "与 Laravel Cookie 的对应关系", en: "Laravel Cookie Mapping"},
	}

	entries := Filter("cookie", "", "")
	for _, section := range sections {
		found := false
		for _, item := range entries {
			if item.HeadingZH == section.zh && item.HeadingEN == section.en {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("cookie catalog missing zh=%q en=%q heading pair: actual entries=%d, want pair present", section.zh, section.en, len(entries))
		}
	}

	summary, ok := SummaryFor("cookie")
	if !ok || summary.Total != 50 || summary.Planned != 50 || summary.Remaining != 50 {
		t.Fatalf("cookie summary = %#v, found=%v; want total=50 planned=50 remaining=50", summary, ok)
	}
}

func TestContainerDocumentationCoverage(t *testing.T) {
	entries := Filter("container", "", "")
	docsRoot, ok := FindDocsRoot(".")
	if !ok {
		t.Skip("sibling docs checkout is not available")
	}

	for _, locale := range []struct {
		path    string
		heading func(Entry) string
	}{
		{path: filepath.Join(docsRoot, "zh_CN", "container.md"), heading: func(item Entry) string { return item.HeadingZH }},
		{path: filepath.Join(docsRoot, "en", "container.md"), heading: func(item Entry) string { return item.HeadingEN }},
	} {
		content, err := os.ReadFile(locale.path)
		if err != nil {
			t.Fatalf("read container documentation %s: %v", locale.path, err)
		}
		lines := strings.Split(string(content), "\n")
		for index, line := range lines {
			line = strings.TrimSpace(line)
			depth := len(line) - len(strings.TrimLeft(line, "#"))
			if depth < 2 || depth > 4 || len(line) <= depth || line[depth] != ' ' {
				continue
			}
			hasChild := false
			for _, next := range lines[index+1:] {
				next = strings.TrimSpace(next)
				nextDepth := len(next) - len(strings.TrimLeft(next, "#"))
				if nextDepth < 2 || nextDepth > 4 || len(next) <= nextDepth || next[nextDepth] != ' ' {
					continue
				}
				hasChild = nextDepth > depth
				break
			}
			if hasChild {
				continue
			}
			heading := strings.TrimSpace(line[depth:])
			found := false
			for _, item := range entries {
				if locale.heading(item) == heading {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("container documentation %s leaf heading %q has no catalog entry; actual entries=%d, want heading mapped", locale.path, heading, len(entries))
			}
		}
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
