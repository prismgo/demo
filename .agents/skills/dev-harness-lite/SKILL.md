---
name: dev-harness-lite
description: PrismGo 小功能与简单 bug 的轻量开发闭环（手动触发）：可选小卡留档、S1 澄清、S2 TDD、S3 并发验证与代码审查、S4 分仓提交归档。用户要求 harness、轻量工作流或任务卡时使用。
---

# PrismGo 轻量开发闭环（lite）

你是编排器。全程保持原有 `S0 → S1 → S2 → S3 → S4`；不得删减、合并或改名 S1–S4。
路径基准是同时包含 `demo/`、`framework/`、`docs/` 的 PrismGo 工作区根。以下命令均从工作区根执行：

```bash
W="bash demo/.agents/skills/dev-harness-lite/scripts/wf.sh"
D="bash demo/.agents/skills/dev-harness-lite/scripts/wipDiff.sh"
```

只按任务需要读取 `demo/AGENTS.md` 的地图索引。查看或改动 `framework/` 前必须先读
`framework/AGENTS.md` 与 `framework/CODE_INDEX.md`；涉及外部服务再读
`demo/.dev/docs/local-test-environment.md`；涉及文档再读 `demo/.dev/docs/documentation.md`；交付前读
`demo/.dev/docs/verification.md`。

**边界**：用户选了 lite 就按 lite 做完，不建议换别的 workflow。遇到多仓、公开契约、DB 迁移、金额、权限或
并发状态机等敏感面，只用一句话提示风险，然后继续。修改归属始终是：Demo/联调工具在 `demo/`，框架实现在
`framework/`，框架文档在 `docs/`；根目录不是 Git 仓库。框架行为变化必须同步 `docs/`，并保持
`demo/go.mod` 的本地 `replace github.com/prismgo/framework => ../framework`。

小卡只用于防失忆、防上下文漂移和保留关键决策，不是详细计划或证据档。不开卡仍争取单会话完成；开卡时以
`demo/.dev/_task/` 内的卡为续点，不依赖聊天上下文。该目录是本地、gitignored 的 harness 状态，不进入业务提交。

## S0 启动：一次问完四项

一次性询问：

1. **是否开小卡**：开（进入 board，完成后归档）/ 不开（不产生 harness 状态文件）。
2. **工作区模式**：普通 / worktree。普通使用各仓主检出；worktree 保持主检出不动。
3. **上下文限制模式**：启用（默认，上限 500K，可改 N）/ 不启用（`continuous`，连续做完不暂停）。
4. **是否记录耗时**：记录（默认）/ 不记录。记录后每步耗时落 `<id>.assets/timing.jsonl`，归档时汇总到卡的
   `## 耗时统计`。打点寄生在已有 `wf` 命令中，不额外增加回合。不开卡则此项自动作废。

严禁代选。拿到答案后，从需求生成 kebab-case `<slug>`，并确认所有受影响仓库（逗号分隔）。新任务在建卡或
开分支前，对每个受影响仓执行 `git status --short`；干净时才可 `git switch main` 和
`git pull --ff-only origin main`。失败即停，不 stash/reset，也不覆盖用户改动。续做已有卡则跳过，直接
`$W status <id>`。

若开小卡：

```bash
$W lite-new <slug> <demo,framework,docs> <owner> --workspace <normal|worktree> \
  --timing <on|off> --loop-mode <context|continuous> [--loop-limit 500]
```

命令只生成轻量卡并加入 board，不生成详细计划或证据目录。随后用 `wf fill` 写卡，完成后再 `$W branch <id>`，
然后按 S1 末尾强制暂停。分支固定为 `lite/<slug>`，每个受影响仓使用同名分支；多仓准备结果逐项写入
`branched_repos`，外部 Git 操作导致中断时修复对应仓后重跑即可续做。

若不开小卡：在每个受影响仓创建 `lite/<slug>` 分支；worktree 放在
`.worktrees/lite-<slug>/<repo>`。上下文限制模式无卡可落盘时，暂停前把“已完成 / 下一步 / 未决问题 / 各仓分支”
直接写进回复正文，供 compact 后续做。

需要外部服务时，用 `./demo/dev status` 和对应连接变量做环境门禁；纯单元测试不启动 Docker。变量缺失的真实集成
测试必须明确跳过，不能把依赖健康误报成框架验证通过。

## S1 澄清、DoD 与小卡定稿

