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
cp "${source_root}/.env.example" "${fixture_root}/workspace/dev/.env.example"

cat >"${fixture_root}/bin/git" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >>"${GIT_CALLS_FILE}"
[[ "$1" == "clone" ]]
mkdir -p "$3"
EOF
chmod +x "${fixture_root}/bin/git"

cat >"${fixture_root}/bin/go" <<'EOF'
#!/usr/bin/env bash
if [[ "${1:-} ${2:-}" == "work init" ]]; then
    : >go.work
fi
EOF
chmod +x "${fixture_root}/bin/go"

GIT_CALLS_FILE="${fixture_root}/git-calls" \
    PATH="${fixture_root}/bin:${PATH}" \
    "${fixture_root}/workspace/dev/dev" init

expected_calls="$(cat <<EOF
clone git@github.com:prismgo/framework.git ${fixture_root}/workspace/framework
clone git@github.com:prismgo/docs.git ${fixture_root}/workspace/docs
EOF
)"
[[ "$(cat "${fixture_root}/git-calls")" == "${expected_calls}" ]]
[[ -d "${fixture_root}/workspace/framework" ]]
[[ -d "${fixture_root}/workspace/docs" ]]

echo "init clones missing repositories: PASS"
