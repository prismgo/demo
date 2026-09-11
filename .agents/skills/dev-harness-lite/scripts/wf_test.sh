#!/usr/bin/env bash
set -euo pipefail

source_skill="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fixture="$(mktemp -d)"
trap 'rm -rf "${fixture}"' EXIT

skill="${fixture}/demo/.agents/skills/dev-harness-lite"
mkdir -p "$(dirname "${skill}")" "${fixture}/demo/.dev"
cp -R "${source_skill}" "${skill}"
wf="${skill}/scripts/wf.sh"
wip_diff="${skill}/scripts/wipDiff.sh"

output="$(bash "${wf}" lite-new cache-contract framework,docs codex --workspace normal --timing on --loop-mode context --loop-limit 500)"
[[ "${output}" == *"0001-cache-contract.md"* ]]
card="${fixture}/demo/.dev/_task/tasks/0001-cache-contract.md"
[[ -f "${card}" ]]
grep -Fq 'repos: [framework, docs]' "${card}"
grep -Fq 'stage: S1' "${card}"
grep -Fq 'branch: lite/cache-contract' "${card}"
grep -Fq 'active_feature: -' "${card}"
grep -Fq 'attempt_key: -' "${card}"
grep -Fq 'attempts: -' "${card}"
grep -Fq 'budget_used: 0' "${card}"
grep -Fq 'pause_state: none' "${card}"

if bash "${wf}" lite-new Bad_Slug demo >/dev/null 2>&1; then echo "invalid slug accepted" >&2; exit 1; fi
if bash "${wf}" lite-new bad-repo framework,../api >/dev/null 2>&1; then echo "repository traversal accepted" >&2; exit 1; fi
if bash "${wf}" lite-new bad-mode demo --workspace other >/dev/null 2>&1; then echo "invalid workspace accepted" >&2; exit 1; fi

printf '%s\n' '缓存读取遵循公开契约；非目标：不改数据库。' | bash "${wf}" fill 0001 需求描述 >/dev/null
printf '%s\n' '无' | bash "${wf}" fill 0001 外部文档索引 >/dev/null
printf '%s\n' '在公开 cache seam 做竖切片 TDD。' | bash "${wf}" fill 0001 'Agent 设计摘要' >/dev/null
printf '%s\n' '- [ ] 目标包测试通过 [test]' '- [ ] 中英文文档同步 [docs]' | bash "${wf}" fill 0001 '验收标准 DoD' >/dev/null
printf '%s\n' '- [ ] F1 缓存契约行为' '- [ ] F2 文档同步' | bash "${wf}" fill 0001 'Feature 清单' >/dev/null
printf '%s\n' '采用现有公开接口，避免新增抽象。' | bash "${wf}" fill 0001 决策 >/dev/null
printf '%s\n' '当前待进入 S2。' | bash "${wf}" fill 0001 Handoff >/dev/null
if bash "${wf}" lite-check 0001 F1 'bypass TDD' >/dev/null 2>&1; then
    echo "lite-check accepted a feature before S2/lite-start" >&2
    exit 1
fi

for repo in framework docs; do
    mkdir -p "${fixture}/${repo}"
    git -C "${fixture}/${repo}" init -q -b main
    git -C "${fixture}/${repo}" config user.email test@example.com
    git -C "${fixture}/${repo}" config user.name test
    printf 'fixture\n' >"${fixture}/${repo}/README.md"
    git -C "${fixture}/${repo}" add README.md
    git -C "${fixture}/${repo}" commit -qm init
done

# Branch setup must preflight every affected repository before switching any of
# them, otherwise a dirty later repository leaves the workspace half-switched.
printf 'local change\n' >>"${fixture}/docs/README.md"
if bash "${wf}" branch 0001 >/dev/null 2>&1; then
    echo "branch accepted a dirty affected repository" >&2
    exit 1
fi
[[ "$(git -C "${fixture}/framework" branch --show-current)" == 'main' ]]
git -C "${fixture}/docs" restore README.md

occupied="${fixture}/occupied-docs"
git -C "${fixture}/docs" worktree add -q -b lite/cache-contract "${occupied}" main
if bash "${wf}" branch 0001 >/dev/null 2>&1; then
    echo "branch ignored an existing worktree using the feature branch" >&2
    exit 1
