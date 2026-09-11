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

generated_key="$(sed -n 's/^APP_KEY=//p' "${fixture_root}/workspace/demo/.env")"
[[ "${generated_key}" =~ ^base64:[A-Za-z0-9+/]{43}=$ ]]

sed -i 's|^APP_KEY=.*$|APP_KEY=base64:custom-key-must-be-preserved|' \
    "${fixture_root}/workspace/demo/.env"
PATH="${fixture_root}/bin:${PATH}" "${fixture_root}/workspace/demo/dev" init
grep -Fxq 'APP_KEY=base64:custom-key-must-be-preserved' \
    "${fixture_root}/workspace/demo/.env"

echo "init application key: PASS"
