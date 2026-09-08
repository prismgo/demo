#!/usr/bin/env bash
set -euo pipefail

repo_input="${1:-}"
base="${2:-}"
output="${3:-}"

if [[ -z "${repo_input}" || -z "${base}" || -z "${output}" ]]; then
    echo "usage: wipDiff.sh <repo-path> <fixed-point> </tmp/output.patch>" >&2
    exit 2
fi
if [[ "${output}" != /tmp/* ]]; then
    echo "wipDiff: output must be under /tmp" >&2
    exit 2
fi

command -v flock >/dev/null || { echo "wipDiff: flock is required" >&2; exit 2; }
repo_path="$(cd "${repo_input}" && pwd)"
git -C "${repo_path}" rev-parse --verify "${base}^{commit}" >/dev/null
exec 8>"${output}.lock"
flock -x 8

first="$(mktemp /tmp/prismgo-wip-first.XXXXXX)"
second="$(mktemp /tmp/prismgo-wip-second.XXXXXX)"
index_file="$(mktemp /tmp/prismgo-wip-index.XXXXXX)"
trap 'rm -f "${first}" "${second}" "${index_file}"' EXIT

generate_patch() {
    local destination="$1"
    rm -f "${index_file}"
    GIT_INDEX_FILE="${index_file}" git -C "${repo_path}" read-tree "${base}"
    GIT_INDEX_FILE="${index_file}" git -C "${repo_path}" add -A -- .
    GIT_INDEX_FILE="${index_file}" git -C "${repo_path}" diff --cached --binary "${base}" >"${destination}"
}

stable=false
for _ in 1 2 3; do
    generate_patch "${first}"
    generate_patch "${second}"
    if cmp -s "${first}" "${second}"; then stable=true; break; fi
done
${stable} || { echo "wipDiff: worktree changed while capturing patch" >&2; exit 3; }
[[ -s "${second}" ]] || { echo "wipDiff: empty change set against ${base}" >&2; exit 3; }
mv "${second}" "${output}"

changed="$(grep -c '^diff --git ' "${output}" || true)"
printf 'patch=%s\nhash=%s\nchanged_files=%s\n' \
    "${output}" "$(git hash-object "${output}")" "${changed}"