fi
[[ "$(git -C "${fixture}/framework" branch --show-current)" == 'main' ]]
git -C "${fixture}/docs" worktree remove "${occupied}"

bash "${wf}" branch 0001 >/dev/null
[[ "$(git -C "${fixture}/framework" branch --show-current)" == 'lite/cache-contract' ]]
[[ "$(git -C "${fixture}/docs" branch --show-current)" == 'lite/cache-contract' ]]

printf 'tracked WIP\n' >>"${fixture}/framework/README.md"
printf 'untracked WIP\n' >"${fixture}/framework/new-file.txt"
patch_file="${fixture}/review.patch"
patch_output="$(bash "${wip_diff}" "${fixture}/framework" main "${patch_file}")"
[[ "${patch_output}" == *"patch=${patch_file}"* ]]
grep -Fq 'tracked WIP' "${patch_file}"
grep -Fq 'untracked WIP' "${patch_file}"
[[ -z "$(git -C "${fixture}/framework" diff --cached)" ]]
git -C "${fixture}/framework" restore README.md
rm "${fixture}/framework/new-file.txt"

if bash "${wf}" lite-progress 0001 S3 'skip S2' >/dev/null 2>&1; then
    echo "stage machine allowed S1 to skip S2" >&2
    exit 1
fi
if bash "${wf}" lite-progress 0001 S2 'continued without compact' >/dev/null 2>&1; then
    echo "S2 started before the mandatory S1 pause/clear" >&2
    exit 1
fi
bash "${wf}" pause 0001 'waiting for compact after S1' >/dev/null
bash "${wf}" status 0001 --cleared >/dev/null
bash "${wf}" lite-progress 0001 S2 '已完成：S1；当前：F1 red；下一步：green；阻塞：无' >/dev/null
bash "${wf}" lite-start 0001 F1 >/dev/null
grep -Fq 'active_feature: F1' "${card}"
status_output="$(bash "${wf}" status 0001)"
[[ "${status_output}" == *"active=F1"* ]]
bash "${wf}" lite-fail 0001 cache-miss 'first failed cycle' >/dev/null
bash "${wf}" lite-fail 0001 other-issue 'different failed cycle' >/dev/null
bash "${wf}" lite-fail 0001 cache-miss 'second failed cycle' >/dev/null
if bash "${wf}" lite-fail 0001 cache-miss 'third failed cycle' >/dev/null 2>&1; then
    echo "third failure did not stop the workflow" >&2
    exit 1
fi
grep -Fq 'attempt_key: cache-miss' "${card}"
grep -Fq 'attempts: cache-miss=3,other-issue=1' "${card}"
bash "${wf}" lite-check 0001 F1 '目标包测试 PASS' >/dev/null
grep -Fq 'active_feature: -' "${card}"
grep -Fq 'attempts: cache-miss=3,other-issue=1' "${card}"
grep -Fq -- '- [x] F1 缓存契约行为' "${card}"
grep -Fq -- '- [ ] F2 文档同步' "${card}"
grep -Fq 'stage: S2' "${card}"
if bash "${wf}" lite-check 0001 F1 duplicate >/dev/null 2>&1; then echo "duplicate check accepted" >&2; exit 1; fi
if bash "${wf}" lite-archive 0001 >/dev/null 2>&1; then echo "incomplete card archived" >&2; exit 1; fi
if bash "${wf}" lite-progress 0001 S3 'F2 is incomplete' >/dev/null 2>&1; then
    echo "stage machine entered S3 with an incomplete feature" >&2
    exit 1
fi
bash "${wf}" lite-start 0001 F2 >/dev/null
bash "${wf}" lite-check 0001 F2 '文档检查 PASS' >/dev/null

bash "${wf}" sect 0001 Feature | grep -Fq 'F2 文档同步'
status_output="$(bash "${wf}" status 0001)"
[[ "${status_output}" == *"features=2/2"* ]] || { echo "unexpected status: ${status_output}" >&2; exit 1; }
bash "${wf}" budget 0001 499 >/dev/null
if bash "${wf}" budget 0001 500 >/dev/null 2>&1; then
    echo "context budget limit was not enforced" >&2
    exit 1
