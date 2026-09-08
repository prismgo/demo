#!/usr/bin/env bash
set -euo pipefail

source_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fixture_root="$(mktemp -d)"
trap 'rm -rf "${fixture_root}"' EXIT

mkdir -p "${fixture_root}/workspace/dev/.dev/scripts"
mkdir -p "${fixture_root}/workspace/framework" "${fixture_root}/workspace/docs"
cp "${source_root}/dev" "${fixture_root}/workspace/dev/dev"
cp "${source_root}/.dev/scripts/test-environment.sh" \
    "${fixture_root}/workspace/dev/.dev/scripts/test-environment.sh"
cp "${source_root}/AGENTS.md" "${fixture_root}/workspace/dev/AGENTS.md"

"${fixture_root}/workspace/dev/dev" init
"${fixture_root}/workspace/dev/dev" init

[[ "$(readlink "${fixture_root}/workspace/AGENTS.md")" == "dev/AGENTS.md" ]]
[[ "$(readlink "${fixture_root}/workspace/CLAUDE.md")" == "dev/AGENTS.md" ]]

echo "init is idempotent: PASS"
