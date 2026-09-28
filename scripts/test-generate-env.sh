#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
temporary_directory="$(mktemp -d)"
cleanup() { rm -rf "$temporary_directory"; }
trap cleanup EXIT

mkdir -p "$temporary_directory/bin"
cat > "$temporary_directory/bin/terraform" <<'FAKE_TERRAFORM'
#!/usr/bin/env bash
set -Eeuo pipefail
if [[ "${1:-}" == -chdir=* ]]; then shift; fi
if [[ "${1:-}" != output || "${2:-}" != -raw ]]; then exit 2; fi
case "${3:-}" in
  tenant_id) printf 'tenant-from-test' ;;
  api_client_id) printf 'api-client-from-test' ;;
  api_scope) printf 'api://test/access_as_user' ;;
  cli_client_id) printf 'cli-client-from-test' ;;
  redirect_uri) printf 'http://localhost:8765/callback' ;;
  acr_name) printf 'spxdemo12345678' ;;
  acr_login_server) printf 'spxdemo12345678.azurecr.io' ;;
  bootstrap_resource_group_name) printf 'spx-demo-bootstrap' ;;
  *) exit 1 ;;
esac
FAKE_TERRAFORM
chmod +x "$temporary_directory/bin/terraform"

export PATH="$temporary_directory/bin:${PATH}"
export ENV_FILE="$temporary_directory/.env"
export CLI_CONFIG_FILE="$temporary_directory/config.env"
printf 'DJANGO_SECRET_KEY=preserve-this-secret\n' > "$ENV_FILE"
printf 'WORKER_CALLBACK_TOKEN=preserve-this-worker-token\n' >> "$ENV_FILE"

"$ROOT_DIR/scripts/generate-env.sh" >/dev/null

grep -qx 'DJANGO_SECRET_KEY=preserve-this-secret' "$ENV_FILE"
grep -qx 'ENTRA_TENANT_ID=tenant-from-test' "$ENV_FILE"
grep -qx 'ENTRA_ALLOWED_AUDIENCES=api-client-from-test' "$ENV_FILE"
grep -qx 'PLATFORM_API_SCOPE=api://test/access_as_user' "$ENV_FILE"
grep -qx 'PLATFORM_ACR_NAME=spxdemo12345678' "$ENV_FILE"
grep -qx 'PLATFORM_ACR_LOGIN_SERVER=spxdemo12345678.azurecr.io' "$ENV_FILE"
grep -qx 'WORKER_CALLBACK_TOKEN=preserve-this-worker-token' "$ENV_FILE"
[[ "$(stat -c '%a' "$ENV_FILE")" == "600" ]]
grep -qx 'PLATFORM_TENANT_ID=tenant-from-test' "$CLI_CONFIG_FILE"
grep -qx 'PLATFORM_CLI_CLIENT_ID=cli-client-from-test' "$CLI_CONFIG_FILE"
if grep -q 'DJANGO_SECRET_KEY' "$CLI_CONFIG_FILE"; then
  printf 'CLI config must not contain the Django secret.\n' >&2
  exit 1
fi
if grep -q 'WORKER_CALLBACK_TOKEN' "$CLI_CONFIG_FILE"; then
  printf 'CLI config must not contain the worker token.\n' >&2
  exit 1
fi
[[ "$(stat -c '%a' "$CLI_CONFIG_FILE")" == "600" ]]

printf 'generate-env.sh checks passed.\n'
