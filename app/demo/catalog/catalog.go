// Package catalog maps public documentation topics to runnable demos and tests.
package catalog

import (
	"fmt"
	"sort"
	"strings"
)

// Level describes how a documentation example is verified.
type Level string

const (
	// LevelCompile verifies that an example compiles.
	LevelCompile Level = "compile"
	// LevelHermetic verifies an example without external services.
	LevelHermetic Level = "hermetic"
	// LevelScenario verifies an application-level scenario.
	LevelScenario Level = "scenario"
	// LevelIntegration verifies an example against external services.
	LevelIntegration Level = "integration"
)

// Status distinguishes delivered coverage from planned and manual checks.
type Status string

const (
	// StatusImplemented marks delivered coverage.
	StatusImplemented Status = "implemented"
	// StatusPlanned marks coverage that has not been implemented.
	StatusPlanned Status = "planned"
	// StatusManual marks coverage verified outside the automated catalog.
	StatusManual Status = "manual"
)

// Entry maps one documentation topic to its demo and verification target.
// HeadingZH and HeadingEN hold the real localized Markdown headings while
// Section remains a stable, language-neutral catalog label.
type Entry struct {
	Feature      string   `json:"feature"`
	Since        string   `json:"since"`
	DocumentZH   string   `json:"document_zh"`
	DocumentEN   string   `json:"document_en"`
	Section      string   `json:"section"`
	HeadingZH    string   `json:"heading_zh"`
	HeadingEN    string   `json:"heading_en"`
	Example      string   `json:"example"`
	Case         string   `json:"case"`
	Test         string   `json:"test"`
	Level        Level    `json:"level"`
	Status       Status   `json:"status"`
	Requirements []string `json:"requirements,omitempty"`
	Notes        string   `json:"notes,omitempty"`
}

