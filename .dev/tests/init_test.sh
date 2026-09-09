#!/usr/bin/env bash
set -euo pipefail

source_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fixture_root="$(mktemp -d)"
trap 'rm -rf "${fixture_root}"' EXIT

mkdir -p "${fixture_root}/workspace/demo/.dev/scripts" "${fixture_root}/bin"
mkdir -p "${fixture_root}/workspace/framework" "${fixture_root}/workspace/docs"
cp "${source_root}/dev" "${fixture_root}/workspace/demo/dev"
cp "${source_root}/.dev/scripts/test-environment.sh" \
    "${fixture_root}/workspace/demo/.dev/scripts/test-environment.sh"
cp "${source_root}/AGENTS.md" "${fixture_root}/workspace/demo/AGENTS.md"
cp -R "${source_root}/.agents" "${fixture_root}/workspace/demo/.agents"
cp "${source_root}/.env.example" "${fixture_root}/workspace/demo/.env.example"

cat >"${fixture_root}/bin/go" <<'EOF'
#!/usr/bin/env bash
if [[ "${1:-} ${2:-}" == "work init" ]]; then
    : >go.work
fi
EOF
chmod +x "${fixture_root}/bin/go"

PATH="${fixture_root}/bin:${PATH}" "${fixture_root}/workspace/demo/dev" init

[[ "$(readlink "${fixture_root}/workspace/AGENTS.md")" == "demo/AGENTS.md" ]]
[[ "$(readlink "${fixture_root}/workspace/CLAUDE.md")" == "demo/AGENTS.md" ]]
[[ "$(readlink "${fixture_root}/workspace/.agents/skills/dev-harness-lite")" == "../../demo/.agents/skills/dev-harness-lite" ]]
[[ "$(readlink "${fixture_root}/workspace/.claude/skills/dev-harness-lite")" == "../../demo/.agents/skills/dev-harness-lite" ]]
[[ "$(readlink "${fixture_root}/workspace/.agents/skills/code-review")" == "../../demo/.agents/skills/code-review" ]]
[[ "$(readlink "${fixture_root}/workspace/.claude/skills/code-review")" == "../../demo/.agents/skills/code-review" ]]
[[ -f "${fixture_root}/workspace/.agents/skills/dev-harness-lite/SKILL.md" ]]
[[ -f "${fixture_root}/workspace/.claude/skills/dev-harness-lite/SKILL.md" ]]
[[ -f "${fixture_root}/workspace/.agents/skills/code-review/SKILL.md" ]]
[[ -f "${fixture_root}/workspace/.claude/skills/code-review/SKILL.md" ]]
[[ -f "${fixture_root}/workspace/.agents/skills/dev-harness-lite/agents/openai.yaml" ]]
[[ -f "${fixture_root}/workspace/.claude/skills/dev-harness-lite/agents/openai.yaml" ]]
[[ -f "${fixture_root}/workspace/.agents/skills/code-review/agents/openai.yaml" ]]
[[ ! -e "${fixture_root}/workspace/.agents/skills/code-review/agents/oopenai.yaml" ]]
grep -Fq 'allow_implicit_invocation: false' \
    "${fixture_root}/workspace/.agents/skills/dev-harness-lite/agents/openai.yaml"
[[ -f "${fixture_root}/workspace/demo/.env" ]]

echo "init creates root agent instruction and bundled skill links: PASS"
