#!/usr/bin/env bash
set -euo pipefail

source_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fixture_root="$(mktemp -d)"
trap 'rm -rf "${fixture_root}"' EXIT

mkdir -p "${fixture_root}/workspace/demo/.dev/scripts"
cp "${source_root}/dev" "${fixture_root}/workspace/demo/dev"
cp "${source_root}/.dev/scripts/test-environment.sh" \
    "${fixture_root}/workspace/demo/.dev/scripts/test-environment.sh"
cp "${source_root}/AGENTS.md" "${fixture_root}/workspace/demo/AGENTS.md"
cp -R "${source_root}/.agents" "${fixture_root}/workspace/demo/.agents"
printf 'keep parent\n' >"${fixture_root}/workspace/.agents"

if output="$("${fixture_root}/workspace/demo/dev" init 2>&1)"; then
    echo "init unexpectedly accepted a conflicting skill parent" >&2
    exit 1
fi

[[ "${output}" == *"Refusing to replace"* ]]
[[ "$(cat "${fixture_root}/workspace/.agents")" == "keep parent" ]]
[[ ! -e "${fixture_root}/workspace/AGENTS.md" ]]
[[ ! -e "${fixture_root}/workspace/CLAUDE.md" ]]

echo "init rejects conflicting skill parent before installing links: PASS"
