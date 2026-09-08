#!/usr/bin/env bash
set -euo pipefail

source_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fixture_root="$(mktemp -d)"
cleanup() {
    chmod -R u+w "${fixture_root}" 2>/dev/null || true
    rm -rf "${fixture_root}"
}
trap cleanup EXIT

mkdir -p "${fixture_root}/workspace/dev/.dev/scripts" "${fixture_root}/bin"
mkdir -p "${fixture_root}/workspace/framework" "${fixture_root}/workspace/docs"
cp "${source_root}/dev" "${fixture_root}/workspace/dev/dev"
cp "${source_root}/.dev/scripts/test-environment.sh" \
    "${fixture_root}/workspace/dev/.dev/scripts/test-environment.sh"
cp "${source_root}/AGENTS.md" "${fixture_root}/workspace/dev/AGENTS.md"
cp -R "${source_root}/.agents" "${fixture_root}/workspace/dev/.agents"
cp "${source_root}/.env.example" "${fixture_root}/workspace/dev/.env.example"

cat >"${fixture_root}/bin/go" <<'EOF'
#!/usr/bin/env bash
if [[ "${1:-} ${2:-}" == "work init" ]]; then
    : >go.work
fi
EOF
chmod +x "${fixture_root}/bin/go"

PATH="${fixture_root}/bin:${PATH}" "${fixture_root}/workspace/dev/dev" init &
first_pid=$!
PATH="${fixture_root}/bin:${PATH}" "${fixture_root}/workspace/dev/dev" init &
second_pid=$!
wait "${first_pid}"
wait "${second_pid}"
chmod a-w "${fixture_root}/workspace/.agents" "${fixture_root}/workspace/.agents/skills" \
    "${fixture_root}/workspace/.claude" "${fixture_root}/workspace/.claude/skills"
PATH="${fixture_root}/bin:${PATH}" "${fixture_root}/workspace/dev/dev" init
chmod u+w "${fixture_root}/workspace/.agents" "${fixture_root}/workspace/.agents/skills" \
    "${fixture_root}/workspace/.claude" "${fixture_root}/workspace/.claude/skills"

[[ "$(readlink "${fixture_root}/workspace/AGENTS.md")" == "dev/AGENTS.md" ]]
[[ "$(readlink "${fixture_root}/workspace/CLAUDE.md")" == "dev/AGENTS.md" ]]
[[ "$(readlink "${fixture_root}/workspace/.agents/skills/dev-harness-lite")" == "../../dev/.agents/skills/dev-harness-lite" ]]
[[ "$(readlink "${fixture_root}/workspace/.claude/skills/dev-harness-lite")" == "../../dev/.agents/skills/dev-harness-lite" ]]
[[ "$(readlink "${fixture_root}/workspace/.agents/skills/code-review")" == "../../dev/.agents/skills/code-review" ]]
[[ "$(readlink "${fixture_root}/workspace/.claude/skills/code-review")" == "../../dev/.agents/skills/code-review" ]]

echo "init is idempotent: PASS"
