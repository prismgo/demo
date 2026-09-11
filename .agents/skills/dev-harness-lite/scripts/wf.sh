#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
skill_root="$(cd "${script_dir}/.." && pwd)"
demo_root="$(cd "${skill_root}/../../.." && pwd)"
workspace_root="$(cd "${demo_root}/.." && pwd)"
state_root="${demo_root}/.dev/_task"
tasks_dir="${state_root}/tasks"
archive_dir="${state_root}/archive"
template="${skill_root}/templates/card.template.md"
board_script="${script_dir}/board.sh"
today="$(date +%F)"

die() { echo "wf: $*" >&2; exit 2; }

usage() {
    cat <<'EOF'
PrismGo dev-harness-lite state helper

  wf.sh lite-new <slug> <repo[,repo...]> [owner] [--workspace normal|worktree]
                 [--timing on|off] [--loop-mode context|continuous] [--loop-limit N]
  wf.sh status [<id>] [--cleared]
  wf.sh sect <id> <section-prefix>
  wf.sh fill <id> <section>                 # body from stdin
  wf.sh branch <id>
  wf.sh lite-add-repos <id> <repo[,repo...]> [--allow-dirty]
  wf.sh lite-merge <id>
  wf.sh lite-progress <id> <S1|S2|S3|S4> <checkpoint>
  wf.sh lite-start <id> <Ftag>
  wf.sh lite-fail <id> <issue-key> <checkpoint>
  wf.sh lite-check <id> <Ftag> [checkpoint]
  wf.sh budget <id> <used-tokens>
  wf.sh pause <id> [reason]
  wf.sh report <id> <budget|timeout|max-attempts|done>
  wf.sh timing <id> [--full]
  wf.sh lite-archive <id>
  wf.sh wtclean <id>
  wf.sh board [--check]

State lives in demo/.dev/_task and is gitignored.
EOF
}

ensure_state() { mkdir -p "${tasks_dir}" "${archive_dir}"; }

field() {
    local file="$1" key="$2"
    awk -v key="${key}" '
        /^---$/ { markers++; next }
        markers == 1 && index($0, key ":") == 1 {
            sub(/^[^:]+:[[:space:]]*/, ""); gsub(/^\[|\]$/, ""); print; exit
        }
    ' "${file}"
}

