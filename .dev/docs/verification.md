# 验证与交付

## 按范围验证

| 变更范围 | 最低验证 |
|---|---|
| 根 Demo 或联调配置 | 在根目录运行 `go test ./...` |
| 单个框架包 | 在 `framework/` 运行目标包测试、覆盖率和静态检查 |
| 跨框架组件 | 在 `framework/` 运行相关包测试，并视风险运行 `make ci` |
| 文档 | 核对链接、目录导航、代码示例和中英文同步情况 |

## 外部依赖测试

数据库、Redis、缓存、队列或 Horizon 变更需要真实服务时，先阅读 [本地测试环境](local-test-environment.md)，再按最小范围启动依赖：

```bash
./dev up <service...>
./dev status
source .dev/runtime/test.env
```

RabbitMQ 集成测试也可以通过统一入口运行：

```bash
./dev test ./queue/... ./horizon/...
```

报告中必须区分测试实际执行、因环境门禁而跳过、以及框架尚未支持对应 driver。容器健康只证明依赖可连接，不证明框架兼容性。

## 框架常用命令

```bash
cd framework
go test ./<package>/...
make covdata PACKAGES=./<package>
golangci-lint run --verbose ./<package>/...
gofmt -w <changed-go-files>
```

全量检查使用：

```bash
cd framework
make ci
```

## 交付检查

- 分别查看 `framework/` 与 `docs/` 的 Git 差异，不混淆两个仓库。
- 只报告实际执行的命令和结果；未执行的检查要说明原因。
- Go 代码变更需报告覆盖率范围、类型、总覆盖率和明显低覆盖区域。
- 检查本次变更是否产生死代码、孤立代码或兼容回退；发现后先报告，不擅自扩大清理范围。
