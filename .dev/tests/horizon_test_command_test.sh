#!/usr/bin/env bash
set -euo pipefail

source_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fixture_root="$(mktemp -d)"
trap 'rm -rf "${fixture_root}"' EXIT

mkdir -p \
    "${fixture_root}/workspace/demo/.dev/docker" \
    "${fixture_root}/workspace/demo/.dev/scripts" \
    "${fixture_root}/workspace/demo/app/demo/horizon" \
    "${fixture_root}/workspace/framework" \
    "${fixture_root}/bin"
cp "${source_root}/dev" "${fixture_root}/workspace/demo/dev"
cp "${source_root}/.dev/docker/defaults.env" \
    "${fixture_root}/workspace/demo/.dev/docker/defaults.env"
cp "${source_root}/.dev/docker/compose.yaml" \
    "${fixture_root}/workspace/demo/.dev/docker/compose.yaml"
cp "${source_root}/.dev/scripts/test-environment.sh" \
    "${fixture_root}/workspace/demo/.dev/scripts/test-environment.sh"

command_log="${fixture_root}/commands.log"
cat >"${fixture_root}/bin/docker" <<'EOF'
#!/usr/bin/env bash
printf 'docker cwd=%s args=%s\n' "${PWD}" "$*" >>"${COMMAND_LOG}"
EOF
cat >"${fixture_root}/bin/go" <<'EOF'
#!/usr/bin/env bash
printf 'go cwd=%s args=%s\n' "${PWD}" "$*" >>"${COMMAND_LOG}"
EOF
chmod +x "${fixture_root}/bin/docker" "${fixture_root}/bin/go"

COMMAND_LOG="${command_log}" PATH="${fixture_root}/bin:${PATH}" \
    "${fixture_root}/workspace/demo/dev" test-horizon

grep -Eq 'docker .*args=compose .* up -d --wait --wait-timeout 240 redis rabbitmq$' "${command_log}"
grep -Fq \
    "go cwd=${fixture_root}/workspace/demo args=test -count=1 -v -timeout=45s ./app/demo/horizon -run ^TestHorizonWithRealQueue$" \
    "${command_log}"

echo "test-horizon starts Redis and RabbitMQ and runs the Demo Horizon integration test: PASS"