一问一答，每次只问一个真实歧义，禁止脑补；提供三条较详细且互斥的方案，标出推荐方案和理由。形成 3–6 条
可执行 DoD，每条按验证方式标 `[test]` / `[curl]` / `[db]` / `[docs]`，圈定非目标，并明确将要测试的公开 seam。
把 DoD、非目标和 seam 读给用户确认后才实现；S2 必须使用已安装的 `tdd` skill，并遵循其 seam 确认规则。

开卡时分段填：

- **需求描述**：用户目标、范围、DoD 与非目标，短而可执行。
- **外部文档索引**：需求、issue、设计稿等链接；没有写“无”，不要复制全文。
- **Agent 设计摘要**：改哪些模块、核心做法、公开 seam 和测试切口；不写详细实施计划。
- **验收标准 DoD**：3–6 条可执行检查项。
- **Feature 清单**：可独立完成和验证的增量 `F1/F2...`，通常 1–3 项，不为形式切碎。
- **决策**：只记影响实现的选择和原因。

```bash
$W fill <id> 需求描述
$W fill <id> 外部文档索引
$W fill <id> "Agent 设计摘要"
$W fill <id> "验收标准 DoD"
$W fill <id> "Feature 清单"
$W fill <id> 决策
```

**写完卡强制暂停**：先 `$W branch <id>`，再一次运行
`$W status <id> && $W pause <id> "等 /compact 或 /clear 后续 S2"` 自检可续并打点，然后停止，只输出：
“小卡 `<id>` 已写完，请 `/compact` 或 `/clear` 后回来，我用 `$W status <id> --cleared` 续做。”
禁止同一上下文继续 S2；只有真的 compact/clear 后才加 `--cleared`。

## S2 TDD（主会话直接做）

先读取并使用 `tdd` skill。按 Feature 顺序做竖切片红 → 绿，一次只写一个行为测试和最小实现；重构留到 S3 review。
只跑受影响包，所有单测都经过输出过滤器并以 pipeline 退出码为准：

```bash
set -o pipefail; <包级测试命令> 2>&1 | bash demo/.agents/skills/dev-harness-lite/scripts/testFilter.sh
```

S2 禁止 `go test ./...` 等全量测试。主 agent、验收者和审查者的单测都走 filter；不绿不进 S3。

开卡时不得手改卡或只靠聊天记忆。开始 Feature、阶段切换、定位失败、完成 Feature、准备 compact 时用命令原子更新：

```bash
$W lite-start <id> <Ftag>
$W lite-progress <id> <S1|S2|S3|S4> "已完成：…；当前：…；下一步：…；阻塞：无/…"
$W lite-fail <id> <issue-key> "已证实根因和下一步"
$W lite-check <id> <Ftag> "<测试/验证结论>"
```

`lite-start` 设置当前 Feature；`lite-fail` 在整张卡内按 issue key 累计，同一问题第 3 次失败会在记录后返回非零并
要求交人，切换 Feature 或穿插其他 issue 都不会清零；`lite-check` 只允许在 S2 勾选已经 `lite-start` 的 Feature。
从 S3/S4 回到 S2 时，对已完成 Feature 再运行 `lite-start` 会先重新打开该项，修复后必须重新 `lite-check`。
所有 Feature 完成后才能进入 S3；`lite-progress` 只允许 S1→S2→S3→S4 及 S3/S4 回到修复阶段。
compact 或新会话后先 `$W status <id>`，只按未完成 Feature、最近进度、决策和 Handoff 续做。

## S3 并发验证 + 代码审查

准备同一份事实包：每个受影响仓的 base、完整 WIP diff、DoD、非目标、服务地址和包级测试命令。先用
`$D <repo-path> <base> /tmp/<repo>-<id>-review.patch` 生成不修改 Git index 的 tracked + untracked 完整 patch 和 hash，
再明确要求 `code-review` 使用 **working-tree/WIP target**；不得使用会漏掉 S2 未提交改动的 `<base>...HEAD`。开卡时用
`$W sect <id> <段标题前缀>` 只取必要段，不读整卡。预计超过 10K 字符的 diff、测试或日志落 `/tmp`，主链只保留
摘要与路径。

- `[test]`：主会话跑目标包验证；交付要求的全量验证留到 S4。
- `[curl]` / `[db]`：主会话验证，写操作必须复核最终状态。
- `[docs]`：检查对应中英文内容、链接、导航与示例。
- 同时调用已安装的 `code-review` skill 审查固定点到当前 WIP 的完整 diff；generator ≠ evaluator。该 skill 默认并发检查
  correctness、concurrency、performance，lite 的阻断轴保持原规则：只有 correctness 的 CONFIRMED 阻断 S4；
  其他轴发现的真实风险也必须修复或在 Handoff 明确记录。每轮完整 patch 落 `/tmp`；复审把上轮与当前 patch 的
  `diff -u` 作为增量输入，同时提供当前受影响文件。
