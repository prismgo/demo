#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
demo_root="$(cd "${script_dir}/../../../.." && pwd)"
state_root="${demo_root}/.dev/_task"
board_file="${state_root}/board.md"
recent_index="${state_root}/archive-recent"
check_only=false
[[ "${1:-}" == "--check" ]] && check_only=true

mkdir -p "${state_root}/tasks" "${state_root}/archive"
if [[ "${DEV_HARNESS_WF_LOCKED:-}" != 1 ]]; then
    command -v flock >/dev/null || { echo "board: flock is required" >&2; exit 2; }
    exec 9>"${state_root}/.wf.lock"
    flock -x 9
fi

if [[ ! -f "${recent_index}" ]]; then
    recent_tmp="$(mktemp "${state_root}/.recent.XXXXXX")"
    find "${state_root}/archive" -maxdepth 1 -type f -name '*.md' -printf '%f\n' | sort | tail -100 >"${recent_tmp}"
    mv "${recent_tmp}" "${recent_index}"
fi

emit_rows() {
    if [[ $# -eq 0 ]]; then
        printf '| — | — | — | — | — | — | — | — |\n'
        return
    fi
    awk '
        function emit() {
            if (!seen) return
            printf "| %s | %s | %s | %s | %s | %s/%s | %s | %s |\n", id, slug, repos, workspace, stage, done+0, total+0, status, updated
        }
        function value(line) { sub(/^[^:]+:[[:space:]]*/, "", line); return line }
        FNR == 1 {
            emit(); seen=1; markers=0; id=slug=repos=workspace=stage=status=updated=""; done=total=inside=0
        }
        /^---$/ { markers++; next }
        markers == 1 && /^id:/ { id=value($0); next }
        markers == 1 && /^slug:/ { slug=value($0); next }
        markers == 1 && /^repos:/ { repos=value($0); gsub(/^\[|\]$/, "", repos); next }
        markers == 1 && /^workspace:/ { workspace=value($0); next }
        markers == 1 && /^stage:/ { stage=value($0); next }
        markers == 1 && /^status:/ { status=value($0); next }
        markers == 1 && /^updated:/ { updated=value($0); next }
        /^## Feature 清单$/ { inside=1; next }
        inside && /^## / { inside=0 }
        inside && /^- \[[ x]\] F[^[:space:]]+[[:space:]]/ { total++ }
        inside && /^- \[x\] F[^[:space:]]+[[:space:]]/ { done++ }
        END { emit() }
    ' "$@"
}

mapfile -t active_files < <(find "${state_root}/tasks" -maxdepth 1 -type f -name '*.md' -print | sort)
archive_files=()
while IFS= read -r basename; do
    [[ -n "${basename}" && -f "${state_root}/archive/${basename}" ]] && archive_files+=("${state_root}/archive/${basename}")
done <"${recent_index}"

render() {
    printf '# PrismGo Lite 工作流看板\n\n'
    printf '> 由 `dev-harness-lite/scripts/board.sh` 从任务卡生成，请勿手改；归档区仅保留最近 100 项。\n\n'
    printf '## 进行中\n\n'
    printf '| id | slug | repos | workspace | stage | features | status | updated |\n'
    printf '|---|---|---|---|---|---|---|---|\n'
    emit_rows "${active_files[@]}"
    printf '\n## 已归档（最近 100）\n\n'
    printf '| id | slug | repos | workspace | stage | features | status | updated |\n'
    printf '|---|---|---|---|---|---|---|---|\n'
    emit_rows "${archive_files[@]}"
}

tmp_file="$(mktemp "${state_root}/.board.XXXXXX")"
trap 'rm -f "${tmp_file}"' EXIT
render >"${tmp_file}"
if ${check_only}; then
    [[ -f "${board_file}" ]] && diff -q "${board_file}" "${tmp_file}" >/dev/null && { echo "board is current"; exit 0; }
    diff -u "${board_file}" "${tmp_file}" 2>/dev/null || true
    exit 1
fi
mv "${tmp_file}" "${board_file}"
trap - EXIT
echo "updated ${board_file}"
