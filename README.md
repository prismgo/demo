<p align="center">
  <img src=".dev/assets/logo.png" width="250">
</p>

<div align="center">

**PrismGo —— 像写 Laravel 一样写 Go**

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
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

Catalog 当前覆盖 **29 个模块、87 个条目**：已实现 41 个、计划中 43 个、手工验证 3 个。以下是 README 更新时的快照；`app/demo/catalog/` 是唯一数据源，实时进度以 `demo:list` 输出为准。

| 模块 | 覆盖范围 | 进度 | 剩余 | 状态 |
|---|---|---:|---:|---|
| `cache` | 缓存存储、标签与锁 | 0/1 | 1 | 计划中 |
| `commands` | Console 命令发现 | 1/1 | 0 | 已实现 |
| `config` | 环境与应用配置 | 0/1 | 1 | 计划中 |
| `console` | Artisan 风格命令与终端 IO | 0/1 | 1 | 计划中 |
| `container` | 依赖绑定与解析 | 0/1 | 1 | 计划中 |
| `cookie` | Cookie 值与排队写入 | 0/1 | 1 | 计划中 |
| `database` | 数据库连接与连接池 | 0/1 | 1 | 计划中 |
| `encryption` | 应用数据加密 | 0/1 | 1 | 计划中 |
| `event` | 同步、异步与队列事件 | 0/1 | 1 | 计划中 |
| `exception` | 异常报告与渲染 | 0/1 | 1 | 计划中 |
| `facade` | 框架 Facade 访问 | 0/1 | 1 | 计划中 |
| `filesystem` | 本地、公共与云文件系统 | 0/1 | 1 | 计划中 |
| `horizon` | 队列监控与 Worker 管理 | 0/1 | 1 | 计划中 |
| `http-server` | HTTP 服务启动与生命周期 | 0/1 | 1 | 计划中 |
| `installation` | 应用安装流程 | 0/1 | — | 手工验证 |
| `lens` | Lens 开发流程 | 0/1 | — | 手工验证 |
| `lifecycle` | 应用启动与关闭 | 0/1 | 1 | 计划中 |
| `logger` | 多通道应用日志 | 0/1 | 1 | 计划中 |
| `queue` | 队列、任务与 Worker | 40/59 | 19 | 进行中 |
| `ratelimit` | 请求与操作限流 | 0/1 | 1 | 计划中 |
| `redis` | Redis 连接与操作 | 0/1 | 1 | 计划中 |
| `route` | HTTP 路由注册 | 0/1 | 1 | 计划中 |
| `schema` | 数据库 Schema 与迁移构建器 | 0/1 | 1 | 计划中 |
| `service-provider` | Service Provider 注册与生命周期 | 0/1 | 1 | 计划中 |
| `session` | 服务端 Session 存储 | 0/1 | 1 | 计划中 |
| `starter` | 生成应用的起始模板 | 0/1 | — | 手工验证 |
| `support` | 通用框架辅助函数 | 0/1 | 1 | 计划中 |
| `timer` | 定时任务定义 | 0/1 | 1 | 计划中 |
| `translation` | 应用翻译与复数处理 | 0/1 | 1 | 计划中 |

在工作区根目录查看实时地图与条目明细：

```bash
go run ./demo demo:list
go run ./demo demo:show queue
go run ./demo demo:show queue --status=planned
go run ./demo demo:show queue --level=integration
go run ./demo demo:list --json
```

状态含义：`已实现` 表示模块全部条目已落地；`进行中` 表示部分条目已落地；`计划中` 表示尚无已实现条目；`手工验证` 表示由安装器、Lens 等外部流程验证，因此不计入“剩余”数量。

## 本地开发工作区与 `dev` 脚本

PrismGo 本地开发采用三仓并列结构。**`demo/`、`framework/`、`docs/` 位于同一个工作区根目录下，互不嵌套**；三个目录各自是独立 Git 仓库，工作区根目录本身不是 Git 仓库。

### 安装与初始化

初始化需要 Go 1.25+、Git、GitHub SSH 访问权限和 `flock`。如需启动数据库、Redis 或 RabbitMQ，还需要 Docker Engine 与 Compose v2。

先创建一个空的工作区目录，只克隆 Demo 仓库；目录名必须是 `demo`：

```bash
mkdir workspace
cd workspace
git clone git@github.com:prismgo/demo.git demo
./demo/dev init
```

`init` 会自动完成剩余工作：

1. 将缺失的 `framework/` 和 `docs/` 克隆到 `demo/` 的同级目录。
2. 在缺失时创建 `demo/.env`，已有文件不会被覆盖。
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

./demo/dev test ./queue/... ./horizon/...
./demo/dev down
```

Demo 自身的默认测试使用隔离驱动，不需要启动 Docker：

```bash
cd demo
GOWORK=off go test . ./app/... ./bootstrap/... ./config/... ./database/... ./routes/...
```

完整的服务地址、环境变量、安全提示和排障方法见[本地测试环境说明](.dev/docs/local-test-environment.md)。
