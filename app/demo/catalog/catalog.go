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
	baselineEntry("event", "Listeners", "注册事件与监听器", "Registering Events and Listeners", "demo:event list", "list", "TestEventDemo", LevelHermetic, StatusPlanned),
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
	baselineEntry("filesystem", "OSS disk switching and ACL mapping", "OSS 驱动", "OSS Driver", "demo:filesystem oss-driver --disk=oss", "oss-driver", "TestFilesystemDemoOSSDriver", LevelIntegration, StatusPlanned, "oss"),
	baselineEntry("filesystem", "Custom driver registration", "注册 Driver", "Registering a Driver", "demo:filesystem custom-driver", "custom-driver", "TestFilesystemDemoCustomDriver", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Custom driver replacement and lifecycle", "注册 Driver", "Registering a Driver", "demo:filesystem custom-driver-lifecycle", "custom-driver-lifecycle", "TestFilesystemDemoCustomDriverLifecycle", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Driver contract", "Driver 接口", "Driver Interface", "demo:filesystem driver-contract", "driver-contract", "TestFilesystemDemoDriverContract", LevelCompile, StatusPlanned),
	baselineEntry("filesystem", "Optional driver capabilities and fallbacks", "Driver 接口", "Driver Interface", "demo:filesystem optional-driver-capabilities", "optional-driver-capabilities", "TestFilesystemDemoOptionalDriverCapabilities", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Driver factory context", "DriverFactoryContext 字段说明", "DriverFactoryContext Fields", "demo:filesystem driver-factory-context", "driver-factory-context", "TestFilesystemDemoDriverFactoryContext", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Manual manager construction and cleanup", "手动初始化", "Manual Initialization", "demo:filesystem manual-manager", "manual-manager", "TestFilesystemDemoManualManager", LevelScenario, StatusPlanned),
	baselineEntry("filesystem", "Manager construction from application config", "手动初始化", "Manual Initialization", "demo:filesystem manager-from-config", "manager-from-config", "TestFilesystemDemoManagerFromConfig", LevelScenario, StatusPlanned),
	baselineEntry("filesystem", "Filesystem error constants", "错误常量", "Error Constants", "demo:filesystem errors", "errors", "TestFilesystemDemoErrors", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "Local driver capability matrix", "内置驱动能力矩阵", "Built-in Driver Capability Matrix", "demo:filesystem local-capabilities", "local-capabilities", "TestFilesystemDemoLocalCapabilities", LevelHermetic, StatusPlanned),
	baselineEntry("filesystem", "OSS driver capability matrix", "内置驱动能力矩阵", "Built-in Driver Capability Matrix", "demo:filesystem oss-capabilities --disk=oss", "oss-capabilities", "TestFilesystemDemoOSSCapabilities", LevelIntegration, StatusPlanned, "oss"),
	baselineEntry("filesystem", "Laravel filesystem API mapping", "与 Laravel Filesystem 的对应关系", "Laravel Filesystem Mapping", "demo:filesystem laravel-compatibility", "laravel-compatibility", "TestFilesystemDemoLaravelCompatibility", LevelCompile, StatusPlanned),
	baselineEntry("filesystem", "Filesystem usage guidance", "使用建议", "Best Practices", "demo:filesystem best-practices", "best-practices", "TestFilesystemDemoBestPractices", LevelCompile, StatusPlanned),
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
	baselineEntry("queue", "Bulk transport dispatch", "内置连接能力矩阵", "Built-in Connection Capability Matrix", "demo:queue bulk --connection=redis", "bulk", "TestQueueDemoBulkTransport", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Delayed transport dispatch", "内置连接能力矩阵", "Built-in Connection Capability Matrix", "demo:queue transport-delay --connection=redis", "transport-delay", "TestQueueDemoTransportDelay", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Blocking pop and queue priority", "内置连接能力矩阵", "Built-in Connection Capability Matrix", "demo:queue blocking-pop --connection=redis", "blocking-pop", "TestQueueDemoBlockingPop", LevelIntegration, StatusImplemented, "redis", "rabbitmq"),
	baselineEntry("queue", "Redis retry-after reservation", "内置连接能力矩阵", "Built-in Connection Capability Matrix", "demo:queue redis-retry-after --connection=redis", "redis-retry-after", "TestQueueDemoRedisRetryAfter", LevelIntegration, StatusImplemented, "redis"),
	baselineEntry("queue", "RabbitMQ unsupported retry-after", "内置连接能力矩阵", "Built-in Connection Capability Matrix", "demo:queue rabbitmq-retry-after --connection=rabbitmq", "rabbitmq-retry-after", "TestQueueDemoRabbitMQRetryAfter", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("queue", "RabbitMQ publisher confirms", "RabbitMQ 配置", "RabbitMQ Configuration", "demo:queue rabbitmq-confirm --connection=rabbitmq", "rabbitmq-confirm", "TestQueueDemoRabbitMQPublisherConfirm", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("queue", "RabbitMQ reconnect lifecycle", "RabbitMQ 配置", "RabbitMQ Configuration", "demo:queue rabbitmq-reconnect --connection=rabbitmq", "rabbitmq-reconnect", "TestQueueDemoRabbitMQReconnect", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("queue", "RabbitMQ topology declaration", "RabbitMQ 配置", "RabbitMQ Configuration", "demo:queue rabbitmq-topology --connection=rabbitmq", "rabbitmq-topology", "TestQueueDemoRabbitMQTopology", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("queue", "RabbitMQ delay modes", "RabbitMQ 配置", "RabbitMQ Configuration", "demo:queue rabbitmq-delay-modes --connection=rabbitmq", "rabbitmq-delay-modes", "TestQueueDemoRabbitMQDelayModes", LevelIntegration, StatusImplemented, "rabbitmq"),
	baselineEntry("queue", "Poison envelope rejection", "内置连接能力矩阵", "Built-in Connection Capability Matrix", "demo:queue poison-rejection --connection=rabbitmq", "poison-rejection", "TestQueueDemoPoisonEnvelopeRejection", LevelIntegration, StatusImplemented, "rabbitmq"),
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
