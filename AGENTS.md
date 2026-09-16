# PrismGo 本地开发工作区

项目根目录用于组织 PrismGo 本地联调工作区：`demo/` 是本地 Demo 与开发工具仓库，`framework/` 是框架源码仓库，`docs/` 是框架文档仓库。本文件位于 `demo/AGENTS.md`，并由 `./demo/dev init` 链接到项目根目录。

## 铁律

| 规则 | 要求 |
|---|---|
| 修改归属 | Demo 与联调工具改在 `demo/`；框架实现改在 `framework/`；框架文档改在 `docs/`；项目根目录只组织这三个仓库。 |
| 就近遵循 | 进入子目录后，同时遵循其中更具体的 `AGENTS.md`；冲突时以更深层规则为准。 |
| 先读再改 | 按下表只读取与任务相关的说明；修改框架 Go 代码前先读 `framework/CODE_INDEX.md`。 |
| Go 风格前置 | 任务只要会编写或修改 Go 代码，必须在第一次改动前完整读取 `demo/.dev/docs/go-style.md`，并在实现过程中直接将其作为验收条件；显式规范高于相邻旧代码，旧违规不得作为继续复制的先例。 |
| 保持本地联调 | `demo/go.mod` 中的 `github.com/prismgo/framework` 必须指向同级的本地 `framework/`。 |
| Catalog 进度同步 | 修改 `demo/app/demo/catalog/` 中会影响条目或进度统计的内容时，必须在同一变更中同步更新 `demo/README.md` 与 `demo/README_en.md` 的 Catalog 总数、模块进度和剩余数量，并用 `demo:list` 核对一致性。 |
| 文档同步 | `framework/` 功能变更时，同步更新 `docs/` 中的对应文档。 |
| 分仓处理 | 根目录不是提交仓库；`demo/`、`framework/` 与 `docs/` 是三个独立 Git 仓库，检查和提交必须分别执行。 |

## Go 实现阶段风格闭环

Go 风格必须在实现阶段就地保证，不得留到独立 code review 才发现。每个涉及 Go 代码的任务执行以下闭环：

1. 写代码前，根据 `go-style.md` 主动确认本次变更涉及的命名、API、错误、context、并发、注释和测试规则；不得仅依赖记忆或相邻文件。
2. 实现时持续按规范约束新增代码。复用相邻模式前先确认它不与显式规则冲突；发现旧代码违规时，不扩大无关清理，但新代码必须使用合规写法。
3. 每完成一个逻辑变更，立即对本次改动的 Go 文件运行 `gofmt` 和 `goimports`，并检查当前 diff，确认没有无必要别名、错误丢失上下文、未收尾 goroutine、无说明的可变全局状态、缺失导出注释，以及缺少 actual/want/定位上下文的测试失败信息。如 `goimports` 不可用，必须手工完成同等的导入分组与别名检查，并在交付中报告工具缺失。

## 开发命令

在项目根目录执行初始化：

```bash
./demo/dev init
```

该命令会创建根目录的 Agent 指令链接，并把 `demo/.agents/skills/` 中的项目 skills（当前包含 `dev-harness-lite` 与 `code-review`）软链到根目录的 `.agents/skills/` 与 `.claude/skills/`；随后克隆缺失的 `framework/` 与 `docs/` 仓库，将 `demo/.env.example` 复制为缺失的 `demo/.env`，为仍为空的 `APP_KEY` 生成本地随机密钥，创建或更新本地 `go.work`，并下载框架 Go 依赖。已有非空配置不会被覆盖。

初始化完成后，从项目根目录启动 Demo HTTP 服务器：

```bash
go run ./demo serve
```

查看所有可执行命令及参数入口：

```bash
go run ./demo list
```

需要数据库、Redis 或队列等本地服务时，先按任务范围启动依赖，例如 `./demo/dev up mysql redis rabbitmq`；服务状态与连接变量通过 `./demo/dev status` 和 `./demo/dev env` 查看。

## Demo 示例与测试

- `app/demo/catalog/` 维护“Framework 版本 → 功能模块 → 文档章节 → 示例 → 测试”的覆盖目录；用 `go run ./demo demo:list` 查看模块进度总览，用 `go run ./demo demo:show <feature>` 查看子功能详情，并可按 `--level`、`--status` 过滤或使用 `--json` 输出。
- `app/demo/testing/` 提供隔离测试 Application，默认使用 SQLite、memory/file cache、sync queue、file session 和临时目录，不读取开发环境的 `.env`。
- `app/models/`、`app/repositories/`、`app/services/` 与 `database/migrations/` 中的订单示例用于展示模型、仓储、事务和业务服务组合；当前以单元测试和场景测试为主要使用入口。

默认验证不需要 Docker：

```bash
cd demo
GOWORK=off go test . ./app/... ./bootstrap/... ./config/... ./database/... ./routes/...
```

真实集成测试按需通过 `PRISMGO_MYSQL_TEST_DSN`、`PRISMGO_REDIS_TEST_ADDR`、`PRISMGO_REDIS_TEST_URL` 或 `PRISMGO_RABBITMQ_TEST_URL` 启用；变量缺失时相关测试应明确跳过。

## 发布上线流程（framework 与 prismgo）

发布通过 push `vX.Y.Z` tag 触发各仓的 GitHub Actions（Goreleaser）；`framework/` 与 `prismgo/` 是独立仓库，须分别提交、分别打 tag。

本流程仅在用户明确要求发布新版本时启动。日常提交推送（修 bug、升依赖、改 CI 等）与发布无关：不得检测 tag 或 Release 状态，不得主动提醒或建议打 tag，也不要输出“按发布铁律本次不打 tag”这类发布确认。只有用户明确说“发布新版本”时，才按“验证 → 提交推送 → 等用户确认 Actions 通过 → 打 tag 发布”执行，且提交推送后不得立即打 tag。