fi
grep -Fq 'budget_used: 500' "${card}"
bash "${wf}" pause 0001 'waiting for compact' >/dev/null
bash "${wf}" status 0001 --cleared >/dev/null
bash "${wf}" lite-progress 0001 S3 '已完成：F1；当前：review；下一步：S4；阻塞：无' >/dev/null
bash "${wf}" lite-fail 0001 review-cache 'review found a regression' >/dev/null
bash "${wf}" lite-progress 0001 S2 '回到 F1 修复 review 问题' >/dev/null
bash "${wf}" lite-start 0001 F1 >/dev/null
grep -Fq -- '- [ ] F1 缓存契约行为' "${card}"
bash "${wf}" lite-check 0001 F1 'review 修复测试 PASS' >/dev/null
bash "${wf}" lite-progress 0001 S3 '增量复审 PASS' >/dev/null
if bash "${wf}" report 0001 arbitrary >/dev/null 2>&1; then
    echo "report accepted an unknown terminal reason" >&2
    exit 1
fi
bash "${wf}" report 0001 done >/dev/null
timing_output="$(bash "${wf}" timing 0001)"
[[ "${timing_output}" == *"事件"* ]] || { echo "unexpected timing: ${timing_output}" >&2; exit 1; }

if bash "${wf}" lite-archive 0001 >/dev/null 2>&1; then
    echo "card archived before entering S4 and merging branches" >&2
    exit 1
fi

printf 'delivered\n' >>"${fixture}/framework/README.md"
git -C "${fixture}/framework" add README.md
git -C "${fixture}/framework" commit -qm 'cache-contract delivery'

docs_main="${fixture}/docs-main"
git -C "${fixture}/docs" worktree add -q "${docs_main}" main
printf 'main delivery\n' >"${docs_main}/README.md"
git -C "${docs_main}" add README.md
git -C "${docs_main}" commit -qm 'main-side delivery'
git -C "${fixture}/docs" worktree remove "${docs_main}"
printf 'feature delivery\n' >"${fixture}/docs/README.md"
git -C "${fixture}/docs" add README.md
git -C "${fixture}/docs" commit -qm 'cache-contract delivery'
bash "${wf}" lite-progress 0001 S4 '验证与 review 已通过；各仓提交已完成；下一步：合并' >/dev/null
if bash "${wf}" lite-archive 0001 >/dev/null 2>&1; then
    echo "card archived before feature branches were merged" >&2
    exit 1
fi
if bash "${wf}" lite-merge 0001 >/dev/null 2>&1; then
    echo "expected the docs merge conflict" >&2
    exit 1
fi
grep -Fq 'merged_repos: [framework]' "${card}"
grep -Fq 'S4 merge FAIL：docs' "${card}"
git -C "${fixture}/docs" merge --abort
git -C "${fixture}/docs" switch lite/cache-contract >/dev/null
if ! git -C "${fixture}/docs" merge --no-edit main >/dev/null 2>&1; then
    printf 'resolved delivery\n' >"${fixture}/docs/README.md"
    git -C "${fixture}/docs" add README.md
    git -C "${fixture}/docs" commit -qm 'cache-contract resolve main'
fi
git -C "${fixture}/docs" switch main >/dev/null
bash "${wf}" lite-merge 0001 >/dev/null
grep -Fq 'merged_repos: [framework, docs]' "${card}"
bash "${wf}" lite-archive 0001 >/dev/null
archived="${fixture}/demo/.dev/_task/archive/0001-cache-contract.md"
[[ -f "${archived}" && ! -e "${card}" ]]
grep -Fq 'status: done' "${archived}"
grep -Fq '事件 ' "${archived}"
[[ -f "${fixture}/demo/.dev/_task/archive/0001.assets/timing.jsonl" ]]
grep -Fq '| 0001 | cache-contract | framework, docs | normal | S4 | 2/2 | done |' "${fixture}/demo/.dev/_task/board.md"
bash "${wf}" board --check >/dev/null

