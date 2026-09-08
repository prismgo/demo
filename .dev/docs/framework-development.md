# 框架开发导航

## 开始前

1. 阅读 `framework/AGENTS.md`，它是框架仓库的具体开发规则。
2. 修改或查找 Go 代码前阅读 `framework/CODE_INDEX.md`，按索引定位包、文件和符号。
3. 先确认问题属于框架实现，而不是 Demo 配置或用法。

## 常用入口

| 目标 | 入口 |
|---|---|
| 应用启动与生命周期 | `framework/foundation/` |
| 依赖注入 | `framework/container/`、`framework/provider/` |
| HTTP 与路由 | `framework/http/`、`framework/route/` |
| CLI | `framework/console/`、`framework/kernel/`、`framework/cmd/` |
| 数据库 | `framework/database/` |
| 缓存、队列、事件 | `framework/cache/`、`framework/queue/`、`framework/event/` |
| 公共契约 | `framework/contracts/` |
| Facade 解析 | `framework/facade/` 及各组件的 `facade.go` |

完整组件清单和符号位置以 `framework/CODE_INDEX.md` 为准。

## 实现边界

- 保持 `contracts/<component>`、组件实现、Facade 三层职责清晰。
- 组件通过 ServiceProvider 接入应用生命周期；沿用相邻包的注册、启动和释放模式。
- 不擅自改变公共 API、加入兼容层或新增第三方依赖。
- 行为变更应添加同包测试；只修测试装配问题时不要改变生产语义。
- 新增包、文件或导出符号后同步维护 `framework/CODE_INDEX.md`。

