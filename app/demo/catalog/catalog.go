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
// intentionally visible: demo:list is both the discovery surface and gap report.
var Entries = []Entry{
	entry("cache", "Cache operations", "缓存操作", "Cache Usage", "demo:cache list", "list", "TestCacheOperations", LevelHermetic, StatusPlanned),
	entry("commands", "Available commands", "查看命令", "Listing Commands", "demo:list", "list", "TestDemoListCommand", LevelScenario, StatusImplemented),
	entry("config", "Quick start", "快速开始", "Quick Start", "demo:config list", "list", "TestConfigDemo", LevelHermetic, StatusPlanned),
	entry("console", "Defining commands", "定义命令", "Defining Commands", "demo:console list", "list", "TestConsoleDemo", LevelScenario, StatusPlanned),
	entry("container", "Binding", "绑定", "Binding", "demo:container list", "list", "TestContainerDemo", LevelHermetic, StatusPlanned),
	entry("cookie", "Creating cookies", "创建 Cookie", "Creating Cookies", "demo:cookie list", "list", "TestCookieDemo", LevelHermetic, StatusPlanned),
	entry("database", "Database connections", "获取数据库连接", "Obtaining a Database Connection", "demo:database list", "list", "TestDatabaseDemo", LevelHermetic, StatusPlanned),
	entry("encryption", "Key rotation", "轮换密钥", "Rotating Keys", "demo:encryption list", "list", "TestEncryptionDemo", LevelHermetic, StatusPlanned),
	entry("event", "Listeners", "注册事件与监听器", "Registering Events and Listeners", "demo:event list", "list", "TestEventDemo", LevelHermetic, StatusPlanned),
	entry("exception", "Exception handling", "异常处理", "Handling Exceptions", "demo:exception list", "list", "TestExceptionDemo", LevelScenario, StatusPlanned),
	entry("facade", "Available facades", "可用 Facade", "Available Facades", "demo:facade list", "list", "TestFacadeDemo", LevelCompile, StatusPlanned),
	entry("filesystem", "Disk access", "获取磁盘实例", "Obtaining Disk Instances", "demo:filesystem list", "list", "TestFilesystemDemo", LevelHermetic, StatusPlanned),
	entry("horizon", "Running Horizon", "运行 Horizon", "Running Horizon", "demo:horizon list", "list", "TestHorizonWithRealQueue", LevelIntegration, StatusPlanned, "redis", "rabbitmq"),
	entry("http-server", "Starting the server", "启动服务", "Starting the Server", "demo:http-server list", "list", "TestHTTPServerDemo", LevelScenario, StatusPlanned),
	manualEntry("installation", "Creating an application", "创建应用", "Creating an Application", "installer smoke test", "TestInstallerSmoke", LevelIntegration, "requires a clean temporary checkout and network access"),
	manualEntry("lens", "Installation", "安装", "Installation", "Lens workflow verification", "TestLensWorkflow", LevelIntegration, "validated by the Lens toolchain rather than an application command"),
	entry("lifecycle", "Lifecycle overview", "生命周期概览", "Lifecycle Overview", "demo:lifecycle list", "list", "TestLifecycleDemo", LevelHermetic, StatusPlanned),
	entry("logger", "Writing log messages", "编写日志消息", "Writing Log Messages", "demo:logger list", "list", "TestLoggerDemo", LevelHermetic, StatusPlanned),
	entry("queue", "Creating jobs", "创建任务", "Creating Jobs", "demo:queue basic --connection=sync", "basic-sync", "TestBasicScenarioDispatchesAndProcessesSyncJob", LevelHermetic, StatusImplemented),
	entry("queue", "Creating jobs", "创建任务", "Creating Jobs", "demo:queue basic --connection=redis", "basic-redis", "TestQueueDemoBasicWithRealRedis", LevelIntegration, StatusImplemented, "redis"),
	entry("queue", "Creating jobs", "创建任务", "Creating Jobs", "demo:queue basic --connection=rabbitmq", "basic-rabbitmq", "TestQueueDemoBasicWithRealRabbitMQ", LevelIntegration, StatusImplemented, "rabbitmq"),
	entry("queue", "Job strategies", "任务策略接口", "Job Strategy Interfaces", "demo:queue strategies", "strategies", "TestQueueDemoStrategies", LevelHermetic, StatusImplemented),
	entry("queue", "Unique jobs", "唯一任务", "Unique Jobs", "demo:queue unique", "unique", "TestQueueDemoUnique", LevelHermetic, StatusImplemented),
	entry("queue", "Debounced jobs", "防抖任务", "Debounced Jobs", "demo:queue debounce --connection=redis", "debounce", "TestQueueDemoDebounce", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	entry("queue", "Job middleware", "任务中间件", "Job Middleware", "demo:queue middleware", "middleware", "TestQueueDemoMiddleware", LevelHermetic, StatusImplemented),
	entry("queue", "Dispatching jobs", "分发任务", "Dispatching Jobs", "demo:queue dispatch --connection=redis", "dispatch", "TestQueueDemoDispatch", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	entry("queue", "Job chaining", "任务链", "Job Chaining", "demo:queue chain --connection=redis", "chain", "TestQueueDemoChain", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	entry("queue", "Job batching", "任务批处理", "Job Batching", "demo:queue batch --connection=redis", "batch", "TestQueueDemoBatch", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	entry("queue", "Running workers", "运行队列 Worker", "Running The Queue Worker", "demo:queue worker --connection=redis", "worker", "TestQueueDemoWorker", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	entry("queue", "Failed jobs", "处理失败任务", "Dealing With Failed Jobs", "demo:queue failure", "failure", "TestQueueDemoFailure", LevelIntegration, StatusPlanned, "redis", "rabbitmq"),
	entry("queue", "Failed job commands", "清理失败任务", "Dealing With Failed Jobs", "demo:queue failed-commands", "failed-commands", "TestQueueDemoFailedCommands", LevelScenario, StatusPlanned),
	entry("queue", "Worker restart", "Worker 与部署", "Queue Workers and Deployment", "demo:queue restart", "restart", "TestQueueDemoRestart", LevelIntegration, StatusPlanned, "redis", "rabbitmq"),
	entry("queue", "Lifecycle events", "生命周期事件", "Lifecycle Events", "demo:queue events", "events", "TestQueueDemoEvents", LevelIntegration, StatusPlanned, "redis", "rabbitmq"),
	entry("queue", "Encrypted payloads", "加密 Payload", "Encrypted Payloads", "demo:queue encryption", "encryption", "TestQueueDemoEncryption", LevelHermetic, StatusPlanned),
	entry("queue", "Custom drivers", "自定义驱动", "Custom Drivers", "demo:queue custom-driver", "custom-driver", "TestQueueDemoCustomDriver", LevelHermetic, StatusPlanned),
	entry("queue", "Error constants", "错误常量", "Error Constants", "demo:queue errors", "errors", "TestQueueDemoErrors", LevelHermetic, StatusPlanned),
	entry("queue", "Redis connection", "Redis 连接", "Redis Connection", "demo:queue redis --connection=redis", "redis", "TestQueueDemoRedisBoundaries", LevelIntegration, StatusPlanned, "redis"),
	entry("queue", "RabbitMQ configuration", "RabbitMQ 配置", "RabbitMQ Configuration", "demo:queue rabbitmq --connection=rabbitmq", "rabbitmq", "TestQueueDemoRabbitMQBoundaries", LevelIntegration, StatusPlanned, "rabbitmq"),
	entry("ratelimit", "Feature overview", "功能概览", "Feature Overview", "demo:ratelimit list", "list", "TestRateLimitDemo", LevelScenario, StatusPlanned),
	entry("redis", "Interacting with Redis", "与 Redis 交互", "Interacting With Redis", "demo:redis list", "list", "TestRedisWithRealServer", LevelIntegration, StatusPlanned, "redis"),
	entry("route", "Quick start", "快速开始", "Quick Start", "demo:route list", "list", "TestRouteDemo", LevelScenario, StatusPlanned),
	entry("schema", "Builder entry points", "Builder 入口", "Builder Entry Points", "demo:schema list", "list", "TestSchemaDemo", LevelHermetic, StatusPlanned),
	entry("service-provider", "Writing providers", "编写服务提供者", "Writing Service Providers", "demo:provider list", "list", "TestProviderDemo", LevelHermetic, StatusPlanned),
	entry("session", "Using sessions", "Session 使用", "Session Usage", "demo:session list", "list", "TestSessionDemo", LevelScenario, StatusPlanned),
	manualEntry("starter", "First endpoint", "添加第一个接口", "Add Your First Endpoint", "starter HTTP smoke test", "TestStarterSmoke", LevelScenario, "runs against a clean generated starter"),
	entry("support", "Feature overview", "功能概览", "Feature Overview", "demo:support list", "list", "TestSupportDemo", LevelHermetic, StatusPlanned),
	entry("timer", "Defining tasks", "定义调度任务", "Defining Scheduled Tasks", "demo:timer list", "list", "TestTimerDemo", LevelHermetic, StatusPlanned),
	entry("translation", "Overview", "概述", "Overview", "demo:translation list", "list", "TestTranslationDemo", LevelHermetic, StatusPlanned),
}

func entry(feature, section, headingZH, headingEN, example, caseName, test string, level Level, status Status, requirements ...string) Entry {
	return Entry{
		Feature: feature, DocumentZH: "zh_CN/" + feature + ".md", DocumentEN: "en/" + feature + ".md",
		Section: section, HeadingZH: headingZH, HeadingEN: headingEN, Example: example, Case: caseName,
		Test: test, Level: level, Status: status, Requirements: requirements,
	}
}

func manualEntry(feature, section, headingZH, headingEN, example, test string, level Level, notes string) Entry {
	item := entry(feature, section, headingZH, headingEN, example, "manual", test, level, StatusManual)
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
			strings.TrimSpace(item.Example) == "" || strings.TrimSpace(item.Test) == "" || strings.TrimSpace(item.Case) == "" {
			return fmt.Errorf("catalog entry %d has an empty required field", index)
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