- review 必须额外检查本次 diff 是否删除、改写或削弱既有测试断言；发现即按 correctness CONFIRMED 报告。
- 验收者和审查者不得把长原始输出返回主链：输出落 `/tmp`，只回 `PASS/FAIL`、核心原因、`file:line`、最小复现
  命令和原始输出路径，每项不超过 5 行。

可并行的验证与 review 同批启动，避免串行等待。任一验证失败或 correctness 有 CONFIRMED：合并错误一次说清，先
`$W lite-fail` 计数并记录已证实根因，再 `$W lite-progress` 回 S2，`lite-start` 重新打开对应 Feature 后修复并重跑。
同一问题第 3 轮仍不过就停止交人。

## S4 提交、合并与归档

先按 `demo/.dev/docs/verification.md` 对每个受影响仓执行最低验证；框架 Go 变更还要报告目标覆盖率和静态检查结果。
验证 PASS、review 阻断项清零后，分别精确 `git add` 和提交，commit subject 以 `<slug>` 开头，不 push。三个仓是
独立 Git 仓库，不在工作区根提交，也不把本地任务卡加入提交。

开卡时确认所有 Feature 已由 `lite-check` 打勾。提交后先进入 S4，再合并；归档必须最后执行：

```bash
$W lite-progress <id> S4 "验证与 review 已通过；各仓提交已完成；下一步：合并"
$W lite-merge <id>
# 仅 worktree 模式：$W wtclean <id>
$W fill <id> Handoff
$W lite-archive <id>
```

`lite-merge` 将每个受影响仓的 `lite/<slug>` 用 `--no-ff` 合回各自 main，并在 `merged_repos` 中逐仓记录成功结果；
中途冲突时保留已完成仓和失败续点，解决或 abort 冲突后重跑即可跳过已完成仓。worktree 模式必须随后运行
`wtclean`；`lite-archive` 会拒绝非 S4、未合并、dirty、未回到 main 或仍有 worktree 的任务。最后 `$W board`
重生看板。停在这里，不 push，等用户明确说“推”。

最终报告不超过 15 行：改动、文件、DoD 逐条结果、review、各仓 commit hash、未运行/跳过项和非目标。

## 铁律与省时省 token 机制

- S1 写完卡必须停下；状态只能用 `wf.sh` 改，普通进度用 `lite-progress`，Feature 完成用 `lite-check`。
- 停下等用户前把 `pause` 串进收口命令，不单发计时命令：`$W status <id> && $W pause <id> "<等什么>"`。
- 验证先于宣告胜利；测试原始输出、review 长输出及大 diff 不进主上下文，落盘后只读摘要。
- **上下文预算**：宿主报告当前 token 用量时，把预算检查寄生在进度命令同一工具回合：
  `$W lite-progress ... && $W budget <id> <used>`。`loop_mode=context` 达 `loop_limit` 会返回非零；随后完成当前 S 步即停，
  用 `$W report <id> budget && $W pause <id> "等 compact/clear"` 留续点，不进入下一 S 步或 Feature。不开卡则把同样续点写进回复。
  选 `continuous` 才连续做完。
- **一次工具回合完成机械连招**：下一条命令不需要看上一条输出才能决定是否执行时，用 `&&` 串联；有退出码分支
  时不要硬串。可并行的只读检查、验证和 review 同批派发。
- **并发安全**：所有 `wf.sh` 状态、看板和 Git 编排命令共用任务目录锁；并行验证可以同时跑，但状态更新仍由
  `wf.sh` 串行落盘。`demo/dev init` 也按工作区加锁，避免并发初始化留下半套链接。
- **渐进读取**：先 `rg`/`rg --files` 定位，再读命中附近；续做先 `status`，取卡片段用 `sect`，不整卡重读；
  框架入口先用 `CODE_INDEX.md`，避免扫仓。
- **失败收敛**：同一轮把所有已知错误汇总后再修；复审只看增量；同一问题最多 3 轮。
- **耗时零额外回合**：计时事件寄生在 `lite-new/lite-progress/lite-start/lite-fail/lite-check/budget/lite-merge/pause/report/lite-archive`；等待时间由
  `pause` 剔除，归档自动汇总。计时失败不得改变主命令输出和退出码。
- 不使用无人值守 `codex exec`/`claude -p` 批量跑任务；不自动 stash/reset，不覆盖用户改动，不 push。