### framework 发布

1. 全量验证：`cd framework && go test ./...`（有跨组件风险时 `make ci`），并确认工作区干净。
2. 改版本号：把 `framework/version/version.go` 的 `const Framework` 设为 `x.y.z`。Release workflow 强制 `tag 去掉 v` 必须等于该常量，否则校验失败并自动删除 tag。
3. 提交推送：`gofmt` 后提交（`bump version to x.y.z`）并 `git push origin main`。
4. 等待确认：push 后不要立即打 tag；等用户在 GitHub 验证 Actions 通过并明确通知发布，收到确认后再 `git tag vx.y.z && git push origin vx.y.z`，触发 Goreleaser。
5. 校验：`https://github.com/prismgo/framework/releases.atom` 出现该 tag，且模块代理已索引（`https://proxy.golang.org/github.com/prismgo/framework/@v/vx.y.z.mod`）。

### prismgo 发布

1. 升级框架依赖：`cd prismgo && GOWORK=off go get github.com/prismgo/framework@vx.y.z`。
2. 清理依赖：`GOWORK=off go mod tidy`，移除已摘出为扩展包的 oss/rabbitmq 等未用间接依赖。
3. 复核 `go.mod`：`go` 指令保持 1.25+，关键间接依赖未被顶到需要更高工具链。
4. 运行验证：`GOWORK=off go run ./` 正常，且 `GOWORK=off go build ./...`、`go vet ./...`、`go test ./...` 通过。
5. 提交推送：按仓库习惯提交（`chore(deps): ...`）并 `git push origin main`。
6. 等待确认：push 后不要立即打 tag；等用户在 GitHub 验证 Actions 通过并明确通知发布，收到确认后再 `git tag vx.y.z && git push origin vx.y.z` 触发 Goreleaser，并校验 `https://github.com/prismgo/prismgo/releases.atom`。

### 发布铁律与教训

| 规则 | 要求 |
|---|---|
| 按需发布 | 只有用户明确要求发布新版本时才进入发布流程并检查 tag/Release；日常提交推送不得检测 tag、不得主动提醒或建议发布。 |
| 已发布版本不可变 | Go module 代理对已缓存版本只读：一旦 `proxy.golang.org/.../@v/vX.Y.Z.info` 固定了 `Origin.Hash`，重打或移动 tag 不会改变 `go get` 结果。发错版本只能新发 patch（如 v0.3.1），禁止复用或移动已发布 tag。 |
| 先推送再打 tag | tag 必须指向已 push 到 `main` 的提交；未推送就发布会造成 Release 与源码不一致。 |
| 待确认再发布 | 进入发布流程后，push 后不得立即打 tag，须等用户验证 GitHub Actions 通过并明确通知发布后再 tag。 |
| 版本号先对齐 | framework 打 tag 前必须已把 `version.Framework` 改为对应版本，否则 workflow 校验失败并自动删 tag。 |
| 警惕 go 指令漂移 | 依赖升级（含 dependabot）可能因某依赖要求更高工具链而自动抬高 `go` 指令（例如 x/term v0.46 要求 go 1.26）；合并前需确认符合项目 Go 版本策略。 |
| 发布通道 | 发布由 tag push 触发云端 Goreleaser，本地无需 `gh` 或 token；若要修改或删除已发布的 Release，需在 GitHub UI 或用带 token 的 API 操作。 |
| 分仓独立 | `framework/` 与 `prismgo/` 各自 commit、push、tag，不得混在一次提交里。 |

## 地图索引

| 何时读取 | 文档 | 用于确认 |
|---|---|---|
| 任务开始时无法确定应修改 Demo、框架还是文档，或需要处理本地依赖与 Git 状态时 | [`demo/.dev/docs/workspace.md`](demo/.dev/docs/workspace.md) | 三个仓库的职责边界、本地联调方式和独立仓库约束 |
| 准备编写、修改或审查 Go 代码时 | [`demo/.dev/docs/go-style.md`](demo/.dev/docs/go-style.md) | Go 命名、导入、API、错误处理、并发、测试及变更纪律 |
| 准备查看、搜索、修改或验证 `framework/` 内任何内容时，必须先读 | [`framework/AGENTS.md`](framework/AGENTS.md) | 框架仓库的铁律，以及架构、Go 开发、测试和交付细则的进一步索引 |
| 开发或排查框架能力，需要从工作区角度判断入口、代码层次或 Demo 联调方式时 | [`demo/.dev/docs/framework-development.md`](demo/.dev/docs/framework-development.md) | 框架常用入口、三层结构、源码修改边界及 `CODE_INDEX.md` 使用要求 |
| 修改或测试数据库、Redis、缓存、队列、Horizon 等依赖外部服务的能力时 | [`demo/.dev/docs/local-test-environment.md`](demo/.dev/docs/local-test-environment.md) | `./demo/dev` 启停方式、测试连接变量、真实集成测试和 Skip 判定规则 |
| 准备修改 `docs/`，或框架行为变化可能需要同步用户文档与双语内容时 | [`demo/.dev/docs/documentation.md`](demo/.dev/docs/documentation.md) | 文档仓库结构、内容归属、源码核对和中英文同步要求 |
| 修改完成准备验证或交付，或需要判断应在哪个仓库运行哪些检查时 | [`demo/.dev/docs/verification.md`](demo/.dev/docs/verification.md) | Demo、框架和文档的最低验证范围，以及结果报告与分仓检查要求 |
