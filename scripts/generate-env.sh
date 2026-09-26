#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BOOTSTRAP_DIR="${ROOT_DIR}/infra/bootstrap"
ENV_FILE="${ENV_FILE:-${ROOT_DIR}/.env}"
CLI_CONFIG_FILE="${CLI_CONFIG_FILE:-${XDG_CONFIG_HOME:-${HOME}/.config}/spx-platform/config.env}"

if ! command -v terraform >/dev/null 2>&1; then
  printf 'terraform is required to generate %s\n' "$ENV_FILE" >&2
  exit 1
fi
if ! command -v openssl >/dev/null 2>&1; then
  printf 'openssl is required to generate a Django secret\n' >&2
  exit 1
fi

terraform_output() {
  local output_name="$1"
  local value
  if ! value="$(terraform -chdir="$BOOTSTRAP_DIR" output -raw "$output_name" 2>/dev/null)"; then
    printf 'Terraform output %q is unavailable. Run terraform apply first.\n' "$output_name" >&2
    exit 1
  fi
  if [[ -z "$value" ]]; then
    printf 'Terraform output %q is empty. Run terraform apply first.\n' "$output_name" >&2
    exit 1
  fi
  printf '%s' "$value"
}

existing_secret=""
if [[ -f "$ENV_FILE" ]]; then
  existing_secret="$(sed -n 's/^DJANGO_SECRET_KEY=//p' "$ENV_FILE" | head -n 1)"
fi
if [[ -z "$existing_secret" || "$existing_secret" == "replace-for-local-development" ]]; then
  existing_secret="$(openssl rand -hex 32)"
fi

tenant_id="$(terraform_output tenant_id)"
api_client_id="$(terraform_output api_client_id)"
api_scope="$(terraform_output api_scope)"
cli_client_id="$(terraform_output cli_client_id)"
redirect_uri="$(terraform_output redirect_uri)"

write_atomic() {
  local target="$1"
  local directory
  local temporary_file
  directory="$(dirname "$target")"
  mkdir -p "$directory"
  temporary_file="$(mktemp "${target}.tmp.XXXXXX")"
  cat > "$temporary_file"
  chmod 600 "$temporary_file"
  mv -f "$temporary_file" "$target"
}

umask 077
write_atomic "$ENV_FILE" <<EOF
DJANGO_SECRET_KEY=${existing_secret}
DJANGO_DEBUG=true
ENTRA_TENANT_ID=${tenant_id}
ENTRA_ALLOWED_AUDIENCES=${api_client_id}
ENTRA_REQUIRED_SCOPE=access_as_user
PLATFORM_API_URL=http://localhost:8000
PLATFORM_TENANT_ID=${tenant_id}
PLATFORM_CLI_CLIENT_ID=${cli_client_id}
PLATFORM_API_SCOPE=${api_scope}
PLATFORM_REDIRECT_URI=${redirect_uri}
EOF

write_atomic "$CLI_CONFIG_FILE" <<EOF
PLATFORM_API_URL=http://localhost:8000
PLATFORM_TENANT_ID=${tenant_id}
PLATFORM_CLI_CLIENT_ID=${cli_client_id}
PLATFORM_API_SCOPE=${api_scope}
PLATFORM_REDIRECT_URI=${redirect_uri}
EOF

printf 'Generated %s and %s with Terraform values.\n' "$ENV_FILE" "$CLI_CONFIG_FILE"