find_card() {
    local id="$1" file matches=()
    for file in "${tasks_dir}/${id}-"*.md "${archive_dir}/${id}-"*.md; do
        [[ -e "${file}" ]] && matches+=("${file}")
    done
    [[ ${#matches[@]} -eq 1 ]] || return 1
    printf '%s\n' "${matches[0]}"
}

active_card() {
    local id="$1" file
    file="$(find_card "${id}")" || die "card not found: ${id}"
    [[ "${file}" == "${tasks_dir}/"* ]] || die "archived card is read-only: ${id}"
    printf '%s\n' "${file}"
}

rewrite_field() {
    local file="$1" key="$2" value="$3" tmp
    tmp="$(mktemp "${state_root}/.card.XXXXXX")"
    awk -v key="${key}" -v value="${value}" '
        /^---$/ { markers++; print; next }
        markers == 1 && index($0, key ":") == 1 { print key ": " value; changed=1; next }
        { print }
        END { if (!changed) exit 3 }
    ' "${file}" >"${tmp}" || { rm -f "${tmp}"; die "card has no field: ${key}"; }
    mv "${tmp}" "${file}"
}

replace_section() {
    local file="$1" section="$2" content="$3" tmp
    grep -Fxq "## ${section}" "${file}" || die "unknown section: ${section}"
    tmp="$(mktemp "${state_root}/.card.XXXXXX")"
    awk -v heading="## ${section}" -v content="${content}" '
        $0 == heading {
            print; print ""
            while ((getline line < content) > 0) print line
            close(content); replacing=1; next
        }
        replacing && /^## / { replacing=0 }
        replacing { next }
        { print }
    ' "${file}" >"${tmp}"
    mv "${tmp}" "${file}"
}

append_section() {
    local file="$1" section="$2" entry="$3" tmp
    [[ -n "${entry}" && "${entry}" != *$'\n'* ]] || die "entry must be one non-empty line"
    tmp="$(mktemp "${state_root}/.card.XXXXXX")"
    awk -v heading="## ${section}" -v entry="${entry}" '
        $0 == heading { inside=1; found=1; print; next }
        inside && /^## / { print entry; print ""; inside=0 }
        { print }
        END { if (!found) exit 3; if (inside) print entry }
    ' "${file}" >"${tmp}" || { rm -f "${tmp}"; die "unknown section: ${section}"; }
    mv "${tmp}" "${file}"
}

normalize_repos() {
    local input="$1" repo normalized="" seen=","
    IFS=',' read -r -a values <<<"${input}"
    [[ ${#values[@]} -gt 0 ]] || die "at least one repo is required"
    for repo in "${values[@]}"; do
        repo="${repo//[[:space:]]/}"
        [[ -n "${repo}" ]] || die "repository path must not be empty"
        [[ "${repo}" != /* && "${repo}" != "." && "${repo}" != ".." &&
            "${repo}" != ../* && "${repo}" != */../* && "${repo}" != */.. &&
            "${repo}" != ./* && "${repo}" != */./* && "${repo}" != */. ]] || \
            die "repository path must stay within the workspace: ${repo}"
        [[ "${seen}" == *",${repo},"* ]] && continue
        normalized="${normalized:+${normalized}, }${repo}"; seen+="${repo},"
    done
    printf '%s\n' "${normalized}"
}

repos_each() {
    local file="$1" values repo
    values="$(field "${file}" repos)"; values="${values// /}"
    IFS=',' read -r -a repos <<<"${values}"
    for repo in "${repos[@]}"; do printf '%s\n' "${repo}"; done
}

next_id() {
    local counter="${state_root}/next-id" file base raw maximum=0 number next tmp
    if [[ -f "${counter}" ]]; then
        next="$(<"${counter}")"
        [[ "${next}" =~ ^[1-9][0-9]*$ ]] || die "invalid next-id counter: ${next}"
    else
        # One-time migration for workspaces created before the persistent counter.
        for file in "${tasks_dir}"/*.md "${archive_dir}"/*.md; do
            [[ -e "${file}" ]] || continue
            base="${file##*/}"; raw="${base%%-*}"
            [[ "${raw}" =~ ^[0-9]{4}$ ]] || continue
            number=$((10#${raw})); (( number > maximum )) && maximum=${number}
        done
        next=$((maximum + 1))
    fi
    (( next <= 9999 )) || die "lite card id space exhausted"
    tmp="$(mktemp "${state_root}/.next-id.XXXXXX")"
    printf '%s\n' $((next + 1)) >"${tmp}"
    mv "${tmp}" "${counter}"
    printf '%04d\n' "${next}"
}

assets_dir() {
    local file="$1" id directory
    id="$(field "${file}" id)"; directory="$(dirname "${file}")"
    printf '%s/%s.assets\n' "${directory}" "${id}"
}

timing_enabled() { [[ "$(field "$1" timing)" == "on" ]]; }

ensure_not_paused() {
    local file="$1"
    [[ "$(field "${file}" pause_state)" != waiting ]] || die "card is paused; resume with status $(field "${file}" id) --cleared"
}

timing_event() {
    (
        local file="$1" event="$2" detail="${3:-}" assets log now last
        timing_enabled "${file}" || exit 0
        assets="$(assets_dir "${file}")"; log="${assets}/timing.jsonl"; mkdir -p "${assets}"
        now="$(date +%s)"
        last="$(tail -n 1 "${log}" 2>/dev/null | sed -n 's/.*"event":"\([^"]*\)".*/\1/p' || true)"
        if [[ "${last}" == "pause" && "${event}" != "pause" ]]; then
            printf '{"ts":%s,"event":"resume","stage":"%s","detail":"automatic"}\n' \
                "${now}" "$(field "${file}" stage)" >>"${log}"
        fi
        detail="${detail//\\/\\\\}"; detail="${detail//\"/\\\"}"
        printf '{"ts":%s,"event":"%s","stage":"%s","detail":"%s"}\n' \
            "${now}" "${event}" "$(field "${file}" stage)" "${detail}" >>"${log}"
    ) >/dev/null 2>&1 || true
}

refresh() {
    local file="$1"
    rewrite_field "${file}" updated "${today}"
    DEV_HARNESS_WF_LOCKED=1 bash "${board_script}" >/dev/null
}

list_has() {
    local values="${1// /}" wanted="$2" value
    IFS=',' read -r -a items <<<"${values}"
    for value in "${items[@]}"; do [[ "${value}" == "${wanted}" ]] && return 0; done
    return 1
}

mark_repo_merged() {
    local file="$1" repo="$2" merged
    merged="$(field "${file}" merged_repos)"
    list_has "${merged}" "${repo}" && return
    merged="${merged:+${merged}, }${repo}"
    rewrite_field "${file}" merged_repos "[${merged}]"
    append_section "${file}" "进度日志" "- ${today} S4 merge：${repo} 已合并到 main"
    refresh "${file}"; timing_event "${file}" lite-merge "${repo} merged"
}

mark_repo_branched() {
    local file="$1" repo="$2" branched
    branched="$(field "${file}" branched_repos)"
    list_has "${branched}" "${repo}" && return
    branched="${branched:+${branched}, }${repo}"
    rewrite_field "${file}" branched_repos "[${branched}]"
    append_section "${file}" "进度日志" "- ${today} S1 branch：${repo} 已准备"
    refresh "${file}"; timing_event "${file}" branch "${repo}"
}

record_recent_archive() {
    local basename="$1" index="${state_root}/archive-recent" tmp
    tmp="$(mktemp "${state_root}/.recent.XXXXXX")"
    { [[ ! -f "${index}" ]] || cat "${index}"; printf '%s\n' "${basename}"; } | \
        awk 'NF && !seen[$0]++' | tail -100 >"${tmp}"
    mv "${tmp}" "${index}"
}

increment_attempt() {
    local values="$1" wanted="$2" entry key count found=false result="" next
    [[ "${values}" == - ]] && values=""
    IFS=',' read -r -a entries <<<"${values}"
    for entry in "${entries[@]}"; do
        [[ -n "${entry}" ]] || continue
        key="${entry%%=*}"; count="${entry#*=}"
        [[ "${count}" =~ ^[0-9]+$ ]] || die "invalid attempts entry: ${entry}"
        if [[ "${key}" == "${wanted}" ]]; then count=$((count + 1)); next="${count}"; found=true; fi
        result="${result:+${result},}${key}=${count}"
    done
    if ! ${found}; then next=1; result="${result:+${result},}${wanted}=1"; fi
    printf '%s\n%s\n' "${result}" "${next}"
}

cmd_lite_new() {
    local slug="${1:-}" repo_input="${2:-}" owner="${3:-agent}" workspace="normal" timing="on"
    local loop_mode="context" loop_limit="500" id card repos card_body
    [[ "${slug}" =~ ^[a-z0-9]+(-[a-z0-9]+)*$ ]] || die "slug must be kebab-case"
    [[ -n "${repo_input}" ]] || die "usage: wf.sh lite-new <slug> <repos> [owner] [options]"
    shift 2
    if [[ $# -gt 0 && "$1" != --* ]]; then owner="$1"; shift; fi
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --workspace) workspace="${2:-}"; shift 2 ;;
            --timing) timing="${2:-}"; shift 2 ;;
            --loop-mode) loop_mode="${2:-}"; shift 2 ;;
            --loop-limit) loop_limit="${2:-}"; shift 2 ;;
            *) die "unknown option: $1" ;;
        esac
    done
    case "${workspace}" in normal|worktree) ;; *) die "workspace must be normal or worktree" ;; esac
    case "${timing}" in on|off) ;; *) die "timing must be on or off" ;; esac
    case "${loop_mode}" in context|continuous) ;; *) die "loop-mode must be context or continuous" ;; esac
    [[ "${loop_limit}" =~ ^[1-9][0-9]*$ ]] || die "loop-limit must be a positive integer"
    repos="$(normalize_repos "${repo_input}")"; id="$(next_id)"; card="${tasks_dir}/${id}-${slug}.md"
    card_body="$(<"${template}")"
    card_body="${card_body//\{\{ID\}\}/${id}}"
    card_body="${card_body//\{\{SLUG\}\}/${slug}}"
    card_body="${card_body//\{\{OWNER\}\}/${owner}}"
    card_body="${card_body//\{\{REPOS\}\}/${repos}}"
    card_body="${card_body//\{\{WORKSPACE\}\}/${workspace}}"
    card_body="${card_body//\{\{TIMING\}\}/${timing}}"
    card_body="${card_body//\{\{LOOP_MODE\}\}/${loop_mode}}"
    card_body="${card_body//\{\{LOOP_LIMIT\}\}/${loop_limit}}"
    card_body="${card_body//\{\{DATE\}\}/${today}}"
    printf '%s\n' "${card_body}" >"${card}"
    timing_event "${card}" lite-new "card created"
    bash "${board_script}" >/dev/null
    echo "created ${card}"
}

