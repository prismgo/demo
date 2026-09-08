# 工作区边界

## 目录职责

| 路径 | 职责 | Git 边界 |
|---|---|---|
| `/` | PrismGo 本地 Demo、联调入口与本地依赖配置 | 非独立 Git 仓库 |
| `framework/` | `github.com/prismgo/framework` 的源码、测试与框架变更记录 | 独立仓库 |
| `docs/` | PrismGo 对外文档的中英文内容 | 独立仓库 |
| `.dev/docs/` | 本工作区的 Agent 开发说明 | 工作区文件 |

## 修改归属

- 框架能力、公开 API、组件实现和框架测试放在 `framework/`。
- 用户文档、指南和 API 使用说明放在 `docs/`。
- Demo 只用于复现、验证和展示框架用法，不在 Demo 中复制或绕过框架实现。
- 任务同时涉及源码与文档时，分别检查 `framework/`、`docs/` 的状态和差异。

## 本地依赖

根模块通过以下替换直接使用工作区源码：

```go
replace github.com/prismgo/framework => ./framework
```

不要用已发布版本验证尚未发布的本地框架改动。
