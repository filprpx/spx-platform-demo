#!/usr/bin/env bash
set -Eeuo pipefail

missing=0

require_command() {
  local command_name="$1"
  local purpose="$2"
  if command -v "$command_name" >/dev/null 2>&1; then
    printf '✓ %-12s %s\n' "$command_name" "$purpose"
  else
    printf '✗ %-12s missing — %s\n' "$command_name" "$purpose" >&2
    missing=$((missing + 1))
  fi
}

require_command az 'Azure authentication and Entra bootstrap'
require_command terraform 'Entra bootstrap infrastructure'
require_command docker 'API container runtime'
require_command go 'CLI build and tests'
require_command python3 'host-side Celery worker runtime'
require_command openssl 'local secret and worker-token generation'

if command -v python3 >/dev/null 2>&1; then
  if python3 -m venv --help >/dev/null 2>&1; then
    printf '✓ %-12s Python virtual environments are available\n' 'python3 venv'
  else
    printf '✗ %-12s unavailable — install Python virtualenv support (for example, python3-venv)\n' 'python3 venv' >&2
    missing=$((missing + 1))
  fi
fi

if command -v docker >/dev/null 2>&1; then
  if docker compose version >/dev/null 2>&1; then
    printf '✓ %-12s Docker Compose is available\n' 'docker compose'
  else
    printf '✗ %-12s missing — install the Docker Compose plugin\n' 'docker compose' >&2
    missing=$((missing + 1))
  fi

  if docker info >/dev/null 2>&1; then
    printf '✓ %-12s Docker daemon is reachable\n' 'docker info'
  else
    printf '✗ %-12s unavailable — start Docker Desktop or the Docker daemon\n' 'docker info' >&2
    missing=$((missing + 1))
  fi
fi

if command -v az >/dev/null 2>&1; then
  if az account show >/dev/null 2>&1; then
    printf '✓ %-12s Azure CLI is authenticated\n' 'az login'
  else
    printf '✗ %-12s not authenticated — run: az login\n' 'az login' >&2
    missing=$((missing + 1))
  fi
fi

if [[ ":${PATH}:" != *":${HOME}/.local/bin:"* ]]; then
  printf '! %-12s %s\n' 'PATH' 'warning: ~/.local/bin is not on PATH; installed spx command may need an explicit PATH update'
fi

if (( missing > 0 )); then
  printf '\nDependency check failed with %d issue(s).\n' "$missing" >&2
  exit 1
fi

printf 'All required dependencies are available.\n'
