#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUNTIME_DIR="${ROOT_DIR}/.local/spx"
PID_FILE="${RUNTIME_DIR}/worker.pid"
LOG_FILE="${RUNTIME_DIR}/worker.log"

if [[ ! -f "${ROOT_DIR}/.env" ]]; then
  printf 'Missing .env; run make env first.\n' >&2
  exit 1
fi
if [[ ! -x "${ROOT_DIR}/api/.venv/bin/celery" ]]; then
  printf 'Missing API virtualenv; run make api-venv first.\n' >&2
  exit 1
fi

mkdir -p "$RUNTIME_DIR"

if [[ -f "$PID_FILE" ]]; then
  pid="$(<"$PID_FILE")"
  if [[ "$pid" =~ ^[0-9]+$ ]] && kill -0 "$pid" 2>/dev/null; then
    command_line="$(ps -p "$pid" -o command= 2>/dev/null || true)"
    if [[ "$command_line" == *"celery"* && "$command_line" == *"config.celery"* && "$command_line" == *"provisioning"* ]]; then
      printf 'Worker is already running (PID %s).\n' "$pid"
      exit 0
    fi
    printf 'Refusing to use PID file: PID %s belongs to another process.\n' "$pid" >&2
    exit 1
  fi
  rm -f "$PID_FILE"
fi

set -a
source "${ROOT_DIR}/.env"
set +a

(
  cd "${ROOT_DIR}/api"
  exec env \
    CELERY_BROKER_URL=redis://127.0.0.1:6379/0 \
    PLATFORM_API_URL=http://localhost:8000 \
    .venv/bin/celery -A config.celery worker --queues=provisioning --loglevel=INFO
) >>"$LOG_FILE" 2>&1 < /dev/null &
worker_pid=$!
printf '%s\n' "$worker_pid" > "$PID_FILE"

sleep 1
if ! kill -0 "$worker_pid" 2>/dev/null; then
  printf 'Worker failed to start. Check %s.\n' "$LOG_FILE" >&2
  rm -f "$PID_FILE"
  exit 1
fi

printf 'Started background worker (PID %s).\n' "$worker_pid"
printf 'Worker log: %s\n' "$LOG_FILE"
