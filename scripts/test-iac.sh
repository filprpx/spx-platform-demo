#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MAKEFILE="${ROOT_DIR}/Makefile"

grep -q '^iac-pipeline:' "$MAKEFILE"
grep -q '^iac-destroy:' "$MAKEFILE"
grep -q '^deploy-app:' "$MAKEFILE"
grep -q 'scripts/execute-pipeline.py' "$MAKEFILE"
grep -q 'scripts/destroy-workloads.py' "$MAKEFILE"
grep -q 'scripts/deploy-app.py' "$MAKEFILE"
grep -q '$(MAKE) iac-destroy' "$MAKEFILE"
grep -q 'PLATFORM_ACR_NAME' "${ROOT_DIR}/scripts/generate-env.sh"
grep -q 'azurerm_container_registry' "${ROOT_DIR}/infra/bootstrap/main.tf"
grep -q 'azurerm_container_app' "${ROOT_DIR}/api/apps/provisioning/templates/terraform/main.tf.j2"
grep -q 'AcrPull' "${ROOT_DIR}/api/apps/provisioning/templates/terraform/main.tf.j2"
test -f "${ROOT_DIR}/examples/simple-api/Dockerfile"
test -f "${ROOT_DIR}/examples/simple-api/app.py"

if make --no-print-directory help | grep -q 'execute-pipeline'; then
  printf 'The old execute-pipeline target must not appear in help output.\n' >&2
  exit 1
fi

printf 'IaC and deployment target checks passed.\n'