cmd_lite_add_repos() {
    local id="${1:-}" repo_input="${2:-}" option="${3:-}" file branch existing requested combined repo repo_path current
    local -a added=()
    [[ -n "${id}" && -n "${repo_input}" ]] || die "usage: wf.sh lite-add-repos <id> <repo[,repo...]> [--allow-dirty]"
    [[ -z "${option}" || "${option}" == --allow-dirty ]] || die "unknown lite-add-repos option: ${option}"
    file="$(active_card "${id}")"; ensure_not_paused "${file}"
    case "$(field "${file}" stage)" in S1|S2) ;; *) die "lite-add-repos requires stage S1 or S2" ;; esac
    [[ "$(field "${file}" workspace)" == normal ]] || die "lite-add-repos currently requires normal workspace"
    branch="$(field "${file}" branch)"; existing="$(field "${file}" repos)"; requested="$(normalize_repos "${repo_input}")"

    requested="${requested// /}"
    IFS=',' read -r -a repos <<<"${requested}"
    for repo in "${repos[@]}"; do
        list_has "${existing}" "${repo}" && continue
        repo_path="${workspace_root}/${repo}"
        [[ -d "${repo_path}/.git" || -f "${repo_path}/.git" ]] || die "not a git repo: ${repo_path}"
        [[ "${option}" == --allow-dirty || -z "$(git -C "${repo_path}" status --short)" ]] || \
            die "dirty repository cannot be added without --allow-dirty: ${repo}"
        current="$(git -C "${repo_path}" branch --show-current)"
        [[ "${current}" == "${branch}" ]] || die "repository must already be on ${branch}: ${repo} (${current})"
        added+=("${repo}")
    done

    (( ${#added[@]} > 0 )) || { echo "repositories already tracked: ${requested}"; return; }
    combined="$(normalize_repos "${existing},${requested}")"
    rewrite_field "${file}" repos "[${combined}]"
    for repo in "${added[@]}"; do mark_repo_branched "${file}" "${repo}"; done
    echo "added repositories to ${id}: ${added[*]}"
}

feature_counts() {
    awk '
        /^## Feature 清单$/ { inside=1; next }
        inside && /^## / { exit }
        inside && /^- \[[ x]\] F[^[:space:]]+[[:space:]]/ { total++ }
        inside && /^- \[x\] F[^[:space:]]+[[:space:]]/ { done++ }
        END { print total + 0, done + 0 }
    ' "$1"
}

print_summary() {
    local file="$1" total done
    read -r total done < <(feature_counts "${file}")
    printf '%s %-28s repos=%-22s workspace=%-8s stage=%s features=%s/%s active=%s last_failure=%s failures=%s budget=%s/%s branched=%s merged=%s pause=%s status=%s updated=%s\n' \
        "$(field "${file}" id)" "$(field "${file}" slug)" "$(field "${file}" repos)" \
        "$(field "${file}" workspace)" "$(field "${file}" stage)" "${done}" "${total}" \
        "$(field "${file}" active_feature)" "$(field "${file}" attempt_key)" "$(field "${file}" attempts)" \
        "$(field "${file}" budget_used)" "$(field "${file}" loop_limit)" "$(field "${file}" branched_repos)" \
        "$(field "${file}" merged_repos)" "$(field "${file}" pause_state)" \
        "$(field "${file}" status)" "$(field "${file}" updated)"
}

cmd_status() {
    local id="" cleared=false file found=false arg
    for arg in "$@"; do
        case "${arg}" in --cleared) cleared=true ;; *) [[ -z "${id}" ]] || die "too many status arguments"; id="${arg}" ;; esac
    done
    if [[ -n "${id}" ]]; then
        file="$(find_card "${id}")" || die "card not found: ${id}"
        if ${cleared}; then
            [[ "${file}" == "${tasks_dir}/"* ]] || die "archived card cannot be cleared: ${id}"
            [[ "$(field "${file}" pause_state)" == waiting ]] || die "card is not waiting for compact/clear: ${id}"
            rewrite_field "${file}" pause_state cleared
            refresh "${file}"; timing_event "${file}" cleared "context reset"
        fi
        print_summary "${file}"; echo
        awk '
            /^## Feature 清单$/ || /^## 决策$/ || /^## 进度日志$/ || /^## Handoff$/ { show=1 }
            show && /^## / && !(/^## Feature 清单$/ || /^## 决策$/ || /^## 进度日志$/ || /^## Handoff$/) { show=0 }
            show { print }
        ' "${file}"
        return
    fi
    for file in "${tasks_dir}"/*.md; do [[ -e "${file}" ]] || continue; print_summary "${file}"; found=true; done
    ${found} || echo "(无)"
}

cmd_sect() {
    local id="${1:-}" prefix="${2:-}" file matches
    [[ -n "${id}" && -n "${prefix}" ]] || die "usage: wf.sh sect <id> <section-prefix>"
    file="$(find_card "${id}")" || die "card not found: ${id}"
    matches="$(grep -c "^## ${prefix}" "${file}" || true)"; [[ "${matches}" == 1 ]] || die "section prefix must match once: ${prefix}"
    awk -v prefix="## ${prefix}" 'index($0,prefix)==1 {show=1} show && /^## / && index($0,prefix)!=1 {exit} show {print}' "${file}"
}

cmd_fill() {
    local id="${1:-}" section="${2:-}" file content
    [[ -n "${id}" && -n "${section}" ]] || die "usage: wf.sh fill <id> <section>"
    case "${section}" in 需求描述|外部文档索引|"Agent 设计摘要"|"验收标准 DoD"|"Feature 清单"|决策|Handoff) ;;
        *) die "fill cannot replace managed or unknown section: ${section}" ;; esac
    file="$(active_card "${id}")"; ensure_not_paused "${file}"
    content="$(mktemp)"; trap 'rm -f "${content}"' RETURN
    cat >"${content}"; [[ -s "${content}" ]] || die "section body must not be empty"
    replace_section "${file}" "${section}" "${content}"; rm -f "${content}"; trap - RETURN
    refresh "${file}"; timing_event "${file}" fill "${section}"; echo "updated ${section} in ${file}"
}

cmd_lite_progress() {
    local id="${1:-}" stage="${2:-}" note="${3:-}" file current
    [[ -n "${id}" && -n "${stage}" && -n "${note}" ]] || die "usage: wf.sh lite-progress <id> <S1|S2|S3|S4> <note>"
    case "${stage}" in S1|S2|S3|S4) ;; *) die "invalid lite stage: ${stage}" ;; esac
    file="$(active_card "${id}")"; current="$(field "${file}" stage)"
    ensure_not_paused "${file}"
    if [[ "${current}" == S1 && "${stage}" == S2 ]]; then
        [[ "$(field "${file}" pause_state)" == cleared ]] || die "S1 must pause and clear before S2"
    fi
    case "${current}:${stage}" in
        S1:S1|S1:S2|S2:S2|S2:S3|S3:S2|S3:S3|S3:S4|S4:S2|S4:S3|S4:S4) ;;
        *) die "invalid lite stage transition: ${current} -> ${stage}" ;;
    esac
    if [[ "${current}" == S2 && "${stage}" == S3 ]]; then
        read -r total done < <(feature_counts "${file}")
        [[ "${total}" -gt 0 && "${done}" == "${total}" ]] || die "cannot enter S3: incomplete features"
        [[ "$(field "${file}" active_feature)" == - ]] || die "cannot enter S3: active feature is not checked"
    fi
    rewrite_field "${file}" stage "${stage}"
    append_section "${file}" "进度日志" "- ${today} ${stage}：${note}"
    refresh "${file}"; timing_event "${file}" lite-progress "${note}"; echo "stage=${stage} ${file}"
}

cmd_lite_check() {
    local id="${1:-}" tag="${2:-}" note="${3:-}" file tmp matches
    [[ -n "${id}" && "${tag}" =~ ^F[0-9]+([.][A-Za-z0-9_-]+)*$ ]] || die "usage: wf.sh lite-check <id> <Ftag> [note]"
    file="$(active_card "${id}")"; ensure_not_paused "${file}"
    [[ "$(field "${file}" stage)" == S2 ]] || die "lite-check requires stage S2"
    [[ "$(field "${file}" active_feature)" == "${tag}" ]] || die "lite-check requires lite-start ${id} ${tag}"
    matches="$(awk -v tag="${tag}" '$0 ~ /^- \[ \] / { rest=substr($0,7); split(rest,a," "); if (a[1] == tag) n++ } END { print n+0 }' "${file}")"
    [[ "${matches}" == 1 ]] || die "unchecked feature tag must occur exactly once: ${tag}"
    tmp="$(mktemp "${state_root}/.card.XXXXXX")"
    awk -v tag="${tag}" '$0 ~ /^- \[ \] / { rest=substr($0,7); split(rest,a," "); if (a[1] == tag) sub(/^- \[ \]/,"- [x]") } {print}' "${file}" >"${tmp}"
    mv "${tmp}" "${file}"
    rewrite_field "${file}" active_feature -
    rewrite_field "${file}" attempt_key -
    append_section "${file}" "进度日志" "- ${today} ${tag} done：${note:-验证通过}"
    refresh "${file}"; timing_event "${file}" lite-check "${tag}: ${note:-PASS}"; echo "checked ${tag} in ${file}"
}

cmd_lite_start() {
    local id="${1:-}" tag="${2:-}" file unchecked checked active tmp
    [[ -n "${id}" && "${tag}" =~ ^F[0-9]+([.][A-Za-z0-9_-]+)*$ ]] || die "usage: wf.sh lite-start <id> <Ftag>"
    file="$(active_card "${id}")"; ensure_not_paused "${file}"
    [[ "$(field "${file}" stage)" == S2 ]] || die "lite-start requires stage S2"
    read -r unchecked checked < <(awk -v tag="${tag}" '
        /^- \[[ x]\] / { rest=substr($0,7); split(rest,a," "); if (a[1] == tag) { if ($0 ~ /^- \[x\]/) checked++; else unchecked++ } }
        END { print unchecked+0, checked+0 }
    ' "${file}")
    [[ $((unchecked + checked)) == 1 ]] || die "feature tag must occur exactly once: ${tag}"
    if [[ "${checked}" == 1 ]]; then
        tmp="$(mktemp "${state_root}/.card.XXXXXX")"
        awk -v tag="${tag}" '$0 ~ /^- \[x\] / { rest=substr($0,7); split(rest,a," "); if (a[1] == tag) sub(/^- \[x\]/,"- [ ]") } {print}' "${file}" >"${tmp}"
        mv "${tmp}" "${file}"
        append_section "${file}" "进度日志" "- ${today} ${tag} reopen：review 修复重新进入 TDD"
    fi
    active="$(field "${file}" active_feature)"
    if [[ "${active}" != "${tag}" ]]; then
        rewrite_field "${file}" active_feature "${tag}"
        rewrite_field "${file}" attempt_key -
    fi
    append_section "${file}" "进度日志" "- ${today} ${tag} start：进入 TDD 竖切片"
    refresh "${file}"; timing_event "${file}" lite-start "${tag}"; echo "active ${tag} in ${file}"
}

cmd_lite_fail() {
    local id="${1:-}" key="${2:-}" note="${3:-}" file stage attempts
    local -a updated
    [[ -n "${id}" && "${key}" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*$ && -n "${note}" ]] || \
        die "usage: wf.sh lite-fail <id> <issue-key> <checkpoint>"
    file="$(active_card "${id}")"; ensure_not_paused "${file}"; stage="$(field "${file}" stage)"
    [[ "${stage}" == S2 || "${stage}" == S3 ]] || die "lite-fail requires stage S2 or S3"
    mapfile -t updated < <(increment_attempt "$(field "${file}" attempts)" "${key}")
    attempts="${updated[1]}"
    rewrite_field "${file}" attempt_key "${key}"; rewrite_field "${file}" attempts "${updated[0]}"
    append_section "${file}" "进度日志" "- ${today} ${stage} fail ${key} (${attempts}/3)：${note}"
    refresh "${file}"; timing_event "${file}" lite-fail "${key} ${attempts}/3: ${note}"
    if (( attempts >= 3 )); then
        echo "max attempts reached for ${key}: stop and hand off" >&2
        return 3
    fi
    echo "attempt ${attempts}/3 for ${key} recorded"
}

cmd_budget() {
    local id="${1:-}" used="${2:-}" file mode limit
    [[ -n "${id}" && "${used}" =~ ^[0-9]+$ ]] || die "usage: wf.sh budget <id> <used-tokens>"
    file="$(active_card "${id}")"; ensure_not_paused "${file}"
    mode="$(field "${file}" loop_mode)"; limit="$(field "${file}" loop_limit)"
    [[ "${limit}" =~ ^[1-9][0-9]*$ ]] || die "invalid loop_limit field: ${limit}"
    rewrite_field "${file}" budget_used "${used}"; refresh "${file}"; timing_event "${file}" budget "${used}/${limit}"
    if [[ "${mode}" == context && "${used}" -ge "${limit}" ]]; then
        echo "budget=due used=${used} limit=${limit}"
        return 3
    fi
    echo "budget=continue used=${used} limit=${limit} mode=${mode}"
}

cmd_branch() {
    local id="${1:-}" file branch workspace repo repo_path wt_path current
    [[ -n "${id}" ]] || die "usage: wf.sh branch <id>"
    file="$(active_card "${id}")"; ensure_not_paused "${file}"
    branch="$(field "${file}" branch)"; workspace="$(field "${file}" workspace)"
    # Preflight every repository before mutating any checkout.
    while IFS= read -r repo; do
        repo_path="${workspace_root}/${repo}"; [[ -d "${repo_path}/.git" || -f "${repo_path}/.git" ]] || die "not a git repo: ${repo_path}"
        [[ -z "$(git -C "${repo_path}" status --short)" ]] || die "dirty repo, refusing branch switch: ${repo}"
        git -C "${repo_path}" show-ref --verify --quiet refs/heads/main || die "missing main branch: ${repo}"
        current="$(git -C "${repo_path}" branch --show-current)"
        if [[ "${workspace}" == normal ]]; then
            [[ "${current}" == main || "${current}" == "${branch}" ]] || die "repo must be on main or ${branch}: ${repo}"
            if [[ "${current}" != "${branch}" ]] && git -C "${repo_path}" show-ref --verify --quiet "refs/heads/${branch}" && \
                git -C "${repo_path}" worktree list --porcelain | grep -Fqx "branch refs/heads/${branch}"; then
                die "feature branch is checked out in another worktree: ${repo} ${branch}"
            fi
        else
            [[ "${current}" == main ]] || die "worktree mode requires the primary checkout on main: ${repo}"
            wt_path="${workspace_root}/.worktrees/$(field "${file}" id)-$(field "${file}" slug)/${repo}"
            if [[ -e "${wt_path}" ]]; then
                [[ -d "${wt_path}" && "$(git -C "${wt_path}" branch --show-current 2>/dev/null)" == "${branch}" ]] || \
                    die "existing worktree path is not ${branch}: ${wt_path}"
            elif git -C "${repo_path}" show-ref --verify --quiet "refs/heads/${branch}" && \
                git -C "${repo_path}" worktree list --porcelain | grep -Fqx "branch refs/heads/${branch}"; then
                die "feature branch is checked out in another worktree: ${repo} ${branch}"
            fi
        fi
    done < <(repos_each "${file}")

    while IFS= read -r repo; do
        repo_path="${workspace_root}/${repo}"
        [[ -z "$(git -C "${repo_path}" status --short)" ]] || die "repo changed after branch preflight: ${repo}"
        current="$(git -C "${repo_path}" branch --show-current)"
        if [[ "${workspace}" == normal ]]; then
            [[ "${current}" == main || "${current}" == "${branch}" ]] || die "branch changed after preflight: ${repo}"
            if git -C "${repo_path}" show-ref --verify --quiet "refs/heads/${branch}"; then git -C "${repo_path}" switch "${branch}"
            else git -C "${repo_path}" switch -c "${branch}"; fi
        else
            wt_path="${workspace_root}/.worktrees/$(field "${file}" id)-$(field "${file}" slug)/${repo}"
            if [[ -e "${wt_path}" ]]; then mark_repo_branched "${file}" "${repo}"; continue; fi
            mkdir -p "$(dirname "${wt_path}")"
            [[ "${current}" == main ]] || die "primary checkout changed after preflight: ${repo}"
            if git -C "${repo_path}" show-ref --verify --quiet "refs/heads/${branch}"; then git -C "${repo_path}" worktree add "${wt_path}" "${branch}"
            else git -C "${repo_path}" worktree add -b "${branch}" "${wt_path}" main; fi
        fi
        mark_repo_branched "${file}" "${repo}"
    done < <(repos_each "${file}")
    echo "branch ready: ${branch} (${workspace})"
}

cmd_lite_merge() {
    local id="${1:-}" file branch workspace repo repo_path wt_path current
    [[ -n "${id}" ]] || die "usage: wf.sh lite-merge <id>"
    file="$(active_card "${id}")"; ensure_not_paused "${file}"
    [[ "$(field "${file}" stage)" == S4 ]] || die "merge requires stage S4"
    branch="$(field "${file}" branch)"; workspace="$(field "${file}" workspace)"

    while IFS= read -r repo; do
        repo_path="${workspace_root}/${repo}"; [[ -z "$(git -C "${repo_path}" status --short)" ]] || die "dirty primary checkout: ${repo}"
        git -C "${repo_path}" show-ref --verify --quiet refs/heads/main || die "missing main branch: ${repo}"
        git -C "${repo_path}" show-ref --verify --quiet "refs/heads/${branch}" || die "missing feature branch: ${repo} ${branch}"
        current="$(git -C "${repo_path}" branch --show-current)"
        if [[ "${workspace}" == normal ]]; then
            [[ "${current}" == "${branch}" || "${current}" == main ]] || die "unexpected current branch: ${repo} ${current}"
        else
            [[ "${current}" == main ]] || die "worktree mode requires the primary checkout on main: ${repo}"
            wt_path="${workspace_root}/.worktrees/${id}-$(field "${file}" slug)/${repo}"
            [[ ! -e "${wt_path}" || -z "$(git -C "${wt_path}" status --short)" ]] || die "dirty feature worktree: ${wt_path}"
        fi
    done < <(repos_each "${file}")

    while IFS= read -r repo; do
        repo_path="${workspace_root}/${repo}"
        [[ -z "$(git -C "${repo_path}" status --short)" ]] || die "repo changed after merge preflight: ${repo}"
        if git -C "${repo_path}" merge-base --is-ancestor "${branch}" main; then
            mark_repo_merged "${file}" "${repo}"
            continue
        fi
        [[ "$(git -C "${repo_path}" branch --show-current)" == main ]] || git -C "${repo_path}" switch main
        if ! git -C "${repo_path}" merge --no-ff --no-edit "${branch}"; then
            append_section "${file}" "进度日志" "- ${today} S4 merge FAIL：${repo}；先解决或 abort 该仓 merge，再重跑 lite-merge"
            refresh "${file}"; timing_event "${file}" lite-merge-fail "${repo}"
            echo "merge failed in ${repo}; completed repositories are recorded in merged_repos" >&2
            return 2
        fi
        mark_repo_merged "${file}" "${repo}"
    done < <(repos_each "${file}")
    echo "merged ${branch} into main"
}

cmd_pause() {
    local id="${1:-}" reason="${2:-}" file
    [[ -n "${id}" ]] || die "usage: wf.sh pause <id> [reason]"
    file="$(active_card "${id}")"; rewrite_field "${file}" pause_state waiting; refresh "${file}"
    timing_event "${file}" pause "${reason:-waiting for user}"
    echo "paused ${id}: ${reason:-waiting for user}"
}

timing_summary() {
    local file="$1" log
    log="$(assets_dir "${file}")/timing.jsonl"
    [[ -f "${log}" ]] || { echo "不记录"; return; }
    awk '
        { if (match($0,/"ts":[0-9]+/)) { ts=substr($0,RSTART+5,RLENGTH-5)+0 }
          if (match($0,/"event":"[^"]+"/)) ev=substr($0,RSTART+9,RLENGTH-10)
          if (!first) first=ts; last=ts; events++
          if (ev=="pause") pause_at=ts
          if (ev=="resume" && pause_at) { waiting += ts-pause_at; pause_at=0 }
        }
        END { wall=last-first; if (wall<0) wall=0; active=wall-waiting; if(active<0)active=0;
              printf "事件 %d；墙钟 %ds；等待 %ds；有效 %ds",events,wall,waiting,active }
    ' "${log}"
}

cmd_timing() {
    local id="${1:-}" full="${2:-}" file log
    [[ -n "${id}" ]] || die "usage: wf.sh timing <id> [--full]"
    file="$(find_card "${id}")" || die "card not found: ${id}"; echo "$(timing_summary "${file}")"
    log="$(assets_dir "${file}")/timing.jsonl"; [[ "${full}" == --full && -f "${log}" ]] && cat "${log}"
    return 0
}

cmd_report() {
    local id="${1:-}" reason="${2:-}" file content
    [[ -n "${id}" && -n "${reason}" ]] || die "usage: wf.sh report <id> <reason>"
    case "${reason}" in budget|timeout|max-attempts|done) ;; *) die "invalid report reason: ${reason}" ;; esac
    file="$(active_card "${id}")"; ensure_not_paused "${file}"; content="$(mktemp)"
    printf '### 终局报告\n\n- 日期：%s\n- 原因：%s\n- 阶段：%s\n- 下一步：按最近进度日志续做。\n' \
        "${today}" "${reason}" "$(field "${file}" stage)" >"${content}"
    replace_section "${file}" Handoff "${content}"; rm -f "${content}"
    refresh "${file}"; timing_event "${file}" report "${reason}"; echo "reported ${reason} in ${file}"
}

cmd_lite_archive() {
    local id="${1:-}" file total done content source_assets destination_assets repo repo_path branch workspace wt_path
    [[ -n "${id}" ]] || die "usage: wf.sh lite-archive <id>"
    file="$(active_card "${id}")"; ensure_not_paused "${file}"
    [[ "$(field "${file}" stage)" == S4 ]] || die "cannot archive: stage must be S4"
    read -r total done < <(feature_counts "${file}")
    [[ "${total}" -gt 0 ]] || die "cannot archive: no features"
    [[ "${done}" == "${total}" ]] || die "cannot archive: $((total-done)) feature(s) remain"
    branch="$(field "${file}" branch)"; workspace="$(field "${file}" workspace)"
    while IFS= read -r repo; do
        repo_path="${workspace_root}/${repo}"
        [[ -z "$(git -C "${repo_path}" status --short)" ]] || die "cannot archive: dirty repo ${repo}"
        [[ "$(git -C "${repo_path}" branch --show-current)" == main ]] || die "cannot archive: ${repo} is not on main"
        git -C "${repo_path}" merge-base --is-ancestor "${branch}" main || die "cannot archive: ${branch} is not merged in ${repo}"
        list_has "$(field "${file}" merged_repos)" "${repo}" || die "cannot archive: ${repo} is missing from merged_repos"
        if [[ "${workspace}" == worktree ]]; then
            wt_path="${workspace_root}/.worktrees/${id}-$(field "${file}" slug)/${repo}"
            [[ ! -e "${wt_path}" ]] || die "cannot archive: worktree still exists ${wt_path}"
        fi
    done < <(repos_each "${file}")
    timing_event "${file}" lite-archive "done"
    content="$(mktemp)"; timing_summary "${file}" >"${content}"; replace_section "${file}" "耗时统计" "${content}"; rm -f "${content}"
    grep -q '<[^>]*>' "${file}" && die "cannot archive: card still contains placeholders"
    rewrite_field "${file}" stage S4; rewrite_field "${file}" status done; rewrite_field "${file}" updated "${today}"
    source_assets="$(assets_dir "${file}")"; destination_assets="${archive_dir}/${id}.assets"
    mv "${file}" "${archive_dir}/$(basename "${file}")"; [[ ! -d "${source_assets}" ]] || mv "${source_assets}" "${destination_assets}"
    record_recent_archive "$(basename "${file}")"
    DEV_HARNESS_WF_LOCKED=1 bash "${board_script}" >/dev/null; echo "archived ${archive_dir}/$(basename "${file}")"
}

cmd_wtclean() {
    local id="${1:-}" file branch repo repo_path wt_path
    [[ -n "${id}" ]] || die "usage: wf.sh wtclean <id>"
    file="$(find_card "${id}")" || die "card not found: ${id}"; [[ "$(field "${file}" workspace)" == worktree ]] || die "card is not worktree mode"
    branch="$(field "${file}" branch)"
    while IFS= read -r repo; do
        repo_path="${workspace_root}/${repo}"; wt_path="${workspace_root}/.worktrees/${id}-$(field "${file}" slug)/${repo}"
        [[ -e "${wt_path}" ]] || continue
        [[ -z "$(git -C "${wt_path}" status --short)" ]] || die "dirty worktree: ${wt_path}"
        git -C "${repo_path}" merge-base --is-ancestor "${branch}" main || die "branch not merged into main: ${repo} ${branch}"
        git -C "${repo_path}" worktree remove "${wt_path}"
    done < <(repos_each "${file}")
    echo "worktrees cleaned: ${id}"
}

ensure_state
command -v flock >/dev/null || die "flock is required"
exec 9>"${state_root}/.wf.lock"
flock -x 9
export DEV_HARNESS_WF_LOCKED=1
command="${1:-help}"; shift || true
case "${command}" in
    help|-h|--help) usage ;;
    lite-new) cmd_lite_new "$@" ;;
    new) cmd_lite_new "$@" ;;
    status) cmd_status "$@" ;;
    sect) cmd_sect "$@" ;;
    fill) cmd_fill "$@" ;;
    branch) cmd_branch "$@" ;;
    lite-add-repos) cmd_lite_add_repos "$@" ;;
    lite-merge) cmd_lite_merge "$@" ;;
    lite-progress) cmd_lite_progress "$@" ;;
    lite-start) cmd_lite_start "$@" ;;
    lite-fail) cmd_lite_fail "$@" ;;
    lite-check) cmd_lite_check "$@" ;;
    budget) cmd_budget "$@" ;;
    pause) cmd_pause "$@" ;;
    report) cmd_report "$@" ;;
    timing) cmd_timing "$@" ;;
    lite-archive) cmd_lite_archive "$@" ;;
    archive) cmd_lite_archive "$@" ;;
    wtclean) cmd_wtclean "$@" ;;
    board) bash "${board_script}" "$@" ;;
    *) die "unknown command: ${command} (run wf.sh help)" ;;
esac
