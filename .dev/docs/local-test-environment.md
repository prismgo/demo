# 本地测试环境

本工作区使用 Docker Compose 提供 MySQL、PostgreSQL、SQL Server、Redis 和 RabbitMQ，SQLite 由宿主机上的测试进程直接使用。所有基础设施配置位于 `dev/.dev/docker/`，统一入口是项目根目录的 `./dev/dev`。

## 快速开始

前置条件：Docker daemon 正在运行，并安装了支持 `docker compose` 的 Compose v2。

```bash
./dev/dev init
./dev/dev doctor
./dev/dev up
```

`./dev/dev up` 会启动全部容器、等待健康检查、初始化 SQL Server 测试数据库，并生成 `dev/.dev/runtime/test.env`。重复执行是安全的。

只启动当前任务需要的依赖：

```bash
./dev/dev up mysql redis
./dev/dev up rabbitmq
./dev/dev up sqlite
```

## 常用命令

| 命令 | 行为 |
|---|---|
| `./dev/dev init` | 创建根目录 Agent 指令链接，把 `dev/.agents/skills/` 下的项目 skills 软链到 `.agents`/`.claude`，克隆缺失仓库，初始化 `dev/.env` 和 `go.work`，并下载框架 Go 依赖 |
| `./dev/dev up [service...]` | 启动全部或指定服务并等待健康 |
| `./dev/dev status` | 查看容器状态、端口、账号和环境变量文件 |
| `./dev/dev logs [service]` | 持续查看全部或指定服务日志 |
| `./dev/dev env` | 重新生成并输出测试环境变量 |
| `./dev/dev test [go-test-args]` | 启动环境并在 `framework/` 中执行测试 |
| `./dev/dev down` | 停止容器，保留数据库和队列数据 |
| `./dev/dev reset` | 删除全部测试数据并重新创建；需要交互确认 |

服务名为 `mysql`、`postgres`、`sqlserver`、`redis`、`rabbitmq`、`sqlite`。

## 默认连接

所有容器端口默认监听 `0.0.0.0`。本机使用下表中的 `127.0.0.1` 连接；同一局域网内的设备将 `127.0.0.1` 替换为运行 Docker 的宿主机局域网 IP 即可。主机防火墙仍需允许对应端口。

| 服务 | 地址 | 数据库/账号 |
|---|---|---|
| MySQL | `127.0.0.1:13306` | `prismgo_test`，`prismgo/prismgo` |
| PostgreSQL | `127.0.0.1:15432` | `prismgo_test`，`prismgo/prismgo` |
| SQL Server | `127.0.0.1:11433` | `prismgo_test`，`sa/PrismGo_Test123!` |
| SQLite | `.dev/runtime/sqlite/prismgo_test.sqlite` | 文件数据库，无服务进程 |
| Redis | `127.0.0.1:16379` | DB 0，无密码 |
| RabbitMQ | `127.0.0.1:15673` | `prismgo/prismgo`，vhost `/` |
| RabbitMQ 管理界面 | `http://127.0.0.1:25673` | `prismgo/prismgo` |

这些账号只用于受信任局域网内的开发测试，禁止暴露到公网或用于生产环境。公共或不受信任网络下，应在 `.dev/docker/local.env` 设置 `BIND_ADDRESS=127.0.0.1`。

宿主机端口特意避开了各服务的标准端口。仍有端口冲突时：

```bash
cp .dev/docker/local.env.example .dev/docker/local.env
```

然后在 `local.env` 中覆盖端口或镜像版本。该文件不会提交。

## 测试环境变量

运行 `./dev/dev env` 后可加载统一连接变量：

```bash
source dev/.dev/runtime/test.env
```

变量包括：

- `PRISMGO_MYSQL_TEST_DSN`
- `PRISMGO_SQLITE_TEST_DSN`
- `PRISMGO_POSTGRES_TEST_DSN`
- `PRISMGO_SQLSERVER_TEST_DSN`
- `PRISMGO_REDIS_TEST_ADDR`、`PRISMGO_REDIS_TEST_URL`
- `PRISMGO_RABBITMQ_TEST_URL`

现有 RabbitMQ 集成测试已经读取 `PRISMGO_RABBITMQ_TEST_URL`。可直接运行：

```bash
./dev/dev test ./queue/... ./horizon/...
```

## Agent 测试规则

Agent 修改依赖外部服务的代码时，应执行以下流程：

1. 只启动任务需要的服务，例如 `./dev/dev up rabbitmq`。
2. 运行 `./dev/dev status`，确认目标容器为 `healthy`。
3. 通过 `./dev/dev test <packages>`，或先 `source dev/.dev/runtime/test.env` 再运行框架测试。
4. 报告实际启动的服务、执行的命令，以及测试通过、失败或跳过的数量。
5. 环境变量缺失导致的 `Skip` 不是集成测试通过；必须确认测试确实执行。
6. 普通清理使用 `./dev/dev down`；只有明确需要干净数据时才使用会删除卷的 `./dev/dev reset`。

当前框架生产实现正式打开 MySQL 和 SQLite。PostgreSQL、SQL Server 容器用于准备兼容性测试环境，但框架尚未引入对应 GORM driver，不能仅根据容器健康状态宣称框架已经支持它们。Redis 单元和组件测试目前大量使用 `miniredis`；是否使用真实 Redis，应以目标测试读取的环境变量和测试逻辑为准。

## 故障排查

- 端口被占用：在 `dev/.dev/docker/local.env` 覆盖对应宿主机端口，再运行 `./dev/dev up`。
- 容器未健康：运行 `./dev/dev logs <service>` 查看启动日志。
- 修改初始化账号后仍使用旧账号：初始化变量只对空数据卷生效；确认不需要旧数据后运行 `./dev/dev reset`。
- SQL Server 启动慢：首次启动需要更多时间和内存，脚本最多等待 240 秒。
- ARM 主机：SQL Server 默认固定为 `linux/amd64`，通常需要 Docker 的跨架构模拟，速度会比其他服务慢。