// Entries is the initial documentation coverage baseline. Planned entries are
// intentionally visible: demo:list summarizes them and demo:show remains the gap report.
var Entries = []Entry{
	baselineEntry("cache", "Cache operations", "缓存操作", "Cache Usage", "demo:cache list", "list", "TestCacheOperations", LevelHermetic, StatusPlanned),
	baselineEntry("commands", "Available commands", "查看命令", "Listing Commands", "demo:list", "list", "TestDemoListCommand", LevelScenario, StatusImplemented),
	baselineEntry("config", "Quick start", "快速开始", "Quick Start", "demo:config list", "list", "TestConfigDemo", LevelHermetic, StatusPlanned),
	baselineEntry("console", "Defining commands", "定义命令", "Defining Commands", "demo:console list", "list", "TestConsoleDemo", LevelScenario, StatusPlanned),
	baselineEntry("container", "Binding", "绑定", "Binding", "demo:container list", "list", "TestContainerDemo", LevelHermetic, StatusPlanned),
	baselineEntry("cookie", "Creating cookies", "创建 Cookie", "Creating Cookies", "demo:cookie list", "list", "TestCookieDemo", LevelHermetic, StatusPlanned),
	baselineEntry("database", "Database connections", "获取数据库连接", "Obtaining a Database Connection", "demo:database list", "list", "TestDatabaseDemo", LevelHermetic, StatusPlanned),
	baselineEntry("encryption", "Key rotation", "轮换密钥", "Rotating Keys", "demo:encryption list", "list", "TestEncryptionDemo", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Listeners", "注册事件与监听器", "Registering Events and Listeners", "demo:event list", "list", "TestEventDemo", LevelHermetic, StatusPlanned),
	baselineEntry("exception", "Exception handling", "异常处理", "Handling Exceptions", "demo:exception list", "list", "TestExceptionDemo", LevelScenario, StatusPlanned),
	baselineEntry("facade", "Available facades", "可用 Facade", "Available Facades", "demo:facade list", "list", "TestFacadeDemo", LevelCompile, StatusPlanned),
	baselineEntry("filesystem", "Disk access", "获取磁盘实例", "Obtaining Disk Instances", "demo:filesystem list", "list", "TestFilesystemDemo", LevelHermetic, StatusPlanned),
	baselineEntry("horizon", "Running Horizon", "运行 Horizon", "Running Horizon", "demo:horizon list", "list", "TestHorizonWithRealQueue", LevelIntegration, StatusPlanned, "redis", "rabbitmq"),
	baselineEntry("http-server", "Starting the server", "启动服务", "Starting the Server", "demo:http-server list", "list", "TestHTTPServerDemo", LevelScenario, StatusPlanned),
	manualEntry("installation", "Creating an application", "创建应用", "Creating an Application", "installer smoke test", "TestInstallerSmoke", LevelIntegration, "requires a clean temporary checkout and network access"),
	manualEntry("lens", "Installation", "安装", "Installation", "Lens workflow verification", "TestLensWorkflow", LevelIntegration, "validated by the Lens toolchain rather than an application command"),
	baselineEntry("lifecycle", "Lifecycle overview", "生命周期概览", "Lifecycle Overview", "demo:lifecycle list", "list", "TestLifecycleDemo", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Writing log messages", "编写日志消息", "Writing Log Messages", "demo:logger list", "list", "TestLoggerDemo", LevelHermetic, StatusPlanned),
	baselineEntry("queue", "Driver prerequisites", "驱动注意事项与前提条件", "Driver Prerequisites", "demo:queue driver-prerequisites", "driver-prerequisites", "TestQueueDemoDriverPrerequisites", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Configuration loading", "配置文件", "Config File", "demo:queue config", "config", "TestQueueDemoConfiguration", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Payload encoding configuration", "顶层配置", "Global", "demo:queue payload-encoding", "payload-encoding", "TestQueueDemoPayloadEncoding", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Sync connection configuration", "Sync 连接", "Sync Connection", "demo:queue sync-connection", "sync-connection", "TestQueueDemoSyncConnectionConfiguration", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Failed store configuration", "失败任务存储", "Failed Jobs Store", "demo:queue failed-store --connection=redis", "failed-store", "TestQueueDemoFailedStoreConfiguration", LevelIntegration, StatusImplemented, "redis"),
	baselineEntry("queue", "Batch store configuration", "批次状态存储", "Batching Store", "demo:queue batch-store --connection=redis", "batch-store", "TestQueueDemoBatchStoreConfiguration", LevelIntegration, StatusImplemented, "redis"),
	baselineEntry("queue", "Restart store configuration", "重启信号存储", "Restart Store", "demo:queue restart-store --connection=redis", "restart-store", "TestQueueDemoRestartStoreConfiguration", LevelIntegration, StatusImplemented, "redis"),
	baselineEntry("queue", "Creating jobs", "创建任务", "Creating Jobs", "demo:queue basic --connection=sync", "basic-sync", "TestBasicScenarioDispatchesAndProcessesSyncJob", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Creating jobs", "创建任务", "Creating Jobs", "demo:queue basic --connection=redis", "basic-redis", "TestQueueDemoBasicWithRealRedis", LevelIntegration, StatusImplemented, "redis"),
	baselineEntry("queue", "Creating jobs", "创建任务", "Creating Jobs", "demo:queue basic --connection=rabbitmq", "basic-rabbitmq", "TestQueueDemoBasicWithRealRabbitMQ", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("queue", "Job strategy interfaces", "任务策略接口", "Job Strategy Interfaces", "demo:queue strategies", "strategies", "TestQueueDemoStrategies", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Strategy envelope and option precedence", "任务策略接口", "Job Strategy Interfaces", "demo:queue strategy-envelope", "strategy-envelope", "TestQueueDemoStrategyEnvelope", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Unique jobs", "唯一任务", "Unique Jobs", "demo:queue unique", "unique", "TestQueueDemoUnique", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Unique dispatch options", "唯一任务", "Unique Jobs", "demo:queue unique-options", "unique-options", "TestQueueDemoUniqueDispatchOptions", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Debounced jobs", "防抖任务", "Debounced Jobs", "demo:queue debounce --connection=redis", "debounce", "TestQueueDemoDebounce", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Debounce dispatch options", "防抖任务", "Debounced Jobs", "demo:queue debounce-options --connection=redis", "debounce-options", "TestQueueDemoDebounceDispatchOptions", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Job middleware", "任务中间件", "Job Middleware", "demo:queue middleware", "middleware", "TestQueueDemoMiddleware", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Overlap release policy", "防止任务重叠", "Preventing Job Overlaps", "demo:queue overlap-release", "overlap-release", "TestQueueDemoOverlapReleasePolicy", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Dispatching jobs", "分发任务", "Dispatching Jobs", "demo:queue dispatch --connection=redis", "dispatch", "TestQueueDemoDispatch", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Job chaining", "任务链", "Job Chaining", "demo:queue chain --connection=redis", "chain", "TestQueueDemoChain", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Job control errors", "错误处理", "Error Handling", "demo:queue job-control --connection=redis", "job-control", "TestQueueDemoJobControlErrors", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Job batching", "任务批处理", "Job Batching", "demo:queue batch --connection=redis", "batch", "TestQueueDemoBatch", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Worker runtime boundaries", "运行队列 Worker", "Running The Queue Worker", "demo:queue worker --connection=redis", "worker", "TestQueueDemoWorker", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Queue worker command", "queue:work 命令", "The queue:work Command", "demo:queue worker-command --connection=redis", "worker-command", "TestQueueDemoWorkerCommand", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Job expiration and timeout policies", "任务过期与超时", "Job Expirations and Timeouts", "demo:queue expiration --connection=redis", "expiration", "TestQueueDemoExpirationPolicies", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Failed job archival and callback", "处理失败任务", "Dealing With Failed Jobs", "demo:queue failure --connection=redis", "failure", "TestQueueDemoFailure", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Failed store maintenance", "清理失败任务", "Cleaning Up After Failed Jobs", "demo:queue failed-commands", "failed-commands", "TestQueueDemoFailedCommands", LevelScenario, StatusImplemented),
	baselineEntry("queue", "Failed cleanup command paths", "清理失败任务", "Cleaning Up After Failed Jobs", "demo:queue failed-command-paths", "failed-command-paths", "TestQueueDemoFailedCleanupCommandPaths", LevelScenario, StatusImplemented),
	baselineEntry("queue", "Failed job retry command", "重试失败任务", "Retrying Failed Jobs", "demo:queue failed-retry --connection=redis", "failed-retry", "TestQueueDemoFailedRetryCommand", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Worker restart", "Worker 与部署", "Queue Workers and Deployment", "demo:queue restart --connection=redis", "restart", "TestQueueDemoRestart", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Job processing lifecycle events", "生命周期事件", "Lifecycle Events", "demo:queue events --connection=redis", "events", "TestQueueDemoEvents", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Failed job lifecycle event", "生命周期事件", "Lifecycle Events", "demo:queue failed-event --connection=redis", "failed-event", "TestQueueDemoFailedEvent", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Batch lifecycle events", "生命周期事件", "Lifecycle Events", "demo:queue batch-events --connection=redis", "batch-events", "TestQueueDemoBatchEvents", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Poison envelope lifecycle event", "生命周期事件", "Lifecycle Events", "demo:queue poison-event --connection=redis", "poison-event", "TestQueueDemoPoisonEnvelopeEvent", LevelIntegration, StatusPlanned, "redis", "rabbitmq"),
	baselineEntry("queue", "Infrastructure lifecycle events", "生命周期事件", "Lifecycle Events", "demo:queue infrastructure-events --connection=rabbitmq", "infrastructure-events", "TestQueueDemoInfrastructureEvents", LevelIntegration, StatusPlanned, "rabbitmq"),
	baselineEntry("queue", "Encrypted payloads", "加密 Payload", "Encrypted Payloads", "demo:queue encryption", "encryption", "TestQueueDemoEncryption", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Missing encryption key failure", "加密 Payload", "Encrypted Payloads", "demo:queue encryption-missing-key", "encryption-missing-key", "TestQueueDemoEncryptionMissingKey", LevelHermetic, StatusPlanned),
	baselineEntry("queue", "Custom driver registration", "自定义驱动", "Custom Drivers", "demo:queue custom-driver", "custom-driver", "TestQueueDemoCustomDriver", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Custom queue contract", "Queue — 队列传输连接", "Queue — Transport Connection", "demo:queue custom-queue-contract", "custom-queue-contract", "TestQueueDemoCustomQueueContract", LevelHermetic, StatusPlanned),
	baselineEntry("queue", "Custom reserved job lifecycle", "ReservedJob — 已保留任务", "ReservedJob — Held Job", "demo:queue custom-reserved-job", "custom-reserved-job", "TestQueueDemoCustomReservedJob", LevelHermetic, StatusPlanned),
	baselineEntry("queue", "Custom pop session provider", "可选接口", "Optional Interfaces", "demo:queue custom-pop-session", "custom-pop-session", "TestQueueDemoCustomPopSession", LevelHermetic, StatusPlanned),
	baselineEntry("queue", "Custom consumer intent leaser", "可选接口", "Optional Interfaces", "demo:queue custom-consumer-intent", "custom-consumer-intent", "TestQueueDemoCustomConsumerIntent", LevelHermetic, StatusPlanned),
	baselineEntry("queue", "Core error constants", "错误常量", "Error Constants", "demo:queue errors", "errors", "TestQueueDemoErrors", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Job state error constants", "错误常量", "Error Constants", "demo:queue job-errors", "job-errors", "TestQueueDemoJobStateErrors", LevelHermetic, StatusPlanned),
	baselineEntry("queue", "Connection error constants", "错误常量", "Error Constants", "demo:queue connection-errors", "connection-errors", "TestQueueDemoConnectionErrors", LevelHermetic, StatusPlanned),
	baselineEntry("queue", "Poison envelope errors", "错误常量", "Error Constants", "demo:queue poison-errors --connection=redis", "poison-errors", "TestQueueDemoPoisonErrors", LevelIntegration, StatusPlanned, "redis", "rabbitmq"),
	baselineEntry("queue", "RabbitMQ error constants", "错误常量", "Error Constants", "demo:queue rabbitmq-errors --connection=rabbitmq", "rabbitmq-errors", "TestQueueDemoRabbitMQErrors", LevelIntegration, StatusPlanned, "rabbitmq"),
	baselineEntry("queue", "Redis transport boundaries", "Redis 连接", "Redis Connection", "demo:queue redis --connection=redis", "redis", "TestQueueDemoRedisBoundaries", LevelIntegration, StatusImplemented, "redis"),
	baselineEntry("queue", "RabbitMQ transport boundaries", "RabbitMQ 配置", "RabbitMQ Configuration", "demo:queue rabbitmq --connection=rabbitmq", "rabbitmq", "TestQueueDemoRabbitMQBoundaries", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("queue", "Bulk transport dispatch", "内置连接能力矩阵", "Built-in Connection Capability Matrix", "demo:queue bulk --connection=redis", "bulk", "TestQueueDemoBulkTransport", LevelIntegration, StatusPlanned, "redis", "rabbitmq"),
	baselineEntry("queue", "Delayed transport dispatch", "内置连接能力矩阵", "Built-in Connection Capability Matrix", "demo:queue transport-delay --connection=redis", "transport-delay", "TestQueueDemoTransportDelay", LevelIntegration, StatusPlanned, "redis", "rabbitmq"),
	baselineEntry("queue", "Blocking pop and queue priority", "内置连接能力矩阵", "Built-in Connection Capability Matrix", "demo:queue blocking-pop --connection=redis", "blocking-pop", "TestQueueDemoBlockingPop", LevelIntegration, StatusPlanned, "redis", "rabbitmq"),
	baselineEntry("queue", "Redis retry-after reservation", "内置连接能力矩阵", "Built-in Connection Capability Matrix", "demo:queue redis-retry-after --connection=redis", "redis-retry-after", "TestQueueDemoRedisRetryAfter", LevelIntegration, StatusPlanned, "redis"),
	baselineEntry("queue", "RabbitMQ unsupported retry-after", "内置连接能力矩阵", "Built-in Connection Capability Matrix", "demo:queue rabbitmq-retry-after --connection=rabbitmq", "rabbitmq-retry-after", "TestQueueDemoRabbitMQRetryAfter", LevelIntegration, StatusPlanned, "rabbitmq"),
	baselineEntry("queue", "RabbitMQ publisher confirms", "RabbitMQ 配置", "RabbitMQ Configuration", "demo:queue rabbitmq-confirm --connection=rabbitmq", "rabbitmq-confirm", "TestQueueDemoRabbitMQPublisherConfirm", LevelIntegration, StatusPlanned, "rabbitmq"),
	baselineEntry("queue", "RabbitMQ reconnect lifecycle", "RabbitMQ 配置", "RabbitMQ Configuration", "demo:queue rabbitmq-reconnect --connection=rabbitmq", "rabbitmq-reconnect", "TestQueueDemoRabbitMQReconnect", LevelIntegration, StatusPlanned, "rabbitmq"),
	baselineEntry("queue", "RabbitMQ topology declaration", "RabbitMQ 配置", "RabbitMQ Configuration", "demo:queue rabbitmq-topology --connection=rabbitmq", "rabbitmq-topology", "TestQueueDemoRabbitMQTopology", LevelIntegration, StatusPlanned, "rabbitmq"),
	baselineEntry("queue", "RabbitMQ delay modes", "RabbitMQ 配置", "RabbitMQ Configuration", "demo:queue rabbitmq-delay-modes --connection=rabbitmq", "rabbitmq-delay-modes", "TestQueueDemoRabbitMQDelayModes", LevelIntegration, StatusPlanned, "rabbitmq"),
	baselineEntry("queue", "Poison envelope rejection", "内置连接能力矩阵", "Built-in Connection Capability Matrix", "demo:queue poison-rejection --connection=redis", "poison-rejection", "TestQueueDemoPoisonEnvelopeRejection", LevelIntegration, StatusPlanned, "redis", "rabbitmq"),
	baselineEntry("ratelimit", "Feature overview", "功能概览", "Feature Overview", "demo:ratelimit list", "list", "TestRateLimitDemo", LevelScenario, StatusPlanned),
	baselineEntry("redis", "Interacting with Redis", "与 Redis 交互", "Interacting With Redis", "demo:redis list", "list", "TestRedisWithRealServer", LevelIntegration, StatusPlanned, "redis"),
	baselineEntry("route", "Quick start", "快速开始", "Quick Start", "demo:route list", "list", "TestRouteDemo", LevelScenario, StatusPlanned),
	baselineEntry("schema", "Builder entry points", "Builder 入口", "Builder Entry Points", "demo:schema list", "list", "TestSchemaDemo", LevelHermetic, StatusPlanned),
	baselineEntry("service-provider", "Writing providers", "编写服务提供者", "Writing Service Providers", "demo:provider list", "list", "TestProviderDemo", LevelHermetic, StatusPlanned),
	baselineEntry("session", "Using sessions", "Session 使用", "Session Usage", "demo:session list", "list", "TestSessionDemo", LevelScenario, StatusPlanned),
	manualEntry("starter", "First endpoint", "添加第一个接口", "Add Your First Endpoint", "starter HTTP smoke test", "TestStarterSmoke", LevelScenario, "runs against a clean generated starter"),
	baselineEntry("support", "Feature overview", "功能概览", "Feature Overview", "demo:support list", "list", "TestSupportDemo", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Defining tasks", "定义调度任务", "Defining Scheduled Tasks", "demo:timer list", "list", "TestTimerDemo", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Overview", "概述", "Overview", "demo:translation list", "list", "TestTranslationDemo", LevelHermetic, StatusPlanned),
}

func baselineEntry(feature, section, headingZH, headingEN, example, caseName, test string, level Level, status Status, requirements ...string) Entry {
	return versionedEntry(SinceInitial, feature, section, headingZH, headingEN, example, caseName, test, level, status, requirements...)
}

// versionedEntry keeps release provenance explicit for capabilities added after
// the initial catalog baseline.
func versionedEntry(since, feature, section, headingZH, headingEN, example, caseName, test string, level Level, status Status, requirements ...string) Entry {
	return Entry{
		Feature: feature, Since: since, DocumentZH: "zh_CN/" + feature + ".md", DocumentEN: "en/" + feature + ".md",
		Section: section, HeadingZH: headingZH, HeadingEN: headingEN, Example: example, Case: caseName,
		Test: test, Level: level, Status: status, Requirements: requirements,
	}
}

func manualEntry(feature, section, headingZH, headingEN, example, test string, level Level, notes string) Entry {
	item := baselineEntry(feature, section, headingZH, headingEN, example, "manual", test, level, StatusManual)
	item.Notes = notes
	return item
}

// All returns a copy sorted by feature and case.
func All() []Entry {
	result := make([]Entry, len(Entries))
	for i, item := range Entries {
		result[i] = cloneEntry(item)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Feature == result[j].Feature {
			return result[i].Case < result[j].Case
		}
		return result[i].Feature < result[j].Feature
	})
	return result
}

// Filter returns catalog entries matching non-empty filters.
func Filter(feature string, level Level, status Status) []Entry {
	feature = strings.TrimSpace(feature)
	result := make([]Entry, 0)
	for _, item := range All() {
		if feature != "" && item.Feature != feature {
			continue
		}
		if level != "" && item.Level != level {
			continue
		}
		if status != "" && item.Status != status {
			continue
		}
		result = append(result, item)
	}
	return result
}

// Find resolves one feature/case pair.
func Find(feature, caseName string) (Entry, bool) {
	for _, item := range Entries {
		if item.Feature == feature && item.Case == caseName {
			return cloneEntry(item), true
		}
	}
	return Entry{}, false
}

func cloneEntry(item Entry) Entry {
	item.Requirements = append([]string(nil), item.Requirements...)
	return item
}

// Validate checks catalog completeness and uniqueness without accessing docs.
func Validate(entries []Entry) error {
	seen := make(map[string]struct{}, len(entries))
	for index, item := range entries {
		if strings.TrimSpace(item.Feature) == "" || strings.TrimSpace(item.Section) == "" ||
			strings.TrimSpace(item.Example) == "" || strings.TrimSpace(item.Test) == "" || strings.TrimSpace(item.Case) == "" ||
			strings.TrimSpace(item.Since) == "" {
			return fmt.Errorf("catalog entry %d has an empty required field", index)
		}
		if _, ok := LookupFeature(item.Feature); !ok {
			return fmt.Errorf("catalog entry %s/%s has unknown feature metadata", item.Feature, item.Case)
		}
		if strings.TrimSpace(item.DocumentZH) == "" || strings.TrimSpace(item.DocumentEN) == "" ||
			strings.TrimSpace(item.HeadingZH) == "" || strings.TrimSpace(item.HeadingEN) == "" {
			return fmt.Errorf("catalog entry %s/%s has incomplete documentation mapping", item.Feature, item.Case)
		}
		if !validLevel(item.Level) {
			return fmt.Errorf("catalog entry %s/%s has unknown level %q", item.Feature, item.Case, item.Level)
		}
		if !validStatus(item.Status) {
			return fmt.Errorf("catalog entry %s/%s has unknown status %q", item.Feature, item.Case, item.Status)
		}
		key := item.Feature + "\x00" + item.Case
		if _, exists := seen[key]; exists {
			return fmt.Errorf("catalog contains duplicate feature/case %s/%s", item.Feature, item.Case)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validLevel(level Level) bool {
	return level == LevelCompile || level == LevelHermetic || level == LevelScenario || level == LevelIntegration
}

func validStatus(status Status) bool {
	return status == StatusImplemented || status == StatusPlanned || status == StatusManual
}
