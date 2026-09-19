<p align="center">
  <img src=".dev/assets/logo.png" width="250">
</p>

<div align="center">

**PrismGo —— 像写 Laravel 一样写 Go**

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Module](https://img.shields.io/badge/module-github.com%2Fprismgo%2Fframework-blue)](https://github.com/prismgo/framework)
[![Coverage](https://codecov.io/gh/prismgo/framework/branch/main/graph/badge.svg)](https://codecov.io/gh/prismgo/framework)
[![Latest Version](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fproxy.golang.org%2Fgithub.com%2Fprismgo%2Fframework%2F%40latest&query=%24.Version&label=version)](https://pkg.go.dev/github.com/prismgo/framework?tab=versions)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)

简体中文 | [English](README_en.md)

</div>

---

## 简介

PrismGo Demo 是框架的**可运行示例与本地联调项目**，通过 `go.mod` 的本地 `replace` 直接使用同级 `framework/` 源码。

Demo 项目主要承担四件事：

- **展示用法**：以路由、命令、模型、仓储、事务、队列任务和定时任务等真实应用代码展示框架组合方式。
- **维护覆盖目录**：Catalog 将“Framework 版本 → 功能模块 → 文档章节 → 示例 → 测试”串成可查询的进度地图。
- **隔离验证**：默认测试 Application 使用 SQLite、内存/文件驱动和临时目录，不依赖开发环境的 `.env`。
- **本地联调**：`./dev` 统一初始化工作区，并按需管理 MySQL、PostgreSQL、SQL Server、Redis 和 RabbitMQ。

| 路径 | 作用 |
|---|---|
| `app/demo/catalog/` | Catalog 数据、模块进度和文档映射 |
| `app/demo/` | 可执行的功能示例与场景测试 |
| `app/models/`、`app/repositories/`、`app/services/` | 订单领域的模型、仓储和业务服务示例 |
| `app/demo/testing/` | 与本地开发环境隔离的测试 Application |
| `database/migrations/` | Demo 数据表迁移 |
| `.dev/`、`dev` | 工作区初始化、本地依赖和联调工具 |

## Catalog 进度地图

Catalog 当前覆盖 **29 个模块、1443 个条目**：已实现 1376 个、计划中 64 个、手工验证 3 个。以下是 README 更新时的快照；`app/demo/catalog/` 是唯一数据源，实时进度以 `demo:list` 输出为准。

| 模块 | 覆盖范围 | 进度 | 剩余 | 状态 |
|---|---|---:|---:|---|
| `cache` | 缓存存储、标签与锁 | 57/57 | 0 | 已实现 |
| `commands` | 应用命令与选项 | 101/101 | 0 | 已实现 |
| `config` | 环境与应用配置 | 62/62 | 0 | 已实现 |
| `console` | Artisan 风格命令与终端 IO | 70/70 | 0 | 已实现 |
| `container` | 依赖绑定与解析 | 47/47 | 0 | 已实现 |
| `cookie` | Cookie 值与排队写入 | 50/50 | 0 | 已实现 |
| `database` | 数据库连接、模型、迁移与索引维护 | 0/62 | 62 | 计划中 |
| `encryption` | 应用数据加密 | 0/1 | 1 | 计划中 |
| `event` | 同步、异步与队列事件 | 57/57 | 0 | 已实现 |
| `exception` | 异常报告与渲染 | 77/77 | 0 | 已实现 |
| `facade` | 框架 Facade 访问 | 12/12 | 0 | 已实现 |
| `filesystem` | 本地、公共与云文件系统 | 67/67 | 0 | 已实现 |
| `horizon` | 队列监控与 Worker 管理 | 1/1 | 0 | 已实现 |
| `http-server` | HTTP 服务启动与生命周期 | 29/29 | 0 | 已实现 |
| `installation` | 应用安装流程 | 0/1 | — | 手工验证 |
| `lens` | Lens 开发流程 | 0/1 | — | 手工验证 |
| `lifecycle` | 应用启动与关闭 | 57/57 | 0 | 已实现 |
| `logger` | 多通道应用日志 | 40/40 | 0 | 已实现 |
| `queue` | 队列、任务与 Worker | 59/59 | 0 | 已实现 |
| `ratelimit` | 请求与操作限流 | 74/74 | 0 | 已实现 |
| `redis` | Redis 连接与操作 | 93/93 | 0 | 已实现 |
| `route` | HTTP 路由注册 | 87/87 | 0 | 已实现 |
| `schema` | 数据库 Schema 与迁移构建器 | 143/143 | 0 | 已实现 |
| `service-provider` | Service Provider 注册与生命周期 | 42/42 | 0 | 已实现 |
| `session` | 服务端 Session 存储 | 50/50 | 0 | 已实现 |
| `starter` | 生成应用的起始模板 | 0/1 | — | 手工验证 |
| `support` | 通用框架辅助函数 | 0/1 | 1 | 计划中 |
| `timer` | 定时任务定义 | 62/62 | 0 | 已实现 |
| `translation` | 应用翻译与复数处理 | 39/39 | 0 | 已实现 |

在工作区根目录查看实时地图与条目明细：

```bash
go run ./demo demo:list
go run ./demo demo:show queue
go run ./demo demo:show queue --status=planned
go run ./demo demo:show queue --level=integration
go run ./demo demo:list --json
go run ./demo demo:cache list
go run ./demo demo:event list
go run ./demo demo:event queued-redis --connection=redis
go run ./demo demo:console list
go run ./demo demo:console signature --json
go run ./demo demo:cache get --json
go run ./demo demo:container list
go run ./demo demo:container list-entries --json
go run ./demo demo:container singleton --json
go run ./demo demo:cookie list
go run ./demo demo:cookie session-queue --json
go run ./demo demo:session list
go run ./demo demo:session regenerate --json
go run ./demo demo:session redis-driver --connection=redis
go run ./demo demo:session redis-lock --connection=redis
go run ./demo demo:redis list
go run ./demo demo:redis config --json
go run ./demo demo:redis strings --json
go run ./demo demo:redis publish
go run ./demo demo:redis default-connection --json
go run ./demo demo:redis connection-reuse --json
go run ./demo demo:redis command-executed-event --json
go run ./demo demo:redis connection-listener --json
go run ./demo demo:redis cache-flush --json
go run ./demo demo:redis queue-ready --json
go run ./demo demo:redis horizon-metrics --json
go run ./demo demo:lifecycle list
go run ./demo demo:lifecycle run-context --json
go run ./demo demo:commands list
go run ./demo demo:commands serve-reload --json
go run ./demo demo:http-server list
go run ./demo demo:http-server routes --json
go run ./demo demo:filesystem list
go run ./demo demo:filesystem all-directories --json
go run ./demo demo:route list
go run ./demo demo:route list-entries --json
go run ./demo demo:route facade --json
go run ./demo demo:route url-escaping
go run ./demo demo:route api-resource --json
go run ./demo demo:route domain
go run ./demo demo:route list-command --json
go run ./demo demo:route throttle-route
go run ./demo demo:route provider-singleton
go run ./demo demo:route best-practices --json
go run ./demo demo:provider list
go run ./demo demo:provider singleton --json
go run ./demo demo:provider deferred-resolution --json
go run ./demo demo:provider terminate-order
go run ./demo demo:ratelimit list
go run ./demo demo:ratelimit architecture
go run ./demo demo:ratelimit per-minute --json
go run ./demo demo:ratelimit store-resolution
go run ./demo demo:ratelimit quick-start
go run ./demo demo:ratelimit throttle
go run ./demo demo:ratelimit throttle-for
go run ./demo demo:ratelimit success-headers
go run ./demo demo:ratelimit route-compatibility
go run ./demo demo:ratelimit hit --json
go run ./demo demo:ratelimit cache-layout
go run ./demo demo:ratelimit hashed-key
go run ./demo demo:ratelimit cache-errors
go run ./demo demo:ratelimit redis-store --store=redis
go run ./demo demo:ratelimit redis-errors --store=redis
go run ./demo demo:timer list
go run ./demo demo:timer architecture
go run ./demo demo:timer every --json
go run ./demo demo:timer weekly-on
go run ./demo demo:schema list
go run ./demo demo:schema architecture
go run ./demo demo:schema create --json
go run ./demo demo:schema change-column
go run ./demo demo:schema drop-all-tables
go run ./demo demo:schema morphs --json
go run ./demo demo:schema soft-deletes
go run ./demo demo:schema enum-set
go run ./demo demo:schema unsigned
go run ./demo demo:schema default
go run ./demo demo:schema drop-morphs
go run ./demo demo:schema unique-index
go run ./demo demo:schema index
go run ./demo demo:schema foreign-dialect
go run ./demo demo:schema drop-foreign
go run ./demo demo:schema tables
go run ./demo demo:schema has-columns
go run ./demo demo:schema columns
go run ./demo demo:schema indexes
go run ./demo demo:schema when-missing-column
go run ./demo demo:schema sync-models
go run ./demo demo:schema create-database
```

在 `demo/` 中运行本地 OSS HTTP 集成验收（不需要云端凭证）：

```bash
PRISMGO_FILESYSTEM_LOCAL_OSS_TEST=1 go test ./app/demo/filesystem -run TestFilesystemDemoLocalOSSIntegration -count=1
```

`ratelimit` 的 72 个场景为编译级、hermetic 或 scenario，默认不依赖外部服务；`redis-store` 与 `redis-errors` 为 integration，加载 `demo/.dev/runtime/test.env` 后运行 `go test ./app/demo/ratelimit -run 'TestRateLimitDemoRedis(Store|Errors)' -count=1`，或设置 `CACHE_LIMITER_DRIVER=redis` 与 Redis 连接变量后执行 `go run ./demo demo:ratelimit redis-store --store=redis`。

`schema` 已实现 143 个场景，覆盖 catalog 全部条目。`architecture`、`sqlite-extension`、`sqlite-connection-scope`、`create-dialect-options`、`drop-all-types`、`stored-as`、`virtual-as`、`from`、`instant`、`lock`、`change-semantics`、`metadata-types`、`foreign-key-toggle-dialects`、`sync-models-boundaries`、`dialect-compatibility`、`laravel-compatibility`、`ensure-extension`、`ensure-vector-extension` 为编译级；`drop-all-tables` 与 `drop-all-views` 在独立 SQLite 临时库上验证；其余 hermetic 场景在隔离 SQLite 上用 `Blueprint` 定义与字段/索引元数据断言，包括整数族、字符串与文本族、UUID/ULID、JSON、枚举/集合、空间与向量类型、时间族、外键 ID、多态字段、nullable/not-null，以及字段修饰符、删除约定字段、索引创建，和索引删除、表/视图/Schema/类型/字段元数据检查，以及最后一批新增的字段/索引元数据与命名列表（`columns`、`column-type`、`has-index`、`indexes`）、字段与索引条件执行（`when-has-column`、`when-missing-column`、`when-missing-index`）、SyncModels 过渡场景（`sync-models`、`sync-models-columns`、`sync-models-defaults`）。`default-*`、`morph-*`、`explicit-tag-precedence`、`change-column`、`raw`、`unique-modifier`、`comment`、`first`、`after`、`charset`、`collation`、`use-current`、`use-current-on-update`、`invisible`、`drop-constrained-foreign-id`、`rename-index`、`drop-primary`、`foreign-keys`、`disable-foreign-keys`、`enable-foreign-keys`、`without-foreign-keys`、`create-database`、`drop-database` 为 integration，需真实 MySQL（SQLite 忽略长度、精度与未命名的内联唯一约束，且无法重命名索引或删除主键；`create-database`/`drop-database` 使用本地测试环境专门授权的 `prismgo_schema_demo_test` 一次性库）。外键场景 `constrained`、`constrained-explicit`、`foreign`、`foreign-actions`、`cascade-actions`、`restrict-actions`、`null-actions`、`no-action-actions`、`foreign-name`、`drop-foreign` 同样需要真实 MySQL，`foreign-dialect` 按方言验证外键 SQL 边界（MySQL 生成约束，SQLite 在 `CREATE TABLE` 中跳过）；`fulltext-index`、`spatial-index` 在 SQLite 降级为普通索引，`index-naming` 校验默认命名与超过 64 字符时的 SHA1 截断。列类型场景在 MySQL 下额外断言精确 DDL：加载 `demo/.dev/runtime/test.env` 后运行 `go test ./app/demo/schema -count=1 -v`，其中 `TestSchemaDemoMySQLColumnTypes` 会校验 `varchar(120)`、`decimal(10,2) unsigned`、`enum(...)`、索引、注释、`INVISIBLE`、`ON UPDATE CURRENT_TIMESTAMP`、字符集与排序规则等真实 MySQL 属性。SQLite 集成验收通过 `PRISMGO_SQLITE_TEST_DSN` 指向真实 SQLite 服务：加载 `demo/.dev/runtime/test.env` 后运行 `go test ./app/demo/schema -run TestSchemaDemoSQLiteIntegration -count=1`，其中 `...Batch` 接受两方言通用场景，`...DialectBoundary` 断言 MySQL 专属场景在 SQLite 明确报错，`...ColumnUnique` 校验字段级 `.Unique()` 在 SQLite 注册具名唯一索引并强制唯一。`spatial-types`（`geography`）与 `vector` 只做 SQLite 断言，因为当前 MySQL 服务端拒绝这两类列定义；`soft-deletes`、`morphs`、`nullable-morphs` 的普通索引依赖框架建表后补发 `ALTER TABLE ... ADD INDEX`。前批同时修复了框架两处 MySQL 列定义缺陷：把 `CHARACTER SET` / `COLLATE` 移到数据类型之后（原先追加在 `NOT NULL` 之后会被 MySQL 拒绝），并让字段级 `.Unique()` 生成按默认规则命名的唯一索引，使 `Unique(false)` 能按名删除。

状态含义：`已实现` 表示模块全部条目已落地；`进行中` 表示部分条目已落地；`计划中` 表示尚无已实现条目；`手工验证` 表示由安装器、Lens 等外部流程验证，因此不计入“剩余”数量。

## 本地开发工作区与 `dev` 脚本

PrismGo 本地开发采用三仓并列结构。**`demo/`、`framework/`、`docs/` 位于同一个工作区根目录下，互不嵌套**；三个目录各自是独立 Git 仓库，工作区根目录本身不是 Git 仓库。

### 安装与初始化

初始化需要 Go 1.26+、Git、GitHub SSH 访问权限和 `flock`。如需启动数据库、Redis 或 RabbitMQ，还需要 Docker Engine 与 Compose v2。

先创建一个空的工作区目录，只克隆 Demo 仓库；目录名必须是 `demo`：

```bash
mkdir workspace
cd workspace
git clone git@github.com:prismgo/demo.git demo
./demo/dev init
```

`init` 会自动完成剩余工作：

1. 将缺失的 `framework/` 和 `docs/` 克隆到 `demo/` 的同级目录。
2. 在缺失时创建 `demo/.env`，并为仍为空的 `APP_KEY` 生成本地随机密钥；已有非空配置不会被覆盖。
3. 创建或更新 `go.work`，加入 `demo/` 和 `framework/`。
4. 创建工作区 Agent 指令与项目 skill 链接。
5. 下载 Framework 的 Go 依赖。

### 三仓目录结构

`./demo/dev init` 执行完成后，目录结构如下：

```text
workspace/                           # 工作区根目录（名称自定，不是 Git 仓库）
├── demo/                            # 仓库 1：示例、Catalog、联调工具
│   ├── .git/
│   ├── dev                          # 工作区管理入口
│   └── go.mod                       # replace framework => ../framework
├── framework/                       # 仓库 2：框架源码与测试
│   └── .git/
├── docs/                            # 仓库 3：中英文用户文档
│   └── .git/
├── go.work                          # 同时加载 demo/ 与 framework/
├── AGENTS.md -> demo/AGENTS.md
└── CLAUDE.md -> demo/AGENTS.md
```

其中 `demo/go.mod` 通过 `../framework` 使用本地框架源码；Catalog 再把 Demo 示例和测试映射到 `docs/` 的文档章节。修改和提交必须进入对应仓库分别进行。

| 仓库 | 修改内容 | 不应放入 |
|---|---|---|
| `demo/` | Demo、Catalog、场景测试、本地依赖与开发工具 | 框架公开 API 的真实实现 |
| `framework/` | 框架实现、公开 API、组件与框架测试 | Demo 专属业务示例 |
| `docs/` | 面向使用者的中英文指南与 API 说明 | 框架或 Demo 实现代码 |

初始化后可检查环境、查看 Catalog，并启动 Demo HTTP 服务：

```bash
./demo/dev doctor              # 需要 Docker；检查依赖与 Compose 配置
go run ./demo demo:list        # 查看功能覆盖进度
go run ./demo serve            # 启动 Demo HTTP 服务
```

不需要外部服务时可以跳过 `doctor` 和 `dev up`，直接运行 Catalog、Demo 或默认测试。

### 常用 `dev` 命令

| 命令 | 用法 |
|---|---|
| `./demo/dev help` | 查看命令列表和可用服务名 |
| `./demo/dev up [service...]` | 启动全部或指定依赖，并等待健康检查 |
| `./demo/dev status` | 查看服务健康状态、本地端口和连接信息 |
| `./demo/dev logs [service]` | 持续查看全部或指定服务的日志 |
| `./demo/dev env` | 重新生成并输出 `demo/.dev/runtime/test.env` |
| `./demo/dev test [go-test-args]` | 启动依赖，并在 `framework/` 中运行指定集成测试 |
| `./demo/dev test-horizon` | 启动 Redis 与 RabbitMQ，运行 Horizon 双队列真实消费测试 |
| `./demo/dev down` | 停止服务但保留本地数据 |
| `./demo/dev reset [--yes]` | 删除全部本地测试数据并重建环境；请谨慎使用 |

可选服务名为 `mysql`、`postgres`、`sqlserver`、`redis`、`rabbitmq` 和 `sqlite`。通常只启动当前任务需要的依赖：

```bash
./demo/dev up mysql redis
./demo/dev status
./demo/dev logs redis
```

加载脚本生成的测试连接变量，或直接通过统一入口运行框架集成测试：

```bash
./demo/dev env
source demo/.dev/runtime/test.env

./demo/dev test ./queue/...
./demo/dev test-horizon
./demo/dev down
```

Demo 自身的默认测试使用隔离驱动，不需要启动 Docker：

```bash
cd demo
GOWORK=off go test . ./app/... ./bootstrap/... ./config/... ./database/... ./routes/...
```

`commands` 的迁移与 Seeder 场景需要专用 MySQL 测试库。先执行 `./demo/dev up mysql`，创建名称以 `prismgo_commands_` 开头的空数据库并授权测试用户，再设置 `PRISMGO_COMMANDS_MYSQL_TEST_DSN` 运行 `go test ./app/demo/commands -run TestCommandsDemoMySQLIntegration -v`。场景会在运行前后清空该专用库的表；不要指向共享的 `prismgo_test`。

`commands` 的队列场景使用真实 Redis。执行 `./demo/dev up redis`、加载 `demo/.dev/runtime/test.env` 后运行 `go test ./app/demo/commands -run TestCommandsDemoRedisIntegration -v`；场景使用独立键前缀并在结束时清理。

`session` 的 `redis-driver` 与 `redis-lock` 场景同样使用真实 Redis。加载 `demo/.dev/runtime/test.env` 后运行 `go test ./app/demo/session -run 'TestSessionDemoRedis(Driver|Lock)' -v`，或执行 `go run ./demo demo:session redis-driver --connection=redis` 与 `go run ./demo demo:session redis-lock --connection=redis`；场景使用唯一 `prismgo_demo_session_*` 前缀并在结束时清理键。

`redis` 的连接、命令、连接管理、事件、容器集成，以及 cache/queue/horizon 的 Redis 集成场景使用真实 Redis。加载 `demo/.dev/runtime/test.env` 后运行 `go test ./app/demo/redis -run 'TestRedisDemo' -v`（无 `PRISMGO_REDIS_TEST_URL` 时集成用例明确跳过），或执行 `go run ./demo demo:redis list` 与 `go run ./demo demo:redis strings --json`；集成场景从 `PRISMGO_REDIS_TEST_URL` 推导 host/port/database，证明默认连接与 `cache` 连接指向真实服务，通过 Facade 解析连接并分别使用唯一 `prismgo_demo_redis_*`、`prismgo_demo_redis_cache_*`、`prismgo_demo_redis_queue_*`、`prismgo_demo_redis_horizon_*` 前缀清理键。`redis` 的配置、连接生命周期、容器解析与 `horizon-config` 场景（`config`、`url`、`manager`、`purge`、`close`、`provider-registration`、`container-factory`、`event-sensitive-parameters` 等）为 hermetic，不需要 Redis。`horizon-config` 通过显式 ConfigReader 解析 `horizon` 配置，其余 Horizon 场景直接读写真实 Redis Store；cache/queue 场景分别在 `CACHE_STORE=redis`、`QUEUE_CONNECTION=redis` 下验证驱动、TTL、原子/批量/标签操作、ready list、delayed zset、阻塞 pop 与失败任务存储。

`route` 的 87 个场景为编译级或 hermetic，不依赖任何外部服务。运行 `go run ./demo demo:route list` 查看条目，或直接执行 `go run ./demo demo:route facade --json`、`go run ./demo demo:route url-escaping`、`go run ./demo demo:route api-resource --json`；第三批限流与调试场景可用 `go run ./demo demo:route throttle-route`、`go run ./demo demo:route list-command --json`、`go run ./demo demo:route provider-singleton` 验证。启动 `go run ./demo serve` 后请求 `/api/route-demo/users/42` 可验证真实 HTTP Server 挂载、`WhereNumber` 约束与命名路由 URL 生成（`/api/route-demo/users/abc` 返回 404），请求 `/api/route-demo/posts/7` 验证参数绑定、`/api/route-demo/photos/9` 验证 API 资源路由、`/api/route-demo/meta/42` 验证 `route.current` 注入与命名 URL 生成、`/api/route-demo/legacy` 验证重定向。

`service-provider` 的 42 个场景为编译级、hermetic 或 scenario，不依赖任何外部服务。运行 `go run ./demo demo:provider list` 查看条目，或用 `go run ./demo demo:provider singleton --json`、`go run ./demo demo:provider commands`、`go run ./demo demo:provider default-order`、`go run ./demo demo:provider deferred-resolution --json`、`go run ./demo demo:provider terminate-order --json`、`go run ./demo demo:provider full-lifecycle` 验证注册顺序、延迟加载与可终止提供者。启动 `go run ./demo serve` 后请求 `/api/provider-demo/greeting` 可验证 `AppServiceProvider.Register` 写入容器的服务在请求期仍可解析，返回 `{"message":"hello from provider"}`。

`commands` 的 `fresh-drop-types` 场景使用真实 PostgreSQL。执行 `./demo/dev up postgres`、加载 `demo/.dev/runtime/test.env` 后运行 `go test ./app/demo/commands -run TestCommandsDemoPostgresIntegration -v`；测试创建独立的 `prismgo_commands_` schema，验证枚举类型和表被清除后删除该 schema。

完整的服务地址、环境变量、安全提示和排障方法见[本地测试环境说明](.dev/docs/local-test-environment.md)。
