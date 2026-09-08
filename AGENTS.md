# PrismGo 本地开发工作区

本目录用于联调 PrismGo：根目录是本地 Demo，`framework/` 是框架源码仓库，`docs/` 是框架文档仓库。

## 铁律

| 规则 | 要求 |
|---|---|
| 修改归属 | 框架实现改在 `framework/`；框架文档改在 `docs/`；根目录只承载 Demo 与联调配置。 |
| 就近遵循 | 进入子目录后，同时遵循其中更具体的 `AGENTS.md`；冲突时以更深层规则为准。 |
| 先读再改 | 按下表只读取与任务相关的说明；修改框架 Go 代码前先读 `framework/CODE_INDEX.md`。 |
| 保持本地联调 | 根模块的 `github.com/prismgo/framework` 必须通过 `go.mod` 指向本地 `framework/`。 |
| 文档同步 | `framework/` 功能变更时，同步更新 `docs/` 中的对应文档。 |
| 分仓处理 | 根目录不是提交仓库；`framework/` 与 `docs/` 是独立 Git 仓库，检查和提交必须分别执行。 |

## 地图索引

| 何时读取 | 文档 | 用于确认 |
|---|---|---|
| 任务开始时无法确定应修改 Demo、框架还是文档，或需要处理本地依赖与 Git 状态时 | [`.dev/docs/workspace.md`](.dev/docs/workspace.md) | 三个工作区的职责边界、本地联调方式和独立仓库约束 |
| 准备查看、搜索、修改或验证 `framework/` 内任何内容时，必须先读 | [`framework/AGENTS.md`](framework/AGENTS.md) | 框架仓库的铁律，以及架构、Go 开发、测试和交付细则的进一步索引 |
| 开发或排查框架能力，需要从工作区角度判断入口、代码层次或 Demo 联调方式时 | [`.dev/docs/framework-development.md`](.dev/docs/framework-development.md) | 框架常用入口、三层结构、源码修改边界及 `CODE_INDEX.md` 使用要求 |
| 修改或测试数据库、Redis、缓存、队列、Horizon 等依赖外部服务的能力时 | [`.dev/docs/local-test-environment.md`](.dev/docs/local-test-environment.md) | `./dev` 启停方式、测试连接变量、真实集成测试和 Skip 判定规则 |
| 准备修改 `docs/`，或框架行为变化可能需要同步用户文档与双语内容时 | [`.dev/docs/documentation.md`](.dev/docs/documentation.md) | 文档仓库结构、内容归属、源码核对和中英文同步要求 |
| 修改完成准备验证或交付，或需要判断应在哪个仓库运行哪些检查时 | [`.dev/docs/verification.md`](.dev/docs/verification.md) | Demo、框架和文档的最低验证范围，以及结果报告与分仓检查要求 |