# Worktree delivery is not complete until the feature worktree is removed.
bash "${wf}" lite-new worktree-flow framework codex --workspace worktree --timing off --loop-mode continuous >/dev/null
card2="${fixture}/demo/.dev/_task/tasks/0002-worktree-flow.md"
printf '%s\n' '验证 worktree 交付门禁；非目标：不改公开 API。' | bash "${wf}" fill 0002 需求描述 >/dev/null
printf '%s\n' '无' | bash "${wf}" fill 0002 外部文档索引 >/dev/null
printf '%s\n' '通过 CLI seam 验证合并和清理顺序。' | bash "${wf}" fill 0002 'Agent 设计摘要' >/dev/null
printf '%s\n' '- [ ] worktree 已合并并清理 [test]' | bash "${wf}" fill 0002 '验收标准 DoD' >/dev/null
printf '%s\n' '- [ ] F1 worktree 交付' | bash "${wf}" fill 0002 'Feature 清单' >/dev/null
printf '%s\n' '使用独立 worktree。' | bash "${wf}" fill 0002 决策 >/dev/null
printf '%s\n' '等待 S2。' | bash "${wf}" fill 0002 Handoff >/dev/null
bash "${wf}" branch 0002 >/dev/null
wt="${fixture}/.worktrees/0002-worktree-flow/framework"
[[ "$(git -C "${wt}" branch --show-current)" == 'lite/worktree-flow' ]]
bash "${wf}" pause 0002 'waiting for compact after S1' >/dev/null
bash "${wf}" status 0002 --cleared >/dev/null
bash "${wf}" lite-progress 0002 S2 '进入 F1' >/dev/null
bash "${wf}" lite-start 0002 F1 >/dev/null
printf 'worktree\n' >>"${wt}/README.md"
git -C "${wt}" add README.md
git -C "${wt}" commit -qm 'worktree-flow delivery'
bash "${wf}" lite-check 0002 F1 'worktree 测试 PASS' >/dev/null
bash "${wf}" lite-progress 0002 S3 'review PASS' >/dev/null
bash "${wf}" lite-progress 0002 S4 '准备合并' >/dev/null
bash "${wf}" lite-merge 0002 >/dev/null
if bash "${wf}" lite-archive 0002 >/dev/null 2>&1; then
    echo "worktree card archived before cleanup" >&2
    exit 1
fi
bash "${wf}" wtclean 0002 >/dev/null
bash "${wf}" lite-archive 0002 >/dev/null
[[ -f "${fixture}/demo/.dev/_task/archive/0002-worktree-flow.md" && ! -e "${wt}" ]]
grep -Fq '不记录' "${fixture}/demo/.dev/_task/archive/0002-worktree-flow.md"

pids=()
for number in 1 2 3 4 5 6; do
    bash "${wf}" lite-new "concurrent-${number}" demo codex --timing off --loop-mode continuous >/dev/null &
    pids+=("$!")
done
for pid in "${pids[@]}"; do wait "${pid}"; done
[[ "$(find "${fixture}/demo/.dev/_task/tasks" -maxdepth 1 -name '*.md' | wc -l | tr -d ' ')" == 6 ]]
[[ "$(find "${fixture}/demo/.dev/_task/tasks" -maxdepth 1 -name '*.md' -printf '%f\n' | cut -d- -f1 | sort -u | wc -l | tr -d ' ')" == 6 ]]
bash "${wf}" board --check >/dev/null

# Repository lists accept independent nested Git repositories in the workspace.
mkdir -p "${fixture}/ext/rabbitmq"
git -C "${fixture}/ext/rabbitmq" init -q -b main
git -C "${fixture}/ext/rabbitmq" config user.email test@example.com
git -C "${fixture}/ext/rabbitmq" config user.name test
printf 'fixture\n' >"${fixture}/ext/rabbitmq/README.md"
git -C "${fixture}/ext/rabbitmq" add README.md
git -C "${fixture}/ext/rabbitmq" commit -qm init
nested_output="$(bash "${wf}" lite-new nested-repo ext/rabbitmq codex --timing off --loop-mode continuous)"
[[ "${nested_output}" == *"0009-nested-repo.md"* ]]
grep -Fq 'repos: [ext/rabbitmq]' "${fixture}/demo/.dev/_task/tasks/0009-nested-repo.md"
bash "${wf}" branch 0009 >/dev/null
[[ "$(git -C "${fixture}/ext/rabbitmq" branch --show-current)" == 'lite/nested-repo' ]]

echo "dev-harness-lite S1-S4 workflow tests: PASS"
