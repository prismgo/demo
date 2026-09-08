#!/usr/bin/env bash
set -euo pipefail

source_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fixture_root="$(mktemp -d)"
trap 'rm -rf "${fixture_root}"' EXIT

mkdir -p "${fixture_root}/workspace/dev/.dev/scripts" "${fixture_root}/bin"
cp "${source_root}/dev" "${fixture_root}/workspace/dev/dev"
cp "${source_root}/.dev/scripts/test-environment.sh" \
    "${fixture_root}/workspace/dev/.dev/scripts/test-environment.sh"
cp "${source_root}/AGENTS.md" "${fixture_root}/workspace/dev/AGENTS.md"
cp -R "${source_root}/.agents" "${fixture_root}/workspace/dev/.agents"
printf 'keep me\n' >"${fixture_root}/workspace/AGENTS.md"

cat >"${fixture_root}/bin/git" <<'EOF'
#!/usr/bin/env bash
exit 99
EOF
chmod +x "${fixture_root}/bin/git"

if output="$(PATH="${fixture_root}/bin:${PATH}" \
    "${fixture_root}/workspace/dev/dev" init 2>&1)"; then
    echo "init unexpectedly accepted a conflicting AGENTS.md" >&2
    exit 1
fi

[[ "${output}" == *"Refusing to replace"* ]]
[[ "$(cat "${fixture_root}/workspace/AGENTS.md")" == "keep me" ]]
[[ ! -e "${fixture_root}/workspace/framework" ]]
[[ ! -e "${fixture_root}/workspace/docs" ]]

echo "init rejects conflicting instruction files: PASS"
