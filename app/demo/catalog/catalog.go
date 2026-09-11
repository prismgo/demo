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
	baselineEntry("cache", "Manager and repository architecture", "简介", "Introduction", "demo:cache architecture", "architecture", "TestCacheDemoArchitecture", LevelCompile, StatusPlanned),
	baselineEntry("cache", "Configuration loading", "配置文件", "Config File", "demo:cache config", "config", "TestCacheDemoConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Driver prerequisites", "驱动前置条件", "Driver Prerequisites", "demo:cache driver-prerequisites", "driver-prerequisites", "TestCacheDemoDriverPrerequisites", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Top-level configuration", "顶层配置", "Top-Level Configuration", "demo:cache top-level-config", "top-level-config", "TestCacheDemoTopLevelConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Memory store configuration", "Memory Store", "Memory Store", "demo:cache memory-config", "memory-config", "TestCacheDemoMemoryConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Redis store configuration", "Redis Store", "Redis Store", "demo:cache redis-config", "redis-config", "TestCacheDemoRedisConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "File store configuration", "File Store", "File Store", "demo:cache file-config", "file-config", "TestCacheDemoFileConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Failover store configuration", "Failover Store", "Failover Store", "demo:cache failover-config", "failover-config", "TestCacheDemoFailoverConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Lock configuration", "锁配置", "Lock Configuration", "demo:cache lock-config", "lock-config", "TestCacheDemoLockConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Flexible refresh configuration", "热点刷新配置", "Hot Key Refresh Configuration", "demo:cache flexible-config", "flexible-config", "TestCacheDemoFlexibleConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Package-level facade", "使用包级 Facade", "Using the Package-Level Facade", "demo:cache facade", "facade", "TestCacheDemoFacade", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Named store operations", "使用指定 Store", "Using a Specific Store", "demo:cache named-store", "named-store", "TestCacheDemoNamedStore", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Repository dependency injection", "获取 Repository 实例", "Obtaining a Repository Instance", "demo:cache repository", "repository", "TestCacheDemoRepository", LevelCompile, StatusPlanned),
	baselineEntry("cache", "Missing named store errors", "获取缓存实例", "Obtaining a Cache Instance", "demo:cache missing-store", "missing-store", "TestCacheDemoMissingStore", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Generic retrieval and cache misses", "基础读取", "Basic Retrieval", "demo:cache get", "get", "TestCacheDemoGet", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Value and lazy fallbacks", "带默认值读取", "Retrieval with Default Value", "demo:cache fallbacks", "fallbacks", "TestCacheDemoFallbacks", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Typed convenience retrieval", "便捷类型读取", "Typed Convenience Methods", "demo:cache typed-retrieval", "typed-retrieval", "TestCacheDemoTypedRetrieval", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Cache key existence", "判断存在性", "Checking Existence", "demo:cache existence", "existence", "TestCacheDemoExistence", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Put and Set storage", "基础写入", "Basic Storage", "demo:cache put", "put", "TestCacheDemoPutAndSet", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Permanent storage", "永久写入", "Permanent Storage", "demo:cache forever", "forever", "TestCacheDemoForever", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Atomic conditional storage", "条件写入（仅当 key 不存在）", "Conditional Storage (Write Only If Key Does Not Exist)", "demo:cache add", "add", "TestCacheDemoAdd", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Bulk storage aliases", "批量写入", "Bulk Storage", "demo:cache put-many", "put-many", "TestCacheDemoPutMany", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Retrieve and store", "读穿缓存（Remember）", "Retrieve and Store", "demo:cache remember", "remember", "TestCacheDemoRemember", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Permanent retrieve and store aliases", "永久读穿", "Remember Forever", "demo:cache remember-forever", "remember-forever", "TestCacheDemoRememberForever", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Stale-while-revalidate windows", "Stale-While-Revalidate（Flexible）", "Stale-While-Revalidate (Flexible)", "demo:cache flexible", "flexible", "TestCacheDemoFlexible", LevelScenario, StatusPlanned),
	baselineEntry("cache", "Touch TTL semantics", "延长 TTL", "Touch TTL", "demo:cache touch", "touch", "TestCacheDemoTouch", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Bulk retrieval aliases", "批量读取", "Retrieving Multiple Items", "demo:cache many", "many", "TestCacheDemoMany", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Atomic retrieve and delete", "取后删除", "Retrieve and Delete", "demo:cache pull", "pull", "TestCacheDemoPull", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Single-key removal aliases", "删除单个 key", "Removing a Single Key", "demo:cache forget", "forget", "TestCacheDemoForget", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Bulk removal", "批量删除", "Bulk Removal", "demo:cache forget-many", "forget-many", "TestCacheDemoForgetMany", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Store flush aliases and prefix safety", "清空整个 store", "Clearing an Entire Store", "demo:cache flush", "flush", "TestCacheDemoFlush", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Atomic integer counters", "计数器", "Counters", "demo:cache counters", "counters", "TestCacheDemoCounters", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Lock acquisition and release", "基础用法", "Basic Usage", "demo:cache lock", "lock", "TestCacheDemoLock", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Lock callback auto-release", "回调自动释放", "Automatic Release via Callback", "demo:cache lock-callback", "lock-callback", "TestCacheDemoLockCallback", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Blocking lock wait and timeout", "阻塞等待", "Blocking Wait", "demo:cache lock-block", "lock-block", "TestCacheDemoLockBlock", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Lock owner restoration and force release", "跨流程释放", "Cross-Process Release", "demo:cache lock-restore", "lock-restore", "TestCacheDemoLockRestore", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Lock namespace cleanup", "批量清理锁", "Bulk Lock Cleanup", "demo:cache lock-flush", "lock-flush", "TestCacheDemoLockFlush", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Funnel concurrency limiting", "并发限制器（Funnel）", "Concurrency Limiter (Funnel)", "demo:cache funnel", "funnel", "TestCacheDemoFunnel", LevelScenario, StatusPlanned),
	baselineEntry("cache", "Overlapping task prevention", "防止任务重叠（WithoutOverlapping）", "Preventing Task Overlap (WithoutOverlapping)", "demo:cache without-overlapping", "without-overlapping", "TestCacheDemoWithoutOverlapping", LevelScenario, StatusPlanned),
	baselineEntry("cache", "Memory tagged cache", "Tagged Cache", "Tagged Cache", "demo:cache tags-memory", "tags-memory", "TestCacheDemoMemoryTags", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Redis tagged cache", "Tagged Cache", "Tagged Cache", "demo:cache tags-redis", "tags-redis", "TestCacheDemoRedisTags", LevelIntegration, StatusPlanned, "redis"),
	baselineEntry("cache", "Unsupported tagged cache", "Tagged Cache", "Tagged Cache", "demo:cache tags-unsupported", "tags-unsupported", "TestCacheDemoUnsupportedTags", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Request-scoped memoization", "Memoization", "Memoization", "demo:cache memo", "memo", "TestCacheDemoMemo", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Failover behavior and events", "Failover", "Failover", "demo:cache failover", "failover", "TestCacheDemoFailover", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Custom driver registration and capabilities", "自定义 Driver", "Custom Drivers", "demo:cache custom-driver", "custom-driver", "TestCacheDemoCustomDriver", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Manager resource lifecycle", "资源生命周期", "Resource Lifecycle", "demo:cache resource-lifecycle", "resource-lifecycle", "TestCacheDemoResourceLifecycle", LevelScenario, StatusPlanned),
	baselineEntry("cache", "Lifecycle event dispatch", "Cache Events", "Cache Events", "demo:cache events", "events", "TestCacheDemoEvents", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Cache event payload contract", "事件列表", "Event List", "demo:cache event-contract", "event-contract", "TestCacheDemoEventContract", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Deferred refresh context", "Deferred Context", "Deferred Context", "demo:cache deferred", "deferred", "TestCacheDemoDeferred", LevelScenario, StatusPlanned),
	baselineEntry("cache", "Cache and lock key prefixes", "Key 前缀", "Key Prefixes", "demo:cache key-prefixes", "key-prefixes", "TestCacheDemoKeyPrefixes", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Payload encoding conventions", "数据编码约定", "Data Encoding Conventions", "demo:cache encoding", "encoding", "TestCacheDemoEncoding", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Error constants", "错误常量", "Error Constants", "demo:cache errors", "errors", "TestCacheDemoErrors", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Memory store capability matrix", "内置 Store 能力矩阵", "Built-in Store Capability Matrix", "demo:cache memory-capabilities", "memory-capabilities", "TestCacheDemoMemoryCapabilities", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "File store capability matrix", "内置 Store 能力矩阵", "Built-in Store Capability Matrix", "demo:cache file-capabilities", "file-capabilities", "TestCacheDemoFileCapabilities", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Redis store capability matrix", "内置 Store 能力矩阵", "Built-in Store Capability Matrix", "demo:cache redis-capabilities", "redis-capabilities", "TestCacheDemoRedisCapabilities", LevelIntegration, StatusPlanned, "redis"),
	baselineEntry("cache", "Failover store capability matrix", "内置 Store 能力矩阵", "Built-in Store Capability Matrix", "demo:cache failover-capabilities", "failover-capabilities", "TestCacheDemoFailoverCapabilities", LevelHermetic, StatusPlanned),
	baselineEntry("cache", "Laravel cache API mapping", "与 Laravel Cache 的对应关系", "Laravel Cache Mapping", "demo:cache laravel-compatibility", "laravel-compatibility", "TestCacheDemoLaravelCompatibility", LevelCompile, StatusPlanned),
	baselineEntry("commands", "Available commands", "查看命令", "Listing Commands", "demo:list", "list", "TestDemoListCommand", LevelScenario, StatusImplemented),
	baselineEntry("config", "Quick start", "快速开始", "Quick Start", "demo:config list", "list", "TestConfigDemo", LevelHermetic, StatusPlanned),
	baselineEntry("console", "Defining commands", "定义命令", "Defining Commands", "demo:console list", "list", "TestConsoleDemo", LevelScenario, StatusPlanned),
	baselineEntry("container", "Binding", "绑定", "Binding", "demo:container list", "list", "TestContainerDemo", LevelHermetic, StatusPlanned),
	baselineEntry("cookie", "Creating cookies", "创建 Cookie", "Creating Cookies", "demo:cookie list", "list", "TestCookieDemo", LevelHermetic, StatusPlanned),
	baselineEntry("database", "Database connections", "获取数据库连接", "Obtaining a Database Connection", "demo:database list", "list", "TestDatabaseDemo", LevelHermetic, StatusPlanned),
	baselineEntry("encryption", "Key rotation", "轮换密钥", "Rotating Keys", "demo:encryption list", "list", "TestEncryptionDemo", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Event bus architecture", "简介", "Introduction", "demo:event architecture", "architecture", "TestEventDemoArchitecture", LevelCompile, StatusPlanned),
	baselineEntry("event", "Explicit listener registration", "手动注册监听器", "Manually Registering Listeners", "demo:event manual-registration", "manual-registration", "TestEventDemoManualRegistration", LevelCompile, StatusPlanned),
	baselineEntry("event", "Struct listener registration", "手动注册监听器", "Manually Registering Listeners", "demo:event struct-registration", "struct-registration", "TestEventDemoStructRegistration", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Closure listener registration", "闭包监听器", "Closure Listeners", "demo:event closure-listener", "closure-listener", "TestEventDemoClosureListener", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Prefix wildcard listeners", "通配符监听器", "Wildcard Event Listeners", "demo:event wildcard-prefix", "wildcard-prefix", "TestEventDemoPrefixWildcard", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Catch-all wildcard listeners", "通配符监听器", "Wildcard Event Listeners", "demo:event wildcard-all", "wildcard-all", "TestEventDemoCatchAllWildcard", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Exact-only Forget and Has semantics", "通配符监听器", "Wildcard Event Listeners", "demo:event wildcard-exact-operations", "wildcard-exact-operations", "TestEventDemoWildcardExactOperations", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Event definitions", "定义事件", "Defining Events", "demo:event event-definition", "event-definition", "TestEventDemoEventDefinition", LevelCompile, StatusPlanned),
	baselineEntry("event", "Event naming conventions", "定义事件", "Defining Events", "demo:event event-naming", "event-naming", "TestEventDemoEventNaming", LevelCompile, StatusPlanned),
	baselineEntry("event", "Serializable event payload boundaries", "定义事件", "Defining Events", "demo:event payload-boundaries", "payload-boundaries", "TestEventDemoPayloadBoundaries", LevelCompile, StatusPlanned),
	baselineEntry("event", "Struct listeners", "结构体监听器", "Struct Listeners", "demo:event struct-listener", "struct-listener", "TestEventDemoStructListener", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Function listeners", "函数式监听器", "Function Listeners", "demo:event listener-func", "listener-func", "TestEventDemoListenerFunc", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Non-stoppable event propagation", "停止事件传播", "Stopping The Propagation Of An Event", "demo:event propagation", "propagation", "TestEventDemoPropagation", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Queued listener prerequisites", "Queued Event Listeners", "Queued Event Listeners", "demo:event queue-prerequisites", "queue-prerequisites", "TestEventDemoQueuePrerequisites", LevelCompile, StatusPlanned),
	baselineEntry("event", "ShouldQueue listener dispatch", "实现 ShouldQueue 接口", "Implementing ShouldQueue", "demo:event should-queue", "should-queue", "TestEventDemoShouldQueue", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Queued listener wrapper", "使用 Queued 包装器", "Using the Queued Wrapper", "demo:event queued-wrapper", "queued-wrapper", "TestEventDemoQueuedWrapper", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Synchronous queued listener execution", "使用 Queued 包装器", "Using the Queued Wrapper", "demo:event queued-sync", "queued-sync", "TestEventDemoQueuedSync", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Redis worker queued listener execution", "使用 Queued 包装器", "Using the Queued Wrapper", "demo:event queued-redis --connection=redis", "queued-redis", "TestEventDemoQueuedRedis", LevelIntegration, StatusPlanned, "redis"),
	baselineEntry("event", "Queued listener connection, queue, and delay", "自定义队列连接、队列名与延迟", "Customizing the Queue Connection, Queue Name, & Delay", "demo:event queue-routing-options", "queue-routing-options", "TestEventDemoQueueRoutingOptions", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Queued listener retry, backoff, and timeout", "自定义队列连接、队列名与延迟", "Customizing the Queue Connection, Queue Name, & Delay", "demo:event queue-retry-options", "queue-retry-options", "TestEventDemoQueueRetryOptions", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Queued event factory registration", "注册事件工厂", "Registering Event Factories", "demo:event event-factory", "event-factory", "TestEventDemoEventFactory", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Queued event factory validation", "注册事件工厂", "Registering Event Factories", "demo:event event-factory-validation", "event-factory-validation", "TestEventDemoEventFactoryValidation", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Unregistered queued event fallback", "注册事件工厂", "Registering Event Factories", "demo:event raw-queued-event", "raw-queued-event", "TestEventDemoRawQueuedEvent", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Event dispatch", "派发事件", "Dispatching Events", "demo:event dispatch", "dispatch", "TestEventDemoDispatch", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Dispatch failure isolation", "派发事件", "Dispatching Events", "demo:event dispatch-isolation", "dispatch-isolation", "TestEventDemoDispatchIsolation", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Nil context and event dispatch", "派发事件", "Dispatching Events", "demo:event nil-dispatch", "nil-dispatch", "TestEventDemoNilDispatch", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Event subscriber implementation", "编写 Event Subscribers", "Writing Event Subscribers", "demo:event subscriber", "subscriber", "TestEventDemoSubscriber", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Event subscriber registration", "注册 Event Subscribers", "Registering Event Subscribers", "demo:event subscriber-registration", "subscriber-registration", "TestEventDemoSubscriberRegistration", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Package-level facade operations", "全局门面", "Global Facade", "demo:event facade", "facade", "TestEventDemoFacade", LevelScenario, StatusPlanned),
	baselineEntry("event", "Goroutine-based async listeners", "异步监听器", "Async Listeners", "demo:event async", "async", "TestEventDemoAsync", LevelScenario, StatusPlanned),
	baselineEntry("event", "Async listener durability boundaries", "异步监听器", "Async Listeners", "demo:event async-durability", "async-durability", "TestEventDemoAsyncDurability", LevelCompile, StatusPlanned),
	baselineEntry("event", "Application lifecycle events", "应用生命周期", "Application Lifecycle", "demo:event app-lifecycle", "app-lifecycle", "TestEventDemoApplicationLifecycle", LevelScenario, StatusPlanned),
	baselineEntry("event", "Provider lifecycle events", "Provider 生命周期", "Provider Lifecycle", "demo:event provider-lifecycle", "provider-lifecycle", "TestEventDemoProviderLifecycle", LevelScenario, StatusPlanned),
	baselineEntry("event", "HTTP server lifecycle events", "HTTP 服务生命周期", "HTTP Server Lifecycle", "demo:event server-lifecycle", "server-lifecycle", "TestEventDemoServerLifecycle", LevelScenario, StatusPlanned),
	baselineEntry("event", "HTTP request lifecycle events", "HTTP 请求生命周期", "HTTP Request Lifecycle", "demo:event request-lifecycle", "request-lifecycle", "TestEventDemoRequestLifecycle", LevelScenario, StatusPlanned),
	baselineEntry("event", "HTTP request finished ordering", "HTTP 请求生命周期", "HTTP Request Lifecycle", "demo:event request-finished-ordering", "request-finished-ordering", "TestEventDemoRequestFinishedOrdering", LevelScenario, StatusPlanned),
	baselineEntry("event", "Console lifecycle events", "Console 生命周期", "Console Lifecycle", "demo:event console-lifecycle", "console-lifecycle", "TestEventDemoConsoleLifecycle", LevelScenario, StatusPlanned),
	baselineEntry("event", "Vendor publish lifecycle event", "Vendor Publish 事件", "Vendor Publish Event", "demo:event vendor-publish-event", "vendor-publish-event", "TestEventDemoVendorPublishEvent", LevelScenario, StatusPlanned),
	baselineEntry("event", "Synchronous listener error isolation", "错误与 Panic 处理", "Error and Panic Handling", "demo:event listener-error-isolation", "listener-error-isolation", "TestEventDemoListenerErrorIsolation", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Synchronous listener panic isolation", "错误与 Panic 处理", "Error and Panic Handling", "demo:event listener-panic-isolation", "listener-panic-isolation", "TestEventDemoListenerPanicIsolation", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Async listener failure isolation", "错误与 Panic 处理", "Error and Panic Handling", "demo:event async-failure-isolation", "async-failure-isolation", "TestEventDemoAsyncFailureIsolation", LevelScenario, StatusPlanned),
	baselineEntry("event", "Queued listener failure handling", "错误与 Panic 处理", "Error and Panic Handling", "demo:event queued-failure-handling", "queued-failure-handling", "TestEventDemoQueuedFailureHandling", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Strong-consistency boundary", "错误与 Panic 处理", "Error and Panic Handling", "demo:event consistency-boundary", "consistency-boundary", "TestEventDemoConsistencyBoundary", LevelCompile, StatusPlanned),
	baselineEntry("event", "Isolated dispatcher testing", "测试", "Testing", "demo:event isolated-testing", "isolated-testing", "TestEventDemoIsolatedTesting", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Synchronous queue listener testing", "测试", "Testing", "demo:event queued-sync-testing", "queued-sync-testing", "TestEventDemoQueuedSyncTesting", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Real worker listener testing", "测试", "Testing", "demo:event queued-worker-testing --connection=redis", "queued-worker-testing", "TestEventDemoQueuedWorkerTesting", LevelIntegration, StatusPlanned, "redis"),
	baselineEntry("event", "Event interface contract", "Event", "Event", "demo:event event-interface", "event-interface", "TestEventDemoEventInterface", LevelCompile, StatusPlanned),
	baselineEntry("event", "Listener interface contract", "Listener", "Listener", "demo:event listener-interface", "listener-interface", "TestEventDemoListenerInterface", LevelCompile, StatusPlanned),
	baselineEntry("event", "ListenerFunc adapter contract", "ListenerFunc", "ListenerFunc", "demo:event listener-func-interface", "listener-func-interface", "TestEventDemoListenerFuncInterface", LevelCompile, StatusPlanned),
	baselineEntry("event", "Dispatcher interface contract", "Dispatcher", "Dispatcher", "demo:event dispatcher-interface", "dispatcher-interface", "TestEventDemoDispatcherInterface", LevelCompile, StatusPlanned),
	baselineEntry("event", "Subscriber interface contract", "Subscriber", "Subscriber", "demo:event subscriber-interface", "subscriber-interface", "TestEventDemoSubscriberInterface", LevelCompile, StatusPlanned),
	baselineEntry("event", "ShouldQueue interface contract", "ShouldQueue", "ShouldQueue", "demo:event should-queue-interface", "should-queue-interface", "TestEventDemoShouldQueueInterface", LevelCompile, StatusPlanned),
	baselineEntry("event", "AsyncListener interface contract", "AsyncListener", "AsyncListener", "demo:event async-listener-interface", "async-listener-interface", "TestEventDemoAsyncListenerInterface", LevelCompile, StatusPlanned),
	baselineEntry("event", "QueueOptionsProvider interface contract", "QueueOptionsProvider", "QueueOptionsProvider", "demo:event queue-options-interface", "queue-options-interface", "TestEventDemoQueueOptionsInterface", LevelCompile, StatusPlanned),
	baselineEntry("event", "Event ServiceProvider registration", "ServiceProvider", "ServiceProvider", "demo:event provider-register", "provider-register", "TestEventDemoProviderRegister", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Event ServiceProvider queue bootstrapping", "ServiceProvider", "ServiceProvider", "demo:event provider-boot", "provider-boot", "TestEventDemoProviderBoot", LevelHermetic, StatusPlanned),
	baselineEntry("event", "Laravel events compatibility matrix", "与 Laravel 13 的差异", "Laravel 13 Differences", "demo:event laravel-compatibility", "laravel-compatibility", "TestEventDemoLaravelCompatibility", LevelCompile, StatusPlanned),
	baselineEntry("exception", "Exception handling", "异常处理", "Handling Exceptions", "demo:exception list", "list", "TestExceptionDemo", LevelScenario, StatusPlanned),
	baselineEntry("facade", "Available facades", "可用 Facade", "Available Facades", "demo:facade list", "list", "TestFacadeDemo", LevelCompile, StatusPlanned),
	baselineEntry("filesystem", "Manager and repository architecture", "简介", "Introduction", "demo:filesystem architecture", "architecture", "TestFilesystemDemoArchitecture", LevelCompile, StatusPlanned),
	baselineEntry("filesystem", "Configuration loading", "配置文件", "Configuration File", "demo:filesystem config", "config", "TestFilesystemDemoConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Local driver prerequisites", "驱动前置条件", "Driver Prerequisites", "demo:filesystem local-prerequisites", "local-prerequisites", "TestFilesystemDemoLocalPrerequisites", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "OSS driver prerequisites", "驱动前置条件", "Driver Prerequisites", "demo:filesystem oss-prerequisites --disk=oss", "oss-prerequisites", "TestFilesystemDemoOSSPrerequisites", LevelIntegration, StatusPlanned, "oss"),
	baselineEntry("filesystem", "Top-level configuration", "顶层配置", "Top-Level Configuration", "demo:filesystem top-level-config", "top-level-config", "TestFilesystemDemoTopLevelConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Local disk configuration", "Local 磁盘", "Local Disk", "demo:filesystem local-config", "local-config", "TestFilesystemDemoLocalConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Public disk configuration", "Public 磁盘", "Public Disk", "demo:filesystem public-config", "public-config", "TestFilesystemDemoPublicConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "OSS disk configuration", "OSS 磁盘", "OSS Disk", "demo:filesystem oss-config --disk=oss", "oss-config", "TestFilesystemDemoOSSConfiguration", LevelIntegration, StatusPlanned, "oss"),
	baselineEntry("filesystem", "Symbolic link configuration", "符号链接配置", "Symbolic Link Configuration", "demo:filesystem links-config", "links-config", "TestFilesystemDemoLinksConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Disk configuration fields", "磁盘配置字段说明", "Disk Configuration Fields", "demo:filesystem disk-config", "disk-config", "TestFilesystemDemoDiskConfiguration", LevelCompile, StatusPlanned),
	baselineEntry("filesystem", "Package-level facade", "包级 Facade", "Package-Level Facade", "demo:filesystem facade", "facade", "TestFilesystemDemoFacade", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Named disk operations", "指定磁盘", "Named Disks", "demo:filesystem named-disk", "named-disk", "TestFilesystemDemoNamedDisk", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Cloud disk selection", "云磁盘", "Cloud Disk", "demo:filesystem cloud", "cloud", "TestFilesystemDemoCloudDisk", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Interface selection guidance", "接口选择建议", "Interface Selection Guide", "demo:filesystem interface-selection", "interface-selection", "TestFilesystemDemoInterfaceSelection", LevelCompile, StatusPlanned),
	baselineEntry("filesystem", "Whole-file reads", "读取全部内容", "Reading File Contents", "demo:filesystem get", "get", "TestFilesystemDemoGet", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "JSON deserialization", "读取并反序列化 JSON", "Reading and Deserializing JSON", "demo:filesystem json", "json", "TestFilesystemDemoJSON", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Streaming reads with metadata", "流式读取", "Streaming Reads", "demo:filesystem open-stream", "open-stream", "TestFilesystemDemoOpenStream", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Streaming read alias", "流式读取", "Streaming Reads", "demo:filesystem read-stream", "read-stream", "TestFilesystemDemoReadStream", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Downloads to writers", "下载到 Writer", "Downloading to a Writer", "demo:filesystem download", "download", "TestFilesystemDemoDownload", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "File existence aliases", "判断文件存在", "Determining File Existence", "demo:filesystem file-existence", "file-existence", "TestFilesystemDemoFileExistence", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Directory existence", "判断目录存在", "Determining Directory Existence", "demo:filesystem directory-existence", "directory-existence", "TestFilesystemDemoDirectoryExistence", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "String byte and reader writes", "基础写入", "Basic Writes", "demo:filesystem put", "put", "TestFilesystemDemoPut", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Streaming write aliases", "流式写入", "Streaming Writes", "demo:filesystem put-reader", "put-reader", "TestFilesystemDemoPutReader", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Prepend and append semantics", "前置追加与末尾追加", "Prepending and Appending", "demo:filesystem prepend-append", "prepend-append", "TestFilesystemDemoPrependAppend", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Write visibility and content type options", "PutOptions 说明", "PutOptions Reference", "demo:filesystem put-options", "put-options", "TestFilesystemDemoPutOptions", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Uploads with original names", "PutFile", "PutFile", "demo:filesystem put-file", "put-file", "TestFilesystemDemoPutFile", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Uploads with specified names", "PutFileAs", "PutFileAs", "demo:filesystem put-file-as", "put-file-as", "TestFilesystemDemoPutFileAs", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Upload persistence fields", "推荐业务字段", "Recommended Database Fields", "demo:filesystem upload-fields", "upload-fields", "TestFilesystemDemoUploadFields", LevelCompile, StatusPlanned),
	baselineEntry("filesystem", "Copy and move operations", "复制与移动", "Copying and Moving", "demo:filesystem copy-move", "copy-move", "TestFilesystemDemoCopyMove", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Cross-disk operation guard", "复制与移动", "Copying and Moving", "demo:filesystem cross-disk-guard", "cross-disk-guard", "TestFilesystemDemoCrossDiskGuard", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Multi-file deletion", "删除文件", "Deleting Files", "demo:filesystem delete", "delete", "TestFilesystemDemoDelete", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "File size metadata", "文件元信息", "File Metadata", "demo:filesystem size", "size", "TestFilesystemDemoSize", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Last-modified metadata", "文件元信息", "File Metadata", "demo:filesystem last-modified", "last-modified", "TestFilesystemDemoLastModified", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Complete file metadata", "文件元信息", "File Metadata", "demo:filesystem file-info", "file-info", "TestFilesystemDemoFileInfo", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "MIME type detection", "文件元信息", "File Metadata", "demo:filesystem mime-type", "mime-type", "TestFilesystemDemoMimeType", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "SHA-256 checksums", "文件校验和", "File Checksums", "demo:filesystem checksum", "checksum", "TestFilesystemDemoChecksum", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Physical and logical paths", "物理路径", "Physical Paths", "demo:filesystem path", "path", "TestFilesystemDemoPath", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Directory creation", "目录操作", "Directories", "demo:filesystem make-directory", "make-directory", "TestFilesystemDemoMakeDirectory", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Shallow file listing", "目录操作", "Directories", "demo:filesystem files", "files", "TestFilesystemDemoFiles", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Recursive file listing", "目录操作", "Directories", "demo:filesystem all-files", "all-files", "TestFilesystemDemoAllFiles", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Shallow directory listing", "目录操作", "Directories", "demo:filesystem directories", "directories", "TestFilesystemDemoDirectories", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Recursive directory listing", "目录操作", "Directories", "demo:filesystem all-directories", "all-directories", "TestFilesystemDemoAllDirectories", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Safe directory deletion", "目录操作", "Directories", "demo:filesystem delete-directory", "delete-directory", "TestFilesystemDemoDeleteDirectory", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Public URL generation", "公开 URL", "Public URLs", "demo:filesystem public-url", "public-url", "TestFilesystemDemoPublicURL", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Public disk symbolic links", "公开磁盘符号链接", "Public Disk Symbolic Links", "demo:filesystem storage-link", "storage-link", "TestFilesystemDemoStorageLink", LevelScenario, StatusPlanned),
	baselineEntry("filesystem", "Relative and forced symbolic links", "公开磁盘符号链接", "Public Disk Symbolic Links", "demo:filesystem storage-link-options", "storage-link-options", "TestFilesystemDemoStorageLinkOptions", LevelScenario, StatusPlanned),
	baselineEntry("filesystem", "Safe symbolic link removal", "公开磁盘符号链接", "Public Disk Symbolic Links", "demo:filesystem storage-unlink", "storage-unlink", "TestFilesystemDemoStorageUnlink", LevelScenario, StatusPlanned),
	baselineEntry("filesystem", "Temporary signed URLs", "临时签名 URL", "Temporary Signed URLs", "demo:filesystem temporary-url", "temporary-url", "TestFilesystemDemoTemporaryURL", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Temporary URL capability detection", "临时签名 URL", "Temporary Signed URLs", "demo:filesystem temporary-url-capability", "temporary-url-capability", "TestFilesystemDemoTemporaryURLCapability", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "OSS temporary upload URLs", "临时上传 URL", "Temporary Upload URLs", "demo:filesystem temporary-upload-url --disk=oss", "temporary-upload-url", "TestFilesystemDemoTemporaryUploadURL", LevelIntegration, StatusPlanned, "oss"),
	baselineEntry("filesystem", "Temporary upload URL capability detection", "临时上传 URL", "Temporary Upload URLs", "demo:filesystem temporary-upload-capability", "temporary-upload-capability", "TestFilesystemDemoTemporaryUploadCapability", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Local signed URL verification", "校验本地签名 URL", "Verifying Local Signed URLs", "demo:filesystem verify-temporary-url", "verify-temporary-url", "TestFilesystemDemoVerifyTemporaryURL", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Local disk visibility semantics", "可见性", "Visibility", "demo:filesystem local-visibility", "local-visibility", "TestFilesystemDemoLocalVisibility", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "OSS object visibility ACLs", "可见性", "Visibility", "demo:filesystem oss-visibility --disk=oss", "oss-visibility", "TestFilesystemDemoOSSVisibility", LevelIntegration, StatusPlanned, "oss"),
	baselineEntry("filesystem", "OSS disk switching and ACL mapping", "OSS 驱动", "OSS Driver", "go test ./app/demo/filesystem -run TestFilesystemDemoOSSDriver", "oss-driver", "TestFilesystemDemoOSSDriver", LevelIntegration, StatusImplemented, "oss"),
	baselineEntry("filesystem", "Custom driver registration", "注册 Driver", "Registering a Driver", "demo:filesystem custom-driver", "custom-driver", "TestFilesystemDemoCustomDriver", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Custom driver replacement and lifecycle", "注册 Driver", "Registering a Driver", "demo:filesystem custom-driver-lifecycle", "custom-driver-lifecycle", "TestFilesystemDemoCustomDriverLifecycle", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Driver contract", "Driver 接口", "Driver Interface", "demo:filesystem driver-contract", "driver-contract", "TestFilesystemDemoDriverContract", LevelCompile, StatusPlanned),
	baselineEntry("filesystem", "Optional driver capabilities and fallbacks", "Driver 接口", "Driver Interface", "demo:filesystem optional-driver-capabilities", "optional-driver-capabilities", "TestFilesystemDemoOptionalDriverCapabilities", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Driver factory context", "DriverFactoryContext 字段说明", "DriverFactoryContext Fields", "demo:filesystem driver-factory-context", "driver-factory-context", "TestFilesystemDemoDriverFactoryContext", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Manual manager construction and cleanup", "手动初始化", "Manual Initialization", "demo:filesystem manual-manager", "manual-manager", "TestFilesystemDemoManualManager", LevelScenario, StatusPlanned),
	baselineEntry("filesystem", "Manager construction from application config", "手动初始化", "Manual Initialization", "demo:filesystem manager-from-config", "manager-from-config", "TestFilesystemDemoManagerFromConfig", LevelScenario, StatusPlanned),
	baselineEntry("filesystem", "Filesystem error constants", "错误常量", "Error Constants", "demo:filesystem errors", "errors", "TestFilesystemDemoErrors", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Local driver capability matrix", "驱动能力矩阵", "Driver Capability Matrix", "demo:filesystem local-capabilities", "local-capabilities", "TestFilesystemDemoLocalCapabilities", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "OSS driver capability matrix", "驱动能力矩阵", "Driver Capability Matrix", "demo:filesystem oss-capabilities --disk=oss", "oss-capabilities", "TestFilesystemDemoOSSCapabilities", LevelIntegration, StatusPlanned, "oss"),
	baselineEntry("filesystem", "Laravel filesystem API mapping", "与 Laravel Filesystem 的对应关系", "Laravel Filesystem Mapping", "demo:filesystem laravel-compatibility", "laravel-compatibility", "TestFilesystemDemoLaravelCompatibility", LevelCompile, StatusPlanned),
	baselineEntry("filesystem", "Filesystem usage guidance", "使用建议", "Best Practices", "demo:filesystem best-practices", "best-practices", "TestFilesystemDemoBestPractices", LevelCompile, StatusPlanned),
	baselineEntry("horizon", "Running Horizon", "运行 Horizon", "Running Horizon", "demo:horizon list", "list", "TestHorizonWithRealQueue", LevelIntegration, StatusPlanned, "redis", "rabbitmq"),
	baselineEntry("http-server", "Starting the server", "启动服务", "Starting the Server", "demo:http-server list", "list", "TestHTTPServerDemo", LevelScenario, StatusPlanned),
	manualEntry("installation", "Creating an application", "创建应用", "Creating an Application", "installer smoke test", "TestInstallerSmoke", LevelIntegration, "requires a clean temporary checkout and network access"),
	manualEntry("lens", "Installation", "安装", "Installation", "Lens workflow verification", "TestLensWorkflow", LevelIntegration, "validated by the Lens toolchain rather than an application command"),
	baselineEntry("lifecycle", "Lifecycle overview", "生命周期概览", "Lifecycle Overview", "demo:lifecycle list", "list", "TestLifecycleDemo", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Channel-based logging architecture", "简介", "Introduction", "demo:logger architecture", "architecture", "TestLoggerDemoArchitecture", LevelCompile, StatusPlanned),
	baselineEntry("logger", "Configuration loading and environment overrides", "配置文件", "Config File", "demo:logger config", "config", "TestLoggerDemoConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Single file driver and directory creation", "Single 和 Daily 通道", "Single and Daily Channels", "demo:logger single-driver", "single-driver", "TestLoggerDemoSingleDriver", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Daily file rotation and dated filenames", "Single 和 Daily 通道", "Single and Daily Channels", "demo:logger daily-driver", "daily-driver", "TestLoggerDemoDailyDriver", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Standard error driver", "Stderr 通道", "Stderr Channel", "demo:logger stderr-driver", "stderr-driver", "TestLoggerDemoStderrDriver", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Null driver output suppression", "Null 通道", "Null Channel", "demo:logger null-driver", "null-driver", "TestLoggerDemoNullDriver", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Default channel configuration", "顶层配置", "Top-Level Configuration", "demo:logger top-level-config", "top-level-config", "TestLoggerDemoTopLevelConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Channel driver, formatter, level, path, and stack options", "通道配置", "Channel Configuration", "demo:logger channel-config", "channel-config", "TestLoggerDemoChannelConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Environment-specific logging configuration", "部署建议", "Deployment Recommendations", "demo:logger deployment-config", "deployment-config", "TestLoggerDemoDeploymentConfiguration", LevelCompile, StatusPlanned),
	baselineEntry("logger", "Stack channel fan-out", "构建日志堆栈", "Building Log Stacks", "demo:logger stack", "stack", "TestLoggerDemoStack", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Stack child level and formatter isolation", "构建日志堆栈", "Building Log Stacks", "demo:logger stack-channel-isolation", "stack-channel-isolation", "TestLoggerDemoStackChannelIsolation", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Log level thresholds", "日志级别", "Log Levels", "demo:logger levels", "levels", "TestLoggerDemoLevels", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Package-level logging facade", "使用 Facade", "Using the Facade", "demo:logger facade", "facade", "TestLoggerDemoFacade", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Formatted facade methods", "使用 Facade", "Using the Facade", "demo:logger formatted-facade", "formatted-facade", "TestLoggerDemoFormattedFacade", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Fatal logging process exit", "使用 Facade", "Using the Facade", "demo:logger fatal", "fatal", "TestLoggerDemoFatal", LevelScenario, StatusPlanned),
	baselineEntry("logger", "Single structured context field", "单个字段", "Single Field", "demo:logger with-field", "with-field", "TestLoggerDemoWithField", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Multiple structured context fields", "多个字段", "Multiple Fields", "demo:logger with-fields", "with-fields", "TestLoggerDemoWithFields", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Error context attachment", "错误上下文", "Error Context", "demo:logger with-error", "with-error", "TestLoggerDemoWithError", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Error stacktrace line formatting", "错误上下文", "Error Context", "demo:logger error-stacktrace", "error-stacktrace", "TestLoggerDemoErrorStacktrace", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Manager and channel context extraction", "Context 字段提取", "Context Extraction", "demo:logger context-extractor", "context-extractor", "TestLoggerDemoContextExtractor", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Context extraction without an extractor", "Context 字段提取", "Context Extraction", "demo:logger context-noop", "context-noop", "TestLoggerDemoContextNoop", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Named channel logging", "写入特定通道", "Writing to Specific Channels", "demo:logger named-channel", "named-channel", "TestLoggerDemoNamedChannel", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Missing named channel fallback", "写入特定通道", "Writing to Specific Channels", "demo:logger missing-channel", "missing-channel", "TestLoggerDemoMissingChannel", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Named channel context chaining", "写入特定通道", "Writing to Specific Channels", "demo:logger channel-context", "channel-context", "TestLoggerDemoChannelContext", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Facade Manager resolution", "接口参考", "Interface Reference", "demo:logger resolve-manager", "resolve-manager", "TestLoggerDemoResolveManager", LevelScenario, StatusPlanned),
	baselineEntry("logger", "Default channel name lookup", "接口参考", "Interface Reference", "demo:logger default-name", "default-name", "TestLoggerDemoDefaultName", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Driver interface contract", "自定义 Driver", "Custom Drivers", "demo:logger driver-contract", "driver-contract", "TestLoggerDemoDriverContract", LevelCompile, StatusPlanned),
	baselineEntry("logger", "Custom driver registration", "自定义 Driver", "Custom Drivers", "demo:logger custom-driver", "custom-driver", "TestLoggerDemoCustomDriver", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Custom driver replacement", "自定义 Driver", "Custom Drivers", "demo:logger custom-driver-replacement", "custom-driver-replacement", "TestLoggerDemoCustomDriverReplacement", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Custom formatter registration", "自定义 Formatter", "Custom Formatters", "demo:logger custom-formatter", "custom-formatter", "TestLoggerDemoCustomFormatter", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Formatter channel and custom parameters", "自定义 Formatter", "Custom Formatters", "demo:logger formatter-params", "formatter-params", "TestLoggerDemoFormatterParams", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Laravel-style line formatter", "内置 Formatter", "Built-in Formatters", "demo:logger line-formatter", "line-formatter", "TestLoggerDemoLineFormatter", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "logrus text formatter", "内置 Formatter", "Built-in Formatters", "demo:logger text-formatter", "text-formatter", "TestLoggerDemoTextFormatter", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "JSON line formatter", "内置 Formatter", "Built-in Formatters", "demo:logger json-formatter", "json-formatter", "TestLoggerDemoJSONFormatter", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Programmatic Manager initialization", "程序化初始化", "Programmatic Initialization", "demo:logger manual-manager", "manual-manager", "TestLoggerDemoManualManager", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Service Provider singleton and lazy initialization", "Service Provider", "Service Provider", "demo:logger provider-lifecycle", "provider-lifecycle", "TestLoggerDemoProviderLifecycle", LevelScenario, StatusPlanned),
	baselineEntry("logger", "Service Provider automatic cleanup", "Service Provider", "Service Provider", "demo:logger provider-cleanup", "provider-cleanup", "TestLoggerDemoProviderCleanup", LevelScenario, StatusPlanned),
	baselineEntry("logger", "Manager driver resource cleanup", "资源清理", "Resource Cleanup", "demo:logger resource-cleanup", "resource-cleanup", "TestLoggerDemoResourceCleanup", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Writes after Manager close", "资源清理", "Resource Cleanup", "demo:logger closed-writes", "closed-writes", "TestLoggerDemoClosedWrites", LevelHermetic, StatusPlanned),
	baselineEntry("logger", "Global logrus compatibility bridge", "与全局 logrus 的关系", "Relationship with Global logrus", "demo:logger global-logrus", "global-logrus", "TestLoggerDemoGlobalLogrus", LevelScenario, StatusPlanned),
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
	baselineEntry("queue", "Poison envelope lifecycle event", "生命周期事件", "Lifecycle Events", "demo:queue poison-event --connection=redis", "poison-event", "TestQueueDemoPoisonEnvelopeEvent", LevelIntegration, StatusImplemented, "redis"),
	baselineEntry("queue", "Infrastructure lifecycle events", "生命周期事件", "Lifecycle Events", "demo:queue infrastructure-events --connection=rabbitmq", "infrastructure-events", "TestQueueDemoInfrastructureEvents", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("queue", "Encrypted payloads", "加密 Payload", "Encrypted Payloads", "demo:queue encryption", "encryption", "TestQueueDemoEncryption", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Missing encryption key failure", "加密 Payload", "Encrypted Payloads", "demo:queue encryption-missing-key", "encryption-missing-key", "TestQueueDemoEncryptionMissingKey", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Custom driver registration", "自定义驱动", "Custom Drivers", "demo:queue custom-driver", "custom-driver", "TestQueueDemoCustomDriver", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Custom queue contract", "Queue — 队列传输连接", "Queue — Transport Connection", "demo:queue custom-queue-contract", "custom-queue-contract", "TestQueueDemoCustomQueueContract", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Custom reserved job lifecycle", "ReservedJob — 已保留任务", "ReservedJob — Held Job", "demo:queue custom-reserved-job", "custom-reserved-job", "TestQueueDemoCustomReservedJob", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Custom pop session provider", "可选接口", "Optional Interfaces", "demo:queue custom-pop-session", "custom-pop-session", "TestQueueDemoCustomPopSession", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Custom consumer intent leaser", "可选接口", "Optional Interfaces", "demo:queue custom-consumer-intent", "custom-consumer-intent", "TestQueueDemoCustomConsumerIntent", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Core error constants", "错误常量", "Error Constants", "demo:queue errors", "errors", "TestQueueDemoErrors", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Job state error constants", "错误常量", "Error Constants", "demo:queue job-errors", "job-errors", "TestQueueDemoJobStateErrors", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Connection error constants", "错误常量", "Error Constants", "demo:queue connection-errors", "connection-errors", "TestQueueDemoConnectionErrors", LevelHermetic, StatusImplemented),
	baselineEntry("queue", "Poison envelope errors", "错误常量", "Error Constants", "demo:queue poison-errors --connection=redis", "poison-errors", "TestQueueDemoPoisonErrors", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "RabbitMQ error constants", "错误常量", "Error Constants", "demo:queue rabbitmq-errors --connection=rabbitmq", "rabbitmq-errors", "TestQueueDemoRabbitMQErrors", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("queue", "Redis transport boundaries", "Redis 连接", "Redis Connection", "demo:queue redis --connection=redis", "redis", "TestQueueDemoRedisBoundaries", LevelIntegration, StatusImplemented, "redis"),
	baselineEntry("queue", "RabbitMQ transport boundaries", "RabbitMQ 配置", "RabbitMQ Configuration", "demo:queue rabbitmq --connection=rabbitmq", "rabbitmq", "TestQueueDemoRabbitMQBoundaries", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("queue", "Bulk transport dispatch", "连接能力矩阵", "Connection Capability Matrix", "demo:queue bulk --connection=redis", "bulk", "TestQueueDemoBulkTransport", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Delayed transport dispatch", "连接能力矩阵", "Connection Capability Matrix", "demo:queue transport-delay --connection=redis", "transport-delay", "TestQueueDemoTransportDelay", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Blocking pop and queue priority", "连接能力矩阵", "Connection Capability Matrix", "demo:queue blocking-pop --connection=redis", "blocking-pop", "TestQueueDemoBlockingPop", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Redis retry-after reservation", "连接能力矩阵", "Connection Capability Matrix", "demo:queue redis-retry-after --connection=redis", "redis-retry-after", "TestQueueDemoRedisRetryAfter", LevelIntegration, StatusImplemented, "redis"),
	baselineEntry("queue", "RabbitMQ unsupported retry-after", "连接能力矩阵", "Connection Capability Matrix", "demo:queue rabbitmq-retry-after --connection=rabbitmq", "rabbitmq-retry-after", "TestQueueDemoRabbitMQRetryAfter", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("queue", "RabbitMQ publisher confirms", "RabbitMQ 配置", "RabbitMQ Configuration", "demo:queue rabbitmq-confirm --connection=rabbitmq", "rabbitmq-confirm", "TestQueueDemoRabbitMQPublisherConfirm", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("queue", "RabbitMQ reconnect lifecycle", "RabbitMQ 配置", "RabbitMQ Configuration", "demo:queue rabbitmq-reconnect --connection=rabbitmq", "rabbitmq-reconnect", "TestQueueDemoRabbitMQReconnect", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("queue", "RabbitMQ topology declaration", "RabbitMQ 配置", "RabbitMQ Configuration", "demo:queue rabbitmq-topology --connection=rabbitmq", "rabbitmq-topology", "TestQueueDemoRabbitMQTopology", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("queue", "RabbitMQ delay modes", "RabbitMQ 配置", "RabbitMQ Configuration", "demo:queue rabbitmq-delay-modes --connection=rabbitmq", "rabbitmq-delay-modes", "TestQueueDemoRabbitMQDelayModes", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("queue", "Poison envelope rejection", "连接能力矩阵", "Connection Capability Matrix", "demo:queue poison-rejection --connection=rabbitmq", "poison-rejection", "TestQueueDemoPoisonEnvelopeRejection", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("ratelimit", "Rate limiter architecture", "架构概览", "Architecture", "demo:ratelimit architecture", "architecture", "TestRateLimitDemoArchitecture", LevelCompile, StatusPlanned),
	baselineEntry("ratelimit", "Limiter cache configuration", "配置项说明", "Configuration Items", "demo:ratelimit config", "config", "TestRateLimitDemoConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Dedicated store and default-store fallback", "配置行为", "Configuration Items", "demo:ratelimit store-resolution", "store-resolution", "TestRateLimitDemoStoreResolution", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Single-process memory limiter state", "配置行为", "Configuration Items", "demo:ratelimit memory-store", "memory-store", "TestRateLimitDemoMemoryStore", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Shared Redis limiter state", "配置行为", "Configuration Items", "demo:ratelimit redis-store --store=redis", "redis-store", "TestRateLimitDemoRedisStore", LevelIntegration, StatusPlanned, "redis"),
	baselineEntry("ratelimit", "Cache configuration registration", "配置注册", "Configuration", "demo:ratelimit config-registration", "config-registration", "TestRateLimitDemoConfigurationRegistration", LevelCompile, StatusPlanned),
	baselineEntry("ratelimit", "Application startup initialization", "应用启动时自动初始化", "Quick Start", "demo:ratelimit auto-initialization", "auto-initialization", "TestRateLimitDemoAutomaticInitialization", LevelScenario, StatusPlanned),
	baselineEntry("ratelimit", "Named limiter route quick start", "注册命名限流器并挂载到路由", "Quick Start", "demo:ratelimit quick-start", "quick-start", "TestRateLimitDemoQuickStart", LevelScenario, StatusPlanned),
	baselineEntry("ratelimit", "Explicit standalone limiter construction", "独立程序或测试中显式创建", "Quick Start", "demo:ratelimit explicit-limiter", "explicit-limiter", "TestRateLimitDemoExplicitLimiter", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Named limiter registration", "注册命名限流器", "Named Limiters", "demo:ratelimit named-registration", "named-registration", "TestRateLimitDemoNamedRegistration", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Named limiter lookup", "读取已注册的命名限流器", "Named Limiters", "demo:ratelimit named-lookup", "named-lookup", "TestRateLimitDemoNamedLookup", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Unregistered limiter pass-through", "读取已注册的命名限流器", "Named Limiters", "demo:ratelimit unregistered-pass-through", "unregistered-pass-through", "TestRateLimitDemoUnregisteredPassThrough", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Independent counters for multiple rules", "多条规则同时生效", "Named Limiters", "demo:ratelimit multi-rule-counters", "multi-rule-counters", "TestRateLimitDemoMultipleRuleCounters", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Any over-limit rule blocks a request", "多条规则同时生效", "Named Limiters", "demo:ratelimit multi-rule-block", "multi-rule-block", "TestRateLimitDemoMultipleRuleBlock", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Limit value contract", "Limit 结构体", "Limit Builders", "demo:ratelimit limit-contract", "limit-contract", "TestRateLimitDemoLimitContract", LevelCompile, StatusPlanned),
	baselineEntry("ratelimit", "Custom fixed-window builder", "构造器列表", "Limit Builders", "demo:ratelimit every", "every", "TestRateLimitDemoEvery", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Per-second limit builder", "构造器列表", "Limit Builders", "demo:ratelimit per-second", "per-second", "TestRateLimitDemoPerSecond", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Per-minute limit builder", "构造器列表", "Limit Builders", "demo:ratelimit per-minute", "per-minute", "TestRateLimitDemoPerMinute", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Multi-minute limit builder", "构造器列表", "Limit Builders", "demo:ratelimit per-minutes", "per-minutes", "TestRateLimitDemoPerMinutes", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Per-hour limit builder", "构造器列表", "Limit Builders", "demo:ratelimit per-hour", "per-hour", "TestRateLimitDemoPerHour", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Per-day limit builder", "构造器列表", "Limit Builders", "demo:ratelimit per-day", "per-day", "TestRateLimitDemoPerDay", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Disabled limit builder", "构造器列表", "Limit Builders", "demo:ratelimit none", "none", "TestRateLimitDemoNone", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Dimension key isolation", "By", "By", "demo:ratelimit by", "by", "TestRateLimitDemoBy", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Duplicate-key fallback", "FallbackKey", "FallbackKey", "demo:ratelimit fallback-key", "fallback-key", "TestRateLimitDemoFallbackKey", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "After callback counts matching responses", "After", "After", "demo:ratelimit after-count", "after-count", "TestRateLimitDemoAfterCount", LevelScenario, StatusPlanned),
	baselineEntry("ratelimit", "After callback skips non-matching responses", "After", "After", "demo:ratelimit after-skip", "after-skip", "TestRateLimitDemoAfterSkip", LevelScenario, StatusPlanned),
	baselineEntry("ratelimit", "After rules expose pre-count response headers", "After", "After", "demo:ratelimit after-headers", "after-headers", "TestRateLimitDemoAfterHeaders", LevelScenario, StatusPlanned),
	baselineEntry("ratelimit", "Custom over-limit response", "Response", "Response", "demo:ratelimit custom-response", "custom-response", "TestRateLimitDemoCustomResponse", LevelScenario, StatusPlanned),
	baselineEntry("ratelimit", "Runtime result contract", "Result 结构体", "Result", "demo:ratelimit result-contract", "result-contract", "TestRateLimitDemoResultContract", LevelCompile, StatusPlanned),
	baselineEntry("ratelimit", "Global limiter middleware", "使用全局限流器", "Gin Middleware", "demo:ratelimit throttle", "throttle", "TestRateLimitDemoThrottle", LevelScenario, StatusPlanned),
	baselineEntry("ratelimit", "Explicit limiter middleware", "使用显式限流器实例", "Gin Middleware", "demo:ratelimit throttle-for", "throttle-for", "TestRateLimitDemoThrottleFor", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Disabled middleware rules pass through", "中间件行为", "Gin Middleware", "demo:ratelimit disabled-rule", "disabled-rule", "TestRateLimitDemoDisabledRule", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Default over-limit response", "默认响应", "Gin Middleware", "demo:ratelimit default-response", "default-response", "TestRateLimitDemoDefaultResponse", LevelScenario, StatusPlanned),
	baselineEntry("ratelimit", "Successful request rate-limit headers", "响应头", "Gin Middleware", "demo:ratelimit success-headers", "success-headers", "TestRateLimitDemoSuccessHeaders", LevelScenario, StatusPlanned),
	baselineEntry("ratelimit", "Over-limit response headers", "响应头", "Gin Middleware", "demo:ratelimit over-limit-headers", "over-limit-headers", "TestRateLimitDemoOverLimitHeaders", LevelScenario, StatusPlanned),
	baselineEntry("ratelimit", "Tightest-rule success headers", "响应头", "Gin Middleware", "demo:ratelimit tightest-headers", "tightest-headers", "TestRateLimitDemoTightestHeaders", LevelScenario, StatusPlanned),
	baselineEntry("ratelimit", "Route-group middleware mounting", "分组挂载", "Gin Middleware", "demo:ratelimit route-group", "route-group", "TestRateLimitDemoRouteGroup", LevelScenario, StatusPlanned),
	baselineEntry("ratelimit", "Legacy route package compatibility", "route 包兼容用法", "Compatibility Through the route Package", "demo:ratelimit route-compatibility", "route-compatibility", "TestRateLimitDemoRouteCompatibility", LevelScenario, StatusPlanned),
	baselineEntry("ratelimit", "Package-level facade operations", "包级 facade", "Manual Counting API", "demo:ratelimit facade", "facade", "TestRateLimitDemoFacade", LevelScenario, StatusPlanned),
	baselineEntry("ratelimit", "Explicit limiter instance operations", "包级 facade", "Manual Counting API", "demo:ratelimit instance", "instance", "TestRateLimitDemoInstance", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Record one attempt", "Hit", "Hit", "demo:ratelimit hit", "hit", "TestRateLimitDemoHit", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Default decay for non-positive windows", "Hit", "Hit", "demo:ratelimit hit-default-decay", "hit-default-decay", "TestRateLimitDemoHitDefaultDecay", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Increment by a custom amount", "Increment", "Increment / Decrement", "demo:ratelimit increment", "increment", "TestRateLimitDemoIncrement", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Default increment amount", "Increment", "Increment / Decrement", "demo:ratelimit increment-default", "increment-default", "TestRateLimitDemoIncrementDefault", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Decrement by a custom amount", "Decrement", "Increment / Decrement", "demo:ratelimit decrement", "decrement", "TestRateLimitDemoDecrement", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Default decrement amount", "Decrement", "Increment / Decrement", "demo:ratelimit decrement-default", "decrement-default", "TestRateLimitDemoDecrementDefault", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Read current attempts", "Attempts", "Attempts", "demo:ratelimit attempts", "attempts", "TestRateLimitDemoAttempts", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Missing attempt counter", "Attempts", "Attempts", "demo:ratelimit attempts-missing", "attempts-missing", "TestRateLimitDemoMissingAttempts", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Detect exhausted attempt limit", "TooManyAttempts", "TooManyAttempts", "demo:ratelimit too-many-attempts", "too-many-attempts", "TestRateLimitDemoTooManyAttempts", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Disabled non-positive attempt limit", "TooManyAttempts", "TooManyAttempts", "demo:ratelimit non-positive-limit", "non-positive-limit", "TestRateLimitDemoNonPositiveLimit", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Expired-window counter cleanup", "TooManyAttempts", "TooManyAttempts", "demo:ratelimit expired-window", "expired-window", "TestRateLimitDemoExpiredWindow", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Remaining attempts and RetriesLeft alias", "Remaining / RetriesLeft", "Remaining / RetriesLeft", "demo:ratelimit remaining", "remaining", "TestRateLimitDemoRemaining", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Remaining attempts lower bound", "Remaining / RetriesLeft", "Remaining / RetriesLeft", "demo:ratelimit remaining-floor", "remaining-floor", "TestRateLimitDemoRemainingFloor", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Seconds until availability", "AvailableIn", "AvailableIn", "demo:ratelimit available-in", "available-in", "TestRateLimitDemoAvailableIn", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Elapsed availability lower bound", "AvailableIn", "AvailableIn", "demo:ratelimit available-in-elapsed", "available-in-elapsed", "TestRateLimitDemoAvailableInElapsed", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Reset attempts while preserving timer", "ResetAttempts", "ResetAttempts / Clear", "demo:ratelimit reset-attempts", "reset-attempts", "TestRateLimitDemoResetAttempts", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Clear attempts and timer", "Clear", "ResetAttempts / Clear", "demo:ratelimit clear", "clear", "TestRateLimitDemoClear", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Successful atomic attempt", "Attempt", "Attempt", "demo:ratelimit attempt-success", "attempt-success", "TestRateLimitDemoAttemptSuccess", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Blocked atomic attempt", "Attempt", "Attempt", "demo:ratelimit attempt-blocked", "attempt-blocked", "TestRateLimitDemoAttemptBlocked", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Failed callback is not counted", "Attempt", "Attempt", "demo:ratelimit attempt-error", "attempt-error", "TestRateLimitDemoAttemptError", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Nil atomic-attempt callback", "Attempt", "Attempt", "demo:ratelimit attempt-nil", "attempt-nil", "TestRateLimitDemoAttemptNil", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Control-character key cleaning", "CleanRateLimiterKey", "CleanRateLimiterKey", "demo:ratelimit clean-key", "clean-key", "TestRateLimitDemoCleanKey", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Hashed middleware dimension keys", "哈希 key", "Hashed Keys", "demo:ratelimit hashed-key", "hashed-key", "TestRateLimitDemoHashedKey", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Plain middleware dimension keys", "哈希 key", "Hashed Keys", "demo:ratelimit unhashed-key", "unhashed-key", "TestRateLimitDemoUnhashedKey", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Manual keys remain caller-controlled", "哈希 key", "Hashed Keys", "demo:ratelimit manual-key", "manual-key", "TestRateLimitDemoManualKey", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Business and isolation key design", "设计原则", "Key Design", "demo:ratelimit key-design", "key-design", "TestRateLimitDemoKeyDesign", LevelCompile, StatusPlanned),
	baselineEntry("ratelimit", "Counter and timer cache layout", "工作机制", "How It Works", "demo:ratelimit cache-layout", "cache-layout", "TestRateLimitDemoCacheLayout", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Fixed-window lifecycle", "判断流程", "How It Works", "demo:ratelimit fixed-window", "fixed-window", "TestRateLimitDemoFixedWindow", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Middleware cache key format", "中间件 key 格式", "How It Works", "demo:ratelimit middleware-key", "middleware-key", "TestRateLimitDemoMiddlewareKey", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Cache operation error propagation", "错误处理", "Error Handling", "demo:ratelimit cache-errors", "cache-errors", "TestRateLimitDemoCacheErrors", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Redis connection error propagation", "错误处理", "Error Handling", "demo:ratelimit redis-errors --store=redis", "redis-errors", "TestRateLimitDemoRedisErrors", LevelIntegration, StatusPlanned, "redis"),
	baselineEntry("ratelimit", "Non-integer counter errors", "错误处理", "Error Handling", "demo:ratelimit counter-type-error", "counter-type-error", "TestRateLimitDemoCounterTypeError", LevelHermetic, StatusPlanned),
	baselineEntry("ratelimit", "Fail-open or fail-closed error policy", "推荐处理方式", "Error Handling", "demo:ratelimit error-policy", "error-policy", "TestRateLimitDemoErrorPolicy", LevelCompile, StatusPlanned),
	baselineEntry("ratelimit", "Laravel 13 compatibility boundaries", "与 Laravel 13 的边界", "Boundaries Compared With Laravel 13", "demo:ratelimit laravel-compatibility", "laravel-compatibility", "TestRateLimitDemoLaravelCompatibility", LevelCompile, StatusPlanned),
	baselineEntry("redis", "Interacting with Redis", "与 Redis 交互", "Interacting With Redis", "demo:redis list", "list", "TestRedisWithRealServer", LevelIntegration, StatusPlanned, "redis"),
	baselineEntry("route", "Quick start", "快速开始", "Quick Start", "demo:route list", "list", "TestRouteDemo", LevelScenario, StatusPlanned),
	baselineEntry("schema", "Builder entry points", "Builder 入口", "Builder Entry Points", "demo:schema list", "list", "TestSchemaDemo", LevelHermetic, StatusPlanned),
	baselineEntry("service-provider", "Writing providers", "编写服务提供者", "Writing Service Providers", "demo:provider list", "list", "TestProviderDemo", LevelHermetic, StatusPlanned),
	baselineEntry("session", "Using sessions", "Session 使用", "Session Usage", "demo:session list", "list", "TestSessionDemo", LevelScenario, StatusPlanned),
	manualEntry("starter", "First endpoint", "添加第一个接口", "Add Your First Endpoint", "starter HTTP smoke test", "TestStarterSmoke", LevelScenario, "runs against a clean generated starter"),
	baselineEntry("support", "Feature overview", "功能概览", "Feature Overview", "demo:support list", "list", "TestSupportDemo", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Scheduler architecture", "简介", "Introduction", "demo:timer architecture", "architecture", "TestTimerDemoArchitecture", LevelCompile, StatusPlanned),
	baselineEntry("timer", "Long-running process requirements", "环境要求", "Requirements", "demo:timer requirements", "requirements", "TestTimerDemoRequirements", LevelCompile, StatusPlanned),
	baselineEntry("timer", "Cron command startup", "启动命令", "Start Command", "demo:timer cron-command", "cron-command", "TestTimerDemoCronCommand", LevelScenario, StatusPlanned),
	baselineEntry("timer", "Signal-driven graceful shutdown", "停止调度器", "Stopping The Scheduler", "demo:timer signal-shutdown", "signal-shutdown", "TestTimerDemoSignalShutdown", LevelScenario, StatusPlanned),
	baselineEntry("timer", "Application schedule registration", "注册任务", "Registering Tasks", "demo:timer registration", "registration", "TestTimerDemoRegistration", LevelCompile, StatusPlanned),
	baselineEntry("timer", "Process manager deployment", "生产环境部署", "Production Deployment", "demo:timer deployment", "deployment", "TestTimerDemoDeployment", LevelCompile, StatusPlanned),
	baselineEntry("timer", "Local timezone scheduling", "配置", "Configuration", "demo:timer timezone", "timezone", "TestTimerDemoTimezone", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Debug success logging", "配置", "Configuration", "demo:timer debug-logging", "debug-logging", "TestTimerDemoDebugLogging", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Overlap cache configuration", "配置", "Configuration", "demo:timer overlap-cache", "overlap-cache", "TestTimerDemoOverlapCacheConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Exception reporter configuration", "配置", "Configuration", "demo:timer exception-config", "exception-config", "TestTimerDemoExceptionConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Command task registration", "通过命令签名注册", "Registering Commands By Signature", "demo:timer command", "command", "TestTimerDemoCommand", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Command signature and resolver validation", "通过命令签名注册", "Registering Commands By Signature", "demo:timer command-validation", "command-validation", "TestTimerDemoCommandValidation", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Closure task registration", "通过闭包注册", "Registering Closures", "demo:timer call", "call", "TestTimerDemoCall", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Fixed duration intervals", "固定间隔", "Fixed Intervals", "demo:timer every", "every", "TestTimerDemoEvery", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Second-level frequency helpers", "秒级调度", "Second-Level Scheduling", "demo:timer second-frequencies", "second-frequencies", "TestTimerDemoSecondFrequencies", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Minute-level frequency helpers", "分钟级调度", "Minute-Level Scheduling", "demo:timer minute-frequencies", "minute-frequencies", "TestTimerDemoMinuteFrequencies", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Hourly scheduling", "小时级调度", "Hourly Scheduling", "demo:timer hourly", "hourly", "TestTimerDemoHourly", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Hourly minute offsets", "小时级调度", "Hourly Scheduling", "demo:timer hourly-at", "hourly-at", "TestTimerDemoHourlyAt", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Odd and stepped-hour scheduling", "小时级调度", "Hourly Scheduling", "demo:timer hour-steps", "hour-steps", "TestTimerDemoHourSteps", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Daily scheduling", "每日调度", "Daily Scheduling", "demo:timer daily", "daily", "TestTimerDemoDaily", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Daily scheduling at a time", "每日调度", "Daily Scheduling", "demo:timer daily-at", "daily-at", "TestTimerDemoDailyAt", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Twice-daily hour scheduling", "每日调度", "Daily Scheduling", "demo:timer twice-daily", "twice-daily", "TestTimerDemoTwiceDaily", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Twice-daily time scheduling", "每日调度", "Daily Scheduling", "demo:timer twice-daily-at", "twice-daily-at", "TestTimerDemoTwiceDailyAt", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Composable hit time", "`At(timeValue)` — 设置命中时间", "`At(timeValue)` - Set The Hit Time", "demo:timer at", "at", "TestTimerDemoAt", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Weekly scheduling", "每周调度", "Weekly Scheduling", "demo:timer weekly", "weekly", "TestTimerDemoWeekly", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Weekly day and time scheduling", "`WeeklyOn(dayOfWeek, timeValue...)`", "`WeeklyOn(dayOfWeek, timeValue...)`", "demo:timer weekly-on", "weekly-on", "TestTimerDemoWeeklyOn", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Weekday and weekend constraints", "每周调度", "Weekly Scheduling", "demo:timer weekday-groups", "weekday-groups", "TestTimerDemoWeekdayGroups", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Named weekday constraints", "每周调度", "Weekly Scheduling", "demo:timer named-weekdays", "named-weekdays", "TestTimerDemoNamedWeekdays", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Custom weekday constraints", "`Days(days)` — 自定义星期", "`Days(days)` - Custom Weekdays", "demo:timer days", "days", "TestTimerDemoDays", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Monthly scheduling", "每月调度", "Monthly Scheduling", "demo:timer monthly", "monthly", "TestTimerDemoMonthly", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Monthly day and time scheduling", "`MonthlyOn(dayOfMonth, timeValue...)`", "`MonthlyOn(dayOfMonth, timeValue...)`", "demo:timer monthly-on", "monthly-on", "TestTimerDemoMonthlyOn", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Twice-monthly scheduling", "`TwiceMonthly(args...)`", "`TwiceMonthly(args...)`", "demo:timer twice-monthly", "twice-monthly", "TestTimerDemoTwiceMonthly", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Last day of month scheduling", "`LastDayOfMonth(timeValue...)`", "`LastDayOfMonth(timeValue...)`", "demo:timer last-day-of-month", "last-day-of-month", "TestTimerDemoLastDayOfMonth", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Multiple days of month constraints", "`DaysOfMonth(days...)`", "`DaysOfMonth(days...)`", "demo:timer days-of-month", "days-of-month", "TestTimerDemoDaysOfMonth", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Quarterly scheduling", "季度 / 年度调度", "Quarterly / Yearly Scheduling", "demo:timer quarterly", "quarterly", "TestTimerDemoQuarterly", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Quarter-relative day scheduling", "`QuarterlyOn(args...)`", "`QuarterlyOn(args...)`", "demo:timer quarterly-on", "quarterly-on", "TestTimerDemoQuarterlyOn", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Yearly scheduling", "季度 / 年度调度", "Quarterly / Yearly Scheduling", "demo:timer yearly", "yearly", "TestTimerDemoYearly", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Yearly month, day, and time scheduling", "`YearlyOn(args...)`", "`YearlyOn(args...)`", "demo:timer yearly-on", "yearly-on", "TestTimerDemoYearlyOn", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Overlapping task prevention", "防重叠执行", "Preventing Overlapping Runs", "demo:timer without-overlapping", "without-overlapping", "TestTimerDemoWithoutOverlapping", LevelScenario, StatusPlanned),
	baselineEntry("timer", "Cross-process overlap prevention", "防重叠执行", "Preventing Overlapping Runs", "demo:timer cross-process-overlap --connection=redis", "cross-process-overlap", "TestTimerDemoCrossProcessOverlap", LevelIntegration, StatusPlanned, "redis"),
	baselineEntry("timer", "Scheduler start lifecycle", "`Start(ctx)` — 启动调度", "`Start(ctx)` - Start Scheduling", "demo:timer start", "start", "TestTimerDemoStart", LevelScenario, StatusPlanned),
	baselineEntry("timer", "Scheduler stop lifecycle", "`Stop()` — 停止调度", "`Stop()` - Stop Scheduling", "demo:timer stop", "stop", "TestTimerDemoStop", LevelScenario, StatusPlanned),
	baselineEntry("timer", "Registered task summary", "查看已注册任务", "Inspecting Registered Tasks", "demo:timer summary", "summary", "TestTimerDemoSummary", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Task error isolation and reporting", "任务返回 error", "Task Returns An Error", "demo:timer task-error", "task-error", "TestTimerDemoTaskError", LevelScenario, StatusPlanned),
	baselineEntry("timer", "Task panic isolation and reporting", "任务 panic", "Task Panics", "demo:timer task-panic", "task-panic", "TestTimerDemoTaskPanic", LevelScenario, StatusPlanned),
	baselineEntry("timer", "Successful task debug output", "任务成功执行", "Task Runs Successfully", "demo:timer task-success", "task-success", "TestTimerDemoTaskSuccess", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Exception Reporter dependency", "异常上报依赖", "Exception Reporting Dependency", "demo:timer exception-reporter", "exception-reporter", "TestTimerDemoExceptionReporter", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Fixed interval execution model", "两种调度模式", "Two Scheduling Modes", "demo:timer fixed-interval-model", "fixed-interval-model", "TestTimerDemoFixedIntervalModel", LevelScenario, StatusPlanned),
	baselineEntry("timer", "Calendar execution model", "两种调度模式", "Two Scheduling Modes", "demo:timer calendar-model", "calendar-model", "TestTimerDemoCalendarModel", LevelScenario, StatusPlanned),
	baselineEntry("timer", "Immediate run compatibility", "启动时立即执行", "Immediate Run On Startup", "demo:timer immediate-run", "immediate-run", "TestTimerDemoImmediateRun", LevelScenario, StatusPlanned),
	baselineEntry("timer", "Parallel tasks and serial task runs", "并发与串行", "Parallel and Serial Execution", "demo:timer execution-concurrency", "execution-concurrency", "TestTimerDemoExecutionConcurrency", LevelScenario, StatusPlanned),
	baselineEntry("timer", "Time string parsing and validation", "时间字符串", "Time Strings", "demo:timer time-parsing", "time-parsing", "TestTimerDemoTimeParsing", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Minute and hour list parsing", "分钟和小时列表", "Minute and Hour Lists", "demo:timer offset-parsing", "offset-parsing", "TestTimerDemoOffsetParsing", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Weekday value parsing", "星期值", "Weekday Values", "demo:timer weekday-parsing", "weekday-parsing", "TestTimerDemoWeekdayParsing", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Legacy Timer interval type", "Timer 基础类型", "Timer Base Types", "demo:timer timer-type", "timer-type", "TestTimerDemoTimerType", LevelCompile, StatusPlanned),
	baselineEntry("timer", "ResolvedCommand bridge contract", "ResolvedCommand", "ResolvedCommand", "demo:timer resolved-command", "resolved-command", "TestTimerDemoResolvedCommand", LevelCompile, StatusPlanned),
	baselineEntry("timer", "CommandResolver contract and injection", "CommandResolver", "CommandResolver", "demo:timer command-resolver", "command-resolver", "TestTimerDemoCommandResolver", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Standalone Schedule construction", "NewSchedule", "NewSchedule", "demo:timer new-schedule", "new-schedule", "TestTimerDemoNewSchedule", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Scheduled task naming", "ScheduledTask 元信息方法", "ScheduledTask Metadata Methods", "demo:timer name", "name", "TestTimerDemoName", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Scheduled task descriptions", "ScheduledTask 元信息方法", "ScheduledTask Metadata Methods", "demo:timer description", "description", "TestTimerDemoDescription", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Scheduled task defaults", "默认配置", "Defaults", "demo:timer defaults", "defaults", "TestTimerDemoDefaults", LevelHermetic, StatusPlanned),
	baselineEntry("timer", "Independent scheduler usage", "独立使用调度器（不走 Console Kernel）", "Using The Scheduler Independently Without The Console Kernel", "demo:timer standalone", "standalone", "TestTimerDemoStandalone", LevelScenario, StatusPlanned),
	baselineEntry("translation", "Translator architecture", "概述", "Overview", "demo:translation architecture", "architecture", "TestTranslationDemoArchitecture", LevelCompile, StatusPlanned),
	baselineEntry("translation", "Locale configuration", "环境变量", "Environment Variables", "demo:translation config", "config", "TestTranslationDemoConfiguration", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Automatic service provider registration", "注册服务提供者", "Registering the Service Provider", "demo:translation provider", "provider", "TestTranslationDemoServiceProvider", LevelScenario, StatusPlanned),
	baselineEntry("translation", "Grouped short-key translations", "短键（Short Key）", "Short Keys", "demo:translation short-keys", "short-keys", "TestTranslationDemoShortKeys", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Nested short-key translations", "短键（Short Key）", "Short Keys", "demo:translation nested-keys", "nested-keys", "TestTranslationDemoNestedKeys", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "JSON-key translations", "以翻译字符串作为键（JSON Key）", "JSON Keys", "demo:translation json-keys", "json-keys", "TestTranslationDemoJSONKeys", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Translation key and group conflicts", "以翻译字符串作为键（JSON Key）", "JSON Keys", "demo:translation key-conflicts", "key-conflicts", "TestTranslationDemoKeyConflicts", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Namespaced translations", "命名空间翻译（Namespace）", "Namespaces", "demo:translation namespaces", "namespaces", "TestTranslationDemoNamespaces", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Translator dependency resolution", "使用 Translator 实例", "Using a Translator Instance", "demo:translation translator", "translator", "TestTranslationDemoTranslatorInstance", LevelCompile, StatusPlanned),
	baselineEntry("translation", "Package-level facade retrieval", "使用 Facade 便捷方法", "Using Facade Helpers", "demo:translation facade", "facade", "TestTranslationDemoFacade", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Default missing-key result", "使用 Facade 便捷方法", "Using Facade Helpers", "demo:translation missing-default", "missing-default", "TestTranslationDemoMissingDefault", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Per-call locale selection", "指定语言环境", "Specifying a Locale", "demo:translation locale-argument", "locale-argument", "TestTranslationDemoLocaleArgument", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Translation key existence", "检查翻译键是否存在", "Checking Whether a Key Exists", "demo:translation has", "has", "TestTranslationDemoHas", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Locale-specific key existence", "检查翻译键是否存在", "Checking Whether a Key Exists", "demo:translation has-for-locale", "has-for-locale", "TestTranslationDemoHasForLocale", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Whole and nested group retrieval", "获取整个分组", "Retrieving a Whole Group", "demo:translation get-map", "get-map", "TestTranslationDemoGetMap", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Placeholder replacement", "替换翻译字符串中的参数", "Replacing Parameters", "demo:translation replacements", "replacements", "TestTranslationDemoReplacements", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Placeholder case conversion", "占位符大小写规则", "Placeholder Case Rules", "demo:translation replacement-case", "replacement-case", "TestTranslationDemoReplacementCase", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Custom and Stringer value formatting", "自定义类型格式化（Stringable）", "Custom Formatting With Stringable", "demo:translation stringable", "stringable", "TestTranslationDemoStringable", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Implicit plural variants", "基础管道语法", "Basic Pipe Syntax", "demo:translation plural-pipe", "plural-pipe", "TestTranslationDemoPluralPipe", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Explicit plural intervals", "显式区间语法", "Explicit Intervals", "demo:translation plural-intervals", "plural-intervals", "TestTranslationDemoPluralIntervals", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Plural replacement values", "在复数化字符串中使用占位符", "Placeholders in Pluralized Strings", "demo:translation plural-replacements", "plural-replacements", "TestTranslationDemoPluralReplacements", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Plural count placeholder", "在复数化字符串中使用占位符", "Placeholders in Pluralized Strings", "demo:translation plural-count", "plural-count", "TestTranslationDemoPluralCount", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Current locale access and mutation", "获取与设置当前语言环境", "Current Locale", "demo:translation locale", "locale", "TestTranslationDemoLocale", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Current locale validation", "获取与设置当前语言环境", "Current Locale", "demo:translation locale-validation", "locale-validation", "TestTranslationDemoLocaleValidation", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Fallback locale lookup and mutation", "回退语言环境", "Fallback Locale", "demo:translation fallback", "fallback", "TestTranslationDemoFallback", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Fallback locale validation", "回退语言环境", "Fallback Locale", "demo:translation fallback-validation", "fallback-validation", "TestTranslationDemoFallbackValidation", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Custom locale resolution order", "自定义语言环境解析链", "Custom Locale Resolution", "demo:translation locale-resolver", "locale-resolver", "TestTranslationDemoLocaleResolver", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Package language-file overrides", "覆盖包语言文件", "Overriding Package Language Files", "demo:translation namespace-overrides", "namespace-overrides", "TestTranslationDemoNamespaceOverrides", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Runtime translation lines", "运行时添加翻译行", "Adding Lines at Runtime", "demo:translation add-lines", "add-lines", "TestTranslationDemoAddLines", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Runtime line precedence", "运行时添加翻译行", "Adding Lines at Runtime", "demo:translation add-lines-precedence", "add-lines-precedence", "TestTranslationDemoAddLinesPrecedence", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Missing-key handlers", "缺失翻译键处理", "Handling Missing Keys", "demo:translation missing-handler", "missing-handler", "TestTranslationDemoMissingHandler", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Grouped translation paths", "分组翻译路径（AddPath）", "Group Paths", "demo:translation group-paths", "group-paths", "TestTranslationDemoGroupPaths", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "JSON translation paths", "JSON 翻译路径（AddJSONPath）", "JSON Paths", "demo:translation json-paths", "json-paths", "TestTranslationDemoJSONPaths", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Custom loader replacement", "自定义加载器", "Custom Loaders", "demo:translation custom-loader", "custom-loader", "TestTranslationDemoCustomLoader", LevelCompile, StatusPlanned),
	baselineEntry("translation", "Translator interface contract", "Translator 接口", "Translator", "demo:translation translator-contract", "translator-contract", "TestTranslationDemoTranslatorContract", LevelCompile, StatusPlanned),
	baselineEntry("translation", "Loader interface contract and inspection", "Loader 接口", "Loader", "demo:translation loader-contract", "loader-contract", "TestTranslationDemoLoaderContract", LevelCompile, StatusPlanned),
	baselineEntry("translation", "Plural selector interface contract", "Selector 接口", "Selector", "demo:translation selector-contract", "selector-contract", "TestTranslationDemoSelectorContract", LevelCompile, StatusPlanned),
	baselineEntry("translation", "Facade state reset", "测试", "Testing", "demo:translation reset", "reset", "TestTranslationDemoReset", LevelHermetic, StatusPlanned),
	baselineEntry("translation", "Isolated translator testing", "测试", "Testing", "demo:translation isolated", "isolated", "TestTranslationDemoIsolated", LevelHermetic, StatusPlanned),
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
